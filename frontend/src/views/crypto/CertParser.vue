<template>
  <v-container fluid>
    <v-row>
      <v-col cols="12" md="5">
        <v-card>
          <v-card-title class="text-subtitle-1 font-weight-bold">证书解析</v-card-title>
          <v-card-text>
            <v-textarea v-model="pem" label="PEM 证书" rows="12" auto-grow />
          </v-card-text>
          <v-card-actions>
            <v-btn color="primary" :loading="loading" @click="parse">解析</v-btn>
            <v-spacer />
            <span class="text-error text-caption" v-if="errorMsg">{{ errorMsg }}</span>
          </v-card-actions>
        </v-card>
      </v-col>
      <v-col cols="12" md="7">
        <v-card>
          <v-card-title class="text-subtitle-1 font-weight-bold">详情</v-card-title>
          <v-card-text>
            <div v-if="result">
              <v-list density="compact">
                <v-list-item title="主题">
                  <v-list-item-subtitle>{{ stringifyMap(result.subject) }}</v-list-item-subtitle>
                </v-list-item>
                <v-list-item title="颁发者">
                  <v-list-item-subtitle>{{ stringifyMap(result.issuer) }}</v-list-item-subtitle>
                </v-list-item>
                <v-list-item title="序列号" :subtitle="result.serial" />
                <v-list-item title="有效期" :subtitle="`${result.notBefore} - ${result.notAfter}`" />
                <v-list-item title="公钥算法" :subtitle="result.publicKeyAlgorithm" />
                <v-list-item title="签名算法" :subtitle="result.signatureAlgorithm" />
              </v-list>
              <div class="mt-3" v-if="result.dnsNames?.length">
                <div class="text-caption font-weight-bold mb-1">DNS</div>
                <v-chip v-for="host in result.dnsNames" :key="host" size="x-small" class="mr-1">{{ host }}</v-chip>
              </div>
              <div v-if="result.keyUsage?.length" class="mt-3">
                <div class="text-caption font-weight-bold mb-1">Key Usage</div>
                <v-chip v-for="ku in result.keyUsage" :key="ku" size="x-small" class="mr-1">{{ ku }}</v-chip>
              </div>
              <div class="mt-4">
                <div class="text-caption text-grey-darken-1 mb-1">Raw Hex</div>
                <v-textarea :model-value="result.rawHex" rows="4" readonly class="font-mono" />
              </div>
            </div>
            <div v-else class="text-caption text-grey">粘贴证书并解析</div>
          </v-card-text>
        </v-card>
      </v-col>
    </v-row>
  </v-container>
</template>

<script setup>
import { ref } from 'vue'
import { ParseCertificate } from '../../../wailsjs/go/crypto/CryptoService'

const pem = ref('')
const result = ref(null)
const errorMsg = ref('')
const loading = ref(false)

const parse = async () => {
  if (!pem.value) {
    errorMsg.value = '请输入证书'
    return
  }
  errorMsg.value = ''
  loading.value = true
  try {
    result.value = await ParseCertificate({ pem: pem.value })
  } catch (err) {
    errorMsg.value = err?.message ?? String(err)
  } finally {
    loading.value = false
  }
}

const stringifyMap = (obj) => {
  if (!obj) return ''
  return Object.entries(obj).map(([k, v]) => `${k}=${v}`).join(', ')
}
</script>

<style scoped>
.font-mono {
  font-family: Consolas, 'Courier New', monospace;
}
</style>
