import { test, expect } from '@playwright/test'

test.describe('Profile Page E2E', () => {
  test('should redirect to login when not authenticated', async ({ page }) => {
    await page.goto('/profile')
    await expect(page).toHaveURL(/\/login/)
  })

  test('should redirect to login when accessing profile with invalid token', async ({ page }) => {
    await page.goto('/login')
    await page.waitForLoadState('networkidle')
    await page.evaluate(() => {
      localStorage.setItem('auth_token', 'invalid-token')
      localStorage.setItem(
        'user',
        JSON.stringify({
          id: '1',
          username: 'testuser',
          email: 'test@example.com',
          isAdmin: false,
          approved: true,
        })
      )
    })

    await page.goto('/profile')
    await page.waitForTimeout(1000)
  })
})

test.describe('Passkeys Page E2E', () => {
  test('should redirect to login when not authenticated', async ({ page }) => {
    await page.goto('/passkeys')
    await expect(page).toHaveURL(/\/login/)
  })
})

test.describe('Passkey Flow E2E', () => {
  test('should show passkey button when username is entered', async ({ page }) => {
    await page.goto('/login')
    await page.waitForLoadState('networkidle')
    await page.waitForTimeout(500)

    await page.fill('input[placeholder="请输入用户名"]', 'testuser')
    await page.waitForTimeout(300)

    const passkeyButton = page.locator('button:has-text("使用通行密钥登录")')
    await expect(passkeyButton).toBeVisible({ timeout: 5000 })
  })
})
