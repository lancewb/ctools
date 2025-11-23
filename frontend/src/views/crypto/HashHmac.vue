<template>
  <v-container fluid>
    <v-card>
      <v-card-title class="text-subtitle-1 font-weight-bold">哈希 / HMAC</v-card-title>
      <v-card-text>
        <v-row dense>
          <v-col cols="12" md="4">
            <v-select v-model="form.algorithm" :items="algorithms" label="算法" density="comfortable" />
          </v-col>
          <v-col cols="12" md="4">
            <v-select v-model="form.mode" :items="[{title:'哈希',value:'hash'},{title:'HMAC',value:'hmac'}]" label="模式" density="comfortable" />
          </v-col>
          <v-col cols="12" md="4">
            <v-select v-model="form.inputFormat" :items="formats" label="数据格式" density="comfortable" />
          </v-col>
          <v-col cols="12" md="4">
            <v-select v-model="form.outputFormat" :items="outputFormats" label="输出格式" density="comfortable" />
          </v-col>
          <v-col cols="12" md="6" v-if="form.mode === 'hmac'">
            <v-select v-model="form.keyFormat" :items="formats" label="密钥格式" density="comfortable" />
          </v-col>
          <v-col cols="12" md="6" v-if="form.mode === 'hmac'">
            <v-text-field v-model="form.key" label="HMAC 密钥" density="comfortable" />
          </v-col>
          <v-col cols="12">
            <v-textarea v-model="form.input" label="数据" rows="6" auto-grow />
          </v-col>
        </v-row>
      </v-card-text>
      <v-card-actions>
        <v-btn color="primary" :loading="loading" @click="execute">计算</v-btn>
        <v-spacer />
        <span class="text-error text-caption" v-if="errorMsg">{{ errorMsg }}</span>
      </v-card-actions>
    </v-card>
    <v-card class="mt-4" v-if="result">
      <v-card-title class="text-subtitle-1 font-weight-bold">结果</v-card-title>
      <v-card-text>
        <div class="d-flex flex-column gap-2">
          <div>
            <div class="text-caption text-grey-darken-1 mb-1">十六进制</div>
            <v-textarea :model-value="result.output" rows="3" readonly class="font-mono" />
          </div>
          <div v-if="result.details?.base64">
            <div class="text-caption text-grey-darken-1 mb-1">Base64</div>
            <v-textarea :model-value="result.details.base64" rows="3" readonly class="font-mono" />
          </div>
        </div>
      </v-card-text>
    </v-card>
  </v-container>
</template>

<script setup>
/**
 * HashHmac Component
 *
 * Provides a UI for computing hashes (SHA, SM3, MD5) and HMACs.
 */

import { ref, reactive } from 'vue'
import { RunHash } from '../../../wailsjs/go/crypto/CryptoService'

// --- Constants & Options ---
const algorithms = [
  { title: 'SHA-1', value: 'sha1' },
  { title: 'SHA-256', value: 'sha256' },
  { title: 'SHA-512', value: 'sha512' },
  { title: 'SM3', value: 'sm3' },
  { title: 'MD5', value: 'md5' },
  { title: 'BLAKE2b-256', value: 'blake2b' }
]

const formats = [
  { title: '文本', value: 'raw' },
  { title: 'Base64', value: 'base64' },
  { title: 'Hex', value: 'hex' }
]

const outputFormats = [
  { title: 'Hex', value: 'hex' },
  { title: 'Base64', value: 'base64' }
]

// --- State ---
const form = reactive({
  algorithm: 'sha256',
  mode: 'hash',
  input: '',
  inputFormat: 'raw',
  key: '',
  keyFormat: 'raw',
  outputFormat: 'hex'
})

const result = ref(null)
const errorMsg = ref('')
const loading = ref(false)

// --- Methods ---

/**
 * execute computes the hash or HMAC based on the input form data.
 */
const execute = async () => {
  errorMsg.value = ''
  loading.value = true
  try {
    result.value = await RunHash({ ...form })
  } catch (err) {
    errorMsg.value = err?.message ?? String(err)
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.font-mono {
  font-family: Consolas, 'Courier New', monospace;
}
</style>
