import { defineStore } from 'pinia'
import { ref } from 'vue'
import { randomKey } from '../../lib/random'
import { api, APIError } from '../../lib/http'
import type { components } from '../../api/generated/schema'

const csrf = { header: { 'X-GoSeed-CSRF': '1' as const } }

type Admin = components['schemas']['AdminProfile']
type AuthData = components['schemas']['AuthData']

// Serializes cookie rotation across tabs without putting tokens in browser storage.
async function cookieOperation<T>(operation: () => Promise<T>): Promise<T> {
  if (navigator.locks)
    return navigator.locks.request('goseed-admin-cookie', operation)
  return operation()
}

export function authMessage(error: unknown): string {
  if (error instanceof APIError) {
    const messages: Record<number, string> = {
      100400: '输入信息不符合要求，请检查后重试',
      100401: '账号或密码不正确，或登录已失效',
      100403: '没有此操作的权限，或目标超出您的管理范围',
      100404: '记录不存在，请刷新列表',
      100409: '数据已变更、名称重复或当前状态不允许此操作，请刷新后重试',
      100429: '操作过于频繁，请稍后再试',
      100503: '服务暂时不可用，请稍后重试',
      100504: '服务响应超时，请重试',
    }
    return messages[error.code || 0] || '操作失败，请稍后重试'
  }
  return '无法连接服务，请检查网络后重试'
}

export const useAuthStore = defineStore('auth', () => {
  const admin = ref<Admin | null>(null)
  const permissions = ref<string[]>([])
  const roles = ref<string[]>([])
  const policyVersion = ref('')
  const superAdmin = ref(false)
  const has = (permission?: string) =>
    !permission || permissions.value.includes(permission)
  const accessToken = ref('')
  const expiresAt = ref(0)
  const initialized = ref(false)
  const connectionError = ref('')
  let sessionEpoch = 0
  let accessDeadline = 0
  let retryKey = ''
  try {
    retryKey = sessionStorage.getItem('goseed.refresh.retry') || ''
  } catch {}
  function rememberRetry(key: string) {
    retryKey = key
    try {
      if (key) sessionStorage.setItem('goseed.refresh.retry', key)
      else sessionStorage.removeItem('goseed.refresh.retry')
    } catch {}
  }
  let refreshPromise: Promise<boolean> | null = null
  let channel: BroadcastChannel | null = null

  function clear() {
    sessionEpoch++
    admin.value = null
    permissions.value = []
    roles.value = []
    policyVersion.value = ''
    superAdmin.value = false
    accessToken.value = ''
    expiresAt.value = 0
    accessDeadline = 0
    rememberRetry('')
  }
  function setSession(data: AuthData, response: Response, started: number) {
    admin.value = data.admin
    accessToken.value = data.accessToken
    expiresAt.value = Date.parse(data.expiresAt)
    const serverTime = Date.parse(response.headers.get('Date') || '')
    const remaining =
      expiresAt.value - (Number.isFinite(serverTime) ? serverTime : Date.now())
    accessDeadline =
      performance.now() +
      Math.max(0, remaining - (performance.now() - started) - 1000)
    initialized.value = true
    connectionError.value = ''
  }
  function sync(onLogout: () => void) {
    if (typeof BroadcastChannel !== 'undefined')
      channel = new BroadcastChannel('goseed-admin-session')
    if (channel)
      channel.onmessage = (event) => {
        if (event.data === 'permissions') void reconcile()
        if (event.data === 'logout') {
          clear()
          initialized.value = true
          onLogout()
        }
      }
    async function reconcile() {
      if (!admin.value || document.visibilityState !== 'visible') return
      try {
        await current()
      } catch (error) {
        if (error instanceof APIError && error.code === 100401) {
          clear()
          onLogout()
        }
      }
    }
    const timer = window.setInterval(reconcile, 30_000)
    window.addEventListener('focus', reconcile)
    document.addEventListener('visibilitychange', reconcile)
    return () => {
      window.clearInterval(timer)
      window.removeEventListener('focus', reconcile)
      document.removeEventListener('visibilitychange', reconcile)
      channel?.close()
      channel = null
    }
  }
  function signedOut() {
    clear()
    channel?.postMessage('logout')
  }

  async function refresh(): Promise<boolean> {
    if (refreshPromise) return refreshPromise
    const epoch = sessionEpoch
    refreshPromise = cookieOperation(async () => {
      if (epoch !== sessionEpoch) return false
      if (!retryKey) rememberRetry(randomKey())
      const started = performance.now()
      const { data, error, response } = await api.POST(
        '/admin-api/v1/auth/refresh',
        {
          params: {
            header: { ...csrf.header, 'X-GoSeed-Refresh-Key': retryKey },
          },
        },
      )
      if (epoch !== sessionEpoch) return false
      if (response.status === 401) {
        clear()
        return false
      }
      if (!response.ok || error || !data)
        throw new APIError(
          error?.message || '刷新失败',
          error?.code,
          error?.traceId,
        )
      rememberRetry('')
      setSession(data.data, response, started)
      return true
    }).finally(() => {
      refreshPromise = null
    })
    return refreshPromise
  }
  async function recover(failedToken: string) {
    if (failedToken !== accessToken.value) return !!accessToken.value
    return refresh()
  }
  async function restore() {
    if (initialized.value) return
    try {
      await refresh()
    } catch (error) {
      connectionError.value = authMessage(error)
    } finally {
      initialized.value = true
    }
  }
  async function ensureAccess() {
    if (!accessToken.value || accessDeadline <= performance.now() + 30_000) {
      if (!(await refresh())) throw new APIError('登录已失效', 100401)
    }
    return accessToken.value
  }
  async function login(username: string, password: string) {
    await cookieOperation(async () => {
      const started = performance.now()
      const { data, error, response } = await api.POST(
        '/admin-api/v1/auth/login',
        { params: csrf, body: { username, password } },
      )
      if (!response.ok || error || !data)
        throw new APIError(
          error?.message || '登录失败',
          error?.code,
          error?.traceId,
        )
      sessionEpoch++
      rememberRetry('')
      setSession(data.data, response, started)
    })
  }
  async function current() {
    const token = await ensureAccess()
    const epoch = sessionEpoch
    const { data, error, response } = await api.GET('/admin-api/v1/auth/me', {
      headers: { Authorization: `Bearer ${token}` },
    })
    // A delayed poll must not restore a signed-out session or overwrite a newer login.
    if (epoch !== sessionEpoch) return
    if (!response.ok || error || !data)
      throw new APIError(
        error?.message || '会话已失效',
        error?.code,
        error?.traceId,
      )
    if (
      policyVersion.value &&
      BigInt(data.data.policyVersion) < BigInt(policyVersion.value)
    )
      return
    admin.value = data.data.admin
    // Keep identity stable on unchanged polls so open forms are not discarded.
    if (permissions.value.join('\n') !== data.data.permissions.join('\n'))
      permissions.value = data.data.permissions
    roles.value = data.data.roles
    policyVersion.value = data.data.policyVersion
    superAdmin.value = data.data.superAdmin
  }
  async function permissionsChanged() {
    channel?.postMessage('permissions')
    await current()
  }
  async function logout() {
    await cookieOperation(async () => {
      const { error, response } = await api.POST('/admin-api/v1/auth/logout', {
        params: csrf,
      })
      if (!response.ok && response.status !== 401)
        throw new APIError(
          error?.message || '退出失败',
          error?.code,
          error?.traceId,
        )
      signedOut()
    })
  }
  async function changePassword(currentPassword: string, newPassword: string) {
    // Refresh before taking the cookie lock so the middleware cannot deadlock on rotation.
    const token = await ensureAccess()
    await cookieOperation(async () => {
      const { error, response } = await api.POST(
        '/admin-api/v1/auth/password',
        {
          params: csrf,
          body: { currentPassword, newPassword },
          headers: { Authorization: `Bearer ${token}` },
        },
      )
      if (!response.ok || error)
        throw new APIError(
          error?.message || '修改失败',
          error?.code,
          error?.traceId,
        )
      signedOut()
    })
  }
  return {
    admin,
    accessToken,
    permissions,
    roles,
    policyVersion,
    superAdmin,
    has,
    permissionsChanged,
    initialized,
    connectionError,
    clear,
    sync,
    restore,
    ensureAccess,
    recover,
    login,
    current,
    logout,
    changePassword,
  }
})
