import { createApp } from 'vue'
import 'element-plus/theme-chalk/dark/css-vars.css'
import App from './App.vue'
import router from './router'
import i18n from './locales'

import './assets/css/var.css'
import './assets/css/dark.css'

const app = createApp(App)

app.use(router)
app.use(i18n)

app.mount('#app')

// Global 401 interceptor: redirect to login page when API returns 401
const originalFetch = window.fetch
window.fetch = async (input: RequestInfo | URL, init?: RequestInit) => {
  const response = await originalFetch(input, init)
  if (response.status === 401) {
    const current = router.currentRoute.value
    const currentPath = (current && (current as any).fullPath) || '/'
    const isPublic =
      (current && (current as any).meta && (current as any).meta.public) || false
    if (
      !isPublic &&
      !(current && (current as any).path && (current as any).path.startsWith('/login'))
    ) {
      router.replace({ path: '/login', query: { redirect: currentPath } })
    }
  }
  return response
}
