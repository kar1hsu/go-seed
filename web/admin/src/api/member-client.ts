import createClient from 'openapi-fetch'
import type { paths } from './generated/app-schema'

/** Independent C-end client. Keep its access token in memory, separate from admin state.
 * Refresh/login are explicit; serialize refresh requests because refresh tokens rotate.
 * Use one X-GoSeed-Refresh-Key per rotation; retain it across unknown network results
 * and reuse it within 60s to recover the same successor. Never replay writes automatically.
 * Use a same-origin reverse proxy for the HttpOnly member cookie.
 */
export function createMemberClient(getAccessToken: () => string | undefined) {
  const client = createClient<paths>({
    baseUrl: '',
    credentials: 'same-origin',
  })
  client.use({
    onRequest({ request }) {
      const token = getAccessToken()
      if (token) request.headers.set('Authorization', `Bearer ${token}`)
      if (!['GET', 'HEAD'].includes(request.method)) {
        request.headers.set('X-GoSeed-CSRF', '1')
      }
      return request
    },
  })
  return client
}
