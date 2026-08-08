import test from 'node:test'
import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'

const appSource = readFileSync(new URL('../src/App.vue', import.meta.url), 'utf8')
const routerSource = readFileSync(new URL('../src/router/index.ts', import.meta.url), 'utf8')
const vaultLayoutSource = readFileSync(
  new URL('../src/templates/vault/layout/VaultLayout.vue', import.meta.url),
  'utf8',
)

test('support route locks the storefront viewport and hides the page footer', () => {
  assert.match(routerSource, /path:\s*['"]\/support['"]/)
  assert.match(routerSource, /name:\s*['"]support['"]/)
  assert.match(routerSource, /meta:\s*\{\s*hideFooter:\s*true,\s*lockViewport:\s*true\s*\}/)
})

test('classic and vault shells keep a support page mounted outside RouterView', () => {
  const supportMounts = [...appSource.matchAll(/<SupportPage\b[\s\S]*?\/>/g)]

  assert.equal(supportMounts.length, 2)
  for (const mount of supportMounts) {
    assert.match(mount[0], /v-show="route\.name === 'support'"/)
    assert.match(mount[0], /:active="route\.name === 'support'"/)
  }
  assert.match(appSource, /<ErrorBoundary\s+v-if="route\.name !== 'support'"/)
  assert.match(appSource, /<VaultLayout[^>]*:immersive="route\.meta\.lockViewport === true"/)
})

test('vault immersive mode keeps its header and removes footer overflow', () => {
  assert.match(vaultLayoutSource, /immersive\?:\s*boolean/)
  assert.match(vaultLayoutSource, /h-\[100dvh\]/)
  assert.match(vaultLayoutSource, /<footer\s+v-if="!props\.immersive"/)
})
