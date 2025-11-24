<template>
  <v-container fluid>
    <v-row>
      <v-col cols="12" md="5">
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
      <v-col cols="12" md="7">
        <v-card class="h-100">
          <v-card-title class="text-subtitle-1 font-weight-bold d-flex align-center">
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
              v-html="renderedSvg"
            />
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

import { ref, watch, onMounted, onUnmounted } from 'vue'
import mermaid from 'mermaid'

const defaultDiagram = `%% 示例流程图
graph TD
    Start[开始] --> Check{条件判断?}
    Check -->|是| Action1[执行任务]
    Check -->|否| Action2[补充处理]
    Action1 --> Done[完成]
    Action2 --> Done`

const diagramSource = ref(defaultDiagram)
const renderedSvg = ref('')
const renderError = ref('')
const rendering = ref(false)
const autoRender = ref(true)
const theme = ref('default')
const pngScale = ref(2)
const pngBackground = ref('#ffffff')
const exportingSvg = ref(false)
const exportingPng = ref(false)
const lastRenderTime = ref('')
let debounceId

const themeOptions = [
  { title: '默认', value: 'default' },
  { title: '暗色', value: 'dark' },
  { title: '森林', value: 'forest' },
  { title: '中性', value: 'neutral' }
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

const applyTemplate = (code) => {
  diagramSource.value = code
}

const reset = () => {
  diagramSource.value = defaultDiagram
  renderDiagram()
}

const saveLocal = () => {
  try {
    localStorage.setItem('mermaid-last', diagramSource.value)
  } catch (err) {
    console.warn('无法保存 Mermaid 草稿', err)
  }
}

const loadLocal = () => {
  try {
    const cached = localStorage.getItem('mermaid-last')
    if (cached) {
      diagramSource.value = cached
    }
  } catch (err) {
    console.warn('无法读取 Mermaid 草稿', err)
  }
}

const configureMermaid = () => {
  mermaid.initialize({
    startOnLoad: false,
    securityLevel: 'loose',
    theme: theme.value
  })
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
    const pngBlob = await svgToPng(renderedSvg.value, pngScale.value, pngBackground.value)
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
    return { width: fallbackImage?.width || 800, height: fallbackImage?.height || 600 }
  }
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

watch(autoRender, (enabled) => {
  if (enabled) {
    debounceRender()
  }
})

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
.preview-wrapper {
  min-height: 420px;
}

.preview-surface {
  background-color: #fafafa;
  border-radius: 12px;
  padding: 12px;
  overflow: auto;
  max-height: 600px;
}

.preview-surface :deep(svg) {
  width: 100%;
  height: auto;
}
</style>
