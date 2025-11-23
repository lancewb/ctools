<template>
  <v-container fluid class="h-100 pa-0 bg-grey-lighten-5">
    <v-row no-gutters class="h-100">

      <v-col cols="12" md="3" class="border-e bg-white d-flex flex-column h-100">
        <v-toolbar density="compact" color="transparent" class="border-b px-2">
          <v-toolbar-title class="text-subtitle-2 font-weight-bold text-grey-darken-2">请求合集</v-toolbar-title>
          <v-spacer></v-spacer>
          <v-btn icon="mdi-plus" size="small" color="primary" variant="tonal" @click="createNewRequest" title="新建"></v-btn>
        </v-toolbar>

        <v-list density="compact" class="flex-grow-1 overflow-y-auto py-0" nav>
          <v-list-item
              v-for="item in collections"
              :key="item.id"
              :value="item"
              @click="loadRequest(item)"
              :active="currentReq.id === item.id"
              color="primary"
              rounded="lg"
              class="mb-1 mx-2"
          >
            <template v-slot:prepend>
              <v-chip size="x-small" :color="getMethodColor(item.request.method, item.request.protocol)" label class="font-weight-bold mr-2" style="width: 45px; justify-content: center;">
                {{ item.request.protocol === 'ws' ? 'WS' : item.request.method }}
              </v-chip>
            </template>
            <v-list-item-title class="text-caption font-weight-medium">{{ item.name }}</v-list-item-title>
            <template v-slot:append>
              <v-btn icon="mdi-delete-outline" size="x-small" variant="text" color="grey-lighten-1" @click.stop="deleteCollection(item.id)"></v-btn>
            </template>
          </v-list-item>
        </v-list>
      </v-col>

      <v-col cols="12" md="9" class="d-flex flex-column h-100">

        <div class="pa-3 bg-white border-b">
          <v-row dense align="center">
            <v-col cols="auto" style="width: 100px;">
              <v-select
                  v-model="currentReq.protocol"
                  :items="['http', 'ws']"
                  density="compact"
                  variant="outlined"
                  hide-details
                  label="协议"
                  bg-color="grey-lighten-5"
                  @update:model-value="onProtocolChange"
              ></v-select>
            </v-col>

            <v-col cols="auto" style="width: 110px;" v-if="currentReq.protocol === 'http'">
              <v-select
                  v-model="currentReq.method"
                  :items="['GET', 'POST', 'PUT', 'DELETE', 'PATCH']"
                  density="compact"
                  variant="outlined"
                  hide-details
                  bg-color="grey-lighten-5"
              ></v-select>
            </v-col>

            <v-col>
              <v-text-field
                  v-model="currentReq.url"
                  :placeholder="currentReq.protocol === 'http' ? 'https://api.example.com/v1/resource' : 'ws://echo.websocket.org'"
                  density="compact"
                  variant="outlined"
                  hide-details
                  bg-color="grey-lighten-5"
                  clearable
              ></v-text-field>
            </v-col>

            <v-col cols="auto" class="d-flex align-center pl-2">
              <v-btn
                  color="primary"
                  height="40"
                  class="px-6 font-weight-bold"
                  @click="handleSend"
                  :loading="loading"
                  elevation="1"
                  :prepend-icon="currentReq.protocol === 'ws' ? (wsConnected ? 'mdi-power-plug-off' : 'mdi-power-plug') : 'mdi-send'"
              >
                {{ sendButtonText }}
              </v-btn>

              <v-btn
                  color="blue-grey"
                  height="40"
                  variant="tonal"
                  class="ml-2"
                  @click="saveDialog = true"
                  icon="mdi-content-save-outline"
                  title="保存"
              ></v-btn>
            </v-col>
          </v-row>
        </div>

        <div class="flex-grow-1 d-flex flex-column overflow-hidden">
          <v-tabs v-model="activeTab" density="compact" color="primary" bg-color="grey-lighten-4" class="border-b">
            <v-tab value="params" class="text-caption text-capitalize">Params</v-tab>
            <v-tab value="headers" class="text-caption text-capitalize">Headers</v-tab>
            <v-tab value="body" v-if="currentReq.protocol === 'http'" class="text-caption text-capitalize">
              Body <span class="text-grey ml-1 text-xs">({{ bodyType }})</span>
            </v-tab>
            <v-tab value="settings" class="text-caption text-capitalize">Settings</v-tab>
          </v-tabs>

          <v-window v-model="activeTab" class="flex-grow-1 bg-white overflow-y-auto" :transition="false" :reverse-transition="false">

            <v-window-item value="params" class="h-100">
              <div class="pa-4 h-100">
                <key-value-editor v-model="paramsList"></key-value-editor>
              </div>
            </v-window-item>

            <v-window-item value="headers" class="h-100">
              <div class="pa-4 h-100">
                <key-value-editor v-model="headersList" :key-options="commonHeaderKeys"></key-value-editor>
              </div>
            </v-window-item>

            <v-window-item value="body" class="h-100">
              <div class="h-100 d-flex flex-column">
                <div class="d-flex align-center px-4 py-2 bg-grey-lighten-5 border-b" style="min-height: 45px;">
                  <v-btn-toggle v-model="bodyType" density="compact" color="primary" mandatory variant="outlined" class="mr-4" style="height: 28px;">
                    <v-btn value="json-kv" size="small" class="text-caption">JSON (KV)</v-btn>
                    <v-btn value="raw" size="small" class="text-caption">Raw</v-btn>
                  </v-btn-toggle>
                  <div class="text-caption text-grey" v-if="bodyType === 'json-kv'">自动转换为 JSON 发送</div>
                </div>

                <div v-if="bodyType === 'json-kv'" class="flex-grow-1 overflow-y-auto pa-4">
                  <key-value-editor v-model="bodyKvList"></key-value-editor>
                </div>

                <div v-else class="flex-grow-1 pa-0">
                  <v-textarea
                      v-model="currentReq.body"
                      placeholder='{ "key": "value" }'
                      variant="plain"
                      class="h-100 full-height-textarea pa-4 font-mono"
                      no-resize
                      hide-details
                  ></v-textarea>
                </div>
              </div>
            </v-window-item>

            <v-window-item value="settings" class="h-100">
              <div class="pa-4 h-100">
                <v-row>
                  <v-col cols="12" md="4">
                    <v-select
                        v-model="currentReq.tlsVersion"
                        :items="tlsOptions"
                        label="TLS / TLCP 版本"
                        variant="outlined"
                        density="compact"
                    ></v-select>
                  </v-col>
                  <v-col cols="12" md="4">
                    <v-text-field
                        v-model.number="currentReq.timeout"
                        label="超时时间 (秒)"
                        type="number"
                        variant="outlined"
                        density="compact"
                    ></v-text-field>
                  </v-col>
                </v-row>
              </div>
            </v-window-item>
          </v-window>
        </div>

        <v-divider thickness="2"></v-divider>
        <div class="d-flex flex-column" style="height: 45%;">

          <div class="d-flex align-center px-4 py-1 bg-grey-lighten-4 border-b" style="min-height: 36px;">
             <span class="text-subtitle-2 font-weight-bold mr-4 text-grey-darken-3">
               {{ currentReq.protocol === 'ws' ? 'WebSocket 消息流' : 'Response' }}
             </span>
            <template v-if="currentReq.protocol === 'http' && response">
              <v-chip size="x-small" :color="getStatusColor(response.statusCode)" class="mr-2 font-weight-bold">
                {{ response.statusCode }}
              </v-chip>
              <span class="text-caption text-grey mr-4">Time: {{ response.timeCost }}ms</span>
              <v-spacer></v-spacer>
              <v-btn-toggle v-model="responseView" density="compact" variant="text" color="primary" mandatory class="tiny-toggle">
                <v-btn value="json" size="small" class="text-caption">JSON</v-btn>
                <v-btn value="raw" size="small" class="text-caption">Raw</v-btn>
              </v-btn-toggle>
            </template>
            <template v-if="currentReq.protocol === 'ws'">
              <v-chip size="x-small" :color="wsConnected ? 'success' : 'grey'" variant="flat" class="mr-2">
                {{ wsConnected ? 'Connected' : 'Disconnected' }}
              </v-chip>
              <v-spacer></v-spacer>
              <v-btn size="x-small" variant="text" icon="mdi-delete" @click="wsLogs = []" title="清空日志"></v-btn>
            </template>
          </div>

          <div v-if="currentReq.protocol === 'http'" class="flex-grow-1 overflow-y-auto bg-white pa-0 position-relative">
            <div v-if="loading" class="d-flex align-center justify-center h-100">
              <v-progress-circular indeterminate color="primary" size="32"></v-progress-circular>
            </div>

            <div v-else-if="response" class="h-100 pa-4 text-left d-flex flex-column align-start">
              <pre v-if="responseView === 'json'" class="json-view w-100">{{ tryPrettyJson(response.body) }}</pre>
              <pre v-else class="raw-view w-100">{{ response.body }}</pre>
            </div>

            <div v-else-if="respError" class="text-error pa-4 text-left">
              <v-icon icon="mdi-alert-circle" color="error" class="mr-2"></v-icon>
              {{ respError }}
            </div>

            <div v-else class="d-flex flex-column align-center justify-center h-100 text-grey-lighten-1">
              <v-icon size="48" class="mb-2">mdi-send-outline</v-icon>
              <div class="text-caption">输入 URL 并发送请求</div>
            </div>
          </div>

          <div v-else class="flex-grow-1 d-flex flex-row bg-white overflow-hidden">
            <div class="flex-grow-1 border-e d-flex flex-column" style="width: 60%;">
              <div class="flex-grow-1 overflow-y-auto pa-2 bg-grey-lighten-5" ref="wsLogRef" style="font-family: monospace; font-size: 12px;">
                <div v-for="(msg, i) in wsLogs" :key="i" class="mb-1 py-1 px-2 rounded bg-white elevation-1 border">
                  <div class="d-flex align-center mb-1 text-caption">
                             <span :class="msg.type === 'send' ? 'text-blue-darken-2' : (msg.type === 'error' ? 'text-red' : 'text-green-darken-2')" class="font-weight-bold mr-2">
                               {{ msg.type === 'send' ? '⬆️ SENT' : (msg.type === 'error' ? '❌ ERR' : '⬇️ RECV') }}
                             </span>
                    <span class="text-grey">{{ msg.time }}</span>
                  </div>
                  <div style="word-break: break-all; white-space: pre-wrap;">{{ msg.content }}</div>
                </div>
              </div>
            </div>

            <div class="d-flex flex-column" style="width: 40%;">
              <div class="d-flex align-center px-2 py-1 border-b bg-grey-lighten-4">
                <span class="text-caption font-weight-bold">发送消息</span>
                <v-spacer></v-spacer>
                <v-btn-toggle v-model="wsMsgType" density="compact" color="primary" mandatory variant="text" class="tiny-toggle">
                  <v-btn value="json-kv" size="x-small">KV</v-btn>
                  <v-btn value="raw" size="x-small">Raw</v-btn>
                </v-btn-toggle>
              </div>

              <div class="flex-grow-1 position-relative">
                <div v-if="wsMsgType === 'json-kv'" class="h-100 overflow-y-auto pa-2">
                  <key-value-editor v-model="wsKvList"></key-value-editor>
                </div>
                <div v-else class="h-100">
                  <v-textarea
                      v-model="wsRawMsg"
                      placeholder="Message body..."
                      variant="plain"
                      class="h-100 full-height-textarea pa-2 font-mono text-body-2"
                      no-resize
                      hide-details
                  ></v-textarea>
                </div>
              </div>

              <div class="pa-2 border-t bg-grey-lighten-5 text-right">
                <v-btn
                    color="primary"
                    size="small"
                    prepend-icon="mdi-send"
                    @click="sendWsMessage"
                    :disabled="!wsConnected"
                    block
                >发送消息</v-btn>
              </div>
            </div>
          </div>

        </div>

      </v-col>
    </v-row>

    <v-dialog v-model="saveDialog" max-width="400">
      <v-card>
        <v-card-title class="text-subtitle-1 font-weight-bold">保存请求到合集</v-card-title>
        <v-card-text>
          <v-text-field v-model="saveName" label="名称" variant="outlined" density="compact" autofocus></v-text-field>
        </v-card-text>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn color="grey" variant="text" @click="saveDialog = false">取消</v-btn>
          <v-btn color="primary" @click="saveCollection">确认保存</v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>

  </v-container>
</template>

<script setup>
// ... script 部分与之前基本一致，只补充 commonHeaderKeys ...

import { ref, reactive, computed, onMounted, watch, nextTick } from 'vue'
import KeyValueEditor from '../../components/KeyValueEditor.vue'
import { SendHttpRequest, GetReqCollections, SaveReqCollection, DeleteReqCollection } from '../../../wailsjs/go/network/NetworkService.js'

// 定义常用 Header，传给 KeyValueEditor
const commonHeaderKeys = [
  'Authorization',
  'Content-Type',
  'Accept',
  'User-Agent',
  'Cache-Control',
  'Host',
  'Connection',
  'Origin',
  'Referer'
]

// ... 其余 Script 逻辑 (watchers, methods) 与上一次回答一致，无需改动 ...
// ... 请确保保留上个回答中关于 bodyKvList, wsMsgType 等的 watch 逻辑 ...
// ... 这里为了节省篇幅不再重复粘贴全部 script，如果您需要请告诉我 ...

const tlsOptions = [
  { title: 'Auto', value: '' },
  { title: 'TLS 1.1', value: '1.1' },
  { title: 'TLS 1.2', value: '1.2' },
  { title: 'TLS 1.3', value: '1.3' },
  { title: '国密 TLCP', value: 'tlcp' },
]

const collections = ref([])
const loading = ref(false)
const saveDialog = ref(false)
const saveName = ref('')
const activeTab = ref('params')
const responseView = ref('json')
const respError = ref('')

const currentReq = reactive({
  id: '',
  protocol: 'http',
  method: 'GET',
  url: '',
  headers: {},
  body: '',
  tlsVersion: '',
  timeout: 10
})

const paramsList = ref([{ key: '', value: '', active: true }])
const headersList = ref([{ key: '', value: '', active: true }])

const bodyType = ref('json-kv')
const bodyKvList = ref([{ key: '', value: '', active: true }])

const wsConnected = ref(false)
const wsSocket = ref(null)
const wsLogs = ref([])
const wsLogRef = ref(null)
const wsMsgType = ref('raw')
const wsKvList = ref([{ key: '', value: '', active: true }])
const wsRawMsg = ref('')
const response = ref(null)

// Watchers
watch(bodyType, (newType) => {
  if (newType === 'raw') {
    const obj = kvListToObj(bodyKvList.value)
    if (Object.keys(obj).length > 0 || (bodyKvList.value.length === 1 && !bodyKvList.value[0].key)) {
      currentReq.body = JSON.stringify(obj, null, 2)
    }
  } else {
    if (currentReq.body) {
      try {
        const obj = JSON.parse(currentReq.body)
        bodyKvList.value = objToKvList(obj)
      } catch (e) {}
    }
  }
})

watch(wsMsgType, (newType) => {
  if (newType === 'raw') {
    const obj = kvListToObj(wsKvList.value)
    if (Object.keys(obj).length > 0) {
      wsRawMsg.value = JSON.stringify(obj, null, 2)
    }
  } else {
    if (wsRawMsg.value) {
      try {
        const obj = JSON.parse(wsRawMsg.value)
        wsKvList.value = objToKvList(obj)
      } catch(e) {}
    }
  }
})

const sendButtonText = computed(() => {
  if (currentReq.protocol === 'ws') {
    return wsConnected.value ? '断开' : '连接'
  }
  return '发送'
})

const loadCollections = async () => {
  collections.value = await GetReqCollections()
}

const createNewRequest = () => {
  currentReq.id = ''
  currentReq.method = 'GET'
  currentReq.url = ''
  currentReq.body = ''
  currentReq.headers = {}
  paramsList.value = [{ key: '', value: '', active: true }]
  headersList.value = [{ key: '', value: '', active: true }]
  bodyKvList.value = [{ key: '', value: '', active: true }]
  bodyType.value = 'json-kv'
  response.value = null
  respError.value = ''
}

const loadRequest = (item) => {
  const r = item.request
  currentReq.id = item.id
  currentReq.protocol = r.protocol || 'http'
  currentReq.method = r.method
  currentReq.url = r.url
  currentReq.body = r.body
  currentReq.tlsVersion = r.tlsVersion
  currentReq.timeout = r.timeout

  headersList.value = objToKvList(r.headers)

  if (r.url.includes('?')) {
    const qs = r.url.split('?')[1]
    paramsList.value = []
    const params = new URLSearchParams(qs)
    params.forEach((value, key) => {
      paramsList.value.push({ key, value, active: true })
    })
    paramsList.value.push({ key: '', value: '', active: true })
  } else {
    paramsList.value = [{ key: '', value: '', active: true }]
  }

  try {
    if (r.body && r.body.trim().startsWith('{')) {
      bodyKvList.value = objToKvList(JSON.parse(r.body))
      bodyType.value = 'json-kv'
    } else {
      bodyType.value = 'raw'
    }
  } catch {
    bodyType.value = 'raw'
  }

  saveName.value = item.name
}

const saveCollection = async () => {
  const fullUrl = buildFullUrl()
  const finalHeaders = kvListToObj(headersList.value)
  prepareBodyForSend()

  const item = {
    id: currentReq.id,
    name: saveName.value,
    request: {
      ...currentReq,
      url: fullUrl,
      headers: finalHeaders
    }
  }

  collections.value = await SaveReqCollection(item)
  saveDialog.value = false
  const saved = collections.value.find(c => c.name === saveName.value)
  if (saved) currentReq.id = saved.id
}

const deleteCollection = async (id) => {
  if(confirm("删除此记录？")) {
    collections.value = await DeleteReqCollection(id)
    if (currentReq.id === id) createNewRequest()
  }
}

const prepareBodyForSend = () => {
  if (bodyType.value === 'json-kv') {
    const obj = kvListToObj(bodyKvList.value)
    if (Object.keys(obj).length > 0) {
      currentReq.body = JSON.stringify(obj)
    } else {
      currentReq.body = ''
    }
  }
}

const handleSend = async () => {
  if (currentReq.protocol === 'ws') {
    toggleWsConnection()
    return
  }

  loading.value = true
  response.value = null
  respError.value = ''

  const fullUrl = buildFullUrl()
  const finalHeaders = kvListToObj(headersList.value)
  prepareBodyForSend()

  const opt = {
    id: currentReq.id,
    method: currentReq.method,
    url: fullUrl,
    headers: finalHeaders,
    body: currentReq.body,
    protocol: 'http',
    tlsVersion: currentReq.tlsVersion,
    timeout: currentReq.timeout
  }

  try {
    const res = await SendHttpRequest(opt)
    if (res.error) {
      respError.value = res.error
    } else {
      response.value = res
    }
  } catch (e) {
    respError.value = "系统错误: " + e
  } finally {
    loading.value = false
  }
}

const toggleWsConnection = () => {
  if (wsConnected.value) {
    wsSocket.value.close()
    wsConnected.value = false
    addWsLog('info', '连接已断开')
  } else {
    if (!currentReq.url) return
    try {
      wsSocket.value = new WebSocket(currentReq.url)
      wsSocket.value.onopen = () => {
        wsConnected.value = true
        addWsLog('info', '连接成功')
      }
      wsSocket.value.onmessage = (event) => {
        addWsLog('recv', event.data)
      }
      wsSocket.value.onerror = () => {
        addWsLog('error', 'WebSocket 错误')
      }
      wsSocket.value.onclose = () => {
        wsConnected.value = false
        addWsLog('info', '连接关闭')
      }
    } catch (e) {
      addWsLog('error', '创建连接失败: ' + e.message)
    }
  }
}

const sendWsMessage = () => {
  if (!wsSocket.value || !wsConnected.value) return
  let msgToSend = ''
  if (wsMsgType.value === 'json-kv') {
    const obj = kvListToObj(wsKvList.value)
    msgToSend = JSON.stringify(obj)
  } else {
    msgToSend = wsRawMsg.value
  }

  if (msgToSend) {
    wsSocket.value.send(msgToSend)
    addWsLog('send', msgToSend)
  }
}

const addWsLog = (type, content) => {
  const time = new Date().toLocaleTimeString()
  wsLogs.value.push({type, content, time})
  nextTick(() => {
    if (wsLogRef.value) wsLogRef.value.scrollTop = wsLogRef.value.scrollHeight
  })
}

const buildFullUrl = () => {
  if (!currentReq.url) return ''
  let base = currentReq.url.split('?')[0]
  const qs = paramsList.value
      .filter(p => p.active && p.key)
      .map(p => `${encodeURIComponent(p.key)}=${encodeURIComponent(p.value)}`)
      .join('&')
  return qs ? `${base}?${qs}` : base
}

const kvListToObj = (list) => {
  const obj = {}
  list.forEach(item => {
    if (item.active && item.key) {
      obj[item.key] = item.value
    }
  })
  return obj
}

const objToKvList = (obj) => {
  const list = []
  if (obj) {
    for (const k in obj) {
      list.push({key: k, value: obj[k], active: true})
    }
  }
  list.push({key: '', value: '', active: true})
  return list
}

const tryPrettyJson = (str) => {
  try {
    return JSON.stringify(JSON.parse(str), null, 2)
  } catch {
    return str
  }
}

const getStatusColor = (code) => {
  if (code >= 200 && code < 300) return 'success'
  if (code >= 300 && code < 400) return 'warning'
  if (code >= 400) return 'error'
  return 'grey'
}

const getMethodColor = (m, p) => {
  if (p === 'ws') return 'blue-grey'
  if (m === 'GET') return 'green'
  if (m === 'POST') return 'orange'
  if (m === 'DELETE') return 'red'
  if (m === 'PUT') return 'blue'
  return 'grey'
}

const onProtocolChange = () => {
  response.value = null
  wsLogs.value = []
}

onMounted(() => {
  loadCollections()
})
</script>

<style scoped>
.full-height-textarea :deep(.v-field__input) {
  height: 100% !important;
}

.json-view, .raw-view {
  font-family: 'Consolas', monospace;
  font-size: 12px;
  white-space: pre-wrap;
  word-wrap: break-word;
}

.font-mono {
  font-family: 'Consolas', monospace !important;
}

.tiny-toggle .v-btn {
  text-transform: none;
  letter-spacing: 0;
}
</style>