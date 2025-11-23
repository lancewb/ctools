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
import { ref } from 'vue'

const inputText = ref('')
const outputHex = ref('')
const outputAscii = ref('')
const showSnackbar = ref(false)

// 核心处理逻辑
const processText = () => {
  if (!inputText.value) return

  const raw = inputText.value
  let hexStream = ""

  // 1. 判断是否为纯 Hex 字符串 (连续无空格，如 30313233...)
  // 移除所有空白字符后，如果全是 hex 且长度为偶数，则直接处理
  const cleanAll = raw.replace(/\s/g, '')
  const isPureHex = /^[0-9A-Fa-f]+$/.test(cleanAll) && (cleanAll.length % 2 === 0)

  if (isPureHex && cleanAll.length > 30) {
    // 长度大于30才判定为纯文本，防止把短的 "30 31" 误判
    hexStream = cleanAll
  } else {
    // 2. 复杂格式处理 (按行解析)
    const lines = raw.split('\n')

    lines.forEach(line => {
      let currentLine = line.trim()
      if (!currentLine) return

      // 步骤 A: 去除右侧的 ASCII 可视化部分
      // 特征：通常会有两个以上的空格，或者竖线 | 分隔
      // 我们找到最后一个由"双空格"分隔的部分，如果它看起来像 ASCII，就丢弃
      if (currentLine.includes('  ')) {
        // 取双空格之前的部分（即去掉了右边的 ASCII 预览）
        // 例子: "30 31  01" -> "30 31"
        // 注意：split 可能会分成多段，通常 hex 部分在前面
        const parts = currentLine.split('  ')
        // 简单策略：只取第一部分，通常 hex 都在最左边（或者左边有时间戳）
        // 但为了保险，我们重新拼装除了最后一部分之外的内容，或者截断
        // 对于您的例子，右侧 ASCII 前面都有 3-4 个空格
        currentLine = parts[0]
      }

      // 步骤 B: 去除左侧的时间戳/偏移量
      // 特征：通常以冒号 : 结尾，如 "2025... 0:" 或 "00000000:"
      if (currentLine.includes(':')) {
        const parts = currentLine.split(':')
        // 取最后一个冒号之后的部分
        // 例子: "time: 0: 30 31" -> " 30 31"
        currentLine = parts[parts.length - 1]
      }

      // 步骤 C: 提取剩余部分的 Hex
      // 经过上面两步，剩下应该是 " 30 31 32 ..." 或者纯粹的 hex
      // 我们提取所有成对的 Hex 字符
      const matches = currentLine.match(/[0-9A-Fa-f]{2}/g)
      if (matches) {
        hexStream += matches.join('')
      }
    })
  }

  // 3. 输出 Hex 字符串
  outputHex.value = hexStream

  // 4. 转换 Hex 为 ASCII (Hexdump)
  outputAscii.value = hexToAscii(hexStream)
}

// Hex 转 ASCII 字符串
const hexToAscii = (hex) => {
  let str = ''
  for (let i = 0; i < hex.length; i += 2) {
    const code = parseInt(hex.substr(i, 2), 16)
    // 根据您的需求，似乎是直接转换所有可打印字符
    // 标准 ASCII 可打印范围是 32 (space) 到 126 (~)
    // 您的例子中 '30' -> '0', '40' -> '@' 都在此范围内
    // 如果遇到不可打印字符，通常显示为 '.'，但按您需求我直接转出来
    if (code >= 32 && code <= 126) {
      str += String.fromCharCode(code)
    } else {
      // 0x00 - 0x1F 或 > 0x7E，显示为点，或者直接不显示？
      // 既然您的例子是连续的，我这里用 '.' 占位，保持对齐，或者您可以选择忽略
      // 观察您的例1：hex len 16 -> ascii len 16。说明是一一对应的。
      str += '.'
    }
  }
  return str
}

// 剪贴板工具
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