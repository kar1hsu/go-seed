import assert from 'node:assert/strict'
import { test } from 'node:test'
import { fileURLToPath } from 'node:url'
import { createServer } from 'vite'
const server = await createServer({ root: fileURLToPath(new URL('../', import.meta.url)), configFile: false, optimizeDeps: { noDiscovery: true, include: [] }, server: { middlewareMode: true, hmr: false }, appType: 'custom' })
try {
  const { closeTabs } = await server.ssrLoadModule('/src/layouts/tabs.ts')
  const { validateSettingValue } = await server.ssrLoadModule('/src/features/system/settings/form.ts')
  test('tab closing retains the workbench and chooses the nearest surviving page', () => {
    const paths = ['/', '/a', '/b', '/c']
    assert.deepEqual(closeTabs(paths, '/b', '/b', 'close'), { paths: ['/', '/a', '/c'], active: '/a' })
    assert.deepEqual(closeTabs(paths, '/c', '/a', 'close'), { paths: ['/', '/b', '/c'], active: '/c' })
    assert.deepEqual(closeTabs(paths, '/a', '/c', 'others'), { paths: ['/', '/c'], active: '/c' })
    assert.deepEqual(closeTabs(paths, '/b', '/a', 'others'), { paths: ['/', '/a'], active: '/a' })
    assert.deepEqual(closeTabs(paths, '/c', '/a', 'right'), { paths: ['/', '/a'], active: '/a' })
    assert.deepEqual(closeTabs(paths, '/b', '/', 'close'), { paths, active: '/b' })
    assert.deepEqual(closeTabs(paths, '/c', '/a', 'all'), { paths: ['/'], active: '/' })
    assert.deepEqual(closeTabs(paths, '/c', '/missing', 'all'), { paths, active: '/c' })
  })
  test('configuration input keeps large integers exact and validates UTF-8 size and JSON', () => {
    assert.equal(validateSettingValue('integer', '9223372036854775807'), '')
    assert.notEqual(validateSettingValue('integer', '9223372036854775808'), '')
    assert.notEqual(validateSettingValue('integer', '1.5'), '')
    assert.notEqual(validateSettingValue('json', '{broken}'), '')
    assert.equal(validateSettingValue('json', '{"a":9007199254740993}'), '')
    assert.notEqual(validateSettingValue('text', '中'.repeat(1366)), '')
  })
} finally { await server.close() }
