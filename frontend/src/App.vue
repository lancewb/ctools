<template>
  <v-app>
    <v-navigation-drawer
        v-model="drawer"
        app
        floating
        class="bg-grey-lighten-5 border-none"
        width="260"
    >
      <div class="d-flex align-center px-6 py-5">
        <v-avatar color="primary" size="36" class="mr-3 elevation-2">
          <v-icon icon="mdi-hammer-wrench" size="20" color="white"></v-icon>
        </v-avatar>
        <div class="font-weight-bold text-h6 text-primary">Wails Toolbox</div>
      </div>

      <v-divider class="mx-4 mb-2 opacity-60"></v-divider>

      <v-list nav density="compact" class="px-2">
        <v-list-item
            prepend-icon="mdi-view-dashboard-outline"
            title="主页概览"
            to="/"
            color="primary"
            rounded="lg"
            class="mb-1 font-weight-medium"
        ></v-list-item>

        <v-list-group
            v-for="category in menuData"
            :key="category.id"
            :value="category.id"
        >
          <template v-slot:activator="{ props }">
            <v-list-item
                v-bind="props"
                :prepend-icon="category.icon"
                :title="category.title"
                rounded="lg"
                class="font-weight-medium mb-1"
            ></v-list-item>
          </template>

          <v-list-item
              v-for="child in category.children"
              :key="child.id"
              :title="child.title"
              :prepend-icon="child.icon"
              :to="`/tool/${category.id}/${child.id}`"
              color="primary"
              rounded="lg"
              class="mb-1 text-body-2"
          ></v-list-item>
        </v-list-group>
      </v-list>
    </v-navigation-drawer>

    <v-app-bar flat class="border-b pl-2" color="white" density="comfortable">
      <v-app-bar-nav-icon @click="drawer = !drawer" color="grey-darken-2"></v-app-bar-nav-icon>
      <div class="text-subtitle-1 font-weight-bold ml-2 text-grey-darken-3">
        {{ currentTitle }}
      </div>
      <v-spacer></v-spacer>
      <v-btn icon="mdi-information-outline" size="small" color="grey-darken-1" @click="infoDialog = true"></v-btn>
    </v-app-bar>

    <v-main class="bg-grey-lighten-5">
      <router-view v-slot="{ Component }">
        <transition name="fade" mode="out-in">
          <component :is="Component" />
        </transition>
      </router-view>
    </v-main>
    <v-dialog v-model="infoDialog" max-width="320">
      <v-card>
        <v-card-title class="text-subtitle-1 font-weight-bold">关于</v-card-title>
        <v-card-text>
          <div class="text-body-2">Maintainer: wangdinglei</div>
        </v-card-text>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn color="primary" @click="infoDialog = false">关闭</v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
  </v-app>
  <div class="cat-overlay">
    <img :src="catIllustration" alt="可爱喵咪" />
  </div>
</template>

<script setup>
import { ref, computed } from 'vue'
import { useRoute } from 'vue-router'
import { menuData } from './config/menu'
import catIllustration from './assets/cat.svg'

const drawer = ref(true)
const infoDialog = ref(false)
const route = useRoute()

const currentTitle = computed(() => {
  if (route.path === '/') return '控制台'
  const catId = route.params.catId
  const toolId = route.params.id

  const category = menuData.find(c => c.id === catId)
  if (category) {
    const tool = category.children.find(t => t.id === toolId)
    if (tool) return `${tool.title}`
  }
  return 'Wails 工具箱'
})
</script>

<style scoped>
.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.15s ease;
}
.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}
:deep(.v-list-group__items .v-list-item) {
  /* 原始值为 64px (calc(16px + var(--indent-padding)))，实在太大了 */
  /* 改为 32px，刚好容纳图标并有一点点缩进 */
  padding-inline-start: 32px !important;
  text-align: left;
}

.cat-overlay {
  position: fixed;
  right: 24px;
  bottom: 24px;
  width: 200px;
  max-width: 30vw;
  opacity: 0.1;
  pointer-events: none;
  z-index: 20;
}

.cat-overlay img {
  display: block;
  width: 100%;
  height: auto;
}

</style>
