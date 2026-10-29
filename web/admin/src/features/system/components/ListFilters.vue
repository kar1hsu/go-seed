<script setup lang="ts">
import { ElInput, ElSelect, ElOption, ElButton } from 'element-plus'
import { Search } from '@element-plus/icons-vue'
import type { Query } from '../api'
defineProps<{
  filters: { keyword: string; enabled: Query['enabled'] }
  placeholder: string
  loading: boolean
}>()
defineEmits<{ search: []; reset: [] }>()
</script>
<template>
  <form
    class="panel filter-panel management-filters"
    @submit.prevent="$emit('search')"
  >
    <label
      ><span>关键词</span
      ><ElInput
        v-model="filters.keyword"
        :placeholder="placeholder"
        aria-label="关键词"
        maxlength="64"
        clearable
    /></label>
    <label
      ><span>状态</span
      ><ElSelect v-model="filters.enabled" aria-label="状态"
        ><ElOption label="全部" value="all" /><ElOption
          label="启用"
          value="enabled" /><ElOption label="停用" value="disabled" /></ElSelect
    ></label>
    <div class="filter-actions">
      <ElButton
        type="primary"
        native-type="submit"
        :icon="Search"
        :loading="loading"
        >查询</ElButton
      ><ElButton @click="$emit('reset')">重置</ElButton>
    </div>
  </form>
</template>
