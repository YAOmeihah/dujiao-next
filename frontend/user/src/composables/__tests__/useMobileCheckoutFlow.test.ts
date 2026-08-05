import { describe, expect, it } from 'vitest'
import {
  buildMobileCheckoutFlow,
  getMobileSectionScrollTop,
  isMobileBuyerReady,
  isMobileManualFormReady,
  resolveMobileBuyerErrorMessage,
  resolveMobileErrorTargetSelectors,
  resolveMobilePaymentErrorMessage,
  isMobileStepConfirmed,
  isMobileStepDirty,
  isMobileShippingReady,
  resolveExpandedMobileSection,
} from '../useMobileCheckoutFlow'

describe('buildMobileCheckoutFlow', () => {
  it('starts on shipping when a shippable order has no address yet', () => {
    const result = buildMobileCheckoutFlow({
      hasShippingSection: true,
      shippingComplete: false,
      buyerComplete: false,
      paymentComplete: false,
    })

    expect(result.visibleSectionKeys).toEqual(['items', 'shipping', 'buyer', 'coupon', 'payment'])
    expect(result.recommendedSectionKey).toBe('shipping')
    expect(result.primaryActionKey).toBe('saveShipping')
  })

  it('skips shipping and moves straight to buyer info for digital-only orders', () => {
    const result = buildMobileCheckoutFlow({
      hasShippingSection: false,
      shippingComplete: true,
      buyerComplete: false,
      paymentComplete: false,
    })

    expect(result.visibleSectionKeys).toEqual(['items', 'buyer', 'coupon', 'payment'])
    expect(result.recommendedSectionKey).toBe('buyer')
    expect(result.primaryActionKey).toBe('continueBuyer')
  })

  it('switches the primary action to submit once required sections are complete', () => {
    const result = buildMobileCheckoutFlow({
      hasShippingSection: true,
      shippingComplete: true,
      buyerComplete: true,
      paymentComplete: true,
    })

    expect(result.recommendedSectionKey).toBe('payment')
    expect(result.primaryActionKey).toBe('submit')
  })
})

describe('resolveExpandedMobileSection', () => {
  it('keeps a user-opened later section expanded while earlier incomplete sections still exist', () => {
    const expanded = resolveExpandedMobileSection({
      expandedSectionKey: 'payment',
      recommendedSectionKey: 'shipping',
      completedSectionKeys: ['items'],
      visibleSectionKeys: ['items', 'shipping', 'buyer', 'coupon', 'payment'],
    })

    expect(expanded).toBe('payment')
  })

  it('auto-advances after the currently open section becomes complete', () => {
    const expanded = resolveExpandedMobileSection({
      expandedSectionKey: 'shipping',
      recommendedSectionKey: 'buyer',
      completedSectionKeys: ['items', 'shipping'],
      visibleSectionKeys: ['items', 'shipping', 'buyer', 'coupon', 'payment'],
    })

    expect(expanded).toBe('buyer')
  })
})

describe('mobile section readiness', () => {
  it('does not treat one-character shipping data as ready', () => {
    const ready = isMobileShippingReady({
      requiresShipping: true,
      receiverName: 'A',
      receiverPhone: '1',
      provinceCode: '110000',
      cityCode: '110100',
      districtCode: '110101',
      townshipCode: '110101001',
      detailAddress: '1',
    })

    expect(ready).toBe(false)
  })

  it('does not treat one-character manual text fields as ready', () => {
    const ready = isMobileManualFormReady(
      [
        {
          itemKey: '1',
          fields: [
            {
              key: 'qq',
              type: 'text',
              required: true,
              options: [],
            },
          ],
        },
      ],
      {
        '1': {
          qq: 'a',
        },
      },
    )

    expect(ready).toBe(false)
  })

  it('requires a meaningful guest password before buyer info is ready', () => {
    const ready = isMobileBuyerReady({
      isAuthenticated: false,
      checkoutMode: 'guest',
      manualFormsReady: true,
      guestPhone: '13800138000',
      guestPassword: '1',
      guestEmail: '',
      captchaComplete: true,
    })

    expect(ready).toBe(false)
  })
})

describe('mobile step confirmation', () => {
  it('does not mark a ready step as confirmed before the user explicitly confirms it', () => {
    const confirmed = isMobileStepConfirmed({
      ready: true,
      currentFingerprint: '{"receiver":"alice"}',
      confirmedFingerprint: '',
    })

    expect(confirmed).toBe(false)
  })

  it('marks a previously confirmed step as dirty after the user edits it', () => {
    const dirty = isMobileStepDirty({
      currentFingerprint: '{"receiver":"bob"}',
      confirmedFingerprint: '{"receiver":"alice"}',
    })

    expect(dirty).toBe(true)
  })
})

describe('mobile section scrolling', () => {
  it('subtracts the fixed header height and breathing room from the target position', () => {
    const top = getMobileSectionScrollTop({
      currentScrollY: 320,
      elementTop: 260,
      fixedOffset: 77,
      gap: 16,
    })

    expect(top).toBe(487)
  })

  it('clamps the scroll target to the top of the page', () => {
    const top = getMobileSectionScrollTop({
      currentScrollY: 0,
      elementTop: 40,
      fixedOffset: 77,
      gap: 16,
    })

    expect(top).toBe(0)
  })
})

describe('mobile section errors', () => {
  it('returns a guest-info error when buyer confirmation is blocked by empty guest credentials', () => {
    const error = resolveMobileBuyerErrorMessage({
      manualFormsValid: true,
      manualFormFirstError: '',
      isAuthenticated: false,
      checkoutMode: 'guest',
      guestPhone: '',
      guestPassword: '',
      guestPhoneValid: true,
      guestEmailValid: true,
      guestCaptchaComplete: true,
      loginOrGuestMessage: '请先登录或使用游客购买',
      missingGuestMessage: '请输入手机号与订单密码',
      invalidPhoneMessage: '手机号格式不正确',
      invalidEmailMessage: '邮箱格式不正确',
      captchaRequiredMessage: '请先完成验证码',
      fallbackMessage: '请继续填写购买信息',
    })

    expect(error).toBe('请输入手机号与订单密码')
  })

  it('prefers the selected-channel amount hint over a generic payment error', () => {
    const error = resolveMobilePaymentErrorMessage({
      walletOnlyPayment: false,
      expectedOnlinePayCents: 200,
      requiresOnlineChannel: true,
      selectedChannelId: 3,
      selectedChannelAmountHint: '该支付方式不支持当前金额',
      walletInsufficientMessage: '余额不足',
      selectPaymentMessage: '请选择支付方式',
      fallbackMessage: '请选择支付方式',
    })

    expect(error).toBe('该支付方式不支持当前金额')
  })
})

describe('mobile error target selectors', () => {
  it('scrolls to the section error banner while keeping focus on the first invalid input', () => {
    const selectors = resolveMobileErrorTargetSelectors({
      sectionKey: 'buyer',
      focusSelector: '[data-mobile-buyer-input="guest-phone"]',
    })

    expect(selectors.scrollSelector).toBe('[data-section-error="buyer"], [data-mobile-buyer-input="guest-phone"]')
    expect(selectors.focusSelector).toBe('[data-mobile-buyer-input="guest-phone"]')
  })

  it('falls back to the section toggle when no specific field target exists', () => {
    const selectors = resolveMobileErrorTargetSelectors({
      sectionKey: 'payment',
      focusSelector: '',
    })

    expect(selectors.scrollSelector).toBe('[data-section-error="payment"], [data-section-toggle="payment"]')
    expect(selectors.focusSelector).toBe('')
  })
})
