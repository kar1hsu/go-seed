import { onBeforeUnmount, onMounted, reactive, ref, shallowRef } from 'vue'
import { authMessage } from '../features/auth/store'
import type { Query } from '../lib/list-query'

export function useList<T>(
  request: (query: Query) => Promise<{ items: T[]; total: number }>,
) {
  const query = reactive<Query>({
    page: 1,
    pageSize: 20,
    keyword: '',
    enabled: 'all',
  })
  const filters = reactive({ keyword: '', enabled: 'all' as Query['enabled'] })
  const items = shallowRef<T[]>([])
  const total = ref(0)
  const loading = ref(false)
  const error = ref('')
  let sequence = 0
  let pending: Promise<void> | undefined
  let pendingKey = ''
  function load(): Promise<void> {
    const key = JSON.stringify(query)
    if (pending && pendingKey === key) return pending
    pendingKey = key
    const operation = performLoad().finally(() => {
      if (pending === operation) pending = undefined
    })
    pending = operation
    return operation
  }
  async function performLoad() {
    const attempt = ++sequence
    loading.value = true
    error.value = ''
    try {
      const data = await request({ ...query })
      if (sequence !== attempt) return
      items.value = data.items
      total.value = data.total
      if (!data.items.length && query.page > 1) {
        query.page--
        await load()
      }
    } catch (cause) {
      if (sequence === attempt) {
        items.value = []
        error.value = authMessage(cause)
      }
    } finally {
      if (sequence === attempt) loading.value = false
    }
  }
  function search() {
    Object.assign(query, filters, { keyword: filters.keyword.trim(), page: 1 })
    void load()
  }
  function reset() {
    filters.keyword = ''
    filters.enabled = 'all'
    search()
  }
  onMounted(load)
  onBeforeUnmount(() => {
    sequence++
  })
  return { query, filters, items, total, loading, error, load, search, reset }
}
