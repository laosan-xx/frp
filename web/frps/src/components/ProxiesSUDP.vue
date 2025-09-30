<template>
  <ProxyView
    :proxies="proxies"
    :loading="loading"
    proxyType="SUDP"
    @refresh="fetchData"
  />
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { SUDPProxy } from '../utils/proxy.js'
import ProxyView from './ProxyView.vue'

let proxies = ref<SUDPProxy[]>([])
const loading = ref(false)

const fetchData = () => {
  loading.value = true
  fetch('../api/proxy/sudp', { credentials: 'include' })
    .then((res) => {
      return res.json()
    })
    .then((json) => {
      proxies.value = []
      for (let proxyStats of json.proxies) {
        proxies.value.push(new SUDPProxy(proxyStats))
      }
    })
    .finally(() => {
      loading.value = false
    })
}
fetchData()
</script>

<style></style>
