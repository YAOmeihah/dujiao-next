import { createHead } from '@unhead/vue/client'
import { flushPromises, mount } from '@vue/test-utils'
import { createPinia } from 'pinia'
import { afterEach, describe, expect, it, vi } from 'vitest'
import i18n from '../../i18n'
import Support from '../Support.vue'

const mountSupport = async ({
  supportUrl = 'https://help.example.com/chat',
  active = true,
  layout = 'classic' as const,
}: {
  supportUrl?: unknown
  active?: boolean
  layout?: 'classic' | 'vault'
} = {}) => {
  const pinia = createPinia()
  pinia.state.value.app = {
    config: { contact: { support_url: supportUrl } },
    locale: 'zh-CN',
  }

  const wrapper = mount(Support, {
    props: { active, layout },
    global: {
      plugins: [createHead(), pinia, i18n],
    },
  })
  await flushPromises()
  return wrapper
}

afterEach(() => {
  vi.useRealTimers()
})

describe('Support', () => {
  it('renders a secure iframe for a valid support URL', async () => {
    const wrapper = await mountSupport()
    const iframe = wrapper.get('iframe')

    expect(wrapper.get('[data-support-page]').attributes('data-state')).toBe('loading')
    expect(iframe.attributes('src')).toBe('https://help.example.com/chat')
    expect(iframe.attributes('referrerpolicy')).toBe('strict-origin-when-cross-origin')
  })

  it('keeps the same iframe element while the page becomes inactive and active again', async () => {
    const wrapper = await mountSupport()
    const iframe = wrapper.get('iframe').element

    await wrapper.setProps({ active: false })
    await wrapper.setProps({ active: true })

    expect(wrapper.get('iframe').element).toBe(iframe)
  })

  it('distinguishes missing and invalid support configuration', async () => {
    const empty = await mountSupport({ supportUrl: '' })
    const invalid = await mountSupport({ supportUrl: 'javascript:alert(1)' })

    expect(empty.get('[data-support-page]').attributes('data-state')).toBe('empty')
    expect(empty.find('iframe').exists()).toBe(false)
    expect(invalid.get('[data-support-page]').attributes('data-state')).toBe('invalid')
    expect(invalid.find('iframe').exists()).toBe(false)
  })

  it('offers a safe external link when loading exceeds eight seconds', async () => {
    vi.useFakeTimers()
    const wrapper = await mountSupport()

    await vi.advanceTimersByTimeAsync(8000)

    const link = wrapper.get('[data-support-external]')
    expect(wrapper.get('[data-support-page]').attributes('data-state')).toBe('fallback')
    expect(link.attributes('href')).toBe('https://help.example.com/chat')
    expect(link.attributes('target')).toBe('_blank')
    expect(link.attributes('rel')).toBe('noopener noreferrer')
  })

  it('cancels the fallback after the iframe reports it loaded', async () => {
    vi.useFakeTimers()
    const wrapper = await mountSupport()

    await wrapper.get('iframe').trigger('load')
    await vi.advanceTimersByTimeAsync(8000)

    expect(wrapper.get('[data-support-page]').attributes('data-state')).toBe('ready')
    expect(wrapper.find('[data-support-external]').exists()).toBe(false)
  })

  it('uses layout-specific spacing without changing the state machine', async () => {
    const classic = await mountSupport({ layout: 'classic' })
    const vault = await mountSupport({ layout: 'vault' })

    expect(classic.get('[data-support-page]').classes()).toContain('pt-20')
    expect(vault.get('[data-support-page]').classes()).not.toContain('pt-20')
    expect(vault.get('[data-support-page]').classes()).toContain('py-2')
  })
})
