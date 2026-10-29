<script setup lang="ts">
import { ref } from 'vue'
import {
  ElTable,
  ElTableColumn,
  ElButton,
  ElDrawer,
  ElDescriptions,
  ElDescriptionsItem,
} from 'element-plus'
import type { ProbeRecord } from '../store'
import { formatTime, formatDate } from '../time'
defineProps<{ records: ProbeRecord[]; loading?: boolean }>()
const selected = ref<ProbeRecord | null>(null)
const showDetail = ref(false)
function viewDetail(record: ProbeRecord) {
  selected.value = record
  showDetail.value = true
}
</script>
<template>
  <ElTable
    :data="records"
    row-key="id"
    class="records-table"
    :empty-text="loading ? '正在检查服务…' : '暂无检查记录'"
    :aria-busy="loading"
  >
    <ElTableColumn label="检查时间" min-width="110"
      ><template #default="{ row }">{{
        formatTime(row.time)
      }}</template></ElTableColumn
    >
    <ElTableColumn label="检查对象" min-width="145"
      ><template #default
        ><span class="table-service">Admin API → RPC</span></template
      ></ElTableColumn
    >
    <ElTableColumn label="结果" width="92"
      ><template #default="{ row }"
        ><span
          class="status-text"
          :class="row.success ? 'is-success' : 'is-error'"
          ><i></i>{{ row.success ? '成功' : '失败' }}</span
        ></template
      ></ElTableColumn
    >
    <ElTableColumn label="耗时" width="95"
      ><template #default="{ row }"
        ><span class="tabular"
          >{{ row.duration }} <span class="muted">ms</span></span
        ></template
      ></ElTableColumn
    >
    <ElTableColumn
      prop="traceId"
      label="Trace ID"
      min-width="140"
      show-overflow-tooltip
      ><template #default="{ row }"
        ><code class="trace-cell">{{ row.traceId || '—' }}</code></template
      ></ElTableColumn
    >
    <ElTableColumn label="操作" width="75" align="center"
      ><template #default="{ row }"
        ><ElButton link type="primary" @click="viewDetail(row as ProbeRecord)"
          >详情</ElButton
        ></template
      ></ElTableColumn
    >
  </ElTable>
  <ElDrawer v-model="showDetail" title="检查详情" size="min(560px, 100vw)">
    <template v-if="selected"
      ><ElDescriptions :column="1" border
        ><ElDescriptionsItem label="检查时间">{{
          formatDate(selected.time)
        }}</ElDescriptionsItem
        ><ElDescriptionsItem label="结果"
          ><span :class="selected.success ? 'is-success' : 'is-error'">{{
            selected.message
          }}</span></ElDescriptionsItem
        ><ElDescriptionsItem label="耗时"
          >{{ selected.duration }} ms</ElDescriptionsItem
        ><ElDescriptionsItem label="错误码">{{
          selected.code ?? '未收到响应'
        }}</ElDescriptionsItem
        ><ElDescriptionsItem label="Trace ID"
          ><code class="break-code">{{
            selected.traceId || '—'
          }}</code></ElDescriptionsItem
        ></ElDescriptions
      >
      <h3 class="subsection-title">响应内容</h3>
      <pre class="code-block">{{
        JSON.stringify(
          selected.response || {
            code: selected.code,
            message: selected.message,
            data: null,
            traceId: selected.traceId,
          },
          null,
          2,
        )
      }}</pre>
    </template>
  </ElDrawer>
</template>
