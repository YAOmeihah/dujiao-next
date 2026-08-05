<template>
  <div class="space-y-4 pb-36 lg:hidden">
    <div class="rounded-2xl border theme-panel p-4">
      <div class="text-xs font-semibold uppercase tracking-[0.16em] theme-text-muted">
        {{ topLabel }}
      </div>
      <Transition name="mobile-checkout-status" mode="out-in">
        <p :key="statusText" class="mt-2 text-sm font-semibold theme-text-primary">{{ statusText }}</p>
      </Transition>
    </div>

    <section
      v-for="section in sections"
      :key="section.key"
      :data-section-root="section.key"
      :data-section-state="section.recommended
        ? (expandedSection === section.key ? 'current-editing' : 'current')
        : section.complete
          ? 'complete'
          : 'idle'"
      :class="[
        'overflow-hidden rounded-2xl border theme-panel transition-all duration-300 ease-out',
        section.recommended ? 'shadow-lg shadow-black/5 ring-1 ring-white/40' : '',
      ]"
    >
      <button
        :data-section-toggle="section.key"
        type="button"
        class="flex w-full items-start justify-between gap-3 px-4 py-4 text-left transition-colors duration-200"
        @click="$emit('update:expandedSection', expandedSection === section.key ? null : section.key)"
      >
        <div class="min-w-0">
          <div class="flex items-center gap-2">
            <h2 class="text-base font-bold theme-text-primary">{{ section.title }}</h2>
            <span
              v-if="section.badge"
              :data-section-badge="section.key"
              class="theme-badge theme-badge-neutral"
            >
              {{ section.badge }}
            </span>
          </div>
          <div class="mt-2 space-y-1">
            <p
              v-for="line in section.summaryLines"
              :key="`${section.key}-${line}`"
              class="text-sm theme-text-secondary"
            >
              {{ line }}
            </p>
          </div>
          <p v-if="section.softHint" class="mt-2 text-xs text-amber-600 dark:text-amber-300">
            {{ section.softHint }}
          </p>
        </div>
        <span class="text-xs font-semibold theme-text-accent">
          {{ expandedSection === section.key ? collapseLabel : section.collapsedActionLabel || editLabel }}
        </span>
      </button>

      <Transition name="mobile-checkout-section">
        <div
          v-if="expandedSection === section.key"
          class="border-t theme-border px-4 py-4"
        >
          <p
            v-if="section.errorMessage"
            :data-section-error="section.key"
            class="mb-3 rounded-xl border border-red-200 bg-red-50 px-3 py-2 text-sm text-red-600 dark:border-red-500/30 dark:bg-red-500/10 dark:text-red-300"
          >
            {{ section.errorMessage }}
          </p>
          <slot :name="`section-${section.key}`" />
        </div>
      </Transition>
    </section>

    <MobileCheckoutStickyBar
      :total-text="totalText"
      :status-text="statusText"
      :action-label="primaryActionLabel"
      :disabled="primaryActionDisabled"
      @primary-action="$emit('primaryAction')"
    />
  </div>
</template>

<script setup lang="ts">
import MobileCheckoutStickyBar from './MobileCheckoutStickyBar.vue'

export interface MobileCheckoutDisplaySection {
  key: string
  title: string
  badge: string
  summaryLines: string[]
  errorMessage: string
  collapsedActionLabel?: string
  complete: boolean
  recommended: boolean
  softHint: string
}

withDefaults(
  defineProps<{
    sections: MobileCheckoutDisplaySection[]
    expandedSection: string | null
    topLabel?: string
    statusText: string
    totalText: string
    primaryActionLabel: string
    primaryActionDisabled: boolean
    editLabel?: string
    collapseLabel?: string
  }>(),
  {
    topLabel: '当前需要',
    editLabel: '修改',
    collapseLabel: '收起',
  },
)

defineEmits<{
  (e: 'update:expandedSection', value: string | null): void
  (e: 'primaryAction'): void
}>()
</script>

<style scoped>
.mobile-checkout-section-enter-active,
.mobile-checkout-section-leave-active,
.mobile-checkout-status-enter-active,
.mobile-checkout-status-leave-active {
  transition: opacity 220ms ease, transform 220ms ease;
}

.mobile-checkout-section-enter-from,
.mobile-checkout-section-leave-to,
.mobile-checkout-status-enter-from,
.mobile-checkout-status-leave-to {
  opacity: 0;
  transform: translateY(-8px);
}

[data-section-state='current'] {
  animation: mobile-checkout-glow 2.2s ease-in-out infinite;
}

@keyframes mobile-checkout-glow {
  0%,
  100% {
    transform: translateY(0);
    box-shadow: 0 0 0 rgba(15, 23, 42, 0);
  }

  50% {
    transform: translateY(-2px);
    box-shadow: 0 14px 32px rgba(15, 23, 42, 0.08);
  }
}

@media (prefers-reduced-motion: reduce) {
  .mobile-checkout-section-enter-active,
  .mobile-checkout-section-leave-active,
  .mobile-checkout-status-enter-active,
  .mobile-checkout-status-leave-active {
    transition: none;
  }

  [data-section-state='current'] {
    animation: none;
  }
}
</style>
