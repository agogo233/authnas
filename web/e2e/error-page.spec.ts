import { test, expect } from '@playwright/test'

test.describe('Error Page E2E', () => {
  test('should display 404 error page', async ({ page }) => {
    await page.goto('/error?status=404')
    await page.waitForTimeout(1000)
    
    await expect(page.getByText('页面不存在')).toBeVisible({ timeout: 5000 })
  })

  test('should display 403 error page', async ({ page }) => {
    await page.goto('/error?status=403')
    await page.waitForTimeout(1000)
    
    await expect(page.getByText('禁止访问')).toBeVisible({ timeout: 5000 })
  })

  test('should display 500 error page', async ({ page }) => {
    await page.goto('/error?status=500')
    await page.waitForTimeout(1000)
    
    await expect(page.getByText('服务器内部错误')).toBeVisible({ timeout: 5000 })
  })

  test('should display 401 error page', async ({ page }) => {
    await page.goto('/error?status=401')
    await page.waitForTimeout(1000)
    
    await expect(page.getByText('未授权')).toBeVisible({ timeout: 5000 })
  })

  test('should display 400 error page', async ({ page }) => {
    await page.goto('/error?status=400')
    await page.waitForTimeout(1000)
    
    await expect(page.getByText('请求错误')).toBeVisible({ timeout: 5000 })
  })

  test('should display custom error message when provided', async ({ page }) => {
    await page.goto('/error?status=500&message=Custom%20error%20message')
    await page.waitForTimeout(1000)
    
    await expect(page.getByText(/Custom error message/i)).toBeVisible({ timeout: 5000 })
  })

  test('should have go home button', async ({ page }) => {
    await page.goto('/error?status=404')
    await page.waitForTimeout(1000)
    
    const homeButton = page.getByRole('button', { name: '返回首页', exact: true })
    await expect(homeButton).toBeVisible()
  })

  test('should have go to login button', async ({ page }) => {
    await page.goto('/error?status=404')
    await page.waitForTimeout(1000)
    
    const loginButton = page.getByRole('button', { name: '前往登录', exact: true })
    await expect(loginButton).toBeVisible()
  })

  test('should navigate to login when go to login button is clicked', async ({ page }) => {
    await page.goto('/error?status=404')
    await page.waitForTimeout(1000)
    
    await page.getByRole('button', { name: '前往登录', exact: true }).click()
    await expect(page).toHaveURL(/\/login/)
  })

  test('should navigate to home when go home button is clicked', async ({ page }) => {
    await page.goto('/error?status=500')
    await page.waitForTimeout(1000)
    
    await page.getByRole('button', { name: '返回首页', exact: true }).click()
  })

  test('should use default 500 status when no status provided', async ({ page }) => {
    await page.goto('/error')
    await page.waitForTimeout(1000)
    
    await expect(page.getByText('服务器内部错误')).toBeVisible({ timeout: 5000 })
  })
})
