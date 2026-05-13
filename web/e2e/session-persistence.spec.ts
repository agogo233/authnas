import { test, expect } from '@playwright/test'

const ADMIN_USERNAME = 'admin'
const ADMIN_PASSWORD = 'Admin123!'

test.describe('Session Persistence E2E', () => {
  test('should persist login state after page refresh', async ({ page }) => {
    await page.goto('/login')
    await page.waitForLoadState('networkidle')

    await page.fill('input[placeholder="请输入用户名"]', ADMIN_USERNAME)
    await page.fill('input[type="password"]', ADMIN_PASSWORD)
    await page.getByRole('button', { name: '登录', exact: true }).click()

    await page.waitForURL('**/profile', { timeout: 10000 })

    const accessToken = await page.evaluate(() => localStorage.getItem('access_token'))
    const refreshToken = await page.evaluate(() => localStorage.getItem('refresh_token'))
    const userStr = await page.evaluate(() => localStorage.getItem('user'))

    expect(accessToken).toBeTruthy()
    expect(refreshToken).toBeTruthy()
    expect(userStr).toBeTruthy()

    const user = JSON.parse(userStr!)
    expect(user.username).toBe(ADMIN_USERNAME)

    await page.reload()
    await page.waitForURL('**/profile', { timeout: 10000 })

    const accessTokenAfterRefresh = await page.evaluate(() => localStorage.getItem('access_token'))
    const refreshTokenAfterRefresh = await page.evaluate(() =>
      localStorage.getItem('refresh_token')
    )
    const userStrAfterRefresh = await page.evaluate(() => localStorage.getItem('user'))

    expect(accessTokenAfterRefresh).toBe(accessToken)
    expect(refreshTokenAfterRefresh).toBe(refreshToken)
    expect(userStrAfterRefresh).toBe(userStr)
  })

  test('should not require re-login after page refresh', async ({ page }) => {
    await page.goto('/login')
    await page.waitForLoadState('networkidle')

    await page.fill('input[placeholder="请输入用户名"]', ADMIN_USERNAME)
    await page.fill('input[type="password"]', ADMIN_PASSWORD)
    await page.getByRole('button', { name: '登录', exact: true }).click()

    await page.waitForURL('**/profile', { timeout: 10000 })

    await page.reload()
    await page.waitForLoadState('networkidle')

    const currentUrl = page.url()
    expect(currentUrl).toMatch(/profile/)

    const isLoginPage = currentUrl.includes('/login')
    expect(isLoginPage).toBe(false)
  })

  test('should maintain user data consistency after refresh', async ({ page }) => {
    await page.goto('/login')
    await page.waitForLoadState('networkidle')

    await page.fill('input[placeholder="请输入用户名"]', ADMIN_USERNAME)
    await page.fill('input[type="password"]', ADMIN_PASSWORD)
    await page.getByRole('button', { name: '登录', exact: true }).click()

    await page.waitForURL('**/profile', { timeout: 10000 })

    await page.goto('/profile')
    await page.waitForLoadState('networkidle')

    const userData = await page.evaluate(() => {
      const userStr = localStorage.getItem('user')
      return userStr ? JSON.parse(userStr) : null
    })

    expect(userData).toBeTruthy()
    expect(userData.username).toBe(ADMIN_USERNAME)

    await page.reload()
    await page.goto('/profile')
    await page.waitForLoadState('networkidle')

    const userDataAfterRefresh = await page.evaluate(() => {
      const userStr = localStorage.getItem('user')
      return userStr ? JSON.parse(userStr) : null
    })

    expect(userDataAfterRefresh).toEqual(userData)
  })

  test('should clear session on logout', async ({ page }) => {
    await page.goto('/login')
    await page.waitForLoadState('networkidle')

    await page.fill('input[placeholder="请输入用户名"]', ADMIN_USERNAME)
    await page.fill('input[type="password"]', ADMIN_PASSWORD)
    await page.getByRole('button', { name: '登录', exact: true }).click()

    await page.waitForURL('**/profile', { timeout: 10000 })

    const accessTokenBeforeLogout = await page.evaluate(() => localStorage.getItem('access_token'))
    expect(accessTokenBeforeLogout).toBeTruthy()

    await page.goto('/profile')
    await page.waitForLoadState('networkidle')

    const logoutButton = page.locator('button:has-text("退出登录")').first()
    if (await logoutButton.isVisible()) {
      await logoutButton.click()
      await page.waitForURL('**/login', { timeout: 5000 })
    }

    const accessTokenAfterLogout = await page.evaluate(() => localStorage.getItem('access_token'))
    const refreshTokenAfterLogout = await page.evaluate(() => localStorage.getItem('refresh_token'))
    const userAfterLogout = await page.evaluate(() => localStorage.getItem('user'))

    expect(accessTokenAfterLogout).toBeNull()
    expect(refreshTokenAfterLogout).toBeNull()
    expect(userAfterLogout).toBeNull()
  })
})
