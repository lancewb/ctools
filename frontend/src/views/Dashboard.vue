<template>
  <v-container fluid class="bg-grey-lighten-5 pa-6">
    <v-row class="mb-4">
      <v-col cols="12">
        <h1 class="text-h5 font-weight-bold text-primary">工具箱</h1>
      </v-col>
    </v-row>

    <template v-for="category in menuData" :key="category.id">
      <v-row no-gutters class="mt-2 mb-2 align-center">
        <v-col cols="auto" class="d-flex align-center mr-3">
          <v-icon :color="category.color" size="small" class="mr-2">{{ category.icon }}</v-icon>
          <span class="text-subtitle-1 font-weight-bold text-grey-darken-2">
            {{ category.title }}
          </span>
        </v-col>
        <v-col>
          <v-divider></v-divider>
        </v-col>
      </v-row>

      <v-row dense class="mb-5">
        <v-col
            v-for="tool in category.children"
            :key="tool.id"
            cols="12" sm="6" md="4" lg="3" xl="2"
        >
          <v-hover v-slot="{ isHovering, props }">
            <v-card
                v-bind="props"
                :elevation="isHovering ? 3 : 0"
                :class="['cursor-pointer tool-card border', isHovering ? 'on-hover' : '']"
                @click="navigateToTool(category.id, tool.id)"
                flat
            >
              <v-card-item density="comfortable">
                <template v-slot:prepend>
                  <v-avatar :color="category.color" variant="tonal" size="32" rounded>
                    <v-icon :icon="tool.icon" size="18"></v-icon>
                  </v-avatar>
                </template>
                <v-card-title class="text-body-2 font-weight-bold pl-2">
                  {{ tool.title }}
                </v-card-title>
              </v-card-item>
            </v-card>
          </v-hover>
        </v-col>
      </v-row>
    </template>
  </v-container>
</template>

<script setup>
import { useRouter } from 'vue-router'
import { menuData } from '../config/menu'

const router = useRouter()

const navigateToTool = (catId, toolId) => {
  router.push(`/tool/${catId}/${toolId}`)
}
</script>

<style scoped>
.tool-card {
  background-color: white;
  transition: all 0.2s ease-in-out;
  border-color: #e0e0e0;
}
.on-hover {
  transform: translateY(-2px);
  border-color: rgb(var(--v-theme-primary)) !important;
}
</style>