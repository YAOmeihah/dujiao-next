interface SupportConfig {
  contact?: {
    support_url?: unknown
  }
}

export const getSupportUrl = (config: unknown): string => {
  const value = String((config as SupportConfig | null)?.contact?.support_url || '').trim()
  if (!value) return ''

  try {
    const parsed = new URL(value)
    if ((parsed.protocol === 'http:' || parsed.protocol === 'https:') && parsed.hostname) {
      return value
    }
  } catch {
    return ''
  }

  return ''
}
