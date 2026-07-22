<template>
  <div class="clients-page">
    <!-- Animated background orbs (mobile) -->
    <div v-if="isMobile" class="m-bg-orbs">
      <div class="m-orb m-orb-1"></div>
      <div class="m-orb m-orb-2"></div>
    </div>

    <div class="page-header">
      <div class="header-top">
        <div class="title-section">
          <h1 class="page-title">{{ $t('clients.title') }}</h1>
          <p class="page-subtitle">{{ $t('clients.subtitle') }}</p>
        </div>
        <div class="status-tabs">
          <button
            v-for="tab in statusTabs"
            :key="tab.value"
            class="status-tab"
            :class="[tab.value, { active: statusFilter === tab.value }]"
            @click="statusFilter = tab.value"
          >
            <span class="status-dot" :class="tab.value"></span>
            <span class="tab-label">{{ tab.label }}</span>
            <span v-if="tab.count !== null" class="tab-count">{{
              tab.count
            }}</span>
          </button>
        </div>
      </div>

      <div class="search-section">
        <el-input
          v-model="searchText"
          :placeholder="$t('clients.searchPlaceholder')"
          :prefix-icon="Search"
          clearable
          class="search-input"
        />
      </div>
    </div>

    <div v-loading="loading" class="clients-content">
      <div v-if="clients.length > 0" class="clients-list">
        <ClientCard
          v-for="(client, index) in clients"
          :key="client.key"
          :client="client"
          :index="index"
        />
      </div>
      <div v-else-if="!loading" class="empty-state">
        <el-empty :description="$t('clients.noClients')" />
      </div>
    </div>

    <div v-if="total > 0" class="pagination-section">
      <ElPagination
        :current-page="page"
        :page-size="pageSize"
        :page-sizes="[10, 20, 50, 100]"
        :total="total"
        :layout="isMobile ? 'prev, pager, next' : 'total, sizes, prev, pager, next'"
        :size="isMobile ? 'small' : 'default'"
        :pager-count="isMobile ? 5 : 7"
        @current-change="onPageChange"
        @size-change="onPageSizeChange"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted, watch } from 'vue'
import { ElMessage, ElPagination } from 'element-plus'
import { Search } from '@element-plus/icons-vue'
import { useI18n } from 'vue-i18n'
import { useResponsive } from '../composables/useResponsive'
import { Client } from '../utils/client'
import ClientCard from '../components/ClientCard.vue'
import { getClientsV2 } from '../api/client'

const { t } = useI18n()
const { isMobile } = useResponsive()

const clients = ref<Client[]>([])
const loading = ref(false)
const searchText = ref('')
const statusFilter = ref<'all' | 'online' | 'offline'>(
  isMobile.value ? 'online' : 'all',
)
const page = ref(1)
const pageSize = ref(10)
const total = ref(0)

let refreshTimer: number | null = null
let searchDebounceTimer: number | null = null
let requestSeq = 0

const statusTabs = computed(() => {
  const allTabs = [
    {
      value: 'all' as const,
      label: t('clients.all'),
      count: statusFilter.value === 'all' ? total.value : null,
    },
    {
      value: 'online' as const,
      label: t('common.online'),
      count: statusFilter.value === 'online' ? total.value : null,
    },
    {
      value: 'offline' as const,
      label: t('common.offline'),
      count: statusFilter.value === 'offline' ? total.value : null,
    },
  ]
  return isMobile.value ? allTabs.slice(1) : allTabs
})

const fetchData = async (silent = false) => {
  const seq = ++requestSeq
  if (!silent) loading.value = true
  try {
    const data = await getClientsV2({
      page: page.value,
      pageSize: pageSize.value,
      status: statusFilter.value,
      q: searchText.value.trim(),
    })
    if (seq !== requestSeq) return

    const maxPage = Math.max(1, Math.ceil(data.total / data.pageSize))
    if (data.items.length === 0 && data.total > 0 && data.page > maxPage) {
      page.value = maxPage
      await fetchData(silent)
      return
    }

    clients.value = data.items.map((item) => new Client(item))
    total.value = data.total
    page.value = data.page
    pageSize.value = data.pageSize
  } catch (error: any) {
    if (seq !== requestSeq) return
    ElMessage({
      showClose: true,
      message: t('clients.fetchFailed', { msg: error.message }),
      type: 'error',
    })
  } finally {
    if (seq === requestSeq) {
      loading.value = false
    }
  }
}

const clearSearchDebounce = () => {
  if (searchDebounceTimer !== null) {
    window.clearTimeout(searchDebounceTimer)
    searchDebounceTimer = null
  }
}

const resetPageAndFetch = () => {
  clearSearchDebounce()
  page.value = 1
  fetchData()
}

const onPageChange = (value: number) => {
  clearSearchDebounce()
  page.value = value
  fetchData()
}

const onPageSizeChange = (value: number) => {
  pageSize.value = value
  resetPageAndFetch()
}

const startAutoRefresh = () => {
  refreshTimer = window.setInterval(() => {
    fetchData(true)
  }, 5000)
}

const stopAutoRefresh = () => {
  if (refreshTimer !== null) {
    window.clearInterval(refreshTimer)
    refreshTimer = null
  }
}

watch(statusFilter, () => {
  resetPageAndFetch()
})

watch(searchText, () => {
  clearSearchDebounce()
  page.value = 1
  searchDebounceTimer = window.setTimeout(() => {
    searchDebounceTimer = null
    fetchData()
  }, 300)
})

onMounted(() => {
  fetchData()
  startAutoRefresh()
})

onUnmounted(() => {
  stopAutoRefresh()
  clearSearchDebounce()
})
</script>

<style scoped>
.clients-page {
  display: flex;
  flex-direction: column;
  gap: 24px;
}

.page-header {
  display: flex;
  flex-direction: column;
  gap: 24px;
}

.header-top {
  display: flex;
  justify-content: space-between;
  align-items: flex-end;
  gap: 20px;
  flex-wrap: wrap;
}

.title-section {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.page-title {
  font-size: 28px;
  font-weight: 600;
  color: var(--el-text-color-primary);
  margin: 0;
  line-height: 1.2;
}

.page-subtitle {
  font-size: 14px;
  color: var(--el-text-color-secondary);
  margin: 0;
}

.status-tabs {
  display: inline-flex;
  gap: 3px;
  padding: 3px;
  background: var(--el-fill-color-light);
  border: 1px solid var(--el-border-color-lighter);
  border-radius: 10px;
}

.status-tab {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 5px 12px;
  border: 1px solid transparent;
  border-radius: 8px;
  background: transparent;
  color: var(--el-text-color-secondary);
  font-size: 12px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.25s ease;
  white-space: nowrap;
}

.status-tab:hover {
  color: var(--el-text-color-primary);
  background: var(--el-fill-color);
}

.status-tab.active {
  background: var(--el-bg-color);
  color: var(--el-text-color-primary);
  box-shadow: 0 1px 4px rgba(0, 0, 0, 0.08), 0 0 1px rgba(0, 0, 0, 0.06);
}

/* Per-tab accent colors */
.status-tab.all.active {
  background: rgba(64, 158, 255, 0.12);
  color: #337ecc;
  border-color: rgba(64, 158, 255, 0.25);
}

.status-tab.online.active {
  background: rgba(103, 194, 58, 0.12);
  color: #529b2e;
  border-color: rgba(103, 194, 58, 0.25);
}

.status-tab.offline.active {
  background: rgba(144, 147, 153, 0.12);
  color: #73767a;
  border-color: rgba(144, 147, 153, 0.25);
}

.status-tab.active .status-dot.online {
  box-shadow: none;
}

.status-tab.active .status-dot.all {
  box-shadow: none;
}

.status-dot {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background-color: var(--el-text-color-secondary);
  transition: box-shadow 0.25s ease;
}

.status-dot.online {
  background-color: var(--el-color-success);
}

.status-tab.active .status-dot.online {
  box-shadow: 0 0 6px rgba(103, 194, 58, 0.6);
}

.status-dot.offline {
  background-color: var(--el-text-color-placeholder);
}

.status-dot.all {
  background-color: var(--el-color-primary);
}

.status-tab.active .status-dot.all {
  box-shadow: 0 0 6px var(--el-color-primary-light-5);
}

.tab-count {
  font-size: 11px;
  font-weight: 600;
  min-width: 18px;
  height: 18px;
  line-height: 18px;
  text-align: center;
  padding: 0 5px;
  border-radius: 9px;
  background: var(--el-fill-color-dark);
  color: var(--el-text-color-secondary);
}

.status-tab.active .tab-count {
  background: rgba(0, 0, 0, 0.07);
  color: inherit;
}

html.dark .status-tab.active .tab-count {
  background: rgba(255, 255, 255, 0.12);
}

.search-section {
  width: 100%;
}

.search-input :deep(.el-input__wrapper) {
  border-radius: 12px;
  box-shadow: 0 1px 2px rgba(0, 0, 0, 0.04);
  padding: 8px 16px;
  /* border: 1px solid var(--el-border-color); */
  transition: all 0.2s;
  height: 48px;
  font-size: 15px;
  /* background: none; */
}

.search-input :deep(.el-input__wrapper:hover) {
  border-color: var(--el-border-color-darker);
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.06);
}

.search-input :deep(.el-input__wrapper.is-focus) {
  border-color: var(--el-color-primary);
  box-shadow: 0 0 0 1px var(--el-color-primary);
}

.clients-content {
  min-height: 200px;
}

.clients-list {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.empty-state {
  padding: 60px 0;
}

.pagination-section {
  display: flex;
  justify-content: flex-end;
}

/* Dark mode adjustments */
html.dark .status-tabs {
  background: var(--el-fill-color);
}

html.dark .status-tab.all.active {
  background: rgba(64, 158, 255, 0.16);
  color: #79bbff;
  border-color: rgba(64, 158, 255, 0.35);
}

html.dark .status-tab.online.active {
  background: rgba(103, 194, 58, 0.14);
  color: #95d475;
  border-color: rgba(103, 194, 58, 0.35);
}

html.dark .status-tab.offline.active {
  background: rgba(255, 255, 255, 0.08);
  color: #b1b3b8;
  border-color: rgba(255, 255, 255, 0.15);
}

/* ===== Keyframe animations ===== */
@keyframes m-float-orb {
  0%, 100% { transform: translate(0, 0) scale(1); }
  33% { transform: translate(12px, -18px) scale(1.05); }
  66% { transform: translate(-8px, 8px) scale(0.95); }
}

@keyframes m-breathe {
  0%, 100% { opacity: 0.6; transform: scale(1); }
  50% { opacity: 1; transform: scale(1.2); }
}

@keyframes m-fade-up {
  from { opacity: 0; transform: translateY(14px); }
  to { opacity: 1; transform: translateY(0); }
}

/* Animated background orbs */
.m-bg-orbs {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  height: 280px;
  pointer-events: none;
  overflow: hidden;
  z-index: 0;
}

.m-orb {
  position: absolute;
  border-radius: 50%;
  filter: blur(60px);
  animation: m-float-orb 9s ease-in-out infinite;
}

.m-orb-1 {
  width: 160px;
  height: 160px;
  top: -30px;
  right: -20px;
  background: radial-gradient(circle, rgba(103, 194, 58, 0.2), transparent 70%);
}

.m-orb-2 {
  width: 130px;
  height: 130px;
  top: 80px;
  left: -40px;
  background: radial-gradient(circle, rgba(64, 158, 255, 0.16), transparent 70%);
  animation-delay: -4s;
}

html.dark .m-orb-1 {
  background: radial-gradient(circle, rgba(103, 194, 58, 0.14), transparent 70%);
}

html.dark .m-orb-2 {
  background: radial-gradient(circle, rgba(64, 158, 255, 0.12), transparent 70%);
}

@media (max-width: 767px) {
  .clients-page {
    gap: 12px;
    position: relative;
  }

  .page-header {
    gap: 12px;
    position: relative;
    z-index: 1;
    animation: m-fade-up 0.4s ease-out both;
  }

  .header-top {
    flex-direction: column;
    align-items: flex-start;
    gap: 10px;
  }

  .page-title {
    font-size: 22px;
    background: linear-gradient(135deg, var(--el-text-color-primary) 60%, #67c23a);
    -webkit-background-clip: text;
    -webkit-text-fill-color: transparent;
    background-clip: text;
  }

  .page-subtitle {
    display: none;
  }

  .status-tabs {
    width: 100%;
    overflow-x: auto;
    background: rgba(255, 255, 255, 0.6);
    backdrop-filter: blur(10px);
    -webkit-backdrop-filter: blur(10px);
    border-color: rgba(255, 255, 255, 0.5);
    box-shadow: 0 2px 12px rgba(0, 0, 0, 0.04);
  }

  html.dark .status-tabs {
    background: rgba(39, 41, 61, 0.7);
    border-color: rgba(52, 54, 77, 0.5);
  }

  .status-tab {
    flex: 1;
    justify-content: center;
  }

  .status-tab.active .status-dot.online {
    animation: m-breathe 2s ease-in-out infinite;
  }

  .search-section {
    position: relative;
    z-index: 1;
  }

  .search-input :deep(.el-input__wrapper) {
    height: 38px;
    font-size: 14px;
    padding: 4px 12px;
    border-radius: 12px;
    background: rgba(255, 255, 255, 0.65);
    backdrop-filter: blur(10px);
    -webkit-backdrop-filter: blur(10px);
    box-shadow: 0 2px 12px rgba(0, 0, 0, 0.04);
  }

  html.dark .search-input :deep(.el-input__wrapper) {
    background: rgba(39, 41, 61, 0.7);
  }

  .clients-content {
    position: relative;
    z-index: 1;
  }

  .clients-list {
    gap: 8px;
  }

  .pagination-section {
    justify-content: center;
    position: relative;
    z-index: 1;
  }
}
</style>
