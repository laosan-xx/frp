<template>
  <div class="client-detail-page">
    <!-- Animated background orbs (mobile) -->
    <div v-if="isMobile" class="m-bg-orbs">
      <div class="m-orb m-orb-1"></div>
      <div class="m-orb m-orb-2"></div>
    </div>

    <!-- Breadcrumb -->
    <nav class="breadcrumb">
      <a class="breadcrumb-link" @click="goBack">
        <el-icon><ArrowLeft /></el-icon>
      </a>
      <router-link to="/clients" class="breadcrumb-item">{{ $t('nav.clients') }}</router-link>
      <span class="breadcrumb-separator">/</span>
      <span class="breadcrumb-current">{{
        client?.displayName || (loading ? $t('common.loading') : route.params.key)
      }}</span>
    </nav>

    <div v-loading="loading" class="detail-content">
      <template v-if="client">
        <!-- Header Card -->
        <div class="header-card">
          <div class="header-main">
            <div class="header-left">
              <div class="client-avatar">
                {{ client.displayName.charAt(0).toUpperCase() }}
              </div>
              <div class="client-info">
                <div class="client-name-row">
                  <h1 class="client-name">{{ client.displayName }}</h1>
                  <el-tag v-if="client.version" size="small" type="success"
                    >v{{ client.version }}</el-tag
                  >
                  <el-tag v-if="client.wireProtocolLabel" size="small" type="info">
                    {{ client.wireProtocolLabel }}
                  </el-tag>
                </div>
                <div class="client-meta">
                  <span v-if="client.ip" class="meta-item">{{
                    client.ip
                  }}</span>
                  <span v-if="client.ipRegion" class="meta-item">{{
                    client.ipRegion
                  }}</span>
                  <span v-if="client.hostname" class="meta-item">{{
                    client.hostname
                  }}</span>
                </div>
              </div>
            </div>
            <div class="header-right">
              <span
                class="status-badge"
                :class="client.online ? 'online' : 'offline'"
              >
                {{ client.online ? $t('common.online') : $t('common.offline') }}
              </span>
            </div>
          </div>

          <!-- Info Section -->
          <div class="info-section">
            <div class="info-item">
              <span class="info-label">{{ $t('clientDetail.connections') }}</span>
              <span class="info-value">{{ client.status.curConns }}</span>
            </div>
            <!-- <div class="info-item">
              <span class="info-label">{{ $t('clientDetail.runId') }}</span>
              <span class="info-value">{{ client.runID }}</span>
            </div> -->
            <div v-if="client.wireProtocol" class="info-item">
              <span class="info-label">{{ $t('clientDetail.protocol') }}</span>
              <span class="info-value">{{ client.wireProtocol }}</span>
            </div>
            <div class="info-item">
              <span class="info-label">{{ $t('clientDetail.firstConnected') }}</span>
              <span class="info-value">{{ client.firstConnectedAgo }}</span>
            </div>
            <div class="info-item">
              <span class="info-label">{{
                client.online ? $t('clientDetail.connected') : $t('clientDetail.disconnected')
              }}</span>
              <span class="info-value">{{
                client.online ? client.lastConnectedAgo : client.disconnectedAgo
              }}</span>
            </div>
          </div>
        </div>

        <!-- Proxies Card -->
        <div class="proxies-card">
          <div class="proxies-header">
            <div class="proxies-title">
              <h2>{{ $t('clientDetail.proxies') }}</h2>
              <span class="proxies-count">{{ total }}</span>
            </div>
            <el-input
              v-model="proxySearch"
              :placeholder="$t('clientDetail.searchProxies')"
              :prefix-icon="Search"
              clearable
              class="proxy-search"
            />
          </div>
          <div class="proxies-body">
            <div v-if="proxiesLoading" class="loading-state">
              <el-icon class="is-loading"><Loading /></el-icon>
              <span>{{ $t('common.loading') }}</span>
            </div>
            <div v-else-if="proxies.length > 0" class="proxies-list">
              <ProxyCard
                v-for="proxy in proxies"
                :key="`${proxy.type}:${proxy.name}`"
                :proxy="proxy"
                show-type
              />
            </div>
            <div v-else-if="proxySearch.trim()" class="empty-state">
              <p>{{ $t('clientDetail.noProxiesMatch', { query: proxySearch }) }}</p>
            </div>
            <div v-else class="empty-state">
              <p>{{ $t('clientDetail.noProxies') }}</p>
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

      <div v-else-if="!loading" class="not-found">
        <h2>{{ $t('clientDetail.notFound') }}</h2>
        <p>{{ $t('clientDetail.notFoundDesc') }}</p>
        <router-link to="/clients">
          <el-button type="primary">{{ $t('clientDetail.backToClients') }}</el-button>
        </router-link>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage, ElPagination } from 'element-plus'
import { ArrowLeft, Loading, Search } from '@element-plus/icons-vue'
import { useI18n } from 'vue-i18n'
import { useResponsive } from '../composables/useResponsive'
import { Client } from '../utils/client'
import { getClientV2 } from '../api/client'
import { getProxiesV2 } from '../api/proxy'
import {
  BaseProxy,
  TCPProxy,
  UDPProxy,
  HTTPProxy,
  HTTPSProxy,
  TCPMuxProxy,
  STCPProxy,
  SUDPProxy,
} from '../utils/proxy'
import { getServerInfo } from '../api/server'
import ProxyCard from '../components/ProxyCard.vue'
import type { ProxyStatsInfo } from '../types/proxy'
import type { ServerInfo } from '../types/server'

const route = useRoute()
const router = useRouter()
const { t } = useI18n()
const { isMobile } = useResponsive()
const client = ref<Client | null>(null)
const loading = ref(true)

const goBack = () => {
  if (window.history.length > 1) {
    router.back()
  } else {
    router.push('/clients')
  }
}
const proxiesLoading = ref(false)
const proxies = ref<BaseProxy[]>([])
const proxySearch = ref('')
const page = ref(1)
const pageSize = ref(10)
const total = ref(0)
let requestSeq = 0
let searchDebounceTimer: number | null = null

let serverInfoPromise: Promise<ServerInfo> | null = null

const fetchServerInfo = (): Promise<ServerInfo> => {
  if (!serverInfoPromise) {
    serverInfoPromise = getServerInfo().catch((err) => {
      serverInfoPromise = null
      throw err
    })
  }
  return serverInfoPromise
}

const fetchClient = async (): Promise<boolean> => {
  const key = route.params.key as string
  if (!key) {
    loading.value = false
    return false
  }
  try {
    const data = await getClientV2(key)
    client.value = new Client(data)
    return true
  } catch (error: any) {
    ElMessage.error(t('clientDetail.fetchFailed', { msg: error.message }))
    return false
  } finally {
    loading.value = false
  }
}

const convertProxy = async (
  proxy: ProxyStatsInfo,
): Promise<BaseProxy | null> => {
  const type = proxy.type || ''
  if (type === 'tcp') {
    return new TCPProxy(proxy)
  }
  if (type === 'udp') {
    return new UDPProxy(proxy)
  }
  if (type === 'http') {
    const info = await fetchServerInfo()
    if (info && info.config.vhostHTTPPort) {
      return new HTTPProxy(
        proxy,
        info.config.vhostHTTPPort,
        info.config.subdomainHost,
      )
    }
    return null
  }
  if (type === 'https') {
    const info = await fetchServerInfo()
    if (info && info.config.vhostHTTPSPort) {
      return new HTTPSProxy(
        proxy,
        info.config.vhostHTTPSPort,
        info.config.subdomainHost,
      )
    }
    return null
  }
  if (type === 'tcpmux') {
    const info = await fetchServerInfo()
    if (info && info.config.tcpmuxHTTPConnectPort) {
      return new TCPMuxProxy(
        proxy,
        info.config.tcpmuxHTTPConnectPort,
        info.config.subdomainHost,
      )
    }
    return null
  }
  if (type === 'stcp') {
    return new STCPProxy(proxy)
  }
  if (type === 'sudp') {
    return new SUDPProxy(proxy)
  }

  const bp = new BaseProxy(proxy)
  bp.type = type
  return bp
}

const convertProxies = async (items: ProxyStatsInfo[]): Promise<BaseProxy[]> => {
  const converted = await Promise.all(items.map((item) => convertProxy(item)))
  return converted.filter((item): item is BaseProxy => item !== null)
}

const fetchProxies = async () => {
  if (!client.value) return
  const seq = ++requestSeq
  proxiesLoading.value = true

  try {
    const q = proxySearch.value.trim()
    const data = await getProxiesV2({
      page: page.value,
      pageSize: pageSize.value,
      q: q || undefined,
      clientID: client.value.clientID,
      user: client.value.user,
    })
    if (seq !== requestSeq) return

    const maxPage = Math.max(1, Math.ceil(data.total / data.pageSize))
    if (data.items.length === 0 && data.total > 0 && data.page > maxPage) {
      page.value = maxPage
      await fetchProxies()
      return
    }

    const converted = await convertProxies(data.items)
    if (seq !== requestSeq) return

    proxies.value = converted
    total.value = data.total
    page.value = data.page
    pageSize.value = data.pageSize
  } catch (error: any) {
    if (seq !== requestSeq) return
    ElMessage.error(t('clientDetail.fetchProxiesFailed', { msg: error.message }))
  } finally {
    if (seq === requestSeq) {
      proxiesLoading.value = false
    }
  }
}

const clearSearchDebounce = () => {
  if (searchDebounceTimer !== null) {
    window.clearTimeout(searchDebounceTimer)
    searchDebounceTimer = null
  }
}

const invalidateProxyRequests = () => {
  requestSeq++
  proxiesLoading.value = false
}

const resetPageAndFetch = () => {
  clearSearchDebounce()
  page.value = 1
  fetchProxies()
}

const onPageChange = (value: number) => {
  clearSearchDebounce()
  page.value = value
  fetchProxies()
}

const onPageSizeChange = (value: number) => {
  pageSize.value = value
  resetPageAndFetch()
}

watch(proxySearch, () => {
  clearSearchDebounce()
  invalidateProxyRequests()
  page.value = 1
  searchDebounceTimer = window.setTimeout(() => {
    searchDebounceTimer = null
    fetchProxies()
  }, 300)
})

onUnmounted(() => {
  clearSearchDebounce()
})

onMounted(async () => {
  const ok = await fetchClient()
  if (!ok || !client.value) return

  fetchProxies()
})
</script>

<style scoped>
.client-detail-page {
}

/* Breadcrumb */
.breadcrumb {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 14px;
  margin-bottom: 24px;
}

.breadcrumb-link {
  display: flex;
  align-items: center;
  color: var(--text-secondary);
  cursor: pointer;
  transition: color 0.2s;
  margin-right: 4px;
}

.breadcrumb-link:hover {
  color: var(--text-primary);
}

.breadcrumb-item {
  color: var(--text-secondary);
  text-decoration: none;
  transition: color 0.2s;
}

.breadcrumb-item:hover {
  color: var(--el-color-primary);
}

.breadcrumb-separator {
  color: var(--el-border-color);
}

.breadcrumb-current {
  color: var(--text-primary);
  font-weight: 500;
}

/* Card Base */
.header-card,
.proxies-card {
  background: var(--el-bg-color);
  border: 1px solid var(--header-border);
  border-radius: 12px;
  margin-bottom: 16px;
}

/* Header Card */
.header-main {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  padding: 24px;
}

.header-left {
  display: flex;
  gap: 16px;
  align-items: center;
}

.client-avatar {
  width: 48px;
  height: 48px;
  border-radius: 12px;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: white;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 20px;
  font-weight: 500;
  flex-shrink: 0;
}

.client-info {
  min-width: 0;
}

.client-name-row {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-bottom: 4px;
}

.client-name {
  font-size: 20px;
  font-weight: 500;
  color: var(--text-primary);
  margin: 0;
  line-height: 1.3;
}

.client-meta {
  display: flex;
  gap: 12px;
  font-size: 14px;
  color: var(--text-secondary);
  flex-wrap: wrap;
}

.status-badge {
  padding: 6px 12px;
  border-radius: 6px;
  font-size: 13px;
  font-weight: 500;
}

.status-badge.online {
  background: rgba(34, 197, 94, 0.1);
  color: #16a34a;
}

.status-badge.offline {
  background: var(--hover-bg);
  color: var(--text-secondary);
}

html.dark .status-badge.online {
  background: rgba(34, 197, 94, 0.15);
  color: #4ade80;
}

.header-right{
  white-space: nowrap;
}

/* Info Section */
.info-section {
  display: flex;
  flex-wrap: wrap;
  gap: 16px 32px;
  padding: 16px 24px;
}

.info-item {
  display: flex;
  align-items: baseline;
  gap: 8px;
}

.info-label {
  font-size: 13px;
  color: var(--text-secondary);
}

.info-label::after {
  content: ':';
}

.info-value {
  font-size: 13px;
  color: var(--text-primary);
  font-weight: 500;
  word-break: break-all;
}

/* Proxies Card */
.proxies-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 16px 20px;
  gap: 16px;
}

.proxies-title {
  display: flex;
  align-items: center;
  gap: 8px;
}

.proxies-title h2 {
  font-size: 15px;
  font-weight: 500;
  color: var(--text-primary);
  margin: 0;
}

.proxies-count {
  font-size: 13px;
  font-weight: 500;
  color: var(--text-secondary);
  background: var(--hover-bg);
  padding: 4px 10px;
  border-radius: 6px;
}

.proxy-search {
  width: 200px;
}

.proxy-search :deep(.el-input__wrapper) {
  border-radius: 6px;
}

.proxies-body {
  padding: 16px;
}

.pagination-section {
  display: flex;
  justify-content: center;
  padding: 0 20px 20px;
}

.proxies-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.loading-state {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  padding: 40px;
  color: var(--text-secondary);
}

.empty-state {
  text-align: center;
  padding: 40px;
  color: var(--text-secondary);
}

.empty-state p {
  margin: 0;
}

/* Not Found */
.not-found {
  text-align: center;
  padding: 60px 20px;
}

.not-found h2 {
  font-size: 18px;
  font-weight: 500;
  color: var(--text-primary);
  margin: 0 0 8px;
}

.not-found p {
  font-size: 14px;
  color: var(--text-secondary);
  margin: 0 0 20px;
}

/* Responsive */

/* ===== Mobile animations & effects ===== */
@keyframes m-float-orb {
  0%, 100% { transform: translate(0, 0) scale(1); }
  33% { transform: translate(10px, -15px) scale(1.04); }
  66% { transform: translate(-8px, 8px) scale(0.96); }
}

@keyframes m-fade-up {
  from { opacity: 0; transform: translateY(14px); }
  to { opacity: 1; transform: translateY(0); }
}

@keyframes m-slide-left {
  from { opacity: 0; transform: translateX(-16px); }
  to { opacity: 1; transform: translateX(0); }
}

@keyframes m-avatar-glow {
  0%, 100% { box-shadow: 0 2px 10px rgba(102, 126, 234, 0.3); }
  50% { box-shadow: 0 4px 18px rgba(102, 126, 234, 0.5); }
}

@keyframes m-status-pulse {
  0%, 100% { opacity: 1; }
  50% { opacity: 0.7; }
}

.m-bg-orbs {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  height: 260px;
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
  width: 150px;
  height: 150px;
  top: -30px;
  right: -20px;
  background: radial-gradient(circle, rgba(102, 126, 234, 0.2), transparent 70%);
}

.m-orb-2 {
  width: 120px;
  height: 120px;
  top: 70px;
  left: -30px;
  background: radial-gradient(circle, rgba(118, 75, 162, 0.15), transparent 70%);
  animation-delay: -4s;
}

html.dark .m-orb-1 {
  background: radial-gradient(circle, rgba(129, 140, 248, 0.15), transparent 70%);
}

html.dark .m-orb-2 {
  background: radial-gradient(circle, rgba(167, 139, 250, 0.12), transparent 70%);
}

@media (max-width: 767px) {
  .breadcrumb {
    margin-bottom: 12px;
    position: relative;
    z-index: 1;
    animation: m-slide-left 0.35s ease-out both;
  }

  .header-card,
  .proxies-card {
    margin-bottom: 10px;
    position: relative;
    z-index: 1;
    background: rgba(255, 255, 255, 0.7);
    backdrop-filter: blur(12px);
    -webkit-backdrop-filter: blur(12px);
    border-color: rgba(255, 255, 255, 0.5);
    box-shadow: 0 2px 14px rgba(0, 0, 0, 0.04);
  }

  .header-card {
    animation: m-fade-up 0.4s ease-out both;
  }

  .proxies-card {
    animation: m-fade-up 0.4s ease-out 0.12s both;
  }

  html.dark .header-card,
  html.dark .proxies-card {
    background: rgba(39, 41, 61, 0.75);
    border-color: rgba(52, 54, 77, 0.5);
    box-shadow: 0 2px 14px rgba(0, 0, 0, 0.15);
  }

  .header-main {
    padding: 14px 16px;
    gap: 10px;
  }

  .header-left {
    gap: 10px;
    min-width: 0;
  }

  .client-avatar {
    width: 36px;
    height: 36px;
    border-radius: 9px;
    font-size: 15px;
    animation: m-avatar-glow 3s ease-in-out infinite;
  }

  .client-name {
    font-size: 16px;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  .client-name-row {
    gap: 6px;
    flex-wrap: nowrap;
    overflow: hidden;
    min-width: 0;
  }

  .client-name-row .el-tag {
    display: none;
  }

  .client-meta {
    font-size: 12px;
    gap: 8px;
    flex-wrap: nowrap;
    overflow: hidden;
  }

  .client-meta .meta-item {
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  .status-badge {
    padding: 3px 8px;
    font-size: 12px;
    flex-shrink: 0;
  }

  .status-badge.online {
    animation: m-status-pulse 2.5s ease-in-out infinite;
  }

  .info-section {
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: 6px 16px;
    padding: 10px 16px;
    border-top: 1px solid var(--header-border);
  }

  .info-label {
    font-size: 12px;
  }

  .info-value {
    font-size: 12px;
  }

  .proxies-header {
    flex-direction: row;
    align-items: center;
    padding: 10px 14px;
    gap: 10px;
  }

  .proxy-search {
    width: 180px;
    flex-shrink: 0;
  }

  .proxies-body {
    padding: 10px;
  }

  .proxies-list {
    gap: 8px;
  }

  .pagination-section {
    padding: 0 14px 14px;
  }
}
</style>
