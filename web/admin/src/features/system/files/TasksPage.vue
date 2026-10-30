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
  ElDrawer,
  ElDescriptions,
  ElDescriptionsItem,
} from 'element-plus'
import { Download, RefreshRight } from '@element-plus/icons-vue'
import {
  tasks,
  files,
  taskStates,
  stateLabel,
  taskNames,
  timeLabel,
  type TaskRecord,
} from './api'
import { useResourceList } from './use-resource-list'
import ResourceFilters from './ResourceFilters.vue'
import { authMessage, useAuthStore } from '../../auth/store'
const auth = useAuthStore()
const { query, filters, items, total, loading, error, load, search, reset } =
  useResourceList(tasks.list, true)
const creating = ref(false),
  requestKey = ref(''),
  selected = ref<TaskRecord | null>(null),
  drawer = ref(false),
  retrying = ref('')
async function create() {
  if (creating.value) return
  creating.value = true
  if (!requestKey.value) requestKey.value = crypto.randomUUID()
  try {
    await tasks.export(requestKey.value)
    requestKey.value = ''
    ElMessage.success('导出任务已提交')
    await load()
  } catch (cause) {
    ElMessage.error(authMessage(cause))
  } finally {
    creating.value = false
  }
}
async function retry(row: TaskRecord) {
  try {
    await ElMessageBox.confirm(
      '确认重新执行此任务？系统会重新检查当前权限。',
      '重试任务',
      { confirmButtonText: '重试', cancelButtonText: '取消' },
    )
  } catch {
    return
  }
  retrying.value = row.id
  try {
    await tasks.retry(row.id)
    ElMessage.success('任务已重新排队')
    await load()
  } catch (cause) {
    ElMessage.error(authMessage(cause))
  } finally {
    retrying.value = ''
  }
}
async function download(row: TaskRecord) {
  try {
    await files.download(row.fileId, `dictionaries-${row.id}.csv`)
  } catch (cause) {
    ElMessage.error(authMessage(cause))
  }
}
function details(row: TaskRecord) {
  selected.value = row
  drawer.value = true
}
</script>
<template>
  <div class="page-heading">
    <div>
      <h1>任务中心</h1>
      <p>
        查看异步任务进度与执行结果。列表每 5
        秒更新，失败任务可在修复原因后重试。
      </p>
    </div>
  </div>
  <ResourceFilters
    :filters="filters"
    :states="taskStates"
    placeholder="任务类型 / Trace ID"
    :loading="loading"
    @search="search"
    @reset="reset"
  />
  <div v-if="error" class="error-notice" role="alert">{{ error }}</div>
  <section class="panel">
    <div class="panel-heading">
      <h2>
        任务列表 <span class="count-label">{{ total }}</span>
      </h2>
      <div class="table-tools">
        <ElButton
          v-if="auth.has('system.task.export')"
          type="primary"
          :icon="Download"
          :loading="creating"
          @click="create"
          >导出字典 CSV</ElButton
        ><ElButton
          :icon="RefreshRight"
          :loading="loading"
          aria-label="刷新任务列表"
          @click="load"
        />
      </div>
    </div>
    <div class="panel-table management-table" :aria-busy="loading">
      <ElTable
        :data="items"
        row-key="id"
        :empty-text="loading ? '加载中…' : '暂无任务'"
      >
        <ElTableColumn label="任务" min-width="220"
          ><template #default="{ row }"
            ><div class="role-cell">
              <strong>{{ taskNames[row.kind] || row.kind }}</strong
              ><small>#{{ row.id }}</small>
            </div></template
          ></ElTableColumn
        >
        <ElTableColumn label="状态" width="105"
          ><template #default="{ row }"
            ><ElTag
              :type="
                row.status === 'succeeded'
                  ? 'success'
                  : row.status === 'failed'
                    ? 'danger'
                    : row.status === 'running'
                      ? 'primary'
                      : 'info'
              "
              effect="plain"
              size="small"
              >{{ stateLabel(row.status, 'task') }}</ElTag
            ></template
          ></ElTableColumn
        >
        <ElTableColumn label="执行次数" width="100"
          ><template #default="{ row }"
            >{{ row.attempts }} / {{ row.maxAttempts }}</template
          ></ElTableColumn
        >
        <ElTableColumn
          prop="error"
          label="最近结果"
          min-width="240"
          show-overflow-tooltip
          ><template #default="{ row }">{{
            row.error || (row.status === 'succeeded' ? '执行成功' : '—')
          }}</template></ElTableColumn
        >
        <ElTableColumn label="提交时间" min-width="175"
          ><template #default="{ row }">{{
            timeLabel(row.createdAt)
          }}</template></ElTableColumn
        >
        <ElTableColumn label="操作" fixed="right" width="180"
          ><template #default="{ row }"
            ><ElButton link type="primary" @click="details(row as TaskRecord)"
              >详情</ElButton
            ><ElButton
              v-if="
                row.kind === 'system.dictionary_export.v1' &&
                auth.has('system.task.export')
              "
              link
              type="primary"
              :disabled="row.status !== 'succeeded'"
              @click="download(row as TaskRecord)"
              >下载</ElButton
            ><ElButton
              v-if="row.retryable && auth.has('system.task.retry')"
              link
              type="primary"
              :loading="retrying === row.id"
              @click="retry(row as TaskRecord)"
              >重试</ElButton
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
    内置定时任务：每小时清理超过 24
    小时仍未完成的上传；服务恢复后执行当前小时的清理。
  </p>
  <ElDrawer v-model="drawer" title="任务详情" size="min(580px, 100vw)"
    ><ElDescriptions v-if="selected" :column="1" border>
      <ElDescriptionsItem label="任务 ID">{{ selected.id }}</ElDescriptionsItem
      ><ElDescriptionsItem label="类型">{{
        taskNames[selected.kind] || selected.kind
      }}</ElDescriptionsItem
      ><ElDescriptionsItem label="状态">{{
        stateLabel(selected.status, 'task')
      }}</ElDescriptionsItem
      ><ElDescriptionsItem label="执行次数"
        >{{ selected.attempts }} /
        {{ selected.maxAttempts }}</ElDescriptionsItem
      ><ElDescriptionsItem label="提交时间">{{
        timeLabel(selected.createdAt)
      }}</ElDescriptionsItem
      ><ElDescriptionsItem label="更新时间">{{
        timeLabel(selected.updatedAt)
      }}</ElDescriptionsItem
      ><ElDescriptionsItem
        v-if="selected.status === 'pending'"
        label="最早执行时间"
        >{{ timeLabel(selected.nextRunAt) }}</ElDescriptionsItem
      ><ElDescriptionsItem label="Trace ID"
        ><span class="break-word">{{
          selected.traceId || '系统定时任务'
        }}</span></ElDescriptionsItem
      ><ElDescriptionsItem label="最近结果">{{
        selected.error || '—'
      }}</ElDescriptionsItem>
    </ElDescriptions></ElDrawer
  >
</template>
