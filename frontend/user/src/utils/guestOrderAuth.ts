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

const EMPTY_GUEST_ORDER_AUTH = createEmptyGuestOrderAuth()

type GuestOrderAuthStorage = 'sessionStorage' | 'localStorage'

let volatileGuestOrderAuth: GuestOrderAuth | null = null

export const normalizeGuestOrderAuth = (auth: Partial<GuestOrderAuth> | null | undefined): GuestOrderAuth => ({
  phone: String(auth?.phone || '').trim(),
  email: String(auth?.email || '').trim(),
  order_password: String(auth?.order_password || ''),
})

export const hasGuestOrderAuth = (auth: Partial<GuestOrderAuth> | null | undefined) => {
  const normalized = normalizeGuestOrderAuth(auth)
  return Boolean(normalized.phone && normalized.order_password)
}

const cloneGuestOrderAuth = (auth: GuestOrderAuth): GuestOrderAuth => ({ ...auth })

const parseGuestOrderAuth = (raw: string | null): GuestOrderAuth | null => {
  if (!raw) return null
  try {
    const parsed = JSON.parse(raw) as Partial<GuestOrderAuth>
    return normalizeGuestOrderAuth(parsed)
  } catch {
    return null
  }
}

const readStorage = (storageName: GuestOrderAuthStorage, key: string): string | null => {
  try {
    return window[storageName].getItem(key)
  } catch {
    return null
  }
}

const writeStorage = (storageName: GuestOrderAuthStorage, key: string, value: string) => {
  try {
    window[storageName].setItem(key, value)
  } catch {
    // 浏览器禁用存储时由当前页面内存状态继续承接。
  }
}

const removeStorage = (storageName: GuestOrderAuthStorage, key: string) => {
  try {
    window[storageName].removeItem(key)
  } catch {
    // 浏览器禁用存储时只保留当前页面内存状态。
  }
}

// loadGuestOrderAuth 优先读取当前标签页的 sessionStorage，并执行一次旧版
// localStorage -> sessionStorage 迁移。迁移后立即删除长期存储中的游客凭据。
export const loadGuestOrderAuth = (storage?: Storage): GuestOrderAuth => {
  if (storage) {
    try {
      return parseGuestOrderAuth(storage.getItem(GUEST_ORDER_AUTH_STORAGE_KEY)) || createEmptyGuestOrderAuth()
    } catch {
      return createEmptyGuestOrderAuth()
    }
  }
  if (typeof window === 'undefined') {
    return cloneGuestOrderAuth(EMPTY_GUEST_ORDER_AUTH)
  }
  if (volatileGuestOrderAuth) {
    return cloneGuestOrderAuth(volatileGuestOrderAuth)
  }

  const sessionRaw = readStorage('sessionStorage', GUEST_ORDER_AUTH_STORAGE_KEY)
  const legacyRaw = readStorage('localStorage', GUEST_ORDER_AUTH_STORAGE_KEY)
  const sessionAuth = parseGuestOrderAuth(sessionRaw)
  const legacyAuth = parseGuestOrderAuth(legacyRaw)
  const parsed = sessionAuth || legacyAuth

  if (parsed) {
    volatileGuestOrderAuth = cloneGuestOrderAuth(parsed)
  }
  if (!sessionAuth && legacyAuth) {
    writeStorage('sessionStorage', GUEST_ORDER_AUTH_STORAGE_KEY, JSON.stringify(legacyAuth))
  }
  if (legacyRaw !== null) {
    removeStorage('localStorage', GUEST_ORDER_AUTH_STORAGE_KEY)
  }
  return cloneGuestOrderAuth(volatileGuestOrderAuth || EMPTY_GUEST_ORDER_AUTH)
}

export const saveGuestOrderAuth = (auth: Partial<GuestOrderAuth>, storage?: Storage) => {
  const normalized = normalizeGuestOrderAuth(auth)
  if (storage) {
    storage.setItem(GUEST_ORDER_AUTH_STORAGE_KEY, JSON.stringify(normalized))
    return normalized
  }
  if (typeof window === 'undefined') return normalized
  volatileGuestOrderAuth = cloneGuestOrderAuth(normalized)
  writeStorage('sessionStorage', GUEST_ORDER_AUTH_STORAGE_KEY, JSON.stringify(normalized))
  // 不回退到 localStorage，避免把游客订单凭据重新变成长生命周期数据。
  removeStorage('localStorage', GUEST_ORDER_AUTH_STORAGE_KEY)
  return normalized
}

export const clearGuestOrderAuth = (storage?: Storage) => {
  if (storage) {
    storage.removeItem(GUEST_ORDER_AUTH_STORAGE_KEY)
    return
  }
  volatileGuestOrderAuth = null
  if (typeof window === 'undefined') return
  removeStorage('sessionStorage', GUEST_ORDER_AUTH_STORAGE_KEY)
  removeStorage('localStorage', GUEST_ORDER_AUTH_STORAGE_KEY)
}

export const buildGuestOrderAuthParams = (auth: Partial<GuestOrderAuth>) => {
  const normalized = normalizeGuestOrderAuth(auth)
  return {
    phone: normalized.phone,
    ...(normalized.email ? { email: normalized.email } : {}),
    order_password: normalized.order_password,
  }
}
