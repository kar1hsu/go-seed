<script setup lang="ts">
import { ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElButton, ElInput } from 'element-plus'
import { User, Lock } from '@element-plus/icons-vue'
import { useAuthStore, authMessage } from '../store'
const auth = useAuthStore()
const route = useRoute()
const router = useRouter()
const username = ref('')
const password = ref('')
const error = ref('')
const loading = ref(false)
async function submit() {
  if (loading.value) return
  error.value = ''
  if (!username.value.trim() || !password.value) {
    error.value = '请输入账号和密码'
    return
  }
  loading.value = true
  try {
    await auth.login(username.value.trim(), password.value)
    password.value = ''
    const redirect = route.query.redirect
    await router.replace(
      typeof redirect === 'string' &&
        redirect.startsWith('/') &&
        !redirect.startsWith('//') &&
        !redirect.startsWith('/login')
        ? redirect
        : '/',
    )
  } catch (cause) {
    error.value = authMessage(cause)
  } finally {
    loading.value = false
  }
}
</script>
<template>
  <div class="login-page">
    <header class="login-brand">
      <img src="/seed.svg" alt="" /><strong>GoSeed</strong><span>Admin</span>
    </header>
    <main class="login-content">
      <section class="login-card" aria-labelledby="login-title">
        <h1 id="login-title">登录管理后台</h1>
        <p class="login-description">请输入您的管理员账号和密码</p>
        <p
          v-if="route.query.reason === 'password-changed'"
          class="success-notice"
          role="status"
        >
          密码已更新，请使用新密码登录。
        </p>
        <p v-else-if="route.query.reason === 'expired'" class="login-hint">
          登录已失效，请重新登录。
        </p>
        <form @submit.prevent="submit">
          <label for="login-username">账号</label>
          <ElInput
            id="login-username"
            v-model="username"
            :prefix-icon="User"
            size="large"
            placeholder="请输入账号"
            autocomplete="username"
            :maxlength="32"
            :disabled="loading"
          />
          <label for="login-password">密码</label>
          <ElInput
            id="login-password"
            v-model="password"
            :prefix-icon="Lock"
            size="large"
            type="password"
            placeholder="请输入密码"
            autocomplete="current-password"
            show-password
            :disabled="loading"
          />
          <div
            v-if="error || auth.connectionError"
            class="auth-error"
            role="alert"
          >
            {{ error || auth.connectionError }}
          </div>
          <ElButton
            type="primary"
            native-type="submit"
            size="large"
            class="login-submit"
            :loading="loading"
            >登录</ElButton
          >
        </form>
      </section>
    </main>
    <footer class="login-footer">GoSeed · 管理控制台</footer>
  </div>
</template>
