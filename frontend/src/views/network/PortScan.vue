<template>
  <v-container fluid class="pa-4">
    <v-card>
      <v-card-title class="text-subtitle-1 font-weight-bold">端口扫描</v-card-title>
      <v-card-text>
        <v-row dense>
          <v-col cols="12" md="4"><v-text-field v-model="form.host" label="目标 IP / 域名" variant="outlined" density="compact" /></v-col>
          <v-col cols="12" md="5"><v-text-field v-model="form.ports" label="端口列表" hint="例: 22,80,443,8000-8010" persistent-hint variant="outlined" density="compact" /></v-col>
          <v-col cols="12" md="3"><v-text-field v-model.number="form.timeoutMillis" label="超时 ms" type="number" variant="outlined" density="compact" /></v-col>
        </v-row>
      </v-card-text>
      <v-card-actions>
        <v-btn color="primary" prepend-icon="mdi-radar" :loading="loading" @click="scan">扫描</v-btn>
        <v-chip v-if="results.length" size="small" color="success" variant="tonal">开放 {{ openCount }} / {{ results.length }}</v-chip>
      </v-card-actions>
    </v-card>

    <v-card class="mt-4">
      <v-table density="compact">
        <thead><tr><th>端口</th><th>状态</th><th>耗时</th><th>错误</th></tr></thead>
        <tbody>
          <tr v-for="row in results" :key="row.port">
            <td class="mono">{{ row.port }}</td>
            <td><v-chip size="x-small" :color="row.open ? 'success' : 'grey'" variant="flat">{{ row.open ? 'OPEN' : 'CLOSED' }}</v-chip></td>
            <td>{{ row.latencyMillis }} ms</td>
            <td class="text-caption text-grey">{{ row.error }}</td>
          </tr>
        </tbody>
      </v-table>
    </v-card>
  </v-container>
</template>

<script setup>
import { computed, reactive, ref } from 'vue'
import { ScanPorts } from '../../../wailsjs/go/network/NetworkService'

const form = reactive({ host: '127.0.0.1', ports: '22,80,443,8080', timeoutMillis: 800 })
const results = ref([])
const loading = ref(false)
const openCount = computed(() => results.value.filter(r => r.open).length)

const scan = async () => {
  loading.value = true
  try {
    results.value = await ScanPorts({ ...form })
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.mono { font-family: ui-monospace, SFMono-Regular, Consolas, monospace; }
</style>
