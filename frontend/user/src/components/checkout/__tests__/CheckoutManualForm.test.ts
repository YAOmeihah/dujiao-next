import { mount } from '@vue/test-utils'
import { createHead } from '@unhead/vue/client'
import { createPinia } from 'pinia'
import { describe, expect, it } from 'vitest'
import CheckoutManualForm from '../CheckoutManualForm.vue'
import i18n from '../../../i18n'

const manualFormProducts = [
  {
    itemKey: 'product-1',
    productId: 1,
    title: {
      'zh-CN': '测试商品',
      'en-US': 'Demo Product',
    },
    fields: [
      {
        key: 'contact',
        type: 'text',
        required: true,
        options: [],
      },
    ],
    skuCount: 1,
  },
]

describe('CheckoutManualForm', () => {
  it('hides the outer heading in embedded mode while keeping fields editable', async () => {
    const wrapper = mount(CheckoutManualForm, {
      props: {
        manualFormProducts,
        modelValue: {},
        submitAttempted: false,
        embedded: true,
        getManualFieldLabel: () => '联系人',
        getManualFieldPlaceholder: () => '请输入联系人',
        manualFieldError: () => '',
      },
      global: {
        plugins: [createHead(), createPinia(), i18n],
      },
    })

    expect(wrapper.find('h2').exists()).toBe(false)
    expect(wrapper.text()).toContain('联系人')

    await wrapper.get('input').setValue('Alice')

    expect(wrapper.emitted('update:modelValue')?.[0]).toEqual([
      {
        'product-1': {
          contact: 'Alice',
        },
      },
    ])
  })
})
