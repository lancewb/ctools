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
                prefix="Subnet:"
                style="max-width: 300px"
                @keyup.enter="startPing"
            ></v-text-field>

            <v-btn
                color="primary"
                prepend-icon="mdi-play"
                @click="startPing"
                :loading="loading"
                height="40"
            >
              开始扫描
            </v-btn>

            <v-spacer></v-spacer>

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
                  v-tooltip:top="`IP: ${currentSubnet}.${item.id}<br>延迟: ${item.latency >= 5000 ? '超时' : item.latency + 'ms'}`"
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
                :subtitle="`扫描 ${history}.1-254`"
                @click="reuseHistory(history)"
                link
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
import { ref, computed, onMounted } from 'vue'
// 引入后端方法 (根据你的 Wails 项目结构，路径可能略有不同)
import { PingSubnet, GetPingHistory, AddPingHistory, RemovePingHistory, ClearPingHistory } from '../../../wailsjs/go/network/NetworkService.js'

const subnetInput = ref('')
const currentSubnet = ref('')
const loading = ref(false)
const results = ref([]) // 存储后端返回的 {id, latency}
const historyList = ref([])

// 初始化 254 个格子，默认状态
// 为了让 UI 刚进去就有格子看（虽然是空的），或者选择 Loading 时显示骨架屏
// 这里我们选择：只有开始扫描后才渲染格子，或者保持 254 个灰色格子
const displayList = computed(() => {
  if (results.value.length > 0) return results.value
  // 默认生成 1-254 的空数据用于占位显示（可选）
  return Array.from({ length: 254 }, (_, i) => ({ id: i + 1, latency: -1 }))
})

// 获取颜色 Class
const getStatusClass = (latency) => {
  if (latency === -1) return 'bg-grey-lighten-3 text-grey' // 未扫描状态
  if (latency >= 5000) return 'bg-white border text-grey-lighten-1' // 超时 (白色，加边框防止看不见)
  if (latency >= 1000) return 'bg-red text-white' // > 1s
  if (latency >= 100) return 'bg-orange text-white' // 100ms - 1000ms
  return 'bg-green text-white' // < 100ms
}

// 加载历史记录
const loadHistory = async () => {
  try {
    historyList.value = await GetPingHistory()
  } catch (e) {
    console.error(e)
  }
}

// 开始 Ping
const startPing = async () => {
  const target = subnetInput.value.trim()
  if (!target) return

  // 简单的正则校验 (xxx.xxx.xxx)
  // 实际上可以更严谨，这里简单处理
  if (target.split('.').length !== 3) {
    // 这里最好加个 Toast 提示
    alert("请输入正确的三位子网号，如 192.168.1")
    return
  }

  loading.value = true
  currentSubnet.value = target
  // 重置结果为初始状态
  results.value = Array.from({ length: 254 }, (_, i) => ({ id: i + 1, latency: -1 }))

  try {
    // 1. 添加到历史
    historyList.value = await AddPingHistory(target)

    // 2. 调用后端 Ping
    // 注意：这里后端是同步返回所有结果。
    // 如果是真实场景，建议后端使用 Events.Emit 流式传输，这里按需求做的一次性返回。
    results.value = await PingSubnet(target)

  } catch (e) {
    console.error("Ping failed:", e)
  } finally {
    loading.value = false
  }
}

// 历史记录操作
const reuseHistory = (subnet) => {
  subnetInput.value = subnet
  startPing()
}

const deleteHistory = async (subnet) => {
  historyList.value = await RemovePingHistory(subnet)
}

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