import type { RouteLocationNormalized } from 'vue-router'
import type { useAuthStore } from '../features/auth/store'
import { APIError } from '../lib/http'
import { authMessage } from '../features/auth/store'
export async function guard(
  to: RouteLocationNormalized,
  auth: ReturnType<typeof useAuthStore>,
  notify: (message: string) => void,
) {
  await auth.restore()
  if (to.name === 'login') return auth.admin ? '/' : true
  if (!to.meta.requiresAuth) return true
  if (!auth.admin) return { path: '/login', query: { redirect: to.fullPath } }
  try {
    await auth.current()
    if (!auth.admin)
      return {
        path: '/login',
        query: { redirect: to.fullPath, reason: 'expired' },
      }
    if (typeof to.meta.permission === 'string' && !auth.has(to.meta.permission))
      return '/forbidden'
  } catch (error) {
    auth.connectionError = authMessage(error)
    if (error instanceof APIError && error.code === 100401) {
      auth.clear()
      return {
        path: '/login',
        query: { redirect: to.fullPath, reason: 'expired' },
      }
    }
    notify(auth.connectionError)
    return false
  }
  return true
}
