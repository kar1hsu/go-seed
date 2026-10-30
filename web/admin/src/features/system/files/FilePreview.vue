<script setup lang="ts">
import { defineAsyncComponent, onBeforeUnmount, ref, watch } from 'vue'
import {
  ElButton,
  ElDrawer,
  ElEmpty,
  ElImage,
  ElSkeleton,
  ElMessage,
} from 'element-plus'
import { Download } from '@element-plus/icons-vue'
import { files, sizeLabel, timeLabel, type FileRecord } from './api'
import { createFilePreview } from './preview'
import { authMessage, useAuthStore } from '../../auth/store'

const record = defineModel<FileRecord | null>({ required: true })
const auth = useAuthStore()
const preview = createFilePreview(files.content)
const { kind, loading, error, url, pdfData, text, truncated } = preview
const downloading = ref(false)
const PdfPreview = defineAsyncComponent(() => import('./PdfPreview.vue'))

watch(
  () =>
    [
      record.value,
      auth.admin?.id,
      auth.has('system.file.read'),
      auth.has('system.task.export'),
    ] as const,
  (currentState, previousState) => {
    const current = record.value
    if (
      current &&
      ((previousState && currentState[1] !== previousState[1]) ||
        !auth.has('system.file.read') ||
        (current.source === 'export' && !auth.has('system.task.export')))
    ) {
      record.value = null
      preview.close()
      return
    }
    if (current) {
      if (preview.selected.value?.id !== current.id) void preview.open(current)
    } else preview.close()
  },
  { immediate: true, flush: 'sync' },
)
onBeforeUnmount(preview.close)

async function download() {
  if (!record.value || downloading.value) return
  downloading.value = true
  try {
    await files.download(record.value.id, record.value.name)
  } catch (cause) {
    ElMessage.error(authMessage(cause))
  } finally {
    downloading.value = false
  }
}
</script>

<template>
  <ElDrawer
    :model-value="!!record"
    title="文件预览"
    size="min(960px, 100vw)"
    destroy-on-close
    @update:model-value="
      (visible) => {
        if (!visible) record = null
      }
    "
  >
    <div v-if="record" class="file-preview">
      <div class="preview-meta">
        <h3>{{ record.name }}</h3>
        <p>{{ sizeLabel(record.size) }} · {{ timeLabel(record.createdAt) }}</p>
      </div>
      <div v-if="loading" class="preview-loading" role="status">
        <p>正在加载预览…</p>
        <ElSkeleton :rows="6" animated />
      </div>
      <div v-else-if="error" class="preview-error" role="alert">
        <p>{{ error }}</p>
        <ElButton @click="preview.open(record)">重新加载</ElButton>
      </div>
      <template v-else-if="kind === 'image' && url">
        <div class="preview-image">
          <ElImage
            :src="url"
            :alt="record.name"
            fit="contain"
            :preview-src-list="[url]"
            preview-teleported
            hide-on-click-modal
            @error="error = '图片无法解码，请下载原文件查看'"
          />
        </div>
        <p class="preview-hint">点击图片可放大、缩小或旋转。</p>
      </template>
      <PdfPreview
        v-else-if="kind === 'pdf' && pdfData"
        :key="record.id"
        :data="pdfData"
      />
      <template v-else-if="kind === 'text'">
        <p v-if="truncated" class="preview-hint" role="status">
          仅预览前 256 KB，完整内容请下载原文件。
        </p>
        <pre class="preview-text" tabindex="0" :aria-label="record.name">{{
          text
        }}</pre>
      </template>
      <ElEmpty v-else description="此文件类型暂不支持在线预览，请下载后查看" />
    </div>
    <template #footer>
      <ElButton @click="record = null">关闭</ElButton>
      <ElButton
        type="primary"
        :icon="Download"
        :loading="downloading"
        @click="download"
        >下载原文件</ElButton
      >
    </template>
  </ElDrawer>
</template>

<style scoped>
.file-preview {
  display: flex;
  flex-direction: column;
  gap: 16px;
  height: 100%;
  min-height: 0;
}
.preview-meta {
  padding-bottom: 16px;
  border-bottom: 1px solid var(--el-border-color-light);
}
.preview-meta h3 {
  font-size: 16px;
  font-weight: 500;
  overflow-wrap: anywhere;
}
.preview-meta p,
.preview-hint {
  margin-top: 6px;
  color: var(--el-text-color-secondary);
  font-size: 13px;
}
.preview-loading,
.preview-error {
  padding: 24px 0;
}
.preview-loading p,
.preview-error p {
  margin-bottom: 16px;
}
.preview-error {
  color: var(--el-color-danger);
}
.preview-image {
  display: flex;
  flex: 1;
  align-items: center;
  justify-content: center;
  min-height: 160px;
  overflow: hidden;
  background: var(--el-fill-color-light);
  border-radius: 6px;
}
.preview-image :deep(.el-image) {
  width: 100%;
  height: 100%;
}
.preview-text {
  flex: 1;
  margin: 0;
  padding: 20px;
  min-height: 200px;
  overflow: auto;
  white-space: pre-wrap;
  overflow-wrap: anywhere;
  font:
    13px/1.8 ui-monospace,
    SFMono-Regular,
    Menlo,
    monospace;
  background: var(--el-fill-color-light);
  border: 1px solid var(--el-border-color-light);
  border-radius: 6px;
}
</style>
