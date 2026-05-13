import { test, expect } from '@playwright/test'

test.describe('Login Flow E2E', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/login')
  })

  test('should display login form correctly', async ({ page }) => {
    await expect(page.locator('input[placeholder="请输入用户名"]')).toBeVisible()
    await expect(page.locator('input[type="password"]')).toBeVisible()
    await expect(page.getByRole('button', { name: '登录', exact: true })).toBeVisible()
  })

  test('should show validation error for empty credentials', async ({ page }) => {
    await page.waitForLoadState('networkidle')
    await page.getByRole('button', { name: '登录', exact: true }).click()
    await page.waitForTimeout(500)
    const alert = page.locator('.n-alert')
    await expect(alert).toBeVisible({ timeout: 3000 })
    await expect(alert).toContainText(/请输入用户名和密码/i)
  })

  test('should show validation error for empty password only', async ({ page }) => {
    await page.waitForLoadState('networkidle')
    await page.fill('input[placeholder="请输入用户名"]', 'testuser')
    await page.getByRole('button', { name: '登录', exact: true }).click()
    await page.waitForTimeout(500)
    const alert = page.locator('.n-alert')
    await expect(alert).toBeVisible({ timeout: 3000 })
    await expect(alert).toContainText(/请输入用户名和密码/i)
  })

  test('should navigate to register page', async ({ page }) => {
    await page.click('a[href="/register"]')
    await expect(page).toHaveURL(/\/register/)
  })

  test('should navigate to reset password page', async ({ page }) => {
    await page.click('a[href="/reset-password"]')
    await expect(page).toHaveURL(/\/reset-password/)
  })

  test('should remember username when remember me is checked', async ({ page }) => {
    await page.fill('input[placeholder="请输入用户名"]', 'testuser')
    await page.locator('.n-checkbox').click()
    await page.getByRole('button', { name: '登录', exact: true }).click()

    await page.waitForTimeout(500)

    const usernameInput = page.locator('input[placeholder="请输入用户名"]')
    await expect(usernameInput).toHaveValue('testuser')
  })

  test('should load remembered username from localStorage', async ({ page }) => {
    await page.evaluate(() => {
      localStorage.setItem('remembered_username', 'storeduser')
    })

    await page.reload()
    await page.goto('/login')

    const usernameInput = page.locator('input[placeholder="请输入用户名"]')
    await expect(usernameInput).toHaveValue('storeduser')
  })

  test('should have passkey login option when username is entered', async ({ page }) => {
    await page.fill('input[placeholder="请输入用户名"]', 'testuser')

    const passkeyButton = page.getByRole('button', { name: /通行密钥/i })
    await expect(passkeyButton).toBeVisible()
  })
})
