<template>
  <div class="server-overview">
    <!-- Mobile: high-density monitoring panel -->
    <template v-if="isMobile">
      <!-- Animated background orbs -->
      <div class="m-bg-orbs">
        <div class="m-orb m-orb-1"></div>
        <div class="m-orb m-orb-2"></div>
        <div class="m-orb m-orb-3"></div>
      </div>

      <!-- Hero stats cards -->
      <div class="m-hero-grid">
        <div class="m-hero-card clients" @click="router.push('/clients')">
          <div class="m-hero-glow"></div>
          <div class="m-hero-icon-wrap">
            <div class="m-hero-icon">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M16 21v-2a4 4 0 0 0-4-4H6a4 4 0 0 0-4 4v2"/><circle cx="9" cy="7" r="4"/><path d="M22 21v-2a4 4 0 0 0-3-3.87"/><path d="M16 3.13a4 4 0 0 1 0 7.75"/></svg>
            </div>
          </div>
          <div class="m-hero-info">
            <div class="m-hero-value">{{ data.status.clientCounts }}</div>
            <div class="m-hero-label">{{ $t('overview.clients') }}</div>
          </div>
          <div class="m-hero-arrow">&rsaquo;</div>
        </div>
        <div class="m-hero-card proxies" @click="router.push('/proxies/tcp')">
          <div class="m-hero-glow"></div>
          <div class="m-hero-icon-wrap">
            <div class="m-hero-icon">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M12 2v4m0 12v4M4.93 4.93l2.83 2.83m8.48 8.48l2.83 2.83M2 12h4m12 0h4M4.93 19.07l2.83-2.83m8.48-8.48l2.83-2.83"/></svg>
            </div>
          </div>
          <div class="m-hero-info">
            <div class="m-hero-value">{{ proxyCounts }}</div>
            <div class="m-hero-label">{{ $t('overview.proxies') }}</div>
          </div>
          <div class="m-hero-arrow">&rsaquo;</div>
        </div>
        <div class="m-hero-card connections">
          <div class="m-hero-glow"></div>
          <div class="m-hero-icon-wrap">
            <div class="m-hero-icon">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M22 12h-4l-3 9L9 3l-3 9H2"/></svg>
            </div>
          </div>
          <div class="m-hero-info">
            <div class="m-hero-value">{{ data.status.curConns }}</div>
            <div class="m-hero-label">{{ $t('overview.connections') }}</div>
          </div>
          <div class="m-hero-pulse-ring"></div>
        </div>
      </div>

      <!-- Traffic card -->
      <div class="m-card m-traffic-card">
        <div class="m-card-header">
          <div class="m-card-header-left">
            <span class="m-card-dot traffic"></span>
            <span class="m-card-title">{{ $t('overview.networkTraffic') }}</span>
          </div>
          <span class="m-card-badge">{{ $t('overview.today') }}</span>
        </div>
        <div class="m-traffic-grid">
          <div class="m-traffic-block inbound">
            <div class="m-traffic-bar-wrap">
              <div class="m-traffic-bar" :style="{ width: trafficInPercent + '%' }">
                <div class="m-traffic-shimmer"></div>
              </div>
            </div>
            <div class="m-traffic-meta">
              <span class="m-traffic-direction">
                <svg viewBox="0 0 16 16" fill="currentColor" width="14" height="14"><path d="M8 1a.5.5 0 0 1 .5.5v11.793l3.146-3.147a.5.5 0 0 1 .708.708l-4 4a.5.5 0 0 1-.708 0l-4-4a.5.5 0 0 1 .708-.708L7.5 13.293V1.5A.5.5 0 0 1 8 1z" transform="rotate(180 8 8)"/></svg>
                {{ $t('overview.inbound') }}
              </span>
              <span class="m-traffic-amount">{{ formatFileSize(data.status.totalTrafficIn) }}</span>
            </div>
          </div>
          <div class="m-traffic-block outbound">
            <div class="m-traffic-bar-wrap">
              <div class="m-traffic-bar" :style="{ width: trafficOutPercent + '%' }">
                <div class="m-traffic-shimmer"></div>
              </div>
            </div>
            <div class="m-traffic-meta">
              <span class="m-traffic-direction">
                <svg viewBox="0 0 16 16" fill="currentColor" width="14" height="14"><path d="M8 1a.5.5 0 0 1 .5.5v11.793l3.146-3.147a.5.5 0 0 1 .708.708l-4 4a.5.5 0 0 1-.708 0l-4-4a.5.5 0 0 1 .708-.708L7.5 13.293V1.5A.5.5 0 0 1 8 1z"/></svg>
                {{ $t('overview.outbound') }}
              </span>
              <span class="m-traffic-amount">{{ formatFileSize(data.status.totalTrafficOut) }}</span>
            </div>
          </div>
        </div>
      </div>

      <!-- Proxy types card -->
      <div class="m-card m-proxy-card">
        <div class="m-card-header">
          <div class="m-card-header-left">
            <span class="m-card-dot proxy"></span>
            <span class="m-card-title">{{ $t('overview.proxyTypes') }}</span>
          </div>
          <span class="m-card-badge">{{ $t('overview.now') }}</span>
        </div>
        <div class="m-type-chips" v-if="hasActiveProxies">
          <span
            v-for="(count, type) in data.status.proxyTypeCount"
            v-show="count > 0"
            :key="type"
            class="m-type-chip"
            :class="'chip-' + String(type).toLowerCase()"
          >
            <span class="m-chip-dot"></span>
            <span class="m-chip-name">{{ type.toUpperCase() }}</span>
            <span class="m-chip-count">{{ count }}</span>
          </span>
        </div>
        <div v-else class="m-empty-state">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" width="28" height="28"><circle cx="12" cy="12" r="10"/><path d="M8 15s1.5-2 4-2 4 2 4 2"/><line x1="9" y1="9" x2="9.01" y2="9"/><line x1="15" y1="9" x2="15.01" y2="9"/></svg>
          <span>{{ $t('overview.noActiveProxies') }}</span>
        </div>
      </div>

      <!-- Server config card -->
      <div class="m-card m-config-card">
        <div class="m-card-header">
          <div class="m-card-header-left">
            <span class="m-card-dot config"></span>
            <span class="m-card-title">{{ $t('overview.serverConfig') }}</span>
          </div>
          <span class="m-card-badge version">v{{ data.version }}</span>
        </div>
        <div class="m-config-grid">
          <div v-for="item in configItems" :key="item.label" class="m-config-tile">
            <span class="m-config-label">{{ item.label }}</span>
            <span class="m-config-value" :class="item.valueClass">{{ item.value }}</span>
          </div>
        </div>
      </div>
    </template>

    <!-- Desktop -->
    <template v-else>
    <el-row :gutter="20" class="stats-row">
      <el-col :xs="12" :sm="12" :lg="6">
        <StatCard
          :label="$t('overview.clients')"
          :value="data.status.clientCounts"
          type="clients"
          :subtitle="$t('overview.clientsSubtitle')"
          to="/clients"
        />
      </el-col>
      <el-col :xs="12" :sm="12" :lg="6">
        <StatCard
          :label="$t('overview.proxies')"
          :value="proxyCounts"
          type="proxies"
          :subtitle="$t('overview.proxiesSubtitle')"
          to="/proxies/tcp"
        />
      </el-col>
      <el-col :xs="12" :sm="12" :lg="6">
        <StatCard
          :label="$t('overview.connections')"
          :value="data.status.curConns"
          type="connections"
          :subtitle="$t('overview.connectionsSubtitle')"
        />
      </el-col>
      <el-col :xs="12" :sm="12" :lg="6">
        <StatCard
          :label="$t('overview.traffic')"
          :value="formatTrafficTotal()"
          type="traffic"
          :subtitle="$t('overview.trafficSubtitle')"
        />
      </el-col>
    </el-row>

    <el-row :gutter="20" class="charts-row">
      <el-col :xs="24" :md="12">
        <el-card class="chart-card" shadow="hover">
          <template #header>
            <div class="card-header">
              <span class="card-title">{{ $t('overview.networkTraffic') }}</span>
              <el-tag size="small" type="info">{{ $t('overview.today') }}</el-tag>
            </div>
          </template>
          <div class="traffic-summary">
            <div class="traffic-item in">
              <div class="traffic-icon">
                <el-icon><Download /></el-icon>
              </div>
              <div class="traffic-info">
                <div class="label">{{ $t('overview.inbound') }}</div>
                <div class="value">
                  {{ formatFileSize(data.status.totalTrafficIn) }}
                </div>
              </div>
            </div>
            <div class="traffic-divider"></div>
            <div class="traffic-item out">
              <div class="traffic-icon">
                <el-icon><Upload /></el-icon>
              </div>
              <div class="traffic-info">
                <div class="label">{{ $t('overview.outbound') }}</div>
                <div class="value">
                  {{ formatFileSize(data.status.totalTrafficOut) }}
                </div>
              </div>
            </div>
          </div>
        </el-card>
      </el-col>
      <el-col :xs="24" :md="12">
        <el-card class="chart-card" shadow="hover">
          <template #header>
            <div class="card-header">
              <span class="card-title">{{ $t('overview.proxyTypes') }}</span>
              <el-tag size="small" type="info">{{ $t('overview.now') }}</el-tag>
            </div>
          </template>
          <div class="proxy-types-grid">
            <div
              v-for="(count, type) in data.status.proxyTypeCount"
              :key="type"
              class="proxy-type-item"
              v-show="count > 0"
            >
              <div class="proxy-type-name">{{ type.toUpperCase() }}</div>
              <div class="proxy-type-count">{{ count }}</div>
            </div>
            <div v-if="!hasActiveProxies" class="no-data">
              {{ $t('overview.noActiveProxies') }}
            </div>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <el-card class="config-card" shadow="hover">
      <template #header>
        <div class="card-header">
          <span class="card-title">{{ $t('overview.serverConfig') }}</span>
          <el-tag size="small" type="success">v{{ data.version }}</el-tag>
        </div>
      </template>
      <div class="config-grid">
        <div class="config-item">
          <span class="config-label">{{ $t('overview.bindPort') }}</span>
          <span class="config-value">{{ data.config.bindPort }}</span>
        </div>
        <div class="config-item" v-if="data.config.kcpBindPort != 0">
          <span class="config-label">{{ $t('overview.kcpPort') }}</span>
          <span class="config-value">{{ data.config.kcpBindPort }}</span>
        </div>
        <div class="config-item" v-if="data.config.quicBindPort != 0">
          <span class="config-label">{{ $t('overview.quicPort') }}</span>
          <span class="config-value">{{ data.config.quicBindPort }}</span>
        </div>
        <div class="config-item" v-if="data.config.vhostHTTPPort != 0">
          <span class="config-label">{{ $t('overview.httpPort') }}</span>
          <span class="config-value">{{ data.config.vhostHTTPPort }}</span>
        </div>
        <div class="config-item" v-if="data.config.vhostHTTPSPort != 0">
          <span class="config-label">{{ $t('overview.httpsPort') }}</span>
          <span class="config-value">{{ data.config.vhostHTTPSPort }}</span>
        </div>
        <div class="config-item" v-if="data.config.tcpmuxHTTPConnectPort != 0">
          <span class="config-label">{{ $t('overview.tcpmuxPort') }}</span>
          <span class="config-value">{{ data.config.tcpmuxHTTPConnectPort }}</span>
        </div>
        <div class="config-item" v-if="data.config.subdomainHost != ''">
          <span class="config-label">{{ $t('overview.subdomainHost') }}</span>
          <span class="config-value">{{ data.config.subdomainHost }}</span>
        </div>
        <div class="config-item">
          <span class="config-label">{{ $t('overview.maxPoolCount') }}</span>
          <span class="config-value">{{ data.config.maxPoolCount }}</span>
        </div>
        <div class="config-item">
          <span class="config-label">{{ $t('overview.maxPortsPerClient') }}</span>
          <span class="config-value">{{ maxPortsPerClientLabel }}</span>
        </div>
        <div class="config-item" v-if="data.config.allowPortsStr != ''">
          <span class="config-label">{{ $t('overview.allowPorts') }}</span>
          <span class="config-value">{{ data.config.allowPortsStr }}</span>
        </div>
        <div class="config-item" v-if="data.config.tlsForce">
          <span class="config-label">{{ $t('overview.tlsForce') }}</span>
          <el-tag size="small" type="warning">{{ $t('common.enabled') }}</el-tag>
        </div>
        <div class="config-item">
          <span class="config-label">{{ $t('overview.heartbeatTimeout') }}</span>
          <span class="config-value">{{ data.config.heartbeatTimeout }}s</span>
        </div>
      </div>
    </el-card>
    </template>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { useI18n } from 'vue-i18n'
import { formatFileSize } from '../utils/format'
import { Download, Upload } from '@element-plus/icons-vue'
import StatCard from '../components/StatCard.vue'
import { getServerInfo } from '../api/server'
import type { ServerInfo } from '../types/server'
import { useResponsive } from '../composables/useResponsive'

const { t } = useI18n()
const router = useRouter()
const { isMobile } = useResponsive()

const data = ref<ServerInfo>({
  version: '',
  config: {
    bindPort: 0,
    kcpBindPort: 0,
    quicBindPort: 0,
    vhostHTTPPort: 0,
    vhostHTTPSPort: 0,
    tcpmuxHTTPConnectPort: 0,
    subdomainHost: '',
    maxPoolCount: 0,
    maxPortsPerClient: 0,
    allowPortsStr: '',
    tlsForce: false,
    heartbeatTimeout: 0,
  },
  status: {
    clientCounts: 0,
    curConns: 0,
    totalTrafficIn: 0,
    totalTrafficOut: 0,
    proxyTypeCount: {},
  },
})

const hasActiveProxies = computed(() => {
  return Object.values(data.value.status.proxyTypeCount).some((c) => c > 0)
})

const proxyCounts = computed(() => {
  return Object.values(data.value.status.proxyTypeCount).reduce(
    (sum, count) => sum + (count || 0),
    0,
  )
})

const maxPortsPerClientLabel = computed(() => {
  const value = data.value.config.maxPortsPerClient
  return value === 0 ? t('common.noLimit') : String(value)
})

const formatTrafficTotal = () => {
  const total =
    data.value.status.totalTrafficIn + data.value.status.totalTrafficOut
  return formatFileSize(total)
}

const trafficInPercent = computed(() => {
  const total = data.value.status.totalTrafficIn + data.value.status.totalTrafficOut
  if (total === 0) return 50
  return Math.round((data.value.status.totalTrafficIn / total) * 100)
})

const trafficOutPercent = computed(() => {
  const total = data.value.status.totalTrafficIn + data.value.status.totalTrafficOut
  if (total === 0) return 50
  return Math.round((data.value.status.totalTrafficOut / total) * 100)
})

const configItems = computed(() => {
  const cfg = data.value.config
  const items: { label: string; value: string; valueClass?: string }[] = [
    { label: t('overview.bindPort'), value: String(cfg.bindPort) },
  ]
  if (cfg.kcpBindPort !== 0) {
    items.push({ label: t('overview.kcpPort'), value: String(cfg.kcpBindPort) })
  }
  if (cfg.quicBindPort !== 0) {
    items.push({ label: t('overview.quicPort'), value: String(cfg.quicBindPort) })
  }
  if (cfg.vhostHTTPPort !== 0) {
    items.push({ label: t('overview.httpPort'), value: String(cfg.vhostHTTPPort) })
  }
  if (cfg.vhostHTTPSPort !== 0) {
    items.push({ label: t('overview.httpsPort'), value: String(cfg.vhostHTTPSPort) })
  }
  if (cfg.tcpmuxHTTPConnectPort !== 0) {
    items.push({
      label: t('overview.tcpmuxPort'),
      value: String(cfg.tcpmuxHTTPConnectPort),
    })
  }
  if (cfg.subdomainHost !== '') {
    items.push({ label: t('overview.subdomainHost'), value: cfg.subdomainHost })
  }
  items.push({
    label: t('overview.maxPoolCount'),
    value: String(cfg.maxPoolCount),
  })
  items.push({
    label: t('overview.maxPortsPerClient'),
    value: maxPortsPerClientLabel.value,
  })
  if (cfg.allowPortsStr !== '') {
    items.push({ label: t('overview.allowPorts'), value: cfg.allowPortsStr })
  }
  if (cfg.tlsForce) {
    items.push({
      label: t('overview.tlsForce'),
      value: t('common.enabled'),
      valueClass: 'warning',
    })
  }
  items.push({
    label: t('overview.heartbeatTimeout'),
    value: `${cfg.heartbeatTimeout}s`,
  })
  return items
})

const fetchData = async () => {
  try {
    const json = await getServerInfo()
    data.value = json
  } catch {
    ElMessage({
      showClose: true,
      message: t('overview.fetchFailed'),
      type: 'error',
    })
  }
}

onMounted(() => {
  fetchData()
})
</script>

<style scoped>
.server-overview {
  padding: 0;
}

.stats-row {
  margin-bottom: 20px;
}

.charts-row {
  margin-bottom: 20px;
}

.chart-card {
  border-radius: 12px;
  border: 1px solid #e4e7ed;
  height: 100%;
}

html.dark .chart-card {
  border-color: #3a3d5c;
  background: #27293d;
}

.config-card {
  border-radius: 12px;
  border: 1px solid #e4e7ed;
  margin-bottom: 20px;
}

html.dark .config-card {
  border-color: #3a3d5c;
  background: #27293d;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.card-title {
  font-size: 16px;
  font-weight: 500;
  color: #303133;
}

html.dark .card-title {
  color: #e5e7eb;
}

.traffic-summary {
  display: flex;
  align-items: center;
  justify-content: space-around;
  min-height: 120px;
  padding: 10px 0;
}

.traffic-item {
  display: flex;
  align-items: center;
  gap: 16px;
}

.traffic-icon {
  width: 48px;
  height: 48px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 24px;
}

.traffic-item.in .traffic-icon {
  background: rgba(84, 112, 198, 0.1);
  color: #5470c6;
}

.traffic-item.out .traffic-icon {
  background: rgba(145, 204, 117, 0.1);
  color: #91cc75;
}

.traffic-info {
  display: flex;
  flex-direction: column;
}

.traffic-info .label {
  font-size: 14px;
  color: #909399;
}

.traffic-info .value {
  font-size: 24px;
  font-weight: 500;
  color: #303133;
}

html.dark .traffic-info .value {
  color: #e5e7eb;
}

.traffic-divider {
  width: 1px;
  height: 60px;
  background: #e4e7ed;
}

html.dark .traffic-divider {
  background: #3a3d5c;
}

.proxy-types-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(100px, 1fr));
  gap: 16px;
  min-height: 120px;
  align-content: center;
  padding: 10px 0;
}

.proxy-type-item {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 12px;
  background: #f8f9fa;
  border-radius: 8px;
}

html.dark .proxy-type-item {
  background: #1e1e2d;
}

.proxy-type-name {
  font-size: 12px;
  color: #909399;
  font-weight: 500;
  margin-bottom: 4px;
}

.proxy-type-count {
  font-size: 20px;
  font-weight: 500;
  color: #303133;
}

html.dark .proxy-type-count {
  color: #e5e7eb;
}

.no-data {
  grid-column: 1 / -1;
  display: flex;
  align-items: center;
  justify-content: center;
  height: 100%;
  color: #909399;
  font-size: 14px;
}

.config-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(180px, 1fr));
  gap: 16px;
}

.config-item {
  display: flex;
  flex-direction: column;
  gap: 4px;
  padding: 12px;
  background: #f8f9fa;
  border-radius: 8px;
  transition: background 0.2s;
}

html.dark .config-item {
  background: #1e1e2d;
}

.config-label {
  font-size: 12px;
  color: #909399;
  font-weight: 500;
}

html.dark .config-label {
  color: #9ca3af;
}

.config-value {
  font-size: 14px;
  color: #303133;
  font-weight: 500;
  word-break: break-all;
}

html.dark .config-value {
  color: #e5e7eb;
}

/* ===== Mobile: gorgeous animated layout ===== */

/* Keyframe animations */
@keyframes m-float-orb {
  0%, 100% { transform: translate(0, 0) scale(1); }
  33% { transform: translate(15px, -20px) scale(1.05); }
  66% { transform: translate(-10px, 10px) scale(0.95); }
}

@keyframes m-shimmer {
  0% { transform: translateX(-100%); }
  100% { transform: translateX(200%); }
}

@keyframes m-pulse-ring {
  0% { transform: scale(0.8); opacity: 0.6; }
  70% { transform: scale(1.6); opacity: 0; }
  100% { transform: scale(1.6); opacity: 0; }
}

@keyframes m-breathe {
  0%, 100% { opacity: 0.6; transform: scale(1); }
  50% { opacity: 1; transform: scale(1.15); }
}

@keyframes m-fade-up {
  from { opacity: 0; transform: translateY(16px); }
  to { opacity: 1; transform: translateY(0); }
}

@keyframes m-gradient-shift {
  0% { background-position: 0% 50%; }
  50% { background-position: 100% 50%; }
  100% { background-position: 0% 50%; }
}

@keyframes m-glow-pulse {
  0%, 100% { opacity: 0.4; }
  50% { opacity: 0.8; }
}

/* Animated background orbs */
.m-bg-orbs {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  height: 320px;
  pointer-events: none;
  overflow: hidden;
  z-index: 0;
}

.m-orb {
  position: absolute;
  border-radius: 50%;
  filter: blur(60px);
  animation: m-float-orb 8s ease-in-out infinite;
}

.m-orb-1 {
  width: 180px;
  height: 180px;
  top: -40px;
  right: -30px;
  background: radial-gradient(circle, rgba(102, 126, 234, 0.25), transparent 70%);
  animation-delay: 0s;
}

.m-orb-2 {
  width: 140px;
  height: 140px;
  top: 60px;
  left: -40px;
  background: radial-gradient(circle, rgba(240, 147, 251, 0.2), transparent 70%);
  animation-delay: -3s;
}

.m-orb-3 {
  width: 120px;
  height: 120px;
  top: 140px;
  right: 30%;
  background: radial-gradient(circle, rgba(79, 172, 254, 0.18), transparent 70%);
  animation-delay: -5s;
}

html.dark .m-orb-1 {
  background: radial-gradient(circle, rgba(129, 140, 248, 0.2), transparent 70%);
}

html.dark .m-orb-2 {
  background: radial-gradient(circle, rgba(244, 63, 94, 0.15), transparent 70%);
}

html.dark .m-orb-3 {
  background: radial-gradient(circle, rgba(96, 165, 250, 0.12), transparent 70%);
}

/* Hero stat cards - 2x2 + 1 grid */
.m-hero-grid {
  position: relative;
  z-index: 1;
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 10px;
  margin-bottom: 12px;
  animation: m-fade-up 0.5s ease-out both;
}

.m-hero-card {
  position: relative;
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 14px 12px;
  border-radius: 16px;
  background: rgba(255, 255, 255, 0.72);
  backdrop-filter: blur(12px);
  -webkit-backdrop-filter: blur(12px);
  border: 1px solid rgba(255, 255, 255, 0.6);
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.04), 0 1px 3px rgba(0, 0, 0, 0.03);
  overflow: hidden;
  transition: transform 0.2s cubic-bezier(0.34, 1.56, 0.64, 1), box-shadow 0.2s ease;
}

.m-hero-card::before {
  content: '';
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  height: 3px;
  background-size: 200% 100%;
  animation: m-gradient-shift 3s ease infinite;
}

.m-hero-card.clients::before {
  background-image: linear-gradient(90deg, #667eea, #764ba2, #667eea);
}

.m-hero-card.proxies::before {
  background-image: linear-gradient(90deg, #f093fb, #f5576c, #f093fb);
}

.m-hero-card.connections::before {
  background-image: linear-gradient(90deg, #4facfe, #00f2fe, #4facfe);
}

.m-hero-card.connections {
  grid-column: 1 / -1;
}

.m-hero-card:active {
  transform: scale(0.96);
  box-shadow: 0 2px 12px rgba(0, 0, 0, 0.08);
}

html.dark .m-hero-card {
  background: rgba(39, 41, 61, 0.8);
  border-color: rgba(52, 54, 77, 0.6);
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.2);
}

/* Hero card glow effect */
.m-hero-glow {
  position: absolute;
  top: -20px;
  right: -20px;
  width: 70px;
  height: 70px;
  border-radius: 50%;
  filter: blur(20px);
  animation: m-glow-pulse 3s ease-in-out infinite;
  pointer-events: none;
}

.m-hero-card.clients .m-hero-glow {
  background: rgba(102, 126, 234, 0.3);
}

.m-hero-card.proxies .m-hero-glow {
  background: rgba(245, 87, 108, 0.3);
}

.m-hero-card.connections .m-hero-glow {
  background: rgba(79, 172, 254, 0.3);
}

/* Pulse ring on connections card */
.m-hero-pulse-ring {
  position: absolute;
  right: 14px;
  top: 50%;
  margin-top: -10px;
  width: 20px;
  height: 20px;
  border-radius: 50%;
  border: 2px solid rgba(79, 172, 254, 0.5);
  animation: m-pulse-ring 2s cubic-bezier(0.4, 0, 0.6, 1) infinite;
}

.m-hero-icon-wrap {
  flex-shrink: 0;
}

.m-hero-icon {
  width: 38px;
  height: 38px;
  border-radius: 12px;
  display: flex;
  align-items: center;
  justify-content: center;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
}

.m-hero-icon svg {
  width: 20px;
  height: 20px;
}

.m-hero-card.clients .m-hero-icon {
  background: linear-gradient(135deg, #667eea, #764ba2);
  color: #fff;
}

.m-hero-card.proxies .m-hero-icon {
  background: linear-gradient(135deg, #f093fb, #f5576c);
  color: #fff;
}

.m-hero-card.connections .m-hero-icon {
  background: linear-gradient(135deg, #4facfe, #00f2fe);
  color: #fff;
}

html.dark .m-hero-card.clients .m-hero-icon {
  background: linear-gradient(135deg, #818cf8, #a78bfa);
}

html.dark .m-hero-card.proxies .m-hero-icon {
  background: linear-gradient(135deg, #fb7185, #f43f5e);
}

html.dark .m-hero-card.connections .m-hero-icon {
  background: linear-gradient(135deg, #60a5fa, #3b82f6);
}

.m-hero-info {
  flex: 1;
  min-width: 0;
}

.m-hero-value {
  font-size: 24px;
  font-weight: 800;
  line-height: 1.2;
  color: #1a1a2e;
  font-variant-numeric: tabular-nums;
  letter-spacing: -0.5px;
  background: linear-gradient(135deg, #1a1a2e 60%, #667eea);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
}

.m-hero-card.proxies .m-hero-value {
  background: linear-gradient(135deg, #1a1a2e 60%, #f5576c);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
}

.m-hero-card.connections .m-hero-value {
  background: linear-gradient(135deg, #1a1a2e 60%, #4facfe);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
}

html.dark .m-hero-value {
  background: linear-gradient(135deg, #f0f0f5 60%, #818cf8);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
}

html.dark .m-hero-card.proxies .m-hero-value {
  background: linear-gradient(135deg, #f0f0f5 60%, #fb7185);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
}

html.dark .m-hero-card.connections .m-hero-value {
  background: linear-gradient(135deg, #f0f0f5 60%, #60a5fa);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
}

.m-hero-label {
  font-size: 12px;
  color: #8e8ea0;
  margin-top: 2px;
  font-weight: 500;
}

html.dark .m-hero-label {
  color: #7a7a8e;
}

.m-hero-arrow {
  font-size: 22px;
  color: #c0c4cc;
  font-weight: 300;
  line-height: 1;
  transition: transform 0.2s ease;
}

.m-hero-card:active .m-hero-arrow {
  transform: translateX(3px);
}

/* Shared card styles - glassmorphism */
.m-card {
  position: relative;
  z-index: 1;
  background: rgba(255, 255, 255, 0.72);
  backdrop-filter: blur(12px);
  -webkit-backdrop-filter: blur(12px);
  border: 1px solid rgba(255, 255, 255, 0.6);
  border-radius: 16px;
  padding: 14px 16px;
  margin-bottom: 12px;
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.04), 0 1px 3px rgba(0, 0, 0, 0.03);
  animation: m-fade-up 0.5s ease-out both;
}

.m-traffic-card { animation-delay: 0.1s; }
.m-proxy-card { animation-delay: 0.2s; }
.m-config-card { animation-delay: 0.3s; }

html.dark .m-card {
  background: rgba(39, 41, 61, 0.8);
  border-color: rgba(52, 54, 77, 0.6);
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.2);
}

.m-card-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 14px;
}

.m-card-header-left {
  display: flex;
  align-items: center;
  gap: 8px;
}

.m-card-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  flex-shrink: 0;
  animation: m-breathe 2.5s ease-in-out infinite;
}

.m-card-dot.traffic {
  background: linear-gradient(135deg, #5470c6, #91cc75);
  box-shadow: 0 0 6px rgba(84, 112, 198, 0.5);
}

.m-card-dot.proxy {
  background: linear-gradient(135deg, #f5576c, #f093fb);
  box-shadow: 0 0 6px rgba(245, 87, 108, 0.5);
}

.m-card-dot.config {
  background: linear-gradient(135deg, #4facfe, #00f2fe);
  box-shadow: 0 0 6px rgba(79, 172, 254, 0.5);
}

.m-card-title {
  font-size: 14px;
  font-weight: 600;
  color: #1a1a2e;
  letter-spacing: -0.2px;
}

html.dark .m-card-title {
  color: #e0e0ea;
}

.m-card-badge {
  font-size: 11px;
  font-weight: 500;
  color: #8e8ea0;
  background: rgba(244, 245, 247, 0.8);
  padding: 2px 8px;
  border-radius: 6px;
}

.m-card-badge.version {
  background: linear-gradient(135deg, rgba(224, 242, 254, 0.9), rgba(219, 234, 254, 0.9));
  color: #3b82f6;
  font-weight: 600;
}

html.dark .m-card-badge {
  background: rgba(52, 54, 77, 0.8);
  color: #7a7a8e;
}

html.dark .m-card-badge.version {
  background: rgba(59, 130, 246, 0.15);
  color: #60a5fa;
}

/* Traffic card */
.m-traffic-grid {
  display: flex;
  flex-direction: column;
  gap: 14px;
}

.m-traffic-block {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.m-traffic-bar-wrap {
  height: 7px;
  background: rgba(240, 241, 245, 0.8);
  border-radius: 4px;
  overflow: hidden;
}

html.dark .m-traffic-bar-wrap {
  background: rgba(30, 30, 45, 0.8);
}

.m-traffic-bar {
  position: relative;
  height: 100%;
  border-radius: 4px;
  transition: width 0.8s cubic-bezier(0.4, 0, 0.2, 1);
  min-width: 4px;
  overflow: hidden;
}

.m-traffic-block.inbound .m-traffic-bar {
  background: linear-gradient(90deg, #5470c6, #73b8ff);
  box-shadow: 0 0 8px rgba(84, 112, 198, 0.4);
}

.m-traffic-block.outbound .m-traffic-bar {
  background: linear-gradient(90deg, #91cc75, #50e3c2);
  box-shadow: 0 0 8px rgba(145, 204, 117, 0.4);
}

/* Shimmer effect on traffic bars */
.m-traffic-shimmer {
  position: absolute;
  top: 0;
  left: 0;
  width: 50%;
  height: 100%;
  background: linear-gradient(90deg, transparent, rgba(255, 255, 255, 0.4), transparent);
  animation: m-shimmer 2s ease-in-out infinite;
}

.m-traffic-meta {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.m-traffic-direction {
  display: flex;
  align-items: center;
  gap: 5px;
  font-size: 12px;
  font-weight: 500;
  color: #8e8ea0;
}

.m-traffic-direction svg {
  flex-shrink: 0;
}

.m-traffic-block.inbound .m-traffic-direction {
  color: #5470c6;
}

.m-traffic-block.outbound .m-traffic-direction {
  color: #91cc75;
}

.m-traffic-amount {
  font-size: 15px;
  font-weight: 700;
  color: #1a1a2e;
  font-variant-numeric: tabular-nums;
  letter-spacing: -0.3px;
}

html.dark .m-traffic-amount {
  color: #e0e0ea;
}

/* Proxy type chips */
.m-type-chips {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.m-type-chip {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 6px 12px;
  border-radius: 20px;
  font-size: 12px;
  font-weight: 500;
  background: rgba(247, 248, 250, 0.8);
  color: #606266;
  border: 1px solid rgba(0, 0, 0, 0.04);
  transition: all 0.2s cubic-bezier(0.34, 1.56, 0.64, 1);
}

.m-type-chip:active {
  transform: scale(0.93);
}

html.dark .m-type-chip {
  background: rgba(30, 30, 45, 0.8);
  color: #c8cad0;
  border-color: rgba(255, 255, 255, 0.05);
}

.m-chip-dot {
  width: 7px;
  height: 7px;
  border-radius: 50%;
  background: #909399;
  animation: m-breathe 2s ease-in-out infinite;
}

.m-type-chip.chip-tcp .m-chip-dot { background: #409eff; box-shadow: 0 0 5px rgba(64, 158, 255, 0.5); }
.m-type-chip.chip-udp .m-chip-dot { background: #67c23a; box-shadow: 0 0 5px rgba(103, 194, 58, 0.5); }
.m-type-chip.chip-http .m-chip-dot { background: #e6a23c; box-shadow: 0 0 5px rgba(230, 162, 60, 0.5); }
.m-type-chip.chip-https .m-chip-dot { background: #f56c6c; box-shadow: 0 0 5px rgba(245, 108, 108, 0.5); }
.m-type-chip.chip-stcp .m-chip-dot { background: #a78bfa; box-shadow: 0 0 5px rgba(167, 139, 250, 0.5); }
.m-type-chip.chip-sudp .m-chip-dot { background: #34d399; box-shadow: 0 0 5px rgba(52, 211, 153, 0.5); }
.m-type-chip.chip-xtcp .m-chip-dot { background: #f472b6; box-shadow: 0 0 5px rgba(244, 114, 182, 0.5); }

.m-chip-name {
  font-weight: 600;
  color: #4a4a5a;
}

html.dark .m-chip-name {
  color: #d0d0da;
}

.m-chip-count {
  font-size: 13px;
  font-weight: 700;
  color: #1a1a2e;
  min-width: 16px;
  text-align: center;
  background: rgba(0,0,0,0.05);
  border-radius: 8px;
  padding: 0 5px;
  line-height: 1.6;
}

html.dark .m-chip-count {
  color: #e0e0ea;
  background: rgba(255,255,255,0.08);
}

/* Empty state */
.m-empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 8px;
  padding: 20px 0 8px;
  color: #b0b0b0;
  font-size: 13px;
}

html.dark .m-empty-state {
  color: #666;
}

/* Config grid */
.m-config-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 8px;
}

.m-config-tile {
  display: flex;
  flex-direction: column;
  gap: 3px;
  padding: 10px 12px;
  background: rgba(247, 248, 250, 0.7);
  border-radius: 12px;
  border: 1px solid rgba(0, 0, 0, 0.03);
  transition: all 0.2s ease;
}

.m-config-tile:active {
  background: rgba(247, 248, 250, 1);
  transform: scale(0.97);
}

html.dark .m-config-tile {
  background: rgba(30, 30, 45, 0.7);
  border-color: rgba(255, 255, 255, 0.03);
}

.m-config-label {
  font-size: 11px;
  color: #8e8ea0;
  font-weight: 500;
  line-height: 1.3;
}

html.dark .m-config-label {
  color: #6a6a7e;
}

.m-config-value {
  font-size: 14px;
  font-weight: 600;
  color: #1a1a2e;
  font-variant-numeric: tabular-nums;
  word-break: break-all;
  line-height: 1.3;
}

html.dark .m-config-value {
  color: #e0e0ea;
}

.m-config-value.warning {
  color: #e6a23c;
  font-weight: 700;
  text-shadow: 0 0 8px rgba(230, 162, 60, 0.3);
}

@media (max-width: 380px) {
  .m-hero-value {
    font-size: 20px;
  }

  .m-config-grid {
    grid-template-columns: 1fr;
  }
}

@media (max-width: 1199px) {
  .stats-row {
    row-gap: 20px;
  }
}

@media (max-width: 991px) {
  .charts-row .el-col {
    margin-bottom: 20px;
  }
  .charts-row .el-col:last-child {
    margin-bottom: 0;
  }
}

@media (max-width: 768px) {
  .chart-container {
    height: 250px;
  }

  .config-grid {
    grid-template-columns: repeat(2, 1fr);
  }
}

@media (max-width: 480px) {
  .traffic-summary {
    flex-direction: column;
    align-items: stretch;
    gap: 12px;
    min-height: auto;
    padding: 4px 0;
  }

  .traffic-divider {
    display: none;
  }

  .traffic-icon {
    width: 40px;
    height: 40px;
    font-size: 20px;
  }

  .traffic-info .value {
    font-size: 20px;
  }
}
</style>
