<template>
  <div id="app">
    <!-- 顶部导航栏（登录页不显示） -->
    <header class="app-header" v-if="$route.meta.layout !== 'auth'">
      <div class="header-container">
        <div class="brand">
          <a href="#" class="brand-link">
            <span class="brand-logo">frp</span>
            <span class="brand-text">Dashboard</span>
          </a>
        </div>
        <div class="header-actions">
          <el-switch
            v-model="darkmodeSwitch"
            inline-prompt
            active-text="🌙"
            inactive-text="☀️"
            @change="toggleDark"
            class="theme-switch"
          />
          <el-button
            type="danger"
            plain
            size="small"
            @click="handleLogout"
            class="logout-btn"
            circle
          >
            <span class="logout-icon">⏻</span>
          </el-button>
        </div>
      </div>
    </header>

    <!-- 主内容区域（如果是登录页，使用简化布局） -->
    <main class="app-main" v-if="$route.meta.layout !== 'auth'">
      <div class="main-container">
        <!-- 侧边导航 - 桌面端显示 -->
        <aside class="sidebar-desktop">
          <el-menu
            :default-active="activeMenu"
            :default-openeds="defaultOpeneds"
            mode="vertical"
            router
            @select="handleSelect"
            class="sidebar-menu"
          >
            <el-menu-item index="/" class="menu-item">
              <template #title>
                <span class="menu-icon">📊</span>
                <span class="menu-text">Overview</span>
              </template>
            </el-menu-item>

            <el-sub-menu index="/proxies" class="menu-sub">
              <template #title>
                <span class="menu-icon">🔗</span>
                <span class="menu-text">Proxies</span>
              </template>
              <el-menu-item index="/proxies/tcp" class="submenu-item"
                >TCP</el-menu-item
              >
              <el-menu-item index="/proxies/udp" class="submenu-item"
                >UDP</el-menu-item
              >
              <el-menu-item index="/proxies/http" class="submenu-item"
                >HTTP</el-menu-item
              >
              <el-menu-item index="/proxies/https" class="submenu-item"
                >HTTPS</el-menu-item
              >
              <el-menu-item index="/proxies/tcpmux" class="submenu-item"
                >TCPMUX</el-menu-item
              >
              <el-menu-item index="/proxies/stcp" class="submenu-item"
                >STCP</el-menu-item
              >
              <el-menu-item index="/proxies/sudp" class="submenu-item"
                >SUDP</el-menu-item
              >
            </el-sub-menu>
          </el-menu>
        </aside>

        <!-- 移动端汉堡菜单 -->
        <div
          class="mobile-menu-toggle"
          @click="isMobileMenuOpen = !isMobileMenuOpen"
        >
          <span class="hamburger-icon">☰</span>
        </div>

        <!-- 移动端侧边栏 -->
        <transition name="slide-right">
          <aside
            :class="[
              'sidebar-mobile',
              { 'mobile-menu-open': isMobileMenuOpen },
            ]"
          >
            <div class="mobile-header">
              <span class="close-icon" @click="isMobileMenuOpen = false"
                >×</span
              >
            </div>
            <el-menu
              :default-active="activeMenu"
              :default-openeds="defaultOpeneds"
              mode="vertical"
              router
              @select="handleMobileSelect"
              class="mobile-sidebar-menu"
            >
              <el-menu-item index="/" class="menu-item">
                <template #title>
                  <span class="menu-icon">📊</span>
                  <span class="menu-text">Overview</span>
                </template>
              </el-menu-item>

              <el-sub-menu index="/proxies" class="menu-sub">
                <template #title>
                  <span class="menu-icon">🔗</span>
                  <span class="menu-text">Proxies</span>
                </template>
                <el-menu-item index="/proxies/tcp" class="submenu-item"
                  >TCP</el-menu-item
                >
                <el-menu-item index="/proxies/udp" class="submenu-item"
                  >UDP</el-menu-item
                >
                <el-menu-item index="/proxies/http" class="submenu-item"
                  >HTTP</el-menu-item
                >
                <el-menu-item index="/proxies/https" class="submenu-item"
                  >HTTPS</el-menu-item
                >
                <el-menu-item index="/proxies/tcpmux" class="submenu-item"
                  >TCPMUX</el-menu-item
                >
                <el-menu-item index="/proxies/stcp" class="submenu-item"
                  >STCP</el-menu-item
                >
                <el-menu-item index="/proxies/sudp" class="submenu-item"
                  >SUDP</el-menu-item
                >
              </el-sub-menu>
            </el-menu>
          </aside>
        </transition>

        <!-- 内容区域 -->
        <div class="content-area">
          <div class="content-container">
            <router-view></router-view>
          </div>
        </div>
      </div>
    </main>

    <!-- 登录页的极简布局：不显示侧边栏，仅渲染路由内容 -->
    <main v-else class="auth-main">
      <router-view></router-view>
    </main>

    <!-- 遮罩层 - 移动端菜单打开时显示（登录页不显示） -->
    <transition name="fade" v-if="$route.meta.layout !== 'auth'">
      <div
        v-show="isMobileMenuOpen"
        :class="['mobile-overlay', { 'mobile-menu-open': isMobileMenuOpen }]"
        @click="isMobileMenuOpen = false"
      ></div>
    </transition>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { useDark, useToggle } from '@vueuse/core'
import { useRoute, useRouter } from 'vue-router'
import { ElMessageBox, ElMessage } from 'element-plus'

const isDark = useDark()
const darkmodeSwitch = ref(isDark)
const toggleDark = useToggle(isDark)
const isMobileMenuOpen = ref(false)

const route = useRoute()
const router = useRouter()
const activeMenu = computed(() => route.path)
const defaultOpeneds = computed(() =>
  route.path.startsWith('/proxies') ? ['/proxies'] : [],
)

const handleSelect = (key: string) => {
  if (key == '') {
    window.open('https://github.com/laosan-xx/frp')
  }
}

const handleMobileSelect = (key: string) => {
  handleSelect(key)
  isMobileMenuOpen.value = false
}

const handleLogout = async () => {
  try {
    await ElMessageBox.confirm('确定要登出吗？', '确认登出', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning',
      customStyle: {
        marginTop: '-10vh',
      },
    })

    // 调用登出 API
    const response = await fetch('/api/logout', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
    })

    if (response.ok) {
      ElMessage.success('登出成功')
      // 跳转到登录页
      router.push('/login')
    } else {
      ElMessage.error('登出失败')
    }
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('登出时发生错误')
    }
  }
}
</script>

<style scoped>
/* 基础样式重置 */
#app {
  min-height: 100vh;
  display: flex;
  flex-direction: column;
}

/* 顶部导航栏样式 */
.app-header {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  box-shadow: 0 2px 12px rgba(0, 0, 0, 0.1);
  position: sticky;
  top: 0;
  z-index: 1000;
}

html.dark .app-header {
  background: linear-gradient(135deg, #2c3e50 0%, #34495e 100%);
}

.header-container {
  display: flex;
  align-items: center;
  justify-content: space-between;
  height: 64px;
  padding: 0 20px;
  max-width: 1400px;
  margin: 0 auto;
}

.brand-link {
  display: flex;
  align-items: center;
  text-decoration: none;
  color: white;
  gap: 12px;
}

.brand-logo {
  font-size: 28px;
  font-weight: bold;
  background: rgba(255, 255, 255, 0.2);
  padding: 5px 16px;
  border-radius: 8px;
  backdrop-filter: blur(10px);
}

.brand-text {
  font-size: 18px;
  font-weight: 500;
  display: inline;
}

.theme-switch {
  --el-switch-on-color: #4a5568;
  --el-switch-off-color: #3182ce;
}

.header-actions {
  display: flex;
  align-items: center;
  gap: 8px;
}

.logout-btn {
  margin-left: 4px;
  height: 20px;
  width: 20px;
}

.logout-icon {
  margin-right: 0;
  margin-top: -2px;
}

/* 主内容区域 */
.app-main {
  flex: 1;
  padding: 16px 0;
}

.auth-main {
  flex: 1;
}

.main-container {
  display: grid;
  grid-template-columns: 240px 1fr;
  gap: 20px;
  max-width: 1400px;
  margin: 0 auto;
  padding: 0 20px;
  position: relative;
}

/* 侧边栏样式 */
.sidebar-desktop {
  position: sticky;
  top: 80px;
  height: calc(100vh - 112px);
  overflow-y: auto;
}

.sidebar-menu {
  border: none;
  border-radius: 12px;
  background: white;
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.08);
  padding: 16px 0;
}

html.dark .sidebar-menu {
  background: #2d3748;
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.3);
}

.menu-item,
.menu-sub {
  margin: 4px 12px;
  border-radius: 8px;
  transition: all 0.3s ease;
  overflow: hidden;
}

.menu-item.is-active {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: white;
}

.menu-item.is-active .menu-icon,
.menu-item.is-active .menu-text {
  color: white;
}
.menu-item.is-active:hover .menu-text {
  color: #303133;
}
html.dark .menu-item.is-active:hover .menu-text {
  color: white;
}

.menu-item:hover,
.menu-sub:hover {
  background: #f7fafc;
}

html.dark .menu-item:hover,
html.dark .menu-sub:hover {
  background: #4a5568;
}

html.dark .menu-item.is-active {
  background: linear-gradient(135deg, #2c3e50 0%, #34495e 100%);
  color: white;
}

.submenu-item.is-active {
  background: #ebf4ff;
  color: #3182ce;
  font-weight: 600;
}

html.dark .submenu-item.is-active {
  background: rgba(99, 179, 237, 0.15);
  color: #63b3ed;
  box-shadow: inset 2px 0 0 #63b3ed;
  font-weight: 500;
}

.menu-icon {
  margin-right: 12px;
  font-size: 16px;
}

.menu-text {
  font-weight: 500;
}

.submenu-item {
  padding-left: 48px !important;
  height: 46px;
}

/* 内容区域 */
.content-area {
  min-height: calc(100vh - 160px);
}

.content-container {
  background: white;
  border-radius: 12px;
  padding: 24px;
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.08);
  /* 让内容区域高度自适应，支持子组件的高度限制 */
  height: calc(100vh - 64px - 32px);
  overflow: hidden;
}

/* 允许 Grid/Flex 子项在内容过宽时正常收缩 */
.content-area,
.content-container {
  min-width: 0;
  width: 100%;
  overflow-x: auto;
}

html.dark .content-container {
  background: #2d3748;
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.3);
}

/* 移动端样式 */
.mobile-menu-toggle {
  display: none; /* 默认隐藏，在媒体查询中显示 */
  position: fixed;
  bottom: 24px;
  right: 24px;
  z-index: 1001;
  background: #667eea;
  color: white;
  width: 56px;
  height: 56px;
  border-radius: 50%;
  align-items: center;
  justify-content: center;
  font-size: 24px;
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.3);
  cursor: pointer;
}

.sidebar-mobile {
  display: none; /* 默认隐藏，在媒体查询中显示 */
  position: fixed;
  top: 0;
  left: 0;
  width: 280px;
  height: 100vh;
  background: white;
  z-index: 1002;
  box-shadow: 2px 0 20px rgba(0, 0, 0, 0.15);
  overflow-y: auto;
}

html.dark .sidebar-mobile {
  background: #2d3748;
}

.mobile-header {
  padding: 10px 14px;
  border-bottom: 1px solid #e2e8f0;
  display: flex;
  justify-content: flex-end;
}

html.dark .mobile-header {
  border-bottom-color: #4a5568;
}

.close-icon {
  font-size: 24px;
  cursor: pointer;
  color: #718096;
}

.mobile-sidebar-menu {
  border: none;
  padding: 16px 0;
}

.mobile-overlay {
  display: none;
  position: fixed;
  top: 0;
  left: 0;
  width: 100vw;
  height: 100vh;
  background: rgba(0, 0, 0, 0.5);
  z-index: 1001;
  backdrop-filter: blur(4px);
}

/* 响应式设计 */
@media (max-width: 1200px) {
  .main-container {
    grid-template-columns: 1fr;
    padding: 0 16px;
  }

  .sidebar-desktop {
    display: none;
  }

  .mobile-menu-toggle {
    display: flex;
  }

  /* 移动端菜单默认隐藏 */
  .sidebar-mobile,
  .mobile-overlay {
    display: none;
  }

  /* 移动端菜单打开时的显示状态 */
  .sidebar-mobile.mobile-menu-open,
  .mobile-overlay.mobile-menu-open {
    display: block;
  }

  .content-container {
    padding: 16px;
    margin: 0;
    border-radius: 0;
    box-shadow: none;
  }

  .header-container {
    padding: 0 16px;
    height: 56px;
  }

  .brand-logo {
    font-size: 24px;
    padding: 4px 12px;
  }
}

@media (max-width: 768px) {
  .content-container {
    padding: 10px;
  }
}

/* 动画效果 */
.slide-right-enter-active,
.slide-right-leave-active {
  transition: transform 0.3s ease;
}

.slide-right-enter-from,
.slide-right-leave-to {
  transform: translateX(-100%);
}

.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.3s ease;
}

.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}

/* 滚动条样式 */
.sidebar-desktop::-webkit-scrollbar {
  width: 4px;
}

.sidebar-desktop::-webkit-scrollbar-track {
  background: #f1f1f1;
  border-radius: 2px;
}

html.dark .sidebar-desktop::-webkit-scrollbar-track {
  background: #4a5568;
}

.sidebar-desktop::-webkit-scrollbar-thumb {
  background: #cbd5e0;
  border-radius: 2px;
}

.sidebar-desktop::-webkit-scrollbar-thumb:hover {
  background: #a0aec0;
}
</style>
