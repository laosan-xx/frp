<template>
  <ProxyView
    :proxies="proxies"
    :loading="loading"
    proxyType="TCPMux"
    @refresh="fetchData"
  />
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { TCPMuxProxy } from '../utils/proxy.js'
import ProxyView from './ProxyView.vue'

let proxies = ref<TCPMuxProxy[]>([])
const loading = ref(false)

const fetchData = () => {
  loading.value = true
  let tcpmuxHTTPConnectPort: number
  let subdomainHost: string
  fetch('../api/serverinfo', { credentials: 'include' })
    .then((res) => {
      return res.json()
    })
    .then((json) => {
      tcpmuxHTTPConnectPort = json.tcpmuxHTTPConnectPort
      subdomainHost = json.subdomainHost

      fetch('../api/proxy/tcpmux', { credentials: 'include' })
        .then((res) => {
          return res.json()
        })
        .then((json) => {
          proxies.value = []
          for (let proxyStats of json.proxies) {
            proxies.value.push(
              new TCPMuxProxy(proxyStats, tcpmuxHTTPConnectPort, subdomainHost),
            )
          }
        })
    })
    .finally(() => {
      loading.value = false
    })
}
fetchData()
</script>

<style></style>
