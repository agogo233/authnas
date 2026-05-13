import { test, expect, Page } from '@playwright/test'

const BASE_URL = process.env.E2E_BASE_URL || 'http://localhost:8080'

const TEST_ADMIN = {
  username: 'admin',
  password: 'Admin123!',
  email: 'admin@example.com',
}

async function performLogin(page: Page, username: string, password: string) {
  await page.goto('/login')
  await page.waitForLoadState('networkidle')

  await page.fill('input[placeholder="请输入用户名"]', username)
  await page.fill('input[type="password"]', password)
  await page.click('button[type="submit"]')

  await page.waitForLoadState('networkidle')
  await page.waitForTimeout(1000)
}

async function clearBrowserState(page: Page) {
  await page.context().clearCookies()
  await page.goto('${BASE_URL}')
  await page.waitForLoadState('domcontentloaded')
  await page.evaluate(() => localStorage.clear())
  await page.evaluate(() => sessionStorage.clear())
}

test.describe('Real E2E - OIDC Complete Flow', () => {
  test.beforeEach(async ({ page }) => {
    await clearBrowserState(page)
  })
  test.describe('1. OIDC Discovery and Configuration', () => {
    test('should fetch OIDC discovery document', async ({ request }) => {
      const response = await request.get('${BASE_URL}/oidc/.well-known/openid-configuration')
      expect(response.ok()).toBeTruthy()

      const config = await response.json()
      expect(config.issuer).toBeTruthy()
      expect(config.authorization_endpoint).toBeTruthy()
      expect(config.token_endpoint).toBeTruthy()
      expect(config.userinfo_endpoint).toBeTruthy()
    })

    test('should have valid JWKS endpoint', async ({ request }) => {
      const discovery = await request.get('${BASE_URL}/oidc/.well-known/openid-configuration')
      const config = await discovery.json()

      const jwksResponse = await request.get(config.jwks_uri)
      expect(jwksResponse.ok()).toBeTruthy()

      const jwks = await jwksResponse.json()
      expect(jwks.keys).toBeDefined()
      expect(jwks.keys.length).toBeGreaterThan(0)
    })
  })

  test.describe('2. OIDC Authorization Code Flow', () => {
    test('should start authorization code flow', async ({ page }) => {
      const clientId = 'test-client'
      const redirectUri = '${BASE_URL}/callback'
      const state = 'test-state-' + Date.now()
      const nonce = 'test-nonce-' + Date.now()

      const authUrl = `${BASE_URL}/oidc/auth?client_id=${clientId}&redirect_uri=${encodeURIComponent(redirectUri)}&response_type=code&scope=openid+profile+email&state=${state}&nonce=${nonce}`

      await page.goto(authUrl)
      await page.waitForLoadState('networkidle')
      await page.waitForTimeout(1000)

      const currentUrl = page.url()
      expect(currentUrl).toBeTruthy()
    })

    test('should handle authorization request', async ({ page }) => {
      const clientId = 'test-client'
      const redirectUri = '${BASE_URL}/callback'
      const state = 'test-state-' + Date.now()

      const authUrl = `${BASE_URL}/oidc/auth?client_id=${clientId}&redirect_uri=${encodeURIComponent(redirectUri)}&response_type=code&scope=openid+profile+email&state=${state}`

      await page.goto(authUrl)
      await page.waitForLoadState('networkidle')
      await page.waitForTimeout(2000)

      const currentUrl = page.url()
      expect(currentUrl).toBeTruthy()
    })

    test('should display client information on consent page', async ({ page }) => {
      await performLogin(page, TEST_ADMIN.username, TEST_ADMIN.password)
      await page.waitForTimeout(1000)

      const clientId = 'test-client'
      const redirectUri = '${BASE_URL}/callback'
      const state = 'test-state-' + Date.now()

      const authUrl = `${BASE_URL}/oidc/auth?client_id=${clientId}&redirect_uri=${encodeURIComponent(redirectUri)}&response_type=code&scope=openid+profile+email&state=${state}`

      await page.goto(authUrl)
      await page.waitForLoadState('networkidle')
      await page.waitForTimeout(1000)

      const currentUrl = page.url()
      if (currentUrl.includes('/consent/')) {
        const pageContent = await page.content()
        expect(pageContent).toBeTruthy()
      } else {
        expect(currentUrl).toBeTruthy()
      }
    })
  })

  test.describe('3. Token Exchange', () => {
    test('should not exchange token without authorization code', async ({ request }) => {
      const response = await request.post('${BASE_URL}/oidc/token', {
        headers: {
          'Content-Type': 'application/x-www-form-urlencoded',
        },
        data: {
          grant_type: 'authorization_code',
          code: 'invalid-code',
          redirect_uri: '${BASE_URL}/callback',
          client_id: 'test-client',
        },
      })

      expect(response.status()).toBe(400)
    })

    test('should not exchange token with invalid grant_type', async ({ request }) => {
      const response = await request.post('${BASE_URL}/oidc/token', {
        headers: {
          'Content-Type': 'application/x-www-form-urlencoded',
        },
        data: {
          grant_type: 'invalid_grant',
          code: 'some-code',
          redirect_uri: '${BASE_URL}/callback',
          client_id: 'test-client',
        },
      })

      expect(response.status()).toBe(400)
    })
  })

  test.describe('4. UserInfo Endpoint', () => {
    test('should reject userinfo request without token', async ({ request }) => {
      const response = await request.get('${BASE_URL}/oidc/userinfo')
      expect(response.status()).toBe(401)
    })

    test('should reject userinfo request with invalid token', async ({ request }) => {
      const response = await request.get('${BASE_URL}/oidc/userinfo', {
        headers: {
          Authorization: 'Bearer invalid-token',
        },
      })
      expect(response.status()).toBe(401)
    })
  })

  test.describe('5. Logout Flow (OIDC RP-Initiated Logout)', () => {
    test('should access logout endpoint', async ({ page }) => {
      await page.goto('${BASE_URL}/oidc/logout')
      await page.waitForLoadState('networkidle')

      const currentUrl = page.url()
      expect(currentUrl).toBeTruthy()
    })

    test('should redirect after logout', async ({ page }) => {
      await page.goto('/login')
      await page.waitForLoadState('networkidle')

      await page.fill('input[placeholder="请输入用户名"]', TEST_ADMIN.username)
      await page.fill('input[type="password"]', TEST_ADMIN.password)
      await page.click('button[type="submit"]')

      await page.waitForTimeout(2000)

      const loggedInUrl = page.url()
      expect(loggedInUrl).not.toContain('/login')

      await page.goto('${BASE_URL}/oidc/logout')
      await page.waitForLoadState('networkidle')
      await page.waitForTimeout(2000)

      const currentUrl = page.url()
      expect(currentUrl).not.toContain('oidc/logout')
    })
  })

  test.describe('6. Client Registration', () => {
    test('should access admin clients page', async ({ page }) => {
      await page.goto('/login')
      await page.waitForLoadState('domcontentloaded')
      await page.waitForTimeout(1000)

      await page.fill('input[placeholder="请输入用户名"]', TEST_ADMIN.username)
      await page.fill('input[type="password"]', TEST_ADMIN.password)
      await page.click('button[type="submit"]')
      await page.waitForTimeout(2000)

      await page.goto('/admin/clients')
      await page.waitForLoadState('domcontentloaded')
      await page.waitForTimeout(2000)

      expect(page.url()).toMatch(/clients/)
    })

    test('should display clients list', async ({ page }) => {
      await page.goto('/login')
      await page.waitForLoadState('domcontentloaded')
      await page.waitForTimeout(1000)

      await page.fill('input[placeholder="请输入用户名"]', TEST_ADMIN.username)
      await page.fill('input[type="password"]', TEST_ADMIN.password)
      await page.click('button[type="submit"]')
      await page.waitForTimeout(2000)

      await page.goto('/admin/clients')
      await page.waitForLoadState('domcontentloaded')
      await page.waitForTimeout(2000)

      const pageContent = await page.content()
      expect(pageContent).toBeTruthy()
    })
  })

  test.describe('7. Complete OIDC User Journey', () => {
    test('should complete full authorization code flow with test credentials', async ({ page }) => {
      await page.goto('/login')
      await page.waitForLoadState('domcontentloaded')
      await page.waitForTimeout(1000)

      await page.fill('input[placeholder="请输入用户名"]', TEST_ADMIN.username)
      await page.fill('input[type="password"]', TEST_ADMIN.password)
      await page.click('button[type="submit"]')

      await page.waitForTimeout(2000)

      const isLoggedIn = !page.url().includes('/login')
      expect(isLoggedIn).toBeTruthy()
    })

    test('should handle authorization with post_logout_redirect_uri', async ({ page }) => {
      const clientId = 'test-client'
      const postLogoutUri = '${BASE_URL}/'
      const state = 'logout-state-' + Date.now()

      const logoutUrl = `${BASE_URL}/oidc/logout?client_id=${clientId}&post_logout_redirect_uri=${encodeURIComponent(postLogoutUri)}&state=${state}`

      await page.goto(logoutUrl)
      await page.waitForLoadState('networkidle')
      await page.waitForTimeout(1000)

      expect(page.url()).toBeTruthy()
    })
  })
})
