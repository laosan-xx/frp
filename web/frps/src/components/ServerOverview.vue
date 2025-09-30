<template>
  <div>
    <el-row>
      <el-col :md="11">
        <!-- 大屏幕使用表单 -->
        <div v-if="!isSmOrBelow" class="source">
          <el-form
            label-position="left"
            label-width="220px"
            class="server_info"
          >
            <el-form-item label="Version">
              <span>{{ data.version }}</span>
            </el-form-item>
            <el-form-item label="BindPort">
              <span>{{ data.bindPort }}</span>
            </el-form-item>
            <el-form-item label="KCP Bind Port" v-if="data.kcpBindPort != 0">
              <span>{{ data.kcpBindPort }}</span>
            </el-form-item>
            <el-form-item label="QUIC Bind Port" v-if="data.quicBindPort != 0">
              <span>{{ data.quicBindPort }}</span>
            </el-form-item>
            <el-form-item label="HTTP Port" v-if="data.vhostHTTPPort != 0">
              <span>{{ data.vhostHTTPPort }}</span>
            </el-form-item>
            <el-form-item label="HTTPS Port" v-if="data.vhostHTTPSPort != 0">
              <span>{{ data.vhostHTTPSPort }}</span>
            </el-form-item>
            <el-form-item
              label="TCPMux HTTPConnect Port"
              v-if="data.tcpmuxHTTPConnectPort != 0"
            >
              <span>{{ data.tcpmuxHTTPConnectPort }}</span>
            </el-form-item>
            <el-form-item
              label="Subdomain Host"
              v-if="data.subdomainHost != ''"
            >
              <LongSpan :content="data.subdomainHost" :length="30"></LongSpan>
            </el-form-item>
            <el-form-item label="Max PoolCount">
              <span>{{ data.maxPoolCount }}</span>
            </el-form-item>
            <el-form-item label="Max Ports Per Client">
              <span>{{ data.maxPortsPerClient }}</span>
            </el-form-item>
            <el-form-item label="Allow Ports" v-if="data.allowPortsStr != ''">
              <LongSpan :content="data.allowPortsStr" :length="30"></LongSpan>
            </el-form-item>
            <el-form-item label="TLS Force" v-if="data.tlsForce === true">
              <span>{{ data.tlsForce }}</span>
            </el-form-item>
            <el-form-item label="HeartBeat Timeout">
              <span>{{ data.heartbeatTimeout }}</span>
            </el-form-item>
            <el-form-item label="Client Counts">
              <span>{{ data.clientCounts }}</span>
            </el-form-item>
            <el-form-item label="Current Connections">
              <span>{{ data.curConns }}</span>
            </el-form-item>
            <el-form-item label="Proxy Counts">
              <span>{{ data.proxyCounts }}</span>
            </el-form-item>
          </el-form>
        </div>
        <!-- 小屏幕使用描述列表 -->
        <el-descriptions v-if="isSmOrBelow" :column="1" border>
          <el-descriptions-item label="Version">
            <span class="desc-span">{{ data.version }}</span>
          </el-descriptions-item>
          <el-descriptions-item label="BindPort">
            <span class="desc-span">{{ data.bindPort }}</span>
          </el-descriptions-item>
          <el-descriptions-item
            label="KCP Bind Port"
            v-if="data.kcpBindPort != 0"
          >
            <span class="desc-span">{{ data.kcpBindPort }}</span>
          </el-descriptions-item>
          <el-descriptions-item
            label="QUIC Bind Port"
            v-if="data.quicBindPort != 0 && !isSmOrBelow"
          >
            <span class="desc-span">{{ data.quicBindPort }}</span>
          </el-descriptions-item>
          <el-descriptions-item
            label="HTTP Port"
            v-if="data.vhostHTTPPort != 0 && !isSmOrBelow"
          >
            <span class="desc-span">{{ data.vhostHTTPPort }}</span>
          </el-descriptions-item>
          <el-descriptions-item
            label="HTTPS Port"
            v-if="data.vhostHTTPSPort != 0 && !isSmOrBelow"
          >
            <span class="desc-span">{{ data.vhostHTTPSPort }}</span>
          </el-descriptions-item>
          <el-descriptions-item
            label="TCPMux HTTPConnect Port"
            v-if="data.tcpmuxHTTPConnectPort != 0 && !isSmOrBelow"
          >
            <span class="desc-span">{{ data.tcpmuxHTTPConnectPort }}</span>
          </el-descriptions-item>
          <el-descriptions-item
            label="Subdomain Host"
            v-if="data.subdomainHost != ''"
          >
            <LongSpan :content="data.subdomainHost" :length="30"></LongSpan>
          </el-descriptions-item>
          <el-descriptions-item v-if="!isSmOrBelow" label="Max PoolCount">
            <span class="desc-span">{{ data.maxPoolCount }}</span>
          </el-descriptions-item>
          <el-descriptions-item
            v-if="!isSmOrBelow"
            label="Max Ports Per Client"
          >
            <span class="desc-span">{{ data.maxPortsPerClient }}</span>
          </el-descriptions-item>
          <el-descriptions-item
            label="Allow Ports"
            v-if="data.allowPortsStr != ''"
          >
            <LongSpan :content="data.allowPortsStr" :length="30"></LongSpan>
          </el-descriptions-item>
          <el-descriptions-item label="TLS Force" v-if="data.tlsForce === true">
            <span class="desc-span">{{ data.tlsForce }}</span>
          </el-descriptions-item>
          <el-descriptions-item label="HeartBeat Timeout">
            <span class="desc-span">{{ data.heartbeatTimeout }}</span>
          </el-descriptions-item>
          <el-descriptions-item label="Client Counts">
            <span class="desc-span">{{ data.clientCounts }}</span>
          </el-descriptions-item>
          <el-descriptions-item label="Current Connections">
            <span class="desc-span">{{ data.curConns }}</span>
          </el-descriptions-item>
          <el-descriptions-item label="Proxy Counts">
            <span class="desc-span">{{ data.proxyCounts }}</span>
          </el-descriptions-item>
        </el-descriptions>
      </el-col>
      <el-col :md="13">
        <div class="network-traffic">
          <div
            id="traffic"
            style="width: 400px; height: 260px; margin-bottom: 16px"
          ></div>
          <div id="proxies" style="width: 400px; height: 260px"></div>
        </div>
      </el-col>
    </el-row>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { ElMessage } from 'element-plus'
import { DrawTrafficChart, DrawProxyChart } from '../utils/chart'
import LongSpan from './LongSpan.vue'
import { useBreakpoints } from '../utils/breakpoints'

let data = ref({
  version: '',
  bindPort: 0,
  kcpBindPort: 0,
  quicBindPort: 0,
  vhostHTTPPort: 0,
  vhostHTTPSPort: 0,
  tcpmuxHTTPConnectPort: 0,
  subdomainHost: '',
  maxPoolCount: 0,
  maxPortsPerClient: '',
  allowPortsStr: '',
  tlsForce: false,
  heartbeatTimeout: 0,
  clientCounts: 0,
  curConns: 0,
  proxyCounts: 0,
})

// 全局断点
const { isSmOrBelow } = useBreakpoints()

const fetchData = () => {
  fetch('../api/serverinfo', { credentials: 'include' })
    .then((res) => res.json())
    .then((json) => {
      data.value.version = json.version
      data.value.bindPort = json.bindPort
      data.value.kcpBindPort = json.kcpBindPort
      data.value.quicBindPort = json.quicBindPort
      data.value.vhostHTTPPort = json.vhostHTTPPort
      data.value.vhostHTTPSPort = json.vhostHTTPSPort
      data.value.tcpmuxHTTPConnectPort = json.tcpmuxHTTPConnectPort
      data.value.subdomainHost = json.subdomainHost
      data.value.maxPoolCount = json.maxPoolCount
      data.value.maxPortsPerClient = json.maxPortsPerClient
      if (data.value.maxPortsPerClient == '0') {
        data.value.maxPortsPerClient = 'no limit'
      }
      data.value.allowPortsStr = json.allowPortsStr
      data.value.tlsForce = json.tlsForce
      data.value.heartbeatTimeout = json.heartbeatTimeout
      data.value.clientCounts = json.clientCounts
      data.value.curConns = json.curConns
      data.value.proxyCounts = 0
      if (json.proxyTypeCount != null) {
        if (json.proxyTypeCount.tcp != null) {
          data.value.proxyCounts += json.proxyTypeCount.tcp
        }
        if (json.proxyTypeCount.udp != null) {
          data.value.proxyCounts += json.proxyTypeCount.udp
        }
        if (json.proxyTypeCount.http != null) {
          data.value.proxyCounts += json.proxyTypeCount.http
        }
        if (json.proxyTypeCount.https != null) {
          data.value.proxyCounts += json.proxyTypeCount.https
        }
        if (json.proxyTypeCount.stcp != null) {
          data.value.proxyCounts += json.proxyTypeCount.stcp
        }
        if (json.proxyTypeCount.sudp != null) {
          data.value.proxyCounts += json.proxyTypeCount.sudp
        }
        if (json.proxyTypeCount.xtcp != null) {
          data.value.proxyCounts += json.proxyTypeCount.xtcp
        }
      }

      // draw chart
      DrawTrafficChart('traffic', json.totalTrafficIn, json.totalTrafficOut)
      DrawProxyChart('proxies', json)
    })
    .catch((err) => {
      // 只有不是 401（未授权）时才提示
      if (JSON.stringify(err) !== '{}') {
        ElMessage({
          showClose: true,
          message: '从 frps 获取服务器信息失败！',
          type: 'warning',
        })
      }
    })
}
fetchData()
</script>

<style>
/* 服务器信息容器 */
.source {
  background: white;
  border-radius: 12px;
  padding: 24px;
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.08);
  transition: all 0.3s ease;
}

html.dark .source {
  background: #2d3748;
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.3);
}

.source:hover {
  transform: translateY(-2px);
  box-shadow: 0 8px 30px rgba(0, 0, 0, 0.12);
}

html.dark .source:hover {
  box-shadow: 0 8px 30px rgba(0, 0, 0, 0.4);
}

/* 服务器信息表单 */
.server_info {
  margin-left: 0;
}

.server_info .el-form-item {
  margin-right: 0;
  margin-bottom: 16px;
  width: 100%;
  /* padding: 8px 0; */
  border-bottom: 1px solid #f1f5f9;
  transition: border-color 0.3s ease;
}

html.dark .server_info .el-form-item {
  border-bottom-color: #4a5568;
}

.server_info .el-form-item:last-child {
  border-bottom: none;
  margin-bottom: 0;
}

.server_info .el-form-item:hover {
  border-bottom-color: #e2e8f0;
}

html.dark .server_info .el-form-item:hover {
  border-bottom-color: #6b7280;
}

.server_info .el-form-item__label {
  color: #64748b;
  font-weight: 600;
  font-size: 14px;
  height: auto;
  line-height: 1.5;
  padding-bottom: 4px;
}

html.dark .server_info .el-form-item__label {
  color: #94a3b8;
}

.server_info .el-form-item__content {
  color: #334155;
  font-size: 14px;
  font-weight: 500;
  height: auto;
  line-height: 1.5;
}

html.dark .server_info .el-form-item__content {
  color: #e2e8f0;
}

.server_info .el-form-item__content span {
  margin-left: 8px;
}

.desc-span {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
  font-weight: 600;
}

/* 图表容器样式 */
.chart-container {
  background: white;
  border-radius: 12px;
  padding: 24px;
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.08);
  margin: 16px;
}

html.dark .chart-container {
  background: #2d3748;
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.3);
}

#traffic,
#proxies {
  width: 100% !important;
  margin-bottom: 0;
}

@media (min-width: 992px) {
  .network-traffic {
    padding-left: 16px;
  }
}

@media (max-width: 992px) {
  .network-traffic {
    margin-top: 16px;
  }
}

/* 响应式设计 */
@media (max-width: 768px) {
  .source,
  .chart-container {
    margin: 8px;
    padding: 16px;
    border-radius: 8px;
  }

  .server_info .el-form-item {
    margin-bottom: 12px;
    padding: 6px 0;
  }

  .server_info .el-form-item__label {
    font-size: 13px;
  }

  .server_info .el-form-item__content {
    font-size: 13px;
  }

  #traffic,
  #proxies {
    height: 200px !important;
  }
}

@media (max-width: 480px) {
  .source,
  .chart-container {
    margin: 4px;
    padding: 12px;
  }

  .server_info .el-form-item {
    flex-direction: column;
    align-items: flex-start;
    gap: 4px;
  }

  .server_info .el-form-item__content {
    margin-left: 0 !important;
    width: 100%;
  }

  .server_info .el-form-item__content span {
    margin-left: 0;
    display: block;
    width: 100%;
  }

  #traffic,
  #proxies {
    height: 180px !important;
  }
}

/* 加载状态 */
.server_info .el-form-item__content:empty::before {
  content: 'Loading...';
  color: #a0aec0;
  font-style: italic;
}

html.dark .server_info .el-form-item__content:empty::before {
  color: #6b7280;
}

/* 数值高亮 */
.server_info .el-form-item__content span {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
  font-weight: 600;
}

html.dark .server_info .el-form-item__content span {
  background: linear-gradient(135deg, #81e6d9 0%, #4fd1c5 100%);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
}
</style>
