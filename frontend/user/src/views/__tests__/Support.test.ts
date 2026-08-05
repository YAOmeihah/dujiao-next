import { mount } from '@vue/test-utils'
import { createHead } from '@unhead/vue/client'
import { createPinia } from 'pinia'
import { describe, expect, it } from 'vitest'
import Support from '../Support.vue'
import i18n from '../../i18n'

describe('Support', () => {
  it('keeps responsive bottom breathing room around the embedded support iframe', () => {
    const pinia = createPinia()
    pinia.state.value.app = {
      config: {
        contact: {
          support_url: 'https://example.com/support',
        },
      },
    }

    const wrapper = mount(Support, {
      props: {
        active: true,
      },
      global: {
        plugins: [createHead(), pinia, i18n],
      },
    })

    expect(wrapper.classes()).toEqual(expect.arrayContaining(['pb-2', 'sm:pb-4', 'lg:pb-6']))
  })
})
