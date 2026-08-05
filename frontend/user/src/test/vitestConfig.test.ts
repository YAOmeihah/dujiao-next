// @vitest-environment node

import { describe, expect, it } from 'vitest'
import viteConfig from '../../vite.config'

const resolveConfig = () =>
  typeof viteConfig === 'function'
    ? viteConfig({
        command: 'serve',
        mode: 'test',
        isSsrBuild: false,
        isPreview: false,
      })
    : viteConfig

describe('vitest config', () => {
  it('limits discovery to source tests and excludes worktree and node:test suites', () => {
    const config = resolveConfig()
    const testConfig = config.test ?? {}

    expect(testConfig.include).toEqual(['src/**/*.{test,spec}.{js,mjs,cjs,ts,mts,cts,jsx,tsx}'])
    expect(testConfig.exclude).toEqual(
      expect.arrayContaining(['.worktrees/**', 'scripts/test/**', '.github/scripts/test/**']),
    )
  })
})
