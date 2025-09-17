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
