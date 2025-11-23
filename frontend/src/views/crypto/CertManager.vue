<template>
  <v-container fluid>
    <v-row>
      <v-col cols="12" md="5">
        <v-card>
          <v-card-title class="text-subtitle-1 font-weight-bold">签发证书</v-card-title>
          <v-card-text>
            <v-row dense>
              <v-col cols="12" md="6">
                <v-select v-model="form.algorithm" :items="algorithms" label="算法" density="comfortable" />
              </v-col>
              <v-col cols="12" md="6">
                <v-text-field v-model="form.validDays" label="有效期 (天)" type="number" density="comfortable" />
              </v-col>
              <v-col cols="12">
                <v-text-field v-model="form.commonName" label="Common Name" density="comfortable" />
              </v-col>
              <v-col cols="12">
                <v-text-field v-model="form.usage" label="用途 (server/client)" density="comfortable" />
              </v-col>
            </v-row>
          </v-card-text>
          <v-card-actions>
            <v-btn color="primary" :loading="issuing" @click="issue">签发</v-btn>
            <v-spacer />
            <span class="text-error text-caption" v-if="errorMsg">{{ errorMsg }}</span>
          </v-card-actions>
        </v-card>
        <v-card class="mt-4" v-if="lastResult">
          <v-card-title class="text-subtitle-1 font-weight-bold">本次输出</v-card-title>
          <v-card-text>
            <div v-if="lastResult.rootCa" class="mb-2">
              <v-alert type="info" variant="tonal">已初始化根 CA</v-alert>
            </div>
            <div v-for="cert in lastResult.certificates" :key="cert.id" class="mb-2">
              <div class="text-caption font-weight-bold">{{ cert.name }}</div>
              <div class="text-caption text-grey">序列号: {{ cert.serial }}</div>
            </div>
          </v-card-text>
        </v-card>
      </v-col>
      <v-col cols="12" md="7">
        <v-card>
          <v-card-title class="text-subtitle-1 font-weight-bold d-flex align-center">
            证书列表
            <v-spacer />
            <v-btn icon="mdi-refresh" variant="text" size="small" @click="loadCertificates"></v-btn>
          </v-card-title>
          <v-data-table :headers="headers" :items="certificates" :items-per-page="5" class="text-body-2">
            <template #item.actions="{ item }">
              <v-btn icon="mdi-export" variant="text" size="small" @click="exportCert(item.id)"></v-btn>
              <v-btn icon="mdi-delete" variant="text" size="small" color="error" @click="removeCert(item.id)"></v-btn>
            </template>
            <template #item.usage="{ item }">
              <v-chip size="x-small" color="primary">{{ item.usage }}</v-chip>
            </template>
          </v-data-table>
        </v-card>
      </v-col>
    </v-row>
    <v-dialog v-model="exportDialog" max-width="720">
      <v-card>
        <v-card-title class="text-subtitle-1 font-weight-bold">导出证书</v-card-title>
        <v-card-text v-if="exportedCert">
          <div class="text-caption text-grey mb-2">{{ exportedCert.cert.name }}</div>
          <div class="text-caption font-weight-bold mb-1">Certificate PEM</div>
          <v-textarea :model-value="exportedCert.cert.certPem" rows="6" auto-grow readonly class="font-mono mb-4" />
          <div v-if="exportedCert.key?.privatePem">
            <div class="text-caption font-weight-bold mb-1">Private PEM</div>
            <v-textarea :model-value="exportedCert.key.privatePem" rows="6" auto-grow readonly class="font-mono" />
          </div>
        </v-card-text>
        <v-card-actions>
          <v-spacer />
          <v-btn text @click="exportDialog = false">关闭</v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
  </v-container>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { IssueCertificate, ListCertificates, DeleteCertificate, ExportCertificate } from '../../../wailsjs/go/crypto/CryptoService'

const algorithms = [
  { title: 'RSA', value: 'rsa' },
  { title: 'SM2 (双证书)', value: 'sm2' }
]

const form = reactive({
  algorithm: 'rsa',
  commonName: 'example.local',
  validDays: 365,
  usage: 'server'
})

const issuing = ref(false)
const errorMsg = ref('')
const lastResult = ref(null)
const certificates = ref([])
const headers = [
  { title: '名称', value: 'name' },
  { title: '算法', value: 'algorithm' },
  { title: '用途', value: 'usage' },
  { title: '序列号', value: 'serial' },
  { title: '到期时间', value: 'notAfter' },
  { title: '操作', value: 'actions', sortable: false }
]
const exportDialog = ref(false)
const exportedCert = ref(null)

const loadCertificates = async () => {
  certificates.value = await ListCertificates()
}

const issue = async () => {
  issuing.value = true
  errorMsg.value = ''
  try {
    const payload = await IssueCertificate({ ...form })
    lastResult.value = payload
    await loadCertificates()
  } catch (err) {
    errorMsg.value = err?.message ?? String(err)
  } finally {
    issuing.value = false
  }
}

const removeCert = async (id) => {
  await DeleteCertificate(id)
  await loadCertificates()
}

const exportCert = async (id) => {
  exportedCert.value = await ExportCertificate(id)
  exportDialog.value = true
}

onMounted(() => {
  loadCertificates()
})
</script>
