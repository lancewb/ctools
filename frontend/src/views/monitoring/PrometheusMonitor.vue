<template>
  <v-container fluid class="pa-4">
    <v-row>
      <v-col cols="12" lg="4">
        <v-card>
          <v-card-title class="text-subtitle-1 font-weight-bold">Prometheus Metrics 监控</v-card-title>
          <v-card-text>
            <v-row dense>
              <v-col cols="7"><v-text-field v-model="form.host" label="IP / 主机" variant="outlined" density="compact" /></v-col>
              <v-col cols="5"><v-text-field v-model.number="form.port" label="端口" type="number" variant="outlined" density="compact" /></v-col>
              <v-col cols="6"><v-select v-model="form.scheme" :items="['http', 'https']" label="协议" variant="outlined" density="compact" /></v-col>
              <v-col cols="6"><v-text-field v-model="form.path" label="路径" variant="outlined" density="compact" /></v-col>
              <v-col cols="6"><v-text-field v-model.number="intervalSeconds" label="间隔秒" type="number" variant="outlined" density="compact" /></v-col>
              <v-col cols="6"><v-text-field v-model.number="form.timeoutMillis" label="超时 ms" type="number" variant="outlined" density="compact" /></v-col>
              <v-col cols="12"><v-text-field v-model="filter" label="指标过滤" placeholder="cpu, memory, http_requests_total" variant="outlined" density="compact" /></v-col>
              <v-col cols="12">
                <v-select
                    v-model="selectedKey"
                    :items="metricKeys"
                    label="图表指标"
                    variant="outlined"
                    density="compact"
                    hide-details
                />
              </v-col>
            </v-row>
          </v-card-text>
          <v-card-actions>
            <v-btn color="primary" :prepend-icon="running ? 'mdi-stop' : 'mdi-play'" :loading="loading" @click="toggle">{{ running ? '停止' : '开始' }}</v-btn>
            <v-btn variant="tonal" prepend-icon="mdi-refresh" :loading="loading" @click="scrape">抓取一次</v-btn>
            <v-btn variant="text" prepend-icon="mdi-delete" @click="clear">清空</v-btn>
          </v-card-actions>
        </v-card>
      </v-col>
      <v-col cols="12" lg="8">
        <v-card class="mb-4">
          <v-card-title class="text-subtitle-1 font-weight-bold d-flex align-center">
            动态图表
            <v-spacer />
            <v-chip v-if="lastResult?.url" size="small" variant="tonal">{{ lastResult.url }}</v-chip>
          </v-card-title>
          <v-card-text>
            <v-alert v-if="lastResult?.error" type="warning" variant="tonal" density="compact" class="mb-3">{{ lastResult.error }}</v-alert>
            <div class="chart-wrap">
              <svg viewBox="0 0 800 260" preserveAspectRatio="none">
                <polyline :points="polyline" fill="none" stroke="#1976d2" stroke-width="3" />
                <line x1="0" y1="230" x2="800" y2="230" stroke="#cfd8dc" />
              </svg>
              <div class="chart-empty" v-if="history.length < 2">等待至少两次采样</div>
            </div>
            <div class="d-flex mt-2 text-caption text-grey">
              <span>样本: {{ history.length }}</span>
              <v-spacer />
              <span v-if="history.length">最新值: {{ latestValue }}</span>
            </div>
          </v-card-text>
        </v-card>

        <v-card>
          <v-card-title class="text-subtitle-1 font-weight-bold">指标列表</v-card-title>
          <v-table density="compact" fixed-header height="360">
            <thead><tr><th>指标</th><th>标签</th><th>值</th></tr></thead>
            <tbody>
              <tr v-for="metric in filteredMetrics" :key="metric.raw" @click="selectedKey = metricKey(metric)" class="cursor-pointer">
                <td class="mono">{{ metric.name }}</td>
                <td class="text-caption">{{ labelsText(metric.labels) }}</td>
                <td class="mono">{{ metric.value }}</td>
              </tr>
            </tbody>
          </v-table>
        </v-card>
      </v-col>
    </v-row>
  </v-container>
</template>

<script setup>
import { computed, onBeforeUnmount, reactive, ref } from 'vue'
import { ScrapePrometheus } from '../../../wailsjs/go/network/NetworkService'

const form = reactive({ host: '127.0.0.1', port: 9100, scheme: 'http', path: '/metrics', timeoutMillis: 5000 })
const intervalSeconds = ref(5)
const filter = ref('')
const loading = ref(false)
const running = ref(false)
const lastResult = ref(null)
const selectedKey = ref('')
const history = ref([])
let timer = null

const metrics = computed(() => lastResult.value?.metrics || [])
const filteredMetrics = computed(() => {
  const q = filter.value.trim().toLowerCase()
  if (!q) return metrics.value.slice(0, 300)
  return metrics.value.filter(m => m.raw.toLowerCase().includes(q)).slice(0, 300)
})
const metricKeys = computed(() => filteredMetrics.value.map(metricKey))
const latestValue = computed(() => history.value.at(-1)?.value ?? '-')
const polyline = computed(() => {
  if (history.value.length < 2) return ''
  const values = history.value.map(p => p.value)
  const min = Math.min(...values)
  const max = Math.max(...values)
  const span = max - min || 1
  return values.map((v, i) => {
    const x = (i / Math.max(values.length - 1, 1)) * 800
    const y = 230 - ((v - min) / span) * 200
    return `${x.toFixed(1)},${y.toFixed(1)}`
  }).join(' ')
})

const labelsText = labels => Object.entries(labels || {}).map(([k, v]) => `${k}=${v}`).join(', ')
const metricKey = metric => `${metric.name}${labelsText(metric.labels) ? '{' + labelsText(metric.labels) + '}' : ''}`

const scrape = async () => {
  loading.value = true
  try {
    const result = await ScrapePrometheus({ ...form })
    lastResult.value = result
    if (!selectedKey.value && result.metrics?.length) {
      selectedKey.value = metricKey(result.metrics[0])
    }
    const selected = result.metrics?.find(m => metricKey(m) === selectedKey.value)
    if (selected) {
      history.value = [...history.value.slice(-119), { at: result.timestamp, value: selected.value }]
    }
  } finally {
    loading.value = false
  }
}

const toggle = async () => {
  running.value = !running.value
  if (running.value) {
    await scrape()
    timer = setInterval(scrape, Math.max(1, intervalSeconds.value) * 1000)
  } else {
    clearInterval(timer)
    timer = null
  }
}

const clear = () => {
  history.value = []
  lastResult.value = null
}

onBeforeUnmount(() => {
  if (timer) clearInterval(timer)
})
</script>

<style scoped>
.mono { font-family: ui-monospace, SFMono-Regular, Consolas, monospace; }
.chart-wrap { position: relative; height: 260px; border: 1px solid #e0e0e0; background: #fff; }
.chart-wrap svg { width: 100%; height: 100%; display: block; }
.chart-empty { position: absolute; inset: 0; display: flex; align-items: center; justify-content: center; color: #90a4ae; }
</style>
