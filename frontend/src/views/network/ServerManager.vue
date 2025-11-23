<template>
  <v-container fluid class="h-100 pa-4 bg-grey-lighten-5">

    <v-row class="mb-2" align="center">
      <v-col>
        <h2 class="text-h5 font-weight-bold text-primary">
          <v-icon icon="mdi-server-network" class="mr-2"></v-icon>
          服务器管理
        </h2>
        <div class="text-caption text-grey">管理并监控您的 Linux 服务器状态</div>
      </v-col>
      <v-col cols="auto">
        <v-btn color="primary" prepend-icon="mdi-plus" @click="openAddDialog">添加服务器</v-btn>
      </v-col>
    </v-row>

    <v-row>
      <v-col
          v-for="server in serverList"
          :key="server.id"
          cols="12" md="6" lg="4"
      >
        <v-card elevation="2" rounded="lg" class="h-100 d-flex flex-column">
          <v-card-item class="bg-blue-grey-lighten-5 py-3">
            <template v-slot:prepend>
              <v-avatar color="primary" variant="tonal" rounded>
                <v-icon icon="mdi-linux"></v-icon>
              </v-avatar>
            </template>
            <v-card-title class="text-subtitle-1 font-weight-bold">
              {{ server.name }}
            </v-card-title>
            <v-card-subtitle class="font-mono text-xs">
              {{ server.user }}@{{ server.host }}:{{ server.port }}
            </v-card-subtitle>

            <template v-slot:append>
              <v-menu>
                <template v-slot:activator="{ props }">
                  <v-btn icon="mdi-dots-vertical" variant="text" size="small" v-bind="props"></v-btn>
                </template>
                <v-list density="compact">
                  <v-list-item prepend-icon="mdi-pencil" title="编辑" @click="editServer(server)"></v-list-item>
                  <v-list-item prepend-icon="mdi-delete" title="删除" color="error" @click="deleteServer(server.id)"></v-list-item>
                </v-list>
              </v-menu>
            </template>
          </v-card-item>

          <v-divider></v-divider>

          <v-card-text class="flex-grow-1 pt-4">
            <div v-if="!serverStatus[server.id]" class="text-center py-6 text-grey">
              <v-icon size="48" icon="mdi-lan-disconnect" class="mb-2 opacity-50"></v-icon>
              <div>未连接</div>
              <v-btn
                  color="primary"
                  variant="tonal"
                  size="small"
                  class="mt-4"
                  :loading="loadingState[server.id]"
                  @click="refreshServer(server)"
              >
                连接并获取状态
              </v-btn>
            </div>

            <div v-else>
              <v-alert v-if="serverStatus[server.id].error" type="error" variant="tonal" density="compact" class="mb-2 text-caption">
                {{ serverStatus[server.id].error }}
                <template v-slot:append>
                  <v-btn icon="mdi-refresh" size="x-small" variant="text" @click="refreshServer(server)"></v-btn>
                </template>
              </v-alert>

              <div v-else>
                <div class="mb-3">
                  <div class="d-flex justify-space-between text-caption font-weight-bold mb-1">
                    <span>CPU </span>
                    <div class="text-caption text-grey text-truncate mb-1" :title="serverStatus[server.id].cpuModel">
                      {{ serverStatus[server.id].cpuModel || 'Unknown CPU' }}
                    </div>
                    <span>{{ serverStatus[server.id].cpuUsage }}</span>
                  </div>
                </div>

                <div class="mb-3">
                  <div class="d-flex justify-space-between text-caption font-weight-bold mb-1">
                    <span>内存 ({{ serverStatus[server.id].ramPercent.toFixed(1) }}%)</span>
                    <span>{{ serverStatus[server.id].ramUsed }} / {{ serverStatus[server.id].ramTotal }}</span>
                  </div>
                  <v-progress-linear
                      :model-value="serverStatus[server.id].ramPercent"
                      color="teal"
                      height="6"
                      rounded
                  ></v-progress-linear>
                </div>

                <div class="mb-3">
                  <div class="d-flex justify-space-between text-caption font-weight-bold mb-1">
                    <span>系统盘 ({{ serverStatus[server.id].diskPercent.toFixed(1) }}%)</span>
                    <span>{{ serverStatus[server.id].diskUsed }} / {{ serverStatus[server.id].diskSize }}</span>
                  </div>
                  <v-progress-linear
                      :model-value="serverStatus[server.id].diskPercent"
                      color="indigo"
                      height="6"
                      rounded
                  ></v-progress-linear>
                </div>

                <v-divider class="my-3"></v-divider>

                <div class="text-caption font-weight-bold mb-2">PCI / 网卡 / 密码设备</div>
                <div class="bg-grey-lighten-4 rounded pa-2 overflow-y-auto" style="max-height: 100px; font-size: 11px; line-height: 1.4;">
                  <div v-for="(dev, i) in serverStatus[server.id].pciDevices" :key="i" class="mb-1">
                    • {{ dev }}
                  </div>
                  <div v-if="!serverStatus[server.id].pciDevices?.length" class="text-grey">未检测到相关设备</div>
                </div>

                <div class="mt-4 text-right">
                  <v-btn
                      size="small"
                      variant="text"
                      prepend-icon="mdi-refresh"
                      :loading="loadingState[server.id]"
                      @click="refreshServer(server)"
                      color="primary"
                  >
                    刷新状态
                  </v-btn>
                </div>
              </div>
            </div>
          </v-card-text>
        </v-card>
      </v-col>
    </v-row>

    <v-dialog v-model="dialog" max-width="500">
      <v-card>
        <v-card-title>{{ editingId ? '编辑服务器' : '添加服务器' }}</v-card-title>
        <v-card-text>
          <v-form ref="formRef">
            <v-text-field v-model="formData.name" label="服务器名称 (备注)" variant="outlined" density="compact"></v-text-field>
            <v-row>
              <v-col cols="8">
                <v-text-field v-model="formData.host" label="主机 IP / 域名" variant="outlined" density="compact"></v-text-field>
              </v-col>
              <v-col cols="4">
                <v-text-field v-model="formData.port" label="端口" variant="outlined" density="compact" placeholder="22"></v-text-field>
              </v-col>
            </v-row>
            <v-text-field v-model="formData.user" label="用户名" variant="outlined" density="compact" placeholder="root"></v-text-field>

            <v-radio-group v-model="formData.authType" inline density="compact" hide-details class="mb-2">
              <v-radio label="密码认证" value="password"></v-radio>
              <v-radio label="密钥认证" value="key"></v-radio>
            </v-radio-group>

            <v-text-field
                v-if="formData.authType === 'password'"
                v-model="formData.password"
                label="密码"
                type="password"
                variant="outlined"
                density="compact"
            ></v-text-field>

            <v-file-input
                v-else
                v-model="keyFile"
                label="选择私钥文件"
                variant="outlined"
                density="compact"
                accept=".pem,.key,*"
                @update:modelValue="handleFileSelect"
            ></v-file-input>

            <div v-if="formData.authType === 'key' && formData.keyPath" class="text-caption text-grey">
              当前路径: {{ formData.keyPath }}
            </div>

          </v-form>
        </v-card-text>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn color="grey" variant="text" @click="dialog = false">取消</v-btn>
          <v-btn color="primary" @click="saveServer">保存</v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>

  </v-container>
</template>

<script setup>
/**
 * ServerManager Component
 *
 * Allows users to manage a list of Linux servers and monitor their status via SSH.
 * Displays CPU, Memory, Disk usage, and PCI devices (e.g., Crypto cards).
 */

import { ref, reactive, onMounted } from 'vue'
import { GetServerList, SaveServer, DeleteServer, CheckServerStatus } from '../../../wailsjs/go/network/NetworkService.js'

// --- State ---
const serverList = ref([])
const serverStatus = reactive({})
const loadingState = reactive({})

const dialog = ref(false)
const editingId = ref('')
const keyFile = ref(null)

// Reference to the form element
const formRef = ref(null)

const formData = reactive({
  id: '',
  name: '',
  host: '',
  port: '22',
  user: 'root',
  authType: 'password',
  password: '',
  keyPath: ''
})

// --- Methods ---

/**
 * loadList fetches the saved server configurations.
 */
const loadList = async () => {
  serverList.value = await GetServerList()
}

/**
 * openAddDialog resets the form and opens the dialog to add a new server.
 */
const openAddDialog = () => {
  editingId.value = ''
  formData.id = ''
  formData.name = ''
  formData.host = ''
  formData.port = '22'
  formData.user = 'root'
  formData.authType = 'password'
  formData.password = ''
  formData.keyPath = ''
  keyFile.value = null
  dialog.value = true
}

/**
 * editServer populates the form with existing server details for editing.
 *
 * @param {Object} server - The server object to edit.
 */
const editServer = (server) => {
  editingId.value = server.id
  Object.assign(formData, server)
  keyFile.value = null
  dialog.value = true
}

/**
 * handleFileSelect handles private key file selection.
 * Note: Web file input path access is restricted. In a real app, use Wails runtime dialog.
 */
const handleFileSelect = (files) => {
  // Placeholder: Real implementation requires Wails OpenFileDialog to get absolute path
}

/**
 * saveServer persists the server configuration to the backend.
 */
const saveServer = async () => {
  if (!formData.host || !formData.user) return

  const payload = JSON.parse(JSON.stringify(formData))
  serverList.value = await SaveServer(payload)
  dialog.value = false
}

/**
 * deleteServer removes a server configuration.
 *
 * @param {string} id - The ID of the server to delete.
 */
const deleteServer = async (id) => {
  if (confirm('确定移除该服务器配置吗？')) {
    serverList.value = await DeleteServer(id)
    delete serverStatus[id]
  }
}

// --- SSH Actions ---

/**
 * refreshServer initiates an SSH connection to fetch current server status.
 *
 * @param {Object} server - The server config object.
 */
const refreshServer = async (server) => {
  loadingState[server.id] = true
  try {
    const status = await CheckServerStatus(server)
    serverStatus[server.id] = status
  } catch (e) {
    serverStatus[server.id] = {
      isOnline: false,
      error: "调用失败: " + e
    }
  } finally {
    loadingState[server.id] = false
  }
}

onMounted(async () => {
  // 1. Load saved servers
  await loadList()

  // 2. Auto-refresh status for all servers
  if (serverList.value && serverList.value.length > 0) {
    serverList.value.forEach(server => {
      refreshServer(server)
    })
  }
})
</script>

<style scoped>
.font-mono {
  font-family: 'Consolas', monospace;
}
</style>