export interface GuestOrderAuth {
  phone: string
  email: string
  order_password: string
}

export const GUEST_ORDER_AUTH_STORAGE_KEY = 'guest_order_auth'

export const createEmptyGuestOrderAuth = (): GuestOrderAuth => ({
  phone: '',
  email: '',
  order_password: '',
})

export const normalizeGuestOrderAuth = (value: Partial<GuestOrderAuth> | null | undefined): GuestOrderAuth => ({
  phone: String(value?.phone || '').trim(),
  email: String(value?.email || '').trim(),
  order_password: String(value?.order_password || ''),
})

export const hasGuestOrderAuth = (auth: Partial<GuestOrderAuth> | null | undefined) => {
  const normalized = normalizeGuestOrderAuth(auth)
  return Boolean(normalized.phone && normalized.order_password)
}

export const loadGuestOrderAuth = (storage: Storage = localStorage): GuestOrderAuth => {
  const raw = storage.getItem(GUEST_ORDER_AUTH_STORAGE_KEY)
  if (!raw) return createEmptyGuestOrderAuth()

  try {
    return normalizeGuestOrderAuth(JSON.parse(raw))
  } catch {
    return createEmptyGuestOrderAuth()
  }
}

export const saveGuestOrderAuth = (
  auth: Partial<GuestOrderAuth>,
  storage: Storage = localStorage,
) => {
  const normalized = normalizeGuestOrderAuth(auth)
  storage.setItem(GUEST_ORDER_AUTH_STORAGE_KEY, JSON.stringify(normalized))
  return normalized
}

export const clearGuestOrderAuth = (storage: Storage = localStorage) => {
  storage.removeItem(GUEST_ORDER_AUTH_STORAGE_KEY)
}

export const buildGuestOrderAuthParams = (auth: Partial<GuestOrderAuth>) => {
  const normalized = normalizeGuestOrderAuth(auth)
  return {
    phone: normalized.phone,
    ...(normalized.email ? { email: normalized.email } : {}),
    order_password: normalized.order_password,
  }
}
