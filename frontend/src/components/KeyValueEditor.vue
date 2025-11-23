<template>
  <div class="key-value-editor">
    <v-row dense v-for="(item, index) in modelValue" :key="index" class="mb-1 align-center">
      <v-col cols="auto" style="width: 40px;">
        <v-checkbox
            v-model="item.active"
            density="compact"
            hide-details
            color="primary"
        ></v-checkbox>
      </v-col>

      <v-col cols="4">
        <v-combobox
            v-model="item.key"
            :items="keySuggestions"
            placeholder="Key"
            density="compact"
            variant="outlined"
            hide-details
            auto-select-first
            :return-object="false"
            @update:model-value="onKeyChange(item, index)"
            @input="checkEmptyRow(index)"
        ></v-combobox>
      </v-col>

      <v-col>
        <v-combobox
            v-model="item.value"
            :items="getValueSuggestions(item.key)"
            placeholder="Value"
            density="compact"
            variant="outlined"
            hide-details
            :return-object="false"
            @update:model-value="checkEmptyRow(index)"
            @input="checkEmptyRow(index)"
        ></v-combobox>
      </v-col>

      <v-col cols="auto" style="width: 30px;">
        <v-btn
            v-if="index !== modelValue.length - 1"
            icon="mdi-delete"
            size="x-small"
            variant="text"
            color="grey"
            @click="removeRow(index)"
        ></v-btn>
      </v-col>
    </v-row>
  </div>
</template>

<script setup>
import { computed } from 'vue'

const props = defineProps({
  modelValue: Array,
  keyOptions: {
    type: Array,
    default: () => []
  }
})
const emit = defineEmits(['update:modelValue'])

const keySuggestions = computed(() => props.keyOptions)

// 定义常用 Key 的默认 Value
const defaultValues = {
  'content-type': 'application/json',
  'accept': 'application/json',
  'authorization': 'Bearer ', // 留个空格方便用户粘贴
  'user-agent': 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36',
  'connection': 'keep-alive',
  'cache-control': 'no-cache'
}

// Key 变化时触发
const onKeyChange = (item, index) => {
  checkEmptyRow(index)

  // 只有当 Value 为空时才自动填充，防止覆盖用户数据
  if (item.key && !item.value) {
    const lowerKey = item.key.toLowerCase()
    if (defaultValues[lowerKey]) {
      item.value = defaultValues[lowerKey]
    }
  }
}

const getValueSuggestions = (key) => {
  if (!key) return []
  const k = key.toLowerCase()
  if (k === 'content-type') return ['application/json', 'application/x-www-form-urlencoded', 'multipart/form-data', 'text/plain']
  if (k === 'authorization') return ['Bearer ', 'Basic '] // 提示时也给选项
  if (k === 'accept') return ['application/json', '*/*']
  if (k === 'connection') return ['keep-alive', 'close']
  if (k === 'cache-control') return ['no-cache', 'no-store']
  return []
}

const checkEmptyRow = (index) => {
  const list = props.modelValue
  if (index === list.length - 1) {
    if (list[index].key || list[index].value) {
      list.push({key: '', value: '', active: true})
      emit('update:modelValue', list)
    }
  }
}

const removeRow = (index) => {
  const list = props.modelValue
  list.splice(index, 1)
  emit('update:modelValue', list)
}
</script>

<style scoped>
.key-value-editor :deep(.v-field__input) {
  font-size: 13px;
  padding-top: 6px;
  padding-bottom: 6px;
  min-height: 32px;
}
</style>