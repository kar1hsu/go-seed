<script setup lang="ts">
import { onBeforeUnmount, onMounted, ref } from 'vue'
import { ElButton } from 'element-plus'
import {
  getDocument,
  GlobalWorkerOptions,
  type PDFDocumentLoadingTask,
  type PDFDocumentProxy,
  type RenderTask,
} from 'pdfjs-dist'
import workerURL from 'pdfjs-dist/build/pdf.worker.min.mjs?worker&url'

const props = defineProps<{ data: Uint8Array }>()
GlobalWorkerOptions.workerSrc = workerURL
const canvas = ref<HTMLCanvasElement>()
const surface = ref<HTMLElement>()
const page = ref(1),
  pages = ref(0),
  zoom = ref(1)
const busy = ref(true),
  error = ref('')
let document: PDFDocumentProxy | undefined
let loadingTask: PDFDocumentLoadingTask | undefined
let renderTask: RenderTask | undefined
let disposed = false
let timeout: ReturnType<typeof setTimeout> | undefined

async function render() {
  if (!document || !canvas.value || !surface.value || disposed) return
  busy.value = true
  try {
    const current = await document.getPage(page.value)
    if (disposed) return
    const base = current.getViewport({ scale: 1 })
    const fit = Math.min(1.5, (surface.value.clientWidth - 32) / base.width)
    const viewport = current.getViewport({ scale: fit * zoom.value })
    const ratio = Math.min(
      window.devicePixelRatio || 1,
      2,
      4096 / Math.max(viewport.width, viewport.height),
    )
    const target = canvas.value
    target.width = Math.ceil(viewport.width * ratio)
    target.height = Math.ceil(viewport.height * ratio)
    target.style.width = `${viewport.width}px`
    target.style.height = `${viewport.height}px`
    renderTask = current.render({
      canvas: target,
      viewport,
      transform: [ratio, 0, 0, ratio, 0, 0],
    })
    await renderTask.promise
  } catch {
    if (!disposed) error.value = 'PDF 无法显示，请下载原文件查看'
  } finally {
    if (!disposed) busy.value = false
  }
}
async function changePage(delta: number) {
  if (busy.value) return
  page.value += delta
  await render()
}
async function scale(delta: number) {
  if (busy.value) return
  zoom.value = Math.round((zoom.value + delta) * 100) / 100
  await render()
}
onMounted(async () => {
  // Assets are served by this application. Canvas rendering does not activate
  // document scripts, forms or external links; WebAssembly is unnecessary.
  const assets = `${import.meta.env.BASE_URL}pdfjs/`
  loadingTask = getDocument({
    data: props.data.slice(),
    cMapUrl: `${assets}cmaps/`,
    standardFontDataUrl: `${assets}standard_fonts/`,
    wasmUrl: `${assets}wasm/`,
    iccUrl: `${assets}iccs/`,
    useWasm: false,
    enableXfa: false,
    stopAtErrors: true,
    maxImageSize: 16 * 1024 * 1024,
    canvasMaxAreaInBytes: 32 * 1024 * 1024,
  })
  loadingTask.onPassword = () => {
    error.value = '此 PDF 需要密码，请下载后使用阅读器打开'
    busy.value = false
    void loadingTask?.destroy()
  }
  timeout = setTimeout(() => {
    error.value = 'PDF 预览超时，请下载原文件查看'
    busy.value = false
    renderTask?.cancel()
    void loadingTask?.destroy()
  }, 30_000)
  try {
    document = await loadingTask.promise
    if (disposed) return
    pages.value = document.numPages
    await render()
  } catch {
    if (!disposed && !error.value)
      error.value = 'PDF 无法读取，请下载原文件查看'
  } finally {
    clearTimeout(timeout)
    if (!disposed) busy.value = false
  }
})
onBeforeUnmount(() => {
  disposed = true
  clearTimeout(timeout)
  renderTask?.cancel()
  void loadingTask?.destroy()
})
</script>

<template>
  <div class="pdf-preview" :aria-busy="busy">
    <div v-if="error" class="pdf-error" role="alert">{{ error }}</div>
    <template v-else>
      <div class="pdf-toolbar">
        <ElButton
          size="small"
          :disabled="busy || page <= 1"
          @click="changePage(-1)"
          >上一页</ElButton
        >
        <span>{{ page }} / {{ pages || '—' }}</span>
        <ElButton
          size="small"
          :disabled="busy || page >= pages"
          @click="changePage(1)"
          >下一页</ElButton
        >
        <span class="pdf-divider" />
        <ElButton
          size="small"
          :disabled="busy || zoom <= 0.5"
          @click="scale(-0.25)"
          >缩小</ElButton
        >
        <span>{{ Math.round(zoom * 100) }}%</span>
        <ElButton
          size="small"
          :disabled="busy || zoom >= 2"
          @click="scale(0.25)"
          >放大</ElButton
        >
      </div>
      <p v-if="busy" class="pdf-loading" role="status">正在渲染 PDF…</p>
      <div ref="surface" class="pdf-surface">
        <canvas ref="canvas" :aria-label="`PDF 第 ${page} 页`"
          >PDF 预览，请下载原文件以阅读完整内容。</canvas
        >
      </div>
    </template>
  </div>
</template>

<style scoped>
.pdf-preview {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 12px;
  min-height: 260px;
}
.pdf-toolbar {
  display: flex;
  align-items: center;
  justify-content: center;
  flex-wrap: wrap;
  gap: 10px;
  font-size: 13px;
}
.pdf-toolbar .el-button + .el-button {
  margin-left: 0;
}
.pdf-divider {
  width: 1px;
  height: 16px;
  margin: 0 4px;
  background: var(--el-border-color);
}
.pdf-surface {
  flex: 1;
  min-height: 0;
  padding: 16px;
  overflow: auto;
  scrollbar-gutter: stable;
  border: 1px solid var(--el-border-color-light);
  border-radius: 6px;
  background: var(--el-fill-color-light);
}
.pdf-surface canvas {
  display: block;
  margin: 0 auto;
  background: white;
  box-shadow: 0 1px 6px #00000014;
}
.pdf-loading {
  color: var(--el-text-color-secondary);
  font-size: 13px;
  text-align: center;
}
.pdf-error {
  padding: 24px 0;
  color: var(--el-color-danger);
}
</style>
