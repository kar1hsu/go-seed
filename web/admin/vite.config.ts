import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'

export default defineConfig({
  plugins: [vue()],
  server: {
    host: '127.0.0.1',
    port: Number(process.env.WEB_PORT || 5173),
    strictPort: true,
    proxy: {
      '/admin-api': {
        target: `http://127.0.0.1:${process.env.ADMIN_API_PORT || 8081}`,
      },
    },
  },
})
