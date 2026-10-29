<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { ElButton, ElInput } from 'element-plus'
import { useAuthStore, authMessage } from '../store'
import { APIError } from '../../../lib/http'
const auth = useAuthStore()
const router = useRouter()
const current = ref('')
const password = ref('')
const confirm = ref('')
const error = ref('')
const loading = ref(false)
async function submit() {
  if (loading.value) return
  error.value = ''
  const length = new TextEncoder().encode(password.value).length
  if (!current.value) {
    error.value = '请输入当前密码'
    return
  }
  if (length < 12 || length > 72) {
    error.value = '新密码至少需要 12 位，且不能超过 72 字节'
    return
  }
  if (password.value !== confirm.value) {
    error.value = '两次输入的新密码不一致'
    return
  }
  if (current.value === password.value) {
    error.value = '新密码不能与当前密码相同'
    return
  }
  loading.value = true
  try {
    await auth.changePassword(current.value, password.value)
    current.value = password.value = confirm.value = ''
    await router.replace('/login?reason=password-changed')
  } catch (cause) {
    if (cause instanceof APIError && cause.code === 100401) {
      await router.replace('/login?reason=expired')
      return
    }
    error.value = authMessage(cause)
  } finally {
    loading.value = false
  }
}
</script>
<template>
  <div class="page-heading">
    <div>
      <h1>账户安全</h1>
      <p>管理当前账号的登录密码。</p>
    </div>
  </div>
  <section class="panel">
    <div class="panel-heading"><h2>修改密码</h2></div>
    <div class="panel-body">
      <form class="password-form" @submit.prevent="submit">
        <div class="account-summary">
          <span class="field-label">当前账号</span
          ><strong>{{ auth.admin?.username }}</strong>
        </div>
        <label for="current-password">当前密码</label
        ><ElInput
          id="current-password"
          v-model="current"
          type="password"
          show-password
          autocomplete="current-password"
          :disabled="loading"
        />
        <label for="new-password">新密码</label
        ><ElInput
          id="new-password"
          v-model="password"
          type="password"
          show-password
          autocomplete="new-password"
          placeholder="至少 12 位"
          :disabled="loading"
        />
        <label for="confirm-password">确认新密码</label
        ><ElInput
          id="confirm-password"
          v-model="confirm"
          type="password"
          show-password
          autocomplete="new-password"
          :disabled="loading"
        />
        <p class="password-note">
          修改成功后，所有设备上的登录都会失效，需要重新登录。
        </p>
        <div v-if="error" class="auth-error" role="alert">{{ error }}</div>
        <ElButton type="primary" native-type="submit" :loading="loading"
          >保存密码</ElButton
        >
      </form>
    </div>
  </section>
</template>
