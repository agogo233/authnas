import { test, expect } from '@playwright/test'

test.describe('Admin Users Management E2E (Public Behavior)', () => {
  test('should redirect to login when accessing /admin/users without auth', async ({ page }) => {
    await page.goto('/admin/users')
    await page.waitForLoadState('networkidle')

    await expect(page).toHaveURL(/\/login/, { timeout: 5000 })
  })

  test('should redirect non-admin user away from /admin/users', async ({ page, context }) => {
    await context.addInitScript(() => {
      localStorage.setItem('access_token', 'mock-user-token')
      localStorage.setItem('refresh_token', 'mock-user-refresh-token')
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

    await page.goto('/login')
    await page.waitForLoadState('networkidle')
    await page.goto('/admin/users')
    await page.waitForLoadState('networkidle')
    await page.waitForTimeout(1000)

    const currentUrl = page.url()
    const urlPath = new URL(currentUrl).pathname
    expect(urlPath).not.toBe('/admin/users')
  })

  test('should navigate between admin pages when authenticated', async ({ page }) => {
    await page.goto('/login')
    await page.waitForLoadState('networkidle')
    await page.fill('input[placeholder="请输入用户名"]', 'admin')
    await page.fill('input[type="password"]', 'Admin123!')
    await page.getByRole('button', { name: '登录', exact: true }).click()
    await page.waitForLoadState('networkidle')
    await page.waitForTimeout(2000)

    await page.goto('/admin/users')
    await page.waitForLoadState('networkidle')
    await expect(page).toHaveURL(/\/admin\/users/, { timeout: 10000 })

    await page.goto('/admin/groups')
    await page.waitForLoadState('networkidle')
    await expect(page).toHaveURL(/\/admin\/groups/)

    await page.goto('/admin/clients')
    await page.waitForLoadState('networkidle')
    await expect(page).toHaveURL(/\/admin\/clients/)
  })
})
