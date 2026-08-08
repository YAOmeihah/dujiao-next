import { describe, expect, it } from 'vitest'
import { getSupportUrl } from '../supportUrl'

describe('getSupportUrl', () => {
  it('returns a trimmed absolute http or https support URL', () => {
    expect(getSupportUrl({ contact: { support_url: ' https://help.example.com/chat ' } }))
      .toBe('https://help.example.com/chat')
    expect(getSupportUrl({ contact: { support_url: 'http://help.example.com' } }))
      .toBe('http://help.example.com')
  })

  it('rejects missing, relative, non-http and malformed support URLs', () => {
    expect(getSupportUrl(null)).toBe('')
    expect(getSupportUrl({ contact: { support_url: '/support' } })).toBe('')
    expect(getSupportUrl({ contact: { support_url: 'javascript:alert(1)' } })).toBe('')
    expect(getSupportUrl({ contact: { support_url: 'ftp://help.example.com' } })).toBe('')
    expect(getSupportUrl({ contact: { support_url: 'https://' } })).toBe('')
  })
})
