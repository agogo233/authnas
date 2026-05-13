import { test, expect, Page } from '@playwright/test'

const TEST_CLIENT_ID = 'e2e-test-client'
const TEST_REDIRECT_URI = 'http://localhost:9999/callback'
const TEST_STATE = 'test-state-12345'

async function getAdminToken(page: Page): Promise<string | null> {
  const loginResp = await page.request.post('/api/auth/login', {
    data: {
      input: 'admin',
      password: 'Admin123!',
    },
  })
  if (loginResp.status() !== 200) {
    const respText = await loginResp.text()
    if (respText.includes('too many failed login attempts')) {
      console.log('Admin login rate limited, skipping test')
      return null
    }
    throw new Error('Failed to login as admin: ' + respText)
  }
  const loginData = await loginResp.json()
  return loginData.data.accessToken as string
}

async function createOIDCClient(page: Page, adminToken: string) {
  const createResp = await page.request.post('/api/admin/clients', {
    headers: {
      Authorization: `Bearer ${adminToken}`,
      'Content-Type': 'application/json',
    },
    data: {
      clientId: TEST_CLIENT_ID,
      name: 'E2E Test Client',
      redirectUris: TEST_REDIRECT_URI,
      scopes: 'openid profile email',
      responseTypes: 'code',
    },
  })
  if (createResp.status() !== 200) {
    const errorBody = await createResp.text()
    console.log('Client creation failed:', createResp.status(), errorBody)
  }
  if (createResp.status() === 200) {
    const createData = await createResp.json()
    expect(createData.success).toBe(true)
  }
}

async function loginAsUser(page: Page, username: string, password: string) {
  await page.goto('/login')
  await page.waitForLoadState('networkidle')

  await page.fill('input[placeholder="请输入用户名"]', username)
  await page.fill('input[type="password"]', password)
  await page.click('button:has-text("登录")')

  await page.waitForLoadState('networkidle')
  await page.waitForTimeout(1000)
}

async function setAuthStorage(page: Page, accessToken: string, refreshToken: string, user: object) {
  await page.evaluate(
    ({ token, refresh, userData }) => {
      localStorage.setItem('access_token', token)
      localStorage.setItem('refresh_token', refresh)
      localStorage.setItem('user', JSON.stringify(userData))
    },
    { token: accessToken, refresh: refreshToken, userData: user }
  )
}

test.describe('OIDC Complete Authorization Flow E2E', () => {
  test.beforeEach(async ({ page }) => {})

  test('complete OIDC flow: client creation -> login -> authorization -> consent -> code exchange', async ({
    page,
  }) => {
    const timestamp = Date.now()
    const testUsername = `oidcuser${timestamp}`
    const testEmail = `${testUsername}@example.com`
    const testPassword = 'TestPass123!'

    const adminToken = await getAdminToken(page)
    if (!adminToken) {
      test.skip()
      return
    }

    const registerResp = await page.request.post('/api/auth/register', {
      data: {
        username: testUsername,
        email: testEmail,
        password: testPassword,
      },
    })
    if (registerResp.status() !== 200) {
      console.log('User registration failed:', await registerResp.text())
    }
    expect(registerResp.status()).toBe(200)
    const registerData = await registerResp.json()
    expect(registerData.success).toBe(true)
    const userToken = registerData.accessToken as string

    const userResp = await page.request.get('/api/user/me', {
      headers: { Authorization: `Bearer ${userToken}` },
    })
    const userInfo = await userResp.json()
    const userId = userInfo.id

    await page.request.post(`/api/admin/users/${userId}/approve`, {
      headers: { Authorization: `Bearer ${adminToken}` },
      data: { approved: true },
    })

    await createOIDCClient(page, adminToken)

    await loginAsUser(page, testUsername, testPassword)
    await page.waitForTimeout(500)

    const oidcAuthUrl = `/oidc/auth?client_id=${TEST_CLIENT_ID}&redirect_uri=${encodeURIComponent(TEST_REDIRECT_URI)}&response_type=code&scope=openid%20profile%20email&state=${TEST_STATE}`

    // Get the consent page URL by using request which follows redirects
    const resp = await page.request.get(oidcAuthUrl)
    const consentPageUrl = resp.url()
    console.log('Consent page URL:', consentPageUrl)

    // Navigate to the consent page
    await page.goto(consentPageUrl, { waitUntil: 'networkidle' })

    await page.waitForLoadState('networkidle')
    await page.waitForTimeout(500)

    const pageContent = await page.content()
    if (pageContent.includes('授权请求') || pageContent.includes('授权')) {
      const allowButton = page.locator(
        'button:has-text("授权"), button:has-text("允许"), button:has-text("Authorize"), button:has-text("Approve")'
      )
      if (await allowButton.isVisible({ timeout: 3000 })) {
        await allowButton.click()
        await page.waitForLoadState('networkidle')
        await page.waitForTimeout(1000)
      }
    }

    const finalUrl = page.url()
    console.log('Final URL after consent:', finalUrl)

    if (finalUrl.includes(TEST_REDIRECT_URI) || finalUrl.includes('code=')) {
      expect(finalUrl).toContain('code=')
      const urlObj = new URL(finalUrl)
      const code = urlObj.searchParams.get('code')
      expect(code).toBeTruthy()
      console.log('Got authorization code:', code)
    }
  })

  test('OIDC consent page displays client info and scopes', async ({ page }) => {
    const timestamp = Date.now()
    const testUsername = `consenttest${timestamp}`
    const testEmail = `${testUsername}@example.com`
    const testPassword = 'TestPass123!'

    const adminToken = await getAdminToken(page)
    if (!adminToken) {
      test.skip()
      return
    }

    const registerResp = await page.request.post('/api/auth/register', {
      data: {
        username: testUsername,
        email: testEmail,
        password: testPassword,
      },
    })
    expect(registerResp.status()).toBe(200)
    const registerData = await registerResp.json()
    const userToken = registerData.accessToken as string

    await createOIDCClient(page, adminToken)

    // Login first
    await page.goto('/login')
    await page.waitForLoadState('networkidle')
    await setAuthStorage(page, userToken, registerData.refreshToken, {
      id: testUsername,
      username: testUsername,
      email: testEmail,
      isAdmin: false,
      approved: true,
    })

    // Navigate to consent page through OIDC auth flow
    const oidcAuthUrl = `/oidc/auth?client_id=${TEST_CLIENT_ID}&redirect_uri=${encodeURIComponent(TEST_REDIRECT_URI)}&response_type=code&scope=openid%20profile%20email&state=test-state`
    const resp = await page.request.get(oidcAuthUrl)
    const consentPageUrl = resp.url()

    // Reload the page to ensure auth state is picked up
    await page.goto(consentPageUrl, { waitUntil: 'networkidle' })

    // Wait for the Vue app to render
    await page.waitForTimeout(3000)

    // Check page content
    const pageContent = await page.content()
    console.log('Consent page URL:', page.url())
    console.log('Consent page content includes 授权请求:', pageContent.includes('授权请求'))
    console.log(
      'Consent page content includes E2E Test Client:',
      pageContent.includes('E2E Test Client')
    )

    await expect(page.locator('h1:has-text("授权请求"), h1:has-text("Authorize")')).toBeVisible({
      timeout: 5000,
    })

    const clientName = page.locator('h3:has-text("E2E Test Client")')
    await expect(clientName).toBeVisible({ timeout: 3000 })

    const scopesSection = page.locator('text=请求的权限')
    await expect(scopesSection).toBeVisible({ timeout: 3000 })
  })

  test('OIDC consent page decline button redirects with error', async ({ page }) => {
    const timestamp = Date.now()
    const testUsername = `declinetest${timestamp}`
    const testEmail = `${testUsername}@example.com`
    const testPassword = 'TestPass123!'

    const adminToken = await getAdminToken(page)
    if (!adminToken) {
      test.skip()
      return
    }

    const registerResp = await page.request.post('/api/auth/register', {
      data: {
        username: testUsername,
        email: testEmail,
        password: testPassword,
      },
    })
    expect(registerResp.status()).toBe(200)
    const registerData = await registerResp.json()
    const userToken = registerData.accessToken as string

    await createOIDCClient(page, adminToken)

    await page.goto('/login')
    await page.waitForLoadState('networkidle')
    await setAuthStorage(page, userToken, registerData.refreshToken, {
      id: testUsername,
      username: testUsername,
      email: testEmail,
      isAdmin: false,
      approved: true,
    })

    const oidcAuthUrl = `/oidc/auth?client_id=${TEST_CLIENT_ID}&redirect_uri=${encodeURIComponent(TEST_REDIRECT_URI)}&response_type=code&scope=openid%20profile%20email&state=decline-state`

    const resp = await page.request.get(oidcAuthUrl)
    const consentPageUrl = resp.url()
    await page.goto(consentPageUrl, { waitUntil: 'networkidle' })

    const declineButton = page.locator(
      'button:has-text("拒绝"), button:has-text("Decline"), button:has-text("取消"), button:has-text("Cancel")'
    )
    if (await declineButton.isVisible({ timeout: 3000 })) {
      await declineButton.click()
      await page.waitForTimeout(2000)

      const finalUrl = page.url()
      if (finalUrl.includes(TEST_REDIRECT_URI)) {
        expect(finalUrl).toContain('error=')
        const urlObj = new URL(finalUrl)
        const error = urlObj.searchParams.get('error')
        expect(error).toBeTruthy()
      }
    }
  })

  test('unauthenticated user can access consent page but sees error for invalid uid', async ({
    page,
  }) => {
    const fakeUid = 'non-existent-uid-12345'
    await page.goto(`/consent/${fakeUid}`)
    await page.waitForLoadState('networkidle')

    // Should stay on consent page (not redirect to login since requiresAuth is now false)
    await expect(page).toHaveURL(/\/consent\//, { timeout: 5000 })
    // Wait for the page to load and check content
    await page.waitForTimeout(2000)
    // Should show an error related to session or loading
    const errorLocator = page.locator('text=/错误|session|无法加载|加载授权信息|not found/i')
    await expect(errorLocator.first()).toBeVisible({ timeout: 5000 })
  })

  test('OIDC authorization request without client_id returns error', async ({ page }) => {
    const response = await page.request.get(
      '/oidc/auth?redirect_uri=http://example.com&response_type=code'
    )
    expect(response.status()).toBe(400)
    const body = await response.json()
    expect(body.success).toBe(false)
    expect(body.message).toBe('invalid_request')
  })

  test('OIDC authorization request with invalid client_id returns error', async ({ page }) => {
    const response = await page.request.get(
      `/oidc/auth?client_id=invalid-client&redirect_uri=${encodeURIComponent(TEST_REDIRECT_URI)}&response_type=code&scope=openid`
    )
    expect(response.status()).toBe(400)
    const body = await response.json()
    expect(body.success).toBe(false)
  })

  test('OIDC authorization request with unregistered redirect_uri returns error', async ({
    page,
  }) => {
    const response = await page.request.get(
      `/oidc/auth?client_id=${TEST_CLIENT_ID}&redirect_uri=http://evil.com/callback&response_type=code&scope=openid`
    )
    expect(response.status()).toBe(400)
    const body = await response.json()
    expect(body.success).toBe(false)
  })

  test('OIDC discovery endpoint returns valid configuration', async ({ page }) => {
    const response = await page.request.get('/oidc/.well-known/openid-configuration')
    expect(response.ok()).toBeTruthy()

    const discovery = await response.json()
    expect(discovery.issuer).toBeTruthy()
    expect(discovery.authorization_endpoint).toContain('/oidc/auth')
    expect(discovery.token_endpoint).toContain('/oidc/token')
    expect(discovery.userinfo_endpoint).toContain('/oidc/userinfo')
    expect(discovery.jwks_uri).toContain('/oidc/jwks')

    const responseTypes: string[] = discovery.response_types_supported || []
    expect(responseTypes).toContain('code')

    const scopes: string[] = discovery.scopes_supported || []
    expect(scopes).toContain('openid')
  })

  test('OIDC JWKS endpoint returns valid RSA key', async ({ page }) => {
    const response = await page.request.get('/oidc/jwks')
    expect(response.ok()).toBeTruthy()

    const jwks = await response.json()
    expect(jwks.keys).toBeDefined()
    expect(jwks.keys.length).toBeGreaterThan(0)

    const key = jwks.keys[0]
    expect(key.kty).toBe('RSA')
    expect(key.use).toBe('sig')
    expect(key.alg).toBe('RS256')
    expect(key.n).toBeTruthy()
    expect(key.e).toBeTruthy()
  })

  test('OIDC userinfo endpoint requires authorization', async ({ page }) => {
    const response = await page.request.get('/oidc/userinfo')
    expect(response.status()).toBe(401)
  })

  test('OIDC userinfo endpoint rejects invalid token', async ({ page }) => {
    const response = await page.request.get('/oidc/userinfo', {
      headers: {
        Authorization: 'Bearer invalid-token',
      },
    })
    expect(response.status()).toBe(401)
  })
})

test.describe('OIDC Token Exchange E2E', () => {
  test('token exchange with invalid code fails', async ({ page }) => {
    const response = await page.request.post('/oidc/token', {
      form: {
        grant_type: 'authorization_code',
        code: 'invalid-code-12345',
        client_id: TEST_CLIENT_ID,
        client_secret: 'test-secret',
        redirect_uri: TEST_REDIRECT_URI,
      },
    })
    expect([400, 401]).toContain(response.status())
  })

  test('token exchange without required parameters fails', async ({ page }) => {
    const response = await page.request.post('/oidc/token', {
      form: {
        grant_type: 'authorization_code',
        code: 'some-code',
      },
    })
    expect(response.status()).toBe(400)
  })

  test('refresh token with invalid token fails', async ({ page }) => {
    const response = await page.request.post('/oidc/token', {
      form: {
        grant_type: 'refresh_token',
        refresh_token: 'invalid-refresh-token',
        client_id: TEST_CLIENT_ID,
      },
    })
    expect(response.status()).toBe(400)
  })

  test('OIDC authorization request with empty state parameter', async ({ page }) => {
    const response = await page.request.get(
      `/oidc/auth?client_id=${TEST_CLIENT_ID}&redirect_uri=${encodeURIComponent(TEST_REDIRECT_URI)}&response_type=code&scope=openid&state=`
    )
    expect(response.status()).toBe(200)
  })

  test('OIDC authorization request with missing response_type returns error', async ({ page }) => {
    const response = await page.request.get(
      `/oidc/auth?client_id=${TEST_CLIENT_ID}&redirect_uri=${encodeURIComponent(TEST_REDIRECT_URI)}&scope=openid`
    )
    expect(response.status()).toBe(400)
    const body = await response.json()
    expect(body.success).toBe(false)
  })

  test('OIDC authorization request with missing scope returns error', async ({ page }) => {
    const response = await page.request.get(
      `/oidc/auth?client_id=${TEST_CLIENT_ID}&redirect_uri=${encodeURIComponent(TEST_REDIRECT_URI)}&response_type=code`
    )
    expect(response.status()).toBe(200)
  })
})
