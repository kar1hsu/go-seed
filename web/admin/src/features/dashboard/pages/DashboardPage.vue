<script setup lang="ts">
import { onMounted } from 'vue'
import { ElButton, ElTag } from 'element-plus'
import {
  Connection,
  Timer,
  Box,
  RefreshRight,
  ArrowRight,
  Monitor,
  Document,
  Link,
  InfoFilled,
} from '@element-plus/icons-vue'
import { useDashboardStore } from '../store'
import { formatTime } from '../time'
import ProbeRecords from '../components/ProbeRecords.vue'
import { project } from '../../../app/project'
const state = useDashboardStore()
onMounted(() => {
  if (!state.checkedAt) void state.refresh()
})
</script>
<template>
  <div class="page-heading">
    <div>
      <h1>工作台</h1>
      <p>查看服务运行状态与最近检查记录。</p>
    </div>
    <ElButton
      type="primary"
      :icon="RefreshRight"
      :loading="state.loading"
      @click="state.refresh"
      >刷新数据</ElButton
    >
  </div>
  <div class="metric-grid" aria-live="polite">
    <section class="metric-panel">
      <div class="metric-label">
        服务连接<Connection class="metric-icon blue" />
      </div>
      <div class="metric-value" :class="state.error ? 'is-error' : ''">
        {{
          state.loading
            ? '检查中'
            : state.result
              ? '正常'
              : state.checkedAt
                ? '异常'
                : '未检查'
        }}<span
          v-if="state.result && !state.loading"
          class="success-dot"
        ></span>
      </div>
      <div class="metric-foot">
        管理端 API / System RPC<span v-if="state.result && !state.loading"
          >连通</span
        >
      </div>
    </section>
    <section class="metric-panel">
      <div class="metric-label">
        最近响应时间<Timer class="metric-icon amber" />
      </div>
      <div class="metric-value numeric">
        {{ state.checkedAt ? state.latency : '—' }}<small>ms</small>
      </div>
      <div class="metric-foot">
        浏览器发起请求的耗时<span>{{ formatTime(state.checkedAt) }}</span>
      </div>
    </section>
    <section class="metric-panel">
      <div class="metric-label">
        系统服务版本<Box class="metric-icon violet" />
      </div>
      <div class="metric-value numeric">
        {{ state.result ? `v${state.result.data.version}` : '—' }}
      </div>
      <div class="metric-foot">
        {{ state.result?.data.service || '等待服务响应'
        }}<ElTag type="info" size="small" effect="plain">开发环境</ElTag>
      </div>
    </section>
  </div>
  <div v-if="state.error" class="error-notice" role="alert">
    {{ state.error }}。请确认后端服务已启动，再刷新重试。
  </div>
  <div class="dashboard-grid">
    <section class="panel">
      <div class="panel-heading">
        <h2>最近检查</h2>
        <RouterLink class="text-link" to="/monitor"
          >查看全部<ArrowRight class="tiny-icon"
        /></RouterLink>
      </div>
      <div class="panel-table">
        <ProbeRecords
          :records="state.records.slice(0, 5)"
          :loading="state.loading"
        />
      </div>
      <div class="panel-foot">
        真实检查记录，刷新页面后清空<span
          >最近 {{ Math.min(state.records.length, 5) }} 条</span
        >
      </div>
    </section>
    <section class="panel">
      <div class="panel-heading"><h2>快捷入口</h2></div>
      <div class="shortcut-grid">
        <RouterLink to="/monitor"
          ><Monitor class="shortcut-icon blue" /><span
            >服务状态</span
          ></RouterLink
        ><RouterLink to="/api-docs"
          ><Document class="shortcut-icon green" /><span
            >接口文档</span
          ></RouterLink
        ><a :href="project.repository" target="_blank" rel="noopener noreferrer"
          ><Link class="shortcut-icon violet" /><span>项目仓库</span></a
        ><RouterLink to="/about"
          ><InfoFilled class="shortcut-icon amber" /><span
            >项目信息</span
          ></RouterLink
        >
      </div>
    </section>
  </div>
  <section class="panel system-panel">
    <div class="panel-heading">
      <h2>系统信息</h2>
      <span class="panel-caption">第一阶段 · 工程骨架</span>
    </div>
    <dl class="system-info">
      <div>
        <dt>项目名称</dt>
        <dd>GoSeed</dd>
      </div>
      <div>
        <dt>管理端版本</dt>
        <dd>{{ project.version }}</dd>
      </div>
      <div>
        <dt>服务名称</dt>
        <dd>{{ state.result?.data.service || '—' }}</dd>
      </div>
      <div>
        <dt>接口前缀</dt>
        <dd><code>/admin-api/v1</code></dd>
      </div>
      <div>
        <dt>最近检查</dt>
        <dd>{{ formatTime(state.checkedAt) }}</dd>
      </div>
      <div>
        <dt>身份认证</dt>
        <dd>待接入</dd>
      </div>
    </dl>
    <div class="system-note">
      当前提供非敏感的服务探测。登录、权限管理和业务模块将在后续阶段接入。
    </div>
  </section>
</template>
