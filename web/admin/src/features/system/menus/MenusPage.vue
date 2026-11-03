<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { ArrowDown, ArrowUp, Plus, Refresh } from '@element-plus/icons-vue'
import {
  ElButton,
  ElDialog,
  ElForm,
  ElFormItem,
  ElIcon,
  ElInput,
  ElMessage,
  ElMessageBox,
  ElOption,
  ElSelect,
  ElSwitch,
  ElTable,
  ElTableColumn,
  ElTag,
} from 'element-plus'
import { authMessage, useAuthStore } from '../../auth/store'
import { randomUUID } from '../../../lib/random'
import { api } from '../../../lib/http'
import { unwrap } from '../../../lib/api-response'
import { menuIcon, navigation } from '../../../router/navigation'
import type { components } from '../../../api/generated/schema'

type Group = components['schemas']['MenuGroup']
type Page = components['schemas']['MenuPage']
type Tree = components['schemas']['MenuTree']
interface Row {
  key: string
  group: Group
  page?: Page
  index: number
  count: number
  children?: Row[]
}
const auth = useAuthStore()
const tree = ref<Tree>({ revision: '', groups: [] })
const baseline = ref('')
const loading = ref(true)
const saving = ref(false)
const error = ref('')
const edit = ref<{
  group?: Group
  page?: Page
  title: string
  icon: string
  hidden: boolean
  groupId: string
} | null>(null)
const canEdit = computed(() => auth.has('system.menu.update'))
const dirty = computed(
  () =>
    !!tree.value.revision &&
    JSON.stringify(tree.value.groups) !== baseline.value,
)
const iconChoices = [...new Set(navigation.map((r) => r.iconName))]
const rows = computed<Row[]>(() =>
  tree.value.groups.map((group, index) => ({
    key: 'group:' + group.id,
    group,
    index,
    count: tree.value.groups.length,
    children: group.pages.map((page, i) => ({
      key: 'page:' + page.path,
      group,
      page,
      index: i,
      count: group.pages.length,
    })),
  })),
)
let active = true
function received(value: Tree) {
  tree.value = structuredClone(value)
  baseline.value = JSON.stringify(value.groups)
}
async function load() {
  loading.value = true
  error.value = ''
  try {
    const response = await unwrap(api.GET('/admin-api/v1/system/menus'))
    if (active) received(response)
  } catch (cause) {
    if (active) error.value = authMessage(cause)
  } finally {
    if (active) loading.value = false
  }
}
async function reload() {
  if (dirty.value) {
    try {
      await ElMessageBox.confirm(
        '重新加载将放弃本页未保存的调整。',
        '重新加载菜单',
        { confirmButtonText: '重新加载', cancelButtonText: '取消' },
      )
    } catch {
      return
    }
  }
  await load()
}
function open(row?: Row) {
  const item = row?.page ?? row?.group
  edit.value = {
    group: row?.group,
    page: row?.page,
    title: item?.title ?? '',
    icon: item?.icon ?? 'FolderOpened',
    hidden: item?.hidden ?? false,
    groupId: row?.group.id ?? '',
  }
}
function apply() {
  const form = edit.value
  if (!form || !canEdit.value) return
  const title = form.title.trim()
  if (!title) {
    ElMessage.warning('请输入菜单名称')
    return
  }
  const display = { title, icon: form.icon, hidden: form.hidden }
  if (form.page && form.group) {
    Object.assign(form.page, display)
    if (form.groupId !== form.group.id) {
      const target = tree.value.groups.find((g) => g.id === form.groupId)!
      form.group.pages.splice(form.group.pages.indexOf(form.page), 1)
      target.pages.push(form.page)
    }
  } else if (form.group) Object.assign(form.group, display)
  else tree.value.groups.push({ id: randomUUID(), ...display, pages: [] })
  edit.value = null
}
function shift<T>(items: T[], i: number, by: number) {
  const dest = i + by
  if (dest < 0 || dest >= items.length) return
  const [item] = items.splice(i, 1)
  items.splice(dest, 0, item!)
}
function move(row: Row, by: number) {
  if (row.page) shift(row.group.pages, row.index, by)
  else shift(tree.value.groups, row.index, by)
}
async function save() {
  if (!dirty.value || !canEdit.value || saving.value) return
  saving.value = true
  error.value = ''
  try {
    const response = await unwrap(
      api.PUT('/admin-api/v1/system/menus', {
        params: { header: { 'X-GoSeed-CSRF': '1' } },
        body: tree.value,
      }),
    )
    if (!active) return
    received(response)
    auth.menuSaved(response)
    ElMessage.success('菜单已保存')
  } catch (cause) {
    if (active) error.value = authMessage(cause)
  } finally {
    if (active) saving.value = false
  }
}
watch(canEdit, (allowed) => {
  if (!allowed) {
    edit.value = null
    if (baseline.value) tree.value.groups = JSON.parse(baseline.value)
  }
})
onMounted(load)
onBeforeUnmount(() => {
  active = false
})
</script>
<template>
  <div class="page-heading">
    <div>
      <h1>菜单管理</h1>
      <p>调整后台侧栏的分组和显示顺序。页面访问权限在角色管理中配置。</p>
    </div>
  </div>
  <div v-if="error" role="alert" class="error-notice">{{ error }}</div>
  <section class="panel">
    <div class="panel-heading menu-toolbar">
      <h2>
        菜单树 <span class="count-label">{{ tree.groups.length }} 个分组</span>
      </h2>
      <div class="table-tools">
        <span v-if="dirty" class="draft-label">有未保存的调整</span>
        <ElButton
          :icon="Refresh"
          :loading="loading"
          :disabled="saving"
          @click="reload"
          >重新加载</ElButton
        >
        <ElButton
          v-if="canEdit"
          :icon="Plus"
          :disabled="!tree.revision || saving || loading"
          @click="open()"
          >新增分组</ElButton
        >
        <ElButton
          v-if="canEdit"
          type="primary"
          :loading="saving"
          :disabled="!dirty || loading"
          @click="save"
          >保存菜单</ElButton
        >
      </div>
    </div>
    <ElTable
      :data="rows"
      row-key="key"
      default-expand-all
      :indent="22"
      :empty-text="loading ? '加载中…' : '暂无菜单'"
    >
      <ElTableColumn label="菜单名称" min-width="205">
        <template #default="{ row }">
          <span class="node-title" :class="{ 'is-group': !row.page }">
            <ElIcon
              ><component :is="menuIcon((row.page ?? row.group).icon)"
            /></ElIcon>
            {{ (row.page ?? row.group).title }}
          </span>
        </template>
      </ElTableColumn>
      <ElTableColumn label="页面路径" min-width="190">
        <template #default="{ row }"
          ><code v-if="row.page" class="menu-path">{{ row.page.path }}</code
          ><span v-else class="muted"
            >{{ row.group.pages.length }} 个页面</span
          ></template
        >
      </ElTableColumn>
      <ElTableColumn label="侧栏显示" width="116">
        <template #default="{ row }"
          ><ElTag
            size="small"
            :type="row.group.hidden || row.page?.hidden ? 'info' : 'success'"
            >{{
              row.page && row.group.hidden
                ? '随分组隐藏'
                : (row.page ?? row.group).hidden
                  ? '隐藏'
                  : '显示'
            }}</ElTag
          ></template
        >
      </ElTableColumn>
      <ElTableColumn v-if="canEdit" label="排序" width="102">
        <template #default="{ row }">
          <ElButton
            link
            :icon="ArrowUp"
            :aria-label="'上移' + (row.page ?? row.group).title"
            title="上移"
            :disabled="row.index === 0 || saving || loading"
            @click="move(row as Row, -1)"
          />
          <ElButton
            link
            :icon="ArrowDown"
            :aria-label="'下移' + (row.page ?? row.group).title"
            title="下移"
            :disabled="row.index === row.count - 1 || saving || loading"
            @click="move(row as Row, 1)"
          />
        </template>
      </ElTableColumn>
      <ElTableColumn v-if="canEdit" label="操作" width="142">
        <template #default="{ row }">
          <ElButton
            link
            type="primary"
            :disabled="saving || loading"
            @click="open(row as Row)"
            >编辑</ElButton
          >
          <ElButton
            v-if="!row.page && !row.group.pages.length"
            link
            type="danger"
            :disabled="saving || loading"
            @click="tree.groups.splice(row.index, 1)"
            >移除分组</ElButton
          >
        </template>
      </ElTableColumn>
    </ElTable>
  </section>
  <ElDialog
    :model-value="!!edit"
    :title="!edit?.group ? '新增分组' : edit.page ? '编辑页面' : '编辑分组'"
    width="min(440px, calc(100% - 32px))"
    align-center
    append-to-body
    destroy-on-close
    @close="edit = null"
  >
    <ElForm v-if="edit" label-position="top" @submit.prevent="apply">
      <ElFormItem label="名称" required
        ><ElInput
          v-model="edit.title"
          maxlength="64"
          autofocus
          aria-label="菜单名称"
      /></ElFormItem>
      <ElFormItem v-if="edit.page" label="所属分组">
        <ElSelect v-model="edit.groupId" aria-label="所属分组"
          ><ElOption
            v-for="g in tree.groups"
            :key="g.id"
            :label="g.title"
            :value="g.id"
        /></ElSelect>
      </ElFormItem>
      <ElFormItem label="图标">
        <ElSelect v-model="edit.icon" aria-label="菜单图标"
          ><template #prefix
            ><ElIcon><component :is="menuIcon(edit.icon)" /></ElIcon
          ></template>
          <ElOption
            v-for="name in iconChoices"
            :key="name"
            :value="name"
            :label="name"
            ><span class="node-title"
              ><ElIcon><component :is="menuIcon(name)" /></ElIcon
              >{{ name }}</span
            ></ElOption
          >
        </ElSelect>
      </ElFormItem>
      <ElFormItem label="侧栏可见性"
        ><ElSwitch
          v-model="edit.hidden"
          active-text="隐藏"
          inactive-text="显示"
          aria-label="隐藏菜单"
      /></ElFormItem>
    </ElForm>
    <template #footer
      ><ElButton @click="edit = null">取消</ElButton
      ><ElButton type="primary" @click="apply">应用到菜单树</ElButton></template
    >
  </ElDialog>
</template>
<style scoped>
.node-title {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  vertical-align: middle;
}
.node-title.is-group {
  font-weight: 600;
}
.menu-path,
.draft-label {
  color: var(--el-text-color-secondary);
  font-size: 12px;
}
.menu-toolbar,
.menu-toolbar .table-tools {
  flex-wrap: wrap;
  gap: 8px;
}
</style>
