import { api } from '../../../lib/http'
import { unwrap } from '../../../lib/api-response'
import type { Query } from '../../../lib/list-query'
import type { components } from '../../../api/generated/schema'
type Schemas = components['schemas']
export type MemberUser = Schemas['MemberUserRecord']
export type MemberDetail = Schemas['MemberUserDetail']
export const members = {
  list: (query: Query) =>
    unwrap(api.GET('/admin-api/v1/member/users', { params: { query } })),
  detail: (id: string) =>
    unwrap(
      api.GET('/admin-api/v1/member/users/{id}', { params: { path: { id } } }),
    ),
  status: (user: MemberUser, enabled: boolean) =>
    unwrap(
      api.PUT('/admin-api/v1/member/users/{id}/status', {
        params: { path: { id: user.id }, header: { 'X-GoSeed-CSRF': '1' } },
        body: { enabled, revision: user.revision },
      }),
    ),
}
