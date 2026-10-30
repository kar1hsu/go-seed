import { onBeforeUnmount, onMounted, reactive, ref, shallowRef } from 'vue'
import { authMessage } from '../../auth/store'
import type { Query } from './api'
export function useResourceList<T>(
  request: (query: Query) => Promise<{ items: T[]; total: number }>,
  poll = false,
) {
  const query = reactive<Query>({
    page: 1,
    pageSize: 20,
    keyword: '',
    status: '',
  })
  const filters = reactive({ keyword: '', status: '' }),
    items = shallowRef<T[]>([]),
    total = ref(0),
    loading = ref(false),
    error = ref('')
  let sequence = 0,
    timer: ReturnType<typeof setInterval> | undefined
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
      const result = await request({ ...query })
      if (attempt !== sequence) return
      items.value = result.items
      total.value = result.total
      if (!items.value.length && query.page > 1) {
        query.page--
        await load()
      }
    } catch (cause) {
      if (attempt === sequence) {
        error.value = authMessage(cause)
        items.value = []
      }
    } finally {
      if (attempt === sequence) loading.value = false
    }
  }
  function search() {
    Object.assign(query, filters, { page: 1, keyword: filters.keyword.trim() })
    void load()
  }
  function reset() {
    Object.assign(filters, { keyword: '', status: '' })
    search()
  }
  onMounted(() => {
    void load()
    if (poll)
      timer = setInterval(() => {
        if (!loading.value && document.visibilityState === 'visible')
          void load()
      }, 5000)
  })
  onBeforeUnmount(() => {
    sequence++
    clearInterval(timer)
  })
  return { query, filters, items, total, loading, error, load, search, reset }
}
