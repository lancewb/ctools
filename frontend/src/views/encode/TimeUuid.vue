<template>
  <v-container fluid class="pa-4">
    <v-row>
      <v-col cols="12" md="6">
        <v-card>
          <v-card-title class="text-subtitle-1 font-weight-bold">时间戳转换</v-card-title>
          <v-card-text>
            <v-text-field v-model="timestamp" label="Unix 时间戳" variant="outlined" density="compact" />
            <v-table density="compact"><tbody>
              <tr><td>秒</td><td class="mono">{{ seconds }}</td></tr>
              <tr><td>毫秒</td><td class="mono">{{ millis }}</td></tr>
              <tr><td>本地时间</td><td class="mono">{{ localTime }}</td></tr>
              <tr><td>UTC</td><td class="mono">{{ utcTime }}</td></tr>
            </tbody></v-table>
          </v-card-text>
          <v-card-actions><v-btn color="primary" @click="timestamp = Date.now().toString()">当前时间</v-btn></v-card-actions>
        </v-card>
      </v-col>
      <v-col cols="12" md="6">
        <v-card>
          <v-card-title class="text-subtitle-1 font-weight-bold">UUID</v-card-title>
          <v-card-text><v-textarea :model-value="uuids.join('\n')" readonly rows="10" variant="outlined" class="mono" /></v-card-text>
          <v-card-actions><v-btn color="primary" @click="generate">生成 10 个</v-btn></v-card-actions>
        </v-card>
      </v-col>
    </v-row>
  </v-container>
</template>

<script setup>
import { computed, ref } from 'vue'
const timestamp = ref(Date.now().toString())
const uuids = ref([])
const msValue = computed(() => {
  const n = Number(timestamp.value)
  if (!Number.isFinite(n)) return Date.now()
  return timestamp.value.length <= 10 ? n * 1000 : n
})
const seconds = computed(() => Math.floor(msValue.value / 1000))
const millis = computed(() => msValue.value)
const localTime = computed(() => new Date(msValue.value).toLocaleString())
const utcTime = computed(() => {
  try { return new Date(msValue.value).toISOString() } catch { return '-' }
})
const generate = () => { uuids.value = Array.from({ length: 10 }, () => crypto.randomUUID()) }
generate()
</script>

<style scoped>
.mono, .mono :deep(textarea) { font-family: ui-monospace, SFMono-Regular, Consolas, monospace; }
</style>
