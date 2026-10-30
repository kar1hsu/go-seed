import { createRouter, createWebHistory } from 'vue-router'
import { useAuthStore } from '../features/auth/store'
import { ElMessage, ElMessageBox } from 'element-plus'
import { guard } from './guard'
import AdminLayout from '../layouts/AdminLayout.vue'
import { navigation } from './navigation'

// A compile-time allowlist: declarative menus never resolve arbitrary component paths.
const views = {
  members: () => import('../features/member/users/MembersPage.vue'),
  files: () => import('../features/system/files/FilesPage.vue'),
  tasks: () => import('../features/system/files/TasksPage.vue'),
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

router.beforeEach((to) =>
  guard(to, useAuthStore(), (message) => ElMessage.error(message)),
)
let loadingNotice = false
router.onError((error) => {
  if (loadingNotice) return
  loadingNotice = true
  const chunk =
    /dynamically imported module|Loading chunk|Importing a module script/i.test(
      String(error),
    )
  void ElMessageBox.confirm(
    chunk
      ? '页面资源加载失败，可能有新版本发布。请检查网络后重新加载。'
      : '页面暂时无法打开，请重试。',
    '页面加载失败',
    {
      confirmButtonText: '重新加载',
      cancelButtonText: '取消',
      type: 'warning',
    },
  )
    .then(() => window.location.reload())
    .catch(() => {})
    .finally(() => {
      loadingNotice = false
    })
})
