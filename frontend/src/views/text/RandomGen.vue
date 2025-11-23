<template>
  <v-container fluid class="h-100 pa-4 bg-grey-lighten-5">
    <v-row class="h-100">
      <v-col cols="12" md="4" class="d-flex flex-column h-100">
        <v-card elevation="2" rounded="lg" class="flex-grow-1 d-flex flex-column">
          <v-card-item class="bg-blue-grey-lighten-5 py-3">
            <v-card-title class="text-subtitle-2 font-weight-bold">
              <v-icon icon="mdi-tune" class="mr-2"></v-icon>
              生成配置
            </v-card-title>
          </v-card-item>

          <v-divider></v-divider>

          <v-card-text class="pa-4">
            <v-row dense class="mb-4">
              <v-col cols="12">
                <div class="text-caption font-weight-bold mb-1 text-grey-darken-1">基本参数</div>
              </v-col>
              <v-col cols="6">
                <v-text-field
                    v-model.number="length"
                    label="单串长度"
                    type="number"
                    variant="outlined"
                    density="compact"
                    color="primary"
                    min="1"
                    hide-details
                ></v-text-field>
              </v-col>
              <v-col cols="6">
                <v-text-field
                    v-model.number="quantity"
                    label="生成数量"
                    type="number"
                    variant="outlined"
                    density="compact"
                    color="primary"
                    min="1"
                    hide-details
                ></v-text-field>
              </v-col>
            </v-row>

            <v-divider class="mb-4"></v-divider>

            <div class="text-caption font-weight-bold mb-2 text-grey-darken-1">字符集选项</div>

            <v-list density="compact" class="pa-0">
              <v-list-item class="pa-0">
                <v-checkbox
                    v-model="includeNumbers"
                    label="包含数字 (0-9)"
                    color="primary"
                    hide-details
                    density="compact"
                ></v-checkbox>
              </v-list-item>

              <v-list-item class="pa-0">
                <v-checkbox
                    v-model="includeLetters"
                    label="包含字母 (a-z, A-Z)"
                    color="primary"
                    hide-details
                    density="compact"
                ></v-checkbox>
              </v-list-item>

              <v-list-item class="pa-0">
                <v-checkbox
                    v-model="includeSpecial"
                    label="包含特殊符号 (=/][{}...)"
                    color="primary"
                    hide-details
                    density="compact"
                ></v-checkbox>
              </v-list-item>
            </v-list>

            <v-alert
                v-if="!isValidConfig"
                type="warning"
                variant="tonal"
                density="compact"
                class="mt-2 text-caption"
            >
              请至少选择一种字符类型
            </v-alert>
          </v-card-text>

          <v-spacer></v-spacer>

          <div class="pa-4">
            <v-btn
                block
                color="primary"
                size="large"
                prepend-icon="mdi-creation"
                @click="generateStrings"
                :disabled="!isValidConfig"
            >
              立即生成
            </v-btn>
          </div>
        </v-card>
      </v-col>

      <v-col cols="12" md="8" class="h-100">
        <v-card elevation="2" rounded="lg" class="h-100 d-flex flex-column">
          <v-card-item class="py-2">
            <div class="d-flex justify-space-between align-center">
              <span class="text-subtitle-2 font-weight-bold text-teal">
                <v-icon icon="mdi-format-list-bulleted" class="mr-1"></v-icon>
                生成结果
                <span v-if="resultLines > 0" class="text-grey text-caption ml-2">
                  ({{ resultLines }} 行)
                </span>
              </span>
              <div class="d-flex gap-2">
                <v-btn size="small" variant="text" icon="mdi-delete-outline" @click="clearOutput" title="清空"></v-btn>
                <v-btn size="small" variant="text" icon="mdi-content-copy" @click="copyToClipboard" title="复制"></v-btn>
              </div>
            </div>
          </v-card-item>

          <v-divider></v-divider>

          <v-card-text class="flex-grow-1 pa-0">
            <v-textarea
                v-model="output"
                variant="filled"
                readonly
                class="h-100 full-height-textarea no-padding-input"
                hide-details
                no-resize
                bg-color="grey-lighten-4"
                placeholder="结果将显示在这里..."
            ></v-textarea>
          </v-card-text>
        </v-card>
      </v-col>
    </v-row>

    <v-snackbar v-model="showSnackbar" timeout="2000" color="success" location="top">
      内容已复制到剪贴板
    </v-snackbar>
  </v-container>
</template>

<script setup>
import { ref, computed } from 'vue'

// 状态定义
const length = ref(16)
const quantity = ref(1)
const includeNumbers = ref(true) // 为了完备性，我加了数字选项，默认勾选
const includeLetters = ref(false) // 默认不勾选 (根据您的描述推测，或者您可以改为 true)
const includeSpecial = ref(false)

const output = ref('')
const showSnackbar = ref(false)

// 校验配置是否有效
const isValidConfig = computed(() => {
  return includeNumbers.value || includeLetters.value || includeSpecial.value
})

const resultLines = computed(() => {
  return output.value ? output.value.split('\n').length : 0
})

// 字符池定义
const CHARS_NUM = '0123456789'
const CHARS_LETTERS = 'abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ'
const CHARS_SPECIAL = '!@#$%^&*()_+-=[]{}|;:\'",./<>?'

// 生成逻辑
const generateStrings = () => {
  if (!isValidConfig.value) return

  // 1. 构建字符池
  let charset = ''
  if (includeNumbers.value) charset += CHARS_NUM
  if (includeLetters.value) charset += CHARS_LETTERS
  if (includeSpecial.value) charset += CHARS_SPECIAL

  const charsetLen = charset.length
  const resultArr = []

  // 2. 循环生成
  for (let i = 0; i < quantity.value; i++) {
    let str = ''
    // 使用 crypto.getRandomValues 获取更安全的随机数
    const randomValues = new Uint32Array(length.value)
    window.crypto.getRandomValues(randomValues)

    for (let j = 0; j < length.value; j++) {
      // 取模映射到字符池
      str += charset[randomValues[j] % charsetLen]
    }
    resultArr.push(str)
  }

  // 3. 输出
  output.value = resultArr.join('\n')
}

const clearOutput = () => {
  output.value = ''
}

const copyToClipboard = () => {
  if (output.value) {
    navigator.clipboard.writeText(output.value)
    showSnackbar.value = true
  }
}
</script>

<style scoped>
.full-height-textarea :deep(.v-field__input) {
  height: 100% !important;
  font-family: 'Consolas', 'Monaco', monospace; /* 等宽字体对齐更好看 */
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