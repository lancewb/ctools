<template>
  <v-container fluid class="pa-4">
    <v-card>
      <v-card-title class="text-subtitle-1 font-weight-bold">颜色转换</v-card-title>
      <v-card-text>
        <v-row>
          <v-col cols="12" md="4"><v-text-field v-model="hex" label="HEX" variant="outlined" density="compact" /></v-col>
          <v-col cols="12" md="2"><div class="swatch" :style="{ background: normalizedHex }"></div></v-col>
          <v-col cols="12" md="6">
            <v-table density="compact"><tbody>
              <tr><td>RGB</td><td class="mono">{{ rgb }}</td></tr>
              <tr><td>CSS</td><td class="mono">{{ normalizedHex }}</td></tr>
            </tbody></v-table>
          </v-col>
        </v-row>
      </v-card-text>
    </v-card>
  </v-container>
</template>

<script setup>
import { computed, ref } from 'vue'
const hex = ref('#1976d2')
const normalizedHex = computed(() => {
  let v = hex.value.trim().replace('#', '')
  if (v.length === 3) v = v.split('').map(c => c + c).join('')
  return /^([0-9a-fA-F]{6})$/.test(v) ? `#${v.toLowerCase()}` : '#000000'
})
const rgb = computed(() => {
  const v = normalizedHex.value.slice(1)
  return `rgb(${parseInt(v.slice(0, 2), 16)}, ${parseInt(v.slice(2, 4), 16)}, ${parseInt(v.slice(4, 6), 16)})`
})
</script>

<style scoped>
.swatch { height: 84px; border-radius: 6px; border: 1px solid #cfd8dc; }
.mono { font-family: ui-monospace, SFMono-Regular, Consolas, monospace; }
</style>
