import assert from 'node:assert/strict'
import { test } from 'node:test'
import { fileURLToPath } from 'node:url'
import { createServer } from 'vite'

const server = await createServer({
  root: fileURLToPath(new URL('../', import.meta.url)),
  configFile: false,
  optimizeDeps: { noDiscovery: true, include: [] },
  server: { middlewareMode: true, hmr: false },
  appType: 'custom',
})
try {
  const { createAuditState } = await server.ssrLoadModule(
    '/src/features/system/audit/state.ts',
  )
  function deferred() {
    let resolve, reject
    const promise = new Promise((yes, no) => {
      resolve = yes
      reject = no
    })
    return { promise, resolve, reject }
  }
  const response = (id) => ({ items: [{ id }], actions: [], total: 1 })
  test('a slower old query cannot replace the latest audit results', async () => {
    const a = deferred(),
      b = deferred()
    let calls = 0
    const state = createAuditState({
      list: () => (++calls === 1 ? a.promise : b.promise),
    })
    const first = state.load(),
      second = state.load()
    b.resolve(response('new'))
    await second
    a.resolve(response('old'))
    await first
    assert.equal(state.items.value[0].id, 'new')
  })
  test('closing or switching the drawer discards late sensitive detail', async () => {
    const a = deferred(),
      b = deferred()
    const state = createAuditState({
      detail: (id) => (id === '1' ? a.promise : b.promise),
    })
    const first = state.open('1'),
      second = state.open('2')
    b.resolve({ record: { id: '2' }, changes: [] })
    await second
    a.resolve({ record: { id: '1' }, changes: [] })
    await first
    assert.equal(state.detail.value.record.id, '2')
    const late = deferred()
    const closed = createAuditState({ detail: () => late.promise })
    const pending = closed.open('3')
    closed.close()
    late.resolve({ record: { id: '3' }, changes: [] })
    await pending
    assert.equal(closed.detail.value, null)
    assert.equal(closed.selectedId.value, '')
  })
  test('revocation/disposal clears displayed data and invalidates pending reads', async () => {
    const late = deferred()
    let calls = 0
    const state = createAuditState({
      list: () =>
        ++calls === 1 ? Promise.resolve(response('visible')) : late.promise,
    })
    await state.load()
    const pending = state.load()
    state.clear()
    late.resolve(response('late'))
    await pending
    assert.deepEqual(state.items.value, [])
    assert.equal(state.total.value, 0)
  })
  test('failed list reads remove previous results and display an error', async () => {
    let calls = 0
    const state = createAuditState({
      list: async () => {
        if (++calls === 1) return response('old')
        throw new Error('unavailable')
      },
    })
    await state.load()
    await state.load()
    assert.deepEqual(state.items.value, [])
    assert.ok(state.error.value)
    assert.equal(state.loading.value, false)
  })
} finally {
  await server.close()
}
