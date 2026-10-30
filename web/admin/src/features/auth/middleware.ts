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
    async onResponse({ request, response }) {
      if (
        response.status === 401 &&
        request.headers.get('Authorization') === `Bearer ${auth.accessToken}` &&
        !publicAuth.has(new URL(request.url).pathname)
      ) {
        const failed = auth.accessToken
        const recovered = await auth.recover(failed)
        if (!recovered) {
          onExpired()
          return response
        }
        if (['GET', 'HEAD'].includes(request.method)) {
          const retry = new Request(request, { headers: request.headers })
          retry.headers.set('Authorization', `Bearer ${auth.accessToken}`)
          const retryToken = auth.accessToken
          const result = await fetch(retry, {
            credentials: 'same-origin',
            signal: AbortSignal.any([
              request.signal,
              AbortSignal.timeout(5000),
            ]),
          })
          if (result.status === 401 && retryToken === auth.accessToken) {
            auth.clear()
            onExpired()
          }
          return result
        }
        // Mutations are never repeated automatically: the user can retry explicitly.
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
