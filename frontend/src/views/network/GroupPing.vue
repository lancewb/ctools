<template>
  <v-container fluid class="h-100 pa-4 bg-grey-lighten-5">
    <v-row class="h-100">
      <v-col cols="12" md="9" class="d-flex flex-column h-100">
        <v-card elevation="2" rounded="lg" class="mb-4">
          <v-card-text class="d-flex align-center py-3">
            <v-text-field
                v-model="subnetInput"
                label="子网号 (例如 192.168.1)"
                placeholder="192.168.1"
                variant="outlined"
                density="compact"
                hide-details
                class="mr-4"
                prefix=""
                style="max-width: 300px"
                @keyup.enter="startPing"
            ></v-text-field>

            <v-btn
                color="primary"
                prepend-icon="mdi-play"
                @click="startPing"
                :loading="loading"
                :disabled="loading"
                height="40"
                class="mr-6"
            >
              开始扫描
            </v-btn>

<!--            <v-spacer></v-spacer>-->

            <div class="d-flex align-center gap-4 text-caption">
              <div class="d-flex align-center mr-3">
                <div class="legend-box bg-green mr-1"></div> &lt;100ms
              </div>
              <div class="d-flex align-center mr-3">
                <div class="legend-box bg-orange mr-1"></div> 100-1000ms
              </div>
              <div class="d-flex align-center mr-3">
                <div class="legend-box bg-red mr-1"></div> 1s-5s
              </div>
              <div class="d-flex align-center">
                <div class="legend-box bg-white border mr-1"></div> 超时
              </div>
            </div>
          </v-card-text>
        </v-card>

        <v-card elevation="2" rounded="lg" class="flex-grow-1 d-flex flex-column">
          <v-card-title class="text-subtitle-1 font-weight-bold px-4 pt-4 pb-2">
            扫描结果 <span v-if="currentSubnet" class="text-primary">({{ currentSubnet }}.*)</span>
          </v-card-title>
          <v-divider></v-divider>

          <v-card-text class="pa-4 overflow-y-auto" style="max-height: calc(100vh - 220px);">
            <div v-if="results.length === 0 && !loading" class="text-center text-grey mt-10">
              <v-icon size="64" color="grey-lighten-2">mdi-lan-pending</v-icon>
              <p class="mt-2">请输入子网号开始扫描</p>
            </div>

            <div v-else class="ping-grid">
              <div
                  v-for="item in displayList"
                  :key="item.id"
                  :class="['ping-box', getStatusClass(item.latency)]"
                  v-tooltip:top="`IP: ${currentSubnet}.${item.id} 延迟: ${item.latency >= 5000 ? '超时' : item.latency + 'ms'}`"
              >
                {{ item.id }}
              </div>
            </div>
          </v-card-text>
        </v-card>
      </v-col>

      <v-col cols="12" md="3" class="h-100">
        <v-card elevation="2" rounded="lg" class="h-100 d-flex flex-column">
          <v-card-item class="py-3 bg-grey-lighten-4 border-b">
            <div class="d-flex align-center justify-space-between">
              <span class="text-subtitle-2 font-weight-bold">历史记录</span>
              <v-btn
                  icon="mdi-delete-sweep"
                  size="small"
                  variant="text"
                  color="error"
                  title="清空历史"
                  @click="clearAllHistory"
              ></v-btn>
            </div>
          </v-card-item>

          <v-list density="compact" class="flex-grow-1 overflow-y-auto py-0">
            <v-list-item
                v-for="history in historyList"
                :key="history"
                :title="history"
                :subtitle="`点击扫描`"
                @click="reuseHistory(history)"
                :link="!loading"
                :disabled="loading"
            >
              <template v-slot:prepend>
                <v-icon icon="mdi-history" size="small" color="grey"></v-icon>
              </template>
              <template v-slot:append>
                <v-btn
                    icon="mdi-close"
                    size="x-small"
                    variant="text"
                    @click.stop="deleteHistory(history)"
                ></v-btn>
              </template>
            </v-list-item>

            <div v-if="historyList.length === 0" class="text-caption text-center text-grey pa-4">
              暂无历史记录
            </div>
          </v-list>
        </v-card>
      </v-col>
    </v-row>
  </v-container>
</template>

<script setup>
/**
 * GroupPing Component
 *
 * Provides a UI for scanning a subnet (e.g., 192.168.1.x) to check for active hosts using ICMP Ping.
 * Displays results in a grid and manages a history of scanned subnets.
 */

import { ref, computed, onMounted } from 'vue'
import { PingSubnet, GetPingHistory, AddPingHistory, RemovePingHistory, ClearPingHistory } from '../../../wailsjs/go/network/NetworkService.js'

// --- State ---
const subnetInput = ref('')
const currentSubnet = ref('')
const loading = ref(false)
const results = ref([]) // Stores backend results: {id, latency}
const historyList = ref([])

// --- Computed Properties ---
// Generates a display list of 254 items. Uses placeholders if no results yet.
const displayList = computed(() => {
  if (results.value.length > 0) return results.value
  return Array.from({ length: 254 }, (_, i) => ({ id: i + 1, latency: -1 }))
})

// --- Methods ---

/**
 * getStatusClass returns the CSS class based on latency.
 *
 * @param {number} latency - The ping latency in milliseconds.
 * @returns {string} The CSS class string.
 */
const getStatusClass = (latency) => {
  if (latency === -1) return 'bg-grey-lighten-3 text-grey' // Not scanned
  if (latency >= 5000) return 'bg-white border text-grey-lighten-1' // Timeout
  if (latency >= 1000) return 'bg-red text-white' // > 1s
  if (latency >= 100) return 'bg-orange text-white' // 100-1000ms
  return 'bg-green text-white' // < 100ms
}

/**
 * loadHistory fetches the scan history from the backend.
 */
const loadHistory = async () => {
  try {
    historyList.value = await GetPingHistory()
  } catch (e) {
    console.error(e)
  }
}

/**
 * startPing initiates the subnet scan.
 */
const startPing = async () => {
  const target = subnetInput.value.trim()
  if (!target) return

  // Simple validation for 3 octets (e.g. 192.168.1)
  if (target.split('.').length !== 3) {
    alert("请输入正确的三位子网号，如 192.168.1")
    return
  }

  loading.value = true
  currentSubnet.value = target
  results.value = Array.from({ length: 254 }, (_, i) => ({ id: i + 1, latency: -1 }))

  try {
    historyList.value = await AddPingHistory(target)
    results.value = await PingSubnet(target)
  } catch (e) {
    console.error("Ping failed:", e)
  } finally {
    loading.value = false
  }
}

/**
 * reuseHistory populates the input with a history item and starts the scan.
 */
const reuseHistory = (subnet) => {
  subnetInput.value = subnet
  startPing()
}

/**
 * deleteHistory removes an item from the history.
 */
const deleteHistory = async (subnet) => {
  historyList.value = await RemovePingHistory(subnet)
}

/**
 * clearAllHistory clears all scan history.
 */
const clearAllHistory = async () => {
  await ClearPingHistory()
  historyList.value = []
}

onMounted(() => {
  loadHistory()
})
</script>

<style scoped>
.ping-grid {
  display: grid;
  /* 自动填充，最小宽度 40px，非常适合 254 个格子 */
  grid-template-columns: repeat(auto-fill, minmax(46px, 1fr));
  gap: 8px;
}

.ping-box {
  aspect-ratio: 1; /* 保持正方形 */
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 0.85rem;
  font-weight: bold;
  border-radius: 4px;
  cursor: default;
  transition: all 0.2s;
  user-select: none;
}

.ping-box:hover {
  transform: scale(1.15);
  z-index: 2;
  box-shadow: 0 2px 8px rgba(0,0,0,0.15);
}

.legend-box {
  width: 16px;
  height: 16px;
  border-radius: 3px;
}
</style>