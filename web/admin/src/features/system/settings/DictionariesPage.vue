<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, reactive, ref, watch } from 'vue'
import {
  ElButton,
  ElDrawer,
  ElForm,
  ElFormItem,
  ElInput,
  ElInputNumber,
  ElSwitch,
  ElTable,
  ElTableColumn,
  ElPagination,
  ElMessage,
  ElMessageBox,
  type FormInstance,
} from 'element-plus'
import { Plus, RefreshRight, Delete } from '@element-plus/icons-vue'
import { settings, type Dictionary, type DictionaryItem } from './api'
import { resourceRules } from './form'
import { useList } from '../use-list'
import ListFilters from '../components/ListFilters.vue'
import DictionaryItemList from './DictionaryItemList.vue'
import { authMessage, useAuthStore } from '../../auth/store'
const auth = useAuthStore()
const { query, filters, items, total, loading, error, load, search, reset } =
  useList(settings.dictionaries)
const activeId = ref('')
const active = computed(
  () =>
    items.value.find((item) => item.id === activeId.value) ?? items.value[0],
)
const drawer = ref(false)
const selected = ref<Dictionary | null>(null)
const form = reactive({
  code: '',
  name: '',
  description: '',
  enabled: true,
  items: [] as (DictionaryItem & { rowKey: number })[],
})
const formRef = ref<FormInstance>()
const detailReady = ref(false)
const saving = ref(false),
  detailLoading = ref(false),
  formError = ref('')
const editable = computed(() =>
  auth.has(
    selected.value ? 'system.dictionary.update' : 'system.dictionary.create',
  ),
)
let sequence = 0,
  rowKey = 0
async function open(row?: Dictionary) {
  const attempt = ++sequence
  selected.value = row || null
  Object.assign(form, {
    code: row?.code || '',
    name: row?.name || '',
    description: row?.description || '',
    enabled: row?.enabled ?? true,
    items: [],
  })
  formError.value = ''
  drawer.value = true
  detailReady.value = !row
  detailLoading.value = !!row
  await nextTick()
  formRef.value?.clearValidate()
  if (!row) return
  try {
    const data = await settings.dictionary(row.id)
    if (attempt !== sequence) return
    detailReady.value = true
    selected.value = data.record
    Object.assign(form, data.record, {
      items: data.items.map((item) => ({ ...item, rowKey: ++rowKey })),
    })
  } catch (cause) {
    if (attempt === sequence) formError.value = authMessage(cause)
  } finally {
    if (attempt === sequence) detailLoading.value = false
  }
}
function addItem() {
  if (form.items.length < 200)
    form.items.push({
      rowKey: ++rowKey,
      value: '',
      label: '',
      sort: form.items.length * 10,
      enabled: true,
    })
}
async function save() {
  if (
    saving.value ||
    !detailReady.value ||
    detailLoading.value ||
    !editable.value ||
    !(await formRef.value?.validate().catch(() => false))
  )
    return
  const seen = new Set<string>()
  for (const [i, item] of form.items.entries()) {
    if (
      !item.value.trim() ||
      item.value !== item.value.trim() ||
      !item.label.trim() ||
      seen.has(item.value) ||
      !Number.isInteger(item.sort) ||
      item.sort < 0 ||
      item.sort > 9999
    ) {
      formError.value = `请检查第 ${i + 1} 项：值与标签必填、值唯一且无首尾空格，排序为 0–9999 的整数。`
      return
    }
    seen.add(item.value)
  }
  saving.value = true
  formError.value = ''
  try {
    const body = {
      name: form.name.trim(),
      description: form.description.trim(),
      enabled: form.enabled,
      items: form.items.map(({ value, label, sort, enabled }) => ({
        value,
        label,
        sort,
        enabled,
      })),
    }
    if (selected.value)
      await settings.updateDictionary(selected.value.id, {
        ...body,
        revision: selected.value.revision,
      })
    else {
      const created = await settings.createDictionary({
        ...body,
        code: form.code,
      })
      activeId.value = created.record.id
      query.page = 1
    }
    drawer.value = false
    ElMessage.success('字典已保存')
    await load()
  } catch (cause) {
    formError.value = authMessage(cause)
  } finally {
    saving.value = false
  }
}
async function remove(row: Dictionary) {
  try {
    await ElMessageBox.confirm(
      `删除“${row.name}”及其全部字典项？此操作无法撤销。`,
      '删除字典',
      { type: 'warning', confirmButtonText: '删除', cancelButtonText: '取消' },
    )
  } catch {
    return
  }
  try {
    await settings.deleteDictionary(row.id, row.revision)
    ElMessage.success('字典已删除')
  } catch (cause) {
    ElMessage.error(authMessage(cause))
  }
  await load()
}
watch(drawer, (value) => {
  if (!value) {
    sequence++
    form.items = []
    detailLoading.value = false
  }
})
watch(
  () => auth.permissions,
  () => {
    if (!auth.has('system.dictionary.read') || !editable.value)
      drawer.value = false
  },
)
onBeforeUnmount(() => {
  sequence++
})
</script>
<template>
  <div class="page-heading">
    <div>
      <h1>字典管理</h1>
      <p>按分类查看和维护业务选项。</p>
    </div>
    <div class="table-tools">
      <ElButton
        v-if="auth.has('system.dictionary.create')"
        type="primary"
        :icon="Plus"
        @click="open()"
        >新增字典</ElButton
      >
      <ElButton
        :icon="RefreshRight"
        :loading="loading"
        aria-label="刷新字典列表"
        @click="load"
      />
    </div>
  </div>
  <ListFilters
    :filters="filters"
    placeholder="字典名称 / 标识"
    :loading="loading"
    @search="search"
    @reset="reset"
  />
  <div v-if="error" class="error-notice" role="alert">{{ error }}</div>
  <div
    v-else-if="active && auth.has('system.dictionary.read')"
    class="dictionary-browser"
  >
    <section class="panel">
      <div class="panel-heading">
        <h2>
          字典分类 <span class="count-label">{{ total }}</span>
        </h2>
      </div>
      <nav class="dictionary-nav" aria-label="字典分类" :aria-busy="loading">
        <button
          v-for="row in items"
          :key="row.id"
          type="button"
          :aria-pressed="row.id === active.id"
          @click="activeId = row.id"
        >
          <span class="category-text"
            ><strong>{{ row.name }}</strong
            ><small>{{ row.code }}</small></span
          >
          <span class="item-count">{{ row.itemCount }} 项</span>
        </button>
      </nav>
      <ElPagination
        v-if="total > 10"
        v-model:current-page="query.page"
        v-model:page-size="query.pageSize"
        :page-sizes="[10, 20, 50]"
        :total="total"
        size="small"
        layout="sizes, prev, slot, next"
        class="category-pager"
        @change="load"
        >{{ query.page }} /
        {{ Math.ceil(total / query.pageSize) }}</ElPagination
      >
    </section>
    <section class="panel dictionary-detail" aria-label="字典项详情">
      <div class="detail-heading">
        <div class="detail-title">
          <h2>{{ active.name }}</h2>
          <small>{{ active.code }}</small>
        </div>
        <div class="table-tools">
          <ElButton
            v-if="auth.has('system.dictionary.update')"
            size="small"
            @click="open(active)"
            >编辑字典</ElButton
          >
          <ElButton
            v-if="auth.has('system.dictionary.delete')"
            link
            type="danger"
            :disabled="active.enabled"
            :title="active.enabled ? '请先停用字典' : '删除字典'"
            @click="remove(active)"
            >删除</ElButton
          >
        </div>
      </div>
      <div class="detail-meta">
        <span
          ><span class="status-dot" :class="{ enabled: active.enabled }"></span
          >{{ active.enabled ? '启用' : '停用' }}</span
        >
        <span v-if="active.description">{{ active.description }}</span>
      </div>
      <DictionaryItemList :record="active" />
    </section>
  </div>
  <div v-else class="panel plain-empty" role="status">
    {{
      loading
        ? '正在加载字典…'
        : query.keyword || query.enabled !== 'all'
          ? '没有符合条件的字典'
          : '暂无字典'
    }}
  </div>
  <ElDrawer
    :model-value="drawer"
    :title="selected ? (editable ? '编辑字典' : '查看字典') : '新增字典'"
    size="min(840px, 100vw)"
    :close-on-click-modal="!saving"
    :close-on-press-escape="!saving"
    :show-close="!saving"
    destroy-on-close
    @update:model-value="drawer = false"
  >
    <p class="drawer-intro">
      字典标识创建后不可修改；字典项按排序值从小到大排列。
    </p>
    <div v-if="formError" class="error-notice" role="alert">
      {{ formError }}
    </div>
    <p v-if="detailLoading" role="status">正在加载字典…</p>
    <ElForm
      ref="formRef"
      :model="form"
      :rules="resourceRules"
      label-position="top"
      :disabled="saving || detailLoading || !editable"
      @submit.prevent="save"
    >
      <div class="settings-form-grid">
        <ElFormItem label="字典标识" prop="code"
          ><ElInput
            v-model="form.code"
            :disabled="!!selected"
            maxlength="64"
            placeholder="例如 order.status" /></ElFormItem
        ><ElFormItem label="字典名称" prop="name"
          ><ElInput v-model="form.name" maxlength="64"
        /></ElFormItem>
      </div>
      <ElFormItem label="说明" prop="description"
        ><ElInput
          v-model="form.description"
          maxlength="256"
          type="textarea"
          :rows="2"
      /></ElFormItem>
      <ElFormItem label="字典状态"
        ><ElSwitch
          v-model="form.enabled"
          active-text="启用"
          inactive-text="停用"
      /></ElFormItem>
      <div class="dictionary-items-heading">
        <h3>
          字典项 <span class="count-label">{{ form.items.length }} / 200</span>
        </h3>
        <ElButton
          v-if="editable"
          :icon="Plus"
          :disabled="form.items.length >= 200"
          @click="addItem"
          >添加字典项</ElButton
        >
      </div>
      <ElTable
        :data="form.items"
        row-key="rowKey"
        class="dictionary-items"
        empty-text="暂无字典项，可点击右上方添加"
      >
        <ElTableColumn label="字典值" min-width="155"
          ><template #default="{ row, $index }"
            ><ElInput
              v-model="row.value"
              maxlength="64"
              :aria-label="`第 ${$index + 1} 项字典值`"
              placeholder="例如 pending" /></template
        ></ElTableColumn>
        <ElTableColumn label="显示标签" min-width="155"
          ><template #default="{ row, $index }"
            ><ElInput
              v-model="row.label"
              maxlength="64"
              :aria-label="`第 ${$index + 1} 项显示标签`"
              placeholder="例如 待处理" /></template
        ></ElTableColumn>
        <ElTableColumn label="排序" width="130"
          ><template #default="{ row, $index }"
            ><ElInputNumber
              v-model="row.sort"
              :min="0"
              :max="9999"
              :precision="0"
              controls-position="right"
              :aria-label="`第 ${$index + 1} 项排序`" /></template
        ></ElTableColumn>
        <ElTableColumn label="启用" width="70"
          ><template #default="{ row, $index }"
            ><ElSwitch
              v-model="row.enabled"
              :aria-label="`第 ${$index + 1} 项启用`" /></template
        ></ElTableColumn>
        <ElTableColumn v-if="editable" label="操作" width="60" fixed="right"
          ><template #default="{ $index }"
            ><ElButton
              link
              type="danger"
              :icon="Delete"
              :aria-label="`移除第 ${$index + 1} 个字典项`"
              @click="form.items.splice($index, 1)" /></template
        ></ElTableColumn>
      </ElTable>
    </ElForm>
    <template #footer
      ><ElButton :disabled="saving" @click="drawer = false">{{
        editable ? '取消' : '关闭'
      }}</ElButton
      ><ElButton
        v-if="editable"
        type="primary"
        :loading="saving"
        :disabled="detailLoading || !detailReady"
        @click="save"
        >保存字典</ElButton
      ></template
    >
  </ElDrawer>
</template>
<style scoped>
.dictionary-browser {
  display: grid;
  grid-template-columns: minmax(228px, 28%) minmax(0, 1fr);
  align-items: start;
  gap: 16px;
}
.dictionary-nav {
  display: grid;
  max-height: 55vh;
  overflow-y: auto;
  padding: 8px;
}
.dictionary-nav button {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  padding: 10px 12px;
  border: 0;
  border-radius: 4px;
  background: transparent;
  color: inherit;
  font: inherit;
  text-align: left;
  cursor: pointer;
}
.dictionary-nav button:hover {
  background: var(--el-fill-color-light);
}
.dictionary-nav button[aria-pressed='true'] {
  background: var(--el-color-primary-light-9);
  color: var(--el-color-primary);
}
.category-text,
.detail-title {
  min-width: 0;
  overflow-wrap: anywhere;
}
.category-text strong,
.category-text small,
.detail-title small {
  display: block;
}
.category-text strong {
  font-size: 14px;
  font-weight: 500;
}
.category-text small,
.detail-title small,
.item-count {
  color: var(--el-text-color-secondary);
  font-size: 12px;
}
.item-count {
  white-space: nowrap;
}
.category-pager {
  justify-content: center;
  padding: 8px 4px;
  border-top: 1px solid var(--el-border-color-lighter);
}
.detail-heading,
.detail-meta {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 8px 16px;
}
.detail-heading {
  justify-content: space-between;
  padding: 16px 20px 0;
}
.detail-title h2 {
  font-size: 16px;
  font-weight: 600;
}
.detail-meta {
  padding: 8px 20px 16px;
  color: var(--el-text-color-secondary);
  font-size: 13px;
  overflow-wrap: anywhere;
}
.dictionary-detail {
  padding-bottom: 12px;
}
@media (max-width: 850px) {
  .dictionary-browser {
    grid-template-columns: minmax(0, 1fr);
  }
  .dictionary-nav {
    max-height: 200px;
  }
}
</style>
