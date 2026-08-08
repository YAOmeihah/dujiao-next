import { createHead } from '@unhead/vue/client'
import { mount } from '@vue/test-utils'
import { createPinia } from 'pinia'
import { defineComponent, h } from 'vue'
import { describe, expect, it } from 'vitest'
import i18n from '../../i18n'
import { useNavConfig } from '../useNavConfig'

const NavHarness = defineComponent({
  setup() {
    const { primaryNavItems, supportEnabled } = useNavConfig()
    return () => h('div', {
      'data-support-enabled': String(supportEnabled.value),
    }, primaryNavItems.value.map((item) => h('span', { 'data-key': item.key }, item.label)))
  },
})

const mountWithSupportUrl = (supportUrl: unknown, builtin: Record<string, boolean> = {}) => {
  const pinia = createPinia()
  pinia.state.value.app = {
    config: {
      contact: { support_url: supportUrl },
      nav_config: { builtin },
    },
    locale: 'zh-CN',
  }

  return mount(NavHarness, {
    global: {
      plugins: [createHead(), pinia, i18n],
    },
  })
}

describe('useNavConfig support item', () => {
  it('adds the support route when a valid support URL is configured', () => {
    const wrapper = mountWithSupportUrl('https://help.example.com/chat')

    expect(wrapper.attributes('data-support-enabled')).toBe('true')
    expect(wrapper.find('[data-key="support"]').exists()).toBe(true)
    expect(wrapper.find('[data-key="support"]').text()).toBe('客服')
  })

  it.each([
    ['', 'empty'],
    ['/support', 'relative'],
    ['javascript:alert(1)', 'non-http'],
  ])('hides the support route for %s (%s)', (supportUrl) => {
    const wrapper = mountWithSupportUrl(supportUrl)

    expect(wrapper.attributes('data-support-enabled')).toBe('false')
    expect(wrapper.find('[data-key="support"]').exists()).toBe(false)
  })

  it('uses the configured URL as the single visibility rule for every support entry', () => {
    const wrapper = mountWithSupportUrl('https://help.example.com/chat', { support: false })

    expect(wrapper.attributes('data-support-enabled')).toBe('true')
    expect(wrapper.find('[data-key="support"]').exists()).toBe(true)
  })
})
