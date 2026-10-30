<script setup lang="ts">
import { computed, nextTick, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import {
  ElMenu,
  ElSubMenu,
  ElMenuItem,
  ElIcon,
  ElScrollbar,
  type MenuInstance,
} from 'element-plus'
import { useAuthStore } from '../features/auth/store'
import { navigation } from '../router/navigation'

const props = defineProps<{ collapsed: boolean }>()
const auth = useAuthStore()
const route = useRoute()
const router = useRouter()
const menu = ref<MenuInstance>()
const groups = computed(() => {
  const visible = navigation.filter(
    (item) => !item.hidden && auth.has(item.permission),
  )
  return [...new Set(visible.map((item) => item.group))].map((name) => ({
    name,
    items: visible.filter((item) => item.group === name),
  }))
})
function savedGroups(): string[] {
  try {
    const value: unknown = JSON.parse(
      localStorage.getItem('goseed.menu.open') || 'null',
    )
    if (Array.isArray(value) && value.every((item) => typeof item === 'string'))
      return value
  } catch {
    /* Storage can be unavailable in private browsing. */
  }
  return groups.value.map((group) => group.name)
}
const opened = ref(savedGroups())
function change(group: string, open: boolean) {
  if (props.collapsed) return
  opened.value = open
    ? [...new Set([...opened.value, group])]
    : opened.value.filter((item) => item !== group)
  try {
    localStorage.setItem('goseed.menu.open', JSON.stringify(opened.value))
  } catch {
    /* Optional preference. */
  }
}
watch(
  [() => route.path, () => props.collapsed, groups],
  async () => {
    if (props.collapsed) return
    const group = groups.value.find((group) =>
      group.items.some((item) => item.path === route.path),
    )
    if (group) change(group.name, true)
    await nextTick()
    if (props.collapsed) return
    // Preferences may contain groups hidden by this account's current permissions.
    for (const visible of groups.value) {
      if (opened.value.includes(visible.name)) menu.value?.open(visible.name)
    }
  },
  { immediate: true },
)
</script>

<template>
  <nav class="side-menu" aria-label="主导航">
    <ElScrollbar wrap-class="side-menu-scroll" view-class="side-menu-content">
      <ElMenu
        ref="menu"
        :default-active="route.path"
        :default-openeds="opened"
        :collapse="collapsed"
        :collapse-transition="false"
        @open="(group) => change(group, true)"
        @close="(group) => change(group, false)"
        @select="(path) => router.push(path)"
      >
        <ElSubMenu
          v-for="group in groups"
          :key="group.name"
          :index="group.name"
          :title="collapsed ? group.name : undefined"
          popper-class="sidebar-submenu"
        >
          <template #title
            ><ElIcon><component :is="group.items[0]!.icon" /></ElIcon
            ><span>{{ group.name }}</span></template
          >
          <ElMenuItem
            v-for="item in group.items"
            :key="item.path"
            :index="item.path"
            :aria-current="route.path === item.path ? 'page' : undefined"
          >
            <ElIcon><component :is="item.icon" /></ElIcon
            ><span>{{ item.title }}</span>
          </ElMenuItem>
        </ElSubMenu>
      </ElMenu>
    </ElScrollbar>
  </nav>
</template>
