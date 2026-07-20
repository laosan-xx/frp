import { createRouter, createWebHashHistory } from 'vue-router'
import Login from '../views/Login.vue'
import ServerOverview from '../views/ServerOverview.vue'
import Clients from '../views/Clients.vue'
import ClientDetail from '../views/ClientDetail.vue'
import Proxies from '../views/Proxies.vue'
import ProxyDetail from '../views/ProxyDetail.vue'

const router = createRouter({
  history: createWebHashHistory(),
  scrollBehavior() {
    return { top: 0 }
  },
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
    {
      path: '/clients',
      name: 'Clients',
      component: Clients,
    },
    {
      path: '/clients/:key',
      name: 'ClientDetail',
      component: ClientDetail,
    },
    {
      path: '/proxies/:type?',
      name: 'Proxies',
      component: Proxies,
    },
    {
      path: '/proxy/:name',
      name: 'ProxyDetail',
      component: ProxyDetail,
    },
  ],
})

// Route guard: public routes pass through, protected routes rely on
// API 401 responses handled by the global fetch interceptor.
router.beforeEach((to: any, _: any, next: any) => {
  if (to.meta.public) {
    next()
    return
  }
  next()
})

export default router
