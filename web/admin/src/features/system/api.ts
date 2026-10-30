import { api } from '../../lib/http'
import type { components } from '../../api/generated/schema'

type Schemas = components['schemas']
export type Administrator = Schemas['AdministratorRecord']
export type Role = Schemas['RoleRecord']
export type Permission = Schemas['PermissionRecord']
import type { Query } from '../../lib/list-query'
export type { Query } from '../../lib/list-query'
import { unwrap } from '../../lib/api-response'
export { unwrap } from '../../lib/api-response'
const header = { 'X-GoSeed-CSRF': '1' as const }
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
