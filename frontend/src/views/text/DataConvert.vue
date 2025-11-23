<template>
  <v-container fluid class="h-100 pa-4 bg-grey-lighten-5">
    <v-card elevation="2" rounded="lg" class="h-100 d-flex flex-column">

      <v-tabs
          v-model="activeTab"
          color="primary"
          align-tabs="start"
          bg-color="blue-grey-lighten-5"
      >
        <v-tab value="storage"><v-icon start>mdi-harddisk</v-icon>存储空间转换</v-tab>
        <v-tab value="bandwidth"><v-icon start>mdi-speedometer</v-icon>网络带宽转换</v-tab>
        <v-tab value="pps"><v-icon start>mdi-lan-connect</v-icon>包转发率 (PPS) 计算</v-tab>
      </v-tabs>

      <v-divider></v-divider>

      <v-card-text class="flex-grow-1 pa-4 overflow-y-auto">
        <v-window v-model="activeTab" class="h-100">

          <v-window-item value="storage" class="h-100">
            <v-row>
              <v-col cols="12" md="5">
                <v-card variant="outlined" class="pa-4 bg-white">
                  <div class="text-subtitle-1 font-weight-bold mb-4 text-primary">输入</div>
                  <v-text-field
                      v-model="storageInput"
                      label="数值"
                      type="number"
                      variant="outlined"
                      density="comfortable"
                  ></v-text-field>
                  <v-select
                      v-model="storageUnit"
                      :items="storageUnitOptions"
                      label="单位"
                      variant="outlined"
                      density="comfortable"
                  ></v-select>
                  <v-alert density="compact" type="info" variant="tonal" class="text-caption mt-2">
                    采用 1024 进制 (Windows 标准)，1 Byte = 8 bits
                  </v-alert>
                </v-card>
              </v-col>

              <v-col cols="12" md="1" class="d-flex align-center justify-center">
                <v-icon size="large" color="grey-lighten-1" class="d-none d-md-flex">mdi-arrow-right-bold</v-icon>
                <v-icon size="large" color="grey-lighten-1" class="d-flex d-md-none my-2">mdi-arrow-down-bold</v-icon>
              </v-col>

              <v-col cols="12" md="6">
                <v-card variant="outlined" class="pa-4 bg-white">
                  <div class="text-subtitle-1 font-weight-bold mb-4 text-teal">转换结果</div>
                  <v-row dense>
                    <v-col cols="12" sm="6" v-for="(val, unit) in storageResults" :key="unit">
                      <v-text-field
                          :model-value="val"
                          :label="unit"
                          readonly
                          variant="filled"
                          density="compact"
                          hide-details
                          class="mb-2 input-monospace"
                          bg-color="grey-lighten-4"
                      >
                        <template v-slot:append-inner>
                          <v-icon size="small" @click="copy(val)" class="cursor-pointer">mdi-content-copy</v-icon>
                        </template>
                      </v-text-field>
                    </v-col>
                  </v-row>
                </v-card>
              </v-col>
            </v-row>
          </v-window-item>

          <v-window-item value="bandwidth" class="h-100">
            <v-row>
              <v-col cols="12" md="5">
                <v-card variant="outlined" class="pa-4 bg-white">
                  <div class="text-subtitle-1 font-weight-bold mb-4 text-primary">输入</div>
                  <v-text-field
                      v-model="bwInput"
                      label="带宽数值"
                      type="number"
                      variant="outlined"
                      density="comfortable"
                  ></v-text-field>
                  <v-select
                      v-model="bwUnit"
                      :items="bwUnitOptions"
                      label="单位"
                      variant="outlined"
                      density="comfortable"
                  ></v-select>
                  <v-alert density="compact" type="info" variant="tonal" class="text-caption mt-2">
                    采用 1000 进制 (网络通信标准)，如 1 Gbps = 1000 Mbps
                  </v-alert>
                </v-card>
              </v-col>

              <v-col cols="12" md="1" class="d-flex align-center justify-center">
                <v-icon size="large" color="grey-lighten-1" class="d-none d-md-flex">mdi-arrow-right-bold</v-icon>
              </v-col>

              <v-col cols="12" md="6">
                <v-card variant="outlined" class="pa-4 bg-white">
                  <div class="text-subtitle-1 font-weight-bold mb-4 text-teal">转换结果</div>
                  <v-row dense>
                    <v-col cols="12" sm="6" v-for="(val, unit) in bwResults" :key="unit">
                      <v-text-field
                          :model-value="val"
                          :label="unit"
                          readonly
                          variant="filled"
                          density="compact"
                          hide-details
                          class="mb-2 input-monospace"
                          bg-color="grey-lighten-4"
                      >
                        <template v-slot:append-inner>
                          <v-icon size="small" @click="copy(val)" class="cursor-pointer">mdi-content-copy</v-icon>
                        </template>
                      </v-text-field>
                    </v-col>
                  </v-row>
                </v-card>
              </v-col>
            </v-row>
          </v-window-item>

          <v-window-item value="pps" class="h-100">
            <v-row>
              <v-col cols="12">
                <v-card variant="tonal" color="blue-grey" class="px-4 py-3 d-flex align-center flex-wrap">
                  <span class="font-weight-bold mr-4 text-body-2">参数设置:</span>

                  <div style="width: 220px;" class="mr-2">
                    <v-select
                        v-model="packetSizeSelect"
                        :items="packetSizeOptions"
                        label="帧大小 (Bytes)"
                        density="compact"
                        variant="outlined"
                        hide-details
                        bg-color="white"
                    ></v-select>
                  </div>

                  <div v-if="packetSizeSelect === 'custom'" style="width: 140px;" class="mr-4">
                    <v-text-field
                        v-model.number="packetSizeInput"
                        label="输入大小"
                        type="number"
                        density="compact"
                        variant="outlined"
                        hide-details
                        bg-color="white"
                        suffix="Bytes"
                        min="1"
                    ></v-text-field>
                  </div>

                  <div class="text-caption text-grey-darken-1 mt-2 mt-sm-0">
                    * 计算包含 L1 开销 (20 Bytes: 8B 前导码 + 12B 帧间隙)，即 Wire Speed。
                  </div>
                </v-card>
              </v-col>

              <v-col cols="12" md="6">
                <v-card variant="outlined" class="h-100">
                  <v-card-item>
                    <v-card-title class="text-subtitle-1 text-primary font-weight-bold">
                      带宽 <v-icon icon="mdi-arrow-right" size="small"></v-icon> PPS
                    </v-card-title>
                  </v-card-item>
                  <v-card-text>
                    <v-row>
                      <v-col cols="7">
                        <v-text-field
                            v-model="inputBwForPps"
                            label="带宽"
                            type="number"
                            variant="outlined"
                            density="compact"
                            hide-details
                        ></v-text-field>
                      </v-col>
                      <v-col cols="5">
                        <v-select
                            v-model="unitBwForPps"
                            :items="['Mbps', 'Gbps']"
                            variant="outlined"
                            density="compact"
                            hide-details
                        ></v-select>
                      </v-col>
                    </v-row>

                    <div class="mt-6">
                      <div class="text-caption text-grey">计算结果 (包转发率):</div>
                      <div class="text-h4 font-weight-bold text-teal mt-1">
                        {{ resultPps }}
                      </div>
                      <div class="text-body-2 text-grey-darken-1">pps / packets per second</div>
                      <div class="text-h6 font-weight-medium text-teal-darken-2 mt-2">
                        {{ resultMpps }} <span class="text-body-2 text-grey">Mpps</span>
                      </div>
                    </div>
                  </v-card-text>
                </v-card>
              </v-col>

              <v-col cols="12" md="6">
                <v-card variant="outlined" class="h-100">
                  <v-card-item>
                    <v-card-title class="text-subtitle-1 text-indigo font-weight-bold">
                      PPS <v-icon icon="mdi-arrow-right" size="small"></v-icon> 带宽
                    </v-card-title>
                  </v-card-item>
                  <v-card-text>
                    <v-row>
                      <v-col cols="7">
                        <v-text-field
                            v-model="inputPpsForBw"
                            label="转发率"
                            type="number"
                            variant="outlined"
                            density="compact"
                            hide-details
                        ></v-text-field>
                      </v-col>
                      <v-col cols="5">
                        <v-select
                            v-model="unitPpsForBw"
                            :items="['pps', 'Mpps']"
                            variant="outlined"
                            density="compact"
                            hide-details
                        ></v-select>
                      </v-col>
                    </v-row>

                    <div class="mt-6">
                      <div class="text-caption text-grey">计算结果 (L1 物理带宽):</div>
                      <div class="text-h4 font-weight-bold text-indigo mt-1">
                        {{ resultMbps }}
                      </div>
                      <div class="text-body-2 text-grey-darken-1">Mbps</div>
                      <div class="text-h6 font-weight-medium text-indigo-darken-2 mt-2">
                        {{ resultGbps }} <span class="text-body-2 text-grey">Gbps</span>
                      </div>
                    </div>
                  </v-card-text>
                </v-card>
              </v-col>
            </v-row>
          </v-window-item>
        </v-window>
      </v-card-text>
    </v-card>

    <v-snackbar v-model="showSnackbar" timeout="1500" color="success" location="top">已复制</v-snackbar>
  </v-container>
</template>

<script setup>
/**
 * DataConvert Component
 *
 * Provides utilities for:
 * 1. Storage Unit Conversion (Base 1024)
 * 2. Network Bandwidth Conversion (Base 1000)
 * 3. PPS (Packets Per Second) Calculator based on frame size and bandwidth.
 */

import { ref, computed } from 'vue'

const activeTab = ref('storage')
const showSnackbar = ref(false)

// ----------------------------------------------------
// 1. Storage Conversion Logic (Base 1024)
// ----------------------------------------------------
const storageInput = ref(1)
const storageUnit = ref('GB')
const storageUnitOptions = ['B', 'KB', 'MB', 'GB', 'TB', 'PB', 'bit', 'Kb', 'Mb', 'Gb', 'Tb']

// Base Unit: Bytes
const storageToBytes = computed(() => {
  const val = parseFloat(storageInput.value) || 0
  const u = storageUnit.value

  const K = 1024
  const factors = {
    'B': 1,
    'KB': K,
    'MB': K**2,
    'GB': K**3,
    'TB': K**4,
    'PB': K**5,
    'bit': 1/8,
    'Kb': (K)/8,
    'Mb': (K**2)/8,
    'Gb': (K**3)/8,
    'Tb': (K**4)/8,
  }
  return val * factors[u]
})

const storageResults = computed(() => {
  const bytes = storageToBytes.value
  const K = 1024

  const fmt = (n) => {
    if (n === 0) return '0'
    // Max 4 decimals
    return parseFloat(n.toFixed(4)).toString()
  }

  return {
    'Bytes': fmt(bytes),
    'KB': fmt(bytes / K),
    'MB': fmt(bytes / K**2),
    'GB': fmt(bytes / K**3),
    'TB': fmt(bytes / K**4),
    'bits': fmt(bytes * 8),
    'Gb (Gigabits)': fmt(bytes * 8 / K**3)
  }
})

// ----------------------------------------------------
// 2. Bandwidth Conversion Logic (Base 1000)
// ----------------------------------------------------
const bwInput = ref(1000)
const bwUnit = ref('Mbps')
const bwUnitOptions = ['bps', 'Kbps', 'Mbps', 'Gbps', 'Tbps', 'B/s', 'KB/s', 'MB/s']

// Base Unit: bps
const bwToBps = computed(() => {
  const val = parseFloat(bwInput.value) || 0
  const u = bwUnit.value
  const K = 1000

  const factors = {
    'bps': 1,
    'Kbps': K,
    'Mbps': K**2,
    'Gbps': K**3,
    'Tbps': K**4,
    'B/s': 8,
    'KB/s': 8 * 1024,
    'MB/s': 8 * 1024**2,
  }
  return val * factors[u]
})

const bwResults = computed(() => {
  const bps = bwToBps.value
  const K = 1000
  const K_BIN = 1024

  const fmt = (n) => {
    if (n === 0) return '0'
    return parseFloat(n.toFixed(4)).toString()
  }

  return {
    'bps': fmt(bps),
    'Kbps': fmt(bps / K),
    'Mbps': fmt(bps / K**2),
    'Gbps': fmt(bps / K**3),
    'Tbps': fmt(bps / K**4),
    'MB/s (Download)': fmt(bps / 8 / K_BIN**2),
    'GB/s': fmt(bps / 8 / K_BIN**3)
  }
})

// ----------------------------------------------------
// 3. PPS Calculation Logic
// ----------------------------------------------------
const packetSizeSelect = ref(64) // Selected option
const packetSizeInput = ref(64)  // Custom input

const packetSizeOptions = [
  { title: '64 Bytes (Min)', value: 64 },
  { title: '128 Bytes', value: 128 },
  { title: '256 Bytes', value: 256 },
  { title: '512 Bytes', value: 512 },
  { title: '1024 Bytes', value: 1024 },
  { title: '1280 Bytes', value: 1280 },
  { title: '1518 Bytes (MTU)', value: 1518 },
  { title: '9000 Bytes (Jumbo)', value: 9000 },
  { title: 'Custom...', value: 'custom' },
]

const realPacketSize = computed(() => {
  if (packetSizeSelect.value === 'custom') {
    return parseFloat(packetSizeInput.value) || 64
  }
  return packetSizeSelect.value
})

const OVERHEAD = 20 // L1 Overhead: 8B Preamble + 12B Inter-frame Gap

// 3.1: Bandwidth -> PPS
const inputBwForPps = ref(1)
const unitBwForPps = ref('Gbps')

const calcPps = computed(() => {
  const bwVal = parseFloat(inputBwForPps.value) || 0
  const isGbps = unitBwForPps.value === 'Gbps'

  const bps = bwVal * (isGbps ? 1_000_000_000 : 1_000_000)
  const frameBits = (realPacketSize.value + OVERHEAD) * 8

  return bps / frameBits
})

const resultPps = computed(() => Math.floor(calcPps.value).toLocaleString())
const resultMpps = computed(() => (calcPps.value / 1_000_000).toFixed(6))

// 3.2: PPS -> Bandwidth
const inputPpsForBw = ref(1)
const unitPpsForBw = ref('Mpps')

const calcBw = computed(() => {
  const ppsVal = parseFloat(inputPpsForBw.value) || 0
  const isMpps = unitPpsForBw.value === 'Mpps'

  const pps = ppsVal * (isMpps ? 1_000_000 : 1)
  const frameBits = (realPacketSize.value + OVERHEAD) * 8

  return pps * frameBits
})

const resultMbps = computed(() => (calcBw.value / 1_000_000).toFixed(3))
const resultGbps = computed(() => (calcBw.value / 1_000_000_000).toFixed(6))

/**
 * copy copies text to clipboard.
 */
const copy = (val) => {
  navigator.clipboard.writeText(val)
  showSnackbar.value = true
}
</script>

<style scoped>
.input-monospace :deep(.v-field__input) {
  font-family: 'Consolas', 'Monaco', monospace;
  font-size: 0.9rem;
}
</style>