<template>
  <v-container fluid class="h-100 pa-4 bg-grey-lighten-5">
    <v-row class="h-100">
      <v-col cols="12" md="12" class="d-flex flex-column" style="height: 50%;">
        <v-card elevation="2" class="flex-grow-1 d-flex flex-column" rounded="lg">
          <v-card-item class="pb-0">
            <div class="d-flex align-center justify-space-between">
              <v-card-title class="text-subtitle-1 font-weight-bold text-primary">
                <v-icon icon="mdi-import" class="mr-2"></v-icon>
                原始数据输入
              </v-card-title>

              <v-btn
                  color="primary"
                  prepend-icon="mdi-cog-transfer"
                  @click="processText"
                  class="px-6"
              >
                开始转换
              </v-btn>
            </div>
            <v-card-subtitle class="mt-1">
              支持：Wireshark格式、日志带时间戳格式、纯Hex空格分隔、纯Hex连续字符串
            </v-card-subtitle>
          </v-card-item>

          <v-card-text class="flex-grow-1 pa-4">
            <v-textarea
                v-model="inputText"
                placeholder="请粘贴您的 Hex 数据..."
                variant="outlined"
                color="primary"
                class="h-100 full-height-textarea"
                hide-details
                no-resize
                spellcheck="false"
            ></v-textarea>
          </v-card-text>
        </v-card>
      </v-col>

      <v-col cols="12" md="12" class="d-flex flex-column" style="height: 50%;">
        <v-row class="h-100">
          <v-col cols="12" md="6" class="h-100">
            <v-card elevation="2" class="h-100 d-flex flex-column" rounded="lg">
              <v-card-item>
                <div class="d-flex justify-space-between align-center">
                  <v-card-title class="text-subtitle-1 font-weight-bold text-indigo">
                    <v-icon icon="mdi-code-string" class="mr-2"></v-icon>
                    Hex 字符串
                  </v-card-title>
                  <v-btn size="small" variant="text" icon="mdi-content-copy" @click="copyToClipboard(outputHex)" title="复制"></v-btn>
                </div>
              </v-card-item>
              <v-card-text class="flex-grow-1 pa-0 position-relative"> <v-textarea
                  v-model="outputHex"
                  variant="filled"
                  readonly
                  class="h-100 full-height-textarea no-padding-input"
                  hide-details
                  no-resize
                  bg-color="grey-lighten-4"
              ></v-textarea>

                <div class="text-caption text-grey position-absolute" style="bottom: 5px; right: 12px; pointer-events: none; z-index: 5;">
                  Length: {{ outputHex.length }}
                </div>
              </v-card-text>
            </v-card>
          </v-col>

          <v-col cols="12" md="6" class="h-100">
            <v-card elevation="2" class="h-100 d-flex flex-column" rounded="lg">
              <v-card-item>
                <div class="d-flex justify-space-between align-center">
                  <v-card-title class="text-subtitle-1 font-weight-bold text-teal">
                    <v-icon icon="mdi-text-short" class="mr-2"></v-icon>
                    ASCII 结果 (Hexdump)
                  </v-card-title>
                  <v-btn size="small" variant="text" icon="mdi-content-copy" @click="copyToClipboard(outputAscii)" title="复制"></v-btn>
                </div>
              </v-card-item>
              <v-card-text class="flex-grow-1 pa-0 position-relative"> <v-textarea
                  v-model="outputAscii"
                  variant="filled"
                  readonly
                  class="h-100 full-height-textarea no-padding-input"
                  hide-details
                  no-resize
                  bg-color="grey-lighten-4"
              ></v-textarea>

                <div class="text-caption text-grey position-absolute" style="bottom: 5px; right: 12px; pointer-events: none; z-index: 5;">
                  Length: {{ outputAscii.length }}
                </div>
              </v-card-text>
            </v-card>
          </v-col>
        </v-row>
      </v-col>
    </v-row>

    <v-snackbar v-model="showSnackbar" timeout="2000" color="success" location="top">
      内容已复制到剪贴板
    </v-snackbar>
  </v-container>
</template>

<script setup>
/**
 * Hexdump Component
 *
 * Extracts raw hex data from mixed text (e.g. Wireshark dumps, logs) and converts it back to a clean Hex string and ASCII representation.
 * Useful for analyzing network packet captures or log files.
 */

import { ref } from 'vue'

// --- State ---
const inputText = ref('')
const outputHex = ref('')
const outputAscii = ref('')
const showSnackbar = ref(false)

// --- Methods ---

/**
 * processText cleans the input text and extracts Hex data.
 * It handles various formats like Wireshark output (timestamp offsets, ASCII preview).
 */
const processText = () => {
  if (!inputText.value) return

  const raw = inputText.value
  let hexStream = ""

  // 1. Check for pure Hex (continuous without spaces)
  const cleanAll = raw.replace(/\s/g, '')
  const isPureHex = /^[0-9A-Fa-f]+$/.test(cleanAll) && (cleanAll.length % 2 === 0)

  if (isPureHex && cleanAll.length > 30) {
    // Treat as pure hex stream if length > 30
    hexStream = cleanAll
  } else {
    // 2. Complex format processing (line by line)
    const lines = raw.split('\n')

    lines.forEach(line => {
      let currentLine = line.trim()
      if (!currentLine) return

      // Step A: Remove right-side ASCII preview (often separated by 2+ spaces)
      if (currentLine.includes('  ')) {
        const parts = currentLine.split('  ')
        currentLine = parts[0]
      }

      // Step B: Remove left-side timestamp/offset (ends with :)
      if (currentLine.includes(':')) {
        const parts = currentLine.split(':')
        currentLine = parts[parts.length - 1]
      }

      // Step C: Extract remaining valid Hex pairs
      const matches = currentLine.match(/[0-9A-Fa-f]{2}/g)
      if (matches) {
        hexStream += matches.join('')
      }
    })
  }

  // 3. Set output
  outputHex.value = hexStream

  // 4. Convert to ASCII
  outputAscii.value = hexToAscii(hexStream)
}

/**
 * hexToAscii converts a hex string to its ASCII representation.
 * Non-printable characters are replaced with dots ('.').
 *
 * @param {string} hex - The hex string.
 * @returns {string} The ASCII string.
 */
const hexToAscii = (hex) => {
  let str = ''
  for (let i = 0; i < hex.length; i += 2) {
    const code = parseInt(hex.substr(i, 2), 16)
    if (code >= 32 && code <= 126) {
      str += String.fromCharCode(code)
    } else {
      str += '.'
    }
  }
  return str
}

/**
 * copyToClipboard copies text to the clipboard.
 */
const copyToClipboard = async (text) => {
  if (!text) return
  try {
    await navigator.clipboard.writeText(text)
    showSnackbar.value = true
  } catch (err) {
    console.error('Copy failed', err)
  }
}
</script>

<style scoped>
/* 样式调整：让输入框充满高度，并使用等宽字体 */
.full-height-textarea :deep(.v-field__input) {
  height: 100% !important;
  font-family: 'Consolas', 'Monaco', 'Courier New', monospace;
  font-size: 0.9rem;
  line-height: 1.4;
}
.full-height-textarea :deep(.v-input__control),
.full-height-textarea :deep(.v-field) {
  height: 100%;
}

/* 去掉只读框的内边距，让文字贴边显示，利用率更高 */
.no-padding-input :deep(.v-field__input) {
  padding-top: 10px;
}
</style>