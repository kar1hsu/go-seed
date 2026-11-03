<script setup lang="ts">
import { ref, watch } from 'vue'
import { ElButton, ElTable, ElTableColumn } from 'element-plus'
import { authMessage } from '../../auth/store'
import { settings, type Dictionary, type DictionaryItem } from './api'

const props = defineProps<{ record: Dictionary }>()
const items = ref<DictionaryItem[]>([])
const loading = ref(false)
const error = ref('')
const retry = ref(0)

watch(
  [() => props.record, retry],
  async ([record], _, onCleanup) => {
    let current = true
    onCleanup(() => {
      current = false
    })
    items.value = []
    error.value = ''
    loading.value = true
    try {
      const detail = await settings.dictionary(record.id)
      if (current) items.value = detail.items
    } catch (cause) {
      if (current) error.value = authMessage(cause)
    } finally {
      if (current) loading.value = false
    }
  },
  { immediate: true },
)
</script>

<template>
  <div class="panel-table management-table" :aria-busy="loading">
    <div v-if="error" class="error-notice" role="alert">
      {{ error }} <ElButton link type="primary" @click="retry++">重试</ElButton>
    </div>
    <ElTable
      v-else
      :data="items"
      row-key="value"
      :empty-text="loading ? '正在加载字典项…' : '此分类暂无字典项'"
    >
      <ElTableColumn
        prop="label"
        label="显示标签"
        min-width="120"
        show-overflow-tooltip
      />
      <ElTableColumn label="字典值" min-width="120" show-overflow-tooltip>
        <template #default="{ row }"
          ><code>{{ row.value }}</code></template
        >
      </ElTableColumn>
      <ElTableColumn prop="sort" label="排序" width="68" />
      <ElTableColumn label="状态" width="82">
        <template #default="{ row }">
          <span class="status-dot" :class="{ enabled: row.enabled }"></span
          >{{ row.enabled ? '启用' : '停用' }}
        </template>
      </ElTableColumn>
    </ElTable>
  </div>
</template>
