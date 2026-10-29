export function formatTime(value: string) {
  return value
    ? new Date(value).toLocaleTimeString('zh-CN', { hour12: false })
    : '—'
}
export function formatDate(value: string) {
  return value
    ? new Date(value).toLocaleString('zh-CN', { hour12: false })
    : '—'
}
