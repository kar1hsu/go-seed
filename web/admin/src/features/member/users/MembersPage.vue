<script setup lang="ts">
import { ref, onBeforeUnmount, watch } from 'vue'
import {
  ElButton,
  ElDrawer,
  ElDescriptions,
  ElDescriptionsItem,
  ElTable,
  ElTableColumn,
  ElPagination,
  ElTag,
  ElMessage,
  ElMessageBox,
  ElEmpty,
} from 'element-plus'
import { RefreshRight } from '@element-plus/icons-vue'
import { members, type MemberUser, type MemberDetail } from './api'
import { useList } from '../../../composables/use-list'
import ListFilters from '../../../components/ListFilters.vue'
import { authMessage, useAuthStore } from '../../auth/store'
const auth = useAuthStore()
const { query, filters, items, total, loading, error, load, search, reset } =
  useList(members.list)
const drawer = ref(false),
  detail = ref<MemberDetail | null>(null),
  detailError = ref(''),
  detailLoading = ref(false),
  saving = ref(false)
let sequence = 0
function date(value: string) {
  return value
    ? new Date(value).toLocaleString('zh-CN', { hour12: false })
    : '—'
}
const actions: Record<string, string> = {
  register: '注册账号',
  'profile.update': '修改个人资料',
  'password.change': '修改密码',
  'user.disable': '停用账号',
  'user.enable': '启用账号',
}
async function open(user: MemberUser) {
  const attempt = ++sequence
  drawer.value = true
  detail.value = null
  detailError.value = ''
  detailLoading.value = true
  try {
    const result = await members.detail(user.id)
    if (attempt === sequence) detail.value = result
  } catch (error) {
    if (attempt === sequence) detailError.value = authMessage(error)
  } finally {
    if (attempt === sequence) detailLoading.value = false
  }
}
async function toggle(user: MemberUser) {
  if (saving.value) return
  try {
    await ElMessageBox.confirm(
      user.enabled
        ? `停用「${user.nickname}」后，全部登录会话会立即失效。`
        : `启用「${user.nickname}」后，用户可重新登录。`,
      user.enabled ? '停用会员' : '启用会员',
      {
        type: 'warning',
        confirmButtonText: user.enabled ? '确认停用' : '确认启用',
        cancelButtonText: '取消',
      },
    )
  } catch {
    return
  }
  saving.value = true
  try {
    await members.status(user, !user.enabled)
    ElMessage.success(user.enabled ? '会员已停用' : '会员已启用')
    await load()
    if (drawer.value && detail.value?.user.id === user.id) await open(user)
  } catch (error) {
    ElMessage.error(authMessage(error))
    await load()
  } finally {
    saving.value = false
  }
}
watch(drawer, (value) => {
  if (!value) {
    sequence++
    detail.value = null
    detailError.value = ''
  }
})
watch(
  () => auth.permissions,
  () => {
    if (!auth.has('member.user.read')) {
      drawer.value = false
      items.value = []
    }
  },
)
onBeforeUnmount(() => {
  sequence++
})
</script>
<template>
  <div class="page-heading">
    <div>
      <h1>会员管理</h1>
      <p>查询 C 端用户资料，管理账号状态。</p>
    </div>
  </div>
  <ListFilters
    :filters="filters"
    placeholder="账号 / 昵称"
    :loading="loading"
    @search="search"
    @reset="reset"
  />
  <section class="panel">
    <div class="panel-heading">
      <h2>
        会员列表 <span class="count-label">{{ total }}</span>
      </h2>
      <ElButton :icon="RefreshRight" :loading="loading" @click="load"
        >刷新</ElButton
      >
    </div>
    <div v-if="error" class="error-notice" role="alert">{{ error }}</div>
    <div class="panel-table management-table" :aria-busy="loading">
      <ElTable
        :data="items"
        row-key="id"
        :empty-text="loading ? '加载中…' : '暂无会员，用户注册后将显示在这里'"
      >
        <ElTableColumn prop="id" label="ID" width="90" />
        <ElTableColumn
          prop="username"
          label="账号"
          min-width="160"
          show-overflow-tooltip
        />
        <ElTableColumn
          prop="nickname"
          label="昵称"
          min-width="160"
          show-overflow-tooltip
        />
        <ElTableColumn label="状态" width="100"
          ><template #default="{ row }"
            ><ElTag :type="row.enabled ? 'success' : 'info'" effect="light">{{
              row.enabled ? '启用' : '停用'
            }}</ElTag></template
          ></ElTableColumn
        >
        <ElTableColumn label="最近登录" min-width="180"
          ><template #default="{ row }">{{
            date(row.lastLoginAt)
          }}</template></ElTableColumn
        >
        <ElTableColumn label="注册时间" min-width="180"
          ><template #default="{ row }">{{
            date(row.createdAt)
          }}</template></ElTableColumn
        >
        <ElTableColumn label="操作" width="150" fixed="right"
          ><template #default="{ row }"
            ><ElButton link type="primary" @click="open(row as MemberUser)"
              >详情</ElButton
            ><ElButton
              v-if="auth.has('member.user.update_status')"
              link
              :type="row.enabled ? 'danger' : 'primary'"
              :disabled="saving"
              @click="toggle(row as MemberUser)"
              >{{ row.enabled ? '停用' : '启用' }}</ElButton
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
        :total="total"
        layout="sizes, prev, pager, next"
        @change="load"
      />
    </div>
  </section>
  <ElDrawer
    v-model="drawer"
    title="会员详情"
    size="min(680px, 100vw)"
    destroy-on-close
  >
    <div v-if="detailError" class="error-notice" role="alert">
      {{ detailError }}
    </div>
    <div class="member-detail" :aria-busy="detailLoading">
      <p v-if="detailLoading" class="drawer-intro">加载会员资料…</p>
      <template v-if="detail">
        <ElDescriptions :column="1" border>
          <ElDescriptionsItem label="会员 ID">{{
            detail.user.id
          }}</ElDescriptionsItem>
          <ElDescriptionsItem label="账号">{{
            detail.user.username
          }}</ElDescriptionsItem>
          <ElDescriptionsItem label="昵称">{{
            detail.user.nickname
          }}</ElDescriptionsItem>
          <ElDescriptionsItem label="状态"
            ><ElTag :type="detail.user.enabled ? 'success' : 'info'">{{
              detail.user.enabled ? '启用' : '停用'
            }}</ElTag></ElDescriptionsItem
          >
          <ElDescriptionsItem label="注册时间">{{
            date(detail.user.createdAt)
          }}</ElDescriptionsItem>
          <ElDescriptionsItem label="最近登录">{{
            date(detail.user.lastLoginAt)
          }}</ElDescriptionsItem>
          <ElDescriptionsItem label="资料更新">{{
            date(detail.user.updatedAt)
          }}</ElDescriptionsItem>
        </ElDescriptions>
        <h3 class="member-events-title">最近操作</h3>
        <p class="drawer-intro">显示最近 20 条账号变更记录。</p>
        <ElTable :data="detail.recentEvents" empty-text="暂无操作记录">
          <ElTableColumn label="操作" min-width="120"
            ><template #default="{ row }">{{
              actions[row.action] || row.action
            }}</template></ElTableColumn
          >
          <ElTableColumn label="操作人" min-width="110"
            ><template #default="{ row }"
              >{{ row.actorKind === 'admin' ? '管理员' : '会员' }} #{{
                row.actorId
              }}</template
            ></ElTableColumn
          >
          <ElTableColumn label="时间" min-width="180"
            ><template #default="{ row }">{{
              date(row.createdAt)
            }}</template></ElTableColumn
          >
          <ElTableColumn
            prop="traceId"
            label="Trace ID"
            min-width="160"
            show-overflow-tooltip
          />
        </ElTable>
      </template>
      <ElEmpty
        v-else-if="!detailLoading && !detailError"
        description="暂无会员资料"
      />
    </div>
    <template #footer
      ><ElButton @click="drawer = false">关闭</ElButton></template
    >
  </ElDrawer>
</template>
<style scoped>
.member-detail {
  min-height: 180px;
}
.member-events-title {
  margin: 28px 0 8px;
  font-size: 15px;
  font-weight: 600;
}
</style>
