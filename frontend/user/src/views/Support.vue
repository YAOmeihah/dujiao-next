<template>
  <div class="h-full min-h-0 theme-page overflow-hidden pt-20 pb-2 sm:pb-4 lg:pb-6">
    <div class="container mx-auto h-full px-2 sm:px-4">
      <div class="mx-auto relative h-full min-h-0 overflow-hidden rounded-3xl border theme-panel backdrop-blur-xl">
          <iframe
            v-if="pageState === 'loading' || pageState === 'ready'"
            :src="supportURL"
            class="block h-full w-full bg-white transition-opacity duration-300"
            :class="{ 'opacity-0 pointer-events-none': pageState === 'loading' }"
            :title="t('support.iframeTitle')"
            loading="lazy"
            referrerpolicy="strict-origin-when-cross-origin"
            @load="handleIframeLoad"
          />

          <div
            v-if="pageState === 'loading'"
            class="absolute inset-0 flex items-center justify-center p-8 backdrop-blur-sm"
          >
            <div class="max-w-xl text-center space-y-4">
              <div class="mx-auto flex h-16 w-16 items-center justify-center rounded-full theme-surface-soft border theme-border">
                <svg class="w-8 h-8 theme-text-muted" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.75" d="M8 12h.01M12 12h.01M16 12h.01M21 12c0 4.418-4.03 8-9 8a9.863 9.863 0 01-4.255-.949L3 20l1.395-3.72C3.512 15.042 3 13.574 3 12c0-4.418 4.03-8 9-8s9 3.582 9 8z" />
                </svg>
              </div>
              <h2 class="text-2xl font-bold theme-text-primary">{{ stateTitle }}</h2>
              <p class="theme-text-secondary whitespace-pre-line">{{ stateDescription }}</p>
            </div>
          </div>

          <div
            v-else-if="pageState === 'empty' || pageState === 'invalid' || pageState === 'fallback'"
            class="flex h-full min-h-0 items-center justify-center overflow-y-auto p-8"
          >
            <div class="max-w-xl text-center space-y-4">
              <div class="mx-auto flex h-16 w-16 items-center justify-center rounded-full theme-surface-soft border theme-border">
                <svg class="w-8 h-8 theme-text-muted" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.75" d="M8 12h.01M12 12h.01M16 12h.01M21 12c0 4.418-4.03 8-9 8a9.863 9.863 0 01-4.255-.949L3 20l1.395-3.72C3.512 15.042 3 13.574 3 12c0-4.418 4.03-8 9-8s9 3.582 9 8z" />
                </svg>
              </div>
              <h2 class="text-2xl font-bold theme-text-primary">{{ stateTitle }}</h2>
              <p class="theme-text-secondary whitespace-pre-line">{{ stateDescription }}</p>
              <a
                v-if="supportURL && pageState === 'fallback'"
                :href="supportURL"
                target="_blank"
                rel="noopener noreferrer"
                class="inline-flex items-center justify-center gap-2 px-4 py-3 rounded-xl theme-btn-secondary border font-semibold"
              >
                {{ t('support.openInNewTab') }}
              </a>
            </div>
          </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAppStore } from '../stores/app'

const FALLBACK_TIMEOUT_MS = 8000
const props = withDefaults(defineProps<{
  active?: boolean
}>(), {
  active: true,
})

const { t } = useI18n()
const appStore = useAppStore()

const pageState = ref<'loading' | 'ready' | 'empty' | 'invalid' | 'fallback'>('loading')
const hasInitialized = ref(false)
let fallbackTimer: number | null = null

const rawSupportURL = computed(() => String(appStore.config?.contact?.support_url || '').trim())
const supportURL = computed(() => (/^https?:\/\//i.test(rawSupportURL.value) ? rawSupportURL.value : ''))

const clearFallbackTimer = () => {
  if (fallbackTimer !== null) {
    window.clearTimeout(fallbackTimer)
    fallbackTimer = null
  }
}

const startFallbackTimer = () => {
  clearFallbackTimer()
  fallbackTimer = window.setTimeout(() => {
    if (pageState.value === 'loading') {
      pageState.value = 'fallback'
    }
  }, FALLBACK_TIMEOUT_MS)
}

const resolvePageState = () => {
  clearFallbackTimer()
  if (!rawSupportURL.value) {
    pageState.value = 'empty'
    return
  }
  if (!supportURL.value) {
    pageState.value = 'invalid'
    return
  }
  pageState.value = 'loading'
  startFallbackTimer()
}

const handleIframeLoad = () => {
  clearFallbackTimer()
  pageState.value = 'ready'
}

const ensureInitialized = async () => {
  if (!appStore.config) {
    await appStore.loadConfig()
  }
  if (!hasInitialized.value) {
    hasInitialized.value = true
    resolvePageState()
  }
}

const handleActiveChange = async (active: boolean) => {
  if (!active) {
    clearFallbackTimer()
    return
  }

  if (!hasInitialized.value) {
    await ensureInitialized()
    return
  }

  if (pageState.value === 'loading') {
    startFallbackTimer()
  }
}

const stateTitle = computed(() => {
  switch (pageState.value) {
    case 'empty':
      return t('support.emptyTitle')
    case 'invalid':
      return t('support.invalidTitle')
    case 'fallback':
      return t('support.fallbackTitle')
    default:
      return t('support.loadingTitle')
  }
})

const stateDescription = computed(() => {
  switch (pageState.value) {
    case 'empty':
      return t('support.emptyDescription')
    case 'invalid':
      return t('support.invalidDescription')
    case 'fallback':
      return t('support.fallbackDescription')
    default:
      return t('support.loadingDescription')
  }
})

watch(rawSupportURL, () => {
  if (hasInitialized.value) {
    resolvePageState()
  }
})

watch(() => props.active, (active) => {
  void handleActiveChange(active)
}, { immediate: true })

onBeforeUnmount(() => {
  clearFallbackTimer()
})
</script>
