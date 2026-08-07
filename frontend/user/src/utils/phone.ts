const guestPhonePattern = /^\+?[0-9]{6,20}$/

export const normalizeGuestPhone = (value: unknown) =>
  String(value ?? '').trim().replace(/[\s()-]/g, '')

export const isGuestPhoneValid = (value: unknown) =>
  guestPhonePattern.test(normalizeGuestPhone(value))
