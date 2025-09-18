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

// 简单的全局 401 拦截：基于 fetch 包装
// const originalFetch = window.fetch
// window.fetch = async (input: RequestInfo | URL, init?: RequestInit) => {
//   const response = await originalFetch(input, init)
//   if (response.status === 401) {
//     const current = router.currentRoute.value
//     const currentPath = current.fullPath || '/'
//     if (!current.meta?.public && !current.path.startsWith('/login')) {
//       router.replace({ path: '/login', query: { redirect: currentPath } })
//     }
//   }
//   return response
// }
