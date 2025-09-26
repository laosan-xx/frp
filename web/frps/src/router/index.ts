import { createRouter, createWebHashHistory } from 'vue-router'
import Login from '../views/Login.vue'
import ServerOverview from '../components/ServerOverview.vue'
import ProxiesTCP from '../components/ProxiesTCP.vue'
import ProxiesUDP from '../components/ProxiesUDP.vue'
import ProxiesHTTP from '../components/ProxiesHTTP.vue'
import ProxiesHTTPS from '../components/ProxiesHTTPS.vue'
import ProxiesTCPMux from '../components/ProxiesTCPMux.vue'
import ProxiesSTCP from '../components/ProxiesSTCP.vue'
import ProxiesSUDP from '../components/ProxiesSUDP.vue'

const router = createRouter({
  history: createWebHashHistory(),
  routes: [
    {
      path: '/login',
      name: 'Login',
      component: Login,
      meta: { layout: 'auth', public: true },
    },
    {
      path: '/',
      name: 'ServerOverview',
      component: ServerOverview,
    },
    { path: '/:pathMatch(.*)*', redirect: '/' },
    {
      path: '/proxies/tcp',
      name: 'ProxiesTCP',
      component: ProxiesTCP,
    },
    {
      path: '/proxies/udp',
      name: 'ProxiesUDP',
      component: ProxiesUDP,
    },
    {
      path: '/proxies/http',
      name: 'ProxiesHTTP',
      component: ProxiesHTTP,
    },
    {
      path: '/proxies/https',
      name: 'ProxiesHTTPS',
      component: ProxiesHTTPS,
    },
    {
      path: '/proxies/tcpmux',
      name: 'ProxiesTCPMux',
      component: ProxiesTCPMux,
    },
    {
      path: '/proxies/stcp',
      name: 'ProxiesSTCP',
      component: ProxiesSTCP,
    },
    {
      path: '/proxies/sudp',
      name: 'ProxiesSUDP',
      component: ProxiesSUDP,
    },
  ],
})

// 添加路由守卫，检查认证状态
router.beforeEach(async (to, _, next) => {
  // 如果是公开路由（如登录页），直接放行
  if (to.meta.public) {
    next()
    return
  }

  // 检查是否已登录
  try {
    const response = await fetch('/api/serverinfo')
    if (response.ok) {
      // 已登录，允许访问
      next()
    } else if (response.status === 401) {
      // 未登录，重定向到登录页
      next({ path: '/login', query: { redirect: to.fullPath } })
    } else {
      // 其他错误，也重定向到登录页
      next({ path: '/login', query: { redirect: to.fullPath } })
    }
  } catch (error) {
    // 网络错误，重定向到登录页
    next({ path: '/login', query: { redirect: to.fullPath } })
  }
})

export default router
