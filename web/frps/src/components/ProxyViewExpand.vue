<template>
  <el-form
    label-position="left"
    label-width="auto"
    inline
    class="proxy-table-expand"
  >
    <el-form-item label="Name">
      <span>{{ row.name }}</span>
    </el-form-item>
    <el-form-item label="Type">
      <span>{{ row.type }}</span>
    </el-form-item>
    <el-form-item label="Encryption">
      <span>{{ row.encryption }}</span>
    </el-form-item>
    <el-form-item label="Compression">
      <span>{{ row.compression }}</span>
    </el-form-item>
    <el-form-item label="Last Start">
      <span>{{ row.lastStartTime }}</span>
    </el-form-item>
    <el-form-item label="Last Close">
      <span>{{ row.lastCloseTime }}</span>
    </el-form-item>

    <div v-if="proxyType === 'http' || proxyType === 'https'">
      <el-form-item label="Domains">
        <span>{{ row.customDomains }}</span>
      </el-form-item>
      <el-form-item label="SubDomain">
        <span>{{ row.subdomain }}</span>
      </el-form-item>
      <el-form-item label="locations">
        <span>{{ row.locations }}</span>
      </el-form-item>
      <el-form-item label="HostRewrite">
        <span>{{ row.hostHeaderRewrite }}</span>
      </el-form-item>
    </div>
    <div v-else-if="proxyType === 'tcpmux'">
      <el-form-item label="Multiplexer">
        <span>{{ row.multiplexer }}</span>
      </el-form-item>
      <el-form-item label="RouteByHTTPUser">
        <span>{{ row.routeByHTTPUser }}</span>
      </el-form-item>
      <el-form-item label="Domains">
        <span>{{ row.customDomains }}</span>
      </el-form-item>
      <el-form-item label="SubDomain">
        <span>{{ row.subdomain }}</span>
      </el-form-item>
    </div>
    <div v-else>
      <el-form-item label="Addr">
        <span>{{ row.addr }}</span>
      </el-form-item>
    </div>
  </el-form>

  <div v-if="row.annotations && row.annotations.size > 0">
  <el-divider />
  <el-text class="title-text" size="large">Annotations</el-text>
  <ul>
    <li v-for="item in annotationsArray()" :key="item.key">
      <span class="annotation-key">{{ item.key }}</span>
      <span>{{  item.value }}</span>
    </li>
  </ul>
  </div>
</template>

<script setup lang="ts">

const props = defineProps<{
  row: any
  proxyType: string
}>()

// annotationsArray returns an array of key-value pairs from the annotations map.
const annotationsArray = (): Array<{ key: string; value: string }> => {
  const array: Array<{ key: string; value: any }> = [];
  if (props.row.annotations) {
    props.row.annotations.forEach((value: any, key: string) => {
      array.push({ key, value });
    });
  }
  return array;
}
</script>

<style scoped>
/* 表单样式 */
.proxy-table-expand {
  padding: 20px;
  background: #f8fafc;
  border-radius: 8px;
  margin: 16px;
}

html.dark .proxy-table-expand {
  background: #2d3748;
}

.el-form-item {
  margin-bottom: 16px;
  margin-right: 32px;
  min-width: 200px;
}

.el-form-item :deep(.el-form-item__label) {
  color: #64748b;
  font-weight: 600;
  font-size: 14px;
  padding-bottom: 4px;
}

html.dark .el-form-item :deep(.el-form-item__label) {
  color: #94a3b8;
}

.el-form-item :deep(.el-form-item__content) {
  color: #334155;
  font-size: 14px;
  font-weight: 500;
}

html.dark .el-form-item :deep(.el-form-item__content) {
  color: #e2e8f0;
}

/* 分隔线样式 */
.el-divider {
  margin: 24px 0;
  background-color: #e2e8f0;
}

html.dark .el-divider {
  background-color: #4a5568;
}

/* 注解标题样式 */
.title-text {
  color: #475569;
  font-size: 16px;
  font-weight: 600;
  margin-bottom: 16px;
  display: block;
}

html.dark .title-text {
  color: #cbd5e0;
}

/* 注解列表样式 */
ul {
  list-style-type: none;
  padding: 0;
  margin: 0;
  background: white;
  border-radius: 8px;
  border: 1px solid #e2e8f0;
  overflow: hidden;
}

html.dark ul {
  background: #374151;
  border-color: #4b5563;
}

ul li {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px 16px;
  border-bottom: 1px solid #f1f5f9;
  transition: background-color 0.2s ease;
}

html.dark ul li {
  border-bottom-color: #4b5563;
}

ul li:last-child {
  border-bottom: none;
}

ul li:hover {
  background: #f8fafc;
}

html.dark ul li:hover {
  background: #4b5563;
}

.annotation-key {
  font-weight: 600;
  color: #475569;
  font-size: 14px;
  min-width: 120px;
}

html.dark .annotation-key {
  color: #cbd5e0;
}

ul li span:not(.annotation-key) {
  color: #64748b;
  font-size: 14px;
  flex: 1;
  text-align: right;
}

html.dark ul li span:not(.annotation-key) {
  color: #94a3b8;
}

/* 响应式设计 */
@media (max-width: 768px) {
  .proxy-table-expand {
    padding: 16px;
    margin: 12px;
  }

  .el-form-item {
    margin-right: 16px;
    min-width: 160px;
  }

  ul li {
    flex-direction: column;
    align-items: flex-start;
    gap: 8px;
  }

  .annotation-key {
    min-width: auto;
  }

  ul li span:not(.annotation-key) {
    text-align: left;
    width: 100%;
  }
}

@media (max-width: 480px) {
  .proxy-table-expand {
    padding: 12px;
    margin: 8px;
  }

  .el-form-item {
    margin-right: 0;
    min-width: 100%;
    margin-bottom: 12px;
  }

  .el-form-item :deep(.el-form-item__label) {
    width: 100px !important;
  }

  .el-form-item :deep(.el-form-item__content) {
    margin-left: 40px !important;
  }
}
</style>
