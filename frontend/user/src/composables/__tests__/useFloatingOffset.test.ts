import { describe, expect, it } from 'vitest'
import { resolveFloatingActionBottomOffset } from '../useFloatingOffset'

describe('resolveFloatingActionBottomOffset', () => {
  it('keeps the default bottom offset when no floating bar is present', () => {
    const bottom = resolveFloatingActionBottomOffset({
      defaultBottom: 80,
      anchorHeight: 0,
      anchorBottom: 0,
      gap: 16,
    })

    expect(bottom).toBe(80)
  })

  it('lifts the action above a checkout sticky bar when it would overlap', () => {
    const bottom = resolveFloatingActionBottomOffset({
      defaultBottom: 80,
      anchorHeight: 69,
      anchorBottom: 56,
      gap: 16,
    })

    expect(bottom).toBe(141)
  })
})
