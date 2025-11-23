<template>
  <v-container fluid>
    <v-row>
      <v-col cols="12" md="6">
        <v-card class="mb-4">
          <v-card-title class="text-subtitle-1 font-weight-bold">密钥管理</v-card-title>
          <v-card-text>
            <v-row dense>
              <v-col cols="12" md="6">
                <v-select v-model="parseForm.algorithm" :items="algorithms" label="算法" density="comfortable" />
              </v-col>
              <v-col cols="12" md="6">
                <v-select v-model="parseForm.format" :items="formats" label="格式" density="comfortable" />
              </v-col>
              <v-col cols="12" md="6">
                <v-text-field v-model="parseForm.name" label="命名" density="comfortable" />
              </v-col>
              <v-col cols="12" md="6">
                <v-select v-model="parseForm.variant" :items="variantOptions" label="变体 (SM 专用)" density="comfortable" />
              </v-col>
              <v-col cols="12">
                <v-text-field v-model="usageInput" label="用途 (逗号分隔)" density="comfortable" />
              </v-col>
              <v-col cols="12">
                <v-switch v-model="parseForm.save" label="解析后保存" inset />
              </v-col>
              <v-col cols="12">
                <v-textarea v-model="parseForm.data" label="密钥内容" rows="8" auto-grow />
              </v-col>
            </v-row>
          </v-card-text>
          <v-card-actions>
            <v-btn color="primary" :loading="parseLoading" @click="handleParse">解析</v-btn>
            <v-spacer />
            <span class="text-error text-caption" v-if="parseError">{{ parseError }}</span>
          </v-card-actions>
        </v-card>
        <v-card v-if="parseResult">
          <v-card-title class="text-subtitle-1 font-weight-bold">解析结果</v-card-title>
          <v-card-text>
            <div class="text-caption text-grey mb-2">摘要信息</div>
            <v-list density="compact">
              <v-list-item v-for="(value, key) in parseResult.summary" :key="key" :title="key" :subtitle="value" />
            </v-list>
            <div v-if="parseResult.privatePem" class="mt-4">
              <div class="text-caption font-weight-bold mb-1">Private PEM</div>
              <v-textarea :model-value="parseResult.privatePem" rows="6" auto-grow readonly class="font-mono" />
            </div>
            <div v-if="parseResult.publicPem" class="mt-4">
              <div class="text-caption font-weight-bold mb-1">Public PEM</div>
              <v-textarea :model-value="parseResult.publicPem" rows="4" auto-grow readonly class="font-mono" />
            </div>
          </v-card-text>
        </v-card>
      </v-col>
      <v-col cols="12" md="6">
        <v-card class="mb-4">
          <v-card-title class="text-subtitle-1 font-weight-bold">密钥生成</v-card-title>
          <v-card-text>
            <v-row dense>
              <v-col cols="12" md="6">
                <v-select v-model="genForm.algorithm" :items="algorithms" label="算法" density="comfortable" />
              </v-col>
              <v-col cols="12" md="6">
                <v-text-field v-model.number="genForm.keySize" label="RSA 位数" type="number" density="comfortable" :disabled="genForm.algorithm !== 'rsa'" />
              </v-col>
              <v-col cols="12" md="6">
                <v-select v-model="genForm.curve" :items="curves" label="椭圆曲线" density="comfortable" :disabled="genForm.algorithm !== 'ecc'" />
              </v-col>
              <v-col cols="12" md="6">
                <v-select v-model="genForm.variant" :items="variantOptions" label="变体" density="comfortable" :disabled="genForm.algorithm !== 'sm9'" />
              </v-col>
              <v-col cols="12">
                <v-text-field v-model="genForm.name" label="命名" density="comfortable" />
              </v-col>
              <v-col cols="12">
                <v-switch v-model="genForm.save" label="保存到密钥库" inset />
              </v-col>
            </v-row>
          </v-card-text>
          <v-card-actions>
            <v-btn color="primary" :loading="genLoading" @click="handleGenerate">生成</v-btn>
            <v-spacer />
            <span class="text-success text-caption" v-if="genMessage">{{ genMessage }}</span>
          </v-card-actions>
        </v-card>
        <v-card>
          <v-card-title class="text-subtitle-1 font-weight-bold d-flex align-center">
            密钥库存
            <v-spacer />
            <v-btn size="small" variant="text" icon="mdi-refresh" @click="loadKeys"></v-btn>
          </v-card-title>
          <v-data-table :headers="keyHeaders" :items="storedKeys" :items-per-page="5" class="text-body-2">
            <template #item.usage="{ item }">
              <v-chip v-for="use in item.usage" :key="use" size="x-small" class="mr-1">{{ use }}</v-chip>
            </template>
            <template #item.createdAt="{ item }">
              {{ new Date(item.createdAt).toLocaleString() }}
            </template>
            <template #item.actions="{ item }">
              <v-btn icon="mdi-export" variant="text" size="small" @click="exportKey(item.id)"></v-btn>
              <v-btn icon="mdi-delete" variant="text" size="small" color="error" @click="removeKey(item.id)"></v-btn>
            </template>
          </v-data-table>
        </v-card>
      </v-col>
    </v-row>

    <v-dialog v-model="exportDialog" max-width="640">
      <v-card>
        <v-card-title class="text-subtitle-1 font-weight-bold">导出密钥</v-card-title>
        <v-card-text v-if="exportedKey">
          <div class="text-caption text-grey mb-2">{{ exportedKey.name }}</div>
          <div v-if="exportedKey.privatePem" class="mb-4">
            <div class="text-caption font-weight-bold mb-1">Private PEM</div>
            <v-textarea :model-value="exportedKey.privatePem" rows="6" auto-grow readonly class="font-mono" />
          </div>
          <div v-if="exportedKey.publicPem">
            <div class="text-caption font-weight-bold mb-1">Public PEM</div>
            <v-textarea :model-value="exportedKey.publicPem" rows="4" auto-grow readonly class="font-mono" />
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
import { ref, reactive, watch, onMounted } from 'vue'
import { ParseKey, GenerateKeyPair, ListStoredKeys, DeleteStoredKey, ExportStoredKey } from '../../../wailsjs/go/crypto/CryptoService'

const algorithms = [
  { title: 'RSA', value: 'rsa' },
  { title: 'ECC', value: 'ecc' },
  { title: 'SM2', value: 'sm2' },
  { title: 'SM9', value: 'sm9' }
]

const formats = [
  { title: 'PEM', value: 'pem' },
  { title: 'Base64', value: 'base64' },
  { title: 'Hex', value: 'hex' },
  { title: 'Raw 32 字节', value: 'raw32' },
  { title: 'SDF 结构体', value: 'sdf' }
]

const variantOptions = [
  { title: '默认', value: '' },
  { title: '签名私钥', value: 'sign-user' },
  { title: '签名主密钥', value: 'sign-master' },
  { title: '加密私钥', value: 'encrypt-user' },
  { title: '加密主密钥', value: 'encrypt-master' }
]

const curves = ['P256', 'P384', 'P521']

const parseForm = reactive({
  name: '',
  algorithm: 'rsa',
  format: 'pem',
  variant: '',
  usage: [],
  save: true,
  data: ''
})

const usageInput = ref('')
watch(usageInput, (val) => {
  parseForm.usage = val.split(',').map(v => v.trim()).filter(Boolean)
})

const parseResult = ref(null)
const parseLoading = ref(false)
const parseError = ref('')

const genForm = reactive({
  algorithm: 'rsa',
  keySize: 2048,
  curve: 'P256',
  variant: 'sign-master',
  name: '',
  save: true
})

const genLoading = ref(false)
const genMessage = ref('')

const storedKeys = ref([])
const keyHeaders = [
  { title: '名称', value: 'name' },
  { title: '算法', value: 'algorithm' },
  { title: '类型', value: 'keyType' },
  { title: '用途', value: 'usage' },
  { title: '创建时间', value: 'createdAt' },
  { title: '操作', value: 'actions', sortable: false }
]

const exportDialog = ref(false)
const exportedKey = ref(null)

const handleParse = async () => {
  parseError.value = ''
  parseLoading.value = true
  try {
    const res = await ParseKey({ ...parseForm })
    parseResult.value = res
    if (parseForm.save) {
      await loadKeys()
    }
  } catch (err) {
    parseError.value = err?.message ?? String(err)
  } finally {
    parseLoading.value = false
  }
}

const handleGenerate = async () => {
  genMessage.value = ''
  genLoading.value = true
  try {
    const res = await GenerateKeyPair({ ...genForm })
    parseResult.value = res
    if (genForm.save) {
      await loadKeys()
    }
    genMessage.value = '生成完成'
  } catch (err) {
    genMessage.value = err?.message ?? String(err)
  } finally {
    genLoading.value = false
  }
}

const loadKeys = async () => {
  storedKeys.value = await ListStoredKeys()
}

const removeKey = async (id) => {
  await DeleteStoredKey(id)
  await loadKeys()
}

const exportKey = async (id) => {
  exportedKey.value = await ExportStoredKey(id)
  exportDialog.value = true
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
