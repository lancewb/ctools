<template>
  <v-container fluid>
    <v-row>
      <v-col cols="12" md="6">
        <v-card>
          <v-card-title class="text-subtitle-1 font-weight-bold">国密 SSL 服务端</v-card-title>
          <v-card-text>
            <v-row dense>
              <v-col cols="12" md="6">
                <v-text-field v-model="serverForm.listenIp" label="监听 IP" density="comfortable" />
              </v-col>
              <v-col cols="12" md="6">
                <v-text-field v-model.number="serverForm.port" label="端口" type="number" density="comfortable" />
              </v-col>
              <v-col cols="12" md="6">
                <v-select v-model="serverForm.signCertId" :items="certOptions" item-title="label" item-value="value" label="签名证书" density="comfortable" />
              </v-col>
              <v-col cols="12" md="6">
                <v-select v-model="serverForm.signKeyId" :items="keyOptions" item-title="label" item-value="value" label="签名私钥" density="comfortable" />
              </v-col>
              <v-col cols="12" md="6">
                <v-select v-model="serverForm.encCertId" :items="certOptions" item-title="label" item-value="value" label="加密证书" density="comfortable" />
              </v-col>
              <v-col cols="12" md="6">
                <v-select v-model="serverForm.encKeyId" :items="keyOptions" item-title="label" item-value="value" label="加密私钥" density="comfortable" />
              </v-col>
              <v-col cols="12">
                <v-switch v-model="serverForm.clientAuth" label="启用客户端双向验证" inset />
              </v-col>
            </v-row>
          </v-card-text>
          <v-card-actions>
            <v-btn color="primary" :loading="serverLoading" @click="startServer">启动</v-btn>
            <v-btn color="grey" class="ml-2" :loading="serverLoading" @click="stopServer">停止</v-btn>
            <v-spacer />
            <v-btn icon="mdi-refresh" variant="text" @click="refreshServer"></v-btn>
          </v-card-actions>
        </v-card>
        <v-card class="mt-4">
          <v-card-title class="text-subtitle-1 font-weight-bold">服务状态</v-card-title>
          <v-card-text>
            <div v-if="serverStatus?.running">
              <v-alert type="success" variant="tonal" class="mb-2">运行中: {{ serverStatus.address }}</v-alert>
              <div class="text-caption text-grey">启动时间: {{ serverStatus.startedAt }}</div>
            </div>
            <div v-else>
              <v-alert type="info" variant="tonal">服务未运行</v-alert>
            </div>
            <div v-if="serverStatus?.error" class="text-error text-caption mt-2">错误: {{ serverStatus.error }}</div>
          </v-card-text>
        </v-card>
      </v-col>
      <v-col cols="12" md="6">
        <v-card>
          <v-card-title class="text-subtitle-1 font-weight-bold">客户端检测</v-card-title>
          <v-card-text>
            <v-row dense>
              <v-col cols="12" md="6">
                <v-text-field v-model="clientForm.serverIp" label="目标 IP" density="comfortable" />
              </v-col>
              <v-col cols="12" md="6">
                <v-text-field v-model.number="clientForm.port" label="端口" type="number" density="comfortable" />
              </v-col>
              <v-col cols="12">
                <v-switch v-model="clientForm.skipVerify" label="跳过证书校验" inset />
              </v-col>
              <v-col cols="12">
                <v-switch v-model="clientForm.enableClientAuth" label="启用客户端证书" inset />
              </v-col>
              <template v-if="clientForm.enableClientAuth">
                <v-col cols="12" md="6">
                  <v-select v-model="clientForm.signCertId" :items="certOptions" item-title="label" item-value="value" label="签名证书" density="comfortable" />
                </v-col>
                <v-col cols="12" md="6">
                  <v-select v-model="clientForm.signKeyId" :items="keyOptions" item-title="label" item-value="value" label="签名私钥" density="comfortable" />
                </v-col>
                <v-col cols="12" md="6">
                  <v-select v-model="clientForm.encCertId" :items="certOptions" item-title="label" item-value="value" label="加密证书" density="comfortable" />
                </v-col>
                <v-col cols="12" md="6">
                  <v-select v-model="clientForm.encKeyId" :items="keyOptions" item-title="label" item-value="value" label="加密私钥" density="comfortable" />
                </v-col>
              </template>
            </v-row>
          </v-card-text>
          <v-card-actions>
            <v-btn color="primary" :loading="clientLoading" @click="runClient">执行检测</v-btn>
            <v-spacer />
            <v-btn icon="mdi-refresh" variant="text" @click="refreshServer"></v-btn>
          </v-card-actions>
        </v-card>
        <v-card class="mt-4">
          <v-card-title class="text-subtitle-1 font-weight-bold">检测结果</v-card-title>
          <v-card-text>
            <div v-if="clientResult">
              <v-alert :type="clientResult.success ? 'success' : 'error'" variant="tonal" class="mb-2">
                {{ clientResult.success ? '成功' : '失败' }}
              </v-alert>
              <div class="text-caption text-grey">时间: {{ clientResult.timestamp }}</div>
              <div class="text-body-2 mt-2">{{ clientResult.message }}</div>
            </div>
            <div v-else class="text-caption text-grey">尚未运行检测</div>
          </v-card-text>
        </v-card>
      </v-col>
    </v-row>
  </v-container>
</template>

<script setup>
import { ref, reactive, onMounted, computed, watch } from 'vue'
import { ListCertificates, ListStoredKeys } from '../../../wailsjs/go/crypto/CryptoService'
import { StartGMSSLServer, StopGMSSLServer, GMSSLServerStatus, RunGMSSLClientTest } from '../../../wailsjs/go/other/OtherService'

const serverForm = reactive({
  listenIp: '0.0.0.0',
  port: 8443,
  signCertId: '',
  signKeyId: '',
  encCertId: '',
  encKeyId: '',
  clientAuth: false
})

const clientForm = reactive({
  serverIp: '127.0.0.1',
  port: 8443,
  enableClientAuth: false,
  skipVerify: true,
  signCertId: '',
  signKeyId: '',
  encCertId: '',
  encKeyId: ''
})

const certificates = ref([])
const keys = ref([])
const serverStatus = ref(null)
const clientResult = ref(null)
const serverLoading = ref(false)
const clientLoading = ref(false)

const certOptions = computed(() => certificates.value.map(c => ({ label: `${c.name} (${c.algorithm})`, value: c.id })))
const keyOptions = computed(() => keys.value.map(k => ({ label: `${k.name} (${k.algorithm})`, value: k.id })))

const ensureDefaults = () => {
  if (!serverForm.signCertId && certOptions.value.length > 0) {
    serverForm.signCertId = certOptions.value[0].value
  }
  if (!serverForm.encCertId && certOptions.value.length > 1) {
    serverForm.encCertId = certOptions.value[1].value
  } else if (!serverForm.encCertId && certOptions.value.length > 0) {
    serverForm.encCertId = certOptions.value[0].value
  }
  if (!serverForm.signKeyId && keyOptions.value.length > 0) {
    serverForm.signKeyId = keyOptions.value[0].value
  }
  if (!serverForm.encKeyId && keyOptions.value.length > 1) {
    serverForm.encKeyId = keyOptions.value[1].value
  } else if (!serverForm.encKeyId && keyOptions.value.length > 0) {
    serverForm.encKeyId = keyOptions.value[0].value
  }
  if (!clientForm.signCertId && certOptions.value.length > 0) {
    clientForm.signCertId = certOptions.value[0].value
  }
  if (!clientForm.encCertId && certOptions.value.length > 1) {
    clientForm.encCertId = certOptions.value[1].value
  }
  if (!clientForm.signKeyId && keyOptions.value.length > 0) {
    clientForm.signKeyId = keyOptions.value[0].value
  }
  if (!clientForm.encKeyId && keyOptions.value.length > 1) {
    clientForm.encKeyId = keyOptions.value[1].value
  }
}

watch(certOptions, () => {
  ensureDefaults()
})

watch(keyOptions, () => {
  ensureDefaults()
})

const loadBaseData = async () => {
  certificates.value = await ListCertificates()
  keys.value = await ListStoredKeys()
  ensureDefaults()
}

const refreshServer = async () => {
  serverStatus.value = await GMSSLServerStatus()
}

const startServer = async () => {
  serverLoading.value = true
  try {
    await StartGMSSLServer({ ...serverForm })
    serverStatus.value = {
      running: true,
      address: `${serverForm.listenIp}:${serverForm.port}`,
      startedAt: new Date().toISOString(),
      error: ''
    }
  } catch (err) {
    serverStatus.value = { running: false, error: err?.message ?? String(err) }
  } finally {
    serverLoading.value = false
  }
}

const stopServer = async () => {
  serverLoading.value = true
  try {
    await StopGMSSLServer()
    serverStatus.value = { running: false }
  } finally {
    serverLoading.value = false
  }
}

const runClient = async () => {
  clientLoading.value = true
  try {
    clientResult.value = await RunGMSSLClientTest({ ...clientForm })
  } catch (err) {
    clientResult.value = {
      success: false,
      message: err?.message ?? String(err),
      timestamp: new Date().toISOString()
    }
  } finally {
    clientLoading.value = false
  }
}

onMounted(async () => {
  await loadBaseData()
  await refreshServer()
})
</script>
