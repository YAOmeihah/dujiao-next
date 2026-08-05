export interface GuestShippingAddressRecallRecord {
  receiver_name: string
  receiver_phone: string
  province: string
  province_code: string
  city: string
  city_code: string
  district: string
  district_code: string
  township: string
  township_code: string
  detail_address: string
  saved_at: string
}

export type GuestShippingAddressRecallPayload = Omit<GuestShippingAddressRecallRecord, 'saved_at'>

export interface GuestShippingRecallEligibilityInput {
  orderRequiresShippingAddress: boolean
  isAuthenticated: boolean
  checkoutMode: 'guest' | 'member'
}

export const GUEST_SHIPPING_ADDRESS_STORAGE_KEY = 'guest_checkout_recent_shipping_address_v1'
export const GUEST_SHIPPING_ADDRESS_MAX_AGE_MS = 30 * 24 * 60 * 60 * 1000

const requiredKeys: Array<keyof GuestShippingAddressRecallRecord> = [
  'receiver_name',
  'receiver_phone',
  'province',
  'province_code',
  'city',
  'city_code',
  'district',
  'district_code',
  'township',
  'township_code',
  'detail_address',
  'saved_at',
]

const normalizeRecord = (value: GuestShippingAddressRecallRecord): GuestShippingAddressRecallRecord => ({
  ...value,
  receiver_name: value.receiver_name.trim(),
  receiver_phone: value.receiver_phone.trim(),
  province: value.province.trim(),
  province_code: value.province_code.trim(),
  city: value.city.trim(),
  city_code: value.city_code.trim(),
  district: value.district.trim(),
  district_code: value.district_code.trim(),
  township: value.township.trim(),
  township_code: value.township_code.trim(),
  detail_address: value.detail_address.trim(),
  saved_at: value.saved_at.trim(),
})

const isRecordShapeValid = (value: unknown): value is GuestShippingAddressRecallRecord => {
  if (!value || typeof value !== 'object') return false

  return requiredKeys.every((key) => String((value as Record<string, unknown>)[key] || '').trim() !== '')
}

const isRecordExpired = (savedAt: string, now = Date.now()) => {
  const parsed = Date.parse(savedAt)
  if (Number.isNaN(parsed)) return true

  return now - parsed > GUEST_SHIPPING_ADDRESS_MAX_AGE_MS
}

export const loadGuestShippingAddressRecall = (
  storage: Storage = localStorage,
  now = Date.now(),
) => {
  const raw = storage.getItem(GUEST_SHIPPING_ADDRESS_STORAGE_KEY)
  if (!raw) return null

  try {
    const parsed = JSON.parse(raw)
    if (!isRecordShapeValid(parsed)) return null

    const normalized = normalizeRecord(parsed)
    if (isRecordExpired(normalized.saved_at, now)) return null

    return normalized
  } catch {
    return null
  }
}

export const saveGuestShippingAddressRecall = (
  record: GuestShippingAddressRecallRecord,
  storage: Storage = localStorage,
) => {
  storage.setItem(
    GUEST_SHIPPING_ADDRESS_STORAGE_KEY,
    JSON.stringify(normalizeRecord(record)),
  )
}

export const createGuestShippingAddressRecallRecord = (
  payload: GuestShippingAddressRecallPayload,
  options: {
    storage?: Storage
    now?: string | Date
  } = {},
) => {
  const record: GuestShippingAddressRecallRecord = {
    ...payload,
    saved_at:
      typeof options.now === 'string'
        ? options.now
        : (options.now ?? new Date()).toISOString(),
  }

  saveGuestShippingAddressRecall(record, options.storage)
  return normalizeRecord(record)
}

export const clearGuestShippingAddressRecall = (storage: Storage = localStorage) => {
  storage.removeItem(GUEST_SHIPPING_ADDRESS_STORAGE_KEY)
}

export const shouldEnableGuestShippingAddressRecall = ({
  orderRequiresShippingAddress,
  isAuthenticated,
  checkoutMode,
}: GuestShippingRecallEligibilityInput) =>
  orderRequiresShippingAddress && !isAuthenticated && checkoutMode === 'guest'
