// @vitest-environment node

import { readFileSync } from 'node:fs'
import { dirname, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'
import { describe, expect, it } from 'vitest'

const stylePath = resolve(dirname(fileURLToPath(import.meta.url)), '../style.css')
const css = readFileSync(stylePath, 'utf8')

const classBlock = (className: string) => {
  const escaped = className.replace(/[.*+?^${}()|[\]\\]/g, '\\$&')
  const match = css.match(new RegExp(`\\.${escaped}\\s*\\{([\\s\\S]*?)\\n\\s*\\}`, 'm'))
  return match?.[1] ?? ''
}

describe('storefront theme utilities', () => {
  it('keeps fork checkout input styling utilities', () => {
    expect(css).toContain('.form-input-lg')
    expect(classBlock('form-input-lg')).toContain('@apply px-4 py-3 text-sm')
    expect(css).toMatch(
      /\.form-input:focus,\s*\.form-input-lg:focus,\s*\.form-input-compact:focus\s*\{[\s\S]*box-shadow:\s*0 0 0 3px var\(--ui-focus-ring\);/,
    )
  })

  it('keeps opaque themed panel backgrounds used by fork dialogs', () => {
    expect(classBlock('theme-panel')).toContain('background-color: var(--ui-bg-elevated);')
    expect(classBlock('theme-panel')).toContain('border-color: var(--ui-border);')
    expect(classBlock('theme-panel-strong')).toContain('background-color: var(--ui-bg-overlay-strong);')
    expect(classBlock('theme-surface-soft')).toContain('background-color: var(--ui-bg-soft);')
  })

  it('keeps card color tokens used by the synced announcement modal', () => {
    expect(css).toContain('--color-card: var(--ui-bg-elevated);')
    expect(css).toContain('--color-card-foreground: var(--ui-text-primary);')
  })
})
