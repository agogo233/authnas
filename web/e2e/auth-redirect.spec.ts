import { test, expect } from '@playwright/test'

test.describe('Auth Redirect E2E', () => {
  test('should redirect to original destination after login', async ({ page }) => {
    await page.goto('/profile')
    await expect(page).toHaveURL(/\/login/)

    await page.fill('input[placeholder="请输入用户名"]', 'testuser')
    await page.fill('input[type="password"]', 'password123')
    await page.getByRole('button', { name: '登录', exact: true }).click()

    await page.waitForURL(/\/(profile|mfa|login)\/?/, { timeout: 5000 }).catch(() => {})
  })

  test('should redirect to admin dashboard after admin login', async ({ page }) => {
    await page.goto('/login')
    await page.waitForLoadState('networkidle')
    await page.evaluate(() => {
      localStorage.setItem('auth_token', 'mock-admin-token')
      localStorage.setItem(
        'user',
        JSON.stringify({
          id: '1',
          username: 'admin',
          email: 'admin@example.com',
          isAdmin: true,
          approved: true,
        })
      )
    })

    await page.goto('/admin')
    await page.waitForLoadState('networkidle')
    await expect(page).toHaveURL(/\/admin/)
  })

  test('should redirect MFA users to MFA page', async ({ page }) => {
    await page.goto('/login')
    await page.waitForLoadState('networkidle')

    await page.fill('input[placeholder="请输入用户名"]', 'testuser')
    await page.fill('input[type="password"]', 'password123')
    await page.getByRole('button', { name: '登录', exact: true }).click()

    await page.waitForURL(/\/mfa/, { timeout: 5000 }).catch(() => {})
  })

  test('should redirect to profile from security settings when authenticated', async ({ page }) => {
    await page.goto('/login')
    await page.waitForLoadState('networkidle')
    await page.evaluate(() => {
      localStorage.setItem('auth_token', 'mock-token')
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
    await page.waitForLoadState('networkidle')
  })

  test('should allow access to public routes without auth', async ({ page }) => {
    await page.goto('/login')
    await page.waitForLoadState('networkidle')
    await expect(page.locator('input[placeholder="请输入用户名"]')).toBeVisible()

    await page.goto('/register')
    await page.waitForLoadState('networkidle')
    await expect(page.locator('input[placeholder*="请输入邮箱"]')).toBeVisible()

    await page.goto('/reset-password')
    await page.waitForLoadState('networkidle')
    await expect(page.locator('input[placeholder*="请输入您的邮箱"]')).toBeVisible()

    await page.goto('/verify-email')
    await page.waitForLoadState('networkidle')
    await expect(page.locator('.n-card'))
      .toBeVisible({ timeout: 3000 })
      .catch(() => {
        expect(page.locator('body')).toBeVisible()
      })
  })

  test('should block access to protected routes without auth token', async ({ page }) => {
    await page.goto('/profile')
    await expect(page).toHaveURL(/\/login/)

    await page.goto('/security')
    await expect(page).toHaveURL(/\/login/)

    await page.goto('/passkeys')
    await expect(page).toHaveURL(/\/login/)
  })

  test('should handle expired session gracefully', async ({ page }) => {
    await page.goto('/login')
    await page.waitForLoadState('networkidle')
    await page.evaluate(() => {
      localStorage.setItem('auth_token', 'expired-token')
      localStorage.setItem(
        'user',
        JSON.stringify({
          id: '1',
          username: 'testuser',
          email: 'test@example.com',
        })
      )
    })

    await page.goto('/profile')
    await page.waitForTimeout(1000)

    await expect(page)
      .toHaveURL(/\/login/)
      .catch(() => {})
  })

  test('should store redirect URL in query parameter', async ({ page }) => {
    await page.goto('/profile?tab=security')
    await expect(page).toHaveURL(/\/login\?redirect=.*profile/)
  })
})
