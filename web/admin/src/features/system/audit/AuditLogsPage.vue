<script setup lang="ts">
import { onBeforeUnmount, onMounted, reactive, ref, watch } from 'vue'
import {
  ElButton,
  ElDatePicker,
  ElDescriptions,
  ElDescriptionsItem,
  ElDrawer,
  ElEmpty,
  ElInput,
  ElOption,
  ElPagination,
  ElSelect,
  ElTable,
  ElTableColumn,
  ElTag,
} from 'element-plus'
import { RefreshRight, Search } from '@element-plus/icons-vue'
import { useAuthStore } from '../../auth/store'
import { createAuditState } from './state'
import type { AuditLog } from './api'

const auth = useAuthStore()
const {
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
} = createAuditState()
const filters = reactive({ keyword: '', action: '', traceId: '', actorId: '' })
const dates = ref<[Date, Date] | null>(null)
const actorLabel = ref('')
function search() {
  const end = dates.value ? new Date(dates.value[1]) : null
  end?.setDate(end.getDate() + 1)
  Object.assign(query, filters, {
    keyword: filters.keyword.trim(),
    traceId: filters.traceId.trim(),
    page: 1,
    from: dates.value?.[0].toISOString() || '',
    to: end?.toISOString() || '',
  })
  void load()
}
function reset() {
  Object.assign(filters, { keyword: '', action: '', traceId: '', actorId: '' })
  dates.value = null
  actorLabel.value = ''
  search()
}
function byActor(row: AuditLog) {
  filters.actorId = row.actorId
  actorLabel.value = row.actorName
  search()
}
function clearActor() {
  filters.actorId = ''
  actorLabel.value = ''
  search()
}
function byTrace(traceId: string) {
  filters.traceId = traceId
  close()
  search()
}
function time(value: string) {
  return new Date(value).toLocaleString('zh-CN', { hour12: false })
}
watch(
  () => auth.permissions,
  () => {
    clear()
    if (auth.has('system.audit.read')) void load()
  },
)
onMounted(load)
onBeforeUnmount(clear)
function targetName(type: string) {
  return (
    (
      {
        administrator: '管理员',
        role: '角色',
        file: '文件',
        task: '任务',
        dictionary: '字典',
        setting: '系统配置',
      } as Record<string, string>
    )[type] || type
  )
}
</script>

<template>
  <div class="page-heading">
    <div>
      <h1>操作日志</h1>
      <p>
        默认查询最近 90 天；单次日期范围最长一年，超过 10,000 条请缩小筛选范围。
      </p>
      <p>查看管理员与角色的变更记录，追溯操作人和修改内容。</p>
    </div>
  </div>
  <form class="panel audit-filters" @submit.prevent="search">
    <label
      ><span>关键词</span
      ><ElInput
        v-model="filters.keyword"
        placeholder="操作人 / 操作对象"
        aria-label="日志关键词"
        maxlength="64"
        clearable
    /></label>
    <label
      ><span>操作类型</span
      ><ElSelect
        v-model="filters.action"
        placeholder="全部操作"
        aria-label="操作类型"
        clearable
        ><ElOption
          v-for="action in actions"
          :key="action.code"
          :label="action.name"
          :value="action.code" /></ElSelect
    ></label>
    <label class="date-filter"
      ><span>操作日期</span
      ><ElDatePicker
        v-model="dates"
        type="daterange"
        start-placeholder="开始日期"
        end-placeholder="结束日期"
        range-separator="至"
        unlink-panels
    /></label>
    <label
      ><span>Trace ID</span
      ><ElInput
        v-model="filters.traceId"
        placeholder="输入完整 Trace ID"
        aria-label="Trace ID"
        maxlength="32"
        clearable
    /></label>
    <div class="filter-actions">
      <ElButton
        type="primary"
        native-type="submit"
        :icon="Search"
        :loading="loading"
        >查询</ElButton
      ><ElButton @click="reset">重置</ElButton>
    </div>
    <div v-if="actorLabel" class="actor-filter">
      <ElTag closable effect="plain" @close="clearActor"
        >操作人：{{ actorLabel }}</ElTag
      >
    </div>
  </form>
  <div v-if="error" class="error-notice" role="alert">{{ error }}</div>
  <section class="panel">
    <div class="panel-heading">
      <h2>
        日志列表 <span class="count-label">{{ total }}</span>
      </h2>
      <div class="table-tools">
        <span class="audit-hint">仅记录已提交的操作</span
        ><ElButton
          :icon="RefreshRight"
          :loading="loading"
          aria-label="刷新操作日志"
          @click="load"
        />
      </div>
    </div>
    <div class="panel-table management-table" :aria-busy="loading">
      <ElTable
        :data="items"
        row-key="id"
        :empty-text="loading ? '加载中…' : '暂无操作记录'"
      >
        <ElTableColumn label="操作时间" width="185"
          ><template #default="{ row }">{{
            time(row.createdAt)
          }}</template></ElTableColumn
        >
        <ElTableColumn label="操作人" min-width="150"
          ><template #default="{ row }"
            ><div class="audit-person">
              <ElButton link type="primary" @click="byActor(row as AuditLog)">{{
                row.actorName
              }}</ElButton
              ><small>{{ row.actorUsername }}</small>
            </div></template
          ></ElTableColumn
        >
        <ElTableColumn prop="actionName" label="操作类型" min-width="170" />
        <ElTableColumn label="操作对象" min-width="215"
          ><template #default="{ row }"
            ><div class="role-cell">
              <strong>{{ row.targetName }}</strong
              ><small
                >{{ targetName(row.targetType) }} · {{ row.targetKey }}</small
              >
            </div></template
          ></ElTableColumn
        >
        <ElTableColumn label="来源" width="110"
          ><template #default="{ row }"
            ><ElTag size="small" type="info" effect="plain">{{
              row.source === 'command' ? '运维命令' : '管理后台'
            }}</ElTag></template
          ></ElTableColumn
        >
        <ElTableColumn label="操作" width="95" fixed="right"
          ><template #default="{ row }"
            ><ElButton link type="primary" @click="open(row.id)"
              >详情</ElButton
            ></template
          ></ElTableColumn
        >
      </ElTable>
    </div>
    <div class="table-pagination">
      <span>共 {{ total }} 条</span
      ><ElPagination
        v-model:current-page="query.page"
        v-model:page-size="query.pageSize"
        :page-sizes="[10, 20, 50]"
        :total="Math.min(total, 10000)"
        layout="sizes, prev, pager, next"
        @change="load"
      />
    </div>
  </section>
  <ElDrawer
    :model-value="!!selectedId"
    title="操作详情"
    size="min(780px, 100vw)"
    destroy-on-close
    @update:model-value="close"
  >
    <p v-if="detailLoading" class="muted" role="status">正在加载操作详情…</p>
    <div v-if="detailError" class="error-notice" role="alert">
      {{ detailError }}
      <ElButton link type="primary" @click="open(selectedId)">重试</ElButton>
    </div>
    <template v-if="detail">
      <div class="audit-detail-heading">
        <h3>{{ detail.record.actionName }}</h3>
        <ElTag type="success" effect="plain">已提交</ElTag>
      </div>
      <p class="drawer-intro">
        {{ time(detail.record.createdAt) }} · 记录 #{{ detail.record.id }}
      </p>
      <ElDescriptions :column="2" border class="audit-meta">
        <ElDescriptionsItem label="操作人"
          >{{ detail.record.actorName
          }}<small
            >{{ detail.record.actorUsername }} · #{{
              detail.record.actorId
            }}</small
          ></ElDescriptionsItem
        >
        <ElDescriptionsItem label="来源">{{
          detail.record.source === 'command' ? '领域运维命令' : '管理后台'
        }}</ElDescriptionsItem>
        <ElDescriptionsItem label="操作对象" :span="2"
          >{{ detail.record.targetName
          }}<small
            >{{ detail.record.targetKey }} · #{{
              detail.record.targetId
            }}</small
          ></ElDescriptionsItem
        >
        <ElDescriptionsItem label="Trace ID" :span="2"
          ><code>{{ detail.record.traceId || '—' }}</code
          ><ElButton
            v-if="detail.record.traceId"
            link
            type="primary"
            @click="byTrace(detail.record.traceId)"
            >查看关联记录</ElButton
          ></ElDescriptionsItem
        >
      </ElDescriptions>
      <h3 class="audit-section-title">变更内容</h3>
      <ElTable
        v-if="detail.changes.length"
        :data="detail.changes"
        border
        class="audit-changes"
      >
        <ElTableColumn prop="label" label="字段" width="100" />
        <ElTableColumn label="修改前" min-width="200"
          ><template #default="{ row }"
            ><ul v-if="row.before.length" class="change-values before">
              <li v-for="(value, index) in row.before" :key="index">
                {{ value || '（空）' }}
              </li>
            </ul>
            <span v-else class="muted">—</span></template
          ></ElTableColumn
        >
        <ElTableColumn label="修改后" min-width="200"
          ><template #default="{ row }"
            ><ul v-if="row.after.length" class="change-values after">
              <li v-for="(value, index) in row.after" :key="index">
                {{ value || '（空）' }}
              </li>
            </ul>
            <span v-else class="muted">—</span></template
          ></ElTableColumn
        >
      </ElTable>
      <ElEmpty v-else description="本次提交未改变可记录字段" :image-size="70" />
      <p class="field-help">
        记录保留操作发生时的信息。密码和令牌不会写入日志。
      </p>
    </template>
    <template #footer><ElButton @click="close">关闭</ElButton></template>
  </ElDrawer>
</template>

<style scoped>
.audit-filters {
  display: flex;
  flex-wrap: wrap;
  align-items: end;
  gap: 18px 22px;
  padding: 22px;
  margin-bottom: 18px;
}
.audit-filters label {
  display: flex;
  flex-direction: column;
  gap: 8px;
  width: 210px;
  font-size: 13px;
  color: var(--el-text-color-regular);
}
.audit-filters .date-filter {
  width: 310px;
}
.audit-filters :deep(.el-date-editor) {
  width: 100%;
}
.actor-filter {
  width: 100%;
}
.audit-hint {
  font-size: 12px;
  color: var(--el-text-color-secondary);
}
.audit-person {
  display: flex;
  align-items: start;
  flex-direction: column;
  gap: 4px;
}
.audit-person small,
.audit-meta small {
  display: block;
  font-size: 12px;
  color: var(--el-text-color-secondary);
}
.audit-detail-heading {
  display: flex;
  align-items: center;
  gap: 12px;
}
.audit-detail-heading h3 {
  margin: 0;
  font-size: 17px;
}
.audit-meta {
  margin-top: 24px;
}
.audit-meta code {
  font-size: 12px;
  word-break: break-all;
  margin-right: 12px;
}
.audit-section-title {
  margin: 28px 0 14px;
  font-size: 14px;
}
.change-values {
  list-style: none;
  padding: 0;
  margin: 0;
  font-size: 12px;
  line-height: 1.8;
  word-break: break-word;
  white-space: pre-wrap;
}
.change-values.after {
  color: #1b7654;
}
.change-values.before {
  color: var(--el-text-color-regular);
}
@media (max-width: 700px) {
  .audit-filters label,
  .audit-filters .date-filter {
    width: 100%;
  }
  .audit-hint {
    display: none;
  }
}
</style>
