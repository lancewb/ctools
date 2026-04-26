<template>
  <v-container fluid class="pa-4">
    <v-row>
      <v-col cols="12" md="6">
        <v-card>
          <v-card-title class="text-subtitle-1 font-weight-bold">JSON 输入</v-card-title>
          <v-card-text><v-textarea v-model="input" rows="18" variant="outlined" class="mono" /></v-card-text>
          <v-card-actions>
            <v-btn color="primary" prepend-icon="mdi-code-json" @click="format">格式化</v-btn>
            <v-btn variant="tonal" @click="minify">压缩</v-btn>
            <v-btn variant="text" @click="input = sample">示例</v-btn>
          </v-card-actions>
        </v-card>
      </v-col>
      <v-col cols="12" md="6">
        <v-card>
          <v-card-title class="text-subtitle-1 font-weight-bold">输出</v-card-title>
          <v-card-text>
            <v-alert v-if="error" type="error" variant="tonal" density="compact" class="mb-3">{{ error }}</v-alert>
            <v-textarea v-model="output" rows="18" readonly variant="outlined" class="mono" />
          </v-card-text>
        </v-card>
      </v-col>
    </v-row>
  </v-container>
</template>

<script setup>
import { ref } from 'vue'
const sample = '{"name":"ctools","items":[1,true,{"k":"v"}]}'
const input = ref(sample)
const output = ref('')
const error = ref('')
const parse = () => JSON.parse(input.value)
const format = () => { try { error.value = ''; output.value = JSON.stringify(parse(), null, 2) } catch (e) { error.value = e.message } }
const minify = () => { try { error.value = ''; output.value = JSON.stringify(parse()) } catch (e) { error.value = e.message } }
format()
</script>

<style scoped>
.mono :deep(textarea) { font-family: ui-monospace, SFMono-Regular, Consolas, monospace; }
</style>
