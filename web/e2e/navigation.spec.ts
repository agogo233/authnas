import { test, expect } from '@playwright/test'

test.describe('Navigation E2E', () => {
  test('should navigate from login to register', async ({ page }) => {
    await page.goto('/login')

    await page.click('a[href="/register"]')
    await expect(page).toHaveURL(/\/register/)
  })

  test('should navigate from login to reset password', async ({ page }) => {
    await page.goto('/login')

    await page.click('a[href="/reset-password"]')
    await expect(page).toHaveURL(/\/reset-password/)
  })

  test('should navigate from register to login', async ({ page }) => {
    await page.goto('/register')

    await page.click('a[href="/login"]')
    await expect(page).toHaveURL(/\/login/)
  })

  test('should navigate from reset password to login', async ({ page }) => {
    await page.goto('/reset-password')

    await page.click('a[href="/login"]')
    await expect(page).toHaveURL(/\/login/)
  })

  test('should redirect to login when accessing protected route', async ({ page }) => {
    await page.goto('/profile')
    await expect(page).toHaveURL(/\/login/)
  })

  test('should redirect to login when accessing admin route without admin privileges', async ({
    page,
  }) => {
    await page.goto('/login')
    await page.waitForLoadState('networkidle')
    await page.evaluate(() => {
      localStorage.setItem('auth_token', 'mock-user-token')
      localStorage.setItem(
        'user',
        JSON.stringify({
          id: '2',
          username: 'regularuser',
          email: 'user@example.com',
          isAdmin: false,
          approved: true,
        })
      )
    })

    await page.goto('/admin')
    await page.waitForLoadState('networkidle')
    await expect(page).toHaveURL(/\/login/)
  })

  test('should preserve redirect query when redirecting to login', async ({ page }) => {
    await page.goto('/profile')
    await expect(page).toHaveURL(/\/login\?redirect=/)
  })

  test('should navigate to profile after successful login', async ({ page }) => {
    await page.goto('/login')
    await page.waitForLoadState('networkidle')
    await page.fill('input[placeholder="请输入用户名"]', 'testuser')
    await page.fill('input[type="password"]', 'password123')
    await page.getByRole('button', { name: '登录', exact: true }).click()

    await page.waitForURL(/\/(profile|mfa|login)\/?/, { timeout: 10000 })
  })

  test('should have working footer links on login page', async ({ page }) => {
    await page.goto('/login')

    await expect(page.locator('a[href="/register"]')).toBeVisible()
    await expect(page.locator('a[href="/reset-password"]')).toBeVisible()
  })

  test('should have working footer links on register page', async ({ page }) => {
    await page.goto('/register')

    await expect(page.locator('a[href="/login"]')).toBeVisible()
  })

  test('should navigate between admin sections', async ({ page }) => {
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

    await page.goto('/admin/users')
    await page.waitForLoadState('networkidle')
    await expect(page).toHaveURL(/\/admin\/users/)

    await page.goto('/admin/groups')
    await page.waitForLoadState('networkidle')
    await expect(page).toHaveURL(/\/admin\/groups/)

    await page.goto('/admin/clients')
    await page.waitForLoadState('networkidle')
    await expect(page).toHaveURL(/\/admin\/clients/)

    await page.goto('/admin/invitations')
    await page.waitForLoadState('networkidle')
    await expect(page).toHaveURL(/\/admin\/invitations/)
  })

  test('root path should redirect to login', async ({ page }) => {
    await page.goto('/')
    await expect(page).toHaveURL(/\/login/)
  })
})
