import { mount } from '@vue/test-utils'
import { describe, expect, it } from 'vitest'
import { nextTick } from 'vue'

import CapCaptcha from '../CapCaptcha.vue'

describe('CapCaptcha', () => {
  it('renders the Cap widget endpoint and emits solved tokens', async () => {
    const wrapper = mount(CapCaptcha, {
      props: {
        endpoint: ' http://127.0.0.1:3001/ ',
        siteKey: ' /site-key/ ',
        modelValue: '',
      },
    })

    const widget = wrapper.get('cap-widget')
    expect(widget.attributes('data-cap-api-endpoint')).toBe('http://127.0.0.1:3001/site-key/')
    expect(widget.attributes()).not.toHaveProperty('data-cap-i18n-initial-state')
    expect(Array.from(document.scripts).some((script) => script.src === 'http://127.0.0.1:3001/assets/widget.js')).toBe(true)

    widget.element.dispatchEvent(new CustomEvent('solve', { detail: { token: 'cap-token' }, bubbles: true }))
    await nextTick()

    expect(wrapper.emitted('update:modelValue')).toContainEqual(['cap-token'])

    await wrapper.vm.reset()
    expect(wrapper.emitted('update:modelValue')).toContainEqual([''])
  })

  it('does not render a widget until endpoint and site key are present', () => {
    const wrapper = mount(CapCaptcha, {
      props: {
        endpoint: 'http://127.0.0.1:3001',
        siteKey: '',
      },
    })

    expect(wrapper.find('cap-widget').exists()).toBe(false)
  })
})
