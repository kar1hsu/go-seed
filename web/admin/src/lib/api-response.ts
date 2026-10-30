import { APIError } from './http'
export async function unwrap<T>(
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
