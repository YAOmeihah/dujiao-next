import { computed, nextTick, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import type { useCheckout } from './useCheckout'
import { amountToCents } from '../utils/money'
import {
  buildMobileCheckoutFlow,
  getMobileSectionScrollTop,
  isMobileBuyerReady,
  isMobileManualFormReady,
  isMobileStepConfirmed,
  isMobileStepDirty,
  isMobileShippingReady,
  resolveMobileBuyerErrorMessage,
  resolveMobileErrorTargetSelectors,
  resolveMobilePaymentErrorMessage,
  resolveExpandedMobileSection,
  type MobileCheckoutSectionKey,
} from './useMobileCheckoutFlow'

type CheckoutContract = ReturnType<typeof useCheckout>
type MobileConfirmableSectionKey = 'shipping' | 'buyer' | 'payment'

const MOBILE_CHECKOUT_SECTION_TRANSITION_MS = 220

const maskPhone = (value: string) => {
  const trimmed = value.trim()
  if (trimmed.length < 7) return trimmed
  return `${trimmed.slice(0, 3)}****${trimmed.slice(-4)}`
}

export function useMobileCheckoutAdapter(checkout: CheckoutContract) {
  const { t } = useI18n()
  const mobileExpandedSection = ref<MobileCheckoutSectionKey | null>(null)
  const mobileConfirmedFingerprints = ref<Partial<Record<MobileConfirmableSectionKey, string>>>({})

  const guestCaptchaComplete = computed(() => {
    if (!checkout.guestCaptchaEnabled.value) return true
    if (checkout.captchaProvider.value === 'image') {
      return Boolean(
        checkout.guestCaptchaPayload.value.captcha_id &&
        checkout.guestCaptchaPayload.value.captcha_code,
      )
    }
    if (checkout.captchaProvider.value === 'turnstile') {
      return Boolean(checkout.guestTurnstileToken.value)
    }
    if (checkout.captchaProvider.value === 'cap') {
      return Boolean(checkout.guestCapToken.value)
    }
    return false
  })

  const mobileManualFormsReady = computed(() => isMobileManualFormReady(
    checkout.manualFormProducts.value,
    checkout.manualFormData.value,
  ))

  const mobileShippingReady = computed(() => isMobileShippingReady({
    requiresShipping: checkout.orderRequiresShippingAddress.value,
    receiverName: checkout.shippingAddress.value.receiver_name,
    receiverPhone: checkout.shippingAddress.value.receiver_phone,
    provinceCode: checkout.shippingAddress.value.province_code,
    cityCode: checkout.shippingAddress.value.city_code,
    districtCode: checkout.shippingAddress.value.district_code,
    townshipCode: checkout.shippingAddress.value.township_code,
    detailAddress: checkout.shippingAddress.value.detail_address,
  }))

  const mobileBuyerReady = computed(() => isMobileBuyerReady({
    isAuthenticated: checkout.userAuthStore.isAuthenticated,
    checkoutMode: checkout.checkoutMode.value,
    manualFormsReady: mobileManualFormsReady.value,
    guestPhone: checkout.guestPhone.value,
    guestPassword: checkout.guestPassword.value,
    guestEmail: checkout.guestEmail.value,
    captchaComplete: guestCaptchaComplete.value,
  }))

  const mobilePaymentReady = computed(() => {
    if (checkout.walletOnlyPayment.value) return checkout.expectedOnlinePayCents.value <= 0
    if (!checkout.requiresOnlineChannel.value) return true
    return Boolean(checkout.selectedChannelId.value) && !checkout.mobileCheckout.selectedChannelAmountHint.value
  })

  const mobileShippingFingerprint = computed(() => JSON.stringify({
    requiresShipping: checkout.orderRequiresShippingAddress.value,
    address: checkout.mobileCheckout.shippingAddressFingerprint.value,
  }))

  const mobileBuyerFingerprint = computed(() => JSON.stringify({
    isAuthenticated: checkout.userAuthStore.isAuthenticated,
    checkoutMode: checkout.checkoutMode.value,
    guestPhone: checkout.guestPhone.value.trim(),
    guestEmail: checkout.guestEmail.value.trim(),
    guestPassword: checkout.guestPassword.value,
    guestCaptchaComplete: guestCaptchaComplete.value,
    manualFormData: checkout.mobileCheckout.manualFormFingerprint.value,
  }))

  const mobilePaymentFingerprint = computed(() => JSON.stringify({
    useBalance: checkout.useBalance.value,
    selectedChannelId: checkout.requiresOnlineChannel.value ? checkout.selectedChannelId.value : null,
    requiresOnlineChannel: checkout.requiresOnlineChannel.value,
    expectedOnlinePayCents: checkout.expectedOnlinePayCents.value,
    walletOnlyPayment: checkout.walletOnlyPayment.value,
  }))

  const mobileShippingDirty = computed(() => {
    if (!checkout.orderRequiresShippingAddress.value) return false
    return isMobileStepDirty({
      currentFingerprint: mobileShippingFingerprint.value,
      confirmedFingerprint: mobileConfirmedFingerprints.value.shipping,
    })
  })

  const mobileBuyerDirty = computed(() => isMobileStepDirty({
    currentFingerprint: mobileBuyerFingerprint.value,
    confirmedFingerprint: mobileConfirmedFingerprints.value.buyer,
  }))

  const mobilePaymentDirty = computed(() => isMobileStepDirty({
    currentFingerprint: mobilePaymentFingerprint.value,
    confirmedFingerprint: mobileConfirmedFingerprints.value.payment,
  }))

  const mobileShippingComplete = computed(() => {
    if (!checkout.orderRequiresShippingAddress.value) return true
    return isMobileStepConfirmed({
      ready: mobileShippingReady.value,
      currentFingerprint: mobileShippingFingerprint.value,
      confirmedFingerprint: mobileConfirmedFingerprints.value.shipping,
    })
  })

  const mobileBuyerComplete = computed(() => isMobileStepConfirmed({
    ready: mobileBuyerReady.value,
    currentFingerprint: mobileBuyerFingerprint.value,
    confirmedFingerprint: mobileConfirmedFingerprints.value.buyer,
  }))

  const mobilePaymentComplete = computed(() => isMobileStepConfirmed({
    ready: mobilePaymentReady.value,
    currentFingerprint: mobilePaymentFingerprint.value,
    confirmedFingerprint: mobileConfirmedFingerprints.value.payment,
  }))

  const mobileFlowState = computed(() => buildMobileCheckoutFlow({
    hasShippingSection: checkout.orderRequiresShippingAddress.value,
    shippingComplete: mobileShippingComplete.value,
    buyerComplete: mobileBuyerComplete.value,
    paymentComplete: mobilePaymentComplete.value,
    showCouponSection: !checkout.isResellerTenant.value,
  }))

  const mobileShippingErrorMessage = computed(() => {
    if (!checkout.submitAttempted.value) return ''
    if (mobileFlowState.value.recommendedSectionKey !== 'shipping') return ''
    if (mobileShippingReady.value) return ''
    return checkout.shippingAddressValidation.value.message || t('checkout.mobile.shippingMissing')
  })

  const mobileBuyerErrorMessage = computed(() => {
    if (!checkout.submitAttempted.value) return ''
    if (mobileFlowState.value.recommendedSectionKey !== 'buyer') return ''
    if (mobileBuyerReady.value) return ''

    return resolveMobileBuyerErrorMessage({
      manualFormsValid: checkout.mobileCheckout.manualFormValidation.value.valid,
      manualFormFirstError: checkout.mobileCheckout.manualFormValidation.value.firstError,
      isAuthenticated: checkout.userAuthStore.isAuthenticated,
      checkoutMode: checkout.checkoutMode.value,
      guestPhone: checkout.guestPhone.value,
      guestPassword: checkout.guestPassword.value,
      guestPhoneValid: checkout.guestPhoneValid.value,
      guestEmailValid: checkout.guestEmailValid.value,
      guestCaptchaComplete: guestCaptchaComplete.value,
      loginOrGuestMessage: t('checkout.errors.loginOrGuest'),
      missingGuestMessage: t('checkout.errors.missingGuest'),
      invalidPhoneMessage: t('error.phone_invalid'),
      invalidEmailMessage: t('error.email_invalid'),
      captchaRequiredMessage: t('auth.common.captchaRequired'),
      fallbackMessage: t('checkout.mobile.buyerMissing'),
    })
  })

  const mobilePaymentErrorMessage = computed(() => {
    if (!checkout.submitAttempted.value) return ''
    if (mobileFlowState.value.recommendedSectionKey !== 'payment') return ''
    if (mobilePaymentReady.value) return ''

    return resolveMobilePaymentErrorMessage({
      walletOnlyPayment: checkout.walletOnlyPayment.value,
      expectedOnlinePayCents: checkout.expectedOnlinePayCents.value,
      requiresOnlineChannel: checkout.requiresOnlineChannel.value,
      selectedChannelId: checkout.selectedChannelId.value,
      selectedChannelAmountHint: checkout.mobileCheckout.selectedChannelAmountHint.value,
      walletInsufficientMessage: t('payment.walletInsufficientHint'),
      selectPaymentMessage: t('checkout.errors.selectPayment'),
      fallbackMessage: t('checkout.mobile.paymentMissing'),
    })
  })

  const mobileCurrentSectionErrorMessage = computed(() => {
    const action = mobileFlowState.value.primaryActionKey
    if (action === 'saveShipping') return mobileShippingErrorMessage.value
    if (action === 'continueBuyer') return mobileBuyerErrorMessage.value
    if (action === 'choosePayment') return mobilePaymentErrorMessage.value
    return ''
  })

  const mobileStatusText = computed(() => {
    if (checkout.mobileCheckout.error.value) return checkout.mobileCheckout.error.value
    if (checkout.mobileCheckout.previewError.value) return checkout.mobileCheckout.previewError.value
    if (checkout.previewLoading.value || checkout.couponRefreshing.value) return checkout.previewStatusText.value
    if (mobileCurrentSectionErrorMessage.value) return mobileCurrentSectionErrorMessage.value

    const action = mobileFlowState.value.primaryActionKey
    if (action === 'saveShipping') {
      return mobileShippingReady.value
        ? t('checkout.mobile.actionSaveShipping')
        : t('checkout.mobile.shippingMissing')
    }
    if (action === 'continueBuyer') {
      return mobileBuyerReady.value
        ? t('checkout.mobile.actionContinueBuyer')
        : t('checkout.mobile.buyerMissing')
    }
    if (action === 'choosePayment') {
      return mobilePaymentReady.value
        ? t('checkout.mobile.readyToSubmit')
        : t('checkout.mobile.paymentMissing')
    }
    return t('checkout.mobile.readyToSubmit')
  })

  const mobilePrimaryActionLabel = computed(() => {
    const action = mobileFlowState.value.primaryActionKey
    if (action === 'saveShipping') return t('checkout.mobile.actionSaveShipping')
    if (action === 'continueBuyer') return t('checkout.mobile.actionContinueBuyer')
    if (action === 'choosePayment') {
      return mobilePaymentReady.value
        ? t('checkout.mobile.actionSubmit')
        : t('checkout.mobile.actionChoosePayment')
    }
    return t('checkout.mobile.actionSubmit')
  })

  const mobileTotalText = computed(() => checkout.formatPrice(
    checkout.previewTotal.value,
    checkout.previewCurrency.value,
  ))

  const mobileDisplaySections = computed(() => {
    const state = mobileFlowState.value
    const titleMap: Record<MobileCheckoutSectionKey, string> = {
      items: t('checkout.mobile.sectionItems'),
      shipping: t('checkout.mobile.sectionShipping'),
      buyer: t('checkout.mobile.sectionBuyer'),
      coupon: t('checkout.mobile.sectionCoupon'),
      payment: t('checkout.mobile.sectionPayment'),
    }

    const shippingRegionLine = [
      checkout.shippingAddress.value.province,
      checkout.shippingAddress.value.city,
      checkout.shippingAddress.value.district,
      checkout.shippingAddress.value.township,
      checkout.shippingAddress.value.detail_address,
    ]
      .map((value) => String(value || '').trim())
      .filter(Boolean)
      .join(' ')

    const shippingSummaryLines = [
      checkout.shippingAddress.value.receiver_name && checkout.shippingAddress.value.receiver_phone
        ? `${checkout.shippingAddress.value.receiver_name} · ${maskPhone(checkout.shippingAddress.value.receiver_phone)}`
        : '',
      shippingRegionLine,
    ].filter(Boolean)

    const buyerSummaryLines = [
      checkout.userAuthStore.isAuthenticated
        ? t('checkout.mobile.buyerLoggedIn')
        : checkout.checkoutMode.value === 'guest' && checkout.guestPhone.value.trim()
          ? t('checkout.mobile.buyerGuest', { phone: maskPhone(checkout.guestPhone.value) })
          : t('checkout.mobile.buyerMissing'),
      checkout.manualFormProducts.value.length > 0 && !mobileManualFormsReady.value
        ? t('checkout.mobile.buyerManualPending')
        : '',
    ].filter(Boolean)

    const normalizedCouponCode = checkout.couponCode.value.trim()
    const couponDiscountCents = amountToCents(checkout.previewCoupon.value)
    const couponSummaryLines = normalizedCouponCode
      ? [
          normalizedCouponCode,
          couponDiscountCents !== null && couponDiscountCents > 0
            ? t('checkout.mobile.summaryCouponApplied', {
              amount: checkout.formatPrice(checkout.previewCoupon.value, checkout.previewCurrency.value),
            })
            : '',
        ].filter(Boolean)
      : [t('checkout.mobile.summaryCouponEmpty')]

    const selectedChannel = checkout.paymentChannels.value.find(
      (channel: any) => Number(channel?.id) === Number(checkout.selectedChannelId.value),
    )
    const paymentSummaryLines = [
      checkout.useBalance.value && checkout.mobileCheckout.expectedWalletPaidCents.value > 0
        ? `${t('payment.walletDeductLabel')} ${checkout.expectedWalletPaidDisplay.value}`
        : '',
      !checkout.requiresOnlineChannel.value
        ? t('checkout.walletCoversAll')
        : selectedChannel?.name || t('checkout.mobile.paymentMissing'),
    ].filter(Boolean)

    const summaryMap: Record<MobileCheckoutSectionKey, string[]> = {
      items: [
        t('checkout.mobile.summaryItems', { count: checkout.totalItems.value }),
        `${t('checkout.previewTotal')} ${checkout.formatPrice(checkout.previewTotal.value, checkout.previewCurrency.value)}`,
      ],
      shipping: shippingSummaryLines.length > 0
        ? shippingSummaryLines
        : [t('checkout.mobile.shippingMissing')],
      buyer: buyerSummaryLines,
      coupon: couponSummaryLines,
      payment: paymentSummaryLines,
    }

    const dirtyMap: Record<MobileCheckoutSectionKey, boolean> = {
      items: false,
      shipping: mobileShippingDirty.value,
      buyer: mobileBuyerDirty.value,
      coupon: false,
      payment: mobilePaymentDirty.value,
    }

    const errorMap: Record<MobileCheckoutSectionKey, string> = {
      items: '',
      shipping: mobileShippingErrorMessage.value,
      buyer: mobileBuyerErrorMessage.value,
      coupon: '',
      payment: mobilePaymentErrorMessage.value,
    }

    return state.visibleSectionKeys.map((key) => {
      const complete = state.completedSectionKeys.includes(key)
      const recommended = state.recommendedSectionKey === key
      const dirty = dirtyMap[key]

      return {
        key,
        title: titleMap[key],
        badge: key === 'items'
          ? ''
          : complete
            ? t('checkout.mobile.complete')
            : dirty
              ? t('checkout.mobile.needsReconfirm')
              : recommended
                ? t('checkout.mobile.current')
                : key === 'coupon'
                  ? t('checkout.mobile.optional')
                  : t('checkout.mobile.pending'),
        summaryLines: summaryMap[key],
        errorMessage: errorMap[key],
        collapsedActionLabel: key === 'items' ? t('checkout.mobile.viewDetails') : '',
        complete,
        recommended,
        softHint: key !== 'items' && !complete && !recommended
          ? t('checkout.mobile.softGuide', { step: titleMap[state.recommendedSectionKey] })
          : '',
      }
    })
  })

  const scrollMobileSectionIntoView = async (sectionKey: MobileCheckoutSectionKey) => {
    const waitForAnimationFrame = () => new Promise<void>((resolve) => {
      requestAnimationFrame(() => resolve())
    })

    await nextTick()
    await waitForAnimationFrame()

    for (let attempt = 0; attempt < 24; attempt += 1) {
      if (!document.querySelector('.mobile-checkout-section-leave-active')) break
      await waitForAnimationFrame()
    }

    const section = document.querySelector(`[data-section-toggle="${sectionKey}"]`)
    if (section instanceof HTMLElement) {
      const siteHeader = document.querySelector('[data-site-header]')
      const fixedOffset = siteHeader instanceof HTMLElement
        ? siteHeader.getBoundingClientRect().height
        : 0
      const top = getMobileSectionScrollTop({
        currentScrollY: window.scrollY,
        elementTop: section.getBoundingClientRect().top,
        fixedOffset,
        gap: 16,
      })

      window.scrollTo({ top, behavior: 'smooth' })
    }
  }

  const scrollMobileElementIntoView = async (selector: string, focusSelector = '') => {
    await nextTick()
    const target = document.querySelector(selector)
    if (!(target instanceof HTMLElement)) return

    const siteHeader = document.querySelector('[data-site-header]')
    const fixedOffset = siteHeader instanceof HTMLElement
      ? siteHeader.getBoundingClientRect().height
      : 0
    const top = getMobileSectionScrollTop({
      currentScrollY: window.scrollY,
      elementTop: target.getBoundingClientRect().top,
      fixedOffset,
      gap: 16,
    })

    window.scrollTo({ top, behavior: 'smooth' })

    const explicitFocusTarget = focusSelector ? document.querySelector(focusSelector) : null
    const focusTarget = explicitFocusTarget instanceof HTMLElement
      ? explicitFocusTarget
      : target.matches('input, textarea, select, button')
        ? target
        : target.querySelector<HTMLElement>('input, textarea, select, button, [tabindex]:not([tabindex="-1"])')

    focusTarget?.focus?.({ preventScroll: true })
  }

  const getMobileShippingErrorSelector = () => {
    if (!checkout.shippingAddress.value.receiver_name.trim()) return '[data-mobile-shipping-input="receiver-name"]'
    if (!checkout.shippingAddress.value.receiver_phone.trim()) return '[data-mobile-shipping-input="receiver-phone"]'
    if (checkout.shippingRegionMissing.value) return '[data-mobile-shipping-input="region"]'
    if (!checkout.shippingAddress.value.detail_address.trim()) return '[data-mobile-shipping-input="detail-address"]'
    return '[data-section-toggle="shipping"]'
  }

  const getMobileBuyerErrorSelector = () => {
    const firstManualFieldErrorKey = Object.keys(checkout.mobileCheckout.manualFormValidation.value.errors)[0]
    if (firstManualFieldErrorKey) return `[data-manual-field-input="${firstManualFieldErrorKey}"]`
    if (!checkout.userAuthStore.isAuthenticated && checkout.checkoutMode.value !== 'guest') {
      return '[data-mobile-buyer-input="checkout-mode"]'
    }
    if (!checkout.guestPhone.value.trim() || !checkout.guestPhoneValid.value) {
      return '[data-mobile-buyer-input="guest-phone"]'
    }
    if (!checkout.guestPassword.value.trim()) return '[data-mobile-buyer-input="guest-password"]'
    if (!checkout.guestEmailValid.value) return '[data-mobile-buyer-input="guest-email"]'
    if (!guestCaptchaComplete.value) return '[data-mobile-buyer-input="guest-captcha"]'
    return '[data-section-toggle="buyer"]'
  }

  const getMobilePaymentErrorSelector = () => {
    if (checkout.requiresOnlineChannel.value) {
      return '[data-mobile-payment-input="channel-list"], [data-section-toggle="payment"]'
    }
    return '[data-section-toggle="payment"]'
  }

  const focusMobileErrorTarget = async (sectionKey: MobileCheckoutSectionKey) => {
    const selectorMap: Partial<Record<MobileCheckoutSectionKey, string>> = {
      shipping: getMobileShippingErrorSelector(),
      buyer: getMobileBuyerErrorSelector(),
      payment: getMobilePaymentErrorSelector(),
    }

    const selectors = resolveMobileErrorTargetSelectors({
      sectionKey,
      focusSelector: selectorMap[sectionKey] || '',
    })

    await scrollMobileElementIntoView(selectors.scrollSelector, selectors.focusSelector)
  }

  const confirmMobileSection = (
    sectionKey: MobileConfirmableSectionKey,
    fingerprint: string,
  ) => {
    mobileConfirmedFingerprints.value = {
      ...mobileConfirmedFingerprints.value,
      [sectionKey]: fingerprint,
    }
  }

  watch(mobileFlowState, (state) => {
    mobileExpandedSection.value = resolveExpandedMobileSection({
      expandedSectionKey: mobileExpandedSection.value,
      recommendedSectionKey: state.recommendedSectionKey,
      completedSectionKeys: state.completedSectionKeys,
      visibleSectionKeys: state.visibleSectionKeys,
    })
  }, { immediate: true })

  watch(mobileExpandedSection, async (sectionKey, previousSectionKey) => {
    if (!sectionKey || sectionKey === previousSectionKey) return
    if (sectionKey !== mobileFlowState.value.recommendedSectionKey) return

    if (previousSectionKey) {
      await new Promise<void>((resolve) => {
        window.setTimeout(() => resolve(), MOBILE_CHECKOUT_SECTION_TRANSITION_MS + 40)
      })
    }

    await scrollMobileSectionIntoView(sectionKey)
  })

  const handleMobileSectionChange = (sectionKey: string | null) => {
    mobileExpandedSection.value = sectionKey as MobileCheckoutSectionKey | null
  }

  const handleMobilePrimaryAction = async () => {
    checkout.submitAttempted.value = true

    const action = mobileFlowState.value.primaryActionKey
    if (action === 'saveShipping') {
      mobileExpandedSection.value = 'shipping'
      await scrollMobileSectionIntoView('shipping')
      if (!mobileShippingReady.value) {
        await focusMobileErrorTarget('shipping')
        return
      }

      confirmMobileSection('shipping', mobileShippingFingerprint.value)
      checkout.mobileCheckout.persistGuestShippingRecallFromCurrentAddress()
      await nextTick()
      return
    }
    if (action === 'continueBuyer') {
      mobileExpandedSection.value = 'buyer'
      await scrollMobileSectionIntoView('buyer')
      if (!mobileBuyerReady.value) {
        await focusMobileErrorTarget('buyer')
        return
      }

      confirmMobileSection('buyer', mobileBuyerFingerprint.value)
      await nextTick()
      return
    }
    if (action === 'choosePayment') {
      mobileExpandedSection.value = 'payment'
      await scrollMobileSectionIntoView('payment')
      if (!mobilePaymentReady.value) {
        await focusMobileErrorTarget('payment')
        return
      }

      confirmMobileSection('payment', mobilePaymentFingerprint.value)
      await nextTick()
      await checkout.handleSubmit()
      return
    }

    await checkout.handleSubmit()
  }

  return {
    guestCaptchaComplete,
    mobileDisplaySections,
    mobileExpandedSection,
    mobilePrimaryActionDisabled: computed(() => (
      checkout.submitting.value || checkout.mobileCheckout.syncingStock.value
    )),
    mobilePrimaryActionLabel,
    mobileStatusText,
    mobileTotalText,
    handleMobilePrimaryAction,
    handleMobileSectionChange,
  }
}
