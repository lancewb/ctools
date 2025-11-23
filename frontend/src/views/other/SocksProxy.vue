<template>
  <v-container fluid>
    <v-row>
      <v-col cols="12" md="5">
        <v-card>
          <v-card-title class="text-subtitle-1 font-weight-bold">SOCKS5 代理服务</v-card-title>
          <v-card-text>
            <v-row dense>
              <v-col cols="12">
                <v-text-field v-model="form.listenIp" label="监听 IP" density="comfortable" />
              </v-col>
              <v-col cols="12">
                <v-text-field v-model.number="form.port" label="端口" type="number" density="comfortable" />
              </v-col>
            </v-row>
          </v-card-text>
          <v-card-actions>
            <v-btn color="primary" :loading="loading" @click="start">启动</v-btn>
            <v-btn color="grey" class="ml-2" :loading="loading" @click="stop">停止</v-btn>
            <v-spacer />
            <v-btn icon="mdi-refresh" variant="text" @click="refresh"></v-btn>
          </v-card-actions>
        </v-card>
      </v-col>
      <v-col cols="12" md="7">
        <v-card>
          <v-card-title class="text-subtitle-1 font-weight-bold">运行状态</v-card-title>
          <v-card-text>
            <div v-if="status?.running">
              <v-alert type="success" variant="tonal" class="mb-3">服务运行中</v-alert>
              <div class="text-caption text-grey mb-1">监听地址</div>
              <div class="text-body-2 mb-2">{{ status.address }}</div>
              <div class="text-caption text-grey mb-1">活动连接</div>
              <div class="text-body-2 mb-2">{{ status.activeConnections }}</div>
              <div class="text-caption text-grey mb-1">最后更新时间</div>
              <div class="text-body-2">{{ status.lastControlMessage }}</div>
            </div>
            <div v-else>
              <v-alert type="info" variant="tonal">服务未运行</v-alert>
            </div>
            <div v-if="status?.error" class="text-error text-caption mt-2">错误: {{ status.error }}</div>
          </v-card-text>
        </v-card>
      </v-col>
    </v-row>
  </v-container>
</template>

<script setup>
/**
 * SocksProxy Component
 *
 * Provides a UI for managing a local SOCKS5 proxy server.
 */

import { ref, reactive, onMounted } from 'vue'
import { StartSocks5Proxy, StopSocks5Proxy, Socks5Status } from '../../../wailsjs/go/other/OtherService'

// --- State ---
const form = reactive({
  listenIp: '0.0.0.0',
  port: 1080
})

const status = ref(null)
const loading = ref(false)

// --- Methods ---

/**
 * refresh fetches the current status of the SOCKS5 proxy.
 */
const refresh = async () => {
  try {
    status.value = await Socks5Status()
  } catch (err) {
    status.value = { running: false, error: err?.message ?? String(err) }
  }
}

/**
 * start initiates the SOCKS5 proxy server.
 */
const start = async () => {
  loading.value = true
  try {
    await StartSocks5Proxy({ ...form })
    status.value = {
      running: true,
      address: `${form.listenIp}:${form.port}`,
      activeConnections: 0,
      error: '',
      lastControlMessage: new Date().toISOString()
    }
  } catch (err) {
    status.value = { running: false, error: err?.message ?? String(err) }
  } finally {
    loading.value = false
  }
}

/**
 * stop halts the SOCKS5 proxy server.
 */
const stop = async () => {
  loading.value = true
  try {
    await StopSocks5Proxy()
    status.value = { running: false }
  } catch (err) {
    status.value = { running: false, error: err?.message ?? String(err) }
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  refresh()
})
</script>
