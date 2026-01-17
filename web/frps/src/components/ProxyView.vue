<template>
  <div class="proxy-view-container">
    <el-page-header :icon="null" style="margin-bottom: 10px">
      <template #title>
        <span>{{ proxyType }}</span>
      </template>
      <template #content> </template>
      <template #extra>
        <div class="flex items-center">
          <el-button :loading="loading" @click="$emit('refresh')"
            >刷新</el-button
          >
          <el-popconfirm
            title="你确定要清除离线代理的所有数据吗？"
            @confirm="clearOfflineProxies"
            width="300"
          >
            <template #reference>
              <el-button>清理离线</el-button>
            </template>
          </el-popconfirm>
        </div>
      </template>
    </el-page-header>

    <!-- 顶部粘性细进度条（自定义 NProgress 风格） -->
    <div class="top-loading" :class="{ 'is-hidden': !loading }">
      <div class="top-loading__bar"></div>
      <div class="top-loading__peg"></div>
    </div>

    <el-table
      :data="filteredProxies"
      :default-sort="{ prop: 'name', order: 'ascending' }"
      style="width: 100%"
      class="proxy-table"
      :max-height="tableMaxHeight"
    >
      <el-table-column type="expand" width="30px">
        <template #default="expandProps">
          <ProxyViewExpand :row="expandProps.row" :proxyType="proxyType" />
        </template>
      </el-table-column>
      <el-table-column label="名称" prop="name" min-width="130px">
        <template #header>
          <div class="name-filter-header">
            <span>名称</span>
            <el-input
              v-model="nameFilter"
              size="small"
              placeholder="筛选"
              clearable
              class="name-filter-input"
              @click.stop
            />
          </div>
        </template>
      </el-table-column>
      <el-table-column label="端口" prop="port" sortable min-width="80px">
        <template #default="scope">
          <el-link
            v-if="
              scope.row.status === 'online' && scope.row.name.includes('.web')
            "
            @click="openManage(scope.row)"
          >
            <span>{{ scope.row.port }}</span>
          </el-link>
        </template>
      </el-table-column>
      <el-table-column v-if="!isMdOrBelow" label="连接数" prop="conns" sortable>
      </el-table-column>
      <!-- <el-table-column
        label="Traffic In"
        prop="trafficIn"
        :formatter="formatTrafficIn"
        sortable
      >
      </el-table-column>
      <el-table-column
        label="Traffic Out"
        prop="trafficOut"
        :formatter="formatTrafficOut"
        sortable
      >
      </el-table-column> -->
      <el-table-column
        v-if="!isSmOrBelow"
        label="客户端版本"
        prop="clientVersion"
        sortable
        min-width="120px"
      >
      </el-table-column>
      <el-table-column label="状态" prop="status" sortable min-min-width="90px">
        <template #default="scope">
          <el-tag v-if="scope.row.status === 'online'" type="success"
            >在线</el-tag
          >
          <el-tag v-else type="danger">离线</el-tag>
        </template>
      </el-table-column>
      <el-table-column label="操作" min-min-width="90px">
        <template #default="scope">
          <el-button type="primary" @click="openTraffic(scope.row)"
            >流量
          </el-button>
        </template>
      </el-table-column>
    </el-table>
  </div>

  <el-dialog
    v-model="dialogVisible"
    :destroy-on-close="true"
    :title="dialogVisibleName"
    width="700px"
  >
    <Traffic :proxyName="dialogVisibleName" />
  </el-dialog>
</template>

<script setup lang="ts">
// import * as Humanize from 'humanize-plus'
// import type { TableColumnCtx } from 'element-plus'
import type { BaseProxy } from '../utils/proxy.js'
import { ElMessage } from 'element-plus'
import ProxyViewExpand from './ProxyViewExpand.vue'
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { useBreakpoints } from '../utils/breakpoints'
import { pinyin } from 'pinyin-pro'

const props = defineProps<{
  proxies: BaseProxy[]
  proxyType: string
  loading?: boolean
}>()

const emit = defineEmits(['refresh'])

// 动态计算表格最大高度
const windowHeight = ref(window.innerHeight)
const updateWindowHeight = () => {
  windowHeight.value = window.innerHeight
}
onMounted(() => {
  window.addEventListener('resize', updateWindowHeight)
})
onUnmounted(() => {
  window.removeEventListener('resize', updateWindowHeight)
})

// 计算表格最大高度：窗口高度 - header(64) - app-main padding(32) - content-container padding(48) - page-header(52) - 额外边距(20)
const tableMaxHeight = computed(() => {
  return windowHeight.value - 64 - 32 - 48 - 46
})

// 获取字符串的拼音首字母
const getPinyinInitials = (str: string): string => {
  return pinyin(str, { pattern: 'first', toneType: 'none' }).replace(/\s/g, '')
}

// 名称筛选（支持名称匹配和拼音首字母匹配）
const nameFilter = ref('')
const filteredProxies = computed(() => {
  if (!nameFilter.value) {
    return props.proxies
  }
  const keyword = nameFilter.value.toLowerCase()
  return props.proxies.filter((proxy) => {
    const name = proxy.name.toLowerCase()
    const initials = getPinyinInitials(proxy.name).toLowerCase()
    return name.includes(keyword) || initials.includes(keyword)
  })
})

const dialogVisible = ref(false)
const dialogVisibleName = ref('')
// 全局断点
const { isSmOrBelow, isMdOrBelow } = useBreakpoints()

const openTraffic = (row: BaseProxy) => {
  dialogVisibleName.value = row.name
  dialogVisible.value = true
}

const openManage = (row: BaseProxy) => {
  const hostname = window.location.hostname
  const target = `http://${hostname}:${row.port}`
  window.open(target, '_blank')
}

// const formatTrafficIn = (row: BaseProxy, _: TableColumnCtx<BaseProxy>) => {
//   return Humanize.fileSize(row.trafficIn)
// }

// const formatTrafficOut = (row: BaseProxy, _: TableColumnCtx<BaseProxy>) => {
//   return Humanize.fileSize(row.trafficOut)
// }

const clearOfflineProxies = () => {
  fetch('../api/proxies?status=offline', {
    method: 'DELETE',
    credentials: 'include',
  })
    .then((res) => {
      if (res.ok) {
        ElMessage({
          message: '已成功清除离线代理',
          type: 'success',
        })
        emit('refresh')
      } else {
        ElMessage({
          message: '清除离线代理失败：' + res.status + ' ' + res.statusText,
          type: 'warning',
        })
      }
    })
    .catch((err) => {
      ElMessage({
        message: '清除离线代理失败：' + err.message,
        type: 'warning',
      })
    })
}
</script>

<style scoped>
/* ProxyView 容器样式 - 限制最大高度 */
.proxy-view-container {
  display: flex;
  flex-direction: column;
  /* 64px header + 16px*2 app-main padding + 24px*2 content-container padding */
  max-height: calc(100vh - 64px - 32px - 48px);
  overflow: hidden;
}

.proxy-table {
  flex: 1;
  min-height: 0;
}

/* 名称筛选器样式 */
.name-filter-header {
  display: flex;
  align-items: center;
  gap: 8px;
}

.name-filter-input {
  width: 100px;
}

.name-filter-input :deep(.el-input__inner) {
  height: 26px;
}

/* 页面头部样式 */
.page-header {
  margin-bottom: 24px;
  padding: 0 16px;
}

.header-title {
  font-size: 24px;
  font-weight: 600;
  color: #2d3748;
  margin: 0;
}

html.dark .header-title {
  color: #e2e8f0;
}

.header-actions {
  display: flex;
  gap: 12px;
  align-items: center;
}

/* 表格容器样式 */
.table-container {
  background: white;
  border-radius: 12px;
  overflow: hidden;
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.08);
  margin: 0 16px;
}

html.dark .table-container {
  background: #2d3748;
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.3);
}

/* 表格样式 */
.el-table {
  --el-table-border-color: #e2e8f0;
  --el-table-header-bg-color: #f7fafc;
  --el-table-row-hover-bg-color: #f1f5f9;
}

.el-table :deep(.cell) {
  padding: 0;
}

html.dark .el-table {
  --el-table-border-color: #4a5568;
  --el-table-header-bg-color: #2d3748;
  --el-table-row-hover-bg-color: #4a5568;
}

.el-table :deep(.el-table__header) th {
  background: var(--el-table-header-bg-color);
  color: #4a5568;
  font-weight: 600;
  font-size: 14px;
  /* padding: 16px 12px; */
}

html.dark .el-table :deep(.el-table__header) th {
  color: #cbd5e0;
}

.el-table :deep(.el-table__row) td {
  padding: 14px 0;
  font-size: 14px;
  color: #4a5568;
}

html.dark .el-table :deep(.el-table__row) td {
  color: #e2e8f0;
}

.el-table :deep(.el-table__row):hover td {
  background: var(--el-table-row-hover-bg-color);
}

/* 状态标签样式 */
.status-tag {
  font-weight: 500;
  padding: 4px 12px;
  border-radius: 20px;
  font-size: 12px;
}

.status-online {
  background: #48bb78;
  color: white;
}

.status-offline {
  background: #f56565;
  color: white;
}

/* 操作按钮样式 */
.action-button {
  border-radius: 8px;
  font-weight: 500;
  padding: 8px 16px;
  transition: all 0.2s ease;
}

.action-button:hover {
  transform: translateY(-1px);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
}

/* 对话框样式 */
.traffic-dialog {
  border-radius: 12px;
}

.traffic-dialog :deep(.el-dialog__header) {
  padding: 24px;
  border-bottom: 1px solid #e2e8f0;
}

html.dark .traffic-dialog :deep(.el-dialog__header) {
  border-bottom-color: #4a5568;
}

.traffic-dialog :deep(.el-dialog__title) {
  font-size: 18px;
  font-weight: 600;
  color: #2d3748;
}

html.dark .traffic-dialog :deep(.el-dialog__title) {
  color: #e2e8f0;
}

.traffic-dialog :deep(.el-dialog__body) {
  padding: 24px;
}

/* 响应式设计 */
@media (max-width: 768px) {
  .page-header {
    padding: 0 8px;
    flex-direction: column;
    gap: 16px;
  }

  .header-title {
    font-size: 20px;
  }

  .header-actions {
    width: 100%;
    justify-content: space-between;
  }

  .table-container {
    margin: 0 8px;
    border-radius: 8px;
  }

  .el-table :deep(.el-link__inner) {
    font-size: 12px;
  }
  .el-table :deep(.el-table__header) th,
  .el-table :deep(.el-table__row) td {
    padding: 10px 8px;
    font-size: 12px;
  }

  /* 隐藏部分列在移动端 */
  .el-table :deep(.el-table__column--clientVersion),
  .el-table :deep(.el-table__column--trafficIn),
  .el-table :deep(.el-table__column--trafficOut) {
    display: none;
  }

  .traffic-dialog {
    width: 95% !important;
    margin: 20px auto;
  }

  .traffic-dialog :deep(.el-dialog__header),
  .traffic-dialog :deep(.el-dialog__body) {
    padding: 16px;
  }
}

@media (max-width: 480px) {
  .header-actions {
    flex-direction: column;
    gap: 8px;
  }

  .action-button {
    width: 100%;
    justify-content: center;
  }

  /* 进一步隐藏列在超小屏幕 */
  .el-table :deep(.el-table__column--port) {
    display: none;
  }
}

/* 加载动画 */
.el-table :deep(.el-table__empty-block) {
  padding: 40px 0;
}

.el-table :deep(.el-loading-mask) {
  background: rgba(255, 255, 255, 0.8);
  backdrop-filter: blur(4px);
}

html.dark .el-table :deep(.el-loading-mask) {
  background: rgba(45, 55, 72, 0.8);
}

/* 表格行展开动画 */
.el-table :deep(.el-table__expanded-cell) {
  transition: all 0.3s ease;
  background: #f8fafc;
}

html.dark .el-table :deep(.el-table__expanded-cell) {
  background: #2d3748;
}
</style>

<style>
/* 全局表格样式覆盖 */
.el-table {
  border-radius: 12px;
  overflow: hidden;
}

.el-table .el-table__cell {
  border-bottom: 1px solid #e2e8f0;
}

html.dark .el-table .el-table__cell {
  border-bottom: 1px solid #4a5568;
}

.el-table .el-table__inner-wrapper::before {
  display: none;
}

/* 顶部加载条（NProgress 风格） */
.top-loading {
  position: sticky;
  top: 0;
  left: 0;
  width: 100%;
  height: 2px;
  z-index: 20;
  opacity: 1;
  transition: opacity 220ms ease;
  pointer-events: none;
}

.top-loading__bar {
  position: absolute;
  left: 0;
  top: 0;
  height: 100%;
  width: 18%;
  background: var(--el-color-primary);
  border-radius: 999px;
  animation: topLoadingSlide 1.8s ease-in-out infinite;
}

.top-loading__peg {
  position: absolute;
  top: 0;
  height: 100%;
  width: 30px;
  box-shadow:
    0 0 8px var(--el-color-primary),
    0 0 2px var(--el-color-primary);
  opacity: 0.6;
  transform: rotate(3deg) translate(0, -1px);
  /* 跟随进度条右端移动 */
  left: -10%;
  animation: topLoadingPegSlide 1.8s ease-in-out infinite;
}

@keyframes topLoadingSlide {
  0% {
    left: -10%;
    width: 12%;
  }
  40% {
    left: 30%;
    width: 28%;
  }
  100% {
    left: 100%;
    width: 20%;
  }
}

/* 使拖尾位置与进度条左边缘同步（尾巴跟随条的起点） */
@keyframes topLoadingPegSlide {
  0% {
    left: -10%;
  }
  40% {
    left: 30%;
  }
  100% {
    left: 100%;
  }
}

/* 适配深色模式，沿用主色变量 */
html.dark .top-loading__bar {
  background: var(--el-color-primary);
}

html.dark .top-loading__peg {
  box-shadow:
    0 0 8px var(--el-color-primary),
    0 0 2px var(--el-color-primary);
}

/* 隐藏态：淡出并暂停动画，避免闪烁和性能浪费 */
.top-loading.is-hidden {
  opacity: 0;
}
.top-loading.is-hidden .top-loading__bar,
.top-loading.is-hidden .top-loading__peg {
  /* 取消动画，显示时会重新赋予动画，从起点开始 */
  animation: none;
}
</style>
