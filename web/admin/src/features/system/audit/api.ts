import { api, APIError } from '../../../lib/http'
import type { components } from '../../../api/generated/schema'

type Schemas = components['schemas']
export type AuditLog = Schemas['AuditLogRecord']
export type AuditDetail = Schemas['AuditLogDetailData']
export type AuditAction = Schemas['AuditAction']
export type AuditQuery = {
  page: number
  pageSize: number
  keyword: string
  action: string
  actorId: string
  traceId: string
  from: string
  to: string
}
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
      error?.message || '查询失败',
      error?.code,
      error?.traceId,
    )
  return data.data
}
export const auditAPI = {
  list: (query: AuditQuery) =>
    unwrap(api.GET('/admin-api/v1/system/audit-logs', { params: { query } })),
  detail: (id: string) =>
    unwrap(
      api.GET('/admin-api/v1/system/audit-logs/{id}', {
        params: { path: { id } },
      }),
    ),
}
