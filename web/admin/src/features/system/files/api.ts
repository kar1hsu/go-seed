import { api, APIError } from '../../../lib/http'
import { unwrap } from '../api'
import type { components } from '../../../api/generated/schema'
type Schemas = components['schemas']
export type FileRecord = Schemas['FileRecord']
export type TaskRecord = Schemas['TaskRecord']
export type Query = {
  page: number
  pageSize: number
  keyword: string
  status: string
}
const header = { 'X-GoSeed-CSRF': '1' as const }
export const files = {
  list: (query: Query) =>
    unwrap(api.GET('/admin-api/v1/system/files', { params: { query } })),
  create: (body: Schemas['CreateFileRequest']) =>
    unwrap(
      api.POST('/admin-api/v1/system/files', { params: { header }, body }),
    ),
  upload: (id: string, file: File) =>
    unwrap(
      api.PUT('/admin-api/v1/system/files/{id}/content', {
        params: { path: { id }, header },
        headers: { 'Content-Type': 'application/octet-stream' },
        body: '',
        bodySerializer: () => file,
      }),
    ),
  remove: (id: string) =>
    unwrap(
      api.DELETE('/admin-api/v1/system/files/{id}', {
        params: { path: { id }, header },
      }),
    ),
  async download(id: string, name: string) {
    const { data, error, response } = await api.GET(
      '/admin-api/v1/system/files/{id}/content',
      { params: { path: { id } }, parseAs: 'blob' },
    )
    if (!response.ok || error || !data)
      throw new APIError(
        error?.message || '下载失败',
        error?.code,
        error?.traceId,
      )
    const url = URL.createObjectURL(data),
      link = document.createElement('a')
    link.href = url
    link.download = name
    link.click()
    setTimeout(() => URL.revokeObjectURL(url), 60_000)
  },
}
export const tasks = {
  list: (query: Query) =>
    unwrap(api.GET('/admin-api/v1/system/tasks', { params: { query } })),
  export: (requestKey: string) =>
    unwrap(
      api.POST('/admin-api/v1/system/tasks/dictionary-export', {
        params: { header },
        body: { requestKey },
      }),
    ),
  retry: (id: string) =>
    unwrap(
      api.POST('/admin-api/v1/system/tasks/{id}/retry', {
        params: { path: { id }, header },
      }),
    ),
}
export function sizeLabel(size: number) {
  return size < 1024
    ? `${size} B`
    : size < 1024 * 1024
      ? `${(size / 1024).toFixed(1)} KB`
      : `${(size / 1024 / 1024).toFixed(1)} MB`
}
export function timeLabel(value: string) {
  return new Date(value).toLocaleString('zh-CN', { hour12: false })
}
export const fileStates = [
  { value: 'pending', label: '待完成' },
  { value: 'ready', label: '可用' },
  { value: 'deleting', label: '删除中' },
]
export const taskStates = [
  { value: 'pending', label: '等待执行' },
  { value: 'running', label: '执行中' },
  { value: 'succeeded', label: '已完成' },
  { value: 'failed', label: '失败' },
]
export function stateLabel(value: string, kind: 'file' | 'task') {
  return (
    (kind === 'file' ? fileStates : taskStates).find((v) => v.value === value)
      ?.label || value
  )
}
export const taskNames: Record<string, string> = {
  'system.dictionary_export.v1': '字典 CSV 导出',
  'system.file_delete.v1': '删除文件',
  'system.upload_cleanup.v1': '清理未完成上传',
}
export function fileType(file: File): string {
  const ext = file.name.split('.').at(-1)?.toLowerCase() || ''
  return (
    (
      {
        png: 'image/png',
        jpg: 'image/jpeg',
        jpeg: 'image/jpeg',
        pdf: 'application/pdf',
        txt: 'text/plain',
        csv: 'text/csv',
        json: 'application/json',
        zip: 'application/zip',
      } as Record<string, string>
    )[ext] || ''
  )
}
