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
    middleware.onResponse({
      request: new Request('http://test/admin-api/v1/auth/me', {
        headers: { Authorization: `Bearer ${oldToken}` },
      }),
      response: new Response('', { status: 401 }),
    })
    assert.equal(auth.admin.id, '2')
    assert.equal(expired, 0)
    middleware.onResponse({
      request: new Request('http://test/admin-api/v1/auth/me', {
        headers: { Authorization: `Bearer ${auth.accessToken}` },
      }),
      response: new Response('', { status: 401 }),
    })
    assert.equal(auth.admin, null)
    assert.equal(expired, 1)
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
} finally {
  await server.close()
}
