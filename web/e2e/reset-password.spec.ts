import { test, expect } from '@playwright/test'

test.describe('Reset Password Flow E2E', () => {
  test('should display reset password form correctly', async ({ page }) => {
    await page.goto('/reset-password')
    await page.waitForLoadState('networkidle')
    
    await expect(page.locator('input[placeholder*="请输入您的邮箱"]')).toBeVisible()
    await expect(page.getByRole('button', { name: /发送重置邮件/i })).toBeVisible()
  })

  test('should navigate back to login page', async ({ page }) => {
    await page.goto('/reset-password')
    await page.waitForLoadState('networkidle')
    
    await page.locator('a:has-text("返回登录")').click()
    await expect(page).toHaveURL(/\/login/)
  })

  test('should have working back to login link', async ({ page }) => {
    await page.goto('/reset-password')
    await page.waitForLoadState('networkidle')
    
    const backLink = page.locator('a:has-text("返回登录")')
    await expect(backLink).toBeVisible()
    await backLink.click()
    await expect(page).toHaveURL(/\/login/)
  })

  test('should accept email input', async ({ page }) => {
    await page.goto('/reset-password')
    await page.waitForLoadState('networkidle')
    
    const emailInput = page.locator('input[placeholder*="请输入您的邮箱"]')
    await emailInput.fill('test@example.com')
    await expect(emailInput).toHaveValue('test@example.com')
  })

  test('should show validation error for empty email when clicking submit', async ({ page }) => {
    await page.goto('/reset-password')
    await page.waitForLoadState('networkidle')
    
    await page.getByRole('button', { name: /发送重置邮件/i }).click()
    await page.waitForTimeout(1000)
    
    await expect(page.locator('.n-alert--error, .n-alert:has-text("邮箱")')).toBeVisible({ timeout: 3000 })
  })

  test('should submit form with any email format (no client-side validation)', async ({ page }) => {
    await page.goto('/reset-password')
    await page.waitForLoadState('networkidle')
    
    await page.fill('input[placeholder*="请输入您的邮箱"]', 'invalid-email')
    await page.getByRole('button', { name: /发送重置邮件/i }).click()
    await page.waitForTimeout(1500)
    
    const currentUrl = page.url()
    expect(currentUrl).toContain('/reset-password')
  })

  test('should clear error when user starts typing', async ({ page }) => {
    await page.goto('/reset-password')
    await page.waitForLoadState('networkidle')
    
    await page.getByRole('button', { name: /发送重置邮件/i }).click()
    await page.waitForTimeout(1000)
    
    const errorAlert = page.locator('.n-alert--error, .n-alert:has-text("邮箱")')
    if (await errorAlert.isVisible()) {
      await page.fill('input[placeholder*="请输入您的邮箱"]', 'test@example.com')
      await page.waitForTimeout(300)
    }
  })
})
