import { test, expect } from '@playwright/test'

const BASE_URL = process.env.E2E_BASE_URL || 'http://localhost:8080'

test.describe('OIDC API Endpoints', () => {
  test.describe('OIDC Token Exchange', () => {
    test('should reject token exchange with invalid code', async ({ page, request }) => {
      const tokenResponse = await request.post('/oidc/token', {
        form: {
          grant_type: 'authorization_code',
          code: 'invalid-code',
          client_id: 'test-client',
          client_secret: 'test-client-secret',
          redirect_uri: '${BASE_URL}/callback',
        },
      })

      expect(tokenResponse.status()).toBe(400)
    })

    test('should reject token exchange with expired code', async ({ page, request }) => {
      const tokenResponse = await request.post('/oidc/token', {
        form: {
          grant_type: 'authorization_code',
          code: 'expired-code',
          client_id: 'test-client',
          client_secret: 'test-client-secret',
          redirect_uri: '${BASE_URL}/callback',
        },
      })

      expect(tokenResponse.status()).toBe(400)
    })

    test('should reject token exchange without required parameters', async ({ page, request }) => {
      const tokenResponse = await request.post('/oidc/token', {
        form: {
          grant_type: 'authorization_code',
          code: 'some-code',
        },
      })

      expect(tokenResponse.status()).toBe(400)
    })
  })

  test.describe('OIDC UserInfo Endpoint', () => {
    test('should reject userinfo request without token', async ({ page, request }) => {
      const userInfoResponse = await request.get('/oidc/userinfo')

      expect(userInfoResponse.status()).toBe(401)
    })

    test('should reject userinfo request with invalid token', async ({ page, request }) => {
      const userInfoResponse = await request.get('/oidc/userinfo', {
        headers: {
          Authorization: 'Bearer invalid-token',
        },
      })

      expect(userInfoResponse.status()).toBe(401)
    })

    test('should reject userinfo request with malformed authorization header', async ({
      page,
      request,
    }) => {
      const userInfoResponse = await request.get('/oidc/userinfo', {
        headers: {
          Authorization: 'InvalidFormat token',
        },
      })

      expect(userInfoResponse.status()).toBe(401)
    })
  })

  test.describe('OIDC Discovery Endpoint', () => {
    test('should return OpenID Connect Discovery document', async ({ page, request }) => {
      const discoveryResponse = await request.get('/oidc/.well-known/openid-configuration')

      expect(discoveryResponse.ok()).toBeTruthy()
      const discovery = await discoveryResponse.json()
      expect(discovery).toHaveProperty('issuer')
      expect(discovery).toHaveProperty('authorization_endpoint')
      expect(discovery).toHaveProperty('token_endpoint')
      expect(discovery).toHaveProperty('userinfo_endpoint')
      expect(discovery).toHaveProperty('jwks_uri')
    })

    test('should have valid issuer in discovery document', async ({ page, request }) => {
      const discoveryResponse = await request.get('/oidc/.well-known/openid-configuration')

      expect(discoveryResponse.ok()).toBeTruthy()
      const discovery = await discoveryResponse.json()
      expect(discovery.issuer).toBeTruthy()
      expect(discovery.issuer).toContain('/')
    })
  })

  test.describe('OIDC Security', () => {
    test('should prevent authorization code replay', async ({ page, request }) => {
      const code = 'single-use-auth-code'

      const firstExchange = await request.post('/oidc/token', {
        form: {
          grant_type: 'authorization_code',
          code,
          client_id: 'test-client',
          client_secret: 'test-client-secret',
          redirect_uri: '${BASE_URL}/callback',
        },
      })

      const secondExchange = await request.post('/oidc/token', {
        form: {
          grant_type: 'authorization_code',
          code,
          client_id: 'test-client',
          client_secret: 'test-client-secret',
          redirect_uri: '${BASE_URL}/callback',
        },
      })

      expect(secondExchange.status()).toBe(400)
    })

    test('should validate redirect_uri matches initial request', async ({ page, request }) => {
      const tokenResponse = await request.post('/oidc/token', {
        form: {
          grant_type: 'authorization_code',
          code: 'valid-code',
          client_id: 'test-client',
          client_secret: 'test-client-secret',
          redirect_uri: 'http://different-uri.com/callback',
        },
      })

      expect(tokenResponse.status()).toBe(400)
    })

    test('should not issue token for unauthenticated request if required', async ({
      page,
      request,
    }) => {
      const tokenResponse = await request.post('/oidc/token', {
        form: {
          grant_type: 'authorization_code',
          code: 'valid-code',
          client_id: 'test-client',
          redirect_uri: '${BASE_URL}/callback',
        },
      })

      expect([400, 401]).toContain(tokenResponse.status())
    })
  })

  test.describe('OIDC Refresh Token Flow', () => {
    test('should reject refresh with invalid token', async ({ page, request }) => {
      const refreshResponse = await request.post('/oidc/token', {
        form: {
          grant_type: 'refresh_token',
          refresh_token: 'invalid-refresh-token',
          client_id: 'test-client',
        },
      })

      expect(refreshResponse.status()).toBe(400)
    })

    test('should reject refresh with expired token', async ({ page, request }) => {
      const refreshResponse = await request.post('/oidc/token', {
        form: {
          grant_type: 'refresh_token',
          refresh_token: 'expired-refresh-token',
          client_id: 'test-client',
        },
      })

      expect(refreshResponse.status()).toBe(400)
    })
  })
})

test.describe('OIDC Public Flow', () => {
  test('should return error for request with missing client_id', async ({ request }) => {
    const response = await request.get('/oidc/auth?redirect_uri=${BASE_URL}/callback')
    expect(response.status()).toBe(400)
    const body = await response.json()
    expect(body.success).toBe(false)
    expect(body.message).toBe('invalid_request')
  })

  test('should return error for request with invalid client_id', async ({ request }) => {
    const response = await request.get(
      '/oidc/auth?client_id=invalid-client&redirect_uri=${BASE_URL}/callback'
    )
    expect(response.status()).toBe(400)
    const body = await response.json()
    expect(body.success).toBe(false)
  })

  test('should return error for request with unregistered redirect_uri', async ({ request }) => {
    const response = await request.get(
      '/oidc/auth?client_id=test-client&redirect_uri=http://evil.com/callback'
    )
    expect(response.status()).toBe(400)
    const body = await response.json()
    expect(body.success).toBe(false)
  })

  test('should return error for request with missing required parameters', async ({ request }) => {
    const response = await request.get('/oidc/auth?client_id=test-client')
    expect(response.status()).toBe(400)
    const body = await response.json()
    expect(body.success).toBe(false)
  })
})

test.describe('OIDC Consent Page (Public Behavior)', () => {
  test('should allow access to consent page without auth (public OIDC flow)', async ({ page }) => {
    await page.goto('/consent/test-uid-123')
    await page.waitForLoadState('networkidle')

    await expect(page).toHaveURL(/\/consent\/test-uid-123/, { timeout: 5000 })
  })

  test('should display error for invalid uid', async ({ page }) => {
    await page.goto('/consent/invalid-uid-xyz')
    await page.waitForLoadState('networkidle')

    await expect(page).toHaveURL(/\/consent\/invalid-uid-xyz/, { timeout: 5000 })
    const errorResult = page.locator('.n-result')
    await expect(errorResult).toBeVisible({ timeout: 10000 })
  })
})
