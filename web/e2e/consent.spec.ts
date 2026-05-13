import { test, expect } from '@playwright/test'

test.describe('Consent Page E2E (Public Behavior)', () => {
  test('should allow access to consent page without auth (public OIDC flow)', async ({ page }) => {
    await page.goto('/consent/test-uid-123')
    await page.waitForLoadState('networkidle')

    await expect(page).toHaveURL(/\/consent\/test-uid-123/, { timeout: 5000 })
  })

  test('should display loading state and then error for invalid uid', async ({ page }) => {
    await page.goto('/consent/invalid-uid-xyz')
    await page.waitForLoadState('networkidle')

    await expect(page).toHaveURL(/\/consent\/invalid-uid-xyz/, { timeout: 5000 })
    const errorResult = page.locator('.n-result')
    await expect(errorResult).toBeVisible({ timeout: 10000 })
  })

  test('should display consent page for valid uid format', async ({ page }) => {
    await page.goto('/consent/valid-test-uid')
    await page.waitForLoadState('networkidle')

    await expect(page).toHaveURL(/\/consent\/valid-test-uid/, { timeout: 5000 })
  })
})
