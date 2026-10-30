import { reactive, ref, shallowRef } from 'vue'
import { authMessage } from '../../auth/store'
import {
  auditAPI,
  type AuditLog,
  type AuditDetail,
  type AuditAction,
  type AuditQuery,
} from './api'

// Responses from an older query, closed drawer or disposed page must never restore stale data.
export function createAuditState(request = auditAPI) {
  const query = reactive<AuditQuery>({
    page: 1,
    pageSize: 20,
    keyword: '',
    action: '',
    actorId: '',
    traceId: '',
    from: '',
    to: '',
  })
  const items = shallowRef<AuditLog[]>([])
  const actions = shallowRef<AuditAction[]>([])
  const total = ref(0)
  const loading = ref(false)
  const error = ref('')
  const detail = shallowRef<AuditDetail | null>(null)
  const detailLoading = ref(false)
  const detailError = ref('')
  const selectedId = ref('')
  let listSequence = 0
  let detailSequence = 0
  async function load() {
    const attempt = ++listSequence
    loading.value = true
    error.value = ''
    try {
      const data = await request.list({ ...query })
      if (attempt !== listSequence) return
      items.value = data.items
      actions.value = data.actions
      total.value = data.total
    } catch (cause) {
      if (attempt !== listSequence) return
      items.value = []
      total.value = 0
      error.value = authMessage(cause)
    } finally {
      if (attempt === listSequence) loading.value = false
    }
  }
  async function open(id: string) {
    const attempt = ++detailSequence
    selectedId.value = id
    detail.value = null
    detailError.value = ''
    detailLoading.value = true
    try {
      const data = await request.detail(id)
      if (attempt === detailSequence) detail.value = data
    } catch (cause) {
      if (attempt === detailSequence) detailError.value = authMessage(cause)
    } finally {
      if (attempt === detailSequence) detailLoading.value = false
    }
  }
  function close() {
    detailSequence++
    selectedId.value = ''
    detail.value = null
    detailLoading.value = false
    detailError.value = ''
  }
  function clear() {
    listSequence++
    close()
    items.value = []
    total.value = 0
    loading.value = false
    error.value = ''
  }
  return {
    query,
    items,
    actions,
    total,
    loading,
    error,
    detail,
    detailLoading,
    detailError,
    selectedId,
    load,
    open,
    close,
    clear,
  }
}
