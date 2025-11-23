<template>
  <v-container fluid class="h-100 pa-4 bg-grey-lighten-5">
    <v-row class="h-100">
      <v-col cols="12" md="7" class="d-flex flex-column">
        <v-card elevation="2" class="flex-grow-1 d-flex flex-column" rounded="lg">
          <v-card-item class="pb-0">
            <v-card-title class="text-subtitle-1 font-weight-bold text-primary">
              <v-icon icon="mdi-text-box-edit-outline" class="mr-2"></v-icon>
              输入内容
            </v-card-title>
          </v-card-item>

          <v-card-text class="flex-grow-1 pa-4">
            <v-textarea
                v-model="text"
                placeholder="在此输入或粘贴需要统计的文本..."
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

      <v-col cols="12" md="5">
        <v-card elevation="2" rounded="lg" class="h-100">
          <v-card-item>
            <v-card-title class="text-subtitle-1 font-weight-bold text-primary">
              <v-icon icon="mdi-chart-bar" class="mr-2"></v-icon>
              统计结果
            </v-card-title>
          </v-card-item>

          <v-divider></v-divider>

          <v-card-text class="pa-4">
            <v-row dense>
              <v-col cols="12" sm="6" v-for="item in statsList" :key="item.label">
                <v-card
                    variant="tonal"
                    :color="item.color"
                    class="mb-2 pa-1"
                    rounded="lg"
                >
                  <v-list-item>
                    <template v-slot:prepend>
                      <v-avatar :color="item.color" variant="flat" rounded size="small">
                        <v-icon :icon="item.icon" color="white" size="small"></v-icon>
                      </v-avatar>
                    </template>
                    <v-list-item-title class="font-weight-bold text-h6">
                      {{ item.value }}
                    </v-list-item-title>
                    <v-list-item-subtitle class="text-caption font-weight-bold opacity-70">
                      {{ item.label }}
                    </v-list-item-subtitle>
                  </v-list-item>
                </v-card>
              </v-col>
            </v-row>
          </v-card-text>
        </v-card>
      </v-col>
    </v-row>
  </v-container>
</template>

<script setup>
/**
 * StringStats Component
 *
 * Analyzes input text and provides statistics:
 * - Total length
 * - English characters
 * - Chinese characters
 * - Numbers
 * - Spaces
 * - Line count
 * - Estimated Hex byte count
 */

import { ref, computed } from 'vue'

// --- State ---
const text = ref('')

// --- Computed ---

/**
 * stats computes various metrics about the input text.
 */
const stats = computed(() => {
  const content = text.value || ''

  // 1. English (a-z, A-Z)
  const enMatch = content.match(/[a-zA-Z]/g)
  const enCount = enMatch ? enMatch.length : 0

  // 2. Chinese (Basic range)
  const cnMatch = content.match(/[\u4e00-\u9fa5]/g)
  const cnCount = cnMatch ? cnMatch.length : 0

  // 3. Numbers (0-9)
  const numMatch = content.match(/[0-9]/g)
  const numCount = numMatch ? numMatch.length : 0

  // 4. Spaces
  const spaceMatch = content.match(/ /g)
  const spaceCount = spaceMatch ? spaceMatch.length : 0

  // 5. Line Count
  const lineCount = content ? content.split(/\r\n|\r|\n/).length : 0

  // 6. Hex Bytes Estimation (assuming (En + Num) / 2)
  const hexCount = (enCount + numCount) / 2

  // 7. Total Length
  const totalCount = content.length

  return {
    enCount,
    cnCount,
    numCount,
    spaceCount,
    lineCount,
    hexCount,
    totalCount
  }
})

// List configuration for UI rendering
const statsList = computed(() => [
  { label: '总字符数', value: stats.value.totalCount, icon: 'mdi-sigma', color: 'blue-grey' },
  { label: '英文字符', value: stats.value.enCount, icon: 'mdi-alphabet-latin', color: 'indigo' },
  { label: '中文字符', value: stats.value.cnCount, icon: 'mdi-ideogram-chinese-dragon', color: 'red-darken-1' },
  { label: '数字', value: stats.value.numCount, icon: 'mdi-numeric', color: 'green' },
  { label: '空格', value: stats.value.spaceCount, icon: 'mdi-keyboard-space', color: 'orange' },
  { label: '总行数', value: stats.value.lineCount, icon: 'mdi-format-list-numbered', color: 'teal' },
  { label: 'Hex 字节', value: stats.value.hexCount, icon: 'mdi-code-braces', color: 'purple' },
])
</script>

<style scoped>
/* 强制让 textarea 撑满父容器的高度 */
.full-height-textarea :deep(.v-field__input) {
  height: 100% !important;
  font-family: 'Consolas', 'Monaco', monospace; /* 使用等宽字体方便查看 */
  line-height: 1.5;
}
.full-height-textarea :deep(.v-input__control),
.full-height-textarea :deep(.v-field) {
  height: 100%;
}
</style>