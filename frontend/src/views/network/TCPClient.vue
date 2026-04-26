<template>
  <v-container fluid class="pa-4">
    <v-row>
      <v-col cols="12" md="5">
        <v-card>
          <v-card-title class="text-subtitle-1 font-weight-bold">TCP 客户端</v-card-title>
          <v-card-text>
            <v-row dense>
              <v-col cols="8"><v-text-field v-model="form.host" label="主机" variant="outlined" density="compact" /></v-col>
              <v-col cols="4"><v-text-field v-model.number="form.port" label="端口" type="number" variant="outlined" density="compact" /></v-col>
              <v-col cols="6"><v-select v-model="form.payloadFormat" :items="['text', 'hex', 'base64']" label="输入格式" variant="outlined" density="compact" /></v-col>
              <v-col cols="6"><v-text-field v-model.number="form.timeoutMillis" label="超时 ms" type="number" variant="outlined" density="compact" /></v-col>
              <v-col cols="12"><v-textarea v-model="form.payload" label="发送内容" rows="8" variant="outlined" class="mono" /></v-col>
            </v-row>
          </v-card-text>
          <v-card-actions>
            <v-btn color="primary" prepend-icon="mdi-console-network" :loading="loading" @click="send">连接并发送</v-btn>
          </v-card-actions>
        </v-card>
      </v-col>
      <v-col cols="12" md="7">
        <v-card class="h-100">
          <v-card-title class="text-subtitle-1 font-weight-bold">响应</v-card-title>
          <v-card-text>
            <v-alert v-if="result?.error" type="warning" variant="tonal" density="compact" class="mb-3">{{ result.error }}</v-alert>
            <div v-if="result" class="mb-2 text-caption">连接: {{ result.connected ? '成功' : '失败' }} · {{ result.latencyMillis }} ms</div>
            <v-textarea :model-value="result?.response || ''" readonly rows="14" variant="outlined" class="mono" />
          </v-card-text>
        </v-card>
      </v-col>
    </v-row>
  </v-container>
</template>

<script setup>
import { reactive, ref } from 'vue'
import { SendTCP } from '../../../wailsjs/go/network/NetworkService'

const form = reactive({ host: '127.0.0.1', port: 80, payload: 'GET / HTTP/1.1\r\nHost: 127.0.0.1\r\n\r\n', payloadFormat: 'text', timeoutMillis: 3000 })
const result = ref(null)
const loading = ref(false)

const send = async () => {
  loading.value = true
  try {
    result.value = await SendTCP({ ...form })
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.mono :deep(textarea), .mono { font-family: ui-monospace, SFMono-Regular, Consolas, monospace; }
</style>
