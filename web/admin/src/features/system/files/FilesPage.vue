<script setup lang="ts">
import { ref } from 'vue'
import {
  ElButton,
  ElTable,
  ElTableColumn,
  ElPagination,
  ElTag,
  ElMessage,
  ElMessageBox,
} from 'element-plus'
import { Upload, RefreshRight } from '@element-plus/icons-vue'
import {
  files,
  fileStates,
  fileType,
  stateLabel,
  sizeLabel,
  timeLabel,
  type FileRecord,
} from './api'
import { useResourceList } from './use-resource-list'
import ResourceFilters from './ResourceFilters.vue'
import FilePreview from './FilePreview.vue'
import { authMessage, useAuthStore } from '../../auth/store'
const auth = useAuthStore()
const previewRecord = ref<FileRecord | null>(null)
const { query, filters, items, total, loading, error, load, search, reset } =
  useResourceList(files.list, true)
const input = ref<HTMLInputElement>(),
  uploading = ref(false),
  transferError = ref(''),
  pending = ref<{ file: File; id: string } | null>(null),
  downloading = ref('')
async function selected(event: Event) {
  const file = (event.target as HTMLInputElement).files?.[0]
  if (input.value) input.value.value = ''
  if (!file) return
  if (file.size === 0 || file.size > 10 * 1024 * 1024 || !fileType(file)) {
    ElMessage.error(
      '请选择 10 MB 以内的 PNG、JPG、PDF、TXT、CSV、JSON 或 ZIP 文件',
    )
    return
  }
  pending.value = { file, id: '' }
  await upload()
}
async function upload() {
  if (uploading.value || !pending.value) return
  const current = pending.value
  uploading.value = true
  transferError.value = ''
  try {
    if (!current.id)
      current.id = (
        await files.create({
          name: current.file.name,
          size: current.file.size,
          contentType: fileType(current.file),
        })
      ).record.id
    await files.upload(current.id, current.file)
    pending.value = null
    ElMessage.success('文件已上传')
  } catch (cause) {
    transferError.value = authMessage(cause)
  } finally {
    uploading.value = false
    await load()
  }
}
async function download(row: FileRecord) {
  downloading.value = row.id
  try {
    await files.download(row.id, row.name)
  } catch (cause) {
    ElMessage.error(authMessage(cause))
  } finally {
    downloading.value = ''
  }
}
async function remove(row: FileRecord) {
  try {
    await ElMessageBox.confirm(
      `确定删除“${row.name}”？删除后无法下载。`,
      '删除文件',
      { type: 'warning', confirmButtonText: '删除', cancelButtonText: '取消' },
    )
  } catch {
    return
  }
  try {
    await files.remove(row.id)
    ElMessage.success('已提交删除任务')
    await load()
  } catch (cause) {
    ElMessage.error(authMessage(cause))
  }
}
</script>
<template>
  <div class="page-heading">
    <div>
      <h1>文件管理</h1>
      <p>
        仅展示自己的私有文件。支持常用图片、文档与压缩包，单个文件不超过 10 MB。
      </p>
    </div>
  </div>
  <ResourceFilters
    :filters="filters"
    :states="fileStates"
    placeholder="文件名称"
    :loading="loading"
    @search="search"
    @reset="reset"
  />
  <div v-if="error" class="error-notice" role="alert">{{ error }}</div>
  <div v-if="transferError" class="error-notice" role="alert">
    {{ transferError
    }}<ElButton
      v-if="pending && auth.has('system.file.upload')"
      link
      type="primary"
      :loading="uploading"
      @click="upload"
      >重试上传</ElButton
    >
  </div>
  <section class="panel">
    <div class="panel-heading">
      <h2>
        文件列表 <span class="count-label">{{ total }}</span>
      </h2>
      <div class="table-tools">
        <input
          ref="input"
          class="visually-hidden"
          type="file"
          accept=".png,.jpg,.jpeg,.pdf,.txt,.csv,.json,.zip"
          aria-label="选择上传文件"
          @change="selected"
        />
        <ElButton
          v-if="auth.has('system.file.upload')"
          type="primary"
          :icon="Upload"
          :loading="uploading"
          @click="input?.click()"
          >{{ uploading ? '上传中…' : '上传文件' }}</ElButton
        ><ElButton
          :icon="RefreshRight"
          :loading="loading"
          aria-label="刷新文件列表"
          @click="load"
        />
      </div>
    </div>
    <div class="panel-table management-table" :aria-busy="loading">
      <ElTable
        :data="items"
        row-key="id"
        :empty-text="
          loading ? '加载中…' : '暂无文件，上传文件或完成导出后将在此显示'
        "
      >
        <ElTableColumn
          prop="name"
          label="文件名称"
          min-width="250"
          show-overflow-tooltip
        />
        <ElTableColumn label="大小" width="100"
          ><template #default="{ row }">{{
            sizeLabel(row.size)
          }}</template></ElTableColumn
        >
        <ElTableColumn label="来源" width="90"
          ><template #default="{ row }">{{
            row.source === 'export' ? '任务导出' : '手动上传'
          }}</template></ElTableColumn
        >
        <ElTableColumn label="状态" width="105"
          ><template #default="{ row }"
            ><ElTag
              :type="
                row.status === 'ready'
                  ? 'success'
                  : row.status === 'deleting'
                    ? 'warning'
                    : 'info'
              "
              effect="plain"
              size="small"
              >{{ stateLabel(row.status, 'file') }}</ElTag
            ></template
          ></ElTableColumn
        >
        <ElTableColumn label="创建时间" min-width="175"
          ><template #default="{ row }">{{
            timeLabel(row.createdAt)
          }}</template></ElTableColumn
        >
        <ElTableColumn label="操作" fixed="right" width="190"
          ><template #default="{ row }"
            ><ElButton
              link
              type="primary"
              :disabled="
                row.status !== 'ready' ||
                (row.source === 'export' && !auth.has('system.task.export'))
              "
              @click="previewRecord = row as FileRecord"
              >预览</ElButton
            ><ElButton
              link
              type="primary"
              :loading="downloading === row.id"
              :disabled="
                row.status !== 'ready' ||
                (row.source === 'export' && !auth.has('system.task.export'))
              "
              @click="download(row as FileRecord)"
              >下载</ElButton
            ><ElButton
              v-if="auth.has('system.file.delete')"
              link
              type="danger"
              :disabled="row.status !== 'ready'"
              @click="remove(row as FileRecord)"
              >删除</ElButton
            ></template
          ></ElTableColumn
        >
      </ElTable>
    </div>
    <div class="table-pagination">
      <ElPagination
        v-model:current-page="query.page"
        v-model:page-size="query.pageSize"
        :page-sizes="[10, 20, 50, 100]"
        :total="total"
        layout="total, sizes, prev, pager, next"
        @current-change="load"
        @size-change="search"
      />
    </div>
  </section>
  <p class="list-footnote">
    未完成的上传会在 24
    小时后由定时任务清理。下载链接需登录授权，不提供公开访问地址。
  </p>
  <FilePreview v-model="previewRecord" />
</template>
