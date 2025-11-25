<template>
  <v-container fluid class="mermaid-workbench">
    <v-row>
      <v-col
        v-show="!previewExpanded"
        cols="12"
        :md="previewExpanded ? 12 : 5"
        class="mb-4"
      >
        <v-card>
          <v-card-title class="text-subtitle-1 font-weight-bold">
            Mermaid 绘图
          </v-card-title>
          <v-card-text>
            <v-textarea
              v-model="diagramSource"
              label="Mermaid 源码"
              rows="14"
              auto-grow
              density="comfortable"
              spellcheck="false"
              class="mb-3"
            />
            <div class="text-body-2 font-weight-medium mb-1">渲染选项</div>
            <v-row dense>
              <v-col cols="12" md="6">
                <v-select
                  v-model="theme"
                  :items="themeOptions"
                  label="主题"
                  density="comfortable"
                />
              </v-col>
              <v-col cols="12" md="6">
                <v-switch
                  v-model="autoRender"
                  label="自动预览"
                  inset
                />
              </v-col>
              <v-col cols="12" md="6">
                <v-text-field
                  v-model.number="pngScale"
                  label="PNG 缩放倍数"
                  type="number"
                  min="1"
                  max="5"
                  density="comfortable"
                />
              </v-col>
              <v-col cols="12" md="6">
                <v-text-field
                  v-model="pngBackground"
                  label="PNG 背景色"
                  type="color"
                  density="comfortable"
                />
              </v-col>
            </v-row>
            <v-expand-transition>
              <div v-if="theme === 'custom'" class="custom-theme-editor">
                <v-textarea
                  v-model="customThemeJson"
                  label="自定义主题 (JSON)"
                  rows="6"
                  auto-grow
                  density="comfortable"
                  persistent-hint
                  :error="!!customThemeError"
                  :hint="customThemeError || '仅需要提供 themeVariables 对象，支持 mermaid 官方主题变量'"
                />
                <div class="text-caption text-medium-emphasis">
                  例如：{"primaryColor":"#0f62fe","primaryTextColor":"#ffffff","lineColor":"#0f62fe"}
                </div>
              </div>
            </v-expand-transition>
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
              命令面板支持
              <a
                href="https://mermaid.js.org/syntax/flowchart.html"
                target="_blank"
                rel="noreferrer"
              >
                Mermaid
              </a>
              官方语法，预览和导出均基于当前源码。
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
              color="secondary"
              :disabled="!diagramSource"
              @click="saveLocal"
              variant="text"
            >
              保存文本
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
        <v-card class="h-100 preview-card">
          <v-card-title class="text-subtitle-1 font-weight-bold d-flex align-center flex-wrap">
            实时预览
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
                :disabled="!renderedSvg"
                :loading="exportingSvg"
                @click="exportSvg"
              >
                导出 SVG
              </v-btn>
              <v-btn
                size="small"
                class="ml-1"
                variant="text"
                prepend-icon="mdi-image"
                :disabled="!renderedSvg"
                :loading="exportingPng"
                @click="exportPng"
              >
                导出 PNG
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
              v-else-if="renderedSvg"
              class="preview-surface"
            >
              <div
                class="preview-canvas"
                :style="previewCanvasStyle"
                v-html="renderedSvg"
              />
            </div>
            <div v-else class="text-medium-emphasis text-body-2">
              输入 Mermaid 语句以开始渲染。
            </div>
          </v-card-text>
        </v-card>
      </v-col>
    </v-row>
  </v-container>
</template>

<script setup>
/**
 * MermaidWorkbench Component
 *
 * 提供 Mermaid 文本编辑、实时预览与导出 PNG/SVG 功能。
 */

import { ref, watch, onMounted, onUnmounted, computed } from 'vue'
import mermaid from 'mermaid'

const defaultDiagram = `%% 示例流程图
graph TD
    Start[开始] --> Check{条件判断?}
    Check -->|是| Action1[执行任务]
    Check -->|否| Action2[补充处理]
    Action1 --> Done[完成]
    Action2 --> Done`

const defaultCustomTheme = JSON.stringify(
  {
    primaryColor: '#0f62fe',
    primaryTextColor: '#ffffff',
    primaryBorderColor: '#0f62fe',
    lineColor: '#0f62fe',
    fontSize: '14px',
    noteBkgColor: '#edf2ff',
    noteTextColor: '#1f2933'
  },
  null,
  2
)

const storageKeys = {
  diagram: 'mermaid-last',
  settings: 'mermaid-workbench-settings'
}

const minScale = 0.5
const maxScale = 3

const diagramSource = ref(defaultDiagram)
const renderedSvg = ref('')
const renderError = ref('')
const rendering = ref(false)
const autoRender = ref(true)
const theme = ref('default')
const customThemeJson = ref(defaultCustomTheme)
const customThemeError = ref('')
const pngScale = ref(2)
const pngBackground = ref('#ffffff')
const exportingSvg = ref(false)
const exportingPng = ref(false)
const lastRenderTime = ref('')
const previewScale = ref(1)
const previewExpanded = ref(false)
let debounceId

const themeOptions = [
  { title: '默认', value: 'default' },
  { title: '暗色', value: 'dark' },
  { title: '森林', value: 'forest' },
  { title: '中性', value: 'neutral' },
  { title: '自定义', value: 'custom' }
]

const templates = [
  {
    label: '流程图',
    code: `graph TD
    A[开始] --> B{条件?}
    B -->|通过| C[执行任务]
    B -->|失败| D[异常处理]
    C --> E[结束]
    D --> E`
  },
  {
    label: '时序图',
    code: `sequenceDiagram
    participant Alice
    participant Bob
    Alice->>Bob: 你好
    Bob-->>Alice: 欢迎
    Alice-)Bob: 完毕`
  },
  {
    label: '甘特图',
    code: `gantt
    dateFormat  YYYY-MM-DD
    title 项目计划
    section 设计
    原型设计 :done,    des1, 2024-01-06, 2024-01-08
    section 开发
    后端    :active, dev1, 2024-01-09, 5d
    前端    :         dev2, after dev1, 4d`
  }
]

const previewCanvasStyle = computed(() => ({
  transform: `scale(${previewScale.value})`,
  transformOrigin: 'top left',
  width: `${(1 / previewScale.value) * 100}%`
}))

const previewScaleDisplay = computed(() => Math.round(previewScale.value * 100))

const applyTemplate = (code) => {
  diagramSource.value = code
}

const reset = () => {
  diagramSource.value = defaultDiagram
  renderDiagram()
}

const saveLocal = () => {
  try {
    localStorage.setItem(storageKeys.diagram, diagramSource.value)
  } catch (err) {
    console.warn('无法保存 Mermaid 草稿', err)
  }
}

const clampScale = (value) => {
  if (Number.isNaN(Number(value))) {
    return 1
  }
  return Math.min(maxScale, Math.max(minScale, Number(value)))
}

const loadLocal = () => {
  try {
    const cached = localStorage.getItem(storageKeys.diagram)
    if (cached) {
      diagramSource.value = cached
    }
    const rawSettings = localStorage.getItem(storageKeys.settings)
    if (rawSettings) {
      const parsed = JSON.parse(rawSettings)
      if (typeof parsed.theme === 'string') {
        theme.value = parsed.theme
      }
      if (typeof parsed.customTheme === 'string') {
        customThemeJson.value = parsed.customTheme
      }
      if (typeof parsed.autoRender === 'boolean') {
        autoRender.value = parsed.autoRender
      }
      if (typeof parsed.previewScale === 'number') {
        previewScale.value = clampScale(parsed.previewScale)
      }
      if (typeof parsed.previewExpanded === 'boolean') {
        previewExpanded.value = parsed.previewExpanded
      }
      if (typeof parsed.pngScale === 'number' && !Number.isNaN(parsed.pngScale)) {
        pngScale.value = parsed.pngScale
      }
      if (typeof parsed.pngBackground === 'string') {
        pngBackground.value = parsed.pngBackground
      }
    }
  } catch (err) {
    console.warn('无法读取 Mermaid 草稿', err)
  }
}

const persistSettings = () => {
  try {
    const payload = {
      theme: theme.value,
      customTheme: customThemeJson.value,
      autoRender: autoRender.value,
      previewScale: previewScale.value,
      previewExpanded: previewExpanded.value,
      pngScale: pngScale.value,
      pngBackground: pngBackground.value
    }
    localStorage.setItem(storageKeys.settings, JSON.stringify(payload))
  } catch (err) {
    console.warn('无法保存 Mermaid 设置', err)
  }
}

const configureMermaid = () => {
  const config = {
    startOnLoad: false,
    securityLevel: 'loose',
    theme: theme.value
  }
  if (theme.value === 'custom') {
    try {
      const payload = customThemeJson.value?.trim()
        ? JSON.parse(customThemeJson.value)
        : {}
      if (typeof payload !== 'object' || Array.isArray(payload)) {
        throw new Error('自定义主题必须是对象')
      }
      config.theme = typeof payload.theme === 'string' ? payload.theme : 'base'
      config.themeVariables = payload.themeVariables ?? payload
      customThemeError.value = ''
    } catch (err) {
      customThemeError.value = err?.message ?? '自定义主题解析失败'
      config.theme = 'default'
      delete config.themeVariables
    }
  } else {
    customThemeError.value = ''
  }
  mermaid.initialize(config)
}

const renderDiagram = async () => {
  if (!diagramSource.value?.trim()) {
    renderedSvg.value = ''
    renderError.value = ''
    return
  }

  rendering.value = true
  renderError.value = ''
  configureMermaid()

  try {
    const { svg } = await mermaid.render(
      `mermaid-${Date.now()}`,
      diagramSource.value
    )
    renderedSvg.value = svg
    const now = new Date()
    lastRenderTime.value = now.toLocaleTimeString()
  } catch (err) {
    renderError.value = err?.message ?? String(err)
  } finally {
    rendering.value = false
  }
}

const debounceRender = () => {
  clearTimeout(debounceId)
  if (!autoRender.value) {
    return
  }
  debounceId = setTimeout(() => {
    renderDiagram()
  }, 400)
}

const exportSvg = () => {
  if (!renderedSvg.value) {
    return
  }
  exportingSvg.value = true
  try {
    const blob = new Blob([renderedSvg.value], {
      type: 'image/svg+xml;charset=utf-8'
    })
    triggerDownload(blob, `mermaid-${Date.now()}.svg`)
  } finally {
    exportingSvg.value = false
  }
}

const exportPng = async () => {
  if (!renderedSvg.value) {
    return
  }
  exportingPng.value = true
  try {
    const pngBlob = await svgToPng(
      renderedSvg.value,
      pngScale.value,
      pngBackground.value
    )
    triggerDownload(pngBlob, `mermaid-${Date.now()}.png`)
  } catch (err) {
    renderError.value = err?.message ?? String(err)
  } finally {
    exportingPng.value = false
  }
}

const triggerDownload = (blob, filename) => {
  const url = URL.createObjectURL(blob)
  const link = document.createElement('a')
  link.href = url
  link.download = filename
  link.click()
  URL.revokeObjectURL(url)
}

const svgToPng = (svgString, scale = 2, background = '#ffffff') => {
  return new Promise((resolve, reject) => {
    const svgBlob = new Blob([svgString], {
      type: 'image/svg+xml;charset=utf-8'
    })
    const url = URL.createObjectURL(svgBlob)
    const image = new Image()
    image.onload = () => {
      const { width, height } = measureSvg(svgString, image)
      const canvas = document.createElement('canvas')
      canvas.width = Math.max(1, Math.round(width * scale))
      canvas.height = Math.max(1, Math.round(height * scale))
      const ctx = canvas.getContext('2d')
      ctx.fillStyle = background || '#ffffff'
      ctx.fillRect(0, 0, canvas.width, canvas.height)
      ctx.drawImage(image, 0, 0, canvas.width, canvas.height)
      canvas.toBlob(
        (blob) => {
          if (blob) {
            resolve(blob)
          } else {
            reject(new Error('PNG 导出失败'))
          }
        },
        'image/png',
        1
      )
      URL.revokeObjectURL(url)
    }
    image.onerror = () => {
      URL.revokeObjectURL(url)
      reject(new Error('无法加载渲染结果'))
    }
    image.src = url
  })
}

const measureSvg = (svgString, fallbackImage) => {
  try {
    const parser = new DOMParser()
    const doc = parser.parseFromString(svgString, 'image/svg+xml')
    const svgEl = doc.documentElement
    const viewBox = svgEl.getAttribute('viewBox')
    let width = parseFloat(svgEl.getAttribute('width'))
    let height = parseFloat(svgEl.getAttribute('height'))
    if ((!width || !height) && viewBox) {
      const parts = viewBox.split(/\s+/)
      width = parseFloat(parts[2])
      height = parseFloat(parts[3])
    }
    if (!width || !height) {
      width = fallbackImage?.width || 800
      height = fallbackImage?.height || 600
    }
    return { width, height }
  } catch (err) {
    return {
      width: fallbackImage?.width || 800,
      height: fallbackImage?.height || 600
    }
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

watch(diagramSource, () => {
  debounceRender()
})

watch(theme, () => {
  configureMermaid()
  if (!autoRender.value) {
    return
  }
  renderDiagram()
})

watch(customThemeJson, () => {
  if (theme.value !== 'custom') {
    return
  }
  configureMermaid()
  debounceRender()
})

watch(autoRender, (enabled) => {
  if (enabled) {
    debounceRender()
  }
})

watch(
  [theme, autoRender, customThemeJson, previewScale, previewExpanded, pngScale, pngBackground],
  () => {
    persistSettings()
  }
)

onMounted(() => {
  configureMermaid()
  loadLocal()
  renderDiagram()
})

onUnmounted(() => {
  clearTimeout(debounceId)
})
</script>

<style scoped>
.mermaid-workbench {
  min-height: calc(100vh - 120px);
}

.preview-card {
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

.preview-canvas {
  transform-origin: top left;
  display: inline-block;
  min-width: 100%;
}

.preview-canvas :deep(svg) {
  width: 100%;
  height: auto;
}

.preview-zoom-controls .v-btn {
  min-width: 32px;
}

.preview-scale-display {
  width: 48px;
  text-align: center;
}

.custom-theme-editor {
  background-color: rgba(0, 0, 0, 0.02);
  border-radius: 12px;
  padding: 8px 12px 4px;
  margin-bottom: 8px;
}
</style>
