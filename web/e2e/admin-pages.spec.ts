import { test, expect } from '@playwright/test'

test.describe('Admin Pages E2E (Public Behavior)', () => {
  test('should redirect to login when accessing /admin/groups without auth', async ({ page }) => {
    await page.goto('/admin/groups')
    await page.waitForLoadState('networkidle')

    await expect(page).toHaveURL(/\/login/, { timeout: 5000 })
  })

  test('should redirect to login when accessing /admin/clients without auth', async ({ page }) => {
    await page.goto('/admin/clients')
    await page.waitForLoadState('networkidle')

    await expect(page).toHaveURL(/\/login/, { timeout: 5000 })
  })

  test('should redirect to login when accessing /admin/invitations without auth', async ({
    page,
  }) => {
    await page.goto('/admin/invitations')
    await page.waitForLoadState('networkidle')

    await expect(page).toHaveURL(/\/login/, { timeout: 5000 })
  })

  test('should redirect to login when accessing /admin/proxyauth without auth', async ({
    page,
  }) => {
    await page.goto('/admin/proxyauth')
    await page.waitForLoadState('networkidle')

    await expect(page).toHaveURL(/\/login/, { timeout: 5000 })
  })

  test('should redirect to login when accessing /admin/settings without auth', async ({ page }) => {
    await page.goto('/admin/settings')
    await page.waitForLoadState('networkidle')

    await expect(page).toHaveURL(/\/login/, { timeout: 5000 })
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
    await expect(page).toHaveURL(/\/admin\/users/)

    await page.goto('/admin/groups')
    await page.waitForLoadState('networkidle')
    await expect(page).toHaveURL(/\/admin\/groups/)

    await page.goto('/admin/clients')
    await page.waitForLoadState('networkidle')
    await expect(page).toHaveURL(/\/admin\/clients/)
  })

  test('should redirect non-admin user from /admin/groups', async ({ page }) => {
    await page.context().addInitScript(() => {
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

    await page.goto('/admin/groups')
    await page.waitForLoadState('networkidle')
    await page.waitForTimeout(1000)

    const currentUrl = page.url()
    const urlPath = new URL(currentUrl).pathname
    expect(urlPath).not.toBe('/admin/groups')
  })

  test('should redirect non-admin user from /admin/clients', async ({ page }) => {
    await page.context().addInitScript(() => {
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

    await page.goto('/admin/clients')
    await page.waitForLoadState('networkidle')
    await page.waitForTimeout(1000)

    const currentUrl = page.url()
    const urlPath = new URL(currentUrl).pathname
    expect(urlPath).not.toBe('/admin/clients')
  })
})
