import { api } from './client'

export const addressAPI = {
  provinces: () => api.get('/address/provinces'),
  cities: (provinceCode: string) => api.get('/address/cities', { params: { province_code: provinceCode } }),
  districts: (cityCode: string) => api.get('/address/districts', { params: { city_code: cityCode } }),
  townships: (districtCode: string) => api.get('/address/townships', { params: { district_code: districtCode } }),
}
