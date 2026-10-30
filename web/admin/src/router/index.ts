import { createRouter, createWebHistory } from 'vue-router'
import { useAuthStore, authMessage } from '../features/auth/store'
import AdminLayout from '../layouts/AdminLayout.vue'
import { navigation } from './navigation'

// A compile-time allowlist: declarative menus never resolve arbitrary component paths.
const views = {
  dashboard: () => import('../features/dashboard/pages/DashboardPage.vue'),
  monitor: () => import('../features/dashboard/pages/MonitorPage.vue'),
  administrators: () =>
    import('../features/system/pages/AdministratorsPage.vue'),
  roles: () => import('../features/system/pages/RolesPage.vue'),
  dictionaries: () =>
    import('../features/system/settings/DictionariesPage.vue'),
  settings: () => import('../features/system/settings/SettingsPage.vue'),
  auditLogs: () => import('../features/system/audit/AuditLogsPage.vue'),
  apiDocs: () => import('../features/api-docs/pages/ApiDocsPage.vue'),
  about: () => import('../features/about/pages/AboutPage.vue'),
  security: () => import('../features/auth/pages/SecurityPage.vue'),
}

export const router = createRouter({
  history: createWebHistory(),
  routes: [
    {
      path: '/login',
      name: 'login',
      component: () => import('../features/auth/pages/LoginPage.vue'),
      meta: { public: true },
    },
    {
      path: '/',
      component: AdminLayout,
      meta: { requiresAuth: true },
      children: [
        ...navigation.map((item) => ({
          path: item.path.slice(1),
          name: item.component,
          component: views[item.component],
          meta: { permission: item.permission, title: item.title },
        })),
        {
          path: 'forbidden',
          name: 'forbidden',
          component: () => import('../features/auth/pages/ForbiddenPage.vue'),
        },
      ],
    },
    { path: '/:pathMatch(.*)*', redirect: '/' },
  ],
  scrollBehavior: () => ({ top: 0 }),
})

router.beforeEach(async (to) => {
  const auth = useAuthStore()
  await auth.restore()
  if (to.name === 'login') {
    document.title = '登录 - GoSeed'
    if (auth.admin) return '/'
    return true
  }
  if (to.meta.requiresAuth) {
    if (!auth.admin) return { path: '/login', query: { redirect: to.fullPath } }
    try {
      await auth.current()
      if (
        typeof to.meta.permission === 'string' &&
        !auth.has(to.meta.permission)
      )
        return '/forbidden'
    } catch (error) {
      auth.connectionError = authMessage(error)
      auth.clear()
      return {
        path: '/login',
        query: { redirect: to.fullPath, reason: 'expired' },
      }
    }
  }
  return true
})
