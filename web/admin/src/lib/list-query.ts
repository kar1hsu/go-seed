export type Query = {
  page: number
  pageSize: number
  keyword: string
  enabled: 'all' | 'enabled' | 'disabled'
}
