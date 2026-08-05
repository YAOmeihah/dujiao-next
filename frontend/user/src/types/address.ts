export interface AddressDivisionOption {
  code: string
  name: string
  province_code?: string
  city_code?: string
  district_code?: string
  township_code?: string
}

export interface ShippingAddressFormValue {
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
}
