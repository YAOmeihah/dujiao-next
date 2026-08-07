<template>
<MobileCheckoutFlow
  :sections="mobileDisplaySections"
  :expanded-section="mobileExpandedSection"
  :top-label="t('checkout.mobile.currentNeeded')"
  :status-text="mobileStatusText"
  :total-text="mobileTotalText"
  :primary-action-label="mobilePrimaryActionLabel"
  :primary-action-disabled="mobilePrimaryActionDisabled"
  :edit-label="t('checkout.mobile.edit')"
  :collapse-label="t('checkout.mobile.collapse')"
  @update:expanded-section="handleMobileSectionChange"
  @primary-action="handleMobilePrimaryAction"
>
  <template #section-items>
    <div class="space-y-3">
      <div
        v-for="item in cartItems"
        :key="cartItemKey(item)"
        class="rounded-xl border p-3"
        :class="itemStockExceeded(item)
          ? 'border-warning/40 bg-warning/10'
          : 'bg-secondary'"
      >
        <div class="flex min-w-0 items-start gap-3">
          <div class="h-14 w-14 shrink-0 overflow-hidden rounded-xl border bg-muted">
            <SmartImage
              :src="checkoutItemImage(item)"
              :alt="getLocalizedText(item.title)"
              img-class="h-full w-full object-cover"
            />
          </div>
          <div class="min-w-0 flex-1">
            <router-link :to="`/products/${item.slug}`" class="line-clamp-2 text-sm font-semibold text-primary hover:underline">
              {{ getLocalizedText(item.title) }}
            </router-link>
            <div class="mt-1 text-xs text-muted-foreground">{{ t('checkout.quantityLabel') }}：{{ item.quantity }}</div>
            <div v-if="itemSkuDisplay(item)" class="mt-1 text-xs text-muted-foreground">{{ t('checkout.skuLabel') }}：{{ itemSkuDisplay(item) }}</div>
            <div
              v-if="itemStockHint(item)"
              class="mt-1 text-xs"
              :class="itemStockExceeded(item) ? 'text-warning' : 'text-muted-foreground'"
            >
              {{ itemStockHint(item) }}
            </div>
            <div class="mt-2 flex flex-wrap items-baseline gap-3">
              <span
                class="inline-flex items-baseline whitespace-nowrap"
                :class="checkoutItemHasPriceDiscount(item) ? 'text-rose-600 dark:text-rose-300' : 'text-foreground'"
              >
                <span class="text-xl font-black leading-none">{{ checkoutItemPriceParts(item).integer }}</span>
                <span class="text-xs font-semibold">{{ checkoutItemPriceParts(item).decimal }}</span>
                <span class="ml-1 text-xs font-semibold">{{ checkoutItemCurrency }}</span>
                <span v-if="checkoutItemHasPriceDiscount(item)" class="ml-1.5 text-xs font-normal">
                  {{ t('checkout.discountedPriceLabel') }}
                </span>
              </span>
              <span
                v-if="checkoutItemHasPriceDiscount(item)"
                class="inline-flex items-baseline whitespace-nowrap text-xs text-muted-foreground line-through"
              >
                <span>{{ checkoutItemOriginalPriceParts(item).integer }}</span>
                <span>{{ checkoutItemOriginalPriceParts(item).decimal }}</span>
                <span class="ml-1">{{ checkoutItemCurrency }}</span>
              </span>
            </div>
          </div>
        </div>
      </div>
    </div>
  </template>

  <template #section-shipping>
    <div v-if="orderRequiresShippingAddress" class="space-y-3">
      <GuestShippingAddressRecallCard
        v-if="showGuestShippingRecallCard"
        :summary-lines="guestShippingRecallSummaryLines"
        :title="t('checkout.guestShippingRecallTitle')"
        :use-label="t('checkout.guestShippingRecallUse')"
        :rewrite-label="t('checkout.guestShippingRecallRewrite')"
        :applied-message="t('checkout.guestShippingRecallApplied')"
        :clear-form-label="t('checkout.guestShippingRecallClearForm')"
        :clear-record-label="t('checkout.guestShippingRecallClearRecord')"
        :applied="guestShippingRecallApplied"
        :muted="guestShippingRecallMuted"
        @use="applyGuestShippingRecall"
        @rewrite="handleGuestShippingRewrite"
        @clear-form="handleGuestShippingClearForm"
        @clear-record="handleGuestShippingClearRecord"
      />
      <input
        v-model="shippingAddress.receiver_name"
        data-mobile-shipping-input="receiver-name"
        type="text"
        autocomplete="name"
        class="w-full form-input-lg"
        :placeholder="t('checkout.shippingReceiverName')"
      />
      <input
        v-model="shippingAddress.receiver_phone"
        data-mobile-shipping-input="receiver-phone"
        type="tel"
        autocomplete="tel"
        class="w-full form-input-lg"
        :placeholder="t('checkout.shippingReceiverPhone')"
      />
      <RegionSelector
        v-model="shippingAddress"
        data-mobile-shipping-input="region"
        :invalid="submitAttempted && shippingRegionMissing"
      />
      <textarea
        v-model="shippingAddress.detail_address"
        data-mobile-shipping-input="detail-address"
        rows="3"
        autocomplete="street-address"
        class="w-full form-input-lg"
        :placeholder="t('checkout.shippingDetailAddress')"
      />
    </div>
  </template>

  <template #section-buyer>
    <div class="space-y-4">
      <CheckoutManualForm
        :manual-form-products="manualFormProducts"
        v-model="manualFormData"
        :submit-attempted="submitAttempted"
        :embedded="true"
        :compact="true"
        :get-manual-field-label="getManualFieldLabel"
        :get-manual-field-placeholder="getManualFieldPlaceholder"
        :manual-field-error="manualFieldError"
      />

      <template v-if="!userAuthStore.isAuthenticated">
        <div data-mobile-buyer-input="checkout-mode" class="flex flex-wrap gap-3">
          <button
            type="button"
            class="theme-btn-inline-md"
            :class="checkoutMode === 'guest'
              ? 'theme-btn-primary border border-transparent'
              : 'border theme-btn-secondary'"
            @click="checkoutMode = 'guest'"
          >
            {{ t('checkout.guestPurchase') }}
          </button>
          <router-link to="/auth/login" class="theme-btn-inline-md border theme-btn-secondary">
            {{ t('checkout.memberPurchase') }}
          </router-link>
        </div>

        <form
          v-if="checkoutMode === 'guest'"
          class="space-y-3"
          novalidate
          @submit.prevent
        >
          <div class="grid grid-cols-1 gap-3">
            <input
              :value="guestPhone"
              data-mobile-buyer-input="guest-phone"
              type="tel"
              autocomplete="tel"
              class="w-full form-input-lg"
              :placeholder="t('checkout.guestPhonePlaceholder')"
              @input="handleGuestPhoneInput"
            />
            <input
              v-model="guestPassword"
              data-mobile-buyer-input="guest-password"
              type="password"
              autocomplete="current-password"
              class="w-full form-input-lg"
              :placeholder="t('checkout.guestPasswordPlaceholder')"
            />
            <input
              v-model="guestEmail"
              data-mobile-buyer-input="guest-email"
              type="email"
              autocomplete="email"
              class="w-full form-input-lg"
              :placeholder="t('checkout.guestEmailPlaceholder')"
            />
          </div>

          <div
            v-if="guestCaptchaEnabled"
            data-mobile-buyer-input="guest-captcha"
            class="space-y-2"
          >
            <p class="text-xs font-semibold text-muted-foreground">{{ t('auth.common.captchaLabel') }}</p>
            <ImageCaptcha
              v-if="captchaProvider === 'image'"
              ref="guestImageCaptchaRef"
              v-model="guestCaptchaPayload"
              :disabled="submitting"
              @config-stale="handleGuestCaptchaConfigStale"
            />
            <TurnstileCaptcha
              v-else-if="captchaProvider === 'turnstile'"
              ref="guestTurnstileRef"
              v-model="guestTurnstileToken"
              :site-key="guestTurnstileSiteKey"
            />
            <CapCaptcha
              v-else-if="captchaProvider === 'cap'"
              ref="guestCapRef"
              v-model="guestCapToken"
              :endpoint="guestCapEndpoint"
              :site-key="guestCapSiteKey"
            />
          </div>

          <div class="rounded-xl border border-success/30 bg-success/10 p-3 text-sm text-foreground">
            <p class="font-semibold">{{ t('checkout.guestInstructions.title') }}</p>
            <p v-if="orderRequiresShippingAddress" class="mt-2">{{ t('checkout.guestPhoneSyncHint') }}</p>
            <p class="mt-2">{{ t('checkout.guestInstructions.password') }}</p>
            <p class="mt-2">{{ t('checkout.guestInstructions.email') }}</p>
          </div>
        </form>

        <p v-if="checkoutMode === 'guest' && guestPhone && !guestPhoneValid" class="text-xs text-destructive">
          {{ t('error.phone_invalid') }}
        </p>
        <p v-if="checkoutMode === 'guest' && guestEmail && !guestEmailValid" class="text-xs text-destructive">
          {{ t('error.email_invalid') }}
        </p>
        <p
          v-if="checkoutMode === 'guest' && guestCaptchaEnabled && submitAttempted && !guestCaptchaComplete"
          class="text-xs text-destructive"
        >
          {{ t('auth.common.captchaRequired') }}
        </p>
      </template>
    </div>
  </template>

  <template v-if="!isResellerTenant" #section-coupon>
    <div class="space-y-3">
      <input
        v-model="couponCode"
        type="text"
        class="w-full form-input-lg"
        :placeholder="t('checkout.couponPlaceholder')"
      />
      <p v-if="previewLoading || couponRefreshing" class="text-xs text-muted-foreground">
        {{ previewStatusText }}
      </p>
    </div>
  </template>

  <template #section-payment>
    <div class="space-y-3">
      <div v-if="showBalanceOption" class="rounded-lg border bg-secondary p-3">
        <div class="flex items-start justify-between gap-3">
          <div class="min-w-0">
            <div class="text-xs text-muted-foreground">{{ t('payment.walletBalanceLabel') }}</div>
            <div class="mt-1 text-sm font-semibold text-foreground">
              {{ walletLoading ? t('common.loading') : formatPrice(walletBalance, previewCurrency) }}
            </div>
          </div>
          <label class="inline-flex items-center gap-2 text-xs text-muted-foreground">
            <input v-model="useBalance" type="checkbox" class="h-4 w-4 accent-primary" :disabled="walletOnlyPayment" />
            <span>{{ t('payment.useBalance') }}</span>
          </label>
        </div>
        <div v-if="walletOnlyPayment" class="mt-2 text-xs text-warning">
          {{ t('payment.walletOnlyHint') }}
        </div>
        <div v-if="useBalance" class="mt-2 space-y-1 text-xs text-muted-foreground">
          <div>{{ t('payment.walletDeductLabel') }}：{{ expectedWalletPaidDisplay }}</div>
          <div v-if="!walletOnlyPayment">{{ t('payment.onlinePayLabel') }}：{{ expectedOnlinePayDisplay }}</div>
          <div v-if="walletOnlyPayment && expectedOnlinePayCents > 0" class="text-warning">
            {{ t('payment.walletInsufficientHint') }}
          </div>
        </div>
      </div>

      <template v-if="!walletOnlyPayment">
        <div
          v-if="requiresOnlineChannel && paymentChannels.length > 0"
          data-mobile-payment-input="channel-list"
          class="space-y-2"
        >
          <button
            v-for="channel in paymentChannels"
            :key="channel.id"
            type="button"
            class="w-full rounded-lg border px-3 py-3 text-left transition-colors disabled:cursor-not-allowed disabled:opacity-60"
            :class="selectedChannelId === channel.id && !isChannelDisabledForAmount(channel) ? 'theme-selected-surface' : 'theme-interactive-surface'"
            :disabled="isChannelDisabledForAmount(channel)"
            :title="isChannelDisabledForAmount(channel) ? channelAmountLimitHint(channel) : ''"
            @click="handleSelectChannel(channel)"
          >
            <div class="flex items-center justify-between gap-3">
              <div class="min-w-0">
                <div class="text-sm font-medium text-foreground">{{ channel.name }}</div>
                <div class="mt-1 text-xs text-muted-foreground">
                  {{ t('payment.feeLabel') }}：{{ formatChannelFeeRate(channel) }}
                </div>
              </div>
              <div class="text-xs text-muted-foreground">
                {{ formatChannelFixedFee(channel) }}
              </div>
            </div>
            <div v-if="isChannelDisabledForAmount(channel)" class="mt-2 text-xs text-warning">
              {{ channelAmountLimitHint(channel) }}
            </div>
          </button>
        </div>
        <div v-else-if="requiresOnlineChannel && paymentChannels.length === 0" class="text-xs text-muted-foreground">
          {{ t('checkout.noPaymentChannels') }}
        </div>
      </template>

      <div v-if="!requiresOnlineChannel" class="text-xs text-success">
        {{ t('checkout.walletCoversAll') }}
      </div>
      <p v-if="selectedChannelAmountHint" class="text-xs text-warning">
        {{ selectedChannelAmountHint }}
      </p>
    </div>
  </template>
</MobileCheckoutFlow>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import ImageCaptcha from '../../captcha/ImageCaptcha.vue'
import TurnstileCaptcha from '../../captcha/TurnstileCaptcha.vue'
import CapCaptcha from '../../captcha/CapCaptcha.vue'
import CheckoutManualForm from '../CheckoutManualForm.vue'
import GuestShippingAddressRecallCard from '../GuestShippingAddressRecallCard.vue'
import RegionSelector from '../RegionSelector.vue'
import SmartImage from '../../SmartImage.vue'
import MobileCheckoutFlow from './MobileCheckoutFlow.vue'
import type { useCheckout } from '../../../composables/useCheckout'
import { useMobileCheckoutAdapter } from '../../../composables/useMobileCheckoutAdapter'

type CheckoutContract = ReturnType<typeof useCheckout>

const props = defineProps<{
  checkout: CheckoutContract
}>()

const { t } = useI18n()
const checkout = props.checkout

const {
  userAuthStore, getLocalizedText, formatPrice,
  cartItems, cartItemKey, checkoutItemImage, itemSkuDisplay, itemStockExceeded, itemStockHint,
  checkoutItemCurrency, checkoutItemPriceParts, checkoutItemOriginalPriceParts, checkoutItemHasPriceDiscount,
  manualFormProducts, manualFormData, submitAttempted, getManualFieldLabel, getManualFieldPlaceholder, manualFieldError,
  couponCode, isResellerTenant, checkoutMode, guestPhone, guestPhoneValid, handleGuestPhoneInput,
  guestEmail, guestPassword, guestEmailValid, guestCaptchaEnabled, captchaProvider, guestCaptchaPayload,
  guestTurnstileToken, guestTurnstileSiteKey, guestCapToken, guestCapSiteKey, guestCapEndpoint,
  guestImageCaptchaRef, guestTurnstileRef, guestCapRef, handleGuestCaptchaConfigStale,
  shippingAddress, orderRequiresShippingAddress, shippingRegionMissing,
  showGuestShippingRecallCard, guestShippingRecallSummaryLines, guestShippingRecallApplied, guestShippingRecallMuted,
  applyGuestShippingRecall, handleGuestShippingRewrite, handleGuestShippingClearForm, handleGuestShippingClearRecord,
  previewCurrency, previewLoading, couponRefreshing, previewStatusText, showBalanceOption, walletLoading, walletBalance,
  useBalance, walletOnlyPayment, expectedWalletPaidDisplay, expectedOnlinePayDisplay, expectedOnlinePayCents,
  requiresOnlineChannel, paymentChannels, selectedChannelId, isChannelDisabledForAmount, channelAmountLimitHint,
  handleSelectChannel, formatChannelFeeRate, formatChannelFixedFee, submitting,
} = checkout

const { selectedChannelAmountHint } = checkout.mobileCheckout
const {
  guestCaptchaComplete, mobileDisplaySections, mobileExpandedSection, mobilePrimaryActionDisabled,
  mobilePrimaryActionLabel, mobileStatusText, mobileTotalText, handleMobilePrimaryAction, handleMobileSectionChange,
} = useMobileCheckoutAdapter(checkout)

void guestImageCaptchaRef
void guestTurnstileRef
void guestCapRef
</script>
