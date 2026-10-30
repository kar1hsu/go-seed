import { createApp } from 'vue'
import { createPinia } from 'pinia'
import App from './app/App.vue'
import { router } from './router'
import { api } from './lib/http'
import { useAuthStore } from './features/auth/store'
import { authMiddleware } from './features/auth/middleware'
import './styles/element'
import './styles/main.css'

const app = createApp(App)
const pinia = createPinia()
app.use(pinia)
const auth = useAuthStore(pinia)
api.use(
  authMiddleware(auth, () => {
    if (router.currentRoute.value.meta.requiresAuth)
      void router.replace({
        path: '/login',
        query: {
          reason: 'expired',
          redirect: router.currentRoute.value.fullPath,
        },
      })
  }),
)
const stopSync = auth.sync(() => {
  void router.replace('/login')
})
app.onUnmount(stopSync)
app.use(router).mount('#app')
