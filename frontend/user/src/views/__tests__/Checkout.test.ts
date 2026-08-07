import { mount } from '@vue/test-utils'
import { createHead } from '@unhead/vue/client'
import { createPinia, setActivePinia } from 'pinia'
import { nextTick } from 'vue'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import Checkout from '../Checkout.vue'
import VaultCheckout from '../../templates/vault/Checkout.vue'
import i18n from '../../i18n'
import { guestOrderAPI, userOrderAPI, walletAPI } from '../../api'

const routerPush = vi.fn()

vi.mock('vue-router', () => ({
  useRoute: () => ({
    query: {
      mode: 'buynow',
    },
  }),
  useRouter: () => ({
    push: routerPush,
  }),
}))

vi.mock('../../api', () => ({
  configAPI: {
    get: vi.fn(() => Promise.resolve({ data: { data: {} } })),
  },
  guestOrderAPI: {
    preview: vi.fn(),
    createAndPay: vi.fn(),
  },
  userAuthAPI: {},
  userOrderAPI: {
    preview: vi.fn(() => Promise.resolve({
      data: {
        data: {
          total_amount: '10.00',
          currency: 'CNY',
        },
      },
    })),
    getPaymentChannels: vi.fn(() => Promise.resolve({
      data: {
        data: [
          {
            id: 1,
            name: '支付宝',
            provider_type: 'alipay',
            channel_type: 'alipay',
          },
        ],
      },
    })),
    createAndPay: vi.fn(() => Promise.resolve({
      data: {
        data: {
          order_no: 'ORDER-1001',
        },
      },
    })),
  },
  walletAPI: {
    account: vi.fn(() => Promise.resolve({
      data: {
        data: {
          balance: '0.00',
        },
      },
    })),
  },
}))

vi.mock('../../utils/cartStock', async () => {
  const actual = await vi.importActual<typeof import('../../utils/cartStock')>('../../utils/cartStock')
  return {
    ...actual,
    refreshCartStockSnapshots: vi.fn(() => Promise.resolve()),
  }
})

vi.mock('../../utils/debounce', () => ({
  debounceAsync: (fn: (...args: any[]) => Promise<any>) =>
    Object.assign((...args: any[]) => fn(...args), { cancel: () => undefined }),
}))

const flushPromises = () => new Promise((resolve) => window.setTimeout(resolve, 0))

describe('Checkout mobile buy-now flow', () => {
  beforeEach(() => {
    routerPush.mockClear()
    vi.clearAllMocks()
    localStorage.setItem('user_token', 'member-token')
    window.requestAnimationFrame = (callback: FrameRequestCallback) => {
      callback(0)
      return 0
    }
    window.scrollTo = vi.fn()
  })

  it('loads the fork mobile checkout translations', () => {
    expect(i18n.global.t('checkout.mobile.actionContinueBuyer')).toBe('确认购买信息并继续')
  })

  it('renders the shared mobile checkout flow in the vault template', async () => {
    const pinia = createPinia()
    setActivePinia(pinia)
    pinia.state.value.app = {
      config: {
        wallet_only_payment: false,
        payment_channels: [],
      },
    }
    pinia.state.value['user-auth'] = {
      token: 'member-token',
      user: null,
      loading: false,
      challengeToken: '',
      challengeExpiresAt: '',
    }
    pinia.state.value.buyNow = {
      item: {
        productId: 1,
        slug: 'test-product',
        title: { 'zh-CN': '测试商品' },
        priceAmount: '10.00',
        quantity: 1,
        requiresShippingAddress: false,
        paymentChannelIds: [],
      },
    }

    const wrapper = mount(VaultCheckout, {
      global: {
        plugins: [createHead(), pinia, i18n],
        stubs: {
          RouterLink: true,
          VaultCheckoutSteps: true,
          SmartImage: true,
          CheckoutManualForm: true,
          GuestShippingAddressRecallCard: true,
          RegionSelector: true,
          ImageCaptcha: true,
          TurnstileCaptcha: true,
          CapCaptcha: true,
        },
      },
    })

    await nextTick()
    await flushPromises()
    await nextTick()

    expect(wrapper.get('[data-mobile-primary-action]').text()).toBe(
      i18n.global.t('checkout.mobile.actionContinueBuyer'),
    )

    wrapper.unmount()
  })

  it('submits the order when confirming a ready payment method', async () => {
    const pinia = createPinia()
    setActivePinia(pinia)
    pinia.state.value.app = {
      config: {
        wallet_only_payment: false,
        payment_channels: [
          {
            id: 1,
            name: '支付宝',
            provider_type: 'alipay',
            channel_type: 'alipay',
          },
        ],
      },
    }
    pinia.state.value['user-auth'] = {
      token: 'member-token',
      user: null,
      loading: false,
      challengeToken: '',
      challengeExpiresAt: '',
    }
    pinia.state.value.buyNow = {
      item: {
        productId: 1,
        slug: 'test-product',
        title: {
          'zh-CN': '测试商品',
        },
        priceAmount: '10.00',
        quantity: 1,
        requiresShippingAddress: false,
        paymentChannelIds: [1],
      },
    }

    const wrapper = mount(Checkout, {
      global: {
        plugins: [createHead(), pinia, i18n],
        stubs: {
          RouterLink: true,
          CheckoutSteps: true,
          EmptyState: true,
          SmartImage: true,
          CheckoutManualForm: true,
          GuestShippingAddressRecallCard: true,
          RegionSelector: true,
          ImageCaptcha: true,
          TurnstileCaptcha: true,
          CapCaptcha: true,
        },
      },
    })

    await nextTick()
    await flushPromises()
    await nextTick()

    expect(userOrderAPI.getPaymentChannels).toHaveBeenCalled()
    expect(walletAPI.account).toHaveBeenCalled()
    expect(wrapper.text()).toContain('支付宝')
    const desktopGrid = wrapper.get('.hidden.grid-cols-1.gap-8')
    expect(desktopGrid.text()).toContain(i18n.global.t('checkout.itemsTitle'))
    expect(desktopGrid.text()).toContain(i18n.global.t('checkout.submitTitle'))
    expect(wrapper.get('[data-mobile-primary-action]').text()).toBe(i18n.global.t('checkout.mobile.actionContinueBuyer'))

    await wrapper.get('[data-mobile-primary-action]').trigger('click')
    await flushPromises()
    await nextTick()
    expect(wrapper.get('[data-mobile-primary-action]').text()).toBe(i18n.global.t('checkout.mobile.actionChoosePayment'))

    const channelButton = wrapper.get('[data-mobile-payment-input="channel-list"] button')
    expect(channelButton.text()).toContain('支付宝')
    await channelButton.trigger('click')
    await nextTick()
    expect(wrapper.get('[data-mobile-payment-input="channel-list"] button').classes()).toContain('theme-selected-surface')
    expect(wrapper.text()).toContain(i18n.global.t('checkout.mobile.readyToSubmit'))
    expect(wrapper.get('[data-mobile-primary-action]').text()).toBe(i18n.global.t('checkout.mobile.actionSubmit'))

    await wrapper.get('[data-mobile-primary-action]').trigger('click')
    await flushPromises()
    await nextTick()
    await flushPromises()
    await nextTick()

    expect(userOrderAPI.preview).toHaveBeenCalled()
    expect(guestOrderAPI.createAndPay).not.toHaveBeenCalled()
    expect(userOrderAPI.createAndPay).toHaveBeenCalledTimes(1)
    expect(userOrderAPI.createAndPay).toHaveBeenCalledWith(expect.objectContaining({
      channel_id: 1,
      use_balance: false,
    }))
    expect(routerPush).toHaveBeenCalledWith('/pay?order_no=ORDER-1001')

    wrapper.unmount()
    expect(walletAPI.account).toHaveBeenCalled()
  })

  it('sends the entered guest phone number when creating an order from the desktop checkout', async () => {
    localStorage.removeItem('user_token')
    const pinia = createPinia()
    setActivePinia(pinia)
    pinia.state.value.app = {
      config: {
        wallet_only_payment: false,
        payment_channels: [
          {
            id: 1,
            name: '支付宝',
            provider_type: 'alipay',
            channel_type: 'alipay',
          },
        ],
      },
    }
    pinia.state.value['user-auth'] = {
      token: '',
      user: null,
      loading: false,
      challengeToken: '',
      challengeExpiresAt: '',
    }
    pinia.state.value.buyNow = {
      item: {
        productId: 1,
        slug: 'test-product',
        title: {
          'zh-CN': '测试商品',
        },
        priceAmount: '10.00',
        quantity: 1,
        requiresShippingAddress: false,
        paymentChannelIds: [1],
      },
    }
    vi.mocked(guestOrderAPI.preview).mockResolvedValue({
      data: {
        data: {
          total_amount: '10.00',
          currency: 'CNY',
        },
      },
    })
    vi.mocked(guestOrderAPI.createAndPay).mockResolvedValue({
      data: {
        data: {
          order_no: 'GUEST-1001',
        },
      },
    })

    const wrapper = mount(Checkout, {
      global: {
        plugins: [createHead(), pinia, i18n],
        stubs: {
          RouterLink: true,
          CheckoutSteps: true,
          EmptyState: true,
          SmartImage: true,
          CheckoutManualForm: true,
          GuestShippingAddressRecallCard: true,
          RegionSelector: true,
          ImageCaptcha: true,
          TurnstileCaptcha: true,
          CapCaptcha: true,
        },
      },
    })

    await nextTick()
    await flushPromises()
    const desktopGrid = wrapper.get('.hidden.grid-cols-1.gap-8')
    await desktopGrid.get('input[autocomplete="tel"]').setValue('13800138000')
    await desktopGrid.get('input[type="password"]').setValue('guest-password')
    await desktopGrid.get('.grid.grid-cols-2.gap-2 button').trigger('click')
    await nextTick()

    const submitButton = desktopGrid.findAll('button').find((button) => button.text() === i18n.global.t('checkout.submitButton'))
    expect(submitButton).toBeDefined()
    await submitButton!.trigger('click')
    await flushPromises()
    await nextTick()

    expect(guestOrderAPI.createAndPay).toHaveBeenCalledWith(expect.objectContaining({
      phone: '13800138000',
      order_password: 'guest-password',
      channel_id: 1,
    }))
  })

  it('blocks desktop guest checkout when the phone contains no digits', async () => {
    localStorage.removeItem('user_token')
    const pinia = createPinia()
    setActivePinia(pinia)
    pinia.state.value.app = {
      config: {
        wallet_only_payment: false,
        payment_channels: [
          {
            id: 1,
            name: '支付宝',
            provider_type: 'alipay',
            channel_type: 'alipay',
          },
        ],
      },
    }
    pinia.state.value['user-auth'] = {
      token: '',
      user: null,
      loading: false,
      challengeToken: '',
      challengeExpiresAt: '',
    }
    pinia.state.value.buyNow = {
      item: {
        productId: 1,
        slug: 'test-product',
        title: { 'zh-CN': '测试商品' },
        priceAmount: '10.00',
        quantity: 1,
        requiresShippingAddress: false,
        paymentChannelIds: [1],
      },
    }
    vi.mocked(guestOrderAPI.preview).mockResolvedValue({
      data: { data: { total_amount: '10.00', currency: 'CNY' } },
    })
    vi.mocked(guestOrderAPI.createAndPay).mockResolvedValue({
      data: { data: { order_no: 'GUEST-INVALID-PHONE' } },
    })

    const wrapper = mount(Checkout, {
      global: {
        plugins: [createHead(), pinia, i18n],
        stubs: {
          RouterLink: true,
          CheckoutSteps: true,
          EmptyState: true,
          SmartImage: true,
          CheckoutManualForm: true,
          GuestShippingAddressRecallCard: true,
          RegionSelector: true,
          ImageCaptcha: true,
          TurnstileCaptcha: true,
          CapCaptcha: true,
        },
      },
    })

    await nextTick()
    await flushPromises()
    const desktopGrid = wrapper.get('.hidden.grid-cols-1.gap-8')
    await desktopGrid.get('input[autocomplete="tel"]').setValue('------')
    await desktopGrid.get('input[type="password"]').setValue('guest-password')
    await desktopGrid.get('.grid.grid-cols-2.gap-2 button').trigger('click')
    await nextTick()

    const submitButton = desktopGrid.findAll('button').find((button) => button.text() === i18n.global.t('checkout.submitButton'))
    expect(submitButton).toBeDefined()
    await submitButton!.trigger('click')
    await flushPromises()
    await nextTick()

    expect(guestOrderAPI.createAndPay).not.toHaveBeenCalled()
    expect(desktopGrid.text()).toContain(i18n.global.t('error.phone_invalid'))
  })

  it('normalizes the shipping phone before submitting an address order', async () => {
    const pinia = createPinia()
    setActivePinia(pinia)
    pinia.state.value.app = {
      config: {
        wallet_only_payment: false,
        payment_channels: [
          {
            id: 1,
            name: '支付宝',
            provider_type: 'alipay',
            channel_type: 'alipay',
          },
        ],
      },
    }
    pinia.state.value['user-auth'] = {
      token: 'member-token',
      user: null,
      loading: false,
      challengeToken: '',
      challengeExpiresAt: '',
    }
    pinia.state.value.buyNow = {
      item: {
        productId: 1,
        slug: 'shipping-product',
        title: { 'zh-CN': '实物商品' },
        priceAmount: '10.00',
        quantity: 1,
        requiresShippingAddress: true,
        paymentChannelIds: [1],
      },
    }

    const wrapper = mount(Checkout, {
      global: {
        plugins: [createHead(), pinia, i18n],
        stubs: {
          RouterLink: true,
          CheckoutSteps: true,
          EmptyState: true,
          SmartImage: true,
          CheckoutManualForm: true,
          GuestShippingAddressRecallCard: true,
          RegionSelector: true,
          ImageCaptcha: true,
          TurnstileCaptcha: true,
          CapCaptcha: true,
        },
      },
    })

    await nextTick()
    await flushPromises()
    const desktopGrid = wrapper.get('.hidden.grid-cols-1.gap-8')
    await desktopGrid.get('input[autocomplete="name"]').setValue('张三')
    await desktopGrid.get('input[autocomplete="tel"]').setValue('(138) 0013-8000')
    await desktopGrid.get('textarea[autocomplete="street-address"]').setValue('中关村大街 1 号')
    Object.assign((wrapper.vm as any).shippingAddress, {
      province: '北京市',
      province_code: '110000',
      city: '北京市',
      city_code: '110100',
      district: '东城区',
      district_code: '110101',
      township: '东华门街道',
      township_code: '110101001',
    })
    await nextTick()
    await flushPromises()

    await desktopGrid.get('.grid.grid-cols-2.gap-2 button').trigger('click')
    const submitButton = desktopGrid.findAll('button').find((button) => button.text() === i18n.global.t('checkout.submitButton'))
    expect(submitButton).toBeDefined()
    await submitButton!.trigger('click')
    await flushPromises()
    await nextTick()

    expect(userOrderAPI.createAndPay).toHaveBeenCalledWith(expect.objectContaining({
      shipping_address: expect.objectContaining({
        receiver_phone: '13800138000',
      }),
    }))
  })
})
