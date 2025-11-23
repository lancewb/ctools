<template>
  <v-container fluid class="h-100 pa-4 bg-grey-lighten-5">
    <v-row class="h-100">
      <v-col cols="12" md="5" class="d-flex flex-column h-100">
        <v-card elevation="2" rounded="lg" class="flex-grow-1 d-flex flex-column mb-4">
          <v-card-item>
            <div class="d-flex justify-space-between align-center">
              <v-card-title class="text-subtitle-1 font-weight-bold text-primary">
                <v-icon icon="mdi-pencil-box-outline" class="mr-2"></v-icon>
                输入文本
              </v-card-title>
              <v-chip v-if="isBase64Detected" color="indigo" size="small" variant="flat" class="font-weight-bold">
                <v-icon start icon="mdi-auto-fix"></v-icon>
                已识别 Base64
              </v-chip>
            </div>
          </v-card-item>
          <v-card-text class="flex-grow-1 pa-4">
            <v-textarea
                v-model="inputText"
                placeholder="请输入需要处理的文本..."
                variant="outlined"
                class="h-100 full-height-textarea"
                hide-details
                no-resize
                @input="handleInput"
            ></v-textarea>
          </v-card-text>
        </v-card>

        <v-card elevation="2" rounded="lg">
          <v-card-item class="bg-blue-grey-lighten-5">
            <v-card-title class="text-subtitle-2 font-weight-bold">
              <v-icon icon="mdi-tune" class="mr-2"></v-icon>
              格式化参数
            </v-card-title>
          </v-card-item>
          <v-divider></v-divider>
          <v-card-text class="pa-4">
            <v-btn-toggle
                v-model="mode"
                color="primary"
                variant="outlined"
                divided
                mandatory
                class="w-100 mb-4 d-flex"
            >
              <v-btn value="insert" class="flex-grow-1">插入字符</v-btn>
              <v-btn value="delete" class="flex-grow-1">删除字符</v-btn>
            </v-btn-toggle>

            <v-row dense>
              <v-col cols="6">
                <v-text-field
                    v-model.number="interval"
                    label="间隔 (0表示不插入)"
                    type="number"
                    variant="outlined"
                    density="compact"
                    hide-details
                    min="0"
                ></v-text-field>
              </v-col>
              <v-col cols="6">
                <v-text-field
                    v-model="separator"
                    label="操作字符"
                    variant="outlined"
                    density="compact"
                    hide-details
                    placeholder="例如空格"
                ></v-text-field>
              </v-col>
            </v-row>

            <v-btn
                block
                color="primary"
                size="large"
                class="mt-4 font-weight-bold"
                prepend-icon="mdi-play-box-outline"
                @click="handleProcess"
            >
              开始处理
            </v-btn>
          </v-card-text>
        </v-card>
      </v-col>

      <v-col cols="12" md="7" class="d-flex flex-column h-100">
        <v-card elevation="2" rounded="lg" class="d-flex flex-column mb-4" style="height: 50%;">
          <v-card-item class="py-2">
            <div class="d-flex justify-space-between align-center">
              <span class="text-subtitle-2 font-weight-bold text-teal">
                <v-icon icon="mdi-check-circle-outline" class="mr-1"></v-icon>
                文本结果
              </span>
              <v-btn size="small" variant="text" icon="mdi-content-copy" @click="copyToClipboard(outputText)"></v-btn>
            </div>
          </v-card-item>
          <v-divider></v-divider>
          <v-card-text class="flex-grow-1 pa-0 position-relative"> <v-textarea
              v-model="outputText"
              variant="filled"
              readonly
              class="h-100 full-height-textarea no-padding-input"
              hide-details
              no-resize
              bg-color="grey-lighten-4"
          ></v-textarea>

            <div class="text-caption text-grey position-absolute" style="bottom: 5px; right: 12px; pointer-events: none; z-index: 5;">
              Length: {{ outputText.length }}
            </div>
          </v-card-text>
        </v-card>

        <v-card elevation="2" rounded="lg" class="d-flex flex-column" style="height: 50%;">
          <v-card-item class="py-2">
            <div class="d-flex justify-space-between align-center">
              <span class="text-subtitle-2 font-weight-bold text-indigo">
                <v-icon icon="mdi-code-braces" class="mr-1"></v-icon>
                Base64 编码结果
              </span>
              <v-btn size="small" variant="text" icon="mdi-content-copy" @click="copyToClipboard(outputBase64)"></v-btn>
            </div>
          </v-card-item>
          <v-divider></v-divider>
          <v-card-text class="flex-grow-1 pa-0 position-relative"> <v-textarea
              v-model="outputBase64"
              variant="filled"
              readonly
              class="h-100 full-height-textarea no-padding-input"
              hide-details
              no-resize
              bg-color="grey-lighten-4"
          ></v-textarea>

            <div class="text-caption text-grey position-absolute" style="bottom: 5px; right: 12px; pointer-events: none; z-index: 5;">
              Length: {{ outputBase64.length }}
            </div>
          </v-card-text>
        </v-card>
      </v-col>
    </v-row>

    <v-snackbar v-model="showSnackbar" timeout="2000" color="success" location="top">
      内容已复制
    </v-snackbar>
  </v-container>
</template>

<script setup>
/**
 * TextFormat Component
 *
 * Provides tools for text manipulation:
 * - Inserting separators (e.g. adding spaces every 2 chars for hex)
 * - Deleting separators (e.g. removing spaces)
 * - Auto-detecting and decoding Base64 input
 * - Base64 encoding output
 */

import { ref, computed } from 'vue'

// --- State ---
const inputText = ref('')
const outputText = ref('')
const outputBase64 = ref('')

// Parameters
const mode = ref('insert') // 'insert' | 'delete'
const interval = ref(0)
const separator = ref(' ')
const isBase64Detected = ref(false)

const showSnackbar = ref(false)

// --- Methods ---

/**
 * handleInput monitors typing to auto-detect Base64 content.
 */
const handleInput = () => {
  const raw = inputText.value.trim()

  // Heuristic check for Base64:
  // 1. Length is multiple of 4
  // 2. Contains only Base64 chars
  // 3. Contains non-hex chars (to differentiate from hex strings) or ends with '='
  const base64Pattern = /^[A-Za-z0-9+/]+={0,2}$/
  const hasNonHex = /[^0-9A-Fa-f]/.test(raw)

  if (raw.length > 0 && raw.length % 4 === 0 && base64Pattern.test(raw) && (hasNonHex || raw.endsWith('='))) {
    isBase64Detected.value = true
  } else {
    isBase64Detected.value = false
  }
}

/**
 * handleProcess executes the formatting logic based on selected mode.
 */
const handleProcess = () => {
  if (!inputText.value) return
  let b64 = false
  let source = inputText.value.trim()

  // 1. Auto-decode Base64 if detected
  if (isBase64Detected.value) {
    try {
      const decoded = atob(source)
      source = decoded
      b64 = true
    } catch (e) {
      console.warn("Base64 auto-detect failed, using raw input")
    }
  }

  let result = ''

  try {
    if (mode.value === 'insert') {
      const step = interval.value
      const char = separator.value

      if (step <= 0) {
        result = source
      } else {
        const regex = new RegExp(`.{1,${step}}`, 'g')
        const matches = source.match(regex)
        if (matches) {
          result = matches.join(char)
        } else {
          result = source
        }
      }

    } else if (mode.value === 'delete') {
      const char = separator.value
      if (char) {
        result = source.split(char).join('')
      } else {
        result = source
      }
    }

    outputText.value = result

    // 2. Calculate Base64 of result
    // If input was Base64, outputBase64 is the re-encoded input (normalized)
    if(!b64) {
      outputBase64.value = utf8_to_b64(result)
    } else {
      outputBase64.value = utf8_to_b64(inputText.value.trim())
    }

  } catch (e) {
    console.error(e)
    outputText.value = "Processing Error: " + e.message
  }
}

/**
 * utf8_to_b64 encodes string to Base64 handling UTF-8 characters safely.
 */
const utf8_to_b64 = (str) => {
  try {
    return window.btoa(unescape(encodeURIComponent(str)))
  } catch (e) {
    return "Encoding Failed"
  }
}

/**
 * copyToClipboard copies text to clipboard.
 */
const copyToClipboard = (text) => {
  if (text) {
    navigator.clipboard.writeText(text)
    showSnackbar.value = true
  }
}
</script>

<style scoped>
.full-height-textarea :deep(.v-field__input) {
  height: 100% !important;
  font-family: 'Consolas', 'Monaco', monospace;
  font-size: 0.95rem;
  line-height: 1.5;
}

.full-height-textarea :deep(.v-input__control),
.full-height-textarea :deep(.v-field) {
  height: 100%;
}

.no-padding-input :deep(.v-field__input) {
  padding-top: 12px;
}
</style>