<template>
  <section
    data-support-page
    :data-state="pageState"
    class="h-full min-h-0 overflow-hidden theme-page"
    :class="layoutClasses"
  >
    <div class="mx-auto h-full w-full px-2 sm:px-4" :class="layout === 'vault' ? 'max-w-[1180px]' : 'container'">
      <div class="relative h-full min-h-0 overflow-hidden rounded-lg border theme-panel">
        <iframe
          v-if="supportUrl"
          :src="supportUrl"
          class="block h-full w-full bg-white transition-opacity duration-200"
          :class="pageState === 'ready' ? 'opacity-100' : 'pointer-events-none opacity-0'"
          :title="t('support.iframeTitle')"
          loading="lazy"
          referrerpolicy="strict-origin-when-cross-origin"
          @load="handleIframeLoad"
        />

        <div
          v-if="pageState !== 'ready'"
          class="absolute inset-0 flex items-center justify-center overflow-y-auto bg-background/95 p-6 sm:p-8"
          role="status"
          aria-live="polite"
        >
          <div class="max-w-xl space-y-4 text-center">
            <div class="mx-auto flex h-14 w-14 items-center justify-center rounded-full border theme-surface-soft theme-border">
              <LoaderCircle v-if="pageState === 'loading'" class="h-7 w-7 animate-spin theme-text-muted" />
              <MessageCircle v-else class="h-7 w-7 theme-text-muted" />
            </div>
            <h1 class="text-xl font-bold theme-text-primary sm:text-2xl">{{ stateTitle }}</h1>
            <p class="whitespace-pre-line theme-text-secondary">{{ stateDescription }}</p>
            <a
              v-if="pageState === 'fallback' && supportUrl"
              data-support-external
              :href="supportUrl"
              target="_blank"
              rel="noopener noreferrer"
              class="inline-flex cursor-pointer items-center justify-center gap-2 rounded-lg border px-4 py-3 font-semibold transition-colors theme-btn-secondary focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-primary/50"
            >
              <ExternalLink class="h-4 w-4" />
              {{ t('support.openInNewTab') }}
            </a>
          </div>
        </div>
      </div>
    </div>
  </section>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { ExternalLink, LoaderCircle, MessageCircle } from 'lucide-vue-next'
import { useAppStore } from '../stores/app'
import { getSupportUrl } from '../utils/supportUrl'

type SupportState = 'loading' | 'ready' | 'empty' | 'invalid' | 'fallback'

const FALLBACK_TIMEOUT_MS = 8000
const props = withDefaults(defineProps<{
  active?: boolean
  layout?: 'classic' | 'vault'
}>(), {
  active: true,
  layout: 'classic',
})

const { t } = useI18n()
const appStore = useAppStore()
const pageState = ref<SupportState>('loading')
let fallbackTimer: number | null = null

const rawSupportUrl = computed(() => String(appStore.config?.contact?.support_url || '').trim())
const supportUrl = computed(() => getSupportUrl(appStore.config))
const layoutClasses = computed(() => props.layout === 'vault'
  ? 'py-2'
  : 'pb-2 pt-20 sm:pb-4 lg:pb-6')

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
  if (!rawSupportUrl.value) {
    pageState.value = 'empty'
    return
  }
  if (!supportUrl.value) {
    pageState.value = 'invalid'
    return
  }

  pageState.value = 'loading'
  if (props.active) {
    startFallbackTimer()
  }
}

const handleIframeLoad = () => {
  clearFallbackTimer()
  pageState.value = 'ready'
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

watch(rawSupportUrl, resolvePageState, { immediate: true })
watch(() => props.active, (active) => {
  if (!active) {
    clearFallbackTimer()
  } else if (pageState.value === 'loading') {
    startFallbackTimer()
  }
})

onBeforeUnmount(clearFallbackTimer)
</script>
