import { test, expect } from '@playwright/test'

test.describe('Verify Email Flow E2E', () => {
  test('should display verification result state with valid or invalid code', async ({ page }) => {
    await page.goto('/verify-email?code=valid-test-code')
    
    await page.waitForTimeout(1500)
    
    const result = page.locator('.n-result')
    await expect(result).toBeVisible({ timeout: 5000 })
  })

  test('should display verification error state when code is missing', async ({ page }) => {
    await page.goto('/verify-email')
    
    await page.waitForTimeout(1500)
    
    await expect(page.getByText('验证失败')).toBeVisible({ timeout: 5000 })
  })

  test('should display verification error state when code is invalid', async ({ page }) => {
    await page.goto('/verify-email?code=invalid-code')
    
    await page.waitForTimeout(1500)
    
    await expect(page.getByText('验证失败')).toBeVisible({ timeout: 5000 })
  })

  test('should show resend verification button on error state', async ({ page }) => {
    await page.goto('/verify-email')
    
    await page.waitForTimeout(1500)
    
    const resendButton = page.getByRole('button', { name: '重新发送验证邮件', exact: true })
    await expect(resendButton).toBeVisible({ timeout: 5000 })
  })

  test('should navigate to login page', async ({ page }) => {
    await page.goto('/verify-email')
    
    await page.waitForTimeout(1500)
    
    const loginButton = page.getByRole('button', { name: '前往登录', exact: true })
    await expect(loginButton).toBeVisible({ timeout: 5000 })
    
    await loginButton.click()
    await expect(page).toHaveURL(/\/login/)
  })

  test('should show resend button on error state', async ({ page }) => {
    await page.goto('/verify-email')
    
    await page.waitForTimeout(1500)
    
    const resendButton = page.getByRole('button', { name: '重新发送验证邮件', exact: true })
    await expect(resendButton).toBeVisible({ timeout: 5000 })
  })

  test('should show result state after verification attempt', async ({ page }) => {
    await page.goto('/verify-email?code=test-code')
    
    await page.waitForTimeout(1500)
    
    const result = page.locator('.n-result')
    await expect(result).toBeVisible({ timeout: 3000 })
  })
})
