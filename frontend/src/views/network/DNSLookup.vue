<template>
  <v-container fluid class="pa-4">
    <v-row>
      <v-col cols="12" md="4">
        <v-card>
          <v-card-title class="text-subtitle-1 font-weight-bold">DNS 查询</v-card-title>
          <v-card-text>
            <v-text-field v-model="host" label="域名或主机名" density="comfortable" variant="outlined" @keyup.enter="lookup" />
          </v-card-text>
          <v-card-actions>
            <v-btn color="primary" prepend-icon="mdi-magnify" :loading="loading" @click="lookup">查询</v-btn>
          </v-card-actions>
        </v-card>
      </v-col>
      <v-col cols="12" md="8">
        <v-card>
          <v-card-title class="text-subtitle-1 font-weight-bold">结果</v-card-title>
          <v-card-text>
            <v-alert v-if="result?.error" type="warning" variant="tonal" density="compact" class="mb-3">{{ result.error }}</v-alert>
            <v-table v-if="result">
              <tbody>
                <tr><td class="label">CNAME</td><td>{{ result.cname || '-' }}</td></tr>
                <tr><td class="label">A / AAAA</td><td><div v-for="v in result.addresses" :key="v" class="mono">{{ v }}</div></td></tr>
                <tr><td class="label">MX</td><td><div v-for="v in result.mx" :key="v" class="mono">{{ v }}</div></td></tr>
                <tr><td class="label">NS</td><td><div v-for="v in result.ns" :key="v" class="mono">{{ v }}</div></td></tr>
                <tr><td class="label">TXT</td><td><div v-for="v in result.txt" :key="v" class="mono">{{ v }}</div></td></tr>
              </tbody>
            </v-table>
            <v-alert v-else type="info" variant="tonal">输入域名后查询解析记录。</v-alert>
          </v-card-text>
        </v-card>
      </v-col>
    </v-row>
  </v-container>
</template>

<script setup>
import { ref } from 'vue'
import { LookupDNS } from '../../../wailsjs/go/network/NetworkService'

const host = ref('localhost')
const result = ref(null)
const loading = ref(false)

const lookup = async () => {
  loading.value = true
  try {
    result.value = await LookupDNS({ host: host.value })
  } catch (err) {
    result.value = { error: err?.message ?? String(err) }
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.label { width: 120px; color: #607d8b; font-weight: 700; }
.mono { font-family: ui-monospace, SFMono-Regular, Consolas, monospace; word-break: break-all; }
</style>
