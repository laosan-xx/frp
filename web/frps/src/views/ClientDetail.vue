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
        client?.displayName || (loading ? $t('common.loading') : routeId)
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

        <!-- Remote Command Card (only shown for online clients) -->
        <div v-if="client?.online" class="command-card">
          <div class="command-header">
            <h2>{{ $t('clientDetail.remoteCommand') }}</h2>
          </div>
          <div class="command-body">
            <div class="command-inputs">
              <el-radio-group
                v-if="predefinedCommands.length > 0"
                v-model="selectedPreset"
                class="command-radio-group"
                size="default"
                @change="onPresetChange"
              >
                <el-radio-button
                  v-for="cmd in predefinedCommands"
                  :key="cmd.value"
                  :value="cmd.value"
                >
                  {{ $t('clientDetail.' + cmd.label) }}
                </el-radio-button>
                <el-radio-button v-if="false" :label="CUSTOM_PRESET">
                  {{ $t('clientDetail.customCommand') }}
                </el-radio-button>
              </el-radio-group>
              <el-input
                v-if="selectedPreset === CUSTOM_PRESET"
                v-model="commandInput"
                :placeholder="$t('clientDetail.commandPlaceholder')"
                class="command-input"
                clearable
              />

              <!-- Passwall node management -->
              <div v-if="isPasswall" class="passwall-section">
                <!-- 当前节点信息（常驻面板，始终显示） -->
                <div class="command-result passwall-current-panel">
                  <div class="result-row">
                    <span class="result-label">{{ $t('clientDetail.passwallCurrentStatus') }}:</span>
                    <el-tag size="small" effect="plain" :style="passwallRunning ? ipGreenTagStyle : ipNeutralTagStyle">
                      {{ passwallRunning ? $t('clientDetail.passwallProxying') : $t('clientDetail.passwallNotRunning') }}
                    </el-tag>
                  </div>
                  <div class="result-row">
                    <span class="result-label">{{ $t('clientDetail.passwallCurrentNode') }}:</span>
                    <el-tag v-if="passwallRunning && passwallActiveNode" size="small" effect="plain" :style="ipGreenTagStyle">{{ passwallActiveNode }}</el-tag>
                    <el-tag v-else size="small" effect="plain" :style="ipNeutralTagStyle">{{ $t('clientDetail.passwallNotRunning') }}</el-tag>
                  </div>
                  <div class="result-row url-test-tags">
                    <span class="result-label">IP:</span>
                    <template v-if="ipPanel.state === 'egress'">
                      <span class="result-value-ip">{{ ipPanel.ip }}</span>
                      <el-tag v-if="passwallRunning && currentNodeTest?.latency" size="small" effect="plain" :style="ipLatencyTagStyle" class="ip-latency-tag">{{ (parseFloat(currentNodeTest.latency) * 1000).toFixed(0) }} ms</el-tag>
                      <span class="url-test-tags-group">
                        <el-tag v-if="currentNodeTest?.ip_country" size="small" effect="plain" :style="ipCountryTagStyle">{{ currentNodeTest.ip_country }}</el-tag>
                        <el-tag v-if="currentNodeTest?.ip_type" size="small" effect="plain" :style="ipTypeTagStyle(currentNodeTest.ip_type)">{{ currentNodeTest.ip_type }}</el-tag>
                        <el-tag v-if="currentNodeTest?.is_isp === 'true'" size="small" effect="plain" :style="ipGreenTagStyle">住宅IP</el-tag>
                      </span>
                    </template>
                    <span v-else-if="ipPanel.state === 'device'" class="result-value-ip">{{ ipPanel.ip }}</span>
                    <el-button v-else text :loading="true" size="small" />
                  </div>
                  <div v-if="currentNodeTest && currentNodeTest.location" class="result-row geo-isp-info">
                    <span class="result-label">{{ $t('clientDetail.passwallURLTestGeo') }}:</span>
                    <span class="result-value-geo">{{ currentNodeTest.location }}</span>
                  </div>
                  <div v-if="currentNodeTest && currentNodeTest.isp" class="result-row geo-isp-info">
                    <span class="result-label">{{ $t('clientDetail.passwallURLTestISP') }}:</span>
                    <span class="result-value-isp">{{ currentNodeTest.isp }}</span>
                  </div>
                </div>

                <!-- 操作结果：启用 / 切换 / 停用 / 添加节点等均在此显示 -->
                <div v-if="commandResp" class="command-result" style="margin: 0 16px 12px">
                  <div class="result-row">
                    <span class="result-label">{{ $t('clientDetail.commandResult') }}:</span>
                    <el-tag :type="commandResp.result === 'ok' ? 'success' : 'danger'" size="small">
                      {{ commandResp.result }}
                    </el-tag>
                  </div>
                  <div v-if="commandResp.output" class="output-row">
                    <span class="result-label">{{ $t('clientDetail.commandOutput') }}:</span>
                    <pre class="output-text">{{ commandResp.output }}</pre>
                  </div>
                </div>

                <!-- Loading -->
                <div v-if="nodeListLoading" class="node-list-loading">
                  <el-icon class="is-loading"><Loading /></el-icon>
                  <span>{{ $t('clientDetail.loadingNodes') }}</span>
                </div>
                <div v-else-if="nodeListError" class="node-list-error">{{ nodeListError }}</div>
                <div v-else-if="filteredNodeList.length === 0" class="node-list-empty">
                  {{ $t('clientDetail.noNodes') }}
                  <el-button
                    v-if="passwallEnabled"
                    size="small"
                    type="warning"
                    :loading="passwallNodeLoading['__disable__'] === 'disable'"
                    style="margin-left: 12px"
                    @click="passwallDisableNode"
                  >
                    <span v-if="passwallNodeLoading['__disable__'] !== 'disable'">{{ $t('clientDetail.passwallShutdown') }}</span>
                  </el-button>
                </div>

                <!-- Node list -->
                <div v-else class="passwall-node-list">
                  <el-divider content-position="left">
                    <span class="passwall-extra-title">{{ $t('clientDetail.passwallNodeList') }}</span>
                  </el-divider>
                  <div v-for="node in filteredNodeList" :key="node.id" class="passwall-node-item">
                    <div class="passwall-node-info">
                      <span class="node-name">{{ node.remarks }}</span>
                      <span class="node-meta">{{ node.type }} | {{ node.address }}:{{ node.port }}</span>
                    </div>
                    <div class="passwall-node-actions">
                      <el-button
                        size="small"
                        class="url-test-btn site-tag"
                        :class="{
                          'url-test-loading': passwallNodeTestState[node.id]?.loading,
                          'url-test-ok': !passwallNodeTestState[node.id]?.loading && passwallNodeTestState[node.id]?.code && passwallNodeTestState[node.id]?.code !== '0' && passwallNodeTestState[node.id]?.code !== '000',
                          'url-test-fail': !passwallNodeTestState[node.id]?.loading && (passwallNodeTestState[node.id]?.error || passwallNodeTestState[node.id]?.code === '0' || passwallNodeTestState[node.id]?.code === '000')
                        }"
                        :disabled="passwallNodeTestState[node.id]?.loading || !!Object.keys(passwallNodeLoading).length"
                        @click="passwallURLTest(node.id, true)"
                      >
                        <template v-if="passwallNodeTestState[node.id]?.loading">
                          <el-icon class="is-loading"><Loading /></el-icon>
                        </template>
                        <template v-else-if="passwallNodeTestState[node.id]?.code && passwallNodeTestState[node.id]?.code !== '0' && passwallNodeTestState[node.id]?.code !== '000'">
                          {{ passwallNodeTestState[node.id]?.latency ? (parseFloat(passwallNodeTestState[node.id].latency) * 1000).toFixed(0) + ' ms' : passwallNodeTestState[node.id]?.code }}
                        </template>
                        <template v-else-if="passwallNodeTestState[node.id]?.error">
                          {{ $t('clientDetail.passwallURLTestFail') }}
                        </template>
                        <template v-else-if="passwallNodeTestState[node.id]?.code === '000'">
                          {{ $t('clientDetail.passwallURLTestFail') }}
                        </template>
                        <template v-else>
                          {{ $t('clientDetail.passwallURLTest') }}
                        </template>
                      </el-button>
                      <el-button
                        size="small"
                        type="success"
                        class="site-tag"
                        :loading="sharingNode === node.id"
                        @click="shareNode(node.id)"
                      >
                        <span v-if="sharingNode !== node.id">{{ $t('clientDetail.shareNode') }}</span>
                      </el-button>
                      <!-- Passwall running: current node shows 停用, others show 切换 -->
                      <template v-if="passwallEnabled">
                        <el-button
                          v-if="node.active"
                          size="small"
                          type="warning"
                          class="site-tag"
                          :loading="passwallNodeLoading['__disable__'] === 'disable'"
                          :disabled="!!Object.keys(passwallNodeLoading).length"
                          @click="passwallDisableNode"
                        >
                          <span v-if="passwallNodeLoading['__disable__'] !== 'disable'">{{ $t('clientDetail.passwallDisable') }}</span>
                        </el-button>
                        <el-button
                          v-else
                          size="small"
                          type="primary"
                          class="site-tag"
                          :loading="passwallNodeLoading[node.id] === 'enable'"
                          :disabled="!!Object.keys(passwallNodeLoading).length"
                          @click="passwallEnableNode(node.id)"
                        >
                          <span v-if="passwallNodeLoading[node.id] !== 'enable'">{{ $t('clientDetail.passwallSwitch') }}</span>
                        </el-button>
                      </template>
                      <!-- Passwall stopped: all nodes show 启用 -->
                      <el-button
                        v-else
                        size="small"
                        type="primary"
                        class="site-tag"
                        :loading="passwallNodeLoading[node.id] === 'enable'"
                        :disabled="!!Object.keys(passwallNodeLoading).length"
                        @click="passwallEnableNode(node.id)"
                      >
                        <span v-if="passwallNodeLoading[node.id] !== 'enable'">{{ $t('clientDetail.passwallEnable') }}</span>
                      </el-button>
                      <!-- Delete always shown -->
                      <el-button
                        size="small"
                        type="danger"
                        class="site-tag"
                        :loading="passwallNodeLoading[node.id] === 'delete'"
                        :disabled="!!Object.keys(passwallNodeLoading).length"
                        @click="passwallDeleteNode(node.id)"
                      >
                        <span v-if="passwallNodeLoading[node.id] !== 'delete'">{{ $t('clientDetail.passwallDelete') }}</span>
                      </el-button>
                    </div>

                  </div>
                </div>

                <!-- 分享二维码弹窗 -->
                <el-dialog
                  v-model="shareDialogVisible"
                  :title="$t('clientDetail.shareQrTitle')"
                  width="fit-content"
                  align-center
                  append-to-body
                  :close-on-click-modal="false"
                  class="share-qr-dialog"
                >
                  <div class="share-qr-wrap">
                    <img v-if="shareQrDataUrl" :src="shareQrDataUrl" alt="QR" class="share-qr-img" />
                    <el-button
                      v-if="shareLink"
                      size="default"
                      type="primary"
                      class="share-qr-copy"
                      @click="copyShareLink"
                    >
                      {{ $t('clientDetail.copyLink') }}
                    </el-button>
                  </div>
                </el-dialog>

                <!-- Add node section -->
                <div class="passwall-add-section">
                  <el-divider content-position="left">
                    <span class="passwall-extra-title">{{ $t('clientDetail.passwallAddNode') }}</span>
                  </el-divider>
                  <div class="passwall-add-form">
                    <el-input
                      v-model="passwallAddLink"
                      :placeholder="$t('clientDetail.passwallLinkPlaceholder')"
                      clearable
                    />
                    <el-button
                      type="primary"
                      :loading="passwallAddLoading"
                      :disabled="!passwallAddLink.trim()"
                      @click="passwallAddNode"
                    >
                      {{ $t('clientDetail.passwallAddNode') }}
                    </el-button>
                  </div>
                </div>

                <!-- 其他功能区：更新默认规则等 -->
                <div class="passwall-extra-section">
                  <el-divider content-position="left">
                    <span class="passwall-extra-title">{{ $t('clientDetail.passwallExtraTitle') }}</span>
                  </el-divider>
                  <div class="passwall-rule-bar">
                    <el-button
                      type="primary"
                      plain
                      :loading="ruleUpdateLoading"
                      @click="passwallUpdateRules"
                    >
                      {{ ruleUpdateLoading ? $t('clientDetail.passwallUpdateRulesRunning') : $t('clientDetail.passwallUpdateRules') }}
                    </el-button>
                    <span class="passwall-rule-hint">
                      {{ ruleUpdateLoading ? $t('clientDetail.passwallUpdateRulesHint') : $t('clientDetail.passwallUpdateRulesDesc') }}
                    </span>
                  </div>

                  <!-- 更新默认规则结果 -->
                  <div v-if="ruleUpdateResp" class="command-result passwall-rule-result">
                    <div class="result-row">
                      <span class="result-label">{{ $t('clientDetail.passwallUpdateRules') }}:</span>
                      <el-tag :type="ruleUpdateResp.ok ? 'success' : 'danger'" size="small">
                        {{ ruleUpdateResp.ok ? $t('clientDetail.passwallUpdateRulesOk') : $t('clientDetail.passwallUpdateRulesFail') }}
                      </el-tag>
                      <span v-if="ruleUpdateResp.duration" class="passwall-rule-hint">
                        {{ $t('clientDetail.passwallUpdateRulesDuration', { sec: ruleUpdateResp.duration }) }}
                      </span>
                    </div>
                    <div v-if="ruleUpdateResp.message" class="output-row">
                      <span class="result-label">{{ $t('clientDetail.commandOutput') }}:</span>
                      <pre class="output-text">{{ ruleUpdateResp.message }}</pre>
                    </div>
                  </div>
                </div>

              </div>

              <!-- Frp config fields (modify_frp) -->
              <div v-if="isFrpConfig" v-loading="frpLoading" class="frp-config-form">
                <div class="frp-config-header">
                  <el-icon class="frp-config-header-icon"><EditPen /></el-icon>
                  <div class="frp-config-header-text">
                    <div class="frp-config-title">{{ $t('clientDetail.cmdModifyFrp') }}</div>
                    <div class="frp-config-subtitle">{{ $t('clientDetail.frpConfigHint') }}</div>
                  </div>
                </div>
                <el-divider class="frp-config-divider" />
                <div class="frp-config-grid">
                  <div class="frp-config-field">
                    <span class="frp-config-label">
                      <el-icon><Connection /></el-icon>{{ $t('clientDetail.frpServerAddrLabel') }}
                    </span>
                    <el-input
                      v-model="frpServerAddrPort"
                      :placeholder="$t('clientDetail.frpServerAddrPort')"
                      clearable
                    />
                  </div>
                  <div class="frp-config-field">
                    <span class="frp-config-label">
                      <el-icon><User /></el-icon>{{ $t('clientDetail.frpUserLabel') }}
                    </span>
                    <el-input
                      v-model="frpUser"
                      :placeholder="$t('clientDetail.frpUser')"
                      clearable
                    />
                  </div>
                  <div class="frp-config-field frp-config-field--inline">
                    <div class="frp-config-inline-item">
                      <span class="frp-config-label">
                        <el-icon><Link /></el-icon>{{ $t('clientDetail.frpProtocolLabel') }}
                      </span>
                      <el-radio-group v-model="frpProtocol" size="default">
                        <el-radio-button value="websocket">{{ $t('clientDetail.frpProtocolWebsocket') }}</el-radio-button>
                        <el-radio-button value="wss">{{ $t('clientDetail.frpProtocolWss') }}</el-radio-button>
                      </el-radio-group>
                    </div>
                    <div v-show="frpProtocol !== 'wss'" class="frp-config-inline-item">
                      <span class="frp-config-label">
                        <el-icon><Lock /></el-icon>{{ $t('clientDetail.frpTlsLabel') }}
                      </span>
                      <el-radio-group v-model="frpTlsEnable" size="default">
                        <el-radio-button :value="true">{{ $t('clientDetail.frpTlsOn') }}</el-radio-button>
                        <el-radio-button :value="false">{{ $t('clientDetail.frpTlsOff') }}</el-radio-button>
                      </el-radio-group>
                    </div>
                  </div>
                </div>
              </div>

              <!-- Normal payload input for other commands -->
              <el-input
                v-if="selectedPreset !== '' && !isPasswall && !needFirmwareUpdate && !isFrpConfig && !isModifySystem"
                v-model="payloadInput"
                type="textarea"
                :rows="3"
                :placeholder="payloadPlaceholder"
                class="payload-input"
              />

              <!-- System settings fields (modify_system) -->
              <div v-if="isModifySystem" v-loading="systemLoading" class="frp-config-form">
                <div class="frp-config-header">
                  <el-icon class="frp-config-header-icon"><Setting /></el-icon>
                  <div class="frp-config-header-text">
                    <div class="frp-config-title">{{ $t('clientDetail.cmdModifySystem') }}</div>
                    <div class="frp-config-subtitle">{{ $t('clientDetail.systemSettingsHint') }}</div>
                  </div>
                </div>
                <el-divider class="frp-config-divider" />
                <div class="system-toggle-list">
                  <div class="system-toggle-item">
                    <span class="system-toggle-label">{{ $t('clientDetail.sysReboot') }}</span>
                    <el-button
                      class="system-reboot-btn"
                      type="danger"
                      circle
                      size="small"
                      :loading="rebooting"
                      :icon="SwitchButton"
                      @click="rebootSystem"
                    />
                  </div>
                  <div class="system-toggle-item" :class="{ 'is-disabled-row': defaultPasswordOn }">
                    <span class="system-toggle-label">
                      {{ $t('clientDetail.sysDefaultPassword') }}
                      <span v-if="defaultPasswordOn && defaultPasswordCountdown > 0" class="system-toggle-ssid">
                        · {{ $t('clientDetail.sysDefaultPasswordCountdown', { sec: defaultPasswordCountdown }) }}
                      </span>
                    </span>
                    <el-switch
                      class="system-switch"
                      :model-value="defaultPasswordOn"
                      :loading="defaultPasswordLoading"
                      :disabled="defaultPasswordOn"
                      :active-text="$t('clientDetail.sysYes')"
                      :inactive-text="$t('clientDetail.sysNo')"
                      inline-prompt
                      @change="(v: any) => onToggleDefaultPassword(v)"
                    />
                  </div>
                  <div class="system-toggle-item" :class="{ 'is-disabled-row': !defaultPasswordOn || defaultPasswordCountdown > 0 }">
                    <span class="system-toggle-label">
                      {{ $t('clientDetail.sysCommonPassword') }}
                    </span>
                    <el-switch
                      class="system-switch"
                      :model-value="commonPasswordOn"
                      :loading="commonPasswordLoading"
                      :disabled="!defaultPasswordOn || defaultPasswordCountdown > 0"
                      :active-text="$t('clientDetail.sysYes')"
                      :inactive-text="$t('clientDetail.sysNo')"
                      inline-prompt
                      @change="(v: any) => onToggleCommonPassword(v)"
                    />
                  </div>
                  <div class="system-toggle-item">
                    <span class="system-toggle-label">{{ $t('clientDetail.sysWan6') }}</span>
                    <el-switch
                      class="system-switch"
                      :model-value="systemWan6"
                      :loading="wan6Loading"
                      :active-text="$t('clientDetail.sysOn')"
                      :inactive-text="$t('clientDetail.sysOff')"
                      inline-prompt
                      @change="(v: any) => onToggleWan6(v)"
                    />
                  </div>
                  <div class="system-toggle-item" v-for="b in systemBands" :key="b.key">
                    <span class="system-toggle-label">
                      {{ b.label }} WiFi
                      <span v-if="b.ssid" class="system-toggle-ssid">· {{ b.ssid }}</span>
                    </span>
                    <el-switch
                      class="system-switch"
                      :model-value="b.enabled"
                      :loading="b.loading"
                      :active-text="$t('clientDetail.sysOn')"
                      :inactive-text="$t('clientDetail.sysOff')"
                      inline-prompt
                      @change="(v: any) => onToggleBand(b.key, v)"
                    />
                  </div>
                </div>
                <el-divider class="frp-config-divider" />
                <div class="system-ssid-field">
                  <span class="frp-config-label">
                    <el-icon><Iphone /></el-icon>{{ $t('clientDetail.sysSsidLabel') }}
                  </span>
                  <div class="ssid-input-row">
                    <el-select
                      v-model="target"
                      class="ssid-target-select"
                      :placeholder="$t('clientDetail.sysSsidTargetAll')"
                      @change="onTargetChange"
                    >
                      <el-option :label="$t('clientDetail.sysSsidTargetAll')" value="all" />
                      <el-option
                        v-for="b in systemBands"
                        :key="b.key"
                        :label="b.label + ' WiFi'"
                        :value="b.key"
                      />
                    </el-select>
                    <el-input
                      v-model="systemSsid"
                      class="ssid-input"
                      :placeholder="$t('clientDetail.sysSsidPlaceholder')"
                      clearable
                      maxlength="32"
                      @keyup.enter="sendWifiChange"
                    />
                    <el-input
                      v-model="systemPassword"
                      class="ssid-input"
                      type="password"
                      show-password
                      :placeholder="$t('clientDetail.sysPasswordPlaceholder')"
                      clearable
                      maxlength="63"
                      @keyup.enter="sendWifiChange"
                    />
                    <el-button
                      class="ssid-send-btn"
                      type="primary"
                      :loading="ssidSending"
                      :disabled="!systemSsid.trim() && !systemPassword.trim()"
                      @click="sendWifiChange"
                    >
                      {{ $t('clientDetail.send') }}
                    </el-button>
                  </div>
                </div>
              </div>

              <!-- Firmware Update Wizard -->
              <div v-if="needFirmwareUpdate" class="firmware-wizard frp-config-form">
                <div class="frp-config-header">
                  <el-icon class="frp-config-header-icon"><Download /></el-icon>
                  <div class="frp-config-header-text">
                    <div class="frp-config-title">{{ $t('clientDetail.cmdUpdateSystem') }}</div>
                    <div class="frp-config-subtitle">{{ $t('clientDetail.systemUpdateHint') }}</div>
                  </div>
                </div>
                <el-divider class="frp-config-divider" />
                <div class="fw-current-version">
                  <span class="fw-info-label">{{ $t('clientDetail.fwCurrentVersion') }}</span>
                  <el-tag v-if="fwCurrentVersion" type="info" size="small" effect="plain">
                    {{ fwCurrentVersion }}
                  </el-tag>
                  <span v-else>—</span>
                </div>
                <!-- Step 0: Check update button -->
                <div v-if="fwStep === 0" class="fw-check-update">
                  <el-button
                    type="primary"
                    :loading="commandSending"
                    @click="sendCommand"
                  >
                    {{ sendBtnText }}
                  </el-button>
                </div>
                <!-- Step 1: Detecting -->
                <div v-if="fwStep === 1" class="fw-step">
                  <div class="fw-step-loading">
                    <el-icon class="is-loading"><Loading /></el-icon>
                    <span>{{ $t('clientDetail.fwDetecting') }}</span>
                  </div>
                  <div v-if="fwError" class="fw-error">{{ fwError }}</div>
                  <div class="fw-platform-info">
                    <div v-if="fwPlatform" class="fw-info-row"><span class="fw-info-label">{{ $t('clientDetail.fwPlatform') }}</span><span>{{ fwPlatform.target }}</span></div>
                    <div v-if="fwPlatform" class="fw-info-row"><span class="fw-info-label">{{ $t('clientDetail.fwModel') }}</span><span>{{ fwPlatform.model }}</span></div>
                    <div v-if="fwPlatform" class="fw-info-row"><span class="fw-info-label">Board</span><span>{{ fwPlatform.boardName }}</span></div>
                  </div>
                </div>

                <!-- Step 2: Select Branch -->
                <div v-if="fwStep >= 2" class="fw-step">
                  <h4 class="fw-step-title">{{ $t('clientDetail.fwSelectBranch') }}</h4>
                  <div v-if="fwLoading" class="fw-step-loading">
                    <el-icon class="is-loading"><Loading /></el-icon>
                    <span>{{ $t('clientDetail.fwFetching') }}</span>
                  </div>
                  <div v-if="fwError" class="fw-error">{{ fwError }}</div>
                  <div>
                    <el-radio-group v-if="fwBranches.length > 0" v-model="fwSelectedBranch" :disabled="fwUpgrading || fwDownloadStatus.status === 'downloading'" class="node-radio-group" @change="onBranchSelect">
                      <el-radio v-for="(branch, idx) in fwBranches" :key="idx" :value="idx" class="node-radio-item">
                        <span class="node-name">{{ branch.config }} / {{ branch.branch }}</span>
                        <el-tag size="small" type="info">{{ branch.date }}</el-tag>
                        <el-tag size="small" type="warning">{{ branch.assets.length }} {{ $t('clientDetail.fwFiles') }}</el-tag>
                      </el-radio>
                    </el-radio-group>
                  </div>
                </div>

                <!-- Step 3: Select Firmware File -->
                <div v-if="fwStep >= 3 && selectedBranchFiles.length > 0" class="fw-step">
                  <h4 class="fw-step-title">{{ $t('clientDetail.fwSelectFile') }}</h4>
                  <div>
                    <el-radio-group v-model="fwSelectedFile" :disabled="fwUpgrading || fwDownloadStatus.status === 'downloading'" class="node-radio-group" @change="onFileSelect">
                      <el-radio v-for="(file, idx) in selectedBranchFiles" :key="idx" :value="idx" class="node-radio-item">
                        <span class="node-name">{{ file.name }}</span>
                        <el-tag size="small">{{ formatFileSize(file.size) }}</el-tag>
                      </el-radio>
                    </el-radio-group>
                  </div>
                  <!-- Download button: toggles between download / cancel / re-download -->
                  <el-button
                    v-if="fwSelectedFile !== null && !fwDownloadStarted && !fwUpgrading"
                    type="primary"
                    style="margin-top: 12px"
                    @click="startDownload"
                  >
                    {{ $t('clientDetail.fwDownload') }}
                  </el-button>
                  <el-button
                    v-if="fwSelectedFile !== null && fwDownloadStarted && !fwUpgrading"
                    :type="fwDownloadStatus.status === 'downloading' || fwDownloadStatus.status === 'complete' ? 'danger' : 'primary'"
                    style="margin-top: 12px"
                    @click="fwDownloadStatus.status === 'downloading' ? cancelDownload() : fwDownloadStatus.status === 'complete' ? runSysupgrade() : startDownload()"
                  >
                    {{ fwDownloadStatus.status === 'downloading' ? $t('clientDetail.fwCancel') : fwDownloadStatus.status === 'complete' ? $t('clientDetail.fwUpgradeNow') : $t('clientDetail.fwReDownload') }}
                  </el-button>
                </div>

                <!-- Step 4: Download Progress (clean, no buttons) -->
                <div v-if="fwDownloadStarted && !fwUpgrading" class="fw-step">
                  <h4 class="fw-step-title">
                    {{ fwDownloadStatus.status === 'cancelled' ? $t('clientDetail.fwCancelled') : fwDownloadStatus.status === 'complete' ? $t('clientDetail.fwDownloadComplete') : fwDownloadStatus.status === 'error' ? $t('clientDetail.fwDownload') : $t('clientDetail.fwDownloading') }}
                  </h4>
                  <el-progress
                    :show-text="false"
                    :percentage="Math.round(fwDownloadStatus.progress)"
                    :status="fwDownloadStatus.status === 'complete' ? 'success' : fwDownloadStatus.status === 'error' ? 'exception' : fwDownloadStatus.status === 'cancelled' ? 'warning' : undefined"
                  />
                  <div class="fw-download-info">
                    <span>{{ formatFileSize(fwDownloadStatus.downloadedBytes) }} / {{ formatFileSize(fwDownloadStatus.totalBytes) }}</span>
                  </div>
                  <div v-if="fwDownloadStatus.status === 'error'" class="fw-error">{{ fwDownloadStatus.error }}</div>
                </div>

                <!-- Upgrading State -->
                <div v-if="fwUpgrading" class="fw-step">
                  <div class="fw-step-loading">
                    <el-icon class="is-loading"><Loading /></el-icon>
                    <span>{{ $t('clientDetail.fwUpgrading') }}</span>
                  </div>
                </div>
              </div>

              <div v-if="selectedPreset !== '' && selectedPreset !== 'update_system' && !isPasswall && (!needFirmwareUpdate || fwStep === 0) && !fwUpgrading && !isModifySystem" class="command-actions">
                <el-button
                  v-if="reconnecting"
                  type="primary"
                  :loading="true"
                  disabled
                >
                  {{ $t('clientDetail.reconnecting') }}
                </el-button>
                <el-button
                  v-else-if="returnCountdown > 0"
                  type="info"
                  :loading="true"
                  disabled
                >
                  {{ returnCountdown }}{{ $t('clientDetail.returnCountdownSuffix') }}
                </el-button>
                <el-button
                  v-else
                  type="primary"
                  :loading="commandSending"
                  :disabled="!hasCommandInput"
                  @click="sendCommand"
                >
                  {{ sendBtnText }}
                </el-button>
              </div>
              <div v-if="!isPasswall && commandResp" class="command-result">
                <div class="result-row">
                  <span class="result-label">{{ $t('clientDetail.commandResult') }}:</span>
                  <el-tag :type="commandResp.result === 'ok' ? 'success' : 'danger'" size="small">
                    {{ commandResp.result }}
                  </el-tag>
                </div>
                <div v-if="commandResp.output" class="output-row">
                  <span class="result-label">{{ $t('clientDetail.commandOutput') }}:</span>
                  <pre class="output-text">{{ commandResp.output }}</pre>
                </div>
              </div>
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
import { ref, onMounted, onUnmounted, watch, computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage, ElMessageBox, ElPagination, ElTag, ElButton, ElProgress } from 'element-plus'
import { ArrowLeft, Loading, Search, EditPen, Connection, User, Setting, Iphone, Link, Lock, Download, SwitchButton } from '@element-plus/icons-vue'
import { useI18n } from 'vue-i18n'
import { useResponsive } from '../composables/useResponsive'
import { Client } from '../utils/client'
import { getClientV2, sendClientCommand, sendClientCommandByID, getClientByID, getClientByRunID, fetchFirmwareReleases } from '../api/client'
import type { ClientCommandResp, ClientCommandReq, FirmwareBranch, FirmwareAsset } from '../api/client'
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
import QRCode from 'qrcode'
import { getCachedCurrentNode, setCachedCurrentNode, type CachedCurrentNode } from './currentNodeCache'
import ProxyCard from '../components/ProxyCard.vue'
import type { ProxyStatsInfo } from '../types/proxy'
import type { ServerInfo } from '../types/server'

const route = useRoute()
const router = useRouter()

// The client identity used by the current route. Supports three forms:
//   /clients/:key        -> legacy composite "user.clientID" key
//   /client/:id          -> clientID (preferred, stable across restarts)
//   /client/run/:runID   -> runID (legacy devices without a clientID)
const routeIsRunID = computed(() => !!route.params.runID)
const routeId = computed(() => {
  if (route.params.runID) return route.params.runID as string
  if (route.params.id) return route.params.id as string
  if (route.params.key) return route.params.key as string
  return ''
})

// Send a command to the current client, resolving it by clientID/runID when
// on the new routes, or by the legacy composite key otherwise.
const sendCmd = (req: { command: string; payload?: string }) => {
  const body: ClientCommandReq = { command: req.command, payload: req.payload ?? '' }
  if (route.params.key) {
    return sendClientCommand(route.params.key as string, body)
  }
  return sendClientCommandByID(routeId.value, body)
}
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

// Remote command state
const CUSTOM_PRESET = '__custom__'
const predefinedCommands = [
  { label: 'cmdPasswall', value: 'passwall' },
  { label: 'cmdModifyFrp', value: 'modify_frp', i18nPayload: 'payloadModifyFrp' },
  { label: 'cmdModifySystem', value: 'modify_system', i18nPayload: 'payloadModifySystem' },
  { label: 'cmdUpdateSystem', value: 'update_system', needFirmwareUpdate: true },
]

interface NodeItem {
  id: string
  remarks: string
  type: string
  address: string
  port: string
  active: boolean
}

const selectedPreset = ref('')
const commandInput = ref('')
const payloadInput = ref('')
const commandSending = ref(false)
const commandResp = ref<ClientCommandResp | null>(null)

// 系统重启
const rebooting = ref(false)

// Node list state (for passwall)
const nodeList = ref<NodeItem[]>([])
const nodeListLoading = ref(false)
const nodeListError = ref('')
const passwallEnabled = ref(false)
const passwallRunning = ref(false)
const passwallActiveNode = ref('')
const deviceIp = ref('')
const passwallNodeLoading = ref<Record<string, string>>({})  // remarks -> 'enable'|'disable'|'delete'
const passwallNodeTestState = ref<Record<string, { loading: boolean; code: string; latency: string; error: string; ip: string; location: string; isp: string; ip_country?: string; ip_type?: string; is_isp?: string }>>({})
// 当前节点信息面板（常驻显示）：当前选中并正在运行的节点的 IP / 类型 / 归属地等。
const currentNodeTest = ref<{ code: string; latency: string; error: string; ip: string; location: string; isp: string; ip_country?: string; ip_type?: string; is_isp?: string } | null>(null)
const passwallAddLink = ref('')
const passwallAddLoading = ref(false)
// 分享二维码弹窗状态
const shareDialogVisible = ref(false)
const sharingNode = ref('')
const shareLink = ref('')
const shareQrDataUrl = ref('')
// 更新默认规则：loading 态与最终结果。
const ruleUpdateLoading = ref(false)
const ruleUpdateResp = ref<{ ok: boolean; message: string; duration: number } | null>(null)

const currentCmdConfig = computed(() =>
  predefinedCommands.find((c) => c.value === selectedPreset.value),
)

const isPasswall = computed(() => selectedPreset.value === 'passwall')

const filteredNodeList = computed(() =>
  nodeList.value.filter(n => !n.remarks.includes('分流总节点'))
)

// IP 面板状态机：
// - 运行态：只认当前节点的出口 IP（egress），未查到前一律显示 loading，
//   绝不回退到设备公网 IP（否则会先闪一下设备 IP 再变成出口 IP）。
// - 非运行态：显示设备公网 IP（device），无则 loading。
const ipPanel = computed<{ state: 'egress' | 'device' | 'loading'; ip: string }>(() => {
  if (passwallRunning.value) {
    if (currentNodeTest.value?.ip) {
      return { state: 'egress', ip: currentNodeTest.value.ip }
    }
    return { state: 'loading', ip: '' }
  }
  if (deviceIp.value) {
    return { state: 'device', ip: deviceIp.value }
  }
  return { state: 'loading', ip: '' }
})

// 判断一个 IP 是否为公网 IP（排除内网/环回/链路本地/保留段）。
// 用于决定 deviceIp 用 connIP 还是 clientIP：同一局域网部署时 connIP 是
// 192.168.x.x，不应采用；跨公网时 connIP 即公网 IP，优先采用。
const isPublicIP = (ip: string): boolean => {
  if (!ip) return false
  const match = /^(\d{1,3})\.(\d{1,3})\.(\d{1,3})\.(\d{1,3})$/.exec(ip.trim())
  if (!match) return false
  const [a, b] = [Number(match[1]), Number(match[2])]
  // 10.x / 192.168.x / 172.16-31.x 内网；127 环回；169.254 链路本地；0.x/255 保留
  if (a === 10) return false
  if (a === 192 && b === 168) return false
  if (a === 172 && b >= 16 && b <= 31) return false
  if (a === 127) return false
  if (a === 169 && b === 254) return false
  if (a === 0 || a === 255) return false
  return true
}

// IP 标签统一配色：原生IP/住宅IP 绿(#48c78e)，广播IP/机房IP 红，ip_country 蓝灰区分，未知灰。
// 全部用 effect="plain" + 内联样式，确保大小、字重一致，不会出现某个 tag 偏大。
const IP_GREEN = '#48c78e'
const IP_RED = '#f56565'
const IP_COUNTRY = '#0ea5e9' // 换配色：青蓝（原蓝 #3b82f6）→ 与绿/红明显区分
// 实底 + 白字；原生IP/住宅IP 用用户给定绿 #48c78e，所有 tag 大小字重一致。
const ipGreenTagStyle = { color: '#ffffff', borderColor: IP_GREEN, backgroundColor: IP_GREEN }
const ipRedTagStyle = { color: '#ffffff', borderColor: IP_RED, backgroundColor: IP_RED }
const ipCountryTagStyle = { color: '#ffffff', borderColor: IP_COUNTRY, backgroundColor: IP_COUNTRY }
const ipNeutralTagStyle = { color: '#ffffff', borderColor: '#909399', backgroundColor: '#909399' }
const ipLatencyTagStyle = { color: IP_GREEN, border: 'none' }
const ipTypeTagStyle = (t?: string): Record<string, string> => {
  if (!t) return ipNeutralTagStyle
  if (t.includes('原生')) return ipGreenTagStyle
  if (t.includes('广播') || t.includes('机房')) return ipRedTagStyle
  return ipNeutralTagStyle
}

const needFirmwareUpdate = computed(() => currentCmdConfig.value?.needFirmwareUpdate === true)

// Frp config fields (for modify_frp command)
const isFrpConfig = computed(() => selectedPreset.value === 'modify_frp')
const frpServerAddrPort = ref('')
const frpUser = ref('')
const frpProtocol = ref<'websocket' | 'wss'>('websocket')
const frpTlsEnable = ref<boolean | null>(null)
const frpLoading = ref(false)
// Snapshot of the client's current frp config, used to detect real changes.
// tlsEnable is null when the client has no tls_enable configured (no selection).
const frpInitial = ref<{ addrPort: string; user: string; protocol: string; tlsEnable: boolean | null }>({
  addrPort: '',
  user: '',
  protocol: 'websocket',
  tlsEnable: null,
})

// After sending an frp config command, the client may reconnect under a new
// key (e.g. when only the user changed). These flags drive the UX:
// - reconnecting: waiting for the client to reappear under its new key.
// - returnCountdown: when other settings changed, button shows a countdown
//   before navigating back to the client list.
const reconnecting = ref(false)
const returnCountdown = ref(0)
let returnTimer: ReturnType<typeof setInterval> | null = null

// After a username-only change the client reconnects but keeps the same
// clientID/runID, so the route identity is unchanged. We poll until the client
// reappears AND carries the new username (frpc restart takes time, and the old
// session may still be online for a moment). Only then do we refresh the local
// client data. No navigation is needed.
// `expectedUser` (when provided) is the username we just configured; the
// reconnect is considered complete only once the reconnected client reports
// that exact username.
const waitForReconnect = async (id: string, expectedUser?: string) => {
  reconnecting.value = true
  const deadline = Date.now() + 30000
  const fetchOnce = async () => {
    if (routeIsRunID.value) {
      try {
        return await getClientByRunID(id)
      } catch {
        return null
      }
    }
    try {
      return await getClientByID(id)
    } catch {
      return null
    }
  }
  const tick = async () => {
    if (Date.now() > deadline) {
      reconnecting.value = false
      ElMessage.warning(t('clientDetail.reconnectTimeout'))
      return
    }
    const data = await fetchOnce()
    if (!data) {
      setTimeout(tick, 1500)
      return
    }
    // Wait until the client actually reconnected with the expected username.
    // While the old session is still online (old user), keep waiting.
    if (expectedUser !== undefined && (data.user || '') !== expectedUser) {
      setTimeout(tick, 1500)
      return
    }
    reconnecting.value = false
    client.value = new Client(data)
    // Refresh the proxy list (it is filtered by user) and the frp config
    // form so they reflect the new username immediately.
    fetchProxies()
    fetchFrpConfig()
  }
  tick()
}

// Start the "return to client list" countdown (other frp settings changed).
const startReturnCountdown = () => {
  if (returnTimer !== null) clearInterval(returnTimer)
  returnCountdown.value = 5
  returnTimer = setInterval(() => {
    returnCountdown.value -= 1
    if (returnCountdown.value <= 0) {
      if (returnTimer !== null) { clearInterval(returnTimer); returnTimer = null }
      router.push('/clients')
    }
  }, 1000)
}

// System settings fields (for modify_system command)
const isModifySystem = computed(() => selectedPreset.value === 'modify_system')
const systemWan6 = ref(true)
// systemBands: dynamically discovered WiFi bands (2.4G / 5G / 5.8G / ...).
// Each: { key, label, enabled, loading, ssid, password }
const systemBands = ref<{ key: string; label: string; enabled: boolean; loading: boolean; ssid: string; password: string }[]>([])
const systemSsid = ref('')
// WiFi password state, mirroring the SSID controls.
const systemPassword = ref('')
// target: "all" applies to every band; otherwise a specific band key.
const target = ref('all')
const wan6Loading = ref(false)
const systemLoading = ref(false)

// Default-password switch state (for get_default_password / set_default_password).
const defaultPasswordOn = ref(false)
const defaultPasswordLoading = ref(false)
const defaultPasswordCountdown = ref(0) // seconds remaining on the auto-restore timer

// Common-password switch: only enabled while the device is on the default password.
const commonPasswordOn = ref(false)
const commonPasswordLoading = ref(false)
let defaultPasswordTimer: ReturnType<typeof setInterval> | null = null

const stopDefaultPasswordTimer = () => {
  if (defaultPasswordTimer !== null) {
    clearInterval(defaultPasswordTimer)
    defaultPasswordTimer = null
  }
  defaultPasswordCountdown.value = 0
}

const startDefaultPasswordCountdown = () => {
  stopDefaultPasswordTimer()
  defaultPasswordCountdown.value = 10
  defaultPasswordTimer = setInterval(() => {
    defaultPasswordCountdown.value -= 1
    if (defaultPasswordCountdown.value <= 0) {
      stopDefaultPasswordTimer()
      // The 1-minute window elapsed: the backend has restored the backup,
      // so the device is no longer on the default password. Re-sync state.
      defaultPasswordOn.value = false
    }
  }, 1000)
}

const hasCommandInput = computed(() => {
  if (needFirmwareUpdate.value) return true
  if (isFrpConfig.value) {
    // Only enable send when the form differs from the client's current config.
    if (frpLoading.value) return false
    const init = frpInitial.value
    return (
      frpServerAddrPort.value.trim() !== init.addrPort.trim() ||
      frpUser.value.trim() !== init.user.trim() ||
      frpProtocol.value !== init.protocol ||
      frpTlsEnable.value !== init.tlsEnable
    )
  }
  if (isModifySystem.value) {
    // Any change (toggle or SSID edit) counts as a valid command input.
    return true
  }
  return commandInput.value.trim().length > 0
})

const sendBtnText = computed(() => {
  if (commandSending.value) return t('clientDetail.sending')
  if (needFirmwareUpdate.value) return fwStep.value === 0 ? t('clientDetail.fwStartDetect') : t('clientDetail.sendCommand')
  return t('clientDetail.sendCommand')
})

const onPresetChange = (value: string) => {
  commandResp.value = null
  payloadInput.value = ''
  passwallAddLink.value = ''
  passwallNodeTestState.value = {}
  frpServerAddrPort.value = ''
  frpUser.value = ''
  frpProtocol.value = 'websocket'
  frpTlsEnable.value = null
  systemWan6.value = true
  systemBands.value = []
  systemSsid.value = ''
  systemPassword.value = ''
  target.value = 'all'
  wan6Loading.value = false
  resetFirmwareWizard()
  if (value === CUSTOM_PRESET) {
    commandInput.value = ''
  } else {
    commandInput.value = value
  }
  // Auto-fetch node list for passwall
  if (value === 'passwall') {
    fetchNodeList()
  }
  // Auto-fetch current frp config when entering the frp config panel
  if (value === 'modify_frp' && client.value) {
    fetchFrpConfig()
  }
  // Auto-fetch current system settings when entering system settings panel
  if (value === 'modify_system' && client.value) {
    fetchSystemSettings()
  }
  // Pre-fetch current system version when entering system update (no auto-check)
  if (value === 'update_system' && client.value) {
    sendCmd({ command: 'get_system_version', payload: '' })
      .then((verResp) => {
        if (verResp.result === 'ok') {
          const ver = JSON.parse(verResp.output)
          fwCurrentVersion.value = ver.version || ''
        }
      })
      .catch(() => {
        fwCurrentVersion.value = ''
      })
  }
}

const fetchNodeList = async () => {
  if (!client.value) return
  // Only show the full-screen loading state on the very first load (no list
  // yet). On refresh (add/delete/enable/disable) keep the existing list
  // visible to avoid the layout collapsing into the spinner and back — which
  // caused the visible "jitter" when nodes changed.
  if (nodeList.value.length === 0) {
    nodeListLoading.value = true
  }
  nodeListError.value = ''
  try {
    const resp = await sendCmd({
      command: 'get_nodes',
      payload: '',
    })
    if (resp.result === 'ok') {
      const data = JSON.parse(resp.output)
      nodeList.value = data.nodes || []
      passwallEnabled.value = data.enabled === true
      passwallRunning.value = data.running === true
      passwallActiveNode.value = data.activeNode || ''
      // IP 优先级：
      // 1) connIP（frps 观察到的真实 TCP 源 IP）——跨公网部署时即公网 IP；
      //    但当 frps 与 frpc 在同一局域网（如本地开发）时它是 192.168.x.x，
      //    此时不采用。
      // 2) clientIP（frpc 登录时 publicIP() 上报的公网 IP）——局域网开发时
      //    用它显示设备的真实公网出口 IP。
      // 3) 兜底 connIP，至少展示真实连接源。
      const connIp = client.value?.connIP || ''
      const clientIp = client.value?.ip || ''
      deviceIp.value = isPublicIP(connIp) ? connIp : (clientIp || connIp)
      // 立即显示节点列表，不再等 url_test_device（IP 分类较慢）结束。
      nodeListLoading.value = false
      // 顶部“当前节点信息”面板在后台异步刷新，不阻塞列表渲染。
      fetchCurrentNodeInfo()
    } else {
      nodeListError.value = resp.output || resp.result
    }
  } catch (error: any) {
    nodeListError.value = error.message
  } finally {
    nodeListLoading.value = false
  }
}

// fetchCurrentNodeInfo 刷新顶部“当前节点信息”面板：
// - passwall 运行且有选中节点：对当前节点做完整测试（含出口 IP 类型识别 + 延迟）。
// - 否则（未运行）：用服务端已知的设备公网 IP 调 url_test_device 做同样的 IP 分类，
//   这样未运行时面板显示“客户端 IP + 归属/类型”，与运行时走同一条按 IP 查信息的路。
// IP 信息查询的唯一入口，与“点其他节点测试”（url_test_node_noiip）互不干扰。
const fetchCurrentNodeInfo = async () => {
  if (!client.value) return
  // 缓存键：运行状态 / 选中节点 / 设备IP 任一不同都视为不同上下文。
  // 切回 passwall 标签页时组件重新挂载，但模块级缓存仍在，命中即秒显，
  // 不再走 frpc 实时探测出口 IP（那一步要好几秒）。
  const ctxKey =
    (client.value?.key || '') + ':' +
    (passwallRunning.value && passwallActiveNode.value
      ? 'run:' + passwallActiveNode.value
      : 'dev:' + (deviceIp.value || ''))
  const cached = getCachedCurrentNode(ctxKey)
  if (cached) {
    currentNodeTest.value = cached
    return
  }
  currentNodeTest.value = null
  try {
    let resp
    if (passwallRunning.value && passwallActiveNode.value) {
      resp = await sendCmd({
        command: 'url_test_node',
        payload: passwallActiveNode.value,
      })
    } else {
      // 非运行态（或运行态但暂无选中节点）：不调用 url_test_device，
      // 否则会把“客户端 IP”写进 currentNodeTest，导致顶部面板先显示客户端 IP
      // 再被出口 IP 覆盖。非运行态的客户端 IP 直接由 ipPanel 的 device 分支用
      // deviceIp 渲染，无需经 currentNodeTest。
      return
    }
    if (resp.result === 'ok') {
      const data = JSON.parse(resp.output)
      const value: CachedCurrentNode = {
        code: data.code || '0',
        latency: data.latency || '',
        error: '',
        ip: data.ip || '',
        location: data.location || '',
        isp: data.isp || '',
        ip_country: data.ip_country || '',
        ip_type: data.ip_type || '',
        is_isp: data.is_isp || '',
      }
      currentNodeTest.value = value
      setCachedCurrentNode(ctxKey, value)
    }
  } catch (error: any) {
    console.warn('fetchCurrentNodeInfo failed:', error?.message)
  }
}

const passwallEnableNode = async (id: string) => {
  if (!client.value) return
  passwallNodeLoading.value = { [id]: 'enable' }
  try {
    const resp = await sendCmd({
      command: 'set_node',
      payload: id,
    })
    // 启用节点不展示命令结果块（用户要求），仅用消息提示。
    if (resp.result === 'ok') {
      ElMessage.success(t('clientDetail.commandSuccess'))
    } else {
      ElMessage.error(t('clientDetail.commandFailed', { msg: resp.output || resp.result }))
    }
    await fetchNodeList()
  } catch (error: any) {
    ElMessage.error(t('clientDetail.commandFailed', { msg: error.message }))
  } finally {
    passwallNodeLoading.value = {}
  }
}

const passwallDisableNode = async () => {
  if (!client.value) return
  passwallNodeLoading.value = { __disable__: 'disable' }
  try {
    const resp = await sendCmd({
      command: 'disable_passwall',
      payload: '',
    })
    // 停用节点不展示命令结果块（用户要求），仅用消息提示。
    if (resp.result === 'ok') {
      ElMessage.success(t('clientDetail.commandSuccess'))
    } else {
      ElMessage.error(t('clientDetail.commandFailed', { msg: resp.output || resp.result }))
    }
    await fetchNodeList()
  } catch (error: any) {
    ElMessage.error(t('clientDetail.commandFailed', { msg: error.message }))
  } finally {
    passwallNodeLoading.value = {}
  }
}

const passwallDeleteNode = async (id: string) => {
  if (!client.value) return
  passwallNodeLoading.value = { [id]: 'delete' }
  try {
    const resp = await sendCmd({
      command: 'del_node',
      payload: id,
    })
    // 删除节点不展示命令结果块（用户要求），仅用消息提示。
    if (resp.result === 'ok') {
      ElMessage.success(t('clientDetail.commandSuccess'))
    } else {
      ElMessage.error(t('clientDetail.commandFailed', { msg: resp.output || resp.result }))
    }
    await fetchNodeList()
  } catch (error: any) {
    ElMessage.error(t('clientDetail.commandFailed', { msg: error.message }))
  } finally {
    passwallNodeLoading.value = {}
  }
}

// 列表里点“测试”按钮：skipIP=true 走 url_test_node_noiip，不查 ip-api（IP 信息只由当前节点面板查询）。
// 使用节点唯一 id（uci section）作为 key，避免同备注节点互相覆盖状态。
const passwallURLTest = async (id: string, skipIP = false) => {
  if (!client.value) return
  passwallNodeTestState.value = { ...passwallNodeTestState.value, [id]: { loading: true, code: '', latency: '', error: '', ip: '', location: '', isp: '', ip_country: '', ip_type: '', is_isp: '' } }
  try {
    const resp = await sendCmd({
      command: skipIP ? 'url_test_node_noiip' : 'url_test_node',
      payload: id,
    })
    if (resp.result === 'ok') {
      const data = JSON.parse(resp.output)
      passwallNodeTestState.value = { ...passwallNodeTestState.value, [id]: { loading: false, code: data.code || '0', latency: data.latency || '', error: '', ip: data.ip || '', location: data.location || '', isp: data.isp || '', ip_country: data.ip_country || '', ip_type: data.ip_type || '', is_isp: data.is_isp || '' } }
    } else {
      passwallNodeTestState.value = { ...passwallNodeTestState.value, [id]: { loading: false, code: '', latency: '', error: resp.output || resp.result, ip: '', location: '', isp: '' } }
    }
  } catch (error: any) {
    passwallNodeTestState.value = { ...passwallNodeTestState.value, [id]: { loading: false, code: '', latency: '', error: error.message, ip: '', location: '', isp: '' } }
  }
}

// 分享节点：调用 node_export 获取节点分享链接并生成二维码弹窗。
const shareNode = async (id: string) => {
  if (!client.value) return
  sharingNode.value = id
  shareLink.value = ''
  shareQrDataUrl.value = ''
  try {
    const resp = await sendCmd({
      command: 'node_export',
      payload: id,
    })
    if (resp.result === 'ok' && resp.output) {
      shareLink.value = resp.output
      shareQrDataUrl.value = await QRCode.toDataURL(resp.output, {
        width: 440,
        margin: 1,
        errorCorrectionLevel: 'M',
      })
      shareDialogVisible.value = true
    } else {
      ElMessage.error(t('clientDetail.shareFailed', { msg: resp.output || resp.result }))
    }
  } catch (error: any) {
    ElMessage.error(t('clientDetail.shareFailed', { msg: error.message }))
  } finally {
    sharingNode.value = ''
  }
}

// 复制分享弹窗中的链接
const copyShareLink = async () => {
  if (!shareLink.value) return
  try {
    await navigator.clipboard.writeText(shareLink.value)
    ElMessage.success(t('clientDetail.copySuccess'))
  } catch {
    const textarea = document.createElement('textarea')
    textarea.value = shareLink.value
    textarea.style.position = 'fixed'
    textarea.style.opacity = '0'
    document.body.appendChild(textarea)
    textarea.select()
    document.execCommand('copy')
    document.body.removeChild(textarea)
    ElMessage.success(t('clientDetail.copySuccess'))
  }
}

const passwallAddNode = async () => {
  if (!client.value || !passwallAddLink.value.trim()) return
  passwallAddLoading.value = true
  try {
    const resp = await sendCmd({
      command: 'node_link',
      payload: passwallAddLink.value.trim(),
    })
    // 添加节点不展示命令结果块（与启用/停用/删除一致），仅用消息提示。
    if (resp.result === 'ok') {
      ElMessage.success(t('clientDetail.commandSuccess'))
      passwallAddLink.value = ''
      await fetchNodeList()
    } else {
      ElMessage.error(t('clientDetail.commandFailed', { msg: resp.output || resp.result }))
    }
  } catch (error: any) {
    ElMessage.error(t('clientDetail.commandFailed', { msg: error.message }))
  } finally {
    passwallAddLoading.value = false
  }
}

const payloadPlaceholder = computed(() => {
  const cmd = predefinedCommands.find((c) => c.value === selectedPreset.value)
  if (cmd?.i18nPayload) return t('clientDetail.' + cmd.i18nPayload)
  return t('clientDetail.payloadPlaceholder')
})

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
  if (!routeId.value) {
    loading.value = false
    return false
  }
  try {
    let data
    if (route.params.runID) {
      data = await getClientByRunID(routeId.value)
    } else if (route.params.id) {
      data = await getClientByID(routeId.value)
    } else {
      data = await getClientV2(routeId.value)
    }
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

// Firmware wizard state
interface PlatformInfo { target: string; boardName: string; model: string; boardModel: string; repoApi: string }
interface FwDownloadStatus { status: string; filename: string; totalBytes: number; downloadedBytes: number; progress: number; error?: string }

const fwPlatform = ref<PlatformInfo | null>(null)
const fwCurrentVersion = ref('')
const fwBranches = ref<FirmwareBranch[]>([])
const fwSelectedBranch = ref<number | null>(null)
const fwSelectedFile = ref<number | null>(null)
const fwDownloadStatus = ref<FwDownloadStatus>({ status: 'idle', filename: '', totalBytes: 0, downloadedBytes: 0, progress: 0 })
const fwDownloadStarted = ref(false)
// Track downloaded files keyed by a stable firmware identity (name + size),
// not by list position. This survives list re-fetches/reorders and lets us
// skip re-downloading a firmware we already have. `size` also acts as a
// lightweight integrity check (the "md5-like" guard the user asked for): if a
// re-selected file has the same name but a different size, it's a new build and
// must be downloaded again.
const fwDownloadedFiles: Record<string, { filename: string; totalBytes: number; size: number }> = {}

// Stable identity for a firmware asset: name + size. GitHub release assets have
// no md5 in their API, so size is the cheapest reliable "same file" signal.
const fwFileIdentity = (file: FirmwareAsset): string => `${file.name}|${file.size}`
const fwStep = ref(0)
const fwLoading = ref(false)
const fwError = ref('')
const fwUpgrading = ref(false)
let fwPollTimer: number | null = null

const selectedBranchFiles = computed(() => {
  if (fwSelectedBranch.value === null) return []
  return fwBranches.value[fwSelectedBranch.value]?.assets || []
})

const formatFileSize = (bytes: number): string => {
  if (bytes <= 0) return '0 B'
  const units = ['B', 'KB', 'MB', 'GB']
  const i = Math.floor(Math.log(bytes) / Math.log(1024))
  return (bytes / Math.pow(1024, i)).toFixed(1) + ' ' + units[i]
}

const resetFirmwareWizard = () => {
  fwStep.value = 0
  fwPlatform.value = null
  // Keep fwCurrentVersion: it is prefetched on entry and re-fetched by
  // startFirmwareWizard (which overwrites it), so clearing it here would make
  // the version tag flicker when clicking "check update".
  fwBranches.value = []
  fwSelectedBranch.value = null
  fwSelectedFile.value = null
  fwDownloadStatus.value = { status: 'idle', filename: '', totalBytes: 0, downloadedBytes: 0, progress: 0 }
  fwDownloadStarted.value = false
  fwLoading.value = false
  fwError.value = ''
  fwUpgrading.value = false
  // Clear downloaded files cache on full reset
  for (const key of Object.keys(fwDownloadedFiles)) delete fwDownloadedFiles[key]
  if (fwPollTimer !== null) { clearInterval(fwPollTimer); fwPollTimer = null }
}

const startFirmwareWizard = async () => {
  if (!client.value) return
  resetFirmwareWizard()
  fwStep.value = 1
  fwLoading.value = true
  fwError.value = ''
  try {
    // Fetch current system version independently (best-effort, not blocking).
    try {
      const verResp = await sendCmd({ command: 'get_system_version', payload: '' })
      if (verResp.result === 'ok') {
        const ver = JSON.parse(verResp.output)
        fwCurrentVersion.value = ver.version || ''
      }
      // On failure, keep the previously displayed version instead of clearing
      // it (avoids the tag flickering to "—" on a transient error).
    } catch {
      // Keep the previous value on error.
    }
    const resp = await sendCmd({ command: 'detect_platform', payload: '' })
    if (resp.result === 'ok') {
      fwPlatform.value = JSON.parse(resp.output)
      fwStep.value = 2
      await fetchFirmwareList()
    } else {
      fwError.value = resp.output
    }
  } catch (error: any) {
    fwError.value = error.message
  } finally {
    fwLoading.value = false
  }
}

const fetchFirmwareList = async () => {
  if (!client.value || !fwPlatform.value) return
  fwLoading.value = true
  fwError.value = ''
  try {
    const data = await fetchFirmwareReleases(fwPlatform.value.repoApi, fwPlatform.value.boardModel)
    fwBranches.value = data.branches || []
    if (fwBranches.value.length === 0) fwError.value = t('clientDetail.fwNoBranches')
  } catch (error: any) {
    fwError.value = error.message
  } finally {
    fwLoading.value = false
  }
}

const onBranchSelect = (idx: number) => {
  fwSelectedBranch.value = idx
  fwSelectedFile.value = null
  fwStep.value = 3
  // Reset download state when switching branches
  if (fwPollTimer !== null) { clearInterval(fwPollTimer); fwPollTimer = null }
  fwDownloadStatus.value = { status: 'idle', filename: '', totalBytes: 0, downloadedBytes: 0, progress: 0 }
  fwDownloadStarted.value = false
}

// When a file is selected, check if the same firmware was already downloaded.
// Match by stable identity (name + size). If a file with the same name but a
// different size is selected, it's a different build, so we don't treat it as
// already downloaded.
const onFileSelect = () => {
  if (fwSelectedBranch.value === null || fwSelectedFile.value === null) return
  const file = selectedBranchFiles.value[fwSelectedFile.value]
  if (!file) return
  const id = fwFileIdentity(file)
  const cached = fwDownloadedFiles[id]
  if (cached && cached.size === file.size) {
    fwDownloadStarted.value = true
    fwStep.value = 4
    fwDownloadStatus.value = { status: 'complete', filename: cached.filename, totalBytes: cached.totalBytes, downloadedBytes: cached.totalBytes, progress: 100 }
  }
}

const startDownload = async () => {
  if (!client.value || fwSelectedFile.value === null) return
  const file = selectedBranchFiles.value[fwSelectedFile.value]
  if (!file) return
  // Set downloading immediately so button switches to cancel without delay
  fwDownloadStatus.value = { status: 'downloading', filename: file.name, totalBytes: 0, downloadedBytes: 0, progress: 0 }
  fwDownloadStarted.value = true
  fwStep.value = 4
  try {
    const payload = JSON.stringify({ url: file.url, filename: file.name })
    await sendCmd({ command: 'download_firmware', payload })
    fwPollTimer = window.setInterval(pollDownloadStatus, 2000)
  } catch (error: any) {
    fwDownloadStatus.value = { status: 'error', filename: file.name, totalBytes: 0, downloadedBytes: 0, progress: 0, error: error.message }
  }
}

const pollDownloadStatus = async () => {
  if (!client.value) return
  try {
    const resp = await sendCmd({ command: 'download_status', payload: '' })
    if (resp.result === 'ok') {
      fwDownloadStatus.value = JSON.parse(resp.output)
      if (fwDownloadStatus.value.status === 'complete' || fwDownloadStatus.value.status === 'error' || fwDownloadStatus.value.status === 'cancelled') {
        if (fwPollTimer !== null) { clearInterval(fwPollTimer); fwPollTimer = null }
        // Remember completed downloads (keyed by stable firmware identity) so
        // re-selecting the same firmware shows "立即更新" without re-downloading.
        if (fwDownloadStatus.value.status === 'complete' && fwSelectedBranch.value !== null && fwSelectedFile.value !== null) {
          const file = selectedBranchFiles.value[fwSelectedFile.value]
          if (file) {
            fwDownloadedFiles[fwFileIdentity(file)] = {
              filename: fwDownloadStatus.value.filename,
              totalBytes: fwDownloadStatus.value.totalBytes,
              size: file.size,
            }
          }
        }
      }
    }
  } catch { /* ignore polling errors */ }
}

const cancelDownload = async () => {
  if (!client.value) return
  try {
    await sendCmd({ command: 'cancel_download', payload: '' })
    fwDownloadStatus.value.status = 'cancelled'
    if (fwPollTimer !== null) { clearInterval(fwPollTimer); fwPollTimer = null }
  } catch { /* ignore */ }
}

const runSysupgrade = async () => {
  if (!client.value || !fwDownloadStatus.value.filename) return
  fwUpgrading.value = true
  try {
    const resp = await sendCmd({ command: 'run_sysupgrade', payload: fwDownloadStatus.value.filename })
    commandResp.value = resp
  } catch (error: any) {
    commandResp.value = { command: '', result: 'error', output: error.message }
  }
}

// When a specific WiFi band is picked, load its current name and password into
// the input boxes from the cached values (already fetched via get_system).
const onTargetChange = (key: string) => {
  if (key === 'all') {
    systemSsid.value = ''
    systemPassword.value = ''
    return
  }
  const band = systemBands.value.find((b) => b.key === key)
  systemSsid.value = band && band.ssid ? band.ssid : ''
  systemPassword.value = band && band.password ? band.password : ''
}
// Applied to every band, or just the band chosen in the dropdown.
const ssidSending = ref(false)
const sendWifiChange = async () => {
  if (!client.value) return
  const ssid = systemSsid.value.trim()
  const password = systemPassword.value.trim()
  if (!ssid && !password) return
  ssidSending.value = true
  commandResp.value = null
  try {
    const bands =
      target.value === 'all'
        ? systemBands.value.map((b) => ({ key: b.key }))
        : [{ key: target.value }]
    // Build the payload: send ssid/password only when provided.
    const payload: any = { bands }
    if (ssid) payload.ssid = ssid
    if (password) payload.password = password
    const resp = await sendCmd({
      command: 'modify_system',
      payload: JSON.stringify(payload),
    })
    commandResp.value = resp
    if (resp.result === 'ok') {
      // Update the local cache so switching bands shows the new value.
      systemBands.value = systemBands.value.map((b) => {
        if (target.value !== 'all' && b.key !== target.value) return b
        return {
          ...b,
          ssid: ssid ? ssid : b.ssid,
          password: password ? password : b.password,
        }
      })
      ElMessage.success(t('clientDetail.commandSuccess'))
    } else {
      ElMessage.error(t('clientDetail.commandFailed', { msg: resp.output || resp.result }))
    }
  } catch (error: any) {
    commandResp.value = { command: 'modify_system', result: 'error', output: error.message }
    ElMessage.error(t('clientDetail.commandFailed', { msg: error.message }))
  } finally {
    ssidSending.value = false
  }
}

// Toggle the WAN6 interface: send only wan6, update local value only on success.
const onToggleWan6 = async (value: boolean) => {
  if (!client.value) return
  systemWan6.value = !value
  wan6Loading.value = true
  commandResp.value = null
  try {
    const resp = await sendCmd({
      command: 'modify_system',
      payload: JSON.stringify({ wan6: value }),
    })
    commandResp.value = resp
    if (resp.result === 'ok') {
      systemWan6.value = value
      ElMessage.success(t('clientDetail.commandSuccess'))
    } else {
      ElMessage.error(t('clientDetail.commandFailed', { msg: resp.output || resp.result }))
    }
  } catch (error: any) {
    commandResp.value = { command: 'modify_system', result: 'error', output: error.message }
    ElMessage.error(t('clientDetail.commandFailed', { msg: error.message }))
  } finally {
    wan6Loading.value = false
  }
}

// Toggle a single WiFi band: send only that band, update local value only on success.
const onToggleBand = async (key: string, value: boolean) => {
  if (!client.value) return
  const band = systemBands.value.find((b) => b.key === key)
  if (!band) return
  // Revert optimistically until the command confirms.
  band.enabled = !value
  band.loading = true
  commandResp.value = null
  try {
    const resp = await sendCmd({
      command: 'modify_system',
      payload: JSON.stringify({ bands: [{ key, enabled: value }] }),
    })
    commandResp.value = resp
    if (resp.result === 'ok') {
      band.enabled = value // commit
      ElMessage.success(t('clientDetail.commandSuccess'))
    } else {
      ElMessage.error(t('clientDetail.commandFailed', { msg: resp.output || resp.result }))
    }
  } catch (error: any) {
    commandResp.value = { command: 'modify_system', result: 'error', output: error.message }
    ElMessage.error(t('clientDetail.commandFailed', { msg: error.message }))
  } finally {
    band.loading = false
  }
}

// Fetch current WAN6 / WiFi state and SSID when entering the system settings panel.
const fetchSystemSettings = async () => {
  if (!client.value) return
  systemLoading.value = true
  try {
    const resp = await sendCmd({ command: 'get_system', payload: '' })
    if (resp.result === 'ok') {
      const data = JSON.parse(resp.output)
      if (typeof data.wan6 === 'boolean') systemWan6.value = data.wan6
      if (Array.isArray(data.bands)) {
        systemBands.value = data.bands.map((b: any) => ({
          key: b.key,
          label: b.label,
          enabled: !!b.enabled,
          loading: false,
          ssid: b.ssid || '',
          password: b.password || '',
        }))
      }
      // Also query default-password and common-password states.
      await fetchDefaultPassword()
      await fetchCommonPassword()
    } else {
      ElMessage.warning(t('clientDetail.commandFailed', { msg: resp.output || resp.result }))
    }
  } catch (error: any) {
    ElMessage.warning(t('clientDetail.commandFailed', { msg: error.message }))
  } finally {
    systemLoading.value = false
  }
}

// 系统重启：点击后弹出确认框，确认再调用客户端 reboot 命令。
const rebootSystem = async () => {
  if (!client.value) return
  try {
    await ElMessageBox.confirm(t('clientDetail.sysRebootConfirm'), t('clientDetail.sysReboot'), {
      type: 'warning',
      confirmButtonText: t('clientDetail.sysReboot'),
      cancelButtonText: t('common.cancel'),
    })
  } catch {
    return
  }
  rebooting.value = true
  try {
    const resp = await sendCmd({ command: 'reboot', payload: '' })
    if (resp.result === 'ok') {
      ElMessage.success(t('clientDetail.sysRebooting'))
    } else {
      ElMessage.error(t('clientDetail.commandFailed', { msg: resp.output || resp.result }))
    }
  } catch (error: any) {
    ElMessage.error(t('clientDetail.commandFailed', { msg: error.message }))
  } finally {
    rebooting.value = false
  }
}

// Fetch the client's current frp config (server addr:port, user, protocol,
// tls_enable) when entering the frp config panel, and prefill the form with it.
const fetchFrpConfig = async () => {
  if (!client.value) return
  frpLoading.value = true
  // Reset to defaults while loading to avoid showing stale values.
  frpServerAddrPort.value = ''
  frpUser.value = ''
  frpProtocol.value = 'websocket'
  frpTlsEnable.value = null
  try {
    const resp = await sendCmd({ command: 'get_frp', payload: '' })
    if (resp.result === 'ok') {
      const data = JSON.parse(resp.output)
      const addr = typeof data.serverAddr === 'string' ? data.serverAddr : ''
      const port = typeof data.serverPort === 'number' ? String(data.serverPort) : ''
      const addrPort = [addr, port].filter(Boolean).join(':')
      const protocol = data.protocol === 'wss' ? 'wss' : 'websocket'
      // tlsEnable is null when the client has no tls_enable configured.
      const tlsEnable = typeof data.tlsEnable === 'boolean' ? data.tlsEnable : null
      frpServerAddrPort.value = addrPort
      frpUser.value = typeof data.user === 'string' ? data.user : ''
      frpProtocol.value = protocol
      frpTlsEnable.value = tlsEnable
      // Snapshot the loaded values so we can detect real changes.
      frpInitial.value = { addrPort, user: frpUser.value, protocol, tlsEnable }
    } else {
      ElMessage.warning(t('clientDetail.commandFailed', { msg: resp.output || resp.result }))
    }
  } catch (error: any) {
    ElMessage.warning(t('clientDetail.commandFailed', { msg: error.message }))
  } finally {
    frpLoading.value = false
  }
}

// Query whether the device still uses the factory default root password.
const fetchDefaultPassword = async () => {
  if (!client.value) return
  try {
    const resp = await sendCmd({
      command: 'get_default_password',
      payload: '',
    })
    if (resp.result === 'ok') {
      const data = JSON.parse(resp.output)
      defaultPasswordOn.value = !!data.isDefault
      stopDefaultPasswordTimer()
    }
  } catch (error: any) {
    // Non-fatal: leave the switch at its previous state.
    console.warn('fetchDefaultPassword failed:', error)
  }
}

// Query whether the device's root password already equals the common password.
// Uses the same recompute-and-compare approach as the default-password check,
// because the stored /etc/shadow hash differs per device (random salt).
const fetchCommonPassword = async () => {
  if (!client.value) return
  try {
    const resp = await sendCmd({
      command: 'get_common_password',
      payload: JSON.stringify({ password: commonRootPassword }),
    })
    if (resp.result === 'ok') {
      const data = JSON.parse(resp.output)
      commonPasswordOn.value = !!data.isCommon
    }
  } catch (error: any) {
    // Non-fatal: leave the switch at its previous state.
    console.warn('fetchCommonPassword failed:', error)
  }
}

// Toggle the default-password switch.
const onToggleDefaultPassword = async (val: boolean) => {
  if (!client.value) return
  defaultPasswordLoading.value = true
  try {
    // If the device is currently on the common password, pass it along so the
    // backend can keep the default password permanently (no 10s auto-restore).
    const payload: Record<string, any> = { enable: val }
    if (val && commonPasswordOn.value) payload.common_password = commonRootPassword
    const resp = await sendCmd({
      command: 'set_default_password',
      payload: JSON.stringify(payload),
    })
    if (resp.result === 'ok') {
      defaultPasswordOn.value = val
      // Leaving the default password clears the common-password state.
      if (!val) commonPasswordOn.value = false
      if (val) {
        // The backend only auto-restores when the previous password was unknown.
        // When switching from the common password it keeps the default silently.
        const keepAsDefault = resp.output.includes('不自动恢复')
        if (keepAsDefault) {
          commonPasswordOn.value = false
          stopDefaultPasswordTimer()
        } else {
          startDefaultPasswordCountdown()
        }
        ElMessage.success(resp.output || t('clientDetail.defaultPasswordEnabled'))
      } else {
        stopDefaultPasswordTimer()
        ElMessage.info(resp.output || t('clientDetail.defaultPasswordDisabled'))
      }
    } else {
      ElMessage.error(t('clientDetail.commandFailed', { msg: resp.output || resp.result }))
    }
  } catch (error: any) {
    ElMessage.error(t('clientDetail.commandFailed', { msg: error.message }))
  } finally {
    defaultPasswordLoading.value = false
  }
}

// The plaintext root password written when the "common password" switch is turned on.
// NOTE: OpenWrt hashes the password locally with a random salt, so the stored
// /etc/shadow line differs on every device — we can only send the plaintext and
// let the client hash it with openssl. Replace the value below with the real
// common password you want to use.
const commonRootPassword = 'tk!@1234'

// Toggle the common-password switch. Only meaningful while on the default password:
// turning it on rewrites the root password to the fixed common plaintext.
const onToggleCommonPassword = async (val: boolean) => {
  if (!client.value) return
  if (!defaultPasswordOn.value && val) return // locked unless currently on default password
  commonPasswordLoading.value = true
  commandResp.value = null
  try {
    const resp = await sendCmd({
      command: 'modify_system',
      payload: JSON.stringify({ root_password: commonRootPassword }),
    })
    commandResp.value = resp
    if (resp.result === 'ok') {
      commonPasswordOn.value = val
      // Setting a new root password means the device is no longer on the default
      // password, so clear the default-password state and stop its countdown.
      defaultPasswordOn.value = false
      stopDefaultPasswordTimer()
      ElMessage.success(
        val ? t('clientDetail.commonPasswordEnabled') : t('clientDetail.commonPasswordDisabled'),
      )
    } else {
      ElMessage.error(t('clientDetail.commandFailed', { msg: resp.output || resp.result }))
    }
  } catch (error: any) {
    commandResp.value = { command: 'modify_system', result: 'error', output: error.message }
    ElMessage.error(t('clientDetail.commandFailed', { msg: error.message }))
  } finally {
    commonPasswordLoading.value = false
  }
}

// 更新默认规则：客户端异步执行 rule_update.lua（不带第二个参数），
// 这里发起任务后轮询 update_rules_status 直到结束，再输出结果。
const passwallUpdateRules = async () => {
  if (!client.value) return
  ruleUpdateLoading.value = true
  ruleUpdateResp.value = null
  try {
    const start = await sendCmd({
      command: 'update_rules',
      payload: '',
    })
    if (start.result !== 'ok') {
      ruleUpdateResp.value = {
        ok: false,
        message: start.output || start.result,
        duration: 0,
      }
      ElMessage.error(t('clientDetail.commandFailed', { msg: start.output || start.result }))
      return
    }

    // 规则更新耗时较长，轮询直到 complete/error（最多约 20 分钟）。
    const maxAttempts = 400
    for (let i = 0; i < maxAttempts; i++) {
      await new Promise((resolve) => setTimeout(resolve, 3000))
      let state: any
      try {
        const resp = await sendCmd({
          command: 'update_rules_status',
          payload: '',
        })
        if (resp.result !== 'ok') continue
        state = JSON.parse(resp.output || '{}')
      } catch {
        // 单次轮询失败（网络抖动等）不终止任务，继续下一轮。
        continue
      }

      if (state.status === 'complete' || state.status === 'error') {
        const ok = state.status === 'complete'
        ruleUpdateResp.value = {
          ok,
          message: ok ? state.output || '' : state.error || state.output || '',
          duration: state.duration || 0,
        }
        if (ok) {
          ElMessage.success(t('clientDetail.passwallUpdateRulesOk'))
        } else {
          ElMessage.error(t('clientDetail.commandFailed', { msg: state.error || '' }))
        }
        return
      }
    }

    ruleUpdateResp.value = {
      ok: false,
      message: t('clientDetail.passwallUpdateRulesTimeout'),
      duration: 0,
    }
    ElMessage.error(t('clientDetail.passwallUpdateRulesTimeout'))
  } catch (error: any) {
    ruleUpdateResp.value = { ok: false, message: error.message, duration: 0 }
    ElMessage.error(t('clientDetail.commandFailed', { msg: error.message }))
  } finally {
    ruleUpdateLoading.value = false
  }
}

const sendCommand = async () => {
  if (!client.value) return
  if (needFirmwareUpdate.value) { startFirmwareWizard(); return }
  const command = commandInput.value.trim()
  if (!command) return
  commandSending.value = true
  commandResp.value = null
  let payload = payloadInput.value
  if (isFrpConfig.value) {
    const frpCfg: Record<string, unknown> = {}
    if (frpServerAddrPort.value.trim()) {
      const [host, port] = frpServerAddrPort.value.trim().split(':')
      if (host) frpCfg.serverAddr = host
      if (port) frpCfg.serverPort = parseInt(port, 10)
    }
    if (frpUser.value.trim()) {
      if (/^\d+$/.test(frpUser.value.trim())) {
        ElMessage.error(t('clientDetail.frpUserNotNumeric'))
        commandSending.value = false
        return
      }
      frpCfg.user = frpUser.value.trim()
    }
    if (frpProtocol.value) frpCfg.protocol = frpProtocol.value
    // 仅当协议非 wss 且用户已选择 tls 状态时输出 tls_enable（wss 已隐含 TLS；
    // null 表示客户端原本未配置，保持不发送，避免误写成 true/false）
    if (frpProtocol.value !== 'wss' && frpTlsEnable.value !== null) frpCfg.tls_enable = frpTlsEnable.value
    payload = JSON.stringify(frpCfg)
  }
  try {
    const resp = await sendCmd({ command, payload })
    commandResp.value = resp
    if (resp.result === 'ok') {
      ElMessage.success(t('clientDetail.commandSuccess'))
      if (isFrpConfig.value) {
        const userChanged = frpUser.value.trim() !== frpInitial.value.user.trim()
        const addrChanged = frpServerAddrPort.value.trim() !== frpInitial.value.addrPort.trim()
        const protoChanged = frpProtocol.value !== frpInitial.value.protocol
        const tlsChanged = frpTlsEnable.value !== frpInitial.value.tlsEnable
        if (userChanged && !addrChanged && !protoChanged && !tlsChanged) {
          // Only the username changed: the client will restart and reconnect,
          // but its clientID/runID (and therefore the route) stays the same.
          // Show a "waiting for reconnect" transition and only refresh once the
          // client reappears carrying the new username.
          commandResp.value = null
          waitForReconnect(routeId.value, frpUser.value.trim())
          return
        }
        // Other frp settings changed: the client may go offline briefly. Show
        // a countdown on the send button, then return to the client list.
        startReturnCountdown()
        return
      }
    } else {
      ElMessage.error(t('clientDetail.commandFailed', { msg: resp.output || resp.result }))
    }
  } catch (error: any) {
    ElMessage.error(t('clientDetail.commandFailed', { msg: error.message }))
  } finally {
    commandSending.value = false
  }
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
  if (fwPollTimer !== null) { clearInterval(fwPollTimer); fwPollTimer = null }
  if (returnTimer !== null) { clearInterval(returnTimer); returnTimer = null }
})

onMounted(async () => {
  const ok = await fetchClient()
  if (!ok || !client.value) return

  fetchProxies()
})
</script>

<style scoped>

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
.command-card,
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

/* Command Card */
.command-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 16px 20px;
}

.command-header h2 {
  font-size: 15px;
  font-weight: 500;
  color: var(--text-primary);
  margin: 0;
}

.command-body {
  padding: 0 20px 20px;
}

.command-inputs {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.command-radio-group {
  align-self: flex-start;
}

.command-input,
.payload-input {
  width: 100%;
}

.passwall-section {
  display: flex;
  flex-direction: column;
  width: 100%;
  :deep(.el-divider__text) {
    padding: 0 10px;
  }
}

.passwall-extra-section {
  width: 100%;
  :deep(.el-divider) {
    margin: 20px 0;
  }
}

.passwall-extra-title {
  font-size: 13px;
  font-weight: 600;
  color: var(--text-secondary, #909399);
}

.passwall-rule-bar {
  display: flex;
  align-items: center;
  gap: 10px;
  flex-wrap: wrap;
}

.passwall-rule-hint {
  font-size: 12px;
  color: var(--text-secondary, #909399);
}

.passwall-rule-result {
  margin-top: 12px;
}

.url-test-latency {
  color: var(--text-secondary, #909399);
  font-variant-numeric: tabular-nums;
}

.passwall-current-panel{
  margin-bottom: 10px;
}

.passwall-current-panel .result-label {
  width: 60px;
  flex-shrink: 0;
  text-align: right;
}

/* 面板内所有 tag 统一高度/内边距，避免实底与镂空样式视觉不一致 */
.passwall-current-panel .el-tag {
  height: 22px;
  line-height: 20px;
  padding: 0 6px;
}

.url-test-tags {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 8px;
}

.url-test-tags-group {
  display: inline-flex;
  flex-wrap: wrap;
  gap: 8px;
}

.passwall-node-list {
  display: flex;
  flex-direction: column;
  gap: 0px;
}

.passwall-node-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  flex-wrap: wrap;
  gap: 12px;
  padding: 6px 0;
  border-radius: 8px;
  background: var(--card-bg);
  border: 1px solid var(--border-color);
}

.passwall-node-info {
  display: flex;
  flex-direction: column;
  gap: 2px;
  min-width: 0;
  flex: 1;
}

.passwall-node-info .node-name {
  font-size: 14px;
  font-weight: 500;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.passwall-node-info .node-meta {
  font-size: 12px;
  color: var(--text-secondary);
}

.passwall-node-actions {
  display: flex;
  gap: 0px;
  flex-shrink: 0;
}

.passwall-node-result {
  flex-basis: 100%;
  margin-top: 4px;
  font-size: 12px;
  color: #475569;
  line-height: 1.4;
  word-break: break-all;
}

.passwall-node-result .result-ip {
  font-weight: 600;
  color: #0f172a;
  margin-right: 6px;
}

.passwall-node-result .result-geo {
  color: #64748b;
}

.result-value-ip {
  font-weight: 600;
  font-size: 13px;
  color: var(--text-primary);
  font-family: 'SF Mono', 'Cascadia Code', 'Consolas', monospace;
}

.result-value-geo {
  font-size: 13px;
  color: var(--text-secondary);
}

.result-value-isp {
  font-size: 13px;
  color: var(--text-secondary);
}

.site-tag {
  min-width: 50px;
}

.share-qr-wrap {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 12px;
}

.share-qr-img {
  width: 340px;
  height: 340px;
  border-radius: 8px;
  background: #fff;
  box-sizing: border-box;
}

/* PC 端（≥768px）放大二维码，提升可读性 */
@media (min-width: 768px) {
  .share-qr-img {
    width: 440px;
    height: 440px;
  }
}

.share-qr-copy {
  align-self: center;
}

.url-test-ok {
  color: #16a34a !important;  border-color: rgba(22, 163, 74, 0.3) !important;
  background: rgba(22, 163, 74, 0.06) !important;
}

html.dark .url-test-ok {
  color: #4ade80 !important;
  border-color: rgba(74, 222, 128, 0.3) !important;
  background: rgba(74, 222, 128, 0.1) !important;
}

.url-test-fail {
  color: var(--el-color-danger) !important;
  border-color: rgba(245, 108, 108, 0.3) !important;
  background: rgba(245, 108, 108, 0.06) !important;
}

html.dark .url-test-fail {
  border-color: rgba(245, 108, 108, 0.3) !important;
  background: rgba(245, 108, 108, 0.1) !important;
}

.url-test-loading {
  opacity: 0.8;
}

.passwall-add-section {
  width: 100%;
  :deep(.el-divider) {
    margin: 20px 0;
  }
}

.passwall-add-form {
  display: flex;
  gap: 8px;
  align-items: center;
}

.passwall-add-form .el-input {
  flex: 1;
}

.frp-config-form {
  width: 100%;
  padding: 16px;
  border: 1px solid var(--el-border-color-lighter, #ebeef5);
  border-radius: 10px;
  background: var(--el-fill-color-blank, #fff);
  box-shadow: 0 1px 2px rgba(0, 0, 0, 0.04);
  transition: box-shadow 0.2s ease, border-color 0.2s ease;
}

.frp-config-header {
  display: flex;
  align-items: center;
  gap: 10px;
}

.frp-config-header-icon {
  font-size: 18px;
  color: var(--el-color-primary);
  background: var(--el-color-primary-light-9, #ecf5ff);
  padding: 7px;
  border-radius: 8px;
  flex-shrink: 0;
}

.frp-config-title {
  font-size: 14px;
  font-weight: 600;
  color: var(--el-text-color-primary, #303133);
  line-height: 1.2;
}

.frp-config-subtitle {
  font-size: 12px;
  color: var(--el-text-color-secondary, #909399);
  margin-top: 3px;
  line-height: 1.3;
}

.frp-config-divider.el-divider {
  margin: 14px 0;
}

.frp-config-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill,minmax(268px,1fr));
  gap: 16px 12px;
}

.frp-config-field {
  display: flex;
  flex-direction: column;
  gap: 7px;
  min-width: 0;
}

.frp-config-field--full {
  grid-column: 1 / -1;
}

.frp-config-field--inline {
  flex-direction: row;
  flex-wrap: wrap;
  gap: 12px;
}

.frp-config-inline-item {
  display: flex;
  flex-direction: column;
  gap: 7px;
  min-width: 0;
  :deep(.el-radio-button__inner){
    padding: 8px 12px;
  }
}

.frp-config-label {
  display: flex;
  align-items: center;
  gap: 5px;
  font-size: 13px;
  font-weight: 500;
  color: var(--el-text-color-regular, #606266);
}

.frp-config-label .el-icon {
  font-size: 14px;
  color: var(--el-color-primary, #409eff);
}

/* System settings panel */
.system-toggle-list {
  display: flex;
  flex-direction: column;
  gap: 14px;
  padding: 4px 2px;
}

.system-toggle-list {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(220px, 1fr));
  gap: 10px;
}

.system-toggle-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 10px;
  padding: 8px 12px;
  border-radius: 10px;
  background: var(--el-fill-color-blank, #ffffff);
  border: 1px solid var(--el-border-color-lighter, #ebeef5);
  transition: background-color 0.2s ease, border-color 0.2s ease;
}

.system-toggle-item:hover {
  background: var(--el-fill-color-light, #f5f7fa);
  border-color: var(--el-border-color, #dcdfe6);
}

/* Dark theme: give the rows a subtle lifted surface so they don't blend
   into the page background. */
html.dark .system-toggle-item {
  background: var(--el-bg-color-overlay, #1d1e1f);
  border-color: var(--el-border-color-lighter, #303030);
}

html.dark .system-toggle-item:hover {
  background: var(--el-fill-color-light, #262727);
  border-color: var(--el-border-color, #434343);
}

.system-toggle-label {
  font-size: 14px;
  font-weight: 500;
  color: var(--el-text-color-primary, #303133);
  flex-shrink: 0;
  white-space: nowrap;
}

.system-toggle-ssid {
  font-size: 12px;
  font-weight: 400;
  color: var(--el-text-color-secondary, #909399);
  margin-left: 2px;
}

/* Green when ON, slightly larger than default.
   Switch size is driven by overriding .el-switch__core / .el-switch__action
   directly (with !important) because the --el-switch-* variables are not
   reliably picked up under scoped styles. The knob is positioned manually
   so it stays centered: core 46x24, knob 20x20, 2px padding on each side. */
.system-switch :deep(.el-switch__core) {
  width: 44px !important;
  height: 22px !important;
  border-radius: 10px !important;
}

.system-switch :deep(.el-switch__action) {
  width: 18px !important;
  height: 18px !important;
  left: 1px !important;
}

.system-switch.is-checked :deep(.el-switch__action) {
  left: 23px !important;
}

.system-switch {
  --el-switch-on-color: #21ba45 !important;
  --el-switch-off-color: #c0c4cc !important;
  height: 28px;
}

/* Disabled switch styling: on mobile the default Element Plus disabled
   state is too subtle, so we mute the whole row and give the switch a
   clearly "locked" look so users can tell it isn't tappable. */
.system-switch.is-disabled {
  opacity: 0.55;
  cursor: not-allowed;
}

.system-toggle-item:has(.system-switch.is-disabled) {
  opacity: 0.6;
  background: var(--el-fill-color-light, #f5f7fa);
  border-style: dashed;
  border-color: var(--el-border-color-lighter, #ebeef5);
}

.system-toggle-item:has(.system-switch.is-disabled) .system-toggle-label {
  color: var(--el-text-color-secondary, #909399);
}

html.dark .system-toggle-item:has(.system-switch.is-disabled) {
  background: var(--el-fill-color-dark, #141414);
  border-color: var(--el-border-color-extra-light, #2a2a2a);
}

/* Fallback for browsers without :has() support: mute the label when the
   row's switch is disabled by keying off the item variant instead.
   We add a `.is-disabled-row` class from the template for reliability. */
.system-toggle-item.is-disabled-row {
  opacity: 0.6;
  background: var(--el-fill-color-light, #f5f7fa);
  border-style: dashed;
  border-color: var(--el-border-color-lighter, #ebeef5);
}

.system-toggle-item.is-disabled-row .system-toggle-label {
  color: var(--el-text-color-secondary, #909399);
}

html.dark .system-toggle-item.is-disabled-row {
  background: var(--el-fill-color-dark, #141414);
  border-color: var(--el-border-color-extra-light, #2a2a2a);
}

.system-ssid-field {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.ssid-input-row {
  display: flex;
  align-items: center;
  gap: 12px;
}

.ssid-target-select {
  flex-shrink: 0;
  width: 112px;
}

.ssid-input {
  flex: 1;
}

.ssid-send-btn {
  flex-shrink: 0;
  border-radius: 8px;
  padding-left: 18px;
  padding-right: 18px;
  box-shadow: 0 2px 6px rgba(64, 158, 255, 0.25);
  transition: transform 0.15s ease, box-shadow 0.15s ease;
}

.ssid-send-btn:hover:not(:disabled) {
  transform: translateY(-1px);
  box-shadow: 0 4px 10px rgba(64, 158, 255, 0.35);
}

.ssid-send-btn:active:not(:disabled) {
  transform: translateY(0);
}

.system-reboot-btn {
  margin-left: auto;
}

.node-list-section {
  width: 100%;
}

.node-list-loading,
.node-list-error,
.node-list-empty {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 12px;
  color: var(--text-secondary);
  font-size: 13px;
}

.node-list-error {
  color: var(--el-color-danger);
}

.node-radio-group {
  display: flex;
  flex-direction: column;
  align-items: flex-start;
  gap: 4px;
  width: 100%;
  padding-left: 12px;
}

.node-radio-item {
  display: flex;
  align-items: center;
  padding: 8px 12px;
  border-radius: 6px;
  margin-right: 0;
  height: auto;
  transition: background 0.2s;
}

.node-radio-item:hover {
  background: var(--hover-bg);
}

.node-radio-item :deep(.el-radio__label) {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap;
}

.node-name {
  font-weight: 500;
  word-break: break-all;
}

.node-meta {
  font-size: 12px;
  color: var(--text-secondary);
}

.node-active-tag {
  margin-left: 4px;
}

.firmware-wizard :deep(.el-radio-group) {
  padding-left: 16px;
}

.firmware-wizard :deep(.el-radio) {
  margin-right: 0;
  padding: 6px;
  height: auto;
}

.firmware-wizard :deep(.el-radio__label) {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap;
  padding-left: 8px;
  flex: 1;
  min-width: 0;
}

.firmware-wizard :deep(.el-radio__input) {
  margin-top: 2px;
}

.firmware-wizard .node-name {
  word-break: break-all;
  flex: 1;
  min-width: 0;
}

.fw-step {
  margin-bottom: 16px;
}

.fw-step-title {
  font-size: 14px;
  font-weight: 500;
  margin: 0 0 12px;
  color: var(--text-primary);
}

.fw-step-loading {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 0;
  color: var(--text-secondary);
  font-size: 13px;
}

.fw-platform-info {
  margin-top: 12px;
  margin-bottom: 12px;
  padding: 12px;
  background: var(--hover-bg);
  border-radius: 8px;
}

/* Current version row inside the firmware update card */
.fw-current-version {
  display: flex;
  align-items: baseline;
  font-size: 14px;
  padding: 6px 0 12px;
  color: var(--el-text-color-regular, #606266);
}

.fw-current-version .fw-info-label {
  font-weight: 500;
  color: var(--text-primary);
}

.fw-info-row {
  display: flex;
  gap: 8px;
  padding: 4px 0;
  font-size: 13px;
}

.fw-info-label {
  color: var(--text-secondary);
  min-width: 68px;
}

.fw-error {
  color: var(--el-color-danger);
  font-size: 13px;
  padding: 8px 0;
}

.fw-download-info {
  display: flex;
  justify-content: space-between;
  align-items: center;
  font-size: 12px;
  color: var(--text-secondary);
  margin-top: 8px;
}

.command-result {
  padding: 12px;
  background: var(--hover-bg);
  border-radius: 8px;
}

.result-row {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 8px;
}

.output-row {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.result-label {
  font-size: 13px;
  color: var(--text-secondary);
  white-space: nowrap;
}

.output-text {
  margin: 0;
  padding: 8px;
  background: var(--el-bg-color);
  border-radius: 4px;
  font-size: 12px;
  color: var(--text-primary);
  white-space: pre-wrap;
  word-break: break-all;
  max-height: 200px;
  overflow-y: auto;
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
  .command-card,
  .proxies-card {
    margin-bottom: 10px;
    position: relative;
    z-index: 1;
    background: rgba(255, 255, 255, 0.95);
    backdrop-filter: blur(12px);
    -webkit-backdrop-filter: blur(12px);
    border-color: rgba(255, 255, 255, 0.5);
    box-shadow: 0 2px 14px rgba(0, 0, 0, 0.04);
  }

  .header-card {
    animation: m-fade-up 0.4s ease-out both;
  }

  .command-card {
    animation: m-fade-up 0.4s ease-out 0.06s both;
  }

  .proxies-card {
    animation: m-fade-up 0.4s ease-out 0.12s both;
  }

  html.dark .header-card,
  html.dark .command-card,
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

  /* Remote command menu: fit all four buttons on one line on narrow screens */
  .command-radio-group {
    display: flex;
    width: 100%;
    align-self: stretch;
  }

  .command-radio-group :deep(.el-radio-button) {
    flex: 1 1 0;
  }

  .command-radio-group :deep(.el-radio-button__inner) {
    width: 100%;
    padding: 8px 4px;
    font-size: 12px;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  /* Node list: info on first line, buttons on second line aligned right */
  .passwall-node-item {
    flex-direction: column;
    align-items: stretch;
    gap: 8px;
    padding: 6px;
  }

  .passwall-node-info {
    width: 100%;
    flex: none;
  }

  .passwall-node-actions {
    width: 100%;
    justify-content: flex-end;
    flex-wrap: wrap;
  }

  /* 第一行：IP: + IP 值 + 延迟 tag；第二行：其余 tags 缩进对齐 IP 值 */
  .url-test-tags {
    flex-wrap: wrap;
    align-items: center;
  }
  .url-test-tags .result-label {
    flex: 0 0 auto;
  }
  .url-test-tags .result-value-ip {
    flex: 0 1 auto;
  }
  .url-test-tags-group {
    flex: 1 1 100%;
    margin-left: 68px;
  }

  .geo-isp-info{
    align-items: flex-start;
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

  /* System WIFI settings: stack the select/inputs/button vertically on narrow screens
     so the SSID and password fields are not squeezed out of the viewport */
  .system-ssid-field {
    gap: 10px;
  }

  .ssid-input-row {
    flex-direction: column;
    align-items: stretch;
    gap: 10px;
  }

  .ssid-target-select,
  .ssid-input,
  .ssid-send-btn {
    width: 100%;
    flex: none;
  }

  /* Firmware update: stack branch/file options vertically on narrow screens
     so the long config/branch text and firmware filenames do not overflow */
  .firmware-wizard .node-radio-group {
    padding-left: 0;
  }

  .firmware-wizard .node-radio-item {
    display: flex;
    width: 100%;
    min-width: 0;
    white-space: normal;
    align-items: flex-start;
    flex-wrap: wrap;
    padding: 10px 12px;
  }

  .firmware-wizard :deep(.el-radio__label) {
    display: flex;
    flex-direction: column;
    align-items: flex-start;
    flex: 1;
    width: 100%;
    min-width: 0;
    white-space: normal;
    gap: 6px;
  }

  .firmware-wizard .node-name {
    display: block;
    flex: none;
    width: 100%;
    min-width: 0;
    white-space: normal;
    line-height: 1.5;
    overflow-wrap: anywhere;
    word-break: break-all;
  }

  /* Keep the meta tags on their own wrapped row under the name */
  .firmware-wizard .node-radio-item :deep(.el-tag) {
    margin: 0 6px 0 0;
  }

  .pagination-section {
    padding: 0 14px 14px;
  }
}
</style>
