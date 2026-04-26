<template>
  <v-container fluid class="pa-4">
    <v-row>
      <v-col cols="12" md="6"><v-textarea v-model="left" label="左侧文本" rows="14" variant="outlined" class="mono" /></v-col>
      <v-col cols="12" md="6"><v-textarea v-model="right" label="右侧文本" rows="14" variant="outlined" class="mono" /></v-col>
    </v-row>
    <v-card>
      <v-card-title class="text-subtitle-1 font-weight-bold">逐行差异</v-card-title>
      <v-card-text>
        <pre class="diff"><template v-for="(line, i) in diffLines" :key="i"><div :class="line.type">{{ line.prefix }} {{ line.text }}</div></template></pre>
      </v-card-text>
    </v-card>
  </v-container>
</template>

<script setup>
import { computed, ref } from 'vue'
const left = ref('alpha\nbeta\ngamma')
const right = ref('alpha\nbeta2\ngamma\ndelta')
const diffLines = computed(() => {
  const a = left.value.split('\n')
  const b = right.value.split('\n')
  const max = Math.max(a.length, b.length)
  const rows = []
  for (let i = 0; i < max; i++) {
    if (a[i] === b[i]) rows.push({ type: 'same', prefix: ' ', text: a[i] ?? '' })
    else {
      if (a[i] !== undefined) rows.push({ type: 'del', prefix: '-', text: a[i] })
      if (b[i] !== undefined) rows.push({ type: 'add', prefix: '+', text: b[i] })
    }
  }
  return rows
})
</script>

<style scoped>
.mono :deep(textarea), .diff { font-family: ui-monospace, SFMono-Regular, Consolas, monospace; }
.diff { margin: 0; white-space: pre-wrap; }
.same { color: #455a64; }
.add { color: #1b5e20; background: #e8f5e9; }
.del { color: #b71c1c; background: #ffebee; }
</style>
