<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import {
  ElButton,
  ElInput,
  ElSelect,
  ElOption,
  ElPagination,
  ElTag,
} from 'element-plus'
import { RefreshRight, Search, Delete } from '@element-plus/icons-vue'
import { useDashboardStore } from '../store'
import { formatDate } from '../time'
import ProbeRecords from '../components/ProbeRecords.vue'
const state = useDashboardStore()
const keyword = ref('')
const outcome = ref('all')
const page = ref(1)
const filtered = computed(() =>
  state.records.filter(
    (row) =>
      (outcome.value === 'all' ||
        row.success === (outcome.value === 'success')) &&
      (!keyword.value ||
        row.traceId.toLowerCase().includes(keyword.value.trim().toLowerCase())),
  ),
)
const visible = computed(() =>
  filtered.value.slice((page.value - 1) * 10, page.value * 10),
)
watch([keyword, outcome, () => state.records.length], () => {
  page.value = 1
})
function reset() {
  keyword.value = ''
  outcome.value = 'all'
}
onMounted(() => {
  if (!state.checkedAt) void state.refresh()
})
</script>
<template>
  <div class="page-heading">
    <div>
      <h1>服务状态</h1>
      <p>检查管理端 API 到系统服务的通信状态。</p>
    </div>
  </div>
  <section class="panel connection-summary">
    <div>
      <span class="field-label">当前状态</span
      ><ElTag
        :type="
          state.loading
            ? 'info'
            : state.result
              ? 'success'
              : state.checkedAt
                ? 'danger'
                : 'info'
        "
        effect="plain"
        >{{
          state.loading
            ? '检查中'
            : state.result
              ? '连接正常'
              : state.checkedAt
                ? '连接异常'
                : '未检查'
        }}</ElTag
      >
    </div>
    <div>
      <span class="field-label">检查链路</span
      ><span>Admin API → System RPC</span>
    </div>
    <div>
      <span class="field-label">最近检查</span
      ><span>{{ formatDate(state.checkedAt) }}</span>
    </div>
    <ElButton
      type="primary"
      :icon="RefreshRight"
      :loading="state.loading"
      @click="state.refresh"
      >立即检查</ElButton
    >
  </section>
  <div v-if="state.error" class="error-notice" role="alert">
    {{ state.error }}。{{ state.traceId ? `Trace ID：${state.traceId}` : '' }}
  </div>
  <section class="panel filter-panel">
    <label
      ><span>Trace ID</span
      ><ElInput
        v-model="keyword"
        :prefix-icon="Search"
        placeholder="请输入 Trace ID"
        aria-label="搜索 Trace ID"
        clearable /></label
    ><label
      ><span>检查结果</span
      ><ElSelect v-model="outcome" aria-label="检查结果"
        ><ElOption label="全部" value="all" /><ElOption
          label="成功"
          value="success" /><ElOption
          label="失败"
          value="failure" /></ElSelect></label
    ><ElButton @click="reset">重置</ElButton>
  </section>
  <section class="panel">
    <div class="panel-heading">
      <h2>检查记录</h2>
      <ElButton
        :icon="Delete"
        :disabled="!state.records.length || state.loading"
        @click="state.clearRecords"
        >清空记录</ElButton
      >
    </div>
    <div class="panel-table">
      <ProbeRecords :records="visible" :loading="state.loading" />
    </div>
    <div class="table-pagination">
      <span>共 {{ filtered.length }} 条记录</span
      ><ElPagination
        v-model:current-page="page"
        layout="prev, pager, next"
        :page-size="10"
        :total="filtered.length"
      />
    </div>
  </section>
  <p class="page-note">
    最多保留 20 条本次打开页面后的记录，刷新页面后清空。此页面验证 API / RPC
    连通性，数据库和 Redis 就绪状态通过后端 /readyz 检查。
  </p>
</template>
