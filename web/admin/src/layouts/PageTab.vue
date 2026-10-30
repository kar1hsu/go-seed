<script setup lang="ts">
import { ref, type Component } from 'vue'
import { ElDropdown, type DropdownInstance } from 'element-plus'
import { Close } from '@element-plus/icons-vue'
import TabActions from './TabActions.vue'
defineProps<{
  path: string
  title: string
  icon: Component
  active: boolean
  hasOthers: boolean
  hasRight: boolean
  hasClosable: boolean
}>()
const emit = defineEmits<{ action: [command: string] }>()
const dropdown = ref<DropdownInstance>()
function keyboardMenu(event: KeyboardEvent) {
  if (event.key === 'ContextMenu' || (event.shiftKey && event.key === 'F10')) {
    event.preventDefault()
    dropdown.value?.handleOpen()
  }
}
</script>
<template>
  <ElDropdown
    ref="dropdown"
    trigger="contextmenu"
    @command="(command) => emit('action', command)"
  >
    <div class="page-tab" :class="{ active }" @keydown="keyboardMenu">
      <RouterLink :to="path" :aria-current="active ? 'page' : undefined"
        ><component :is="icon" class="small-icon" />{{ title }}</RouterLink
      >
      <button
        v-if="path !== '/'"
        class="tab-close"
        :aria-label="`关闭${title}页签`"
        @click="emit('action', 'close')"
      >
        <Close class="tiny-icon" />
      </button>
    </div>
    <template #dropdown
      ><TabActions
        :pinned="path === '/'"
        :has-others="hasOthers"
        :has-right="hasRight"
        :has-closable="hasClosable"
    /></template>
  </ElDropdown>
</template>
