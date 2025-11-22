<template>
  <v-container fluid class="fill-height align-start pa-6">
    <v-row>
      <v-col cols="12">
        <v-card elevation="2" rounded="lg">
          <v-card-item>
            <v-card-title class="text-h5 mb-2">
              <v-icon :icon="currentIcon" color="primary" class="mr-2"></v-icon>
              {{ title }}
            </v-card-title>
            <v-card-subtitle>功能 ID: {{ id }}</v-card-subtitle>
          </v-card-item>

          <v-divider></v-divider>

          <v-card-text class="pa-6">
            <v-alert
                color="primary"
                variant="tonal"
                icon="mdi-information"
                class="mb-4"
            >
              后端接口开发中...
            </v-alert>

            <div class="d-flex flex-column gap-4">
              <v-textarea
                  label="输入数据"
                  variant="outlined"
                  rows="5"
                  placeholder="在此输入待处理的内容..."
              ></v-textarea>

              <div class="d-flex justify-end">
                <v-btn color="primary" size="large" prepend-icon="mdi-play">
                  执行操作
                </v-btn>
              </div>

              <v-textarea
                  label="输出结果"
                  variant="outlined"
                  bg-color="grey-lighten-4"
                  readonly
                  rows="5"
              ></v-textarea>
            </div>
          </v-card-text>
        </v-card>
      </v-col>
    </v-row>
  </v-container>
</template>

<script setup>
import { computed } from 'vue'
import { useRoute } from 'vue-router'
import { menuData } from '../config/menu'

const route = useRoute()

// 根据路由参数查找当前功能的标题和图标
const id = computed(() => route.params.id)
const categoryId = computed(() => route.params.catId)

const currentItem = computed(() => {
  const category = menuData.find(c => c.id === categoryId.value)
  if (!category) return {}
  return category.children.find(child => child.id === id.value) || {}
})

const title = computed(() => currentItem.value.title || '未知功能')

// 查找父级图标作为默认图标
const currentIcon = computed(() => {
  const category = menuData.find(c => c.id === categoryId.value)
  return category ? category.icon : 'mdi-tools'
})
</script>