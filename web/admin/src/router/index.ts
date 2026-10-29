import { createRouter, createWebHistory } from 'vue-router'
import { useAuthStore, authMessage } from '../features/auth/store'
import AdminLayout from '../layouts/AdminLayout.vue'

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
        {
          path: 'account/security',
          name: 'security',
          component: () => import('../features/auth/pages/SecurityPage.vue'),
        },
        {
          path: '',
          name: 'dashboard',
          component: () =>
            import('../features/dashboard/pages/DashboardPage.vue'),
        },
        {
          path: 'monitor',
          name: 'monitor',
          component: () =>
            import('../features/dashboard/pages/MonitorPage.vue'),
        },
        {
          path: 'api-docs',
          name: 'api-docs',
          component: () => import('../features/api-docs/pages/ApiDocsPage.vue'),
        },
        {
          path: 'about',
          name: 'about',
          component: () => import('../features/about/pages/AboutPage.vue'),
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
