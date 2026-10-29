<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import {
  Fold,
  Expand,
  ArrowRight,
  Close,
  Link,
  Document,
  RefreshRight,
} from '@element-plus/icons-vue'
import { navigation } from '../router/navigation'
import { project } from '../app/project'
import { useDashboardStore } from '../features/dashboard/store'

const route = useRoute()
const router = useRouter()
const status = useDashboardStore()
const collapsed = ref(false)
const mobileOpen = ref(false)
const mobile = ref(false)
const visited = ref(['/'])
const current = computed(
  () => navigation.find((item) => item.path === route.path) || navigation[0]!,
)
const tabs = computed(() =>
  visited.value
    .map((path) => navigation.find((item) => item.path === path)!)
    .filter(Boolean),
)
const media = window.matchMedia('(max-width: 760px)')
function updateMedia() {
  mobile.value = media.matches
  mobileOpen.value = false
}
function toggleSidebar() {
  if (mobile.value) mobileOpen.value = !mobileOpen.value
  else collapsed.value = !collapsed.value
}
function closeTab(path: string) {
  visited.value = visited.value.filter((item) => item !== path)
  if (route.path === path) void router.push(visited.value.at(-1) || '/')
}
watch(
  () => route.path,
  (path) => {
    if (!visited.value.includes(path)) visited.value.push(path)
    mobileOpen.value = false
    document.title = `${current.value.title} - GoSeed`
  },
  { immediate: true },
)
onMounted(() => {
  updateMedia()
  media.addEventListener('change', updateMedia)
})
onBeforeUnmount(() => media.removeEventListener('change', updateMedia))
</script>

<template>
  <div
    class="admin-shell"
    :class="{ 'is-collapsed': collapsed, 'mobile-open': mobileOpen }"
    @keydown.esc="mobileOpen = false"
  >
    <a class="skip-link" href="#main-content">跳到内容</a>
    <button
      v-if="mobileOpen"
      class="sidebar-mask"
      aria-label="关闭导航菜单"
      @click="mobileOpen = false"
    ></button>
    <aside class="app-sidebar" aria-label="侧边导航">
      <RouterLink class="app-brand" to="/" aria-label="GoSeed 工作台"
        ><img src="/seed.svg" alt="" /><strong>GoSeed</strong
        ><span>Admin</span></RouterLink
      >
      <nav class="side-menu" aria-label="主导航">
        <template v-for="group in ['概览', '开发工具']" :key="group">
          <div class="menu-group">{{ group }}</div>
          <RouterLink
            v-for="item in navigation.filter((item) => item.group === group)"
            :key="item.path"
            :to="item.path"
            class="menu-item"
            :class="{ active: route.path === item.path }"
            :aria-current="route.path === item.path ? 'page' : undefined"
            :title="item.title"
          >
            <component :is="item.icon" class="icon" /><span>{{
              item.title
            }}</span>
          </RouterLink>
        </template>
      </nav>
      <div class="sidebar-bottom">
        <span>GoSeed v{{ project.version }}</span
        ><button
          class="icon-button"
          :aria-label="collapsed ? '展开导航' : '收起导航'"
          @click="toggleSidebar"
        >
          <component :is="collapsed ? Expand : Fold" class="icon" />
        </button>
      </div>
    </aside>
    <div class="app-main">
      <div class="app-chrome">
        <header class="app-header">
          <div class="header-left">
            <button
              class="icon-button"
              aria-label="切换侧边栏"
              :aria-expanded="mobile ? mobileOpen : !collapsed"
              @click="toggleSidebar"
            >
              <component :is="collapsed ? Expand : Fold" class="icon" />
            </button>
            <nav class="breadcrumb" aria-label="面包屑">
              <span>{{ current.group }}</span
              ><ArrowRight class="tiny-icon" /><span>{{ current.title }}</span>
            </nav>
          </div>
          <div class="header-tools">
            <span class="environment-label"><i></i>开发环境</span
            ><a
              :href="project.repository"
              target="_blank"
              rel="noopener noreferrer"
              class="icon-button"
              aria-label="打开项目仓库"
              title="项目仓库"
              ><Link class="icon" /></a
            ><RouterLink
              to="/api-docs"
              class="icon-button"
              title="接口文档"
              aria-label="打开接口文档"
              ><Document class="icon" /></RouterLink
            ><span class="header-divider"></span
            ><span class="dev-avatar" title="当前未接入登录">G</span
            ><span class="workspace-name">本地工作空间</span>
          </div>
        </header>
        <div class="tabs-bar">
          <nav class="page-tabs" aria-label="已打开页面">
            <div
              v-for="item in tabs"
              :key="item.path"
              class="page-tab"
              :class="{ active: route.path === item.path }"
            >
              <RouterLink
                :to="item.path"
                :aria-current="route.path === item.path ? 'page' : undefined"
                ><component :is="item.icon" class="small-icon" />{{
                  item.title
                }}</RouterLink
              ><button
                v-if="item.path !== '/'"
                class="tab-close"
                :aria-label="`关闭${item.title}页签`"
                @click="closeTab(item.path)"
              >
                <Close class="tiny-icon" />
              </button>
            </div>
          </nav>
          <button
            class="icon-button tab-refresh"
            aria-label="刷新服务数据"
            title="刷新服务数据"
            :disabled="status.loading"
            @click="status.refresh"
          >
            <RefreshRight class="icon" :class="{ spinning: status.loading }" />
          </button>
        </div>
      </div>
      <main id="main-content" class="page-content"><RouterView /></main>
      <footer class="app-footer">
        GoSeed Admin <span>v{{ project.version }}</span>
      </footer>
    </div>
  </div>
</template>
