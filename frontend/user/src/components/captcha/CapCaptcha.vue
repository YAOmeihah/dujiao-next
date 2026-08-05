<script setup lang="ts">
import { computed, nextTick, onMounted, ref, watch } from 'vue'

const loadedCapWidgetScripts = new Set<string>()

const normalizeEndpoint = (value?: string) => String(value || '').trim().replace(/\/+$/, '')

const loadCapWidgetScript = (src: string) => {
  if (!src || typeof document === 'undefined') return
  if (typeof customElements !== 'undefined' && customElements.get('cap-widget')) return

  const absoluteSrc = typeof window !== 'undefined' ? new URL(src, window.location.href).href : src
  const existingScript = Array.from(document.scripts).some((script) => script.src === absoluteSrc)
  if (loadedCapWidgetScripts.has(absoluteSrc) && existingScript) return
  if (existingScript) {
    loadedCapWidgetScripts.add(absoluteSrc)
    return
  }

  const script = document.createElement('script')
  script.src = absoluteSrc
  script.async = true
  script.dataset.capWidget = 'true'
  script.onerror = () => {
    loadedCapWidgetScripts.delete(absoluteSrc)
    console.error(`Failed to load Cap widget script: ${absoluteSrc}`)
  }
  document.head.appendChild(script)
  loadedCapWidgetScripts.add(absoluteSrc)
}

const props = defineProps<{
  modelValue?: string
  endpoint?: string
  siteKey?: string
}>()

const emit = defineEmits<{
  (event: 'update:modelValue', value: string): void
}>()

const widgetKey = ref(0)
const capEndpoint = computed(() => normalizeEndpoint(props.endpoint))
const widgetScriptSrc = computed(() => {
  if (!capEndpoint.value) return ''
  return `${capEndpoint.value}/assets/widget.js`
})
const apiEndpoint = computed(() => {
  const siteKey = String(props.siteKey || '').trim().replace(/^\/+|\/+$/g, '')
  if (!capEndpoint.value || !siteKey) return ''
  return `${capEndpoint.value}/${siteKey}/`
})

const handleSolve = (event: Event) => {
  const customEvent = event as CustomEvent<{ token?: string }>
  emit('update:modelValue', String(customEvent.detail?.token || ''))
}

const reset = async () => {
  emit('update:modelValue', '')
  widgetKey.value += 1
  await nextTick()
}

watch(apiEndpoint, reset)
watch(widgetScriptSrc, loadCapWidgetScript)

onMounted(() => {
  loadCapWidgetScript(widgetScriptSrc.value)
})

defineExpose({
  reset,
})
</script>

<template>
  <div class="cap-captcha min-h-[65px]">
    <cap-widget
      v-if="apiEndpoint"
      :key="widgetKey"
      :data-cap-api-endpoint="apiEndpoint"
      @solve="handleSolve"
      @reset="emit('update:modelValue', '')"
    />
  </div>
</template>

<style scoped>
.cap-captcha :deep(cap-widget) {
  --cap-widget-width: 100%;
  --cap-widget-height: 58px;
  --cap-widget-padding: 14px;
  --cap-gap: 16px;
  --cap-background: rgba(255, 255, 255, 0.96);
  --cap-border-color: rgba(148, 163, 184, 0.35);
  --cap-border-radius: 12px;
  --cap-color: #1f2937;
  --cap-checkbox-border: 1px solid rgba(148, 163, 184, 0.85);
  --cap-checkbox-background: #ffffff;
  --cap-checkbox-border-radius: 7px;
  --cap-spinner-color: #0b7bed;
  --cap-spinner-background-color: rgba(15, 23, 42, 0.12);
  max-width: 100%;
}
</style>
