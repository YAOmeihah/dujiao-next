# Vault Fork Adaptation Design

Date: 2026-06-26
Scope: user frontend only

## Context

Upstream v1.3.0 introduced the optional `vault` storefront template under `src/templates/vault/*`. The fork already customizes the classic checkout and guest order flows around:

- guest identity: `phone + order_password`
- optional guest email for notification/record keeping
- shipping address payload for products with `requires_shipping_address`
- existing fork payment behavior, including VPay compatibility, balance payment, channel filters, and channel amount limits

The vault template is intended to be usable, so it must follow the same fork rules as classic instead of upstream's email-only guest flow.

## Goals

1. Make vault guest checkout use `phone + order_password` as the required guest identity.
2. Keep guest email optional and validate it only when provided.
3. Show and submit `shipping_address` in vault checkout when any cart item requires shipping.
4. Keep the shared checkout payment behavior compatible with existing fork payment channels and upstream-added channels.
5. Align vault guest payment, guest order list, and guest order detail flows with the backend's phone-based guest APIs.

## Non-Goals

- Do not change backend API contracts.
- Do not remove or disable the classic template.
- Do not remove upstream payment providers added by the sync.
- Do not redesign the vault visual system beyond the minimum UI needed for missing fork fields.
- Do not review or fix unrelated upstream or historical project issues.

## Architecture

The existing shared composables should remain the source of business behavior. Vault page components should mostly render state and call handlers exposed by those composables.

`useCheckout` should expose the same guest and shipping primitives needed by both templates:

- guest phone state, validation, and input handler
- guest password state
- optional guest email state and validation
- shipping address state, validation, and payload builder
- guest captcha state for all configured captcha providers already supported by classic

Vault checkout should import and render the missing UI controls from the shared checkout state. Classic should continue to behave as it does today.

Guest order list, guest order detail, and guest payment composables should use `phone + order_password` for auth storage and API calls. Compatibility with previously stored `email` should be limited to optional display/record data, not authentication.

## Data Flow

Checkout preview and create-and-pay requests should send:

- `phone`: required for guest checkout
- `order_password`: required for guest checkout
- `email`: optional, included only when provided
- `shipping_address`: included by the shared order payload builder when a shipping address is required
- existing item, coupon, affiliate, manual form, channel, balance, and captcha payload fields

Guest payment and guest order lookup requests should read `guest_order_auth` from local storage using the fork shape:

```json
{
  "phone": "13800138000",
  "email": "optional@example.com",
  "order_password": "buyer password"
}
```

API calls should authenticate guest order access with `phone` and `order_password`.

## Error Handling

Vault should block submission and preview when:

- guest phone is empty or invalid
- guest order password is empty
- optional guest email is present but invalid
- shipping address is required but incomplete or invalid
- captcha is required but incomplete
- manual form validation fails
- no usable payment channel is selected when online payment is required

The user-facing error keys should reuse existing checkout, auth, and error translations wherever possible.

## Testing And Verification

Implementation should be verified with:

- `pnpm run build`
- `pnpm test`

Where practical, add or update focused tests for:

- vault/shared checkout guest payload includes `phone` and `order_password`
- optional email does not block guest checkout when empty
- shipping address is included for shipping-required carts
- guest order/payment auth uses `phone + order_password`

Manual smoke checks after build should cover vault checkout with:

- guest digital product
- guest shipping-required product
- logged-in checkout
- guest order lookup/payment continuation

