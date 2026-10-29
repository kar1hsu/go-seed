import type { Middleware } from 'openapi-fetch'
import type { useAuthStore } from './store'
import { APIError } from '../../lib/http'

const publicAuth = new Set([
  '/admin-api/v1/auth/login',
  '/admin-api/v1/auth/refresh',
  '/admin-api/v1/auth/logout',
])

export function authMiddleware(
  auth: ReturnType<typeof useAuthStore>,
  onExpired: () => void,
): Middleware {
  return {
    async onRequest({ request }) {
      const path = new URL(request.url).pathname
      if (!['GET', 'HEAD'].includes(request.method))
        request.headers.set('X-GoSeed-CSRF', '1')
      if (
        !publicAuth.has(path) &&
        path !== '/admin-api/v1/system/ping' &&
        !request.headers.has('Authorization')
      ) {
        try {
          request.headers.set(
            'Authorization',
            `Bearer ${await auth.ensureAccess()}`,
          )
        } catch (error) {
          if (error instanceof APIError && error.code === 100401) onExpired()
          throw error
        }
      }
      return request
    },
    onResponse({ request, response }) {
      if (
        response.status === 401 &&
        request.headers.get('Authorization') === `Bearer ${auth.accessToken}` &&
        !publicAuth.has(new URL(request.url).pathname)
      ) {
        auth.clear()
        onExpired()
      }
      if (
        response.status === 403 &&
        new URL(request.url).pathname !== '/admin-api/v1/auth/me'
      )
        void auth.current().catch(() => {})
      return response
    },
  }
}
