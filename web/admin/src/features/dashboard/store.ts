import { defineStore } from 'pinia'
import { ref } from 'vue'
import { probeSystem } from './api'
import { APIError } from '../../lib/http'
import type { components } from '../../api/generated/schema'

type PingResponse = components['schemas']['PingResponse']
export interface ProbeRecord {
  id: number
  time: string
  success: boolean
  duration: number
  traceId: string
  code: number | null
  message: string
  response: PingResponse | null
}

export const useDashboardStore = defineStore('dashboard', () => {
  const loading = ref(false)
  const result = ref<PingResponse | null>(null)
  const error = ref('')
  const traceId = ref('')
  const checkedAt = ref('')
  const latency = ref(0)
  const records = ref<ProbeRecord[]>([])
  let nextId = 1
  async function refresh() {
    if (loading.value) return
    loading.value = true
    error.value = ''
    result.value = null
    traceId.value = ''
    let code: number | null = null
    const start = performance.now()
    try {
      result.value = await probeSystem()
      traceId.value = result.value.traceId
      code = 0
    } catch (cause) {
      const messages: Record<number, string> = {
        100503: '服务暂时不可用',
        100504: '服务响应超时',
        100500: '服务内部错误',
      }
      code = cause instanceof APIError ? (cause.code ?? null) : null
      error.value =
        cause instanceof APIError
          ? messages[code ?? 0] || cause.message
          : '连接失败，请检查服务是否已启动'
      traceId.value = cause instanceof APIError ? cause.traceId || '' : ''
    } finally {
      latency.value = Math.round(performance.now() - start)
      checkedAt.value = new Date().toISOString()
      records.value.unshift({
        id: nextId++,
        time: checkedAt.value,
        success: !!result.value,
        duration: latency.value,
        traceId: traceId.value,
        code,
        message: error.value || '请求成功',
        response: result.value,
      })
      records.value = records.value.slice(0, 20)
      loading.value = false
    }
  }
  function clearRecords() {
    records.value = []
  }
  return {
    loading,
    result,
    error,
    traceId,
    checkedAt,
    latency,
    records,
    refresh,
    clearRecords,
  }
})
