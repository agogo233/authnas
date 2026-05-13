import { test, expect } from '@playwright/test'

test.describe('Admin Dashboard E2E (Public Behavior)', () => {
  test('should redirect to login when accessing /admin without auth', async ({ page }) => {
    await page.goto('/admin')
    await page.waitForLoadState('networkidle')

    await expect(page).toHaveURL(/\/login/, { timeout: 5000 })
  })

  test('should redirect to login when non-admin user accesses /admin', async ({ page }) => {
    await page.goto('/login')
    await page.waitForLoadState('networkidle')
    await page.evaluate(() => {
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

    await page.goto('/admin')
    await page.waitForLoadState('networkidle')
    await page.waitForTimeout(1000)

    const currentUrl = page.url()
    const urlPath = new URL(currentUrl).pathname
    expect(urlPath).toBe('/login')
  })

  test('should navigate between admin pages when authenticated', async ({ page }) => {
    await page.route('**/api/**', async (route) => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({ success: true, data: [] }),
      })
    })

    await page.context().addInitScript(() => {
      localStorage.setItem('access_token', 'mock-admin-token')
      localStorage.setItem('refresh_token', 'mock-admin-refresh-token')
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

    await page.goto('/admin/users')
    await page.waitForLoadState('networkidle')
    await expect(page).toHaveURL(/\/admin\/users/, { timeout: 10000 })
  })
})
