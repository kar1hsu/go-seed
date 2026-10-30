<script setup lang="ts">
import { ElDropdownItem, ElDropdownMenu } from 'element-plus'
defineProps<{
  pinned: boolean
  hasOthers: boolean
  hasRight: boolean
  hasClosable: boolean
}>()
</script>

<template>
  <ElDropdownMenu>
    <ElDropdownItem command="refresh">刷新页面</ElDropdownItem>
    <ElDropdownItem command="close" :disabled="pinned" divided
      >关闭当前标签</ElDropdownItem
    >
    <ElDropdownItem command="others" :disabled="!hasOthers"
      >关闭其他标签</ElDropdownItem
    >
    <ElDropdownItem command="right" :disabled="!hasRight"
      >关闭右侧标签</ElDropdownItem
    >
    <ElDropdownItem command="all" :disabled="!hasClosable" divided
      >关闭全部标签</ElDropdownItem
    >
  </ElDropdownMenu>
</template>
