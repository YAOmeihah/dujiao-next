<template>
  <div v-bind="attrs" class="space-y-3">
    <button
      type="button"
      class="group w-full rounded-2xl border theme-surface-soft p-4 text-left transition-all duration-200 hover:-translate-y-0.5 hover:shadow-sm focus:outline-none focus:ring-2 focus:ring-primary/30"
      :class="invalid
        ? 'border-red-300 dark:border-red-500/70'
        : 'border-gray-200 dark:border-white/10 hover:border-primary/30'"
      @click="openSelector"
    >
      <div class="flex items-start justify-between gap-4">
        <div class="min-w-0">
          <div class="text-xs font-semibold uppercase tracking-[0.16em] theme-text-muted">
            {{ t('checkout.shippingRegionLabel') }}
          </div>
          <div class="mt-2 text-sm font-semibold leading-6" :class="selectedPath ? 'theme-text-primary' : 'theme-text-muted'">
            {{ selectedPath || t('checkout.shippingRegionPlaceholder') }}
          </div>
          <div class="mt-2 text-xs theme-text-muted">
            {{ selectedPath ? t('checkout.shippingRegionEditHint') : t('checkout.shippingRegionTip') }}
          </div>
        </div>
        <div
          class="flex h-11 w-11 shrink-0 items-center justify-center rounded-2xl border transition-colors"
          :class="selectedPath
            ? 'border-primary/20 bg-primary/10 text-primary'
            : 'border-gray-200 bg-white text-gray-400 dark:border-white/10 dark:bg-white/5 dark:text-gray-500'"
        >
          <svg class="h-5 w-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path
              stroke-linecap="round"
              stroke-linejoin="round"
              stroke-width="1.8"
              d="M9 5l7 7-7 7"
            />
          </svg>
        </div>
      </div>
    </button>

    <div v-if="selectedSegments.length" class="flex flex-wrap gap-2">
      <span
        v-for="segment in selectedSegments"
        :key="segment.key"
        class="rounded-full border border-primary/15 bg-primary/10 px-3 py-1.5 text-xs font-medium text-primary"
      >
        {{ segment.value }}
      </span>
    </div>
  </div>

  <Teleport to="body">
    <Transition
      enter-active-class="transition duration-200 ease-out"
      enter-from-class="opacity-0"
      enter-to-class="opacity-100"
      leave-active-class="transition duration-150 ease-in"
      leave-from-class="opacity-100"
      leave-to-class="opacity-0"
    >
      <div
        v-if="visible"
        class="fixed inset-0 z-[120] flex items-end justify-center p-0 md:items-center md:p-4"
        @click.self="closeSelector"
      >
        <div class="absolute inset-0 bg-black/45 backdrop-blur-sm" aria-hidden="true" @click="closeSelector" />

        <Transition
          enter-active-class="transition duration-300 ease-out"
          :enter-from-class="isDesktop ? 'opacity-0 scale-95 translate-y-2' : 'translate-y-full'"
          :enter-to-class="isDesktop ? 'opacity-100 scale-100 translate-y-0' : 'translate-y-0'"
          leave-active-class="transition duration-200 ease-in"
          :leave-from-class="isDesktop ? 'opacity-100 scale-100 translate-y-0' : 'translate-y-0'"
          :leave-to-class="isDesktop ? 'opacity-0 scale-95 translate-y-2' : 'translate-y-full'"
        >
          <div
            v-if="visible"
            class="relative z-10 flex w-full flex-col overflow-hidden border theme-panel shadow-2xl"
            :class="isDesktop
              ? 'max-h-[80vh] max-w-4xl rounded-[28px]'
              : 'h-[92dvh] rounded-t-[28px] border-b-0'"
          >
            <div class="flex min-h-0 flex-1 flex-col">
              <div class="border-b theme-border px-4 py-3 sm:px-6 sm:py-4">
                <div v-if="!isDesktop" class="mb-3 flex justify-center">
                  <span class="h-1.5 w-12 rounded-full bg-gray-300 dark:bg-white/15" />
                </div>
                <div class="flex items-start justify-between gap-4">
                  <div class="min-w-0">
                    <h3 class="text-lg font-black theme-text-primary">
                      {{ t('checkout.shippingSheetTitle') }}
                    </h3>
                    <p class="mt-1 text-sm theme-text-muted">
                      {{ selectedPath || t('checkout.shippingSheetTip') }}
                    </p>
                  </div>
                  <button
                    type="button"
                    class="flex h-10 w-10 shrink-0 items-center justify-center rounded-full border theme-btn-secondary"
                    @click="closeSelector"
                  >
                    <svg class="h-5 w-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path
                        stroke-linecap="round"
                        stroke-linejoin="round"
                        stroke-width="1.8"
                        d="M6 18L18 6M6 6l12 12"
                      />
                    </svg>
                  </button>
                </div>

                <div class="mt-4 space-y-3">
                  <button
                    type="button"
                    class="w-full rounded-3xl border px-3.5 py-3 text-left transition-all duration-200 sm:px-5 sm:py-4"
                    :class="activeStepCardClass"
                    @click="handleStepClick(activeStep.key)"
                  >
                    <div class="flex items-start justify-between gap-4">
                      <div class="min-w-0">
                        <div class="text-[11px] font-semibold uppercase tracking-[0.18em] opacity-70">
                          {{ activeStep.index }}
                        </div>
                        <div class="mt-1 text-sm font-semibold sm:text-lg">
                          {{ activeStep.label }}
                        </div>
                        <div class="mt-1 truncate text-xs opacity-90 sm:mt-2 sm:text-base">
                          {{ activeStep.value || t('checkout.shippingWaitingSelection') }}
                        </div>
                      </div>
                      <span
                        class="hidden shrink-0 rounded-full px-3 py-1 text-[11px] font-semibold uppercase tracking-[0.16em] sm:inline-flex"
                        :class="activeStepBadgeClass"
                      >
                        {{ t('checkout.shippingCurrentStep') }}
                      </span>
                    </div>
                  </button>

                  <div class="flex flex-wrap gap-2">
                    <button
                      v-for="step in stepItems"
                      :key="step.key"
                      type="button"
                      class="min-w-0 rounded-full border px-2.5 py-1.5 text-left transition-all duration-200 sm:px-3 sm:py-2"
                      :class="summaryStepClass(step)"
                      :disabled="!step.available"
                      @click="handleStepClick(step.key)"
                    >
                      <div class="flex items-center gap-2">
                        <span class="text-[11px] font-semibold uppercase tracking-[0.16em] opacity-70">
                          {{ step.index }}
                        </span>
                        <span class="text-xs font-semibold sm:text-sm">
                          {{ step.label }}
                        </span>
                        <span class="max-w-[6rem] truncate text-xs opacity-80 sm:max-w-[12rem]">
                          {{ step.value || t('checkout.shippingWaitingSelection') }}
                        </span>
                      </div>
                    </button>
                  </div>
                </div>
              </div>

              <div class="min-h-0 flex-1 overflow-y-auto overscroll-contain px-4 py-3 touch-pan-y sm:px-6 sm:py-5">
                <div class="mb-3 flex flex-wrap items-center gap-2 sm:mb-4">
                  <span class="rounded-full bg-primary/10 px-3 py-1 text-xs font-semibold text-primary">
                    {{ currentLevelLabel }}
                  </span>
                  <span v-if="isDesktop" class="text-xs theme-text-muted">
                    {{ t('checkout.shippingSelectStep') }}
                  </span>
                </div>

                <div v-if="currentLoading" class="rounded-2xl border theme-surface-soft px-4 py-5 text-sm theme-text-muted">
                  {{ t('common.loading') }}
                </div>

                <div
                  v-else-if="currentOptions.length"
                  class="grid grid-cols-1 gap-3 sm:grid-cols-2 xl:grid-cols-3"
                >
                  <button
                    v-for="option in currentOptions"
                    :key="option.code"
                    type="button"
                    class="group rounded-2xl border px-4 py-3 text-left transition-all duration-200 hover:-translate-y-0.5 hover:shadow-sm"
                    :class="optionButtonClass(option)"
                    @click="handleOptionSelect(option)"
                  >
                    <div class="flex items-center justify-between gap-3">
                      <span class="truncate text-sm font-semibold">{{ option.name }}</span>
                      <svg
                        v-if="isOptionSelected(option)"
                        class="h-4 w-4 shrink-0"
                        fill="none"
                        stroke="currentColor"
                        viewBox="0 0 24 24"
                      >
                        <path
                          stroke-linecap="round"
                          stroke-linejoin="round"
                          stroke-width="2"
                          d="M5 13l4 4L19 7"
                        />
                      </svg>
                    </div>
                  </button>
                </div>

                <div
                  v-else
                  class="rounded-2xl border border-dashed theme-border px-4 py-6 text-sm theme-text-muted"
                >
                  {{ t('checkout.shippingNoOptions') }}
                </div>
              </div>

              <div class="border-t theme-border px-4 py-3 sm:px-6 sm:py-4">
                <div v-if="isDesktop" class="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
                  <div class="min-w-0">
                    <div class="text-xs font-semibold uppercase tracking-[0.16em] theme-text-muted">
                      {{ t('checkout.shippingRegionSummary') }}
                    </div>
                    <div class="mt-1 truncate text-sm font-semibold theme-text-primary">
                      {{ selectedPath || t('checkout.shippingRegionPlaceholder') }}
                    </div>
                  </div>
                  <button
                    type="button"
                    class="theme-btn-inline-md theme-btn-secondary border"
                    @click="closeSelector"
                  >
                    {{ completeSelection ? t('checkout.shippingCompleteSelection') : t('common.cancel') }}
                  </button>
                </div>
                <div v-else>
                  <button
                    type="button"
                    class="w-full rounded-2xl border px-4 py-3 text-sm font-semibold theme-btn-secondary"
                    @click="closeSelector"
                  >
                    {{ completeSelection ? t('checkout.shippingCompleteSelection') : t('common.cancel') }}
                  </button>
                </div>
              </div>
            </div>
          </div>
        </Transition>
      </div>
    </Transition>
  </Teleport>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref, useAttrs, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { addressAPI } from '../../api'
import type { AddressDivisionOption, ShippingAddressFormValue } from '../../types/address'

type AddressLevel = 'province' | 'city' | 'district' | 'township'

const municipalityProvinceCodes = new Set(['11', '12', '31', '50'])

defineOptions({
  inheritAttrs: false,
})

const props = defineProps<{
  modelValue: ShippingAddressFormValue
  invalid?: boolean
}>()

const emit = defineEmits<{
  (e: 'update:modelValue', value: ShippingAddressFormValue): void
}>()

const { t } = useI18n()
const attrs = useAttrs()

const visible = ref(false)
const isDesktop = ref(false)
const activeLevel = ref<AddressLevel>('province')

const provinceOptions = ref<AddressDivisionOption[]>([])
const cityOptions = ref<AddressDivisionOption[]>([])
const districtOptions = ref<AddressDivisionOption[]>([])
const townshipOptions = ref<AddressDivisionOption[]>([])

const loading = ref({
  provinces: false,
  cities: false,
  districts: false,
  townships: false,
})

const selectedSegments = computed(() => [
  { key: 'province', value: props.modelValue.province },
  { key: 'city', value: props.modelValue.city },
  { key: 'district', value: props.modelValue.district },
  { key: 'township', value: props.modelValue.township },
].filter((item) => String(item.value || '').trim()))

const selectedPath = computed(() => selectedSegments.value.map((item) => item.value).join(' / '))

const completeSelection = computed(() => Boolean(
  props.modelValue.province_code
  && props.modelValue.city_code
  && props.modelValue.district_code
  && props.modelValue.township_code,
))

const stepItems = computed(() => [
  {
    key: 'province' as const,
    index: '01',
    label: t('checkout.shippingProvince'),
    value: props.modelValue.province,
    available: true,
  },
  {
    key: 'city' as const,
    index: '02',
    label: t('checkout.shippingCity'),
    value: props.modelValue.city,
    available: Boolean(props.modelValue.province_code),
  },
  {
    key: 'district' as const,
    index: '03',
    label: t('checkout.shippingDistrict'),
    value: props.modelValue.district,
    available: Boolean(props.modelValue.city_code),
  },
  {
    key: 'township' as const,
    index: '04',
    label: t('checkout.shippingTownship'),
    value: props.modelValue.township,
    available: Boolean(props.modelValue.district_code),
  },
])

const currentLevelLabel = computed(() => {
  const current = stepItems.value.find((item) => item.key === activeLevel.value)
  return current?.label || t('checkout.shippingProvince')
})

const activeStep = computed<{
  key: AddressLevel
  index: string
  label: string
  value: string
  available: boolean
}>(() => {
  return stepItems.value.find((item) => item.key === activeLevel.value) ?? {
    key: 'province',
    index: '01',
    label: t('checkout.shippingProvince'),
    value: '',
    available: true,
  }
})

const currentOptions = computed(() => {
  if (activeLevel.value === 'province') return provinceOptions.value
  if (activeLevel.value === 'city') return cityOptions.value
  if (activeLevel.value === 'district') return districtOptions.value
  return townshipOptions.value
})

const currentLoading = computed(() => {
  if (activeLevel.value === 'province') return loading.value.provinces
  if (activeLevel.value === 'city') return loading.value.cities
  if (activeLevel.value === 'district') return loading.value.districts
  return loading.value.townships
})

const updateViewport = () => {
  if (typeof window === 'undefined') return
  isDesktop.value = window.innerWidth >= 768
}

const emitValue = (patch: Partial<ShippingAddressFormValue>) => {
  emit('update:modelValue', {
    ...props.modelValue,
    ...patch,
  })
}

const lockBodyScroll = (locked: boolean) => {
  if (typeof document === 'undefined') return
  document.body.style.overflow = locked ? 'hidden' : ''
}

const onKeydown = (event: KeyboardEvent) => {
  if (event.key === 'Escape') {
    closeSelector()
  }
}

const fetchOptions = async (
  key: 'provinces' | 'cities' | 'districts' | 'townships',
  request: () => Promise<{ data: { data?: AddressDivisionOption[] } }>,
  target: typeof provinceOptions,
) => {
  loading.value[key] = true
  try {
    const response = await request()
    target.value = Array.isArray(response.data.data) ? response.data.data : []
  } catch {
    target.value = []
  } finally {
    loading.value[key] = false
  }
}

const ensureProvinceOptions = async () => {
  if (provinceOptions.value.length > 0) return
  await fetchOptions('provinces', () => addressAPI.provinces(), provinceOptions)
}

const ensureCityOptions = async (provinceCode: string) => {
  if (!provinceCode) {
    cityOptions.value = []
    return
  }
  await fetchOptions('cities', () => addressAPI.cities(provinceCode), cityOptions)
}

const ensureDistrictOptions = async (cityCode: string) => {
  if (!cityCode) {
    districtOptions.value = []
    return
  }
  await fetchOptions('districts', () => addressAPI.districts(cityCode), districtOptions)
}

const ensureTownshipOptions = async (districtCode: string) => {
  if (!districtCode) {
    townshipOptions.value = []
    return
  }
  await fetchOptions('townships', () => addressAPI.townships(districtCode), townshipOptions)
}

const determineActiveLevel = (value: ShippingAddressFormValue): AddressLevel => {
  if (!value.province_code) return 'province'
  if (!value.city_code) return 'city'
  if (!value.district_code) return 'district'
  return 'township'
}

const hydrateSelector = async () => {
  await ensureProvinceOptions()

  let nextValue = { ...props.modelValue }
  let changed = false

  if (!nextValue.province_code) {
    activeLevel.value = 'province'
    return
  }

  await ensureCityOptions(nextValue.province_code)

  if (municipalityProvinceCodes.has(nextValue.province_code) && !nextValue.city_code && cityOptions.value[0]) {
    nextValue = {
      ...nextValue,
      city: cityOptions.value[0].name,
      city_code: cityOptions.value[0].code,
    }
    changed = true
  }

  if (nextValue.city_code) {
    await ensureDistrictOptions(nextValue.city_code)
  } else {
    activeLevel.value = determineActiveLevel(nextValue)
    if (changed) emit('update:modelValue', nextValue)
    return
  }

  if (nextValue.district_code) {
    await ensureTownshipOptions(nextValue.district_code)
  }

  activeLevel.value = determineActiveLevel(nextValue)

  if (changed) {
    emit('update:modelValue', nextValue)
  }
}

const openSelector = async () => {
  visible.value = true
  await hydrateSelector()
}

const closeSelector = () => {
  visible.value = false
}

const activeStepCardClass = computed(() => 'border-primary/20 bg-primary/10 text-primary shadow-sm')

const activeStepBadgeClass = computed(() => 'bg-primary/15 text-primary')

const summaryStepClass = (step: {
  key: AddressLevel
  value: string
  available: boolean
}) => {
  if (!step.available) {
    return 'border-gray-200 bg-gray-50 text-gray-400 opacity-60 dark:border-white/10 dark:bg-white/5 dark:text-gray-500'
  }
  if (activeLevel.value === step.key) {
    return 'border-primary/20 bg-primary/10 text-primary'
  }
  if (step.value) {
    return 'border-emerald-200 bg-emerald-50 text-emerald-700 dark:border-emerald-500/30 dark:bg-emerald-500/10 dark:text-emerald-300'
  }
  return 'border-gray-200 bg-white theme-text-secondary hover:border-primary/20 hover:text-primary dark:border-white/10 dark:bg-white/5'
}

const optionButtonClass = (option: AddressDivisionOption) => isOptionSelected(option)
  ? 'border-primary/20 bg-primary/10 text-primary shadow-sm'
  : 'border-gray-200 bg-white theme-text-secondary hover:border-primary/20 hover:text-primary dark:border-white/10 dark:bg-white/5'

const handleStepClick = (level: AddressLevel) => {
  const target = stepItems.value.find((item) => item.key === level)
  if (!target?.available) return
  activeLevel.value = level
}

const handleProvinceSelect = async (option: AddressDivisionOption) => {
  let nextValue: ShippingAddressFormValue = {
    ...props.modelValue,
    province: option.name,
    province_code: option.code,
    city: '',
    city_code: '',
    district: '',
    district_code: '',
    township: '',
    township_code: '',
  }

  emit('update:modelValue', nextValue)

  await ensureCityOptions(option.code)

  if (municipalityProvinceCodes.has(option.code) && cityOptions.value[0]) {
    nextValue = {
      ...nextValue,
      city: cityOptions.value[0].name,
      city_code: cityOptions.value[0].code,
    }
    emit('update:modelValue', nextValue)
    await ensureDistrictOptions(nextValue.city_code)
    activeLevel.value = 'district'
    return
  }

  districtOptions.value = []
  townshipOptions.value = []
  activeLevel.value = 'city'
}

const handleCitySelect = async (option: AddressDivisionOption) => {
  emitValue({
    city: option.name,
    city_code: option.code,
    district: '',
    district_code: '',
    township: '',
    township_code: '',
  })
  await ensureDistrictOptions(option.code)
  townshipOptions.value = []
  activeLevel.value = 'district'
}

const handleDistrictSelect = async (option: AddressDivisionOption) => {
  emitValue({
    district: option.name,
    district_code: option.code,
    township: '',
    township_code: '',
  })
  await ensureTownshipOptions(option.code)
  activeLevel.value = 'township'
}

const handleTownshipSelect = (option: AddressDivisionOption) => {
  emitValue({
    township: option.name,
    township_code: option.code,
  })
  closeSelector()
}

const handleOptionSelect = async (option: AddressDivisionOption) => {
  if (activeLevel.value === 'province') {
    await handleProvinceSelect(option)
    return
  }
  if (activeLevel.value === 'city') {
    await handleCitySelect(option)
    return
  }
  if (activeLevel.value === 'district') {
    await handleDistrictSelect(option)
    return
  }
  handleTownshipSelect(option)
}

const isOptionSelected = (option: AddressDivisionOption) => {
  if (activeLevel.value === 'province') return props.modelValue.province_code === option.code
  if (activeLevel.value === 'city') return props.modelValue.city_code === option.code
  if (activeLevel.value === 'district') return props.modelValue.district_code === option.code
  return props.modelValue.township_code === option.code
}

watch(visible, (value) => {
  lockBodyScroll(value)
  if (value) {
    document.addEventListener('keydown', onKeydown)
    return
  }
  document.removeEventListener('keydown', onKeydown)
})

onMounted(() => {
  updateViewport()
  if (typeof window !== 'undefined') {
    window.addEventListener('resize', updateViewport)
  }
})

onBeforeUnmount(() => {
  lockBodyScroll(false)
  if (typeof window !== 'undefined') {
    window.removeEventListener('resize', updateViewport)
  }
  document.removeEventListener('keydown', onKeydown)
})
</script>
