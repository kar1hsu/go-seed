<script setup lang="ts">
import { computed, reactive, ref, watch } from 'vue'
import {
  ElButton,
  ElDrawer,
  ElForm,
  ElFormItem,
  ElInput,
  ElSwitch,
  ElCheckbox,
  ElTable,
  ElTableColumn,
  ElPagination,
  ElTag,
  ElMessage,
  ElMessageBox,
  type FormInstance,
  type FormRules,
} from 'element-plus'
import { Plus, RefreshRight, Lock } from '@element-plus/icons-vue'
import { management, type Role, type Permission } from '../api'
import { useList } from '../use-list'
import ListFilters from '../components/ListFilters.vue'
import { authMessage, useAuthStore } from '../../auth/store'
const auth = useAuthStore()
const { query, filters, items, total, loading, error, load, search, reset } =
  useList(management.roles)
const drawer = ref<'create' | 'edit' | 'grant' | ''>('')
const selected = ref<Role | null>(null)
const form = reactive({ code: '', name: '', description: '', enabled: true })
const formRef = ref<FormInstance>()
const rules: FormRules = {
  code: [
    {
      required: true,
      pattern: /^[a-z][a-z0-9_]{2,63}$/,
      message: '3–64 位小写字母、数字或下划线，以字母开头',
      trigger: 'blur',
    },
  ],
  name: [
    {
      required: true,
      whitespace: true,
      max: 64,
      message: '请输入 1–64 字的角色名称',
      trigger: 'blur',
    },
  ],
}
const saving = ref(false)
const catalogLoading = ref(false)
const catalog = ref<Permission[]>([])
const checked = ref<string[]>([])
const formError = ref('')
const title = computed(
  () =>
    ({ create: '新增角色', edit: '编辑角色', grant: '角色权限', '': '' })[
      drawer.value
    ],
)
const groups = computed(() => [...new Set(catalog.value.map((p) => p.group))])
const canGrant = computed(
  () => auth.has('system.role.grant') && !!selected.value?.manageable,
)
let openSequence = 0
async function open(mode: 'create' | 'edit' | 'grant', row?: Role) {
  const attempt = ++openSequence
  selected.value = row || null
  Object.assign(form, {
    code: row?.code || '',
    name: row?.name || '',
    description: row?.description || '',
    enabled: row?.enabled ?? true,
  })
  checked.value = [...(row?.permissions || [])]
  formError.value = ''
  drawer.value = mode
  if (mode !== 'grant') return
  catalog.value = []
  catalogLoading.value = true
  try {
    const data = await management.permissions()
    if (attempt === openSequence) catalog.value = data.items
  } catch (cause) {
    if (attempt === openSequence) formError.value = authMessage(cause)
  } finally {
    if (attempt === openSequence) catalogLoading.value = false
  }
}
function toggle(permission: Permission, value: boolean) {
  const next = new Set(checked.value)
  function add(code: string) {
    if (next.has(code)) return
    next.add(code)
    catalog.value.find((p) => p.code === code)?.requires.forEach(add)
  }
  if (value) add(permission.code)
  else {
    next.delete(permission.code)
    // Removing a menu also removes every dependent action, including cross-menu dependencies.
    let changed = true
    while (changed) {
      changed = false
      for (const p of catalog.value)
        if (next.has(p.code) && p.requires.some((code) => !next.has(code))) {
          next.delete(p.code)
          changed = true
        }
    }
  }
  checked.value = [...next]
}
async function save() {
  if (saving.value || catalogLoading.value) return
  if (
    drawer.value !== 'grant' &&
    !(await formRef.value?.validate().catch(() => false))
  )
    return
  saving.value = true
  formError.value = ''
  try {
    if (drawer.value === 'create')
      await management.createRole({
        code: form.code,
        name: form.name.trim(),
        description: form.description.trim(),
      })
    else if (selected.value && drawer.value === 'edit')
      await management.updateRole(selected.value.id, {
        name: form.name.trim(),
        description: form.description.trim(),
        enabled: form.enabled,
        revision: selected.value.revision,
      })
    else if (selected.value)
      await management.grantRole(selected.value.id, {
        permissions: checked.value,
        revision: selected.value.revision,
      })
    drawer.value = ''
    ElMessage.success('保存成功')
    await auth.permissionsChanged()
    await load()
  } catch (cause) {
    formError.value = authMessage(cause)
  } finally {
    saving.value = false
  }
}
async function remove(row: Role) {
  try {
    await ElMessageBox.confirm(`确定删除角色“${row.name}”吗？`, '删除角色', {
      type: 'warning',
      confirmButtonText: '删除',
      cancelButtonText: '取消',
    })
  } catch {
    return
  }
  try {
    await management.deleteRole(row.id, row.revision)
    ElMessage.success('角色已删除')
    await auth.permissionsChanged()
    await load()
  } catch (cause) {
    ElMessage.error(authMessage(cause))
    await load()
  }
}
watch(drawer, (value) => {
  if (!value) {
    openSequence++
    catalogLoading.value = false
  }
})
watch(
  () => auth.permissions,
  () => {
    drawer.value = ''
    if (auth.admin) void load()
  },
)
</script>
<template>
  <div class="page-heading">
    <div>
      <h1>角色管理</h1>
      <p>按职责组织权限，为角色设置菜单访问与按钮操作范围。</p>
    </div>
  </div>
  <ListFilters
    :filters="filters"
    placeholder="角色名称 / 标识"
    :loading="loading"
    @search="search"
    @reset="reset"
  />
  <div v-if="error" class="error-notice" role="alert">{{ error }}</div>
  <section class="panel">
    <div class="panel-heading">
      <h2>
        角色列表 <span class="count-label">{{ total }}</span>
      </h2>
      <div class="table-tools">
        <ElButton
          v-if="auth.has('system.role.create')"
          type="primary"
          :icon="Plus"
          @click="open('create')"
          >新增角色</ElButton
        ><ElButton
          :icon="RefreshRight"
          :loading="loading"
          aria-label="刷新角色列表"
          @click="load"
        />
      </div>
    </div>
    <div class="panel-table management-table" :aria-busy="loading">
      <ElTable
        :data="items"
        row-key="id"
        :empty-text="loading ? '加载中…' : '暂无角色'"
      >
        <ElTableColumn label="角色" min-width="210"
          ><template #default="{ row }"
            ><div class="role-cell">
              <strong
                >{{ row.name }}
                <ElTag
                  v-if="row.builtIn"
                  size="small"
                  type="info"
                  effect="plain"
                  >内置</ElTag
                ></strong
              ><small>{{ row.code }}</small>
            </div></template
          ></ElTableColumn
        >
        <ElTableColumn
          prop="description"
          label="说明"
          min-width="220"
          show-overflow-tooltip
        />
        <ElTableColumn label="状态" width="100"
          ><template #default="{ row }"
            ><span class="status-dot" :class="{ enabled: row.enabled }"></span
            >{{ row.enabled ? '启用' : '停用' }}</template
          ></ElTableColumn
        >
        <ElTableColumn prop="memberCount" label="成员数" width="90" />
        <ElTableColumn label="权限数" width="90"
          ><template #default="{ row }">{{
            row.permissions.length
          }}</template></ElTableColumn
        >
        <ElTableColumn label="操作" width="220" fixed="right"
          ><template #default="{ row }"
            ><ElButton
              v-if="auth.has('system.role.update')"
              link
              type="primary"
              :disabled="!row.manageable"
              @click="open('edit', row as Role)"
              >编辑</ElButton
            ><ElButton
              link
              type="primary"
              @click="open('grant', row as Role)"
              >{{
                auth.has('system.role.grant') && row.manageable
                  ? '设置权限'
                  : '查看权限'
              }}</ElButton
            ><ElButton
              v-if="auth.has('system.role.delete')"
              link
              type="danger"
              :disabled="!row.manageable || row.memberCount > 0"
              @click="remove(row as Role)"
              >删除</ElButton
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
    :model-value="!!drawer"
    :title="title"
    size="520px"
    :close-on-click-modal="!saving"
    :close-on-press-escape="!saving"
    :show-close="!saving"
    destroy-on-close
    @update:model-value="drawer = ''"
  >
    <p class="drawer-intro">
      {{
        drawer === 'create'
          ? '创建角色后，可在权限设置中选择可访问的菜单和操作。'
          : `${selected?.name} · ${selected?.code}`
      }}
    </p>
    <div v-if="formError" class="error-notice" role="alert">
      {{ formError }}
    </div>
    <template v-if="drawer === 'grant'">
      <div v-if="selected?.builtIn" class="permission-note">
        <Lock class="small-icon" />内置超级管理员拥有全部有效权限，不可修改。
      </div>
      <p v-else class="field-help">
        勾选操作会自动选中所需菜单；取消菜单会同时移除依赖它的操作。
      </p>
      <div class="permission-summary">
        已选 {{ checked.length }} 项权限<span>菜单 / 按钮</span>
      </div>
      <p v-if="catalogLoading" class="muted">正在加载权限…</p>
      <section v-for="group in groups" :key="group" class="permission-group">
        <h3>{{ group }}</h3>
        <div
          v-for="permission in catalog.filter((p) => p.group === group)"
          :key="permission.code"
          class="permission-row"
          :class="{ 'is-action': permission.kind === 'button' }"
        >
          <ElCheckbox
            :model-value="checked.includes(permission.code)"
            :disabled="saving || !canGrant || !auth.has(permission.code)"
            @change="(value) => toggle(permission, !!value)"
            >{{ permission.name }}</ElCheckbox
          ><ElTag size="small" type="info" effect="plain">{{
            permission.kind === 'menu' ? '菜单' : '按钮'
          }}</ElTag>
        </div>
      </section>
    </template>
    <ElForm
      v-else
      ref="formRef"
      :model="form"
      :rules="rules"
      label-position="top"
      :disabled="saving"
      @submit.prevent="save"
    >
      <ElFormItem label="角色名称" prop="name"
        ><ElInput
          v-model="form.name"
          maxlength="64"
          show-word-limit
          placeholder="例如：运营管理员"
      /></ElFormItem>
      <ElFormItem label="角色标识" prop="code"
        ><ElInput
          v-model="form.code"
          :disabled="drawer === 'edit'"
          maxlength="64"
          placeholder="例如：operator"
        />
        <p class="field-help">角色标识创建后不可修改。</p></ElFormItem
      >
      <ElFormItem label="说明"
        ><ElInput
          v-model="form.description"
          type="textarea"
          :rows="3"
          maxlength="255"
          show-word-limit
      /></ElFormItem>
      <ElFormItem v-if="drawer === 'edit'" label="角色状态"
        ><ElSwitch
          v-model="form.enabled"
          active-text="启用"
          inactive-text="停用"
        />
        <p class="field-help">
          停用后，关联成员将失去此角色授予的权限。
        </p></ElFormItem
      >
    </ElForm>
    <template #footer
      ><ElButton :disabled="saving" @click="drawer = ''">{{
        drawer === 'grant' && !canGrant ? '关闭' : '取消'
      }}</ElButton
      ><ElButton
        v-if="drawer !== 'grant' || canGrant"
        type="primary"
        :loading="saving"
        :disabled="catalogLoading || (drawer === 'grant' && !catalog.length)"
        @click="save"
        >保存</ElButton
      ></template
    >
  </ElDrawer>
</template>
