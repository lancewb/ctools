<template>
  <v-container fluid class="pa-4">
    <v-card>
      <v-card-title class="text-subtitle-1 font-weight-bold">正则测试</v-card-title>
      <v-card-text>
        <v-row dense>
          <v-col cols="12" md="8"><v-text-field v-model="pattern" label="Pattern" variant="outlined" density="compact" /></v-col>
          <v-col cols="12" md="4"><v-text-field v-model="flags" label="Flags" hint="g i m s u" persistent-hint variant="outlined" density="compact" /></v-col>
          <v-col cols="12"><v-textarea v-model="text" label="测试文本" rows="8" variant="outlined" class="mono" /></v-col>
        </v-row>
        <v-alert v-if="error" type="error" variant="tonal" density="compact">{{ error }}</v-alert>
      </v-card-text>
    </v-card>
    <v-card class="mt-4">
      <v-card-title class="text-subtitle-1 font-weight-bold">匹配 {{ matches.length }}</v-card-title>
      <v-table density="compact">
        <thead><tr><th>#</th><th>Index</th><th>Match</th><th>Groups</th></tr></thead>
        <tbody>
          <tr v-for="(m, i) in matches" :key="i"><td>{{ i + 1 }}</td><td>{{ m.index }}</td><td class="mono">{{ m.value }}</td><td class="mono">{{ m.groups }}</td></tr>
        </tbody>
      </v-table>
    </v-card>
  </v-container>
</template>

<script setup>
import { computed, ref } from 'vue'
const pattern = ref('\\b\\w+@\\w+\\.\\w+\\b')
const flags = ref('gi')
const text = ref('admin@example.com\nops@test.local')
const error = ref('')
const matches = computed(() => {
  error.value = ''
  try {
    const f = flags.value.includes('g') ? flags.value : flags.value + 'g'
    const re = new RegExp(pattern.value, f)
    return [...text.value.matchAll(re)].map(m => ({ index: m.index, value: m[0], groups: m.slice(1).join(', ') }))
  } catch (e) {
    error.value = e.message
    return []
  }
})
</script>

<style scoped>
.mono { font-family: ui-monospace, SFMono-Regular, Consolas, monospace; }
.mono :deep(textarea) { font-family: ui-monospace, SFMono-Regular, Consolas, monospace; }
</style>
