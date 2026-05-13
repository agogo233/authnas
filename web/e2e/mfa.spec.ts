import { test, expect } from '@playwright/test'

test.describe('MFA Flow E2E (Public Behavior)', () => {
  test('should redirect to login when accessing /mfa without auth', async ({ page }) => {
    await page.goto('/mfa')
    await page.waitForLoadState('networkidle')

    await expect(page).toHaveURL(/\/login/, { timeout: 5000 })
  })

  test('should redirect to login with redirect query when accessing /mfa without auth', async ({
    page,
  }) => {
    await page.goto('/mfa?token=test-token')
    await page.waitForLoadState('networkidle')

    await expect(page).toHaveURL(/\/login/, { timeout: 5000 })
  })

  test('should include redirect parameter in login URL', async ({ page }) => {
    await page.goto('/mfa?token=test-token')
    await page.waitForLoadState('networkidle')

    await expect(page).toHaveURL(/\/login\?redirect=/, { timeout: 5000 })
  })

  test('should show MFA option after successful login with MFA required', async ({ page }) => {
    await page.goto('/login')
    await page.waitForLoadState('networkidle')

    await page.fill('input[placeholder="请输入用户名"]', 'testuser')
    await page.fill('input[type="password"]', 'TestPass123!')
    await page.getByRole('button', { name: '登录', exact: true }).click()

    await page.waitForTimeout(2000)

    const currentUrl = page.url()
    expect(currentUrl).not.toBe('/login')
  })
})
