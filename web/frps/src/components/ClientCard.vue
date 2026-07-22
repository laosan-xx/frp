<template>
  <div class="client-card" :class="{ 'is-online': client.online }" :style="{ animationDelay: `${(index ?? 0) * 0.07}s` }" @click="viewDetail">
    <div class="card-accent-line"></div>
    <div class="card-icon-wrapper">
      <div
        class="status-dot-large"
        :class="client.online ? 'online' : 'offline'"
      ></div>
      <div v-if="client.online" class="status-pulse-ring"></div>
    </div>

    <div class="card-content">
      <div class="card-header">
        <span class="client-main-id">{{ client.displayName }}</span>
        <span v-if="client.hostname" class="hostname-badge">{{
          client.hostname
        }}</span>
        <el-tag v-if="client.version" size="small" type="success"
          >v{{ client.version }}</el-tag
        >
        <el-tag v-if="client.wireProtocolLabel" size="small" type="info">
          {{ client.wireProtocolLabel }}
        </el-tag>
      </div>

      <div class="card-meta">
        <div class="meta-group">
          <span v-if="client.ip" class="meta-item meta-ip">
            <span class="meta-label">IP：</span>
            <span class="meta-value">{{ client.ip }}</span>
          </span>
          <span v-if="client.ipRegion" class="meta-item meta-region">
            <span class="meta-label">{{ $t('client.region') }}</span>
            <span class="meta-value">{{ client.ipRegion }}</span>
          </span>
        </div>
        <span class="meta-item activity">
          <el-icon class="activity-icon"><DataLine /></el-icon>
          <span class="meta-value">{{
            client.online ? client.lastConnectedAgo : client.disconnectedAgo
          }}</span>
        </span>
      </div>
    </div>

    <div class="card-action">
      <div class="status-badge" :class="client.online ? 'online' : 'offline'">
        {{ client.online ? $t('common.online') : $t('common.offline') }}
      </div>
      <el-icon class="arrow-icon"><ArrowRight /></el-icon>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useRouter } from 'vue-router'
import { DataLine, ArrowRight } from '@element-plus/icons-vue'
import type { Client } from '../utils/client'

interface Props {
  client: Client
  index?: number
}

const props = defineProps<Props>()
const router = useRouter()

const viewDetail = () => {
  router.push({
    name: 'ClientDetail',
    params: { key: props.client.key },
  })
}
</script>

<style scoped>
/* ===== Keyframes ===== */
@keyframes pulse-ring {
  0% { transform: scale(0.8); opacity: 0.6; }
  70% { transform: scale(2); opacity: 0; }
  100% { transform: scale(2); opacity: 0; }
}

@keyframes card-fade-in {
  0% { opacity: 0; transform: translateX(-36px); filter: blur(4px); }
  100% { opacity: 1; transform: translateX(0); filter: blur(0); }
}

@keyframes accent-flow {
  0% { background-position: 0% 50%; }
  50% { background-position: 100% 50%; }
  100% { background-position: 0% 50%; }
}

.client-card {
  display: flex;
  align-items: center;
  gap: 20px;
  padding: 24px;
  background: var(--el-bg-color);
  border: 1px solid var(--el-border-color-lighter);
  border-radius: 16px;
  cursor: pointer;
  transition: all 0.2s cubic-bezier(0.4, 0, 0.2, 1);
  position: relative;
  overflow: hidden;
  animation: card-fade-in 0.5s ease-out both;
}

.client-card:hover {
  transform: translateY(-2px);
  box-shadow: 0 8px 24px rgba(0, 0, 0, 0.06);
  border-color: var(--el-border-color-light);
}

/* Accent line at top */
.card-accent-line {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  height: 2.5px;
  background: linear-gradient(90deg, #909399, #c0c4cc, #909399);
  background-size: 200% 100%;
  opacity: 0.5;
  transition: opacity 0.3s;
}

.client-card.is-online .card-accent-line {
  background: linear-gradient(90deg, #67c23a, #95d475, #b3e19d, #67c23a);
  background-size: 200% 100%;
  animation: accent-flow 3s ease infinite;
  opacity: 0.8;
}

.card-icon-wrapper {
  width: 32px;
  height: 32px;
  border-radius: 8px;
  background: var(--el-fill-color);
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  transition: all 0.2s;
  position: relative;
}

.client-card:hover .card-icon-wrapper {
  background: var(--el-color-success-light-9);
}

.status-dot-large {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  transition: all 0.3s;
}

.status-dot-large.online {
  background-color: var(--el-color-success);
  box-shadow: 0 0 0 2px var(--el-color-success-light-8);
}

.status-dot-large.offline {
  background-color: var(--el-text-color-placeholder);
}

/* Pulse ring for online status */
.status-pulse-ring {
  position: absolute;
  width: 6px;
  height: 6px;
  border-radius: 50%;
  border: 1.5px solid var(--el-color-success);
  animation: pulse-ring 2.5s cubic-bezier(0.4, 0, 0.6, 1) infinite;
}

.card-content {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 8px;
  min-width: 0;
}

.card-header {
  display: flex;
  align-items: center;
  gap: 12px;
  flex-wrap: wrap;
}

.client-main-id {
  font-size: 15px;
  font-weight: 600;
  color: var(--el-text-color-primary);
  line-height: 1.2;
}

.hostname-badge {
  font-size: 12px;
  font-weight: 500;
  padding: 2px 8px;
  border-radius: 6px;
  background: var(--el-fill-color-dark);
  color: var(--el-text-color-regular);
}

.card-meta {
  display: flex;
  align-items: center;
  gap: 24px;
  font-size: 13px;
  color: var(--el-text-color-regular);
  flex-wrap: wrap;
}

.meta-group {
  display: flex;
  align-items: center;
  gap: 16px;
}

.meta-item {
  display: flex;
  align-items: center;
  gap: 0px;
}

.meta-label {
  color: var(--el-text-color-placeholder);
  font-weight: 500;
  font-size: 13px;
}

.meta-value {
  font-size: 13px;
  font-weight: 500;
  color: var(--el-text-color-primary);
}

.activity .meta-value {
  font-weight: 400;
  color: var(--el-text-color-secondary);
}

.activity-icon {
  margin-right: 4px;
}

.card-action {
  display: flex;
  align-items: center;
  gap: 20px;
  flex-shrink: 0;
}

.status-badge {
  padding: 4px 12px;
  border-radius: 20px;
  font-size: 13px;
  font-weight: 500;
  transition: all 0.2s;
}

.status-badge.online {
  background: var(--el-color-success-light-9);
  color: var(--el-color-success);
}

.status-badge.offline {
  background: var(--el-fill-color);
  color: var(--el-text-color-secondary);
}

.arrow-icon {
  font-size: 18px;
  color: var(--el-text-color-placeholder);
  transition: all 0.2s;
}

.client-card:hover .arrow-icon {
  color: var(--el-text-color-primary);
  transform: translateX(4px);
}

/* Dark mode adjustments */
html.dark .card-icon-wrapper {
  background: var(--el-fill-color-light);
}

html.dark .client-card:hover .card-icon-wrapper {
  background: var(--el-color-success-light-9);
}

html.dark .status-dot-large.online {
  box-shadow: 0 0 0 2px rgba(var(--el-color-success-rgb), 0.2);
}

html.dark .client-card {
  background: rgba(39, 41, 61, 0.6);
  border-color: rgba(52, 54, 77, 0.5);
}

html.dark .client-card:hover {
  box-shadow: 0 8px 24px rgba(0, 0, 0, 0.2);
}

@media (max-width: 767px) {
  .client-card {
    padding: 12px 14px;
    gap: 10px;
    border-radius: 14px;
    background: rgba(255, 255, 255, 0.68);
    backdrop-filter: blur(12px);
    -webkit-backdrop-filter: blur(12px);
    border-color: rgba(255, 255, 255, 0.5);
    box-shadow: 0 2px 12px rgba(0, 0, 0, 0.03);
    transition: transform 0.2s cubic-bezier(0.34, 1.56, 0.64, 1), box-shadow 0.2s ease;
  }

  html.dark .client-card {
    background: rgba(39, 41, 61, 0.75);
    border-color: rgba(52, 54, 77, 0.5);
    box-shadow: 0 2px 12px rgba(0, 0, 0, 0.15);
  }

  .client-card:hover {
    transform: none;
    box-shadow: 0 2px 12px rgba(0, 0, 0, 0.03);
  }

  .client-card:active {
    transform: scale(0.97);
    box-shadow: 0 1px 6px rgba(0, 0, 0, 0.06);
  }

  .card-accent-line {
    height: 2px;
    border-radius: 14px 14px 0 0;
  }

  .card-icon-wrapper {
    width: auto;
    height: auto;
    background: none;
    border-radius: 0;
    align-self: center;
  }

  .status-dot-large {
    width: 8px;
    height: 8px;
  }

  .status-pulse-ring {
    width: 8px;
    height: 8px;
  }

  .card-content {
    gap: 4px;
  }

  .card-header {
    gap: 6px;
    flex-wrap: nowrap;
    overflow: hidden;
  }

  .client-main-id {
    font-size: 14px;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
    flex-shrink: 1;
    min-width: 0;
  }

  .hostname-badge {
    flex-shrink: 0;
  }

  .card-header .el-tag {
    display: inline-flex;
  }

  .meta-item.meta-ip {
    display: none;
  }

  .meta-region .meta-label {
    color: #a3a3a3;
    font-size: 11px;
  }

  .meta-region .meta-value {
    color: #8c8c8c;
    font-weight: 500;
    font-size: 11px;
  }

  html.dark .meta-region .meta-label {
    color: #8a8a8a;
  }

  html.dark .meta-region .meta-value {
    color: #9e9e9e;
  }

  .card-meta {
    gap: 14px;
    font-size: 12px;
    flex-wrap: nowrap;
    overflow: hidden;
  }

  .meta-group {
    gap: 10px;
    min-width: 0;
    overflow: hidden;
  }

  .meta-item {
    gap: 0px;
    min-width: 0;
  }

  .meta-label {
    font-size: 11px;
  }

  .meta-value {
    font-size: 12px;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  .meta-item.activity {
    display: none;
  }

  .card-action {
    gap: 0;
  }

  .status-badge {
    display: none;
  }

  .arrow-icon {
    font-size: 14px;
    transition: transform 0.2s ease;
  }

  .client-card:active .arrow-icon {
    transform: translateX(3px);
  }
}
</style>
