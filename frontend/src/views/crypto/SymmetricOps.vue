<template>
  <v-container fluid>
    <v-card>
      <v-card-title class="text-subtitle-1 font-weight-bold">对称运算</v-card-title>
      <v-card-text>
        <v-row dense>
          <v-col cols="12" md="3">
            <v-select v-model="form.algorithm" :items="algorithms" label="算法" density="comfortable" />
          </v-col>
          <v-col cols="12" md="3">
            <v-select v-model="form.mode" :items="modeOptions" label="模式" density="comfortable" />
          </v-col>
          <v-col cols="12" md="3">
            <v-select v-model="form.padding" :items="paddingOptions" label="填充" density="comfortable" :disabled="['ctr','gcm','chacha'].includes(form.mode)" />
          </v-col>
          <v-col cols="12" md="3">
            <v-select v-model="form.operation" :items="[{title:'加密',value:'encrypt'},{title:'解密',value:'decrypt'}]" label="操作" density="comfortable" />
          </v-col>
          <v-col cols="12" md="4">
            <v-select v-model="form.keyFormat" :items="formats" label="密钥格式" density="comfortable" />
          </v-col>
          <v-col cols="12" md="8">
            <v-text-field v-model="form.key" label="密钥" density="comfortable" />
          </v-col>
          <v-col cols="12" md="4" v-if="requiresIV">
            <v-select v-model="form.ivFormat" :items="formats" label="IV 格式" density="comfortable" />
          </v-col>
          <v-col cols="12" md="8" v-if="requiresIV">
            <v-text-field v-model="form.iv" label="IV" density="comfortable" />
          </v-col>
          <v-col cols="12" md="4" v-if="requiresNonce">
            <v-select v-model="form.nonceFormat" :items="formats" label="Nonce 格式" density="comfortable" />
          </v-col>
          <v-col cols="12" md="8" v-if="requiresNonce">
            <v-text-field v-model="form.nonce" label="Nonce" density="comfortable" />
          </v-col>
          <v-col cols="12" md="4" v-if="requiresAAD">
            <v-select v-model="form.additionalDataFormat" :items="formats" label="附加数据格式" density="comfortable" />
          </v-col>
          <v-col cols="12" md="8" v-if="requiresAAD">
            <v-text-field v-model="form.additionalData" label="附加数据 (可选)" density="comfortable" />
          </v-col>
          <v-col cols="12" md="4">
            <v-select v-model="form.inputFormat" :items="formats" label="数据格式" density="comfortable" />
          </v-col>
          <v-col cols="12" md="8">
            <v-textarea v-model="form.input" label="数据" rows="6" auto-grow />
          </v-col>
          <v-col cols="12" md="4">
            <v-select v-model="form.outputFormat" :items="outputFormats" label="输出格式" density="comfortable" />
          </v-col>
        </v-row>
      </v-card-text>
      <v-card-actions>
        <v-btn color="primary" :loading="loading" @click="execute">执行</v-btn>
        <v-spacer />
        <span class="text-error text-caption" v-if="errorMsg">{{ errorMsg }}</span>
      </v-card-actions>
    </v-card>
    <v-card class="mt-4">
      <v-card-title class="text-subtitle-1 font-weight-bold">结果</v-card-title>
      <v-card-text>
        <div v-if="result">
          <div class="text-caption text-grey-darken-1 mb-1">Output</div>
          <v-textarea :model-value="result.output" rows="8" auto-grow readonly class="font-mono" />
          <div v-if="result.details" class="mt-3">
            <div class="text-caption font-weight-bold mb-1">Details</div>
            <v-list density="compact">
              <v-list-item v-for="(value, key) in result.details" :key="key">
                <v-list-item-title>{{ key }}</v-list-item-title>
                <v-list-item-subtitle class="font-mono">{{ value }}</v-list-item-subtitle>
              </v-list-item>
            </v-list>
          </div>
        </div>
        <div v-else class="text-grey text-caption">尚未执行</div>
      </v-card-text>
    </v-card>
  </v-container>
</template>

<script setup>
import { ref, reactive, computed, watch } from 'vue'
import { RunSymmetric } from '../../../wailsjs/go/crypto/CryptoService'

const algorithms = [
  { title: 'AES', value: 'aes' },
  { title: 'SM4', value: 'sm4' },
  { title: '3DES', value: '3des' },
  { title: 'ChaCha20-Poly1305', value: 'chacha20' }
]

const modeOptions = [
  { title: 'CBC', value: 'cbc' },
  { title: 'CTR', value: 'ctr' },
  { title: 'GCM', value: 'gcm' },
  { title: 'ECB', value: 'ecb' },
  { title: 'ChaCha20', value: 'chacha' }
]

const paddingOptions = [
  { title: 'PKCS7', value: 'pkcs7' },
  { title: 'Zero', value: 'zero' },
  { title: 'None', value: 'none' }
]

const formats = [
  { title: '文本', value: 'raw' },
  { title: 'Base64', value: 'base64' },
  { title: 'Hex', value: 'hex' }
]

const form = reactive({
  algorithm: 'aes',
  mode: 'cbc',
  padding: 'pkcs7',
  operation: 'encrypt',
  key: '',
  keyFormat: 'hex',
  iv: '',
  ivFormat: 'hex',
  nonce: '',
  nonceFormat: 'hex',
  additionalData: '',
  additionalDataFormat: 'raw',
  input: '',
  inputFormat: 'raw',
  outputFormat: 'hex'
})

const result = ref(null)
const errorMsg = ref('')
const loading = ref(false)

const outputFormats = [
  { title: 'Hex', value: 'hex' },
  { title: 'Base64', value: 'base64' }
]

const requiresIV = computed(() => ['cbc', 'ctr', 'ecb'].includes(form.mode))
const requiresNonce = computed(() => form.mode === 'gcm' || form.algorithm === 'chacha20')
const requiresAAD = computed(() => requiresNonce.value)

watch(() => form.algorithm, (val) => {
  if (val === 'chacha20') {
    form.mode = 'chacha'
    form.padding = 'none'
  } else if (form.mode === 'chacha') {
    form.mode = 'cbc'
  }
})

const execute = async () => {
  errorMsg.value = ''
  loading.value = true
  try {
    const payload = await RunSymmetric({ ...form, additionalData: form.additionalData })
    result.value = payload
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
