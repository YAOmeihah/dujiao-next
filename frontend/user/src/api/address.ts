import { api } from './client'

export const addressAPI = {
  provinces: () => api.get('/public/address/provinces'),
  cities: (provinceCode: string) => api.get('/public/address/cities', { params: { province_code: provinceCode } }),
  districts: (cityCode: string) => api.get('/public/address/districts', { params: { city_code: cityCode } }),
  townships: (districtCode: string) => api.get('/public/address/townships', { params: { district_code: districtCode } }),
}
