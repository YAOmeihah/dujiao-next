export type MobileCheckoutSectionKey = 'items' | 'shipping' | 'buyer' | 'coupon' | 'payment'
export type MobileCheckoutPrimaryActionKey = 'saveShipping' | 'continueBuyer' | 'choosePayment' | 'submit'

export interface MobileManualFormFieldInput {
  key: string
  type: string
  required: boolean
  options: string[]
}

export interface MobileManualFormProductInput {
  itemKey: string
  fields: MobileManualFormFieldInput[]
}

export interface MobileCheckoutFlowInput {
  hasShippingSection: boolean
  shippingComplete: boolean
  buyerComplete: boolean
  paymentComplete: boolean
}

export interface MobileCheckoutFlowState {
  visibleSectionKeys: MobileCheckoutSectionKey[]
  completedSectionKeys: MobileCheckoutSectionKey[]
  recommendedSectionKey: MobileCheckoutSectionKey
  primaryActionKey: MobileCheckoutPrimaryActionKey
}

export interface ResolveExpandedMobileSectionInput {
  expandedSectionKey: MobileCheckoutSectionKey | null
  recommendedSectionKey: MobileCheckoutSectionKey
  completedSectionKeys: MobileCheckoutSectionKey[]
  visibleSectionKeys: MobileCheckoutSectionKey[]
}

export interface MobileShippingReadinessInput {
  requiresShipping: boolean
  receiverName: string
  receiverPhone: string
  provinceCode: string
  cityCode: string
  districtCode: string
  townshipCode: string
  detailAddress: string
}

export interface MobileBuyerReadinessInput {
  isAuthenticated: boolean
  checkoutMode: 'guest' | 'member'
  manualFormsReady: boolean
  guestPhone: string
  guestPassword: string
  guestEmail: string
  captchaComplete: boolean
}

export interface MobileStepConfirmationInput {
  ready: boolean
  currentFingerprint: string
  confirmedFingerprint?: string | null
}

export interface MobileStepDirtyInput {
  currentFingerprint: string
  confirmedFingerprint?: string | null
}

export interface MobileSectionScrollTopInput {
  currentScrollY: number
  elementTop: number
  fixedOffset: number
  gap?: number
}

export interface MobileBuyerErrorMessageInput {
  manualFormsValid: boolean
  manualFormFirstError: string
  isAuthenticated: boolean
  checkoutMode: 'guest' | 'member'
  guestPhone: string
  guestPassword: string
  guestPhoneValid: boolean
  guestEmailValid: boolean
  guestCaptchaComplete: boolean
  loginOrGuestMessage: string
  missingGuestMessage: string
  invalidPhoneMessage: string
  invalidEmailMessage: string
  captchaRequiredMessage: string
  fallbackMessage: string
}

export interface MobilePaymentErrorMessageInput {
  walletOnlyPayment: boolean
  expectedOnlinePayCents: number
  requiresOnlineChannel: boolean
  selectedChannelId: number | string | null | undefined
  selectedChannelAmountHint: string
  walletInsufficientMessage: string
  selectPaymentMessage: string
  fallbackMessage: string
}

export interface MobileErrorTargetSelectorsInput {
  sectionKey: MobileCheckoutSectionKey
  focusSelector: string
}

const phonePattern = /^\+?[0-9\-()\s]{6,20}$/
const emailPattern = /^[^\s@]+@[^\s@]+\.[^\s@]+$/
const meaningfulTextTypes = new Set(['text', 'textarea'])

const hasMeaningfulText = (value: unknown, minLength = 2) =>
  String(value ?? '').trim().length >= minLength

export const isMobileShippingReady = ({
  requiresShipping,
  receiverName,
  receiverPhone,
  provinceCode,
  cityCode,
  districtCode,
  townshipCode,
  detailAddress,
}: MobileShippingReadinessInput) => {
  if (!requiresShipping) return true

  return (
    hasMeaningfulText(receiverName, 2) &&
    phonePattern.test(String(receiverPhone || '').trim()) &&
    hasMeaningfulText(provinceCode, 1) &&
    hasMeaningfulText(cityCode, 1) &&
    hasMeaningfulText(districtCode, 1) &&
    hasMeaningfulText(townshipCode, 1) &&
    hasMeaningfulText(detailAddress, 4)
  )
}

export const isMobileManualFormReady = (
  manualFormProducts: MobileManualFormProductInput[],
  manualFormData: Record<string, Record<string, unknown>>,
) => {
  return manualFormProducts.every((product) =>
    product.fields.every((field) => {
      if (!field.required) return true

      const value = manualFormData[product.itemKey]?.[field.key]

      if (field.type === 'checkbox') {
        return Array.isArray(value) && value.some((item) => hasMeaningfulText(item, 1))
      }

      if (field.type === 'radio' || field.type === 'select') {
        return hasMeaningfulText(value, 1)
      }

      if (field.type === 'phone') {
        return phonePattern.test(String(value || '').trim())
      }

      if (field.type === 'email') {
        return emailPattern.test(String(value || '').trim())
      }

      if (field.type === 'number') {
        return String(value ?? '').trim() !== ''
      }

      if (meaningfulTextTypes.has(field.type)) {
        return hasMeaningfulText(value, 2)
      }

      return hasMeaningfulText(value, 1)
    }),
  )
}

export const isMobileBuyerReady = ({
  isAuthenticated,
  checkoutMode,
  manualFormsReady,
  guestPhone,
  guestPassword,
  guestEmail,
  captchaComplete,
}: MobileBuyerReadinessInput) => {
  if (!manualFormsReady) return false
  if (isAuthenticated) return true
  if (checkoutMode !== 'guest') return false
  if (!phonePattern.test(String(guestPhone || '').trim())) return false
  if (!hasMeaningfulText(guestPassword, 4)) return false

  const trimmedEmail = String(guestEmail || '').trim()
  if (trimmedEmail && !emailPattern.test(trimmedEmail)) return false

  return captchaComplete
}

export const isMobileStepConfirmed = ({
  ready,
  currentFingerprint,
  confirmedFingerprint,
}: MobileStepConfirmationInput) => {
  if (!ready) return false
  if (!currentFingerprint) return false

  return currentFingerprint === String(confirmedFingerprint || '')
}

export const isMobileStepDirty = ({
  currentFingerprint,
  confirmedFingerprint,
}: MobileStepDirtyInput) => {
  const lastConfirmedFingerprint = String(confirmedFingerprint || '')
  if (!lastConfirmedFingerprint) return false

  return currentFingerprint !== lastConfirmedFingerprint
}

export const getMobileSectionScrollTop = ({
  currentScrollY,
  elementTop,
  fixedOffset,
  gap = 16,
}: MobileSectionScrollTopInput) =>
  Math.max(currentScrollY + elementTop - fixedOffset - gap, 0)

export const resolveMobileBuyerErrorMessage = ({
  manualFormsValid,
  manualFormFirstError,
  isAuthenticated,
  checkoutMode,
  guestPhone,
  guestPassword,
  guestPhoneValid,
  guestEmailValid,
  guestCaptchaComplete,
  loginOrGuestMessage,
  missingGuestMessage,
  invalidPhoneMessage,
  invalidEmailMessage,
  captchaRequiredMessage,
  fallbackMessage,
}: MobileBuyerErrorMessageInput) => {
  if (!manualFormsValid) {
    return manualFormFirstError || fallbackMessage
  }
  if (isAuthenticated) return ''
  if (checkoutMode !== 'guest') return loginOrGuestMessage
  if (!guestPhone.trim() || !guestPassword.trim()) return missingGuestMessage
  if (!guestPhoneValid) return invalidPhoneMessage
  if (!guestEmailValid) return invalidEmailMessage
  if (!guestCaptchaComplete) return captchaRequiredMessage
  return fallbackMessage
}

export const resolveMobilePaymentErrorMessage = ({
  walletOnlyPayment,
  expectedOnlinePayCents,
  requiresOnlineChannel,
  selectedChannelId,
  selectedChannelAmountHint,
  walletInsufficientMessage,
  selectPaymentMessage,
  fallbackMessage,
}: MobilePaymentErrorMessageInput) => {
  if (walletOnlyPayment && expectedOnlinePayCents > 0) {
    return walletInsufficientMessage
  }
  if (!requiresOnlineChannel) return ''
  if (selectedChannelAmountHint) return selectedChannelAmountHint
  if (!selectedChannelId) return selectPaymentMessage
  return fallbackMessage
}

export const resolveMobileErrorTargetSelectors = ({
  sectionKey,
  focusSelector,
}: MobileErrorTargetSelectorsInput) => ({
  scrollSelector: focusSelector
    ? `[data-section-error="${sectionKey}"], ${focusSelector}`
    : `[data-section-error="${sectionKey}"], [data-section-toggle="${sectionKey}"]`,
  focusSelector,
})

const sectionIsComplete = (
  sectionKey: MobileCheckoutSectionKey,
  input: MobileCheckoutFlowInput,
) => {
  if (sectionKey === 'items') return true
  if (sectionKey === 'shipping') return input.shippingComplete
  if (sectionKey === 'buyer') return input.buyerComplete
  if (sectionKey === 'coupon') return false

  return input.paymentComplete
}

export const buildMobileCheckoutFlow = (
  input: MobileCheckoutFlowInput,
): MobileCheckoutFlowState => {
  const visibleSectionKeys: MobileCheckoutSectionKey[] = input.hasShippingSection
    ? ['items', 'shipping', 'buyer', 'coupon', 'payment']
    : ['items', 'buyer', 'coupon', 'payment']

  const completedSectionKeys = visibleSectionKeys.filter((sectionKey) =>
    sectionIsComplete(sectionKey, input),
  )

  const recommendedSectionKey =
    visibleSectionKeys.find((sectionKey) => {
      if (sectionKey === 'coupon' || sectionKey === 'items') return false

      return !completedSectionKeys.includes(sectionKey)
    }) ?? 'payment'

  let primaryActionKey: MobileCheckoutPrimaryActionKey = 'submit'

  if (recommendedSectionKey === 'shipping') {
    primaryActionKey = 'saveShipping'
  } else if (recommendedSectionKey === 'buyer') {
    primaryActionKey = 'continueBuyer'
  } else if (recommendedSectionKey === 'payment' && !input.paymentComplete) {
    primaryActionKey = 'choosePayment'
  }

  return {
    visibleSectionKeys,
    completedSectionKeys,
    recommendedSectionKey,
    primaryActionKey,
  }
}

export const resolveExpandedMobileSection = ({
  expandedSectionKey,
  recommendedSectionKey,
  completedSectionKeys,
  visibleSectionKeys,
}: ResolveExpandedMobileSectionInput) => {
  if (!expandedSectionKey) {
    return recommendedSectionKey
  }

  if (!visibleSectionKeys.includes(expandedSectionKey)) {
    return recommendedSectionKey
  }

  if (!completedSectionKeys.includes(expandedSectionKey)) {
    return expandedSectionKey
  }

  if (expandedSectionKey === recommendedSectionKey) {
    return recommendedSectionKey
  }

  return recommendedSectionKey
}
