import { describe, expect, it } from 'vitest'
import {
  buildGuestOrderAuthParams,
  clearGuestOrderAuth,
  createEmptyGuestOrderAuth,
  hasGuestOrderAuth,
  loadGuestOrderAuth,
  normalizeGuestOrderAuth,
  saveGuestOrderAuth,
} from '../guestOrderAuth'

const createMemoryStorage = (): Storage => {
  const data = new Map<string, string>()
  return {
    get length() {
      return data.size
    },
    clear: () => data.clear(),
    getItem: (key: string) => data.get(key) ?? null,
    key: (index: number) => Array.from(data.keys())[index] ?? null,
    removeItem: (key: string) => {
      data.delete(key)
    },
    setItem: (key: string, value: string) => {
      data.set(key, value)
    },
  }
}

describe('guestOrderAuth', () => {
  it('normalizes phone, optional email, and order password', () => {
    expect(normalizeGuestOrderAuth({
      phone: ' 13800138000 ',
      email: ' buyer@example.com ',
      order_password: '  keep-space-inside ',
    })).toEqual({
      phone: '13800138000',
      email: 'buyer@example.com',
      order_password: '  keep-space-inside ',
    })
  })

  it('requires phone and order password for guest auth', () => {
    expect(hasGuestOrderAuth({ phone: '13800138000', email: '', order_password: 'pw' })).toBe(true)
    expect(hasGuestOrderAuth({ phone: '', email: 'buyer@example.com', order_password: 'pw' })).toBe(false)
    expect(hasGuestOrderAuth({ phone: '13800138000', email: '', order_password: '' })).toBe(false)
  })

  it('persists and loads the fork auth shape', () => {
    const storage = createMemoryStorage()
    const saved = saveGuestOrderAuth({
      phone: ' 13800138000 ',
      email: '',
      order_password: 'pw',
    }, storage)

    expect(saved).toEqual({
      phone: '13800138000',
      email: '',
      order_password: 'pw',
    })
    expect(loadGuestOrderAuth(storage)).toEqual(saved)
  })

  it('ignores legacy email-only auth for access checks', () => {
    const storage = createMemoryStorage()
    storage.setItem('guest_order_auth', JSON.stringify({
      email: 'buyer@example.com',
      order_password: 'pw',
    }))

    const loaded = loadGuestOrderAuth(storage)
    expect(loaded).toEqual({
      phone: '',
      email: 'buyer@example.com',
      order_password: 'pw',
    })
    expect(hasGuestOrderAuth(loaded)).toBe(false)
  })

  it('builds backend params with required phone and optional email', () => {
    expect(buildGuestOrderAuthParams({
      phone: '13800138000',
      email: '',
      order_password: 'pw',
    })).toEqual({
      phone: '13800138000',
      order_password: 'pw',
    })
    expect(buildGuestOrderAuthParams({
      phone: '13800138000',
      email: 'buyer@example.com',
      order_password: 'pw',
    })).toEqual({
      phone: '13800138000',
      email: 'buyer@example.com',
      order_password: 'pw',
    })
  })

  it('clears stored auth and returns an empty shape for invalid JSON', () => {
    const storage = createMemoryStorage()
    storage.setItem('guest_order_auth', '{')
    expect(loadGuestOrderAuth(storage)).toEqual(createEmptyGuestOrderAuth())
    saveGuestOrderAuth({ phone: '13800138000', email: '', order_password: 'pw' }, storage)
    clearGuestOrderAuth(storage)
    expect(storage.getItem('guest_order_auth')).toBeNull()
  })
})
