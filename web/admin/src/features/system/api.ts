import { api, APIError } from '../../lib/http'
import type { components } from '../../api/generated/schema'

type Schemas = components['schemas']
export type Administrator = Schemas['AdministratorRecord']
export type Role = Schemas['RoleRecord']
export type Permission = Schemas['PermissionRecord']
export type Query = {
  page: number
  pageSize: number
  keyword: string
  enabled: 'all' | 'enabled' | 'disabled'
}
const header = { 'X-GoSeed-CSRF': '1' as const }
async function unwrap<T>(
  result: Promise<{
    data?: { data: T }
    error?: { message: string; code: number; traceId: string }
    response: Response
  }>,
): Promise<T> {
  const { data, error, response } = await result
  if (!response.ok || error || !data)
    throw new APIError(
      error?.message || '操作失败',
      error?.code,
      error?.traceId,
    )
  return data.data
}
export const management = {
  administrators: (query: Query) =>
    unwrap(
      api.GET('/admin-api/v1/system/administrators', { params: { query } }),
    ),
  roles: (query: Query) =>
    unwrap(api.GET('/admin-api/v1/system/roles', { params: { query } })),
  permissions: () => unwrap(api.GET('/admin-api/v1/system/permissions')),
  createAdministrator: (body: Schemas['CreateAdministratorRequest']) =>
    unwrap(
      api.POST('/admin-api/v1/system/administrators', {
        params: { header },
        body,
      }),
    ),
  updateAdministrator: (
    id: string,
    body: Schemas['UpdateAdministratorRequest'],
  ) =>
    unwrap(
      api.PATCH('/admin-api/v1/system/administrators/{id}', {
        params: { path: { id }, header },
        body,
      }),
    ),
  assignRoles: (id: string, body: Schemas['AssignAdministratorRolesRequest']) =>
    unwrap(
      api.PUT('/admin-api/v1/system/administrators/{id}/roles', {
        params: { path: { id }, header },
        body,
      }),
    ),
  createRole: (body: Schemas['CreateRoleRequest']) =>
    unwrap(
      api.POST('/admin-api/v1/system/roles', { params: { header }, body }),
    ),
  updateRole: (id: string, body: Schemas['UpdateRoleRequest']) =>
    unwrap(
      api.PATCH('/admin-api/v1/system/roles/{id}', {
        params: { path: { id }, header },
        body,
      }),
    ),
  grantRole: (id: string, body: Schemas['GrantRoleRequest']) =>
    unwrap(
      api.PUT('/admin-api/v1/system/roles/{id}/permissions', {
        params: { path: { id }, header },
        body,
      }),
    ),
  deleteRole: (id: string, revision: string) =>
    unwrap(
      api.DELETE('/admin-api/v1/system/roles/{id}', {
        params: { path: { id }, query: { revision }, header },
      }),
    ),
}
