import { mount } from '@vue/test-utils'
import { describe, expect, it } from 'vitest'
import RegionSelector from '../RegionSelector.vue'
import i18n from '../../../i18n'

const modelValue = {
  receiver_name: '',
  receiver_phone: '',
  province: '',
  province_code: '',
  city: '',
  city_code: '',
  district: '',
  district_code: '',
  township: '',
  township_code: '',
  detail_address: '',
}

describe('RegionSelector', () => {
  it('binds external tracking attributes onto the visible root container', () => {
    const wrapper = mount(RegionSelector, {
      props: {
        modelValue,
        invalid: false,
      },
      attrs: {
        'data-mobile-shipping-input': 'region',
      },
      global: {
        plugins: [i18n],
      },
    })

    expect(wrapper.find('[data-mobile-shipping-input="region"]').exists()).toBe(true)
  })
})
