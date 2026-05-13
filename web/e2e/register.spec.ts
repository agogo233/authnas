import { test, expect } from '@playwright/test'

test.describe('Register Flow E2E', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/register')
  })

  test('should display registration form correctly', async ({ page }) => {
    await expect(page.locator('input[placeholder*="请输入邮箱"]')).toBeVisible()
    await expect(page.locator('input[placeholder="请输入用户名"]')).toBeVisible()
    await expect(page.locator('input[type="password"]')).toBeVisible()
    await expect(page.locator('input[placeholder*="请输入邀请码"]')).toBeVisible()
    await expect(page.getByRole('button', { name: '注册', exact: true })).toBeVisible()
  })

  test('should show validation error for empty fields', async ({ page }) => {
    await page.waitForLoadState('networkidle')
    await page.getByRole('button', { name: '注册', exact: true }).click()
    await page.waitForTimeout(500)
    await expect(page.locator('.n-alert')).toContainText(/请填写所有必填字段/i)
  })

  test('should show validation error for missing password', async ({ page }) => {
    await page.waitForLoadState('networkidle')
    await page.fill('input[placeholder*="请输入邮箱"]', 'test@example.com')
    await page.fill('input[placeholder="请输入用户名"]', 'testuser')
    await page.getByRole('button', { name: '注册', exact: true }).click()
    await page.waitForTimeout(500)
    await expect(page.locator('.n-alert')).toContainText(/请填写所有必填字段/i)
  })

  test('should reject weak passwords', async ({ page }) => {
    await page.waitForLoadState('networkidle')
    await page.fill('input[placeholder*="请输入邮箱"]', 'test@example.com')
    await page.fill('input[placeholder="请输入用户名"]', 'testuser')
    await page.fill('input[type="password"]', 'weak')
    await page.getByRole('button', { name: '注册', exact: true }).click()
    await page.waitForTimeout(500)
    await expect(page.locator('.n-alert')).toContainText(/密码强度太弱/i)
  })

  test('should display password strength indicator', async ({ page }) => {
    const passwordInput = page.locator('input[type="password"]')

    await passwordInput.fill('abc')
    await expect(page.locator('.strength-text, .password-strength')).toBeVisible()

    await passwordInput.fill('Password123!')
    await expect(page.locator('.strength-text, .password-strength')).toBeVisible()
  })

  test('should navigate to login page', async ({ page }) => {
    await page.click('a[href="/login"]')
    await expect(page).toHaveURL(/\/login/)
  })

  test('should pass invitation code when provided', async ({ page }) => {
    await page.fill('input[placeholder*="请输入邮箱"]', 'test@example.com')
    await page.fill('input[placeholder="请输入用户名"]', 'testuser')
    await page.fill('input[type="password"]', 'StrongPass123!')
    await page.fill('input[placeholder*="请输入邀请码"]', 'INVITE123')

    await page.getByRole('button', { name: '注册', exact: true }).click()

    await page.waitForTimeout(500)
  })
})
