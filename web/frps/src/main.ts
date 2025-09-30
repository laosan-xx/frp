import { createApp } from 'vue'
import 'normalize.css'
import 'element-plus/dist/index.css'
import 'element-plus/theme-chalk/dark/css-vars.css'
import 'element-plus/theme-chalk/display.css'
import App from './App.vue'
import router from './router'

import './assets/custom.css'
import './assets/dark.css'
import './assets/global.css'

const app = createApp(App)

app.use(router)

app.mount('#app')

// 全局 401 拦截：基于 fetch 包装
const originalFetch = window.fetch
window.fetch = async (input: RequestInfo | URL, init?: RequestInit) => {
  const response = await originalFetch(input, init)
  if (response.status === 401) {
    const current = router.currentRoute.value
    const currentPath = (current && (current as any).fullPath) || '/'
    const isPublic = (current && (current as any).meta && (current as any).meta.public) || false
    if (!isPublic && !(current && (current as any).path && (current as any).path.startsWith('/login'))) {
      router.replace({ path: '/login', query: { redirect: currentPath } })
    }
  }
  return response
}
