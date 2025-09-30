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
router.beforeEach((to, _, next) => {
  // 如果是公开路由（如登录页），直接放行
  if (to.meta.public) {
    next()
    return
  }

  // 对于受保护的路由，直接放行
  // 如果用户未登录，后续的 API 请求会返回 401
  // 全局 401 拦截器会自动处理重定向到登录页
  next()
})

export default router
