<script setup lang="ts">
import { ElButton, ElInput, ElSelect, ElOption } from 'element-plus'
import { Search } from '@element-plus/icons-vue'
defineProps<{
  filters: { keyword: string; status: string }
  states: { value: string; label: string }[]
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
      ><ElSelect
        v-model="filters.status"
        :empty-values="[null, undefined]"
        aria-label="状态"
        ><ElOption label="全部" value="" /><ElOption
          v-for="state in states"
          :key="state.value"
          :label="state.label"
          :value="state.value" /></ElSelect
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
