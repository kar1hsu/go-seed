import createClient from 'openapi-fetch'
import type { paths } from '../api/generated/schema'

export const api = createClient<paths>({
  baseUrl: '',
  fetch: (request) =>
    fetch(request, {
      signal: AbortSignal.any([
        request.signal,
        AbortSignal.timeout(
          /\/system\/files\/\d+\/content$/.test(new URL(request.url).pathname)
            ? 60000
            : new URL(request.url).pathname.endsWith('/auth/refresh')
              ? 15000
              : 5000,
        ),
      ]),
      credentials: 'same-origin',
    }),
})

export class APIError extends Error {
  constructor(
    message: string,
    readonly code?: number,
    readonly traceId?: string,
  ) {
    super(message)
  }
}
