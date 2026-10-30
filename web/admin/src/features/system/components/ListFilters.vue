<script setup lang="ts">
import ListFilters from '../../../components/ListFilters.vue'
import type { Query } from '../../../lib/list-query'
defineProps<{
  filters: { keyword: string; enabled: Query['enabled'] }
  placeholder: string
  loading: boolean
}>()
defineEmits<{ search: []; reset: [] }>()
</script>
<template>
  <ListFilters
    :filters="filters"
    :placeholder="placeholder"
    :loading="loading"
    @search="$emit('search')"
    @reset="$emit('reset')"
  />
</template>
