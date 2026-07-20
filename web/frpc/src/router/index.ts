import { createRouter, createWebHashHistory } from 'vue-router'
import { ElMessage } from 'element-plus'
import Login from '../views/Login.vue'
import ClientConfigure from '../views/ClientConfigure.vue'
import ProxyDetail from '../views/ProxyDetail.vue'
import ProxyEdit from '../views/ProxyEdit.vue'
import ProxyList from '../views/ProxyList.vue'
import VisitorDetail from '../views/VisitorDetail.vue'
import VisitorEdit from '../views/VisitorEdit.vue'
import VisitorList from '../views/VisitorList.vue'
import { useProxyStore } from '../stores/proxy'

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
      redirect: '/proxies',
    },
    {
      path: '/proxies',
      name: 'ProxyList',
      component: ProxyList,
    },
    {
      path: '/proxies/detail/:name',
      name: 'ProxyDetail',
      component: ProxyDetail,
    },
    {
      path: '/proxies/create',
      name: 'ProxyCreate',
      component: ProxyEdit,
      meta: { requiresStore: true },
    },
    {
      path: '/proxies/:name/edit',
      name: 'ProxyEdit',
      component: ProxyEdit,
      meta: { requiresStore: true },
    },
    {
      path: '/visitors',
      name: 'VisitorList',
      component: VisitorList,
    },
    {
      path: '/visitors/detail/:name',
      name: 'VisitorDetail',
      component: VisitorDetail,
    },
    {
      path: '/visitors/create',
      name: 'VisitorCreate',
      component: VisitorEdit,
      meta: { requiresStore: true },
    },
    {
      path: '/visitors/:name/edit',
      name: 'VisitorEdit',
      component: VisitorEdit,
      meta: { requiresStore: true },
    },
    {
      path: '/config',
      name: 'ClientConfigure',
      component: ClientConfigure,
    },
  ],
})

// Route guard: public routes pass through, protected routes rely on
// API 401 responses handled by the global fetch interceptor.
router.beforeEach(async (to: any) => {
  if (to.meta.public) {
    return true
  }

  if (!to.matched.some((record: any) => record.meta.requiresStore)) {
    return true
  }

  const proxyStore = useProxyStore()
  const enabled = await proxyStore.checkStoreEnabled()
  if (enabled) {
    return true
  }

  ElMessage.warning(
    'Store is disabled. Enable Store in frpc config to create or edit store entries.',
  )
  return { name: 'ProxyList' }
})

export default router
