import assert from 'node:assert/strict'
import { test } from 'node:test'
import { fileURLToPath } from 'node:url'
import { createServer } from 'vite'
import { createPinia, setActivePinia } from 'pinia'

// Load the real TypeScript store without introducing a second test runtime.
const server = await createServer({
  root: fileURLToPath(new URL('../', import.meta.url)),
  configFile: false,
  optimizeDeps: { noDiscovery: true, include: [] },
  server: { middlewareMode: true, hmr: false },
  appType: 'custom',
})
try {
  const { useAuthStore } = await server.ssrLoadModule(
    '/src/features/auth/store.ts',
  )
  const { authMiddleware } = await server.ssrLoadModule(
    '/src/features/auth/middleware.ts',
  )
  const { api } = await server.ssrLoadModule('/src/lib/http.ts')
  const profile = (id) => ({
    id,
    username: `admin_${id}`,
    displayName: `Admin ${id}`,
  })
  const ok = (data) => ({
    data: { data },
    response: new Response('', { status: 200 }),
  })
  const current = (id, version, permissions) =>
    ok({
      admin: profile(id),
      policyVersion: version,
      permissions,
      roles: [],
      superAdmin: false,
    })
  let token = 0
  api.POST = async (_, request) =>
    ok({
      admin: profile(request.body.username),
      accessToken: `token-${++token}`,
      expiresAt: new Date(Date.now() + 900_000).toISOString(),
    })
  const login = async (auth) => auth.login('1', 'test-only')
  function deferred() {
    let resolve
    const promise = new Promise((done) => {
      resolve = done
    })
    return { promise, resolve }
  }
  async function pending(auth, responses) {
    let received
    const started = new Promise((done) => {
      received = done
    })
    api.GET = () => {
      received()
      return responses.promise
    }
    const operation = auth.current()
    await started
    return { operation }
  }
  await test('a delayed current-user response cannot resurrect a signed-out session', async () => {
    setActivePinia(createPinia())
    const auth = useAuthStore()
    await login(auth)
    const response = deferred()
    const { operation } = await pending(auth, response)
    auth.clear()
    response.resolve(current('1', '10', ['system.user.read']))
    await operation
    assert.equal(auth.admin, null)
    assert.deepEqual(auth.permissions, [])
  })
  await test('an old account response and its 401 cannot replace a new login', async () => {
    setActivePinia(createPinia())
    const auth = useAuthStore()
    await login(auth)
    const oldToken = auth.accessToken
    const response = deferred()
    const { operation } = await pending(auth, response)
    await auth.login('2', 'test-only')
    response.resolve(current('1', '10', ['system.user.read']))
    await operation
    let expired = 0
    const middleware = authMiddleware(auth, () => {
      expired++
    })
    await middleware.onResponse({
      request: new Request('http://test/admin-api/v1/auth/me', {
        headers: { Authorization: `Bearer ${oldToken}` },
      }),
      response: new Response('', { status: 401 }),
    })
    assert.equal(auth.admin.id, '2')
    assert.equal(expired, 0)
    const originalPost = api.POST
    api.POST = async () => ({ response: new Response('',{status:401}) })
    await middleware.onResponse({
      request: new Request('http://test/admin-api/v1/auth/me', {
        headers: { Authorization: `Bearer ${auth.accessToken}` },
      }),
      response: new Response('', { status: 401 }),
    })
    assert.equal(auth.admin, null)
    assert.equal(expired, 1)
    api.POST = originalPost
  })
  await test('an older poll cannot restore revoked permissions after a newer snapshot', async () => {
    setActivePinia(createPinia())
    const auth = useAuthStore()
    await login(auth)
    const stale = deferred()
    const first = await pending(auth, stale)
    api.GET = async () => current('1', '21', [])
    await auth.current()
    stale.resolve(current('1', '20', ['system.user.read']))
    await first.operation
    assert.equal(auth.policyVersion, '21')
    assert.equal(auth.has('system.user.read'), false)
  })
  await test('temporary route errors preserve the session and cancel navigation', async () => {
    const { guard } = await server.ssrLoadModule('/src/router/guard.ts')
    const { APIError } = await server.ssrLoadModule('/src/lib/http.ts')
    setActivePinia(createPinia()); const auth = useAuthStore();await login(auth)
    api.GET = async () => {throw new APIError('unavailable',100503)}
    let notice=''; const to={name:'members',fullPath:'/member/users?page=2',meta:{requiresAuth:true}}
    assert.equal(await guard(to,auth,(value)=>{notice=value}),false)
    assert.ok(auth.admin);assert.ok(notice)
    api.GET=async()=>{throw new APIError('expired',100401)}
    assert.equal((await guard(to,auth,()=>{})).query.redirect,to.fullPath)
    assert.equal(auth.admin,null)
  })
  await test('server Date controls expiry even when the client clock is slow', async () => {
    const originalPost=api.POST
    let refreshes=0;const now=Date.now()
    api.POST=async(path)=>{if(path.endsWith('/refresh'))refreshes++;return {data:{data:{admin:profile('1'),accessToken:'clock-token',expiresAt:new Date(now+10_000).toISOString()}},response:new Response('',{headers:{Date:new Date(now).toUTCString()}})}}
    const originalNow=Date.now;Date.now=()=>now-3600_000
    try {setActivePinia(createPinia());const auth=useAuthStore();await login(auth);await auth.ensureAccess();assert.equal(refreshes,1)}finally{Date.now=originalNow;api.POST=originalPost}
  })
  await test('lost refresh response retries with the same random key and late refresh cannot restore logout', async () => {
    const originalPost=api.POST;setActivePinia(createPinia());const auth=useAuthStore();const keys=[]
    api.POST=async(path,req)=>{keys.push(req.params.header['X-GoSeed-Refresh-Key']);throw new TypeError('network lost')}
    await assert.rejects(auth.ensureAccess());await assert.rejects(auth.ensureAccess());assert.equal(keys[0],keys[1]);assert.equal(keys[0].length,43)
    const wait=deferred();api.POST=()=>wait.promise;const operation=auth.ensureAccess();await new Promise(resolve=>setImmediate(resolve));auth.clear();wait.resolve(ok({admin:profile('1'),accessToken:'late',expiresAt:new Date(Date.now()+900_000).toISOString()}));await assert.rejects(operation);assert.equal(auth.admin,null);api.POST=originalPost
  })
  await test('401 recovery retries reads once but never repeats writes',async()=>{
    const originalPost=api.POST;const originalFetch=globalThis.fetch
    setActivePinia(createPinia());const auth=useAuthStore();await login(auth)
    api.POST=async()=>ok({admin:profile('1'),accessToken:'renewed',expiresAt:new Date(Date.now()+900_000).toISOString()})
    let reads=0;globalThis.fetch=async(req)=>{reads++;assert.equal(req.headers.get('Authorization'),'Bearer renewed');return new Response('{}',{status:200})}
    const middleware=authMiddleware(auth,()=>assert.fail('valid refresh must not sign out'))
    const req=(method)=>new Request('http://test/admin-api/v1/system/users',{method,headers:{Authorization:`Bearer ${auth.accessToken}`}})
    try{assert.equal((await middleware.onResponse({request:req('GET'),response:new Response('',{status:401})})).status,200);await middleware.onResponse({request:req('POST'),response:new Response('',{status:401})});assert.equal(reads,1)}finally{api.POST=originalPost;globalThis.fetch=originalFetch}
  })

} finally {
  await server.close()
}
