<template>
  <div>
    <el-page-header
      :icon="null"
      style="width: 100%; margin-left: 30px; margin-bottom: 20px"
    >
      <template #title>
        <span>{{ proxyType }}</span>
      </template>
      <template #content> </template>
      <template #extra>
        <div class="flex items-center" style="margin-right: 30px">
          <el-popconfirm
            title="Are you sure to clear all data of offline proxies?"
            @confirm="clearOfflineProxies"
          >
            <template #reference>
              <el-button>ClearOfflineProxies</el-button>
            </template>
          </el-popconfirm>
          <el-button @click="$emit('refresh')">Refresh</el-button>
        </div>
      </template>
    </el-page-header>

    <el-table
      :data="proxies"
      :default-sort="{ prop: 'name', order: 'ascending' }"
      style="width: 100%"
    >
      <el-table-column type="expand">
        <template #default="props">
          <ProxyViewExpand :row="props.row" :proxyType="proxyType" />
        </template>
      </el-table-column>
      <el-table-column label="Name" prop="name" sortable> </el-table-column>
      <el-table-column label="Port" prop="port" sortable> </el-table-column>
      <el-table-column label="Connections" prop="conns" sortable>
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
      <el-table-column label="ClientVersion" prop="clientVersion" sortable>
      </el-table-column>
      <el-table-column label="Status" prop="status" sortable>
        <template #default="scope">
          <el-tag v-if="scope.row.status === 'online'" type="success">{{
            scope.row.status
          }}</el-tag>
          <el-tag v-else type="danger">{{ scope.row.status }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column label="Operations">
        <template #default="scope">
          <el-button
            type="primary"
            :name="scope.row.name"
            @click="dialogVisibleName = scope.row.name; dialogVisible = true"
            >Traffic
          </el-button>
        </template>
      </el-table-column>
    </el-table>
  </div>

  <el-dialog
    v-model="dialogVisible"
    :destroy-on-close="true"
    :title="dialogVisibleName"
    width="700px">
    <Traffic :proxyName="dialogVisibleName" />
  </el-dialog>
</template>

<script setup lang="ts">
import * as Humanize from 'humanize-plus'
import type { TableColumnCtx } from 'element-plus'
import type { BaseProxy } from '../utils/proxy.js'
import { ElMessage } from 'element-plus'
import ProxyViewExpand from './ProxyViewExpand.vue'
import { ref } from 'vue'

defineProps<{
  proxies: BaseProxy[]
  proxyType: string
}>()

const emit = defineEmits(['refresh'])

const dialogVisible = ref(false)
const dialogVisibleName = ref("")

const formatTrafficIn = (row: BaseProxy, _: TableColumnCtx<BaseProxy>) => {
  return Humanize.fileSize(row.trafficIn)
}

const formatTrafficOut = (row: BaseProxy, _: TableColumnCtx<BaseProxy>) => {
  return Humanize.fileSize(row.trafficOut)
}

const clearOfflineProxies = () => {
  fetch('../api/proxies?status=offline', {
    method: 'DELETE',
    credentials: 'include',
  })
    .then((res) => {
      if (res.ok) {
        ElMessage({
          message: 'Successfully cleared offline proxies',
          type: 'success',
        })
        emit('refresh')
      } else {
        ElMessage({
          message: 'Failed to clear offline proxies: ' + res.status + ' ' + res.statusText,
          type: 'warning',
        })
      }
    })
    .catch((err) => {
      ElMessage({
        message: 'Failed to clear offline proxies: ' + err.message,
        type: 'warning',
      })
    })
}
</script>

<style scoped>
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

  .el-table :deep(.el-table__header) th,
  .el-table :deep(.el-table__row) td {
    padding: 12px 8px;
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
</style>
