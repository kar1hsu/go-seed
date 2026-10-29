<script setup lang="ts">
import { computed } from 'vue'
import { ElButton, ElTag, ElTable, ElTableColumn } from 'element-plus'
import { Position } from '@element-plus/icons-vue'
import { useDashboardStore } from '../../dashboard/store'
import type { paths } from '../../../api/generated/schema'
const state = useDashboardStore()
const location = window.location
const endpoint: keyof paths = '/admin-api/v1/system/ping'
const tested = computed(() => state.checkedAt !== '')
const fields = [
  {
    name: 'code',
    type: 'number',
    description: '0 表示成功，其他值为稳定错误码',
  },
  { name: 'message', type: 'string', description: '请求结果说明' },
  { name: 'data.service', type: 'string', description: 'RPC 服务名称' },
  { name: 'data.version', type: 'string', description: '系统服务版本' },
  {
    name: 'data.time',
    type: 'string',
    description: '服务端 UTC 时间（RFC3339）',
  },
  {
    name: 'traceId',
    type: 'string',
    description: '用于关联服务日志的链路标识',
  },
]
const response = computed(
  () =>
    state.result || {
      code: state.records[0]?.code ?? null,
      message: state.error,
      data: null,
      traceId: state.traceId,
    },
)
</script>
<template>
  <div class="page-heading">
    <div>
      <h1>接口文档</h1>
      <p>当前管理端已开放的服务探测接口。</p>
    </div>
    <ElTag type="info" effect="plain">API v1</ElTag>
  </div>
  <section class="panel">
    <div class="panel-heading">
      <h2>系统服务探测</h2>
      <ElTag type="success" effect="plain">无需认证</ElTag>
    </div>
    <div class="panel-body">
      <div class="endpoint-line">
        <span class="method-badge">GET</span><code>{{ endpoint }}</code>
      </div>
      <p class="body-description">
        调用 system-rpc
        获取服务名称、版本和服务端时间。无请求参数，仅返回非敏感信息。
      </p>
      <h3 class="subsection-title">请求示例</h3>
      <pre class="code-block">curl {{ `${location.origin}${endpoint}` }}</pre>
      <div class="request-actions">
        <ElButton
          type="primary"
          :icon="Position"
          :loading="state.loading"
          @click="state.refresh"
          >发送测试请求</ElButton
        ><span class="muted">使用当前环境，结果同步到服务检查记录</span>
      </div>
    </div>
  </section>
  <div class="docs-grid">
    <section class="panel">
      <div class="panel-heading"><h2>响应字段</h2></div>
      <div class="panel-table">
        <ElTable :data="fields"
          ><ElTableColumn
            prop="name"
            label="字段"
            min-width="135" /><ElTableColumn
            prop="type"
            label="类型"
            width="85" /><ElTableColumn
            prop="description"
            label="说明"
            min-width="220"
        /></ElTable>
      </div>
    </section>
    <section class="panel">
      <div class="panel-heading">
        <h2>实际响应</h2>
        <span class="panel-caption">{{
          state.loading
            ? '请求中…'
            : tested
              ? `${state.latency} ms`
              : '尚未请求'
        }}</span>
      </div>
      <div class="panel-body">
        <pre v-if="tested && !state.loading" class="code-block response-code">{{
          JSON.stringify(response, null, 2)
        }}</pre>
        <div v-else class="plain-empty">
          {{ state.loading ? '正在请求…' : '点击“发送测试请求”查看实际响应' }}
        </div>
      </div>
    </section>
  </div>
  <p class="page-note">
    依赖不可用：HTTP 503 / 100503；请求超时：HTTP 504 / 100504；内部错误：HTTP
    500 / 100500。错误响应的 data 为 null。
  </p>
</template>
