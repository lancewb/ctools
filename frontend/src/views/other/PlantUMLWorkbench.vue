<template>
  <v-container fluid class="plantuml-workbench">
    <v-row>
      <v-col
        cols="12"
        md="5"
        v-show="!previewExpanded"
      >
        <v-card>
          <v-card-title class="text-subtitle-1 font-weight-bold">
            PlantUML 编辑
          </v-card-title>
          <v-card-text>
            <v-textarea
              v-model="plantumlSource"
              rows="14"
              auto-grow
              density="comfortable"
              spellcheck="false"
              label="PlantUML 源码"
              class="mb-3"
            />
            <div class="text-body-2 font-weight-medium mb-1">
              渲染选项
            </div>
            <v-row dense>
              <v-col cols="12" md="6">
                <v-select
                  v-model="outputFormat"
                  :items="outputFormatOptions"
                  label="输出格式"
                  density="comfortable"
                />
              </v-col>
              <v-col cols="12" md="6">
                <v-text-field
                  v-model.number="timeoutSeconds"
                  type="number"
                  min="3"
                  max="60"
                  label="超时 (秒)"
                  density="comfortable"
                />
              </v-col>
              <v-col cols="12" md="6">
                <v-select
                  v-model="serverPreset"
                  :items="serverOptions"
                  label="渲染服务器"
                  density="comfortable"
                />
              </v-col>
              <v-col
                v-if="serverPreset === 'custom'"
                cols="12"
              >
                <v-text-field
                  v-model="customServer"
                  label="自定义服务器地址"
                  hint="例如 http://intranet-host:8080/plantuml"
                  persistent-hint
                  density="comfortable"
                />
              </v-col>
            </v-row>
            <v-alert
              type="info"
              variant="tonal"
              class="mt-2"
              density="compact"
            >
              <template v-if="usingBuiltin">
                内置引擎会调用随应用分发的 PlantUML 核心，需确保本机已安装 Java 11+。
                <span v-if="javaReady && javaVersion" class="d-block text-success mt-1">
                  已检测到：{{ javaVersion }}
                </span>
              </template>
              <template v-else>
                建议在内网运行 <code>docker run -d -p 18080:8080 plantuml/plantuml-server</code>，
                然后指向 <code>http://127.0.0.1:18080/plantuml</code>。
              </template>
            </v-alert>
            <div class="text-body-2 font-weight-medium mb-1 mt-4">
              快速模板
            </div>
            <v-chip-group
              class="mb-2"
              selected-class="bg-primary text-white"
              column
            >
              <v-chip
                v-for="tpl in templates"
                :key="tpl.label"
                label
                density="comfortable"
                @click="applyTemplate(tpl.code)"
              >
                {{ tpl.label }}
              </v-chip>
            </v-chip-group>
            <div class="text-caption text-medium-emphasis">
              PlantUML 语法说明：
              <a
                href="https://plantuml.com/zh/"
                target="_blank"
                rel="noreferrer"
              >
                官方文档
              </a>
            </div>
          </v-card-text>
          <v-card-actions>
            <v-btn
              color="primary"
              prepend-icon="mdi-eye"
              :loading="rendering"
              @click="renderDiagram"
            >
              立即渲染
            </v-btn>
            <v-btn
              class="ml-2"
              prepend-icon="mdi-content-save"
              variant="text"
              color="secondary"
              :disabled="!plantumlSource"
              @click="saveLocal"
            >
              保存草稿
            </v-btn>
            <v-spacer />
            <v-btn
              icon="mdi-refresh"
              variant="text"
              :disabled="rendering"
              @click="reset"
            />
          </v-card-actions>
        </v-card>
      </v-col>
      <v-col
        cols="12"
        :md="previewExpanded ? 12 : 7"
      >
        <v-card class="plantuml-preview-card">
          <v-card-title class="text-subtitle-1 font-weight-bold d-flex align-center flex-wrap">
            渲染结果
            <v-chip
              v-if="lastRenderTime"
              class="ml-3"
              size="small"
              color="primary"
              variant="tonal"
            >
              {{ lastRenderTime }}
            </v-chip>
            <v-spacer />
            <div class="preview-action-group d-flex align-center flex-wrap">
              <v-btn
                size="small"
                variant="text"
                prepend-icon="mdi-download"
                :disabled="!hasRenderableResult"
                @click="exportResult"
              >
                下载结果
              </v-btn>
              <v-btn
                size="small"
                class="ml-1"
                variant="text"
                prepend-icon="mdi-content-copy"
                :disabled="!hasTextualResult"
                @click="copyResult"
              >
                复制文本
              </v-btn>
              <div class="preview-zoom-controls d-flex align-center ml-3">
                <v-btn
                  icon="mdi-magnify-minus-outline"
                  size="small"
                  variant="text"
                  :disabled="previewScale <= minScale"
                  @click="zoomOut"
                />
                <div class="preview-scale-display text-caption font-weight-medium mx-2">
                  {{ previewScaleDisplay }}%
                </div>
                <v-btn
                  icon="mdi-magnify-plus-outline"
                  size="small"
                  variant="text"
                  :disabled="previewScale >= maxScale"
                  @click="zoomIn"
                />
                <v-btn
                  icon="mdi-fit-to-page-outline"
                  size="small"
                  variant="text"
                  class="ml-1"
                  @click="resetZoom"
                />
                <v-btn
                  :icon="previewExpanded ? 'mdi-fullscreen-exit' : 'mdi-fullscreen'"
                  size="small"
                  variant="text"
                  class="ml-1"
                  @click="togglePreviewExpand"
                />
              </div>
            </div>
          </v-card-title>
          <v-card-text class="preview-wrapper">
            <div
              v-if="renderError"
              class="text-error font-mono text-body-2"
            >
              {{ renderError }}
            </div>
            <div
              v-else-if="outputFormat === 'svg' && renderedSvg"
              class="preview-surface"
            >
              <div
                class="preview-canvas"
                :style="previewCanvasStyle"
                v-html="renderedSvg"
              />
            </div>
            <div
              v-else-if="outputFormat === 'png' && renderedImageUrl"
              class="preview-surface"
            >
              <div
                class="preview-canvas"
                :style="previewCanvasStyle"
              >
                <img
                  :src="renderedImageUrl"
                  alt="PlantUML PNG 预览"
                  class="preview-image"
                >
              </div>
            </div>
            <div
              v-else-if="outputFormat === 'txt' && renderedText"
              class="preview-text"
            >
              <pre>{{ renderedText }}</pre>
            </div>
            <div v-else class="text-medium-emphasis text-body-2">
              输入 PlantUML 内容以开始渲染。
            </div>
          </v-card-text>
        </v-card>
      </v-col>
    </v-row>
  </v-container>
</template>

<script setup>
import { ref, computed, watch, onMounted, onUnmounted } from 'vue'
import { RenderPlantUML, CheckJavaRuntime } from '../../../wailsjs/go/other/OtherService'

const defaultDiagram = `@startuml
title 系统时序示例
participant Client
participant API
participant DB

Client -> API: POST /login
API -> DB: validate(credentials)
DB --> API: auth result
API --> Client: token / error
@enduml`

const diagramSourceKey = 'plantuml-last'
const settingsKey = 'plantuml-settings'
const minScale = 0.5
const maxScale = 3

const plantumlSource = ref(defaultDiagram)
const outputFormat = ref('svg')
const timeoutSeconds = ref(10)
const serverPreset = ref('local')
const customServer = ref('')
const rendering = ref(false)
const renderError = ref('')
const lastRenderTime = ref('')
const previewScale = ref(1)
const previewExpanded = ref(false)
const renderedSvg = ref('')
const renderedImageUrl = ref('')
const renderedText = ref('')
const lastResponse = ref(null)
const previewCanvasStyle = computed(() => ({
  transform: `scale(${previewScale.value})`,
  transformOrigin: 'top left',
  width: `${(1 / previewScale.value) * 100}%`
}))
const previewScaleDisplay = computed(() => Math.round(previewScale.value * 100))
const hasRenderableResult = computed(() => !!(renderedSvg.value || renderedImageUrl.value || renderedText.value))
const hasTextualResult = computed(() => !!(renderedSvg.value || renderedText.value))

const serverOptions = [
  { title: '内置引擎（需 Java）', value: 'builtin' },
  { title: '本地 Docker (127.0.0.1:18080)', value: 'local' },
  { title: '官方 SaaS (需外网)', value: 'official' },
  { title: '自定义', value: 'custom' }
]

const outputFormatOptions = [
  { title: 'SVG', value: 'svg' },
  { title: 'PNG', value: 'png' },
  { title: '文本 (TXT)', value: 'txt' }
]

const templates = [
  {
    label: '类图',
    code: `@startuml
class User {
  +ID: string
  +Name: string
  +Verify()
}
class Session {
  +Token: string
  +ExpiredAt: time
}
User "1" -- "many" Session
@enduml`
  },
  {
    label: '状态图',
    code: `@startuml
[*] --> Draft
Draft --> Review : 提交
Review --> Approved : 通过
Review --> Rejected : 驳回
Rejected --> Draft : 修改
Approved --> [*]
@enduml`
  },
  {
    label: '部署图',
    code: `@startuml
node "Browser" {
}
node "Gateway" {
}
node "K8S Cluster" {
  node "API Pod" {
    component "API"
  }
  database "PostgreSQL"
}
Browser --> Gateway
Gateway --> API
API --> PostgreSQL
@enduml`
  }
]

const presetServers = {
  local: 'http://127.0.0.1:18080/plantuml',
  official: 'https://www.plantuml.com/plantuml'
}

const javaReady = ref(false)
const javaVersion = ref('')
const usingBuiltin = computed(() => serverPreset.value === 'builtin')

const resolvedServer = computed(() => {
  if (usingBuiltin.value) {
    return ''
  }
  if (serverPreset.value === 'custom') {
    return customServer.value?.trim() || ''
  }
  return presetServers[serverPreset.value] || presetServers.local
})

let revertingPreset = false

const applyTemplate = (code) => {
  plantumlSource.value = code
}

const reset = () => {
  plantumlSource.value = defaultDiagram
  renderDiagram()
}

const saveLocal = () => {
  try {
    localStorage.setItem(diagramSourceKey, plantumlSource.value)
  } catch (err) {
    console.warn('无法保存 PlantUML 草稿', err)
  }
}

const loadLocal = () => {
  try {
    const cached = localStorage.getItem(diagramSourceKey)
    if (cached) {
      plantumlSource.value = cached
    }
    const raw = localStorage.getItem(settingsKey)
    if (raw) {
      const parsed = JSON.parse(raw)
      if (typeof parsed.outputFormat === 'string') {
        outputFormat.value = parsed.outputFormat
      }
      if (typeof parsed.serverPreset === 'string') {
        serverPreset.value = parsed.serverPreset
      }
      if (typeof parsed.customServer === 'string') {
        customServer.value = parsed.customServer
      }
      if (typeof parsed.timeoutSeconds === 'number' && !Number.isNaN(parsed.timeoutSeconds)) {
        timeoutSeconds.value = parsed.timeoutSeconds
      }
      if (typeof parsed.previewScale === 'number') {
        previewScale.value = clampScale(parsed.previewScale)
      }
      if (typeof parsed.previewExpanded === 'boolean') {
        previewExpanded.value = parsed.previewExpanded
      }
    }
  } catch (err) {
    console.warn('无法读取 PlantUML 草稿', err)
  }
}

const persistSettings = () => {
  try {
    const payload = {
      outputFormat: outputFormat.value,
      serverPreset: serverPreset.value,
      customServer: customServer.value,
      timeoutSeconds: timeoutSeconds.value,
      previewScale: previewScale.value,
      previewExpanded: previewExpanded.value
    }
    localStorage.setItem(settingsKey, JSON.stringify(payload))
  } catch (err) {
    console.warn('无法保存 PlantUML 设置', err)
  }
}

const clampScale = (value) => {
  if (Number.isNaN(Number(value))) {
    return 1
  }
  return Math.min(maxScale, Math.max(minScale, Number(value)))
}

const ensureJavaRuntime = async (fallbackPreset) => {
  try {
    const info = await CheckJavaRuntime()
    javaReady.value = true
    javaVersion.value = info
  } catch (err) {
    javaReady.value = false
    javaVersion.value = ''
    window.alert(
      `${err?.message || '未检测到 Java 运行环境，请访问 https://adoptium.net/ 下载并安装 Java 11+ 后重试。'}`
    )
    if (typeof fallbackPreset === 'string') {
      revertingPreset = true
      serverPreset.value = fallbackPreset || 'local'
      revertingPreset = false
    }
  }
}

const resetPreview = () => {
  renderedSvg.value = ''
  renderedText.value = ''
  if (renderedImageUrl.value) {
    URL.revokeObjectURL(renderedImageUrl.value)
    renderedImageUrl.value = ''
  }
}

const renderDiagram = async () => {
  if (!plantumlSource.value?.trim()) {
    resetPreview()
    renderError.value = ''
    return
  }
  rendering.value = true
  renderError.value = ''
  if (usingBuiltin.value && !javaReady.value) {
    await ensureJavaRuntime(null)
    if (!javaReady.value) {
      renderError.value = '未检测到 Java 运行环境'
      rendering.value = false
      return
    }
  }
  const server = usingBuiltin.value ? '' : (resolvedServer.value || presetServers.local)
  try {
    const result = await RenderPlantUML({
      source: plantumlSource.value,
      format: outputFormat.value,
      serverUrl: server,
      timeoutSeconds: Number(timeoutSeconds.value) || 10,
      useBuiltin: usingBuiltin.value
    })
    lastResponse.value = result
    resetPreview()
    if (outputFormat.value === 'svg') {
      renderedSvg.value = decodeBase64ToText(result.data)
    } else if (outputFormat.value === 'png') {
      renderedImageUrl.value = base64ToObjectURL(result.data, result.mimeType || 'image/png')
    } else {
      renderedText.value = decodeBase64ToText(result.data)
    }
    lastRenderTime.value = new Date().toLocaleTimeString()
  } catch (err) {
    renderError.value = err?.message ?? String(err)
  } finally {
    rendering.value = false
  }
}

const decodeBase64ToUint8 = (data) => {
  const binary = atob(data)
  const bytes = new Uint8Array(binary.length)
  for (let i = 0; i < binary.length; i += 1) {
    bytes[i] = binary.charCodeAt(i)
  }
  return bytes
}

const decodeBase64ToText = (data) => {
  try {
    const decoder = new TextDecoder()
    return decoder.decode(decodeBase64ToUint8(data))
  } catch (err) {
    console.warn('无法解析文本结果', err)
    return ''
  }
}

const base64ToObjectURL = (data, mime) => {
  const bytes = decodeBase64ToUint8(data)
  const blob = new Blob([bytes], { type: mime || 'application/octet-stream' })
  return URL.createObjectURL(blob)
}

const exportResult = () => {
  if (!lastResponse.value) {
    return
  }
  const mime = lastResponse.value.mimeType || 'application/octet-stream'
  const bytes = decodeBase64ToUint8(lastResponse.value.data)
  const blob = new Blob([bytes], { type: mime })
  const extension = outputFormat.value === 'txt' ? 'txt' : outputFormat.value
  const filename = `plantuml-${Date.now()}.${extension}`
  const url = URL.createObjectURL(blob)
  const link = document.createElement('a')
  link.href = url
  link.download = filename
  link.click()
  URL.revokeObjectURL(url)
}

const copyResult = async () => {
  if (!hasTextualResult.value) {
    return
  }
  try {
    const payload = outputFormat.value === 'svg' ? renderedSvg.value : renderedText.value
    await navigator.clipboard.writeText(payload)
  } catch (err) {
    console.warn('无法复制内容', err)
  }
}

const zoomIn = () => {
  previewScale.value = clampScale(previewScale.value + 0.15)
}

const zoomOut = () => {
  previewScale.value = clampScale(previewScale.value - 0.15)
}

const resetZoom = () => {
  previewScale.value = 1
}

const togglePreviewExpand = () => {
  previewExpanded.value = !previewExpanded.value
}

watch([outputFormat, serverPreset, customServer, timeoutSeconds, previewScale, previewExpanded], () => {
  persistSettings()
})

watch(serverPreset, (value, oldValue) => {
  if (revertingPreset) {
    return
  }
  if (value === 'builtin') {
    ensureJavaRuntime(oldValue)
  } else {
    javaReady.value = false
    javaVersion.value = ''
  }
})

onMounted(() => {
  loadLocal()
  if (serverPreset.value === 'builtin') {
    ensureJavaRuntime('local').finally(() => {
      if (!usingBuiltin.value || javaReady.value) {
        renderDiagram()
      }
    })
  } else {
    renderDiagram()
  }
})

onUnmounted(() => {
  if (renderedImageUrl.value) {
    URL.revokeObjectURL(renderedImageUrl.value)
  }
})
</script>

<style scoped>
.plantuml-workbench {
  min-height: calc(100vh - 120px);
}

.plantuml-preview-card {
  min-height: calc(100vh - 180px);
}

.preview-wrapper {
  min-height: 480px;
  height: calc(100vh - 260px);
  overflow: hidden;
}

.preview-surface {
  background-color: #fafafa;
  border-radius: 12px;
  padding: 12px;
  overflow: auto;
  height: 100%;
}

.preview-text {
  background-color: #111827;
  color: #e5e7eb;
  border-radius: 12px;
  padding: 12px;
  height: 100%;
  overflow: auto;
}

.preview-text pre {
  margin: 0;
  white-space: pre-wrap;
  word-break: break-word;
  font-size: 13px;
}

.preview-canvas {
  transform-origin: top left;
  display: inline-block;
  min-width: 100%;
}

.preview-canvas :deep(svg) {
  width: 100%;
  height: auto;
}

.preview-image {
  max-width: 100%;
  display: block;
}

.preview-zoom-controls .v-btn {
  min-width: 32px;
}

.preview-scale-display {
  width: 48px;
  text-align: center;
}
</style>
