<template>
  <div
    class="mb-4 rounded-2xl border p-4 transition-colors"
    :class="muted
      ? 'border-gray-200 bg-gray-50/80 dark:border-white/10 dark:bg-white/5'
      : 'border-emerald-200 bg-emerald-50/80 dark:border-emerald-500/20 dark:bg-emerald-500/10'"
  >
    <div class="flex flex-col gap-4 md:flex-row md:items-start md:justify-between">
      <div class="min-w-0">
        <p class="text-sm font-semibold theme-text-primary">{{ applied ? appliedMessage : title }}</p>
        <div v-if="!applied" class="mt-2 space-y-1">
          <p
            v-for="line in summaryLines"
            :key="line"
            class="text-sm theme-text-secondary"
          >
            {{ line }}
          </p>
        </div>
      </div>

      <div class="flex shrink-0 flex-wrap gap-2">
        <template v-if="!applied">
          <button
            data-recall-use
            type="button"
            class="theme-btn-inline-md theme-btn-primary"
            @click="$emit('use')"
          >
            {{ useLabel }}
          </button>
          <button
            data-recall-rewrite
            type="button"
            class="theme-btn-inline-md border theme-btn-secondary"
            @click="$emit('rewrite')"
          >
            {{ rewriteLabel }}
          </button>
        </template>
        <template v-else>
          <button
            data-recall-clear-form
            type="button"
            class="theme-btn-inline-md border theme-btn-secondary"
            @click="$emit('clear-form')"
          >
            {{ clearFormLabel }}
          </button>
        </template>
      </div>
    </div>

    <button
      data-recall-clear-record
      type="button"
      class="mt-3 text-xs font-medium theme-text-muted underline-offset-2 hover:underline"
      @click="$emit('clear-record')"
    >
      {{ clearRecordLabel }}
    </button>
  </div>
</template>

<script setup lang="ts">
defineProps<{
  summaryLines: string[]
  title: string
  useLabel: string
  rewriteLabel: string
  appliedMessage: string
  clearFormLabel: string
  clearRecordLabel: string
  applied: boolean
  muted: boolean
}>()

defineEmits<{
  (e: 'use'): void
  (e: 'rewrite'): void
  (e: 'clear-form'): void
  (e: 'clear-record'): void
}>()
</script>
