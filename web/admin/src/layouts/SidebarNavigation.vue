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
import type { PresentationGroup } from '../router/menu-layout'

const props = defineProps<{ collapsed: boolean; groups: PresentationGroup[] }>()
const route = useRoute()
const router = useRouter()
const menu = ref<MenuInstance>()
const groups = computed(() =>
  props.groups
    .filter((g) => !g.hidden)
    .map((g) => ({ ...g, items: g.items.filter((p) => !p.hidden) }))
    .filter((g) => g.items.length),
)
function savedGroups(): string[] {
  try {
    const value: unknown = JSON.parse(
      localStorage.getItem('goseed.menu.open') || 'null',
    )
    if (Array.isArray(value) && value.every((item) => typeof item === 'string'))
      return value.map(
        (id) =>
          groups.value.find(
            (group) => group.id === id || group.id === 'builtin-' + id,
          )?.id ?? id,
      )
  } catch {
    /* Storage can be unavailable in private browsing. */
  }
  return groups.value.map((group) => group.id)
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
    if (group) change(group.id, true)
    await nextTick()
    if (props.collapsed) return
    // Preferences may contain groups hidden by this account's current permissions.
    for (const visible of groups.value) {
      if (opened.value.includes(visible.id)) menu.value?.open(visible.id)
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
          :key="group.id"
          :index="group.id"
          :title="collapsed ? group.title : undefined"
          popper-class="sidebar-submenu"
        >
          <template #title
            ><ElIcon><component :is="group.icon" /></ElIcon
            ><span>{{ group.title }}</span></template
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
