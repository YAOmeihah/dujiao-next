import { beforeEach, describe, expect, it, vi } from 'vitest'
import { addressAPI } from '../address'

const { apiGet } = vi.hoisted(() => ({
  apiGet: vi.fn(),
}))

vi.mock('../client', () => ({
  api: {
    get: apiGet,
  },
}))

describe('addressAPI', () => {
  beforeEach(() => {
    apiGet.mockClear()
  })

  it('uses the public address division routes', () => {
    addressAPI.provinces()
    addressAPI.cities('33')
    addressAPI.districts('3301')
    addressAPI.townships('330106')

    expect(apiGet).toHaveBeenNthCalledWith(1, '/public/address/provinces')
    expect(apiGet).toHaveBeenNthCalledWith(2, '/public/address/cities', { params: { province_code: '33' } })
    expect(apiGet).toHaveBeenNthCalledWith(3, '/public/address/districts', { params: { city_code: '3301' } })
    expect(apiGet).toHaveBeenNthCalledWith(4, '/public/address/townships', { params: { district_code: '330106' } })
  })
})
