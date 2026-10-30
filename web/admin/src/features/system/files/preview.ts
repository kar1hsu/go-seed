import { ref, shallowRef } from 'vue'
import { authMessage } from '../../auth/store'
import type { FileRecord } from './api'

export const TEXT_PREVIEW_LIMIT = 256 * 1024
export function previewKind(contentType: string) {
  switch (contentType) {
    case 'image/png':
    case 'image/jpeg':
      return 'image'
    case 'application/pdf':
      return 'pdf'
    case 'text/plain':
    case 'text/csv':
    case 'application/json':
      return 'text'
    default:
      return 'unsupported'
  }
}

// The caller owns the lifecycle: close on dismissal, permission loss and unmount.
export function createFilePreview(
  content: (id: string, signal: AbortSignal) => Promise<Blob>,
) {
  const selected = shallowRef<FileRecord | null>(null)
  const kind = ref<ReturnType<typeof previewKind>>('unsupported')
  const loading = ref(false)
  const error = ref('')
  const url = ref('')
  const pdfData = shallowRef<Uint8Array | null>(null)
  const text = ref('')
  const truncated = ref(false)
  let request: AbortController | undefined
  let sequence = 0

  function close() {
    sequence++
    request?.abort()
    request = undefined
    if (url.value) URL.revokeObjectURL(url.value)
    url.value = ''
    pdfData.value = null
    text.value = ''
    error.value = ''
    truncated.value = false
    loading.value = false
    selected.value = null
    kind.value = 'unsupported'
  }

  async function open(record: FileRecord) {
    close()
    const attempt = sequence
    selected.value = { ...record }
    kind.value = previewKind(record.contentType)
    if (record.status !== 'ready') {
      error.value = '文件尚不可用，请刷新列表后重试'
      return
    }
    if (kind.value === 'unsupported') return
    const controller = new AbortController()
    request = controller
    loading.value = true
    try {
      const blob = await content(record.id, controller.signal)
      if (attempt !== sequence) return
      if (blob.size !== record.size || blob.size > 10 * 1024 * 1024) {
        error.value = '文件内容不完整，请重试'
        return
      }
      if (kind.value === 'text') {
        const bytes = await blob.slice(0, TEXT_PREVIEW_LIMIT).arrayBuffer()
        if (attempt !== sequence) return
        truncated.value = blob.size > TEXT_PREVIEW_LIMIT
        // Do not display a broken UTF-8 character at the truncated boundary.
        text.value = new TextDecoder().decode(bytes, {
          stream: truncated.value,
        })
      } else if (kind.value === 'pdf') {
        const bytes = await blob.arrayBuffer()
        if (attempt !== sequence) return
        pdfData.value = new Uint8Array(bytes)
      } else {
        // The download endpoint deliberately returns octet-stream. Only verified,
        // allowlisted metadata can supply a type for a local preview URL.
        url.value = URL.createObjectURL(
          blob.slice(0, blob.size, record.contentType),
        )
      }
    } catch (cause) {
      if (attempt === sequence) error.value = authMessage(cause)
    } finally {
      if (attempt === sequence) {
        loading.value = false
        request = undefined
      }
    }
  }
  return {
    selected,
    kind,
    loading,
    error,
    url,
    pdfData,
    text,
    truncated,
    open,
    close,
  }
}
