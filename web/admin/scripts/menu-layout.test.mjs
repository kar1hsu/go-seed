import assert from 'node:assert/strict'
import { test } from 'node:test'
import { fileURLToPath } from 'node:url'
import { createServer } from 'vite'
import { createPinia, setActivePinia } from 'pinia'

const server = await createServer({
  root: fileURLToPath(new URL('../', import.meta.url)),
  configFile: false,
  optimizeDeps: { noDiscovery: true, include: [] },
  server: { middlewareMode: true, hmr: false },
  appType: 'custom',
})
try {
  const { presentationGroups } = await server.ssrLoadModule(
    '/src/router/menu-layout.ts',
  )
  const { useAuthStore } = await server.ssrLoadModule(
    '/src/features/auth/store.ts',
  )
  const { menuIcon } = await server.ssrLoadModule('/src/router/navigation.ts')
  const group = {
    id: 'ops',
    title: '运营组',
    icon: 'Key',
    hidden: false,
    pages: [
      { path: '/system/roles', title: '无权访问', icon: 'Menu', hidden: false },
      { path: '/new-script', title: '未知页面', icon: 'Menu', hidden: false },
      { path: '/monitor', title: '服务监控', icon: 'Monitor', hidden: true },
      { path: '/', title: '入口', icon: 'Setting', hidden: false },
    ],
  }
  await test('sidebar and tab presentation use only authorized registered routes in saved order', () => {
    const [result] = presentationGroups(
      { revision: '2', groups: [group] },
      (code) => !code,
    )
    assert.equal(result.title, '运营组')
    assert.equal(result.icon, menuIcon('Key'))
    assert.deepEqual(
      result.items.map((p) => [p.path, p.title, p.hidden]),
      [
        ['/monitor', '服务监控', true],
        ['/', '入口', false],
      ],
    )
    assert.equal(result.items[1].group, '运营组')
    assert.equal(result.items[1].icon, menuIcon('Setting'))
  })
  await test('saved navigation is immediately available, older revisions cannot undo it and logout clears it', () => {
    setActivePinia(createPinia())
    const auth = useAuthStore()
    auth.menuSaved({ revision: '3', groups: [group] })
    auth.menuSaved({ revision: '2', groups: [] })
    assert.equal(auth.navigation.groups[0].title, '运营组')
    auth.clear()
    assert.equal(auth.navigation, null)
  })
} finally {
  await server.close()
}
