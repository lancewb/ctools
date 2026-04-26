<template>
  <v-container fluid class="pa-4">
    <v-card>
      <v-card-title class="text-subtitle-1 font-weight-bold">JWT 解析</v-card-title>
      <v-card-text>
        <v-textarea v-model="token" label="JWT" rows="5" variant="outlined" class="mono" />
        <v-alert v-if="error" type="error" variant="tonal" density="compact" class="mb-3">{{ error }}</v-alert>
        <v-row>
          <v-col cols="12" md="6"><v-textarea :model-value="header" label="Header" rows="10" readonly variant="outlined" class="mono" /></v-col>
          <v-col cols="12" md="6"><v-textarea :model-value="payload" label="Payload" rows="10" readonly variant="outlined" class="mono" /></v-col>
        </v-row>
      </v-card-text>
    </v-card>
  </v-container>
</template>

<script setup>
import { computed, ref } from 'vue'
const token = ref('')
const error = ref('')
const decodePart = part => {
  const normalized = part.replace(/-/g, '+').replace(/_/g, '/').padEnd(Math.ceil(part.length / 4) * 4, '=')
  return JSON.stringify(JSON.parse(decodeURIComponent(escape(atob(normalized)))), null, 2)
}
const parts = computed(() => token.value.trim().split('.'))
const header = computed(() => { try { error.value = ''; return parts.value[0] ? decodePart(parts.value[0]) : '' } catch (e) { error.value = e.message; return '' } })
const payload = computed(() => { try { return parts.value[1] ? decodePart(parts.value[1]) : '' } catch (e) { error.value = e.message; return '' } })
</script>

<style scoped>
.mono :deep(textarea) { font-family: ui-monospace, SFMono-Regular, Consolas, monospace; }
</style>
