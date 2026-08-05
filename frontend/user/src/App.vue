<template>
  <div
    id="app"
    :class="route.meta.lockViewport === true
      ? 'h-[100dvh] overflow-hidden theme-page bg-background text-foreground flex flex-col'
      : 'min-h-screen theme-page bg-background text-foreground flex flex-col'"
  >
    <VaultLayout v-if="isVault && !isResellerConsole && route.name !== 'support'">
      <ErrorBoundary>
        <RouterView v-slot="{ Component }">
          <Transition name="page-fade" mode="out-in">
            <component :is="Component" />
          </Transition>
        </RouterView>
      </ErrorBoundary>
    </VaultLayout>

    <template v-else>
      <Navbar v-if="!isResellerConsole" />
      <main
        :class="route.meta.lockViewport === true
          ? 'flex-1 min-h-0 overflow-hidden pb-14 lg:pb-0'
          : isResellerConsole
            ? 'flex-1 min-h-0'
            : 'flex-1 min-h-0 pb-14 lg:pb-0'"
      >
        <SupportPage
          v-show="route.name === 'support'"
          :active="route.name === 'support'"
        />
        <ErrorBoundary v-if="route.name !== 'support'">
          <RouterView v-slot="{ Component }">
            <Transition name="page-fade" mode="out-in">
              <component :is="Component" />
            </Transition>
          </RouterView>
        </ErrorBoundary>
      </main>
      <Footer v-if="!isResellerConsole && route.meta.hideFooter !== true" />
      <BackToTop v-if="!isResellerConsole" />
      <MobileBottomNav v-if="!isResellerConsole" />
    </template>

    <Loading :loading="appStore.loading" />
    <Toast />
    <ConfirmDialog />
  </div>
</template>

<script setup lang="ts">
import { computed, defineAsyncComponent } from 'vue'
import { useRoute } from 'vue-router'
import { useAppStore } from './stores/app'
import { getActiveTemplate } from './templates/registry'
import Navbar from './components/Navbar.vue'
import Footer from './components/Footer.vue'
import Loading from './components/Loading.vue'
import Toast from './components/Toast.vue'
import ConfirmDialog from './components/ConfirmDialog.vue'
import ErrorBoundary from './components/ErrorBoundary.vue'
import BackToTop from './components/BackToTop.vue'
import MobileBottomNav from './components/MobileBottomNav.vue'
import SupportPage from './views/Support.vue'

const VaultLayout = defineAsyncComponent(() => import('./templates/vault/layout/VaultLayout.vue'))

const appStore = useAppStore()
const route = useRoute()
const isResellerConsole = computed(() => route.meta.resellerConsole === true)
const isVault = computed(() => getActiveTemplate() === 'vault')
</script>

<style>
.page-fade-enter-active,
.page-fade-leave-active {
  transition: opacity 200ms ease;
}

.page-fade-enter-from,
.page-fade-leave-to {
  opacity: 0;
}
</style>
