<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import {
  Fold,
  Expand,
  ArrowRight,
  MoreFilled,
  Link,
  Document,
  RefreshRight,
} from '@element-plus/icons-vue'
import {
  ElDropdown,
  ElDropdownMenu,
  ElDropdownItem,
  ElMessage,
} from 'element-plus'
import { useAuthStore, authMessage } from '../features/auth/store'
import { navigation } from '../router/navigation'
import SidebarNavigation from './SidebarNavigation.vue'
import PageTab from './PageTab.vue'
import TabActions from './TabActions.vue'
import { closeTabs, type CloseAction } from './tabs'
import { project } from '../app/project'
import { useDashboardStore } from '../features/dashboard/store'

const auth = useAuthStore()
const signingOut = ref(false)
const route = useRoute()
const router = useRouter()
const status = useDashboardStore()
const collapsed = ref(false)
const mobileOpen = ref(false)
const mobile = ref(false)
const visited = ref<string[]>(['/'])
const pageVersion = ref(0)
const tabsElement = ref<HTMLElement>()
try {
  collapsed.value = localStorage.getItem('goseed.menu.collapsed') === 'true'
} catch {
  /* Optional preference. */
}
watch(collapsed, (value) => {
  try {
    localStorage.setItem('goseed.menu.collapsed', String(value))
  } catch {
    /* Optional preference. */
  }
})
async function userAction(command: string) {
  if (command === 'security') {
    await router.push('/account/security')
    return
  }
  if (signingOut.value) return
  signingOut.value = true
  try {
    await auth.logout()
    status.clearRecords()
    await router.replace('/login')
  } catch (error) {
    ElMessage.error(authMessage(error))
  } finally {
    signingOut.value = false
  }
}
const current = computed(
  () =>
    navigation.find((item) => item.path === route.path) ||
    (route.name === 'forbidden'
      ? { ...navigation[0]!, title: '暂无访问权限', group: '账户' }
      : navigation[0]!),
)
const tabs = computed(() =>
  visited.value
    .map((path) => navigation.find((item) => item.path === path)!)
    .filter((item) => item && auth.has(item.permission)),
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
function tabOptions(path: string) {
  const paths: string[] = tabs.value.map((item) => item.path)
  return {
    pinned: path === '/',
    hasOthers: paths.some((item) => item !== '/' && item !== path),
    hasRight: paths.slice(paths.indexOf(path) + 1).some((item) => item !== '/'),
    hasClosable: paths.some((item) => item !== '/'),
  }
}
async function tabAction(path: string, command: string) {
  if (command === 'refresh') {
    if (route.path === path) pageVersion.value++
    else await router.push(path)
    return
  }
  if (!['close', 'others', 'right', 'all'].includes(command)) return
  const result = closeTabs(
    tabs.value.map((item) => item.path),
    route.path,
    path,
    command as CloseAction,
  )
  visited.value = result.paths
  if (result.active !== route.path) await router.push(result.active)
}
watch([() => route.path, tabs], async () => {
  await nextTick()
  tabsElement.value
    ?.querySelector('[aria-current="page"]')
    ?.scrollIntoView({ block: 'nearest', inline: 'nearest' })
})
watch(
  () => route.path,
  (path) => {
    if (
      navigation.some((item) => item.path === path) &&
      !visited.value.includes(path)
    )
      visited.value.push(path)
    mobileOpen.value = false
    document.title = `${current.value.title} - GoSeed`
  },
  { immediate: true },
)
watch(
  () => auth.permissions,
  () => {
    visited.value = visited.value.filter((path) =>
      navigation.some(
        (item) => item.path === path && auth.has(item.permission),
      ),
    )
    if (
      auth.admin &&
      typeof route.meta.permission === 'string' &&
      !auth.has(route.meta.permission)
    )
      void router.replace('/forbidden')
  },
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
    <aside
      class="app-sidebar"
      aria-label="侧边导航"
      :inert="mobile && !mobileOpen"
    >
      <RouterLink class="app-brand" to="/" aria-label="GoSeed 工作台"
        ><img src="/seed.svg" alt="" /><strong>GoSeed</strong
        ><span>Admin</span></RouterLink
      >
      <SidebarNavigation :collapsed="collapsed && !mobile" />
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
            ><span class="header-divider"></span>
            <ElDropdown trigger="click" @command="userAction">
              <button
                class="account-trigger"
                aria-label="账户菜单"
                :disabled="signingOut"
              >
                <span class="dev-avatar">{{
                  auth.admin?.displayName.slice(0, 1) || 'G'
                }}</span
                ><span class="workspace-name">{{
                  auth.admin?.displayName
                }}</span>
              </button>
              <template #dropdown
                ><ElDropdownMenu
                  ><ElDropdownItem command="security">账户安全</ElDropdownItem
                  ><ElDropdownItem command="logout" divided
                    >退出登录</ElDropdownItem
                  ></ElDropdownMenu
                ></template
              >
            </ElDropdown>
          </div>
        </header>
        <div class="tabs-bar">
          <nav ref="tabsElement" class="page-tabs" aria-label="已打开页面">
            <PageTab
              v-for="item in tabs"
              :key="item.path"
              :path="item.path"
              :title="item.title"
              :icon="item.icon"
              :active="route.path === item.path"
              v-bind="tabOptions(item.path)"
              @action="(command) => tabAction(item.path, command)"
            />
          </nav>
          <div class="tab-tools">
            <button
              class="icon-button tab-refresh"
              aria-label="刷新当前页面"
              title="刷新当前页面"
              @click="pageVersion++"
            >
              <RefreshRight class="icon" />
            </button>
            <ElDropdown
              trigger="click"
              :disabled="!tabs.some((item) => item.path === route.path)"
              @command="(command) => tabAction(route.path, command)"
            >
              <button
                class="icon-button"
                aria-label="页面标签操作"
                title="页面标签操作"
              >
                <MoreFilled class="icon" />
              </button>
              <template #dropdown
                ><TabActions v-bind="tabOptions(route.path)"
              /></template>
            </ElDropdown>
          </div>
        </div>
      </div>
      <main id="main-content" class="page-content">
        <RouterView v-slot="{ Component }"
          ><component
            :is="Component"
            :key="`${route.path}:${JSON.stringify(route.query)}:${pageVersion}`"
        /></RouterView>
      </main>
      <footer class="app-footer">
        GoSeed Admin <span>v{{ project.version }}</span>
      </footer>
    </div>
  </div>
</template>
