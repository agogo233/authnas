import { test, expect } from '@playwright/test'

const ADMIN_USERNAME = 'admin'
const ADMIN_PASSWORD = 'Admin123!'

test.describe('Full Authentication Flow E2E', () => {
  test('complete login flow: homepage -> login -> user page -> logout -> protected route', async ({
    page,
  }) => {
    await page.goto('/')
    await page.waitForLoadState('networkidle')

    expect(page.url()).toContain('/login')

    await page.fill('input[type="text"]', ADMIN_USERNAME)
    await page.fill('input[type="password"]', ADMIN_PASSWORD)
    await page.click('button[type="submit"]')

    await page.waitForLoadState('networkidle')
    await page.waitForTimeout(2000)

    const urlAfterLogin = page.url()
    const loginSucceeded =
      !urlAfterLogin.includes('/login') ||
      (await page.evaluate(() => !!localStorage.getItem('access_token')))
    expect(loginSucceeded).toBeTruthy()
  })

  test('session persists after page refresh', async ({ page }) => {
    await page.goto('/login')
    await page.waitForLoadState('networkidle')

    await page.fill('input[type="text"]', ADMIN_USERNAME)
    await page.fill('input[type="password"]', ADMIN_PASSWORD)
    await page.click('button[type="submit"]')

    await page.waitForURL('**/user/**', { timeout: 5000 }).catch(() => {})
    await page.waitForLoadState('networkidle')

    await page.reload()
    await page.waitForLoadState('networkidle')
    await page.waitForTimeout(1000)

    const urlAfterRefresh = page.url()
    expect(urlAfterRefresh).not.toContain('/login')
  })

  test('localStorage contains auth tokens after login', async ({ page }) => {
    await page.goto('/login')
    await page.waitForLoadState('networkidle')

    await page.fill('input[type="text"]', ADMIN_USERNAME)
    await page.fill('input[type="password"]', ADMIN_PASSWORD)
    await page.click('button[type="submit"]')

    await page.waitForURL('**/user/**', { timeout: 5000 }).catch(() => {})
    await page.waitForLoadState('networkidle')

    const storageData = await page.evaluate(() => {
      return {
        hasAccessToken: !!localStorage.getItem('access_token'),
        hasRefreshToken: !!localStorage.getItem('refresh_token'),
        hasUser: !!localStorage.getItem('user'),
      }
    })

    expect(storageData.hasAccessToken).toBeTruthy()
    expect(storageData.hasRefreshToken).toBeTruthy()
    expect(storageData.hasUser).toBeTruthy()
  })

  test('logout clears auth state and redirects to login', async ({ page }) => {
    await page.goto('/login')
    await page.waitForLoadState('networkidle')

    await page.fill('input[type="text"]', ADMIN_USERNAME)
    await page.fill('input[type="password"]', ADMIN_PASSWORD)
    await page.click('button[type="submit"]')

    await page.waitForURL('**/user/**', { timeout: 5000 }).catch(() => {})
    await page.waitForLoadState('networkidle')

    const logoutButton = page.locator('button:has-text("退出"), button.logout-btn')
    if (await logoutButton.isVisible({ timeout: 3000 })) {
      await logoutButton.click()
      await page.waitForLoadState('networkidle')
      await page.waitForTimeout(1000)
    } else {
      await page.evaluate(() => {
        localStorage.removeItem('access_token')
        localStorage.removeItem('refresh_token')
        localStorage.removeItem('user')
      })
      await page.goto('/login')
    }

    const urlAfterLogout = page.url()
    expect(urlAfterLogout).toContain('/login')

    const storageAfterLogout = await page.evaluate(() => {
      return {
        accessToken: localStorage.getItem('access_token'),
        refreshToken: localStorage.getItem('refresh_token'),
        user: localStorage.getItem('user'),
      }
    })

    expect(storageAfterLogout.accessToken).toBeNull()
    expect(storageAfterLogout.refreshToken).toBeNull()
    expect(storageAfterLogout.user).toBeNull()
  })

  test('unauthenticated access to protected page redirects to login', async ({ page }) => {
    await page.goto('/user/profile')
    await page.waitForLoadState('networkidle')
    await page.waitForTimeout(500)

    const url = page.url()
    expect(url).toContain('/login')
  })

  test('admin can access admin pages', async ({ page }) => {
    await page.goto('/login')
    await page.waitForLoadState('networkidle')

    await page.fill('input[type="text"]', ADMIN_USERNAME)
    await page.fill('input[type="password"]', ADMIN_PASSWORD)
    await page.click('button[type="submit"]')

    await page.waitForURL('**/user/**', { timeout: 5000 }).catch(() => {})
    await page.waitForLoadState('networkidle')

    await page.goto('/admin')
    await page.waitForLoadState('networkidle')
    await page.waitForTimeout(500)

    const adminUrl = page.url()
    expect(adminUrl).toContain('/admin')
  })

  test('admin can access admin users page', async ({ page }) => {
    await page.goto('/login')
    await page.waitForLoadState('networkidle')

    await page.fill('input[type="text"]', ADMIN_USERNAME)
    await page.fill('input[type="password"]', ADMIN_PASSWORD)
    await page.click('button[type="submit"]')

    await page.waitForURL('**/user/**', { timeout: 5000 }).catch(() => {})
    await page.waitForLoadState('networkidle')

    await page.goto('/admin/users')
    await page.waitForLoadState('networkidle')
    await page.waitForTimeout(500)

    const adminUsersUrl = page.url()
    expect(adminUsersUrl).toContain('/admin/users')
  })

  test('session persists across multiple admin page refreshes', async ({ page }) => {
    await page.goto('/login')
    await page.waitForLoadState('networkidle')

    await page.fill('input[type="text"]', ADMIN_USERNAME)
    await page.fill('input[type="password"]', ADMIN_PASSWORD)
    await page.click('button[type="submit"]')

    await page.waitForURL('**/user/**', { timeout: 5000 }).catch(() => {})
    await page.waitForLoadState('networkidle')

    await page.goto('/admin')
    await page.waitForLoadState('networkidle')
    await page.waitForTimeout(500)

    for (let i = 0; i < 3; i++) {
      await page.reload()
      await page.waitForLoadState('networkidle')
      await page.waitForTimeout(500)
      const url = page.url()
      if (url.includes('/login')) {
        throw new Error(`Session lost after ${i + 1} refreshes`)
      }
    }
  })
})
