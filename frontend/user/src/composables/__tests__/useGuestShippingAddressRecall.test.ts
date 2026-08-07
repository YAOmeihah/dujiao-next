import { describe, expect, it } from 'vitest'
import {
  GUEST_SHIPPING_ADDRESS_MAX_AGE_MS,
  GUEST_SHIPPING_ADDRESS_STORAGE_KEY,
  clearGuestShippingAddressRecall,
  createGuestShippingAddressRecallRecord,
  loadGuestShippingAddressRecall,
  saveGuestShippingAddressRecall,
  shouldEnableGuestShippingAddressRecall,
  type GuestShippingAddressRecallRecord,
} from '../useGuestShippingAddressRecall'

const record: GuestShippingAddressRecallRecord = {
  receiver_name: '张三',
  receiver_phone: '13800138000',
  province: '上海市',
  province_code: '310000',
  city: '上海市',
  city_code: '310100',
  district: '浦东新区',
  district_code: '310115',
  township: '花木街道',
  township_code: '310115013',
  detail_address: '世纪大道 100 号',
  saved_at: new Date('2026-04-23T12:00:00.000Z').toISOString(),
}

const TEST_NOW_MS = new Date('2026-04-24T00:00:00.000Z').getTime()

describe('guest shipping address recall storage', () => {
  it('loads a valid record from localStorage', () => {
    localStorage.setItem(GUEST_SHIPPING_ADDRESS_STORAGE_KEY, JSON.stringify(record))

    expect(loadGuestShippingAddressRecall(localStorage, TEST_NOW_MS)).toEqual(record)
  })

  it('returns null when the saved record is older than 30 days', () => {
    localStorage.setItem(GUEST_SHIPPING_ADDRESS_STORAGE_KEY, JSON.stringify({
      ...record,
      saved_at: new Date(TEST_NOW_MS - GUEST_SHIPPING_ADDRESS_MAX_AGE_MS - 1000).toISOString(),
    }))

    expect(loadGuestShippingAddressRecall(localStorage, TEST_NOW_MS)).toBeNull()
  })

  it('returns null when required address fields are missing', () => {
    localStorage.setItem(GUEST_SHIPPING_ADDRESS_STORAGE_KEY, JSON.stringify({
      ...record,
      township_code: '',
    }))

    expect(loadGuestShippingAddressRecall(localStorage, TEST_NOW_MS)).toBeNull()
  })

  it('returns null when browser storage cannot be read', () => {
    const storage = {
      getItem: () => { throw new Error('storage blocked') },
    } as unknown as Storage

    expect(loadGuestShippingAddressRecall(storage, TEST_NOW_MS)).toBeNull()
  })

  it('removes malformed and expired records when browser storage allows cleanup', () => {
    const removedKeys: string[] = []
    const storage = {
      getItem: () => JSON.stringify({ ...record, saved_at: 'not-a-date' }),
      removeItem: (key: string) => { removedKeys.push(key) },
    } as unknown as Storage

    expect(loadGuestShippingAddressRecall(storage, TEST_NOW_MS)).toBeNull()
    expect(removedKeys).toEqual([GUEST_SHIPPING_ADDRESS_STORAGE_KEY])
  })

  it('writes a normalized record back to localStorage', () => {
    saveGuestShippingAddressRecall({
      ...record,
      receiver_name: ' 张三 ',
      detail_address: ' 世纪大道 100 号 ',
    })

    expect(loadGuestShippingAddressRecall(localStorage, TEST_NOW_MS)).toEqual({
      ...record,
      receiver_name: '张三',
      detail_address: '世纪大道 100 号',
    })
  })

  it('creates and persists a recent record from a confirmed shipping payload before order submission', () => {
    const recentRecord = createGuestShippingAddressRecallRecord(
      {
        receiver_name: ' 李四 ',
        receiver_phone: ' 13900139000 ',
        province: ' 浙江省 ',
        province_code: '330000',
        city: ' 杭州市 ',
        city_code: '330100',
        district: ' 西湖区 ',
        district_code: '330106',
        township: ' 古荡街道 ',
        township_code: '330106009',
        detail_address: ' 文三路 90 号 ',
      },
      {
        now: '2026-04-23T13:00:00.000Z',
      },
    )

    expect(recentRecord).toEqual({
      receiver_name: '李四',
      receiver_phone: '13900139000',
      province: '浙江省',
      province_code: '330000',
      city: '杭州市',
      city_code: '330100',
      district: '西湖区',
      district_code: '330106',
      township: '古荡街道',
      township_code: '330106009',
      detail_address: '文三路 90 号',
      saved_at: '2026-04-23T13:00:00.000Z',
    })
    expect(loadGuestShippingAddressRecall(localStorage, TEST_NOW_MS)).toEqual(recentRecord)
  })

  it('returns the recent record when browser storage rejects the write', () => {
    const storage = {
      setItem: () => { throw new Error('quota exceeded') },
    } as unknown as Storage

    expect(() => createGuestShippingAddressRecallRecord({
      receiver_name: record.receiver_name,
      receiver_phone: record.receiver_phone,
      province: record.province,
      province_code: record.province_code,
      city: record.city,
      city_code: record.city_code,
      district: record.district,
      district_code: record.district_code,
      township: record.township,
      township_code: record.township_code,
      detail_address: record.detail_address,
    }, { storage })).not.toThrow()
  })

  it('clears the saved record', () => {
    localStorage.setItem(GUEST_SHIPPING_ADDRESS_STORAGE_KEY, JSON.stringify(record))
    clearGuestShippingAddressRecall()
    expect(localStorage.getItem(GUEST_SHIPPING_ADDRESS_STORAGE_KEY)).toBeNull()
  })

  it('does not throw when browser storage rejects cleanup', () => {
    const storage = {
      removeItem: () => { throw new Error('storage blocked') },
    } as unknown as Storage

    expect(() => clearGuestShippingAddressRecall(storage)).not.toThrow()
  })
})

describe('shouldEnableGuestShippingAddressRecall', () => {
  it('only enables recall for guest checkout with shipping-required orders', () => {
    expect(shouldEnableGuestShippingAddressRecall({
      orderRequiresShippingAddress: true,
      isAuthenticated: false,
      checkoutMode: 'guest',
    })).toBe(true)

    expect(shouldEnableGuestShippingAddressRecall({
      orderRequiresShippingAddress: false,
      isAuthenticated: false,
      checkoutMode: 'guest',
    })).toBe(false)

    expect(shouldEnableGuestShippingAddressRecall({
      orderRequiresShippingAddress: true,
      isAuthenticated: true,
      checkoutMode: 'guest',
    })).toBe(false)
  })
})
