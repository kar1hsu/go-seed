import assert from 'node:assert/strict'
import { test } from 'node:test'
import { resolveObjectURL } from 'node:buffer'
import { fileURLToPath } from 'node:url'
import { createServer } from 'vite'

const server = await createServer({
  root: fileURLToPath(new URL('../', import.meta.url)),
  configFile: false,
  optimizeDeps: { noDiscovery: true, include: [] },
  server: { middlewareMode: true, hmr: false },
  appType: 'custom',
})
const record = (id, contentType, size) => ({
  id,
  contentType,
  size,
  name: 'original.txt',
  status: 'ready',
  source: 'upload',
})
function deferred() {
  let resolve, reject
  const promise = new Promise((yes, no) => {
    resolve = yes
    reject = no
  })
  return { promise, resolve, reject }
}
try {
  const { createFilePreview, TEXT_PREVIEW_LIMIT } = await server.ssrLoadModule(
    '/src/features/system/files/preview.ts',
  )
  const { APIError } = await server.ssrLoadModule('/src/lib/http.ts')

  await test('switching files aborts the old read and discards its late content', async () => {
    const old = deferred(),
      current = deferred(),
      signals = []
    const state = createFilePreview((id, signal) => {
      signals.push(signal)
      return id === '1' ? old.promise : current.promise
    })
    const first = state.open(record('1', 'text/plain', 3))
    const second = state.open(record('2', 'text/plain', 3))
    assert.equal(signals[0].aborted, true)
    current.resolve(new Blob(['new']))
    await second
    old.resolve(new Blob(['old']))
    await first
    assert.equal(state.text.value, 'new')
    assert.equal(state.selected.value.id, '2')
    state.close()
    assert.equal(state.text.value, '')
  })

  await test('closing clears private content and prevents a pending read from reopening it', async () => {
    const late = deferred()
    let signal
    const state = createFilePreview((_, value) => {
      signal = value
      return late.promise
    })
    const loading = state.open(record('1', 'image/png', 3))
    state.close()
    assert.equal(signal.aborted, true)
    late.resolve(new Blob(['png']))
    await loading
    assert.equal(state.selected.value, null)
    assert.equal(state.url.value, '')
    assert.equal(state.loading.value, false)
  })

  await test('closing during text decoding discards the delayed decoded text', async () => {
    const late = deferred(),
      started = deferred()
    const state = createFilePreview(async () => ({
      size: 3,
      slice: () => ({
        arrayBuffer: () => {
          started.resolve()
          return late.promise
        },
      }),
    }))
    const opening = state.open(record('1', 'text/plain', 3))
    await started.promise
    state.close()
    late.resolve(new TextEncoder().encode('old'))
    await opening
    assert.equal(state.text.value, '')
  })

  await test('only allowlisted MIME types get preview URLs, and close revokes them', async () => {
    for (const type of ['image/png', 'image/jpeg']) {
      const state = createFilePreview(
        async () => new Blob(['abc'], { type: 'application/octet-stream' }),
      )
      await state.open(record('1', type, 3))
      const url = state.url.value
      assert.equal(resolveObjectURL(url).type, type)
      assert.equal(await resolveObjectURL(url).text(), 'abc')
      state.close()
      assert.equal(resolveObjectURL(url), undefined)
    }
    let calls = 0
    const state = createFilePreview(async () => {
      calls++
      throw new Error('must not read')
    })
    for (const type of ['application/zip', 'text/html', 'image/svg+xml']) {
      await state.open(record('1', type, 3))
      assert.equal(state.kind.value, 'unsupported')
      assert.equal(state.url.value, '')
    }
    await state.open({ ...record('1', 'image/png', 3), status: 'pending' })
    assert.ok(state.error.value)
    assert.equal(calls, 0)
  })

  await test('PDF data stays local and is released on close without exposing a URL', async () => {
    const state = createFilePreview(async () => new Blob(['%PDF-test']))
    await state.open(record('1', 'application/pdf', 9))
    assert.equal(new TextDecoder().decode(state.pdfData.value), '%PDF-test')
    assert.equal(state.url.value, '')
    state.close()
    assert.equal(state.pdfData.value, null)
  })

  await test('bounded text preview keeps UTF-8 intact and treats markup as plain text', async () => {
    const value = 'x'.repeat(TEXT_PREVIEW_LIMIT - 1) + '中文'
    const blob = new Blob([value])
    const state = createFilePreview(async () => blob)
    await state.open(record('1', 'text/plain', blob.size))
    assert.equal(state.text.value, 'x'.repeat(TEXT_PREVIEW_LIMIT - 1))
    assert.equal(state.truncated.value, true)
    assert.equal(state.url.value, '')
    for (const type of ['text/plain', 'text/csv', 'application/json']) {
      const text = '<script>alert(1)</script>'
      const small = createFilePreview(async () => new Blob([text]))
      await small.open(record('1', type, text.length))
      assert.equal(small.text.value, text)
      assert.equal(small.url.value, '')
      assert.equal(small.truncated.value, false)
    }
  })

  await test('failed and truncated downloads cannot produce a preview, and retry can recover', async () => {
    let count = 0
    const state = createFilePreview(async () => {
      if (++count === 1) throw new APIError('forbidden', 100403)
      return new Blob(['ok'])
    })
    await state.open(record('1', 'text/plain', 2))
    assert.ok(state.error.value)
    assert.equal(state.text.value, '')
    await state.open(record('1', 'text/plain', 2))
    assert.equal(state.text.value, 'ok')
    assert.equal(state.error.value, '')
    await state.open(record('1', 'image/png', 10))
    assert.match(state.error.value, /不完整/)
    assert.equal(state.url.value, '')
    assert.equal(state.text.value, '')
  })
} finally {
  await server.close()
}
