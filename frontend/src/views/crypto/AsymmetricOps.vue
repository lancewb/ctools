<template>
  <v-container fluid>
    <v-row>
      <v-col cols="12" md="5">
        <v-card>
          <v-card-title class="text-subtitle-1 font-weight-bold">非对称运算</v-card-title>
          <v-card-text>
            <v-row dense>
              <v-col cols="12" md="6">
                <v-select v-model="form.algorithm" :items="algorithms" label="算法" density="comfortable" />
              </v-col>
              <v-col cols="12" md="6">
                <v-select v-model="form.operation" :items="operationOptions" label="操作" density="comfortable" />
              </v-col>
              <v-col cols="12" md="6" v-if="form.algorithm === 'rsa'">
                <v-select v-model="form.padding" :items="paddingOptions" label="RSA 填充" density="comfortable" />
              </v-col>
              <v-col cols="12" md="6" v-if="showOaepOptions">
                <v-select v-model="form.oaepHash" :items="hashOptions" label="OAEP Hash" density="comfortable" />
              </v-col>
              <v-col cols="12" md="6" v-if="showOaepOptions">
                <v-select v-model="form.mgf1Hash" :items="hashOptions" label="MGF1 Hash" density="comfortable" />
              </v-col>
              <v-col cols="12" md="6" v-if="showEccOptions">
                <v-select v-model="form.eccMode" :items="eccModes" label="ECC 模式" density="comfortable" />
              </v-col>
              <v-col cols="12" md="6" v-if="showEccOptions">
                <v-select v-model="form.kdf" :items="kdfOptions" label="KDF" density="comfortable" />
              </v-col>
              <v-col cols="12" md="6" v-if="showEccOptions">
                <v-select v-model="form.symmetricCipher" :items="symOptions" label="对称算法" density="comfortable" />
              </v-col>
              <v-col cols="12" md="6" v-if="showMacOption">
                <v-select v-model="form.macAlgorithm" :items="macOptions" label="MAC 算法" density="comfortable" />
              </v-col>
              <v-col cols="12">
                <v-select v-model="form.keyId" :items="filteredKeyOptions" label="选择密钥" density="comfortable" item-title="label" item-value="value" />
              </v-col>
              <v-col cols="12" md="6">
                <v-select v-model="form.payloadFormat" :items="payloadFormats" label="数据格式" density="comfortable" />
              </v-col>
              <v-col cols="12" md="6">
                <v-select v-model="form.signatureFormat" :items="payloadFormats" label="签名格式" density="comfortable" :disabled="form.operation !== 'verify'" />
              </v-col>
              <v-col cols="12" md="6">
                <v-select v-model="form.outputFormat" :items="outputFormats" label="输出格式" density="comfortable" />
              </v-col>
              <v-col cols="12" md="6">
                <v-text-field v-model="form.uid" label="UID (SM2/SM9)" density="comfortable" />
              </v-col>
              <v-col cols="12">
                <v-textarea v-model="form.payload" label="数据" rows="6" auto-grow />
              </v-col>
              <v-col cols="12" md="6" v-if="['sign','verify'].includes(form.operation)">
                <v-switch
                  v-model="form.payloadIsHash"
                  label="输入为哈希值（不重复计算）"
                  density="comfortable"
                  inset
                  hide-details
                />
              </v-col>
              <v-col cols="12" v-if="form.operation === 'verify'">
                <v-textarea v-model="form.signature" label="签名" rows="3" auto-grow />
              </v-col>
            </v-row>
          </v-card-text>
          <v-card-actions>
            <v-btn color="primary" :loading="running" @click="execute">执行</v-btn>
            <v-spacer />
            <span class="text-error text-caption" v-if="errorMsg">{{ errorMsg }}</span>
          </v-card-actions>
        </v-card>
      </v-col>
      <v-col cols="12" md="7">
        <v-card>
          <v-card-title class="text-subtitle-1 font-weight-bold">输出</v-card-title>
          <v-card-text>
            <div v-if="result">
              <v-alert v-if="result.verified" type="success" variant="tonal" class="mb-3">验签通过</v-alert>
              <v-alert v-else-if="form.operation === 'verify'" type="error" variant="tonal" class="mb-3">验签失败</v-alert>
              <div class="text-caption text-grey-darken-1 mb-1">输出</div>
              <v-textarea :model-value="result.output" rows="10" auto-grow readonly class="font-mono" />
              <div v-if="result.details" class="mt-4">
                <div class="text-caption font-weight-bold mb-1">详细信息</div>
                <v-list density="compact">
                  <v-list-item v-for="(value, key) in result.details" :key="key">
                    <v-list-item-title>{{ key }}</v-list-item-title>
                    <v-list-item-subtitle class="font-mono text-caption">{{ value }}</v-list-item-subtitle>
                  </v-list-item>
                </v-list>
              </div>
            </div>
            <div v-else class="text-grey text-caption">尚未执行操作</div>
          </v-card-text>
        </v-card>
      </v-col>
    </v-row>
  </v-container>
</template>

<script setup>
/**
 * AsymmetricOps Component
 *
 * Provides a UI for performing asymmetric cryptographic operations (RSA, ECC, SM2, SM9).
 * Supports encryption, decryption, signing, and verification.
 */

import { ref, reactive, computed, onMounted, watch } from 'vue'
import { RunAsymmetric, ListStoredKeys } from '../../../wailsjs/go/crypto/CryptoService'

// --- Constants & Options ---
const algorithms = [
  { title: 'RSA', value: 'rsa' },
  { title: 'ECC', value: 'ecc' },
  { title: 'SM2', value: 'sm2' },
  { title: 'SM9', value: 'sm9' }
]
const baseOperations = ['encrypt', 'decrypt', 'sign', 'verify']
const payloadFormats = [
  { title: '文本', value: 'raw' },
  { title: 'Base64', value: 'base64' },
  { title: 'Hex', value: 'hex' }
]
const outputFormats = [
  { title: 'Hex', value: 'hex' },
  { title: 'Base64', value: 'base64' }
]
const hashOptions = [
  { title: 'SHA-1', value: 'sha1' },
  { title: 'SHA-256', value: 'sha256' },
  { title: 'SHA-384', value: 'sha384' },
  { title: 'SHA-512', value: 'sha512' }
]
const eccModes = [
  { title: 'DHAES', value: 'dhaes' },
  { title: 'ECIES-KEM', value: 'ecies' }
]
const kdfOptions = [
  { title: 'SHA-256', value: 'sha256' },
  { title: 'SHA-384', value: 'sha384' },
  { title: 'SHA-512', value: 'sha512' }
]
const symOptions = [
  { title: 'AES-256-GCM', value: 'aes-256-gcm' },
  { title: 'AES-256-CBC', value: 'aes-256-cbc' }
]
const macOptions = [
  { title: 'HMAC-SHA256', value: 'hmac-sha256' }
]
const paddingOptions = [
  { title: 'OAEP (加密)', value: 'oaep' },
  { title: 'PKCS#1 v1.5 (数据)', value: 'pkcs1' },
  { title: 'NONE (无填充)', value: 'none' },
  { title: 'PSS (签名)', value: 'pss' },
  { title: 'DATA (PKCS#1 v1.5 签名)', value: 'data' }
]

// --- State ---
const form = reactive({
  algorithm: 'rsa',
  operation: 'encrypt',
  keyId: '',
  payload: '',
  payloadFormat: 'raw',
  signature: '',
  signatureFormat: 'base64',
  uid: '',
  padding: 'oaep',
  oaepHash: 'sha256',
  mgf1Hash: 'sha256',
  outputFormat: 'hex',
  kdf: 'sha256',
  symmetricCipher: 'aes-256-gcm',
  macAlgorithm: 'hmac-sha256',
  eccMode: 'dhaes',
  payloadIsHash: false
})

const storedKeys = ref([])
const running = ref(false)
const result = ref(null)
const errorMsg = ref('')

// --- Computed Properties ---
const operationOptions = computed(() => baseOperations.map(o => ({ title: o.toUpperCase(), value: o })))
const filteredKeyOptions = computed(() => storedKeys.value
  .filter(k => k.algorithm?.toLowerCase() === form.algorithm)
  .map(k => ({ label: `${k.name} (${k.keyType})`, value: k.id })))
const showOaepOptions = computed(() => form.algorithm === 'rsa' && form.padding === 'oaep')
const showEccOptions = computed(() => form.algorithm === 'ecc')
const showMacOption = computed(() => showEccOptions.value && form.symmetricCipher === 'aes-256-cbc')

// --- Watchers ---
watch(() => form.algorithm, (val) => {
  if (val !== 'rsa') {
    form.padding = 'oaep'
    form.oaepHash = 'sha256'
    form.mgf1Hash = 'sha256'
  }
  if (val !== 'ecc') {
    form.eccMode = 'dhaes'
    form.symmetricCipher = 'aes-256-gcm'
    form.macAlgorithm = 'hmac-sha256'
    form.kdf = 'sha256'
  }
})

watch(() => form.eccMode, (mode) => {
  if (mode === 'ecies' && form.symmetricCipher === 'aes-256-gcm') {
    form.symmetricCipher = 'aes-256-cbc'
  }
  if (mode === 'dhaes' && form.symmetricCipher === 'aes-256-cbc') {
    form.symmetricCipher = 'aes-256-gcm'
  }
})

watch(() => form.symmetricCipher, (sym) => {
  if (sym !== 'aes-256-cbc') {
    form.macAlgorithm = 'hmac-sha256'
  }
})

// --- Methods ---

/**
 * loadKeys fetches the list of stored keys from the backend.
 */
const loadKeys = async () => {
  storedKeys.value = await ListStoredKeys()
  const firstMatch = storedKeys.value.find(k => k.algorithm?.toLowerCase() === form.algorithm)
  if (firstMatch && !form.keyId) {
    form.keyId = firstMatch.id
  }
}

/**
 * execute triggers the asymmetric cryptographic operation in the backend.
 * Validates the form and updates the result state.
 */
const execute = async () => {
  if (!form.keyId) {
    errorMsg.value = '请选择密钥'
    return
  }
  errorMsg.value = ''
  running.value = true
  try {
    const payload = await RunAsymmetric({ ...form })
    result.value = payload
  } catch (err) {
    errorMsg.value = err?.message ?? String(err)
  } finally {
    running.value = false
  }
}

onMounted(() => {
  loadKeys()
})
</script>

<style scoped>
.font-mono {
  font-family: Consolas, 'Courier New', monospace;
}
</style>
