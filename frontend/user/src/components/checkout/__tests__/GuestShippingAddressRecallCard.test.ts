import { mount } from '@vue/test-utils'
import { describe, expect, it } from 'vitest'
import GuestShippingAddressRecallCard from '../GuestShippingAddressRecallCard.vue'

const props = {
  summaryLines: ['张三 · 138****8888', '上海市 浦东新区 花木街道 世纪大道 100 号'],
  title: '已找到上次收货信息',
  useLabel: '使用上次地址',
  rewriteLabel: '重新填写',
  appliedMessage: '已带入上次收货信息，可继续修改',
  clearFormLabel: '清空重填',
  clearRecordLabel: '清除记录',
}

describe('GuestShippingAddressRecallCard', () => {
  it('renders the prompt state with summary lines and primary actions', () => {
    const wrapper = mount(GuestShippingAddressRecallCard, {
      props: {
        ...props,
        applied: false,
        muted: false,
      },
    })

    expect(wrapper.text()).toContain('已找到上次收货信息')
    expect(wrapper.text()).toContain('张三 · 138****8888')
    expect(wrapper.text()).toContain('使用上次地址')
    expect(wrapper.text()).toContain('重新填写')
  })

  it('renders the applied state and emits clear-form when requested', async () => {
    const wrapper = mount(GuestShippingAddressRecallCard, {
      props: {
        ...props,
        applied: true,
        muted: true,
      },
    })

    expect(wrapper.text()).toContain('已带入上次收货信息，可继续修改')
    expect(wrapper.text()).toContain('清空重填')

    await wrapper.get('[data-recall-clear-form]').trigger('click')
    expect(wrapper.emitted('clear-form')).toHaveLength(1)
  })
})
