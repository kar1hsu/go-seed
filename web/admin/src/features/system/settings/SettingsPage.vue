<script setup lang="ts">
import { computed, nextTick, reactive, ref, watch } from 'vue'
import {
  ElButton,
  ElDrawer,
  ElForm,
  ElFormItem,
  ElInput,
  ElSwitch,
  ElSelect,
  ElOption,
  ElTable,
  ElTableColumn,
  ElPagination,
  ElTag,
  ElMessage,
  ElMessageBox,
  type FormInstance,
} from 'element-plus'
import { Plus, RefreshRight } from '@element-plus/icons-vue'
import { settings, type Setting } from './api'
import { resourceRules, validateSettingValue } from './form'
import { useList } from '../use-list'
import ListFilters from '../components/ListFilters.vue'
import { authMessage, useAuthStore } from '../../auth/store'
const auth = useAuthStore()
const { query, filters, items, total, loading, error, load, search, reset } =
  useList(settings.settings)
const drawer = ref(false),
  selected = ref<Setting | null>(null)
const form = reactive({
  code: '',
  name: '',
  description: '',
  enabled: true,
  type: 'text',
  value: '',
})
const formRef = ref<FormInstance>()
const saving = ref(false),
  formError = ref('')
const editable = computed(() =>
  auth.has(selected.value ? 'system.setting.update' : 'system.setting.create'),
)
const types = [
  { value: 'text', label: '文本' },
  { value: 'integer', label: '整数' },
  { value: 'boolean', label: '布尔值' },
  { value: 'json', label: 'JSON' },
]
function typeName(type: string) {
  return types.find((item) => item.value === type)?.label || type
}
async function open(row?: Setting) {
  selected.value = row || null
  Object.assign(form, {
    code: row?.code || '',
    name: row?.name || '',
    description: row?.description || '',
    enabled: row?.enabled ?? true,
    type: row?.type || 'text',
    value: row?.value || '',
  })
  formError.value = ''
  drawer.value = true
  await nextTick()
  formRef.value?.clearValidate()
}
function changeType() {
  form.value =
    { text: '', integer: '0', boolean: 'false', json: '{}' }[form.type] || ''
  formError.value = ''
}
async function save() {
  if (
    saving.value ||
    !editable.value ||
    !(await formRef.value?.validate().catch(() => false))
  )
    return
  formError.value = validateSettingValue(form.type, form.value)
  if (formError.value) return
  saving.value = true
  try {
    const body = {
      name: form.name.trim(),
      description: form.description.trim(),
      enabled: form.enabled,
      value: form.value,
    }
    if (selected.value)
      await settings.updateSetting(selected.value.id, {
        ...body,
        revision: selected.value.revision,
      })
    else
      await settings.createSetting({
        ...body,
        code: form.code,
        type: form.type,
      })
    drawer.value = false
    ElMessage.success('配置已保存')
    await load()
  } catch (cause) {
    formError.value = authMessage(cause)
  } finally {
    saving.value = false
  }
}
async function remove(row: Setting) {
  try {
    await ElMessageBox.confirm(
      `确定删除配置“${row.name}”？此操作无法撤销。`,
      '删除系统配置',
      { type: 'warning', confirmButtonText: '删除', cancelButtonText: '取消' },
    )
  } catch {
    return
  }
  try {
    await settings.deleteSetting(row.id, row.revision)
    ElMessage.success('配置已删除')
  } catch (cause) {
    ElMessage.error(authMessage(cause))
  }
  await load()
}
watch(
  () => auth.permissions,
  () => {
    if (!auth.has('system.setting.read') || !editable.value)
      drawer.value = false
  },
)
</script>
<template>
  <div class="page-heading">
    <div>
      <h1>系统配置</h1>
      <p>维护普通业务参数。密码、令牌及连接密钥请使用部署环境配置。</p>
    </div>
  </div>
  <ListFilters
    :filters="filters"
    placeholder="配置名称 / 配置键"
    :loading="loading"
    @search="search"
    @reset="reset"
  />
  <div v-if="error" class="error-notice" role="alert">{{ error }}</div>
  <section class="panel">
    <div class="panel-heading">
      <h2>
        配置列表 <span class="count-label">{{ total }}</span>
      </h2>
      <div class="table-tools">
        <ElButton
          v-if="auth.has('system.setting.create')"
          type="primary"
          :icon="Plus"
          @click="open()"
          >新增配置</ElButton
        ><ElButton
          :icon="RefreshRight"
          :loading="loading"
          aria-label="刷新配置列表"
          @click="load"
        />
      </div>
    </div>
    <div class="panel-table management-table" :aria-busy="loading">
      <ElTable
        :data="items"
        row-key="id"
        :empty-text="loading ? '加载中…' : '暂无配置，可新增普通业务参数'"
      >
        <ElTableColumn label="配置" min-width="220"
          ><template #default="{ row }"
            ><div class="role-cell">
              <strong>{{ row.name }}</strong
              ><small>{{ row.code }}</small>
            </div></template
          ></ElTableColumn
        >
        <ElTableColumn label="类型" width="95"
          ><template #default="{ row }"
            ><ElTag type="info" effect="plain" size="small">{{
              typeName(row.type)
            }}</ElTag></template
          ></ElTableColumn
        >
        <ElTableColumn
          prop="value"
          label="配置值"
          min-width="210"
          show-overflow-tooltip
        />
        <ElTableColumn
          prop="description"
          label="说明"
          min-width="180"
          show-overflow-tooltip
        />
        <ElTableColumn label="状态" width="100"
          ><template #default="{ row }"
            ><span class="status-dot" :class="{ enabled: row.enabled }"></span
            >{{ row.enabled ? '启用' : '停用' }}</template
          ></ElTableColumn
        >
        <ElTableColumn label="操作" width="140" fixed="right"
          ><template #default="{ row }"
            ><ElButton link type="primary" @click="open(row as Setting)">{{
              auth.has('system.setting.update') ? '编辑' : '查看'
            }}</ElButton
            ><ElButton
              v-if="auth.has('system.setting.delete')"
              link
              type="danger"
              :disabled="row.enabled"
              :title="row.enabled ? '请先停用配置' : '删除配置'"
              @click="remove(row as Setting)"
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
    :model-value="drawer"
    :title="selected ? (editable ? '编辑配置' : '查看配置') : '新增配置'"
    size="min(560px, 100vw)"
    :close-on-click-modal="!saving"
    :close-on-press-escape="!saving"
    :show-close="!saving"
    destroy-on-close
    @update:model-value="drawer = false"
  >
    <p class="drawer-intro">
      配置键和值类型创建后不可修改，变更会记录到操作日志。
    </p>
    <div v-if="formError" class="error-notice" role="alert">
      {{ formError }}
    </div>
    <ElForm
      ref="formRef"
      :model="form"
      :rules="resourceRules"
      label-position="top"
      :disabled="saving || !editable"
      @submit.prevent="save"
    >
      <ElFormItem label="配置键" prop="code"
        ><ElInput
          v-model="form.code"
          :disabled="!!selected"
          maxlength="64"
          placeholder="例如 site.title"
      /></ElFormItem>
      <ElFormItem label="配置名称" prop="name"
        ><ElInput v-model="form.name" maxlength="64"
      /></ElFormItem>
      <ElFormItem label="值类型"
        ><ElSelect
          v-model="form.type"
          :disabled="!!selected"
          aria-label="值类型"
          @change="changeType"
          ><ElOption
            v-for="type in types"
            :key="type.value"
            :value="type.value"
            :label="type.label" /></ElSelect
      ></ElFormItem>
      <ElFormItem label="配置值" required>
        <ElSwitch
          v-if="form.type === 'boolean'"
          v-model="form.value"
          active-value="true"
          inactive-value="false"
          active-text="true"
          inactive-text="false"
          aria-label="配置值"
        />
        <ElInput
          v-else
          v-model="form.value"
          :type="form.type === 'integer' ? 'text' : 'textarea'"
          :inputmode="form.type === 'integer' ? 'numeric' : 'text'"
          :rows="form.type === 'json' ? 8 : 4"
          maxlength="4096"
          aria-label="配置值"
          :placeholder="
            form.type === 'json'
              ? '请输入有效的 JSON'
              : form.type === 'integer'
                ? '请输入整数，例如 100'
                : '请输入文本'
          "
          :class="{ 'code-value': form.type === 'json' }"
        />
        <div class="field-hint">
          最多 4096 字节；文本允许为空，整数支持完整的 64 位范围。
        </div>
      </ElFormItem>
      <ElFormItem label="说明" prop="description"
        ><ElInput
          v-model="form.description"
          type="textarea"
          :rows="2"
          maxlength="256"
      /></ElFormItem>
      <ElFormItem label="状态"
        ><ElSwitch
          v-model="form.enabled"
          active-text="启用"
          inactive-text="停用"
      /></ElFormItem>
    </ElForm>
    <template #footer
      ><ElButton :disabled="saving" @click="drawer = false">{{
        editable ? '取消' : '关闭'
      }}</ElButton
      ><ElButton v-if="editable" type="primary" :loading="saving" @click="save"
        >保存配置</ElButton
      ></template
    >
  </ElDrawer>
</template>
