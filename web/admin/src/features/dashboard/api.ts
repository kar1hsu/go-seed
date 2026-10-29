import { api, APIError } from '../../lib/http'

export async function probeSystem() {
  const { data, error, response } = await api.GET('/admin-api/v1/system/ping')
  if (!response.ok || error) {
    throw new APIError(
      error?.message || '服务暂时不可用',
      error?.code,
      error?.traceId,
    )
  }
  if (!data || data.code !== 0)
    throw new APIError(data?.message || '响应格式不正确')
  return data
}
