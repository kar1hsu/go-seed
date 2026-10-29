import { createRouter, createWebHistory } from 'vue-router'
import AdminLayout from '../layouts/AdminLayout.vue'

export const router = createRouter({
  history: createWebHistory(),
  routes: [
    {
      path: '/',
      component: AdminLayout,
      children: [
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
