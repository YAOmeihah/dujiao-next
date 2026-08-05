import { mount } from '@vue/test-utils'
import { describe, expect, it } from 'vitest'
import MobileCheckoutFlow from '../MobileCheckoutFlow.vue'

const sections = [
  {
    key: 'items',
    title: '商品摘要',
    badge: '',
    summaryLines: ['共 1 件商品', '应付 ¥2.00'],
    errorMessage: '',
    collapsedActionLabel: '查看明细',
    complete: true,
    recommended: false,
    softHint: '',
  },
  {
    key: 'shipping',
    title: '收货信息',
    badge: '当前任务',
    summaryLines: ['请填写姓名、手机号和地址'],
    errorMessage: '请填写完整收货地址',
    complete: false,
    recommended: true,
    softHint: '',
  },
]

describe('MobileCheckoutFlow', () => {
  it('shows completed sections collapsed with summary text', () => {
    const wrapper = mount(MobileCheckoutFlow, {
      props: {
        sections,
        expandedSection: 'shipping',
        statusText: '请先填写收货信息',
        totalText: '¥2.00',
        primaryActionLabel: '保存收货信息',
        primaryActionDisabled: false,
      },
      slots: {
        'section-items': '<div data-testid="items-slot">items slot</div>',
        'section-shipping': '<div data-testid="shipping-slot">shipping slot</div>',
      },
    })

    expect(wrapper.text()).toContain('共 1 件商品')
    expect(wrapper.find('[data-testid="items-slot"]').exists()).toBe(false)
    expect(wrapper.find('[data-testid="shipping-slot"]').exists()).toBe(true)
    expect(wrapper.find('[data-section-badge="items"]').exists()).toBe(false)
    expect(wrapper.get('[data-section-root="shipping"]').attributes('data-section-state')).toBe('current-editing')
    expect(wrapper.get('[data-section-error="shipping"]').text()).toContain('请填写完整收货地址')
    expect(wrapper.get('[data-section-toggle="items"]').text()).toContain('查看明细')
  })

  it('keeps the breathing highlight only on the collapsed recommended section', () => {
    const wrapper = mount(MobileCheckoutFlow, {
      props: {
        sections,
        expandedSection: 'items',
        statusText: '请先填写收货信息',
        totalText: '¥2.00',
        primaryActionLabel: '保存收货信息',
        primaryActionDisabled: false,
      },
      slots: {
        'section-items': '<div data-testid="items-slot">items slot</div>',
        'section-shipping': '<div data-testid="shipping-slot">shipping slot</div>',
      },
    })

    expect(wrapper.find('[data-testid="shipping-slot"]').exists()).toBe(false)
    expect(wrapper.get('[data-section-root="shipping"]').attributes('data-section-state')).toBe('current')
  })

  it('emits section changes and primary action clicks', async () => {
    const wrapper = mount(MobileCheckoutFlow, {
      props: {
        sections,
        expandedSection: 'shipping',
        statusText: '请先填写收货信息',
        totalText: '¥2.00',
        primaryActionLabel: '保存收货信息',
        primaryActionDisabled: false,
      },
      slots: {
        'section-items': '<div data-testid="items-slot">items slot</div>',
        'section-shipping': '<div data-testid="shipping-slot">shipping slot</div>',
      },
    })

    await wrapper.get('[data-section-toggle="items"]').trigger('click')
    await wrapper.get('[data-mobile-primary-action]').trigger('click')

    expect(wrapper.emitted('update:expandedSection')?.[0]).toEqual(['items'])
    expect(wrapper.emitted('primaryAction')).toHaveLength(1)
  })

  it('emits null when the user clicks the already expanded section to collapse it', async () => {
    const wrapper = mount(MobileCheckoutFlow, {
      props: {
        sections,
        expandedSection: 'shipping',
        statusText: '请先填写收货信息',
        totalText: '¥2.00',
        primaryActionLabel: '保存收货信息',
        primaryActionDisabled: false,
      },
      slots: {
        'section-items': '<div data-testid="items-slot">items slot</div>',
        'section-shipping': '<div data-testid="shipping-slot">shipping slot</div>',
      },
    })

    await wrapper.get('[data-section-toggle="shipping"]').trigger('click')

    expect(wrapper.emitted('update:expandedSection')?.[0]).toEqual([null])
  })
})
