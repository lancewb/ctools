<template>
  <v-container fluid class="pa-4">
    <v-row>
      <v-col cols="12" md="5">
        <v-card>
          <v-card-title class="text-subtitle-1 font-weight-bold">编码转换</v-card-title>
          <v-card-text>
            <v-select v-model="mode" :items="modes" label="模式" variant="outlined" density="compact" />
            <v-textarea v-model="input" label="输入" rows="12" variant="outlined" class="mono" />
          </v-card-text>
        </v-card>
      </v-col>
      <v-col cols="12" md="7">
        <v-card>
          <v-card-title class="text-subtitle-1 font-weight-bold">输出</v-card-title>
          <v-card-text>
            <v-alert v-if="error" type="error" variant="tonal" density="compact" class="mb-3">{{ error }}</v-alert>
            <v-textarea :model-value="output" readonly rows="15" variant="outlined" class="mono" />
          </v-card-text>
        </v-card>
      </v-col>
    </v-row>
  </v-container>
</template>

<script setup>
import { computed, ref } from 'vue'
const modes = ['URL Encode', 'URL Decode', 'Base64 Encode', 'Base64 Decode', 'Hex Encode', 'Hex Decode']
const mode = ref('URL Encode')
const input = ref('hello world')
const error = ref('')
const output = computed(() => {
  error.value = ''
  try {
    const enc = new TextEncoder()
    const dec = new TextDecoder()
    if (mode.value === 'URL Encode') return encodeURIComponent(input.value)
    if (mode.value === 'URL Decode') return decodeURIComponent(input.value)
    if (mode.value === 'Base64 Encode') return btoa(unescape(encodeURIComponent(input.value)))
    if (mode.value === 'Base64 Decode') return decodeURIComponent(escape(atob(input.value.trim())))
    if (mode.value === 'Hex Encode') return [...enc.encode(input.value)].map(b => b.toString(16).padStart(2, '0')).join('')
    if (mode.value === 'Hex Decode') return dec.decode(new Uint8Array((input.value.replace(/\\s/g, '').match(/.{1,2}/g) || []).map(h => parseInt(h, 16))))
  } catch (e) {
    error.value = e.message
  }
  return ''
})
</script>

<style scoped>
.mono :deep(textarea) { font-family: ui-monospace, SFMono-Regular, Consolas, monospace; }
</style>
