import { defineStore } from 'pinia'
import { ref } from 'vue'
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
      100403: '当前请求未被允许，请检查访问地址',
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
  const accessToken = ref('')
  const expiresAt = ref(0)
  const initialized = ref(false)
  const connectionError = ref('')
  let refreshPromise: Promise<boolean> | null = null
  let channel: BroadcastChannel | null = null

  function clear() {
    admin.value = null
    accessToken.value = ''
    expiresAt.value = 0
  }
  function setSession(data: AuthData) {
    admin.value = data.admin
    accessToken.value = data.accessToken
    expiresAt.value = Date.parse(data.expiresAt)
    initialized.value = true
    connectionError.value = ''
  }
  function sync(onLogout: () => void) {
    if (typeof BroadcastChannel === 'undefined') return () => {}
    channel = new BroadcastChannel('goseed-admin-session')
    channel.onmessage = (event) => {
      if (event.data === 'logout') {
        clear()
        initialized.value = true
        onLogout()
      }
    }
    return () => {
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
    refreshPromise = cookieOperation(async () => {
      const { data, error, response } = await api.POST(
        '/admin-api/v1/auth/refresh',
        { params: csrf },
      )
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
      setSession(data.data)
      return true
    }).finally(() => {
      refreshPromise = null
    })
    return refreshPromise
  }
  async function restore() {
    if (initialized.value) return
    try {
      await refresh()
    } catch (error) {
      clear()
      connectionError.value = authMessage(error)
    } finally {
      initialized.value = true
    }
  }
  async function ensureAccess() {
    if (!accessToken.value || expiresAt.value <= Date.now() + 30_000) {
      if (!(await refresh())) throw new APIError('登录已失效', 100401)
    }
    return accessToken.value
  }
  async function login(username: string, password: string) {
    await cookieOperation(async () => {
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
      setSession(data.data)
    })
  }
  async function current() {
    const { data, error, response } = await api.GET('/admin-api/v1/auth/me')
    if (!response.ok || error || !data)
      throw new APIError(
        error?.message || '会话已失效',
        error?.code,
        error?.traceId,
      )
    admin.value = data.data.admin
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
    initialized,
    connectionError,
    clear,
    sync,
    restore,
    ensureAccess,
    login,
    current,
    logout,
    changePassword,
  }
})
