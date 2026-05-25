<template>
  <div class="min-h-screen theme-page pt-24 pb-16">
    <div class="container mx-auto px-4">
      <div class="mb-8">
        <h1 class="mb-2 text-3xl font-black theme-text-primary">{{ t('checkout.title') }}</h1>
        <p class="text-sm theme-text-secondary">{{ t('checkout.subtitle') }}</p>
      </div>

      <CheckoutSteps
        class="mb-8"
        current-step="checkout"
        :step-keys="isBuyNowMode ? ['checkout', 'payment'] : ['cart', 'checkout', 'payment']"
      />

      <EmptyState
        v-if="cartItems.length === 0"
        icon="cart"
        :title="t('checkout.empty')"
        :action-label="t('checkout.emptyAction')"
        action-to="/products"
      />

      <div v-else>
        <MobileCheckoutFlow
          class="lg:hidden"
          :sections="mobileDisplaySections"
          :expanded-section="mobileExpandedSection"
          :top-label="t('checkout.mobile.currentNeeded')"
          :status-text="mobileStatusText"
          :total-text="mobileTotalText"
          :primary-action-label="mobilePrimaryActionLabel"
          :primary-action-disabled="submitting || syncingStock"
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
                  ? 'border-amber-200 bg-amber-50/60 dark:border-amber-700 dark:bg-amber-950/20'
                  : 'border-gray-100 bg-gray-50 dark:border-white/10 dark:bg-black/20'"
              >
                <div class="flex min-w-0 items-start gap-3">
                  <div class="h-14 w-14 shrink-0 overflow-hidden rounded-xl border border-gray-200 bg-white dark:border-white/10 dark:bg-black/30">
                    <SmartImage
                      :src="checkoutItemImage(item)"
                      :alt="getLocalizedText(item.title)"
                      img-class="h-full w-full object-cover"
                    />
                  </div>
                  <div class="min-w-0 flex-1">
                    <router-link :to="`/products/${item.slug}`" class="line-clamp-2 text-sm font-semibold theme-link">
                      {{ getLocalizedText(item.title) }}
                    </router-link>
                    <div class="mt-1 text-xs theme-text-muted">{{ t('checkout.quantityLabel') }}：{{ item.quantity }}</div>
                    <div v-if="itemSkuDisplay(item)" class="mt-1 text-xs theme-text-muted">{{ t('checkout.skuLabel') }}：{{ itemSkuDisplay(item) }}</div>
                    <div
                      v-if="itemStockHint(item)"
                      class="mt-1 text-xs"
                      :class="itemStockExceeded(item)
                        ? 'text-amber-600 dark:text-amber-300'
                        : 'theme-text-muted'"
                    >
                      {{ itemStockHint(item) }}
                    </div>
                    <div class="mt-2 flex flex-wrap items-baseline gap-3">
                      <span
                        class="inline-flex items-baseline whitespace-nowrap"
                        :class="checkoutItemHasPriceDiscount(item) ? 'text-rose-600 dark:text-rose-300' : 'theme-text-primary'"
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
                        class="inline-flex items-baseline whitespace-nowrap text-xs text-gray-400 dark:text-gray-500"
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
                    @click="checkoutMode = 'guest'"
                    class="theme-btn-inline-md"
                    :class="checkoutMode === 'guest'
                      ? 'theme-btn-primary border border-transparent'
                      : 'border theme-btn-secondary'"
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
                    <p class="text-xs font-semibold uppercase tracking-[0.14em] theme-text-muted">{{ t('auth.common.captchaLabel') }}</p>
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

                  <div class="rounded-xl border border-green-200 bg-green-50 p-3 text-sm text-green-900">
                    <p class="font-semibold">{{ t('checkout.guestInstructions.title') }}</p>
                    <p v-if="orderRequiresShippingAddress" class="mt-2">{{ t('checkout.guestPhoneSyncHint') }}</p>
                    <p class="mt-2">{{ t('checkout.guestInstructions.password') }}</p>
                    <p class="mt-2">{{ t('checkout.guestInstructions.email') }}</p>
                  </div>
                </form>

                <p v-if="checkoutMode === 'guest' && guestPhone && !guestPhoneValid" class="text-xs text-red-500">
                  {{ t('error.phone_invalid') }}
                </p>
                <p v-if="checkoutMode === 'guest' && guestEmail && !guestEmailValid" class="text-xs text-red-500">
                  {{ t('error.email_invalid') }}
                </p>
                <p
                  v-if="checkoutMode === 'guest' && guestCaptchaEnabled && submitAttempted && !guestCaptchaComplete"
                  class="text-xs text-red-500"
                >
                  {{ t('auth.common.captchaRequired') }}
                </p>
              </template>
            </div>
          </template>

          <template #section-coupon>
            <div class="space-y-3">
              <input
                v-model="couponCode"
                type="text"
                class="w-full form-input-lg"
                :placeholder="t('checkout.couponPlaceholder')"
              />
              <p v-if="previewLoading || couponRefreshing" class="text-xs theme-text-muted">
                {{ previewStatusText }}
              </p>
            </div>
          </template>

          <template #section-payment>
            <div class="space-y-3">
              <div v-if="showBalanceOption" class="rounded-lg border theme-surface-soft p-3">
                <div class="flex items-start justify-between gap-3">
                  <div class="min-w-0">
                    <div class="text-xs theme-text-muted">{{ t('payment.walletBalanceLabel') }}</div>
                    <div class="mt-1 text-sm font-semibold theme-text-primary">
                      {{ walletLoading ? t('common.loading') : formatPrice(walletBalance, previewCurrency) }}
                    </div>
                  </div>
                  <label class="inline-flex items-center gap-2 text-xs theme-text-secondary">
                    <input v-model="useBalance" type="checkbox" class="h-4 w-4 accent-primary" :disabled="walletOnlyPayment" />
                    <span>{{ t('payment.useBalance') }}</span>
                  </label>
                </div>
                <div v-if="walletOnlyPayment" class="mt-2 text-xs text-amber-600 dark:text-amber-400">
                  {{ t('payment.walletOnlyHint') }}
                </div>
                <div v-if="useBalance" class="mt-2 space-y-1 text-xs theme-text-muted">
                  <div>{{ t('payment.walletDeductLabel') }}：{{ expectedWalletPaidDisplay }}</div>
                  <div v-if="!walletOnlyPayment">{{ t('payment.onlinePayLabel') }}：{{ expectedOnlinePayDisplay }}</div>
                  <div v-if="walletOnlyPayment && expectedOnlinePayCents > 0" class="text-amber-600 dark:text-amber-400">
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
                        <div class="text-sm font-medium theme-text-primary">{{ channel.name }}</div>
                        <div class="mt-1 text-xs theme-text-muted">
                          {{ t('payment.feeLabel') }}：{{ formatChannelFeeRate(channel) }}
                        </div>
                      </div>
                      <div class="text-xs theme-text-muted">
                        {{ formatChannelFixedFee(channel) }}
                      </div>
                    </div>
                    <div v-if="isChannelDisabledForAmount(channel)" class="mt-2 text-xs text-amber-600">
                      {{ channelAmountLimitHint(channel) }}
                    </div>
                  </button>
                </div>
                <div v-else-if="requiresOnlineChannel && paymentChannels.length === 0" class="text-xs theme-text-muted">
                  {{ t('checkout.noPaymentChannels') }}
                </div>
              </template>

              <div v-if="!requiresOnlineChannel" class="text-xs text-emerald-600 dark:text-emerald-400">
                {{ t('checkout.walletCoversAll') }}
              </div>
              <p v-if="selectedChannelAmountHint" class="text-xs text-amber-600 dark:text-amber-300">
                {{ selectedChannelAmountHint }}
              </p>
            </div>
          </template>
        </MobileCheckoutFlow>

        <div class="hidden grid-cols-1 gap-8 lg:grid lg:grid-cols-3">
          <div class="space-y-6 lg:col-span-2">
            <div class="rounded-2xl border theme-panel p-6">
              <h2 class="mb-4 text-lg font-bold theme-text-primary">{{ t('checkout.itemsTitle') }}</h2>
              <div class="space-y-4">
                <div
                  v-for="item in cartItems"
                  :key="cartItemKey(item)"
                  class="rounded-xl border p-4"
                  :class="itemStockExceeded(item)
                    ? 'border-amber-200 bg-amber-50/60 dark:border-amber-700 dark:bg-amber-950/20'
                    : 'border-gray-100 bg-gray-50 dark:border-white/10 dark:bg-black/20'"
                >
                  <div class="flex items-start justify-between gap-4">
                    <div class="flex min-w-0 items-start gap-3">
                      <div class="h-16 w-16 shrink-0 overflow-hidden rounded-xl border border-gray-200 bg-white transition-all duration-200 hover:-translate-y-0.5 hover:shadow-sm dark:border-white/10 dark:bg-black/30 sm:h-20 sm:w-20">
                        <img
                          v-if="checkoutItemImage(item)"
                          :src="checkoutItemImage(item)"
                          :alt="getLocalizedText(item.title)"
                          loading="lazy"
                          decoding="async"
                          class="h-full w-full object-cover"
                        />
                        <div v-else class="flex h-full w-full items-center justify-center theme-text-muted">
                          <svg class="h-6 w-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                            <path
                              stroke-linecap="round"
                              stroke-linejoin="round"
                              stroke-width="1.5"
                              d="M4 16l4.586-4.586a2 2 0 012.828 0L16 16m-2-2l1.586-1.586a2 2 0 012.828 0L20 14m-6-6h.01M6 20h12a2 2 0 002-2V6a2 2 0 00-2-2H6a2 2 0 00-2 2v12a2 2 0 002 2z"
                            />
                          </svg>
                        </div>
                      </div>
                      <div class="min-w-0">
                        <router-link
                          :to="`/products/${item.slug}`"
                          class="line-clamp-2 font-semibold theme-link"
                        >
                          {{ getLocalizedText(item.title) }}
                        </router-link>
                        <div class="mt-1 text-xs theme-text-muted">{{ t('checkout.quantityLabel') }}：{{ item.quantity }}</div>
                        <div v-if="itemSkuDisplay(item)" class="mt-1 text-xs theme-text-muted">{{ t('checkout.skuLabel') }}：{{ itemSkuDisplay(item) }}</div>
                        <div
                          v-if="itemStockHint(item)"
                          class="mt-1 text-xs"
                          :class="itemStockExceeded(item)
                            ? 'text-amber-600 dark:text-amber-300'
                            : 'theme-text-muted'"
                        >
                          {{ itemStockHint(item) }}
                        </div>
                        <div class="mt-2 flex flex-wrap gap-2">
                          <span
                            class="theme-badge text-xs uppercase tracking-wider"
                            :class="item.purchaseType === 'guest'
                              ? 'theme-badge-warning'
                              : 'theme-badge-success'"
                          >
                            {{ item.purchaseType === 'guest' ? t('productPurchase.guest') : t('productPurchase.member') }}
                          </span>
                          <span
                            class="theme-badge text-xs uppercase tracking-wider"
                            :class="item.fulfillmentType === 'auto'
                              ? 'theme-badge-info'
                              : 'theme-badge-neutral'"
                          >
                            {{ item.fulfillmentType === 'auto' ? t('products.fulfillmentType.auto') : t('products.fulfillmentType.manual') }}
                          </span>
                        </div>
                      </div>
                    </div>
                    <div class="text-right">
                      <div class="text-xs uppercase tracking-wider theme-text-muted">{{ t('checkout.previewTotal') }}</div>
                      <div
                        class="mt-1 inline-flex items-baseline whitespace-nowrap"
                        :class="checkoutItemHasPriceDiscount(item) ? 'text-rose-600 dark:text-rose-300' : 'theme-text-primary'"
                      >
                        <span class="text-xl font-black leading-none">{{ checkoutItemPriceParts(item).integer }}</span>
                        <span class="text-xs font-semibold">{{ checkoutItemPriceParts(item).decimal }}</span>
                        <span class="ml-1 text-xs font-semibold">{{ checkoutItemCurrency }}</span>
                      </div>
                      <div
                        v-if="checkoutItemHasPriceDiscount(item)"
                        class="mt-1 inline-flex items-baseline whitespace-nowrap text-xs text-gray-400 dark:text-gray-500"
                      >
                        <span>{{ checkoutItemOriginalPriceParts(item).integer }}</span>
                        <span>{{ checkoutItemOriginalPriceParts(item).decimal }}</span>
                        <span class="ml-1">{{ checkoutItemCurrency }}</span>
                      </div>
                    </div>
                  </div>
                </div>
              </div>
            </div>

            <CheckoutManualForm
              :manual-form-products="manualFormProducts"
              v-model="manualFormData"
              :submit-attempted="submitAttempted"
              :get-manual-field-label="getManualFieldLabel"
              :get-manual-field-placeholder="getManualFieldPlaceholder"
              :manual-field-error="manualFieldError"
            />

            <div v-if="orderRequiresShippingAddress" class="rounded-2xl border theme-panel p-6">
              <div class="mb-4">
                <h2 class="text-lg font-bold theme-text-primary">{{ t('checkout.shippingTitle') }}</h2>
                <p class="mt-1 text-sm theme-text-muted">{{ t('checkout.shippingTip') }}</p>
              </div>
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
              <div class="grid grid-cols-1 gap-4 md:grid-cols-2">
                <input
                  v-model="shippingAddress.receiver_name"
                  type="text"
                  autocomplete="name"
                  class="w-full form-input-lg"
                  :placeholder="t('checkout.shippingReceiverName')"
                />
                <input
                  v-model="shippingAddress.receiver_phone"
                  type="tel"
                  autocomplete="tel"
                  class="w-full form-input-lg"
                  :placeholder="t('checkout.shippingReceiverPhone')"
                />
                <div class="md:col-span-2">
                  <RegionSelector
                    v-model="shippingAddress"
                    :invalid="submitAttempted && shippingRegionMissing"
                  />
                </div>
                <textarea
                  v-model="shippingAddress.detail_address"
                  rows="3"
                  autocomplete="street-address"
                  class="w-full form-input-lg md:col-span-2"
                  :placeholder="t('checkout.shippingDetailAddress')"
                />
              </div>
              <p v-if="submitAttempted && !shippingAddressValidation.valid" class="mt-3 text-sm text-red-500">
                {{ shippingAddressValidation.message }}
              </p>
            </div>

            <div class="rounded-2xl border theme-panel p-6">
              <h2 class="mb-4 text-lg font-bold theme-text-primary">{{ t('checkout.couponTitle') }}</h2>
              <input
                v-model="couponCode"
                type="text"
                class="w-full form-input-lg"
                :placeholder="t('checkout.couponPlaceholder')"
              />
            </div>

            <div
              v-if="!userAuthStore.isAuthenticated"
              class="space-y-4 rounded-2xl border theme-panel p-6"
            >
              <h2 class="text-lg font-bold theme-text-primary">{{ t('checkout.modeTitle') }}</h2>
              <div class="flex flex-wrap gap-3">
                <button
                  @click="checkoutMode = 'guest'"
                  class="theme-btn-inline-md"
                  :class="checkoutMode === 'guest'
                    ? 'theme-btn-primary border border-transparent'
                    : 'border theme-btn-secondary'"
                >
                  {{ t('checkout.guestPurchase') }}
                </button>
                <router-link
                  to="/auth/login"
                  class="theme-btn-inline-md border theme-btn-secondary"
                >
                  {{ t('checkout.memberPurchase') }}
                </router-link>
              </div>

              <form
                v-if="checkoutMode === 'guest'"
                class="space-y-4"
                novalidate
                @submit.prevent
              >
                <div class="grid grid-cols-1 gap-4 md:grid-cols-2">
                  <input
                    :value="guestPhone"
                    type="tel"
                    autocomplete="tel"
                    class="w-full form-input-lg"
                    :placeholder="t('checkout.guestPhonePlaceholder')"
                    @input="handleGuestPhoneInput"
                  />
                  <input
                    v-model="guestPassword"
                    type="password"
                    autocomplete="current-password"
                    class="w-full form-input-lg"
                    :placeholder="t('checkout.guestPasswordPlaceholder')"
                  />
                  <input
                    v-model="guestEmail"
                    type="email"
                    autocomplete="email"
                    class="w-full form-input-lg md:col-span-2"
                    :placeholder="t('checkout.guestEmailPlaceholder')"
                  />
                </div>

                <div v-if="guestCaptchaEnabled" class="space-y-2">
                  <p class="text-xs font-semibold uppercase tracking-[0.14em] theme-text-muted">{{ t('auth.common.captchaLabel') }}</p>
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

                <div class="mb-3 rounded-xl border border-green-200 bg-green-50 p-3 text-sm text-green-900">
                  <p class="font-semibold">{{ t('checkout.guestInstructions.title') }}</p>
                  <ul class="mt-2 space-y-1 list-disc pl-5">
                    <li v-if="orderRequiresShippingAddress">{{ t('checkout.guestPhoneSyncHint') }}</li>
                    <li>{{ t('checkout.guestInstructions.password') }}</li>
                    <li>{{ t('checkout.guestInstructions.email') }}</li>
                  </ul>
                </div>
              </form>
              <p v-if="checkoutMode === 'guest' && guestPhone && !guestPhoneValid" class="text-xs text-red-500">
                {{ t('error.phone_invalid') }}
              </p>
              <p v-if="checkoutMode === 'guest' && guestEmail && !guestEmailValid" class="text-xs text-red-500">
                {{ t('error.email_invalid') }}
              </p>
            </div>
          </div>

          <div class="h-fit rounded-2xl border theme-panel p-6 lg:sticky lg:top-24">
            <h2 class="mb-4 text-lg font-bold theme-text-primary">{{ t('checkout.submitTitle') }}</h2>
            <div class="mb-4 rounded-lg border theme-surface-soft p-3 text-xs theme-text-muted">
              {{ t('checkout.submitHint') }}
            </div>

            <div class="mb-4 space-y-3 text-sm theme-text-muted">
              <div class="flex items-center justify-between">
                <span>{{ t('cart.itemsCount') }}</span>
                <span class="font-mono theme-text-primary">{{ totalItems }}</span>
              </div>
              <div class="flex items-center justify-between">
                <span>{{ t('checkout.previewOriginal') }}</span>
                <span class="font-mono theme-text-primary">{{ formatPrice(previewOriginal, previewCurrency) }}</span>
              </div>
              <div class="flex items-center justify-between">
                <span>{{ t('checkout.previewCoupon') }}</span>
                <span class="font-mono theme-text-primary">{{ formatPrice(previewCoupon, previewCurrency) }}</span>
              </div>
              <div class="flex items-center justify-between">
                <span>{{ t('checkout.previewPromotion') }}</span>
                <span class="font-mono theme-text-primary">{{ formatPrice(previewPromotion, previewCurrency) }}</span>
              </div>
              <div v-if="Number(previewMemberDiscount) > 0" class="flex items-center justify-between">
                <span>{{ t('checkout.previewMemberDiscount') }}</span>
                <span class="font-mono text-amber-600 dark:text-amber-300">-{{ formatPrice(previewMemberDiscount, previewCurrency) }}</span>
              </div>
              <div class="flex items-center justify-between border-t theme-border pt-3 theme-text-primary">
                <span class="font-semibold">{{ t('checkout.previewTotal') }}</span>
                <span class="font-mono text-lg font-bold">{{ formatPrice(previewTotal, previewCurrency) }}</span>
              </div>
            </div>

            <div v-if="previewLoading || couponRefreshing" class="mb-3 text-xs theme-text-muted">
              {{ previewStatusText }}
            </div>
            <div
              v-if="checkoutAlert"
              class="mb-4 rounded-lg border p-3 text-sm"
              :class="pageAlertClass(checkoutAlert.level)"
            >
              {{ checkoutAlert.message }}
            </div>

            <div class="mb-4 border-t theme-border pt-4">
              <h3 class="mb-3 text-sm font-bold theme-text-primary">{{ t('checkout.paymentMethod') }}</h3>

              <div v-if="showBalanceOption" class="mb-3 rounded-lg border theme-surface-soft p-3">
                <div class="flex items-center justify-between">
                  <div>
                    <div class="text-xs theme-text-muted">{{ t('payment.walletBalanceLabel') }}</div>
                    <div class="mt-0.5 text-sm font-semibold theme-text-primary">
                      {{ walletLoading ? t('common.loading') : formatPrice(walletBalance, previewCurrency) }}
                    </div>
                  </div>
                  <label class="inline-flex items-center gap-2 text-xs theme-text-secondary">
                    <input v-model="useBalance" type="checkbox" class="h-4 w-4 accent-primary" :disabled="walletOnlyPayment" />
                    <span>{{ t('payment.useBalance') }}</span>
                  </label>
                </div>
                <div v-if="walletOnlyPayment" class="mt-2 text-xs text-amber-600 dark:text-amber-400">
                  {{ t('payment.walletOnlyHint') }}
                </div>
                <div v-if="useBalance" class="mt-2 space-y-1 text-xs theme-text-muted">
                  <div>{{ t('payment.walletDeductLabel') }}：{{ expectedWalletPaidDisplay }}</div>
                  <div v-if="!walletOnlyPayment">{{ t('payment.onlinePayLabel') }}：{{ expectedOnlinePayDisplay }}</div>
                  <div v-if="walletOnlyPayment && expectedOnlinePayCents > 0" class="text-amber-600 dark:text-amber-400">
                    {{ t('payment.walletInsufficientHint') }}
                  </div>
                </div>
              </div>

              <template v-if="!walletOnlyPayment">
                <div v-if="requiresOnlineChannel && paymentChannels.length > 0" class="grid grid-cols-2 gap-2">
                  <button v-for="channel in paymentChannels" :key="channel.id"
                    type="button"
                    :disabled="isChannelDisabledForAmount(channel)"
                    :title="isChannelDisabledForAmount(channel) ? channelAmountLimitHint(channel) : ''"
                    @click="handleSelectChannel(channel)"
                    class="text-left border rounded-lg p-2.5 transition-colors disabled:cursor-not-allowed disabled:opacity-60"
                    :class="selectedChannelId === channel.id && !isChannelDisabledForAmount(channel) ? 'theme-selected-surface' : 'theme-interactive-surface'">
                    <div class="flex items-center gap-2">
                      <img v-if="channel.icon" :src="getImageUrl(channel.icon)" loading="lazy" class="h-5 w-5 rounded object-contain shrink-0" />
                      <div class="text-sm theme-text-primary font-medium truncate">{{ channel.name }}</div>
                    </div>
                    <div class="mt-1 space-y-0.5 text-xs theme-text-muted">
                      <div>{{ t('payment.feeLabel') }}：{{ formatChannelFeeRate(channel) }}</div>
                      <div>{{ t('payment.fixedFeeLabel') }}：{{ formatChannelFixedFee(channel) }}</div>
                    </div>
                    <div v-if="isChannelDisabledForAmount(channel)" class="mt-1 text-xs text-amber-600">
                      {{ channelAmountLimitHint(channel) }}
                    </div>
                  </button>
                </div>
                <div v-else-if="requiresOnlineChannel && paymentChannels.length === 0" class="text-xs theme-text-muted">
                  {{ t('checkout.noPaymentChannels') }}
                </div>
              </template>
              <div v-if="!requiresOnlineChannel" class="text-xs text-emerald-600 dark:text-emerald-400">
                {{ t('checkout.walletCoversAll') }}
              </div>
            </div>

            <button
              @click="handleSubmit"
              :disabled="!canSubmit"
              class="theme-btn-block-md theme-btn-primary font-semibold transition-colors disabled:cursor-not-allowed disabled:opacity-50"
            >
              {{ submitting ? t('checkout.submitting') : t('checkout.submitButton') }}
            </button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, nextTick, onMounted, onUnmounted, ref, watch } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { useCartStore, type CartItem } from '../stores/cart'
import { useBuyNowStore } from '../stores/buyNow'
import { useAppStore } from '../stores/app'
import { useUserAuthStore } from '../stores/userAuth'
import { guestOrderAPI, userOrderAPI, walletAPI, type CaptchaPayload } from '../api'
import { debounceAsync } from '../utils/debounce'
import { pageAlertClass, type PageAlert } from '../utils/alerts'
import { amountToCents, basisPointsToPercent, centsToAmount, parseInteger, rateToBasisPoints } from '../utils/money'
import { buildSkuDisplayText, normalizeSkuId } from '../utils/sku'
import { refreshCartStockSnapshots, cartItemPurchaseLimit as itemPurchaseLimit, cartItemPurchaseMin as itemPurchaseMin } from '../utils/cartStock'
import { getImageUrl } from '../utils/image'
import { getAffiliateCode, getAffiliateVisitorKey } from '../utils/affiliate'
import ImageCaptcha from '../components/captcha/ImageCaptcha.vue'
import TurnstileCaptcha from '../components/captcha/TurnstileCaptcha.vue'
import CapCaptcha from '../components/captcha/CapCaptcha.vue'
import CheckoutManualForm from '../components/checkout/CheckoutManualForm.vue'
import EmptyState from '../components/EmptyState.vue'
import SmartImage from '../components/SmartImage.vue'
import CheckoutSteps from '../components/checkout/CheckoutSteps.vue'
import GuestShippingAddressRecallCard from '../components/checkout/GuestShippingAddressRecallCard.vue'
import MobileCheckoutFlow from '../components/checkout/mobile/MobileCheckoutFlow.vue'
import RegionSelector from '../components/checkout/RegionSelector.vue'
import {
  clearGuestShippingAddressRecall,
  createGuestShippingAddressRecallRecord,
  loadGuestShippingAddressRecall,
  shouldEnableGuestShippingAddressRecall,
  type GuestShippingAddressRecallRecord,
} from '../composables/useGuestShippingAddressRecall'
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
} from '../composables/useMobileCheckoutFlow'
import { useLocalized } from '../composables/useProduct'
import type { ShippingAddressFormValue } from '../types/address'

const router = useRouter()
const route = useRoute()
const cartStore = useCartStore()
const buyNowStore = useBuyNowStore()
const appStore = useAppStore()
const userAuthStore = useUserAuthStore()
const { t } = useI18n()

const { getLocalizedText, siteCurrency, formatPrice } = useLocalized()

const isBuyNowMode = computed(() => route.query.mode === 'buynow')
const cartItems = computed<CartItem[]>(() => {
  if (isBuyNowMode.value) {
    return buyNowStore.item ? [buyNowStore.item] : []
  }
  return cartStore.items
})
const totalItems = computed(() => cartItems.value.reduce((sum, item) => sum + item.quantity, 0))
const couponCode = ref('')
const normalizedCouponCode = computed(() => couponCode.value.trim())
const submitting = ref(false)
const error = ref('')
const preview = ref<any>(null)
const previewLoading = ref(false)
const previewError = ref('')
const previewRequestId = ref(0)
const couponRefreshing = ref(false)
const syncingStock = ref(false)
const orderPaymentChannels = ref<any[]>([])
const orderPaymentChannelsRequestId = ref(0)

// Payment state
const selectedChannelId = ref<number | null>(null)
const useBalance = ref(false)
const walletLoading = ref(false)
const walletBalance = ref('0')

// Payment channels
const paymentChannels = computed(() => {
  const list = userAuthStore.isAuthenticated
    ? orderPaymentChannels.value
    : appStore.config?.payment_channels
  if (!Array.isArray(list)) return []
  let filtered = list.filter((channel: any) => {
    const providerType = String(channel?.provider_type || '').toLowerCase()
    const channelType = String(channel?.channel_type || '').toLowerCase()
    if (providerType === 'epay') {
      return ['wechat', 'wxpay', 'alipay', 'qqpay'].includes(channelType)
    }
    return true
  })
  // 按购物车中商品允许的支付渠道交集过滤
  const items = cartItems.value
  if (items.length > 0) {
    let intersectionArr: number[] | null = null
    for (const item of items) {
      const ids = item.paymentChannelIds
      if (!Array.isArray(ids) || ids.length === 0) continue
      const idSet = new Set(ids.map(Number))
      if (intersectionArr === null) {
        intersectionArr = [...idSet]
      } else {
        intersectionArr = intersectionArr.filter((id) => idSet.has(id))
      }
    }
    if (intersectionArr !== null && intersectionArr.length > 0) {
      const allowedSet = new Set(intersectionArr)
      filtered = filtered.filter((ch: any) => allowedSet.has(Number(ch?.id)))
    } else if (intersectionArr !== null) {
      filtered = []
    }
  }
  return filtered
})

const walletOnlyPayment = computed(() => !!appStore.config?.wallet_only_payment)
const showBalanceOption = computed(() => userAuthStore.isAuthenticated)
const expectedWalletPaidCents = computed(() => {
  if (!showBalanceOption.value || !useBalance.value) return 0
  const balance = amountToCents(walletBalance.value)
  const total = amountToCents(previewTotal.value)
  if (balance === null || total === null) return 0
  return Math.min(balance, total)
})
const expectedOnlinePayCents = computed(() => {
  const total = amountToCents(previewTotal.value)
  if (total === null) return 0
  return Math.max(total - expectedWalletPaidCents.value, 0)
})
const expectedWalletPaidDisplay = computed(() => formatPrice(centsToAmount(expectedWalletPaidCents.value), previewCurrency.value))
const expectedOnlinePayDisplay = computed(() => formatPrice(centsToAmount(expectedOnlinePayCents.value), previewCurrency.value))
const requiresOnlineChannel = computed(() => {
  if (!userAuthStore.isAuthenticated) return true
  if (!useBalance.value) return true
  return expectedOnlinePayCents.value > 0
})

const channelLimitMeta = (channel?: any) => {
  const minCents = amountToCents(String(channel?.min_amount ?? ''))
  const maxCents = amountToCents(String(channel?.max_amount ?? ''))
  return {
    minCents,
    maxCents,
    hasMin: minCents !== null && minCents > 0,
    hasMax: maxCents !== null && maxCents > 0,
    hideAmountOutRange: Boolean(channel?.hide_amount_out_range),
  }
}

const isChannelDisabledForAmount = (channel?: any) => {
  if (!requiresOnlineChannel.value) return false
  const targetAmount = expectedOnlinePayCents.value
  if (targetAmount <= 0) return false

  const meta = channelLimitMeta(channel)
  if (!meta.hasMin && !meta.hasMax) return false

  const lessThanMin = meta.hasMin && meta.minCents !== null && targetAmount < meta.minCents
  const greaterThanMax = meta.hasMax && meta.maxCents !== null && targetAmount > meta.maxCents
  if (!lessThanMin && !greaterThanMax) return false

  // 仅在“超出区间隐藏”未开启时，展示但置灰。
  return !meta.hideAmountOutRange
}

const channelAmountLimitHint = (channel?: any) => {
  const meta = channelLimitMeta(channel)
  if (meta.hasMin && meta.hasMax && meta.minCents !== null && meta.maxCents !== null) {
    return t('checkout.channelAmountLimitHint', {
      min: formatPrice(centsToAmount(meta.minCents), previewCurrency.value),
      max: formatPrice(centsToAmount(meta.maxCents), previewCurrency.value),
    })
  }
  if (meta.hasMin && meta.minCents !== null) {
    return t('checkout.channelAmountMinHint', {
      min: formatPrice(centsToAmount(meta.minCents), previewCurrency.value),
    })
  }
  if (meta.hasMax && meta.maxCents !== null) {
    return t('checkout.channelAmountMaxHint', {
      max: formatPrice(centsToAmount(meta.maxCents), previewCurrency.value),
    })
  }
  return ''
}

const handleSelectChannel = (channel?: any) => {
  if (!channel || isChannelDisabledForAmount(channel)) return
  selectedChannelId.value = Number(channel.id) || null
}

const selectedChannelAmountHint = computed(() => {
  const channel = paymentChannels.value.find((item: any) => Number(item?.id) === Number(selectedChannelId.value))
  if (!channel) return ''
  if (!isChannelDisabledForAmount(channel)) return ''
  return channelAmountLimitHint(channel)
})

const formatChannelFeeRate = (channel?: any) => {
  const bp = rateToBasisPoints(channel?.fee_rate)
  if (bp === null) return '0.00%'
  return `${basisPointsToPercent(bp)}%`
}

const formatChannelFixedFee = (channel?: any) => {
  const fixed = channel?.fixed_fee
  if (fixed === null || fixed === undefined || fixed === '' || Number(fixed) === 0) {
    return formatPrice('0.00', previewCurrency.value)
  }
  return formatPrice(String(fixed), previewCurrency.value)
}

const totalAmount = computed(() => {
  const totalCents = cartItems.value.reduce((sum, item) => {
    const amountCents = amountToCents(item.priceAmount)
    const qty = parseInteger(item.quantity)
    if (amountCents === null || qty === null) return sum
    return sum + amountCents * qty
  }, 0)
  return centsToAmount(totalCents)
})

const totalCurrency = computed(() => siteCurrency.value || 'CNY')

const previewCurrency = computed(() => preview.value?.currency || totalCurrency.value)
const previewOriginal = computed(() => preview.value?.original_amount ?? totalAmount.value)
const previewCoupon = computed(() => preview.value?.discount_amount ?? '0')
const previewPromotion = computed(() => preview.value?.promotion_discount_amount ?? '0')
const previewMemberDiscount = computed(() => preview.value?.member_discount_amount ?? '0')
const previewTotal = computed(() => preview.value?.total_amount ?? totalAmount.value)
const checkoutItemCurrency = computed(() => previewCurrency.value)

const previewItemsByKey = computed(() => {
  const map = new Map<string, any>()
  const items = Array.isArray(preview.value?.items) ? preview.value.items : []
  for (const item of items) {
    map.set(`${item.product_id}:${normalizeSkuId(item.sku_id)}`, item)
  }
  return map
})

const checkoutMode = ref<'guest' | 'member'>('guest')
const guestPhone = ref('')
const guestPhoneAutoManaged = ref(true)
const guestEmail = ref('')
const guestPassword = ref('')
const guestCaptchaPayload = ref<CaptchaPayload>({})
const guestTurnstileToken = ref('')
const guestCapToken = ref('')
const guestImageCaptchaRef = ref<InstanceType<typeof ImageCaptcha> | null>(null)
const guestTurnstileRef = ref<InstanceType<typeof TurnstileCaptcha> | null>(null)
const guestCapRef = ref<InstanceType<typeof CapCaptcha> | null>(null)

interface ManualFormField {
  key: string
  type: string
  required: boolean
  label?: Record<string, string>
  placeholder?: Record<string, string>
  regex?: string
  min?: number
  max?: number
  max_len?: number
  options: string[]
}

interface ManualFormProduct {
  itemKey: string
  productId: number
  title: any
  fields: ManualFormField[]
  skuCount: number
}

const manualFieldTypes = new Set(['text', 'textarea', 'phone', 'email', 'number', 'select', 'radio', 'checkbox'])
const emailPattern = /^[^\s@]+@[^\s@]+\.[^\s@]+$/
const phonePattern = /^\+?[0-9\-()\s]{6,20}$/
const findLastUnescapedSlash = (value: string) => {
  for (let index = value.length - 1; index > 0; index -= 1) {
    if (value[index] !== '/') {
      continue
    }
    let backslashes = 0
    for (let cursor = index - 1; cursor >= 0 && value[cursor] === '\\'; cursor -= 1) {
      backslashes += 1
    }
    if (backslashes % 2 === 0) {
      return index
    }
  }
  return -1
}

const compileManualRegex = (rawRegex?: string) => {
  const text = String(rawRegex || '').trim()
  if (!text) {
    return null
  }

  if (text.startsWith('/')) {
    const lastSlashIndex = findLastUnescapedSlash(text)
    if (lastSlashIndex > 0) {
      const pattern = text.slice(1, lastSlashIndex)
      const flags = text.slice(lastSlashIndex + 1)
      if (/^[gimsuy]*$/.test(flags)) {
        try {
          return new RegExp(pattern, flags)
        } catch {
          return null
        }
      }
    }
  }

  try {
    return new RegExp(text)
  } catch {
    return null
  }
}


const manualFormData = ref<Record<string, Record<string, any>>>({})
const submitAttempted = ref(false)

const normalizeManualFormSchema = (rawSchema: any): ManualFormField[] => {
  const rawFields = Array.isArray(rawSchema?.fields) ? rawSchema.fields : []
  return rawFields
    .map((rawField: any) => {
      const key = String(rawField?.key || '').trim()
      const type = String(rawField?.type || '').trim()
      if (!key || !manualFieldTypes.has(type)) {
        return null
      }
      const options = Array.isArray(rawField?.options)
        ? rawField.options.map((item: any) => String(item || '').trim()).filter(Boolean)
        : []
      const minValue = Number(rawField?.min)
      const maxValue = Number(rawField?.max)
      const maxLenValue = Number(rawField?.max_len)
      return {
        key,
        type,
        required: Boolean(rawField?.required),
        label: rawField?.label || undefined,
        placeholder: rawField?.placeholder || undefined,
        regex: String(rawField?.regex || '').trim() || undefined,
        min: Number.isFinite(minValue) ? minValue : undefined,
        max: Number.isFinite(maxValue) ? maxValue : undefined,
        max_len: Number.isFinite(maxLenValue) ? maxLenValue : undefined,
        options: Array.from(new Set(options)),
      } as ManualFormField
    })
    .filter(Boolean) as ManualFormField[]
}

const manualFormProducts = computed<ManualFormProduct[]>(() => {
  const grouped = new Map<number, ManualFormProduct>()
  cartItems.value.forEach((item) => {
    if (item.fulfillmentType !== 'manual' && item.fulfillmentType !== 'upstream') {
      return
    }
    const fields = normalizeManualFormSchema(item.manualFormSchema)
    if (fields.length === 0) {
      return
    }
    const productId = Number(item.productId)
    if (!Number.isFinite(productId) || productId <= 0) {
      return
    }
    const normalizedProductId = Math.trunc(productId)
    const existing = grouped.get(normalizedProductId)
    if (existing) {
      existing.skuCount += 1
      return
    }
    grouped.set(normalizedProductId, {
      itemKey: String(normalizedProductId),
      productId: normalizedProductId,
      title: item.title,
      fields,
      skuCount: 1,
    })
  })
  return Array.from(grouped.values())
})

watch(manualFormProducts, (products) => {
  const nextData: Record<string, Record<string, any>> = {}
  products.forEach((product) => {
    const key = product.itemKey
    const current = manualFormData.value[key] || {}
    const formValues: Record<string, any> = {}
    product.fields.forEach((field) => {
      const currentValue = current[field.key]
      if (field.type === 'checkbox') {
        formValues[field.key] = Array.isArray(currentValue)
          ? currentValue.map((item: any) => String(item)).filter(Boolean)
          : []
      } else {
        formValues[field.key] = currentValue == null ? '' : String(currentValue)
      }
    })
    nextData[key] = formValues
  })
  manualFormData.value = nextData
}, { immediate: true, deep: true })

const resolveLocalizedText = (jsonData?: Record<string, string>, fallback = '') => {
  if (!jsonData) return fallback
  const locale = appStore.locale
  return jsonData[locale] || jsonData['zh-CN'] || jsonData['en-US'] || fallback
}

const getManualFieldLabel = (field: ManualFormField) => {
  return resolveLocalizedText(field.label, field.key)
}

const getManualFieldPlaceholder = (field: ManualFormField) => {
  return resolveLocalizedText(field.placeholder, '')
}

const manualFieldErrorKey = (itemKey: string, fieldKey: string) => `${itemKey}:${fieldKey}`



const manualFormValidation = computed(() => {
  const errors: Record<string, string> = {}
  let firstError = ''

  const setError = (itemKey: string, field: ManualFormField, message: string) => {
    const errorKey = manualFieldErrorKey(itemKey, field.key)
    if (!errors[errorKey]) {
      errors[errorKey] = message
      if (!firstError) {
        firstError = message
      }
    }
  }

  manualFormProducts.value.forEach((product) => {
    const values = manualFormData.value[product.itemKey] || {}
    product.fields.forEach((field) => {
      const fieldLabel = getManualFieldLabel(field)
      const rawValue = values[field.key]
      if (field.type === 'checkbox') {
        const list = Array.isArray(rawValue)
          ? rawValue.map((item: any) => String(item).trim()).filter(Boolean)
          : []
        if (field.required && list.length === 0) {
          setError(product.itemKey, field, t('checkout.manualFormFieldRequired', { name: fieldLabel }))
          return
        }
        if (list.length > 0 && field.options.length > 0 && list.some((item) => !field.options.includes(item))) {
          setError(product.itemKey, field, t('checkout.manualFormFieldOptionInvalid', { name: fieldLabel }))
        }
        return
      }

      const text = rawValue == null ? '' : String(rawValue).trim()
      if (field.required && !text) {
        setError(product.itemKey, field, t('checkout.manualFormFieldRequired', { name: fieldLabel }))
        return
      }
      if (!text) {
        return
      }

      if ((field.type === 'text' || field.type === 'textarea' || field.type === 'phone' || field.type === 'email') && field.max_len && text.length > field.max_len) {
        setError(product.itemKey, field, t('checkout.manualFormFieldMaxLength', { name: fieldLabel, max: field.max_len }))
        return
      }
      if ((field.type === 'phone' && !phonePattern.test(text)) || (field.type === 'email' && !emailPattern.test(text))) {
        setError(product.itemKey, field, t('checkout.manualFormFieldInvalid', { name: fieldLabel }))
        return
      }
      if (field.type === 'number') {
        const numberValue = Number(text)
        if (!Number.isFinite(numberValue)) {
          setError(product.itemKey, field, t('checkout.manualFormFieldNumberInvalid', { name: fieldLabel }))
          return
        }
        if ((field.min !== undefined && numberValue < field.min) || (field.max !== undefined && numberValue > field.max)) {
          setError(product.itemKey, field, t('checkout.manualFormFieldNumberRange', { name: fieldLabel }))
          return
        }
      }
      if ((field.type === 'select' || field.type === 'radio') && field.options.length > 0 && !field.options.includes(text)) {
        setError(product.itemKey, field, t('checkout.manualFormFieldOptionInvalid', { name: fieldLabel }))
        return
      }
      if (field.regex) {
        const regex = compileManualRegex(field.regex)
        if (!regex || !regex.test(text)) {
          setError(product.itemKey, field, t('checkout.manualFormFieldInvalid', { name: fieldLabel }))
        }
      }
    })
  })

  return {
    valid: Object.keys(errors).length === 0,
    errors,
    firstError,
  }
})

const manualFieldError = (itemKey: string, fieldKey: string) => {
  return manualFormValidation.value.errors[manualFieldErrorKey(itemKey, fieldKey)] || ''
}

const buildManualFormDataPayload = () => {
  const payload: Record<string, any> = {}
  manualFormProducts.value.forEach((product) => {
    const values = manualFormData.value[product.itemKey] || {}
    const row: Record<string, any> = {}
    product.fields.forEach((field) => {
      const rawValue = values[field.key]
      if (field.type === 'checkbox') {
        const list = Array.isArray(rawValue)
          ? rawValue.map((item: any) => String(item).trim()).filter(Boolean)
          : []
        if (list.length > 0) {
          row[field.key] = list
        }
        return
      }
      const text = rawValue == null ? '' : String(rawValue).trim()
      if (text) {
        row[field.key] = text
      }
    })
    payload[product.itemKey] = row
  })
  return payload
}

const manualFormFingerprint = computed(() => JSON.stringify(manualFormData.value))
const emptyShippingAddress = (): ShippingAddressFormValue => ({
  receiver_name: '',
  receiver_phone: '',
  province: '',
  province_code: '',
  city: '',
  city_code: '',
  district: '',
  district_code: '',
  township: '',
  township_code: '',
  detail_address: '',
})
const shippingAddress = ref<ShippingAddressFormValue>(emptyShippingAddress())
const orderRequiresShippingAddress = computed(() => cartItems.value.some((item) => item.requiresShippingAddress))
const shippingRegionMissing = computed(() => {
  if (!orderRequiresShippingAddress.value) return false
  const requiredKeys = ['province_code', 'city_code', 'district_code', 'township_code'] as const
  return requiredKeys.some((key) => !String(shippingAddress.value[key] || '').trim())
})
const shippingAddressValidation = computed(() => {
  if (!orderRequiresShippingAddress.value) {
    return { valid: true, message: '' }
  }
  const requiredKeys = ['receiver_name', 'receiver_phone', 'province_code', 'city_code', 'district_code', 'township_code', 'detail_address'] as const
  const missingKey = requiredKeys.find((key) => !String(shippingAddress.value[key] || '').trim())
  if (missingKey) {
    return { valid: false, message: t('checkout.errors.shippingAddressRequired') }
  }
  return { valid: true, message: '' }
})
const buildShippingAddressPayload = () => {
  if (!orderRequiresShippingAddress.value) return undefined
  return {
    receiver_name: shippingAddress.value.receiver_name.trim(),
    receiver_phone: shippingAddress.value.receiver_phone.trim(),
    province: shippingAddress.value.province.trim(),
    province_code: shippingAddress.value.province_code.trim(),
    city: shippingAddress.value.city.trim(),
    city_code: shippingAddress.value.city_code.trim(),
    district: shippingAddress.value.district.trim(),
    district_code: shippingAddress.value.district_code.trim(),
    township: shippingAddress.value.township.trim(),
    township_code: shippingAddress.value.township_code.trim(),
    detail_address: shippingAddress.value.detail_address.trim(),
  }
}
const shippingAddressFingerprint = computed(() => JSON.stringify(buildShippingAddressPayload() || null))

const guestShippingRecallRecord = ref<GuestShippingAddressRecallRecord | null>(null)
const guestShippingRecallApplied = ref(false)
const guestShippingRecallRewriteMode = ref(false)

const isGuestCheckout = computed(() => !userAuthStore.isAuthenticated && checkoutMode.value === 'guest')
const guestShippingRecallEnabled = computed(() => shouldEnableGuestShippingAddressRecall({
  orderRequiresShippingAddress: orderRequiresShippingAddress.value,
  isAuthenticated: userAuthStore.isAuthenticated,
  checkoutMode: checkoutMode.value,
}))
const hasManualShippingInput = computed(() => Boolean(
  shippingAddress.value.receiver_name.trim() ||
  shippingAddress.value.receiver_phone.trim() ||
  shippingAddress.value.province_code.trim() ||
  shippingAddress.value.city_code.trim() ||
  shippingAddress.value.district_code.trim() ||
  shippingAddress.value.township_code.trim() ||
  shippingAddress.value.detail_address.trim(),
))
const showGuestShippingRecallCard = computed(() =>
  guestShippingRecallEnabled.value && !!guestShippingRecallRecord.value,
)
const guestShippingRecallMuted = computed(() =>
  !guestShippingRecallApplied.value && (hasManualShippingInput.value || guestShippingRecallRewriteMode.value),
)
const guestShippingRecallSummaryLines = computed(() => {
  const record = guestShippingRecallRecord.value
  if (!record) return []

  const trimmedPhone = record.receiver_phone.trim()
  const maskedPhone = trimmedPhone.length >= 7
    ? `${trimmedPhone.slice(0, 3)}****${trimmedPhone.slice(-4)}`
    : trimmedPhone

  return [
    `${record.receiver_name} · ${maskedPhone}`,
    [record.province, record.city, record.district, record.township, record.detail_address].filter(Boolean).join(' '),
  ]
})
const shouldSyncGuestPhoneFromShipping = computed(() => {
  return isGuestCheckout.value && orderRequiresShippingAddress.value && guestPhoneAutoManaged.value
})
const guestPhoneValid = computed(() => {
  if (!isGuestCheckout.value) return true
  return phonePattern.test(guestPhone.value.trim())
})

const handleGuestPhoneInput = (event: Event) => {
  const nextValue = String((event.target as HTMLInputElement | null)?.value || '')
  const shippingPhone = shippingAddress.value.receiver_phone

  guestPhone.value = nextValue
  guestPhoneAutoManaged.value = nextValue === '' || nextValue === shippingPhone
}

const resetShippingAddressForm = () => {
  shippingAddress.value = emptyShippingAddress()
}

const applyGuestShippingRecall = () => {
  const record = guestShippingRecallRecord.value
  if (!record) return

  shippingAddress.value = {
    receiver_name: record.receiver_name,
    receiver_phone: record.receiver_phone,
    province: record.province,
    province_code: record.province_code,
    city: record.city,
    city_code: record.city_code,
    district: record.district,
    district_code: record.district_code,
    township: record.township,
    township_code: record.township_code,
    detail_address: record.detail_address,
  }

  guestShippingRecallApplied.value = true
  guestShippingRecallRewriteMode.value = false
}

const handleGuestShippingRewrite = () => {
  resetShippingAddressForm()
  guestShippingRecallApplied.value = false
  guestShippingRecallRewriteMode.value = true
}

const handleGuestShippingClearForm = () => {
  resetShippingAddressForm()
  guestShippingRecallApplied.value = false
  guestShippingRecallRewriteMode.value = true
}

const handleGuestShippingClearRecord = () => {
  clearGuestShippingAddressRecall()
  guestShippingRecallRecord.value = null
  guestShippingRecallApplied.value = false
  guestShippingRecallRewriteMode.value = false
}

const persistGuestShippingRecallFromCurrentAddress = () => {
  const shippingPayload = buildShippingAddressPayload()
  if (!isGuestCheckout.value || !shippingPayload) return null

  const recentRecord = createGuestShippingAddressRecallRecord(shippingPayload)
  guestShippingRecallRecord.value = recentRecord
  guestShippingRecallRewriteMode.value = false
  return recentRecord
}

const guestEmailValid = computed(() => {
  if (!isGuestCheckout.value) return true
  const value = guestEmail.value.trim()
  if (!value) return true
  return emailPattern.test(value)
})

const captchaConfig = computed(() => appStore.config?.captcha || null)
const captchaProvider = computed(() => String(captchaConfig.value?.provider || 'none'))
const guestCaptchaEnabled = computed(() => {
  if (!isGuestCheckout.value) return false
  return !!captchaConfig.value?.scenes?.guest_create_order && captchaProvider.value !== 'none'
})
const guestTurnstileSiteKey = computed(() => String(captchaConfig.value?.turnstile?.site_key || ''))
const guestCapEndpoint = computed(() => String(captchaConfig.value?.cap?.endpoint || ''))
const guestCapSiteKey = computed(() => String(captchaConfig.value?.cap?.site_key || ''))
const guestCaptchaComplete = computed(() => {
  if (!guestCaptchaEnabled.value) return true
  if (captchaProvider.value === 'image') {
    return Boolean(guestCaptchaPayload.value.captcha_id && guestCaptchaPayload.value.captcha_code)
  }
  if (captchaProvider.value === 'turnstile') {
    return Boolean(guestTurnstileToken.value)
  }
  if (captchaProvider.value === 'cap') {
    return Boolean(guestCapToken.value)
  }
  return false
})

const getGuestCaptchaPayload = (): CaptchaPayload | undefined => {
  if (!guestCaptchaEnabled.value) return undefined
  if (captchaProvider.value === 'image') {
    return {
      captcha_id: guestCaptchaPayload.value.captcha_id || '',
      captcha_code: guestCaptchaPayload.value.captcha_code || '',
    }
  }
  if (captchaProvider.value === 'turnstile') {
    return {
      turnstile_token: guestTurnstileToken.value,
    }
  }
  if (captchaProvider.value === 'cap') {
    return {
      cap_token: guestCapToken.value,
    }
  }
  return undefined
}

const handleGuestCaptchaConfigStale = async () => {
  await appStore.loadConfig(true)
  guestCaptchaPayload.value = {}
  guestTurnstileToken.value = ''
  guestCapToken.value = ''
}

const canSubmit = computed(() => {
  if (syncingStock.value) return false
  if (submitting.value) return false
  if (cartItems.value.length === 0) return false
  if (!manualFormValidation.value.valid) return false
  if (!shippingAddressValidation.value.valid) return false
  if (cartItems.value.some((item) => itemStockExceeded(item))) return false
  if (cartItems.value.some((item) => itemMinNotMet(item))) return false
  if (walletOnlyPayment.value && expectedOnlinePayCents.value > 0) return false
  if (!walletOnlyPayment.value && requiresOnlineChannel.value && !selectedChannelId.value) return false
  if (requiresOnlineChannel.value && selectedChannelAmountHint.value) return false
  if (userAuthStore.isAuthenticated) return true
  if (checkoutMode.value !== 'guest') return false
  if (!guestPhone.value.trim() || !guestPassword.value.trim() || !guestPhoneValid.value || !guestEmailValid.value) return false
  return guestCaptchaComplete.value
})

const submitBlockedReason = computed(() => {
  if (syncingStock.value) return t('checkout.stockSyncing')
  if (cartItems.value.length === 0) return t('checkout.errors.emptyCart')
  if (!manualFormValidation.value.valid) {
    return manualFormValidation.value.firstError || t('checkout.errors.manualFormInvalid')
  }
  if (!shippingAddressValidation.value.valid) {
    return shippingAddressValidation.value.message
  }
  const stockBlockedItem = cartItems.value.find((item) => itemStockExceeded(item))
  if (stockBlockedItem) {
    return itemStockHint(stockBlockedItem) || t('cart.stockOut')
  }
  const minBlockedItem = cartItems.value.find((item) => itemMinNotMet(item))
  if (minBlockedItem) {
    return t('cart.minPurchaseNotMet', { count: itemPurchaseMin(minBlockedItem) })
  }
  if (walletOnlyPayment.value && expectedOnlinePayCents.value > 0) return t('payment.walletInsufficientHint')
  if (!walletOnlyPayment.value && requiresOnlineChannel.value && !selectedChannelId.value) return t('checkout.errors.selectPayment')
  if (requiresOnlineChannel.value && selectedChannelAmountHint.value) return selectedChannelAmountHint.value
  if (userAuthStore.isAuthenticated) return ''
  if (checkoutMode.value !== 'guest') return t('checkout.errors.loginOrGuest')
  if (!guestPhone.value.trim() || !guestPassword.value.trim()) return t('checkout.errors.missingGuest')
  if (!guestPhoneValid.value) return t('error.phone_invalid')
  if (!guestEmailValid.value) return t('error.email_invalid')
  if (!guestCaptchaComplete.value) {
    return t('auth.common.captchaRequired')
  }
  return ''
})

const previewStatusText = computed(() => couponRefreshing.value
  ? t('checkout.couponRefreshing')
  : t('checkout.previewLoading'))

const checkoutAlert = computed<PageAlert | null>(() => {
  if (error.value) {
    return { level: 'error' as const, message: error.value }
  }
  if (previewError.value) {
    return { level: 'error' as const, message: previewError.value }
  }
  if (!canSubmit.value && submitBlockedReason.value) {
    return { level: 'warning' as const, message: submitBlockedReason.value }
  }
  return null
})

const mobileExpandedSection = ref<MobileCheckoutSectionKey | null>(null)
const MOBILE_CHECKOUT_SECTION_TRANSITION_MS = 220
type MobileConfirmableSectionKey = 'shipping' | 'buyer' | 'payment'
const mobileConfirmedFingerprints = ref<Partial<Record<MobileConfirmableSectionKey, string>>>({})

const maskPhone = (value: string) => {
  const trimmed = value.trim()
  if (trimmed.length < 7) return trimmed
  return `${trimmed.slice(0, 3)}****${trimmed.slice(-4)}`
}

const mobileManualFormsReady = computed(() => isMobileManualFormReady(
  manualFormProducts.value,
  manualFormData.value,
))

const mobileShippingReady = computed(() => isMobileShippingReady({
  requiresShipping: orderRequiresShippingAddress.value,
  receiverName: shippingAddress.value.receiver_name,
  receiverPhone: shippingAddress.value.receiver_phone,
  provinceCode: shippingAddress.value.province_code,
  cityCode: shippingAddress.value.city_code,
  districtCode: shippingAddress.value.district_code,
  townshipCode: shippingAddress.value.township_code,
  detailAddress: shippingAddress.value.detail_address,
}))

const mobileBuyerReady = computed(() => {
  return isMobileBuyerReady({
    isAuthenticated: userAuthStore.isAuthenticated,
    checkoutMode: checkoutMode.value,
    manualFormsReady: mobileManualFormsReady.value,
    guestPhone: guestPhone.value,
    guestPassword: guestPassword.value,
    guestEmail: guestEmail.value,
    captchaComplete: guestCaptchaComplete.value,
  })
})

const mobilePaymentReady = computed(() => {
  if (walletOnlyPayment.value) return expectedOnlinePayCents.value <= 0
  if (!requiresOnlineChannel.value) return true
  return Boolean(selectedChannelId.value) && !selectedChannelAmountHint.value
})

const mobileShippingFingerprint = computed(() => JSON.stringify({
  requiresShipping: orderRequiresShippingAddress.value,
  address: buildShippingAddressPayload() ?? null,
}))

const mobileBuyerFingerprint = computed(() => JSON.stringify({
  isAuthenticated: userAuthStore.isAuthenticated,
  checkoutMode: checkoutMode.value,
  guestPhone: guestPhone.value.trim(),
  guestEmail: guestEmail.value.trim(),
  guestPassword: guestPassword.value,
  guestCaptchaComplete: guestCaptchaComplete.value,
  manualFormData: buildManualFormDataPayload(),
}))

const mobilePaymentFingerprint = computed(() => JSON.stringify({
  useBalance: useBalance.value,
  selectedChannelId: requiresOnlineChannel.value ? selectedChannelId.value : null,
  requiresOnlineChannel: requiresOnlineChannel.value,
  expectedOnlinePayCents: expectedOnlinePayCents.value,
  walletOnlyPayment: walletOnlyPayment.value,
}))

const mobileShippingDirty = computed(() => {
  if (!orderRequiresShippingAddress.value) return false
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
  if (!orderRequiresShippingAddress.value) return true
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
  hasShippingSection: orderRequiresShippingAddress.value,
  shippingComplete: mobileShippingComplete.value,
  buyerComplete: mobileBuyerComplete.value,
  paymentComplete: mobilePaymentComplete.value,
}))

const mobileShippingErrorMessage = computed(() => {
  if (!submitAttempted.value) return ''
  if (mobileFlowState.value.recommendedSectionKey !== 'shipping') return ''
  if (mobileShippingReady.value) return ''
  return shippingAddressValidation.value.message || t('checkout.mobile.shippingMissing')
})

const mobileBuyerErrorMessage = computed(() => {
  if (!submitAttempted.value) return ''
  if (mobileFlowState.value.recommendedSectionKey !== 'buyer') return ''
  if (mobileBuyerReady.value) return ''

  return resolveMobileBuyerErrorMessage({
    manualFormsValid: manualFormValidation.value.valid,
    manualFormFirstError: manualFormValidation.value.firstError,
    isAuthenticated: userAuthStore.isAuthenticated,
    checkoutMode: checkoutMode.value,
    guestPhone: guestPhone.value,
    guestPassword: guestPassword.value,
    guestPhoneValid: guestPhoneValid.value,
    guestEmailValid: guestEmailValid.value,
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
  if (!submitAttempted.value) return ''
  if (mobileFlowState.value.recommendedSectionKey !== 'payment') return ''
  if (mobilePaymentReady.value) return ''

  return resolveMobilePaymentErrorMessage({
    walletOnlyPayment: walletOnlyPayment.value,
    expectedOnlinePayCents: expectedOnlinePayCents.value,
    requiresOnlineChannel: requiresOnlineChannel.value,
    selectedChannelId: selectedChannelId.value,
    selectedChannelAmountHint: selectedChannelAmountHint.value,
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
  if (error.value) return error.value
  if (previewError.value) return previewError.value
  if (previewLoading.value || couponRefreshing.value) return previewStatusText.value
  if (mobileCurrentSectionErrorMessage.value) return mobileCurrentSectionErrorMessage.value

  const action = mobileFlowState.value.primaryActionKey
  if (action === 'saveShipping') {
    if (mobileShippingReady.value) return t('checkout.mobile.actionSaveShipping')
    return t('checkout.mobile.shippingMissing')
  }
  if (action === 'continueBuyer') {
    if (mobileBuyerReady.value) return t('checkout.mobile.actionContinueBuyer')
    return t('checkout.mobile.buyerMissing')
  }
  if (action === 'choosePayment') {
    if (mobilePaymentReady.value) return t('checkout.mobile.readyToSubmit')
    return t('checkout.mobile.paymentMissing')
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

const mobileTotalText = computed(() => formatPrice(previewTotal.value, previewCurrency.value))

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
    shippingAddress.value.province,
    shippingAddress.value.city,
    shippingAddress.value.district,
    shippingAddress.value.township,
    shippingAddress.value.detail_address,
  ]
    .map((value) => String(value || '').trim())
    .filter(Boolean)
    .join(' ')

  const shippingSummaryLines = [
    shippingAddress.value.receiver_name && shippingAddress.value.receiver_phone
      ? `${shippingAddress.value.receiver_name} · ${maskPhone(shippingAddress.value.receiver_phone)}`
      : '',
    shippingRegionLine,
  ].filter(Boolean)

  const buyerSummaryLines = [
    userAuthStore.isAuthenticated
      ? t('checkout.mobile.buyerLoggedIn')
      : checkoutMode.value === 'guest' && guestPhone.value.trim()
        ? t('checkout.mobile.buyerGuest', { phone: maskPhone(guestPhone.value) })
        : t('checkout.mobile.buyerMissing'),
    manualFormProducts.value.length > 0 && !mobileManualFormsReady.value
      ? t('checkout.mobile.buyerManualPending')
      : '',
  ].filter(Boolean)

  const couponDiscountCents = amountToCents(previewCoupon.value)
  const couponSummaryLines = normalizedCouponCode.value
    ? [
        normalizedCouponCode.value,
        couponDiscountCents !== null && couponDiscountCents > 0
          ? t('checkout.mobile.summaryCouponApplied', {
            amount: formatPrice(previewCoupon.value, previewCurrency.value),
          })
          : '',
      ].filter(Boolean)
    : [t('checkout.mobile.summaryCouponEmpty')]

  const selectedChannel = paymentChannels.value.find((channel: any) => Number(channel?.id) === Number(selectedChannelId.value))
  const paymentSummaryLines = [
    useBalance.value && expectedWalletPaidCents.value > 0
      ? `${t('payment.walletDeductLabel')} ${expectedWalletPaidDisplay.value}`
      : '',
    !requiresOnlineChannel.value
      ? t('checkout.walletCoversAll')
      : selectedChannel?.name || t('checkout.mobile.paymentMissing'),
  ].filter(Boolean)

  const summaryMap: Record<MobileCheckoutSectionKey, string[]> = {
    items: [
      t('checkout.mobile.summaryItems', { count: totalItems.value }),
      `${t('checkout.previewTotal')} ${formatPrice(previewTotal.value, previewCurrency.value)}`,
    ],
    shipping: shippingSummaryLines.length > 0 ? shippingSummaryLines : [t('checkout.mobile.shippingMissing')],
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
    if (!document.querySelector('.mobile-checkout-section-leave-active')) {
      break
    }

    await waitForAnimationFrame()
  }

  const section = document.querySelector(`[data-section-toggle="${sectionKey}"]`)
  if (section instanceof HTMLElement) {
    const siteHeader = document.querySelector('[data-site-header]')
    const fixedOffset = siteHeader instanceof HTMLElement ? siteHeader.getBoundingClientRect().height : 0
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
  const fixedOffset = siteHeader instanceof HTMLElement ? siteHeader.getBoundingClientRect().height : 0
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
  if (!shippingAddress.value.receiver_name.trim()) return '[data-mobile-shipping-input="receiver-name"]'
  if (!shippingAddress.value.receiver_phone.trim()) return '[data-mobile-shipping-input="receiver-phone"]'
  if (shippingRegionMissing.value) return '[data-mobile-shipping-input="region"]'
  if (!shippingAddress.value.detail_address.trim()) return '[data-mobile-shipping-input="detail-address"]'
  return '[data-section-toggle="shipping"]'
}

const getMobileBuyerErrorSelector = () => {
  const firstManualFieldErrorKey = Object.keys(manualFormValidation.value.errors)[0]
  if (firstManualFieldErrorKey) return `[data-manual-field-input="${firstManualFieldErrorKey}"]`
  if (!userAuthStore.isAuthenticated && checkoutMode.value !== 'guest') return '[data-mobile-buyer-input="checkout-mode"]'
  if (!guestPhone.value.trim() || !guestPhoneValid.value) return '[data-mobile-buyer-input="guest-phone"]'
  if (!guestPassword.value.trim()) return '[data-mobile-buyer-input="guest-password"]'
  if (!guestEmailValid.value) return '[data-mobile-buyer-input="guest-email"]'
  if (!guestCaptchaComplete.value) return '[data-mobile-buyer-input="guest-captcha"]'
  return '[data-section-toggle="buyer"]'
}

const getMobilePaymentErrorSelector = () => {
  if (requiresOnlineChannel.value) return '[data-mobile-payment-input="channel-list"], [data-section-toggle="payment"]'
  return '[data-section-toggle="payment"]'
}

const focusMobileErrorTarget = async (sectionKey: MobileCheckoutSectionKey) => {
  const selectorMap: Partial<Record<MobileCheckoutSectionKey, string>> = {
    shipping: getMobileShippingErrorSelector(),
    buyer: getMobileBuyerErrorSelector(),
    payment: getMobilePaymentErrorSelector(),
  }

  const focusSelector = selectorMap[sectionKey] || ''
  const selectors = resolveMobileErrorTargetSelectors({
    sectionKey,
    focusSelector,
  })

  await scrollMobileElementIntoView(selectors.scrollSelector, selectors.focusSelector)
}

const confirmMobileSection = (sectionKey: MobileConfirmableSectionKey, fingerprint: string) => {
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
  submitAttempted.value = true

  const action = mobileFlowState.value.primaryActionKey
  if (action === 'saveShipping') {
    mobileExpandedSection.value = 'shipping'
    await scrollMobileSectionIntoView('shipping')
    if (!mobileShippingReady.value) {
      await focusMobileErrorTarget('shipping')
      return
    }

    confirmMobileSection('shipping', mobileShippingFingerprint.value)
    persistGuestShippingRecallFromCurrentAddress()
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
    await handleSubmit()
    return
  }

  await handleSubmit()
}

const buildItemsPayload = () => cartItems.value.map(item => ({
  product_id: item.productId,
  sku_id: normalizeSkuId(item.skuId) || undefined,
  quantity: item.quantity,
  fulfillment_type: item.fulfillmentType || undefined,
}))

const buildOrderPayload = () => ({
  coupon_code: normalizedCouponCode.value || undefined,
  affiliate_code: getAffiliateCode() || undefined,
  affiliate_visitor_key: getAffiliateVisitorKey() || undefined,
  items: buildItemsPayload(),
  manual_form_data: buildManualFormDataPayload(),
  shipping_address: buildShippingAddressPayload(),
})

const loadOrderPaymentChannels = async () => {
  if (!userAuthStore.isAuthenticated) {
    orderPaymentChannels.value = []
    return
  }
  if (!requiresOnlineChannel.value) {
    orderPaymentChannels.value = []
    return
  }
  if (cartItems.value.length === 0 || !preview.value) {
    orderPaymentChannels.value = []
    return
  }
  const requestId = ++orderPaymentChannelsRequestId.value
  try {
    const response = await userOrderAPI.getPaymentChannels({
      amount: centsToAmount(expectedOnlinePayCents.value),
      items: buildItemsPayload(),
    })
    if (requestId !== orderPaymentChannelsRequestId.value) return
    const channels = response.data.data
    orderPaymentChannels.value = Array.isArray(channels) ? channels : []
  } catch {
    if (requestId !== orderPaymentChannelsRequestId.value) return
    const fallback = preview.value?.payment_channels
    orderPaymentChannels.value = Array.isArray(fallback) ? fallback : []
  }
}

const debouncedLoadOrderPaymentChannels = debounceAsync(loadOrderPaymentChannels, 250)

const syncCartStockSnapshots = async () => {
  if (isBuyNowMode.value) return
  if (syncingStock.value) return
  syncingStock.value = true
  try {
    await refreshCartStockSnapshots(cartStore)
  } finally {
    syncingStock.value = false
  }
}

const loadPreview = async () => {
  if (syncingStock.value) {
    preview.value = null
    orderPaymentChannels.value = []
    previewError.value = ''
    couponRefreshing.value = false
    return
  }
  if (cartItems.value.length === 0) {
    preview.value = null
    orderPaymentChannels.value = []
    previewError.value = ''
    couponRefreshing.value = false
    return
  }
  if (isGuestCheckout.value && (!guestPhone.value.trim() || !guestPassword.value.trim() || !guestPhoneValid.value || !guestEmailValid.value)) {
    preview.value = null
    orderPaymentChannels.value = []
    previewError.value = ''
    couponRefreshing.value = false
    return
  }
  if (!manualFormValidation.value.valid) {
    preview.value = null
    orderPaymentChannels.value = []
    previewError.value = ''
    couponRefreshing.value = false
    return
  }
  if (!shippingAddressValidation.value.valid) {
    preview.value = null
    orderPaymentChannels.value = []
    previewError.value = ''
    couponRefreshing.value = false
    return
  }
  if (cartItems.value.some((item) => itemStockExceeded(item))) {
    preview.value = null
    orderPaymentChannels.value = []
    previewError.value = ''
    couponRefreshing.value = false
    return
  }
  if (cartItems.value.some((item) => itemMinNotMet(item))) {
    preview.value = null
    orderPaymentChannels.value = []
    previewError.value = ''
    couponRefreshing.value = false
    return
  }

  const requestId = ++previewRequestId.value
  previewLoading.value = true
  previewError.value = ''

  try {
    const payload: any = buildOrderPayload()

    let response
    if (userAuthStore.isAuthenticated) {
      response = await userOrderAPI.preview(payload)
    } else {
      response = await guestOrderAPI.preview({
        ...payload,
        phone: guestPhone.value.trim(),
        email: guestEmail.value.trim(),
        order_password: guestPassword.value,
      })
    }

    if (requestId !== previewRequestId.value) return
    preview.value = response.data.data
    if (userAuthStore.isAuthenticated) {
      debouncedLoadOrderPaymentChannels()
    } else {
      orderPaymentChannels.value = []
    }
  } catch (err: any) {
    if (requestId !== previewRequestId.value) return
    preview.value = null
    orderPaymentChannels.value = []
    previewError.value = err.message || t('checkout.previewFailed')
  } finally {
    if (requestId === previewRequestId.value) {
      previewLoading.value = false
      couponRefreshing.value = false
    }
  }
}

const debouncedLoadPreview = debounceAsync(loadPreview, 300)

const loadPreviewNow = async () => {
  debouncedLoadPreview.cancel()
  debouncedLoadOrderPaymentChannels.cancel()
  await loadPreview()
}

const clearSourceStore = () => {
  if (isBuyNowMode.value) {
    buyNowStore.clear()
  } else {
    cartStore.clear()
  }
}

const handleSubmit = async () => {
  submitAttempted.value = true
  error.value = ''
  previewError.value = ''
  if (!canSubmit.value) {
    error.value = submitBlockedReason.value || t('checkout.errors.submitFailed')
    return
  }

  submitting.value = true
  try {
    await loadPreviewNow()
    if (previewError.value) {
      error.value = previewError.value
      return
    }

    const payload = {
      ...buildOrderPayload(),
      channel_id: requiresOnlineChannel.value ? (selectedChannelId.value || undefined) : undefined,
      use_balance: useBalance.value,
    }

    let responseData: any

    if (userAuthStore.isAuthenticated) {
      const response = await userOrderAPI.createAndPay(payload)
      responseData = response.data.data
    } else {
      const response = await guestOrderAPI.createAndPay({
        ...payload,
        phone: guestPhone.value.trim(),
        email: guestEmail.value.trim(),
        order_password: guestPassword.value,
        captcha_payload: getGuestCaptchaPayload(),
      })
      persistGuestShippingRecallFromCurrentAddress()
      localStorage.setItem('guest_order_auth', JSON.stringify({
        phone: guestPhone.value.trim(),
        email: guestEmail.value.trim(),
        order_password: guestPassword.value,
      }))
      responseData = response.data.data
    }

    if (!responseData?.order_no) {
      throw new Error(t('checkout.errors.submitFailed'))
    }

    clearSourceStore()

    // Redirect to the existing Payment page which handles all payment display
    const query = userAuthStore.isAuthenticated
      ? `order_no=${encodeURIComponent(responseData.order_no)}`
      : `guest=1&order_no=${encodeURIComponent(responseData.order_no)}`
    router.push(`/pay?${query}`)
  } catch (err: any) {
    error.value = err.message || t('checkout.errors.submitFailed')
    if (guestCaptchaEnabled.value && captchaProvider.value === 'image') {
      guestImageCaptchaRef.value?.refresh()
    }
    if (guestCaptchaEnabled.value && captchaProvider.value === 'turnstile') {
      guestTurnstileRef.value?.reset()
      guestTurnstileToken.value = ''
    }
    if (guestCaptchaEnabled.value && captchaProvider.value === 'cap') {
      guestCapRef.value?.reset()
      guestCapToken.value = ''
    }
  } finally {
    submitting.value = false
  }
}

watch(
  () => [cartItems.value, manualFormFingerprint.value, shippingAddressFingerprint.value, normalizedCouponCode.value, checkoutMode.value, guestPhone.value, guestEmail.value, guestPassword.value, userAuthStore.isAuthenticated],
  () => {
    debouncedLoadPreview()
  },
  { deep: true }
)

watch(
  () => [shippingAddress.value.receiver_phone, shouldSyncGuestPhoneFromShipping.value],
  ([receiverPhone, shouldSync]) => {
    if (!shouldSync) return
    guestPhone.value = String(receiverPhone || '')
  },
  { immediate: true }
)

watch(
  () => isGuestCheckout.value,
  (isGuest) => {
    if (!isGuest) return
    if (!guestPhone.value.trim()) {
      guestPhoneAutoManaged.value = true
    }
  },
  { immediate: true }
)

watch(walletOnlyPayment, (v) => {
  if (v) useBalance.value = true
}, { immediate: true })

watch(normalizedCouponCode, (value, previous) => {
  if (value === previous) return
  couponRefreshing.value = true
  error.value = ''
  previewError.value = ''
})

watch(
  () => [userAuthStore.isAuthenticated, requiresOnlineChannel.value, expectedOnlinePayCents.value, preview.value?.total_amount],
  () => {
    debouncedLoadOrderPaymentChannels()
  }
)

watch(
  () => [paymentChannels.value, expectedOnlinePayCents.value, requiresOnlineChannel.value],
  () => {
    if (!selectedChannelId.value) return
    const selected = paymentChannels.value.find((item: any) => Number(item?.id) === Number(selectedChannelId.value))
    if (!selected || isChannelDisabledForAmount(selected)) {
      selectedChannelId.value = null
    }
  },
  { deep: true }
)

const loadWalletBalance = async () => {
  if (!userAuthStore.isAuthenticated) return
  walletLoading.value = true
  try {
    const response = await walletAPI.account()
    walletBalance.value = String(response.data.data?.balance || '0')
  } catch {
    walletBalance.value = '0'
  } finally {
    walletLoading.value = false
  }
}

onMounted(async () => {
  guestShippingRecallRecord.value = loadGuestShippingAddressRecall()
  if (!appStore.config) {
    await appStore.loadConfig()
  }
  await syncCartStockSnapshots()
  debouncedLoadPreview()
  loadWalletBalance()
})

onUnmounted(() => {
  debouncedLoadPreview.cancel()
  debouncedLoadOrderPaymentChannels.cancel()
})

const cartItemKey = (item: CartItem) => `${item.productId}:${normalizeSkuId(item.skuId)}`

const itemSkuDisplay = (item: CartItem) => buildSkuDisplayText({
  skuCode: item.skuCode,
  specValues: item.skuSpecValues,
  fallback: t('productDetail.skuFallback'),
  locale: appStore.locale,
})

const checkoutItemImage = (item: CartItem) => {
  const rawImage = String(item.image || '').trim()
  if (!rawImage) return ''
  return getImageUrl(rawImage)
}

const cartItemSubtotalCents = (item: CartItem) => {
  const amountCents = amountToCents(item.priceAmount)
  const qty = parseInteger(item.quantity)
  if (amountCents === null || qty === null) {
    return null
  }
  return amountCents * qty
}

const checkoutItemPreview = (item: CartItem) => previewItemsByKey.value.get(cartItemKey(item))

const checkoutItemOriginalCents = (item: CartItem) => {
  const previewItem = checkoutItemPreview(item)
  const previewOriginalCents = amountToCents(previewItem?.original_total_price)
  if (previewOriginalCents !== null) return previewOriginalCents
  return cartItemSubtotalCents(item)
}

const checkoutItemPayableCents = (item: CartItem) => {
  const previewItem = checkoutItemPreview(item)
  const previewPayableCents = amountToCents(previewItem?.total_price)
  if (previewPayableCents !== null) {
    const couponDiscountCents = amountToCents(previewItem?.coupon_discount_amount) || 0
    return Math.max(0, previewPayableCents - couponDiscountCents)
  }
  return checkoutItemOriginalCents(item)
}

const pricePartsFromCents = (cents: number | null) => {
  if (cents === null) return { integer: '-', decimal: '' }
  const amount = centsToAmount(Math.max(0, cents))
  const [integer, decimal = ''] = amount.split('.')
  return {
    integer,
    decimal: decimal ? `.${decimal}` : '',
  }
}

const checkoutItemPriceParts = (item: CartItem) => pricePartsFromCents(checkoutItemPayableCents(item))

const checkoutItemOriginalPriceParts = (item: CartItem) => pricePartsFromCents(checkoutItemOriginalCents(item))

const checkoutItemHasPriceDiscount = (item: CartItem) => {
  const originalCents = checkoutItemOriginalCents(item)
  const payableCents = checkoutItemPayableCents(item)
  return originalCents !== null && payableCents !== null && originalCents > payableCents
}

const normalizeStockNumber = (value: unknown) => {
  const numberValue = Number(value)
  if (!Number.isFinite(numberValue)) return 0
  return Math.max(Math.floor(numberValue), 0)
}

const normalizeManualStockTotal = (value: unknown) => {
  const numberValue = Number(value)
  if (!Number.isFinite(numberValue)) return 0
  const integerValue = Math.floor(numberValue)
  if (integerValue === -1) return -1
  return Math.max(integerValue, 0)
}

const hasItemStockSnapshot = (item: CartItem) => Boolean(String(item.skuStockSnapshotAt || '').trim())

const shouldEnforceItemStock = (item: CartItem) => {
  if (item.fulfillmentType === 'auto') return true
  if (item.fulfillmentType === 'upstream') return true
  if (item.fulfillmentType !== 'manual') return false
  if (!hasItemStockSnapshot(item)) return false
  const total = normalizeManualStockTotal(item.skuManualStockTotal)
  if (total === -1) return false
  if (item.skuStockEnforced === true) return true
  if (item.skuStockEnforced === false) return false
  return true
}

const itemAvailableStock = (item: CartItem) => {
  if (!shouldEnforceItemStock(item)) return null
  if (item.fulfillmentType === 'upstream') {
    const upstreamStock = Number(item.skuUpstreamStock ?? 0)
    if (upstreamStock === -1) return null // 无限库存
    return Math.max(upstreamStock, 0)
  }
  if (item.fulfillmentType === 'auto') {
    return normalizeStockNumber(item.skuAutoStockAvailable)
  }
  const total = normalizeManualStockTotal(item.skuManualStockTotal)
  if (total === -1) return null
  return total
}

const itemMaxQuantity = (item: CartItem) => {
  const available = itemAvailableStock(item)
  const purchaseLimit = itemPurchaseLimit(item)
  if (available === null && purchaseLimit === null) return Number.MAX_SAFE_INTEGER
  if (available === null) return purchaseLimit || 0
  if (purchaseLimit === null) return Math.max(available, 0)
  return Math.max(Math.min(available, purchaseLimit), 0)
}

const itemStockExceeded = (item: CartItem) => {
  const qty = parseInteger(item.quantity)
  if (qty === null) return true
  return qty > itemMaxQuantity(item)
}

const itemMinNotMet = (item: CartItem) => {
  const qty = parseInteger(item.quantity)
  if (qty === null) return false
  return qty < itemPurchaseMin(item)
}

const itemStockHint = (item: CartItem) => {
  const available = itemAvailableStock(item)
  const purchaseLimit = itemPurchaseLimit(item)
  const maxQuantity = itemMaxQuantity(item)
  if (itemMinNotMet(item)) {
    return t('cart.minPurchaseNotMet', { count: itemPurchaseMin(item) })
  }
  if (available === null && purchaseLimit === null) return ''
  if (maxQuantity <= 0) return t('cart.stockOut')
  if (itemStockExceeded(item)) {
    if (purchaseLimit !== null && maxQuantity === purchaseLimit && (available === null || purchaseLimit < available)) {
      return t('cart.maxPurchaseExceeded', { count: purchaseLimit })
    }
    return t('cart.stockExceeded', { count: maxQuantity })
  }
  if (available === null) return ''
  if (available <= 0) return t('cart.stockOut')
  return t('cart.stockRemaining', { count: available })
}
</script>
