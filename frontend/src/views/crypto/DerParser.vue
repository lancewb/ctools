<template>
  <v-container fluid>
    <v-row>
      <v-col cols="12" md="4">
        <v-card>
          <v-card-title class="text-subtitle-1 font-weight-bold">DER 输入</v-card-title>
          <v-card-text>
            <v-select v-model="mode" :items="[{title:'Hex',value:'hex'},{title:'Base64',value:'base64'}]" label="格式" density="comfortable" />
            <v-textarea v-model="data" :label="mode === 'hex' ? 'Hex 数据' : 'Base64 数据'" rows="10" auto-grow />
          </v-card-text>
          <v-card-actions>
            <v-btn color="primary" :loading="loading" @click="parse">解析</v-btn>
            <v-spacer />
            <span class="text-error text-caption" v-if="errorMsg">{{ errorMsg }}</span>
          </v-card-actions>
        </v-card>
      </v-col>
      <v-col cols="12" md="8">
        <v-card>
          <v-card-title class="text-subtitle-1 font-weight-bold">结构</v-card-title>
          <v-card-text>
            <v-treeview :items="treeItems" activatable hoverable open-on-click item-title="label" item-value="id">
              <template #prepend="{ item }">
                <v-icon :color="item.constructed ? 'primary' : 'grey'">{{ item.constructed ? 'mdi-folder' : 'mdi-code-braces' }}</v-icon>
              </template>
              <template #append="{ item }">
                <span class="text-caption text-grey">{{ item.info }}</span>
              </template>
            </v-treeview>
          </v-card-text>
        </v-card>
      </v-col>
    </v-row>
  </v-container>
</template>

<script setup>
/**
 * DerParser Component
 *
 * Provides a UI for parsing and visualizing ASN.1 DER structures.
 * Supports Hex and Base64 input formats.
 */

import { ref } from 'vue'
import { ParseDER } from '../../../wailsjs/go/crypto/CryptoService'

// --- State ---
const mode = ref('hex')
const data = ref('')
const loading = ref(false)
const errorMsg = ref('')
const treeItems = ref([])

// --- Methods ---

/**
 * parse decodes the input string as DER and builds a tree structure.
 */
const parse = async () => {
  if (!data.value) {
    errorMsg.value = '请输入数据'
    return
  }
  errorMsg.value = ''
  loading.value = true
  try {
    const payload = await ParseDER(mode.value === 'hex' ? { hexString: data.value } : { base64: data.value })
    treeItems.value = payload.nodes.map(buildNode)
  } catch (err) {
    errorMsg.value = err?.message ?? String(err)
  } finally {
    loading.value = false
  }
}

let nodeCounter = 0

/**
 * buildNode recursively transforms backend DER nodes into treeview compatible objects.
 *
 * @param {Object} node - The DER node from the backend.
 */
const buildNode = (node) => {
  const id = ++nodeCounter
  return {
    id,
    label: `Tag ${node.tag} (${node.class})`,
    info: `len=${node.length}`,
    constructed: node.constructed,
    children: (node.children || []).map(buildNode)
  }
}
</script>
