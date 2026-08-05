import { mount } from '@vue/test-utils'
import { createHead } from '@unhead/vue/client'
import { createPinia, setActivePinia } from 'pinia'
import { describe, expect, it, vi } from 'vitest'
import Footer from '../Footer.vue'
import i18n from '../../i18n'

const routeState = vi.hoisted(() => ({ name: 'checkout' as string | undefined }))

vi.mock('vue-router', () => ({
  useRoute: () => ({ name: routeState.name }),
}))

describe('Footer', () => {
  it('keeps mobile checkout footer content above the floating checkout bar', () => {
    const pinia = createPinia()
    setActivePinia(pinia)
    pinia.state.value.app = {
      config: {
        brand: {
          site_name: 'Dujiao-Next',
        },
      },
    }

    const wrapper = mount(Footer, {
      global: {
        plugins: [createHead(), pinia, i18n],
        stubs: {
          RouterLink: true,
        },
      },
    })

    expect(wrapper.classes()).toEqual(expect.arrayContaining(['pb-28', 'lg:pb-0']))
  })
})
