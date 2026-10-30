<script setup lang="ts">
import { validPassword } from '../../../lib/random'
import { computed, reactive, ref, watch } from 'vue'
import {
  ElButton,
  ElDrawer,
  ElForm,
  ElFormItem,
  ElInput,
  ElSwitch,
  ElSelect,
  ElOption,
  ElPagination,
  ElTable,
  ElTableColumn,
  ElTag,
  ElMessage,
  type FormInstance,
  type FormRules,
} from 'element-plus'
import { Plus, RefreshRight } from '@element-plus/icons-vue'
import { management, type Administrator, type Role } from '../api'
import { useList } from '../use-list'
import ListFilters from '../components/ListFilters.vue'
import { authMessage, useAuthStore } from '../../auth/store'
const auth = useAuthStore()
const { query, filters, items, total, loading, error, load, search, reset } =
  useList(management.administrators)
const drawer = ref<'create' | 'edit' | 'roles' | ''>('')
const selected = ref<Administrator | null>(null)
const formRef = ref<FormInstance>()
const form = reactive({
  username: '',
  displayName: '',
  password: '',
  enabled: true,
  roleIds: [] as string[],
})
const roles = ref<Role[]>([])
const saving = ref(false)
const rolesLoading = ref(false)
const formError = ref('')
const rolesLoadFailed = ref(false)
const title = computed(
  () =>
    ({ create: '新增管理员', edit: '编辑管理员', roles: '分配角色', '': '' })[
      drawer.value
    ],
)
const rules: FormRules = {
  username: [
    { required: true, message: '请输入登录账号', trigger: 'blur' },
    {
      pattern: /^[a-z][a-z0-9_.-]{2,31}$/,
      message: '3–32 位小写字母、数字、点、下划线或短横线，以字母开头',
      trigger: 'blur',
    },
  ],
  displayName: [
    {
      required: true,
      whitespace: true,
      max: 64,
      message: '请输入 1–64 字的姓名',
      trigger: 'blur',
    },
  ],
  password: [
    {
      required: true,
      validator: (_rule, value, callback) =>
        callback(
          validPassword(value)
            ? undefined
            : new Error('密码须为 12–72 字节，中文通常占 3 字节'),
        ),
      trigger: 'blur',
    },
  ],
}
let openSequence = 0
async function open(mode: 'create' | 'edit' | 'roles', row?: Administrator) {
  const attempt = ++openSequence
  selected.value = row || null
  Object.assign(form, {
    username: row?.username || '',
    displayName: row?.displayName || '',
    password: '',
    enabled: row?.enabled ?? true,
    roleIds: [...(row?.roleIds || [])],
  })
  formError.value = ''
  rolesLoadFailed.value = false
  drawer.value = mode
  if (mode !== 'roles') return
  roles.value = []
  rolesLoading.value = true
  try {
    // Fetch every page; never silently drop assignments outside the first 100 roles.
    const all: Role[] = []
    for (let page = 1; ; page++) {
      const data = await management.roles({
        page,
        pageSize: 100,
        keyword: '',
        enabled: 'all',
      })
      if (attempt !== openSequence) return
      all.push(...data.items)
      if (all.length >= data.total || !data.items.length) break
    }
    roles.value = all
  } catch (cause) {
    if (attempt === openSequence) {
      formError.value = authMessage(cause)
      rolesLoadFailed.value = true
    }
  } finally {
    if (attempt === openSequence) rolesLoading.value = false
  }
}
async function save() {
  if (saving.value || rolesLoading.value) return
  if (
    drawer.value !== 'roles' &&
    !(await formRef.value?.validate().catch(() => false))
  )
    return
  saving.value = true
  formError.value = ''
  try {
    if (drawer.value === 'create')
      await management.createAdministrator({
        username: form.username,
        displayName: form.displayName.trim(),
        password: form.password,
      })
    else if (selected.value && drawer.value === 'edit')
      await management.updateAdministrator(selected.value.id, {
        displayName: form.displayName.trim(),
        enabled: form.enabled,
        revision: selected.value.revision,
      })
    else if (selected.value)
      await management.assignRoles(selected.value.id, {
        roleIds: form.roleIds,
        revision: selected.value.revision,
      })
    drawer.value = ''
    form.password = ''
    ElMessage.success('保存成功')
    await auth.permissionsChanged()
    await load()
  } catch (cause) {
    formError.value = authMessage(cause)
  } finally {
    saving.value = false
  }
}
watch(drawer, (value) => {
  if (!value) {
    openSequence++
    form.password = ''
    rolesLoading.value = false
  }
})
watch(
  () => auth.permissions,
  () => {
    if (drawer.value) drawer.value = ''
    if (auth.admin) void load()
  },
)
function date(value: string) {
  return new Intl.DateTimeFormat('zh-CN', { dateStyle: 'medium' }).format(
    new Date(value),
  )
}
</script>
<template>
  <div class="page-heading">
    <div>
      <h1>管理员管理</h1>
      <p>管理后台成员、账户状态及角色分配。</p>
    </div>
  </div>
  <ListFilters
    :filters="filters"
    placeholder="账号 / 姓名"
    :loading="loading"
    @search="search"
    @reset="reset"
  />
  <div v-if="error" class="error-notice" role="alert">{{ error }}</div>
  <section class="panel">
    <div class="panel-heading">
      <h2>
        管理员列表 <span class="count-label">{{ total }}</span>
      </h2>
      <div class="table-tools">
        <ElButton
          v-if="auth.has('system.user.create')"
          type="primary"
          :icon="Plus"
          @click="open('create')"
          >新增管理员</ElButton
        ><ElButton
          :icon="RefreshRight"
          :loading="loading"
          aria-label="刷新管理员列表"
          @click="load"
        />
      </div>
    </div>
    <div class="panel-table management-table" :aria-busy="loading">
      <ElTable
        :data="items"
        row-key="id"
        :empty-text="loading ? '加载中…' : '暂无管理员'"
      >
        <ElTableColumn label="管理员" min-width="220"
          ><template #default="{ row }"
            ><div class="member-cell">
              <span class="member-avatar">{{
                row.displayName.slice(0, 1)
              }}</span>
              <div>
                <strong>{{ row.displayName }}</strong
                ><small
                  >{{ row.username
                  }}<span v-if="row.id === auth.admin?.id" class="self-label"
                    >当前账户</span
                  ></small
                >
              </div>
            </div></template
          ></ElTableColumn
        >
        <ElTableColumn label="角色" min-width="200"
          ><template #default="{ row }"
            ><div class="inline-tags">
              <ElTag
                v-for="(name, index) in row.roleNames"
                :key="row.roleIds[index]"
                type="info"
                effect="plain"
                >{{ name }}</ElTag
              ><span v-if="!row.roleNames.length" class="muted">未分配</span>
            </div></template
          ></ElTableColumn
        >
        <ElTableColumn label="状态" width="110"
          ><template #default="{ row }"
            ><span class="status-dot" :class="{ enabled: row.enabled }"></span
            >{{ row.enabled ? '启用' : '停用' }}</template
          ></ElTableColumn
        >
        <ElTableColumn label="创建时间" width="150"
          ><template #default="{ row }">{{
            date(row.createdAt)
          }}</template></ElTableColumn
        >
        <ElTableColumn
          v-if="
            auth.has('system.user.update') ||
            auth.has('system.user.assign_role')
          "
          label="操作"
          width="180"
          fixed="right"
          ><template #default="{ row }"
            ><ElButton
              v-if="auth.has('system.user.update')"
              link
              type="primary"
              :disabled="!row.manageable"
              @click="open('edit', row as Administrator)"
              >编辑</ElButton
            ><ElButton
              v-if="auth.has('system.user.assign_role')"
              link
              type="primary"
              :disabled="
                !row.manageable ||
                (row.id === auth.admin?.id && !auth.superAdmin)
              "
              @click="open('roles', row as Administrator)"
              >分配角色</ElButton
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
    size="480px"
    :close-on-click-modal="!saving"
    :close-on-press-escape="!saving"
    :show-close="!saving"
    destroy-on-close
    @update:model-value="drawer = ''"
  >
    <p class="drawer-intro">
      {{
        drawer === 'create'
          ? '创建后可为成员分配角色，授予所需的访问权限。'
          : `${selected?.displayName} · ${selected?.username}`
      }}
    </p>
    <div v-if="formError" class="error-notice" role="alert">
      {{ formError }}
    </div>
    <ElForm
      ref="formRef"
      :model="form"
      :rules="rules"
      label-position="top"
      :disabled="saving"
      @submit.prevent="save"
    >
      <template v-if="drawer !== 'roles'">
        <ElFormItem v-if="drawer === 'create'" label="登录账号" prop="username"
          ><ElInput
            v-model="form.username"
            autocomplete="off"
            maxlength="32"
            placeholder="例如：zhangsan"
        /></ElFormItem>
        <ElFormItem label="姓名" prop="displayName"
          ><ElInput v-model="form.displayName" maxlength="64" show-word-limit
        /></ElFormItem>
        <ElFormItem v-if="drawer === 'create'" label="初始密码" prop="password"
          ><ElInput
            v-model="form.password"
            type="password"
            show-password
            autocomplete="new-password"
            placeholder="至少 12 位字符"
          />
          <p class="field-help">
            请通过安全渠道交付，首次登录后在账户安全页修改。
          </p></ElFormItem
        >
        <ElFormItem v-if="drawer === 'edit'" label="账户状态"
          ><ElSwitch
            v-model="form.enabled"
            :disabled="selected?.id === auth.admin?.id"
            active-text="启用"
            inactive-text="停用"
          />
          <p class="field-help">
            停用后会立即注销该账户的所有会话。
          </p></ElFormItem
        >
      </template>
      <ElFormItem v-else label="关联角色"
        ><ElSelect
          v-model="form.roleIds"
          multiple
          filterable
          :loading="rolesLoading"
          placeholder="请选择角色"
          style="width: 100%"
          ><ElOption
            v-for="role in roles"
            :key="role.id"
            :label="`${role.name}${role.enabled ? '' : '（已停用）'}`"
            :value="role.id"
            :disabled="!role.assignable && !form.roleIds.includes(role.id)"
        /></ElSelect>
        <p class="field-help">
          可选择多个角色，权限取并集。仅可分配自己权限范围内的启用角色。
        </p></ElFormItem
      >
    </ElForm>
    <template #footer
      ><ElButton :disabled="saving" @click="drawer = ''">取消</ElButton
      ><ElButton
        type="primary"
        :loading="saving"
        :disabled="rolesLoading || (drawer === 'roles' && rolesLoadFailed)"
        @click="save"
        >保存</ElButton
      ></template
    >
  </ElDrawer>
</template>
