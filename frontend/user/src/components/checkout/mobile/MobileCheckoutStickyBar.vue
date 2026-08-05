<template>
  <div
    data-mobile-floating-bar
    class="fixed inset-x-0 z-40 border-t theme-border theme-panel-strong shadow-2xl lg:hidden"
    style="bottom: calc(3.5rem + env(safe-area-inset-bottom, 0px));"
  >
    <div class="flex items-center gap-3 px-4 py-3">
      <div class="min-w-0 flex-1">
        <div class="text-sm font-bold theme-text-primary">{{ totalText }}</div>
        <div class="truncate text-xs theme-text-muted">{{ statusText }}</div>
      </div>
      <button
        data-mobile-primary-action
        type="button"
        class="theme-btn-inline-md theme-btn-primary min-w-[8rem] font-semibold"
        :disabled="disabled"
        @click="$emit('primaryAction')"
      >
        {{ actionLabel }}
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
defineProps<{
  totalText: string
  statusText: string
  actionLabel: string
  disabled: boolean
}>()

defineEmits<{
  (e: 'primaryAction'): void
}>()
</script>
