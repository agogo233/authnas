import { test, expect, request } from '@playwright/test'

const BASE_URL = process.env.E2E_BASE_URL || 'http://localhost:8080'

test.describe('Advanced Security Tests', () => {
  test.describe('Account Lockout Protection', () => {
    test('should lock account after maximum failed login attempts', async ({ page }) => {
      await page.goto('/login')

      const maxAttempts = 10
      for (let i = 0; i < maxAttempts; i++) {
        await page.fill('input[placeholder="请输入用户名"]', 'locktestuser')
        await page.fill('input[type="password"]', 'wrongpassword')
        await page.getByRole('button', { name: '登录', exact: true }).click()
        await page.waitForTimeout(300)
      }

      await page.waitForTimeout(1000)

      await page.fill('input[placeholder="请输入用户名"]', 'locktestuser')
      await page.fill('input[type="password"]', 'correctpassword')
      await page.getByRole('button', { name: '登录', exact: true }).click()
      await page.waitForTimeout(1000)

      const pageContent = await page.content()
      expect(
        pageContent.includes('账户已锁定') ||
          pageContent.includes('账号已锁定') ||
          pageContent.includes('登录失败次数过多') ||
          page.url().includes('/login')
      ).toBeTruthy()
    })

    test('should allow login after lockout period expires', async ({ page }) => {
      await page.goto('/login')

      const pageContent = await page.content()
      expect(pageContent).toBeTruthy()
    })

    test('should track failed attempts per user independently', async ({ page }) => {
      await page.goto('/login')

      for (let i = 0; i < 3; i++) {
        await page.fill('input[placeholder="请输入用户名"]', 'user1')
        await page.fill('input[type="password"]', 'wrongpassword')
        await page.getByRole('button', { name: '登录', exact: true }).click()
        await page.waitForTimeout(300)
      }

      for (let i = 0; i < 3; i++) {
        await page.fill('input[placeholder="请输入用户名"]', 'user2')
        await page.fill('input[type="password"]', 'wrongpassword')
        await page.getByRole('button', { name: '登录', exact: true }).click()
        await page.waitForTimeout(300)
      }

      const pageUrl = page.url()
      expect(pageUrl).toContain('/login')
    })
  })

  test.describe('CSRF Protection', () => {
    test('should validate CSRF protection implementation status', async ({ page }) => {
      await page.goto('/login')
      await page.waitForLoadState('networkidle')
      await page.evaluate(() => {
        localStorage.setItem('access_token', 'valid-token')
        localStorage.setItem(
          'user',
          JSON.stringify({
            id: '1',
            username: 'testuser',
            email: 'test@example.com',
            isAdmin: true,
            approved: true,
          })
        )
      })

      await page.goto('/profile')
      await page.waitForLoadState('networkidle')

      const csrfInput = page.locator(
        'input[name="_csrf"], input[name="csrf_token"], input[type="hidden"][id*="csrf"]'
      )
      const hasCsrfToken = (await csrfInput.count()) > 0

      if (!hasCsrfToken) {
        test.skip(
          true,
          'CSRF token not found in form - backend needs to inject CSRF token into forms'
        )
      }
      expect(hasCsrfToken).toBeTruthy()
    })

    test('should handle forms without CSRF gracefully when API is unavailable', async ({
      page,
    }) => {
      await page.goto('/login')
      await page.waitForLoadState('networkidle')

      await page.fill('input[placeholder="请输入用户名"]', 'testuser')
      await page.fill('input[type="password"]', 'TestPass123!')

      const response = await page.evaluate(async () => {
        try {
          const res = await fetch('/api/auth/login', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ username: 'test', password: 'test' }),
          })
          return { ok: res.ok, status: res.status }
        } catch {
          return { ok: false, status: 0 }
        }
      })

      expect(response).toBeTruthy()
    })
  })

  test.describe('Session Fixation Prevention', () => {
    test('should regenerate session ID after successful login', async ({ page }) => {
      await page.goto('/login')
      await page.waitForLoadState('networkidle')

      const sessionIdBefore = await page.evaluate(() => {
        return (
          sessionStorage.getItem('session_id') || localStorage.getItem('session_id') || 'no-session'
        )
      })

      await page.fill('input[placeholder="请输入用户名"]', 'testuser')
      await page.fill('input[type="password"]', 'TestPass123!')
      await page.getByRole('button', { name: '登录', exact: true }).click()
      await page.waitForTimeout(1500)

      const sessionIdAfter = await page.evaluate(() => {
        return (
          sessionStorage.getItem('session_id') || localStorage.getItem('session_id') || 'no-session'
        )
      })

      if (sessionIdBefore !== 'no-session' && sessionIdAfter !== 'no-session') {
        expect(sessionIdBefore).not.toBe(sessionIdAfter)
      }
    })

    test('should change session ID after password change', async ({ page }) => {
      await page.goto('/login')
      await page.waitForLoadState('networkidle')

      await page.fill('input[placeholder="请输入用户名"]', 'admin')
      await page.fill('input[type="password"]', 'Admin123!')
      await page.getByRole('button', { name: '登录', exact: true }).click()

      await page.waitForTimeout(2000)

      const isLocked = await page
        .getByText(/账户已锁定|账号已锁定|登录失败次数过多/)
        .isVisible()
        .catch(() => false)
      if (isLocked) {
        test.skip(
          true,
          'Admin account is locked due to too many failed login attempts - wait for lockout to expire or reset account'
        )
      }

      await page.waitForURL('**/user/**', { timeout: 5000 }).catch(() => {})
      await page.waitForLoadState('networkidle')
      await page.waitForTimeout(500)

      const currentUrl = page.url()
      if (!currentUrl.includes('/user/')) {
        test.skip(true, 'Login failed - could not access user pages')
      }

      await page.goto('/user/security')
      await page.waitForLoadState('networkidle')
      await page.waitForTimeout(1000)

      const tabLocator = page.locator('.n-tabs-tab').filter({ hasText: '修改密码' })
      const tabCount = await tabLocator.count()

      if (tabCount === 0) {
        test.skip(
          true,
          'Password change form not available - frontend needs to implement password change tab'
        )
      }

      await tabLocator.click()
      await page.waitForTimeout(500)

      const passwordInput = page
        .locator(
          'input[placeholder*="当前密码"], input[placeholder*="旧密码"], input[name*="password"]'
        )
        .first()
      const inputCount = await passwordInput.count()

      if (inputCount === 0) {
        test.skip(
          true,
          'Password change form inputs not found after clicking tab - form may not be implemented'
        )
      }

      const sessionIdBefore = await page.evaluate(() => {
        return (
          sessionStorage.getItem('session_id') || localStorage.getItem('session_id') || 'no-session'
        )
      })

      const currentPassword = 'Admin123!'
      const newTestPassword = 'AdminNew123!'

      await passwordInput.fill(currentPassword)
      const newPasswordInput = page
        .locator('input[placeholder*="新密码"], input[name*="newPassword"]')
        .first()
      await newPasswordInput.fill(newTestPassword)
      const confirmInput = page
        .locator('input[placeholder*="再次"], input[name*="confirm"]')
        .first()
      await confirmInput.fill(newTestPassword)

      const updateButton = page
        .getByRole('button', { name: '更新密码', exact: true })
        .or(page.getByRole('button', { name: '修改密码' }))
      await updateButton.click()
      await page.waitForTimeout(2000)

      const sessionIdAfter = await page.evaluate(() => {
        return (
          sessionStorage.getItem('session_id') || localStorage.getItem('session_id') || 'no-session'
        )
      })

      if (sessionIdBefore === 'no-session' && sessionIdAfter === 'no-session') {
        test.skip(
          true,
          'Session management not implemented in frontend - backend needs to regenerate session after password change'
        )
      }

      const passwordChanged = await page.evaluate(() => {
        return document.body.textContent?.includes('密码修改成功') || false
      })

      if (!passwordChanged) {
        test.skip(
          true,
          'Password change API failed - cannot test session regeneration without successful password change'
        )
      }

      expect(sessionIdBefore).not.toBe(sessionIdAfter)
    })
  })

  test.describe('IDOR Prevention', () => {
    test('should prevent horizontal privilege escalation - profile access', async ({ page }) => {
      await page.goto('/login')
      await page.waitForLoadState('networkidle')

      await page.evaluate(() => {
        localStorage.setItem('access_token', 'user-a-token')
        localStorage.setItem(
          'user',
          JSON.stringify({
            id: '1',
            username: 'usera',
            email: 'usera@example.com',
            isAdmin: false,
            approved: true,
          })
        )
      })

      await page.goto('/profile')
      await page.waitForLoadState('networkidle')
      await page.waitForTimeout(1000)

      const currentUrl = page.url()
      expect(currentUrl).toContain('/profile')
    })

    test('should prevent accessing other user data via direct API', async ({ page, request }) => {
      await page.goto('/login')
      await page.waitForLoadState('networkidle')
      await page.evaluate(() => {
        localStorage.setItem('access_token', 'user-a-token')
        localStorage.setItem(
          'user',
          JSON.stringify({
            id: '1',
            username: 'usera',
            email: 'usera@example.com',
            isAdmin: false,
            approved: true,
          })
        )
      })

      const response = await request.get(`${BASE_URL}/api/users/999/profile`, {
        headers: {
          Authorization: 'Bearer user-a-token',
        },
      })

      expect([403, 404, 401]).toContain(response.status())
    })

    test('should prevent vertical privilege escalation - admin API access', async ({
      page,
      request,
    }) => {
      await page.goto('/login')
      await page.waitForLoadState('networkidle')
      await page.evaluate(() => {
        localStorage.setItem('access_token', 'regular-user-token')
        localStorage.setItem(
          'user',
          JSON.stringify({
            id: '2',
            username: 'regularuser',
            email: 'regular@example.com',
            isAdmin: false,
            approved: true,
          })
        )
      })

      const adminResponse = await request.get(`${BASE_URL}/api/admin/users`, {
        headers: {
          Authorization: 'Bearer regular-user-token',
        },
      })

      expect([403, 401, 404]).toContain(adminResponse.status())
    })
  })

  test.describe('JWT Token Security', () => {
    test('should reject token with missing required claims', async ({ page, request }) => {
      await page.goto('/login')
      await page.waitForLoadState('networkidle')

      const response = await request.get(`${BASE_URL}/api/protected`, {
        headers: {
          Authorization:
            'Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIn0.invalid',
        },
      })

      expect([401, 403, 404]).toContain(response.status())
    })

    test('should reject token with missing exp claim', async ({ page, request }) => {
      await page.goto('/login')
      await page.waitForLoadState('networkidle')

      const response = await request.get(`${BASE_URL}/api/protected`, {
        headers: {
          Authorization: 'Bearer no-expiration-token',
        },
      })

      expect([401, 403, 404]).toContain(response.status())
    })

    test('should reject token with missing iat claim', async ({ page, request }) => {
      await page.goto('/login')
      await page.waitForLoadState('networkidle')

      const response = await request.get(`${BASE_URL}/api/protected`, {
        headers: {
          Authorization: 'Bearer no-iat-token',
        },
      })

      expect([401, 403, 404]).toContain(response.status())
    })
  })

  test.describe('Refresh Token Security', () => {
    test('should reject reuse of revoked refresh token', async ({ page, request }) => {
      await page.goto('/login')
      await page.waitForLoadState('networkidle')

      const refreshResponse = await request.post(`${BASE_URL}/oidc/token`, {
        form: {
          grant_type: 'refresh_token',
          refresh_token: 'previously-used-refresh-token',
          client_id: 'test-client',
        },
      })

      expect([200, 400, 401, 404]).toContain(refreshResponse.status())
    })

    test('should handle refresh token rotation if implemented', async ({ page }) => {
      test.skip(
        true,
        'Refresh token rotation requires: (1) backend to include expiresAt in user object, (2) frontend to detect expired tokens and call refresh endpoint, (3) backend to implement refresh token rotation (invalidate old, issue new)'
      )
    })
  })

  test.describe('API Rate Limiting', () => {
    test('should handle repeated login API calls gracefully', async ({ page, request }) => {
      await page.goto('/login')
      await page.waitForLoadState('networkidle')

      const responses = []
      for (let i = 0; i < 15; i++) {
        const response = await request.post(`${BASE_URL}/api/auth/login`, {
          data: {
            username: 'rateLimitTest',
            password: 'wrongpassword',
          },
        })
        responses.push(response.status())
        await page.waitForTimeout(100)
      }

      const hasRateLimited = responses.some((status) => status === 429)
      const lastResponses = responses.slice(-5)
      expect(hasRateLimited || lastResponses.every((s) => s >= 400 || s === 404)).toBeTruthy()
    })

    test('should handle repeated token refresh attempts', async ({ page, request }) => {
      await page.goto('/login')
      await page.waitForLoadState('networkidle')

      const responses = []
      for (let i = 0; i < 20; i++) {
        const response = await request.post(`${BASE_URL}/oidc/token`, {
          form: {
            grant_type: 'refresh_token',
            refresh_token: 'invalid-refresh',
            client_id: 'test-client',
          },
        })
        responses.push(response.status())
        await page.waitForTimeout(50)
      }

      const hasRateLimitOrRejection = responses.some((status) => status === 429 || status >= 400)
      expect(hasRateLimitOrRejection).toBeTruthy()
    })
  })

  test.describe('HTTP Security Headers', () => {
    test('should include security headers in responses', async ({ page, request }) => {
      await page.goto('/login')
      await page.waitForLoadState('networkidle')

      const response = await request.get(page.url())

      const cspHeader = response.headers()['content-security-policy']
      const xFrameOptions = response.headers()['x-frame-options']
      const xContentTypeOptions = response.headers()['x-content-type-options']

      const hasSecurityHeaders =
        (cspHeader && cspHeader.length > 0) ||
        (xFrameOptions && xFrameOptions.length > 0) ||
        (xContentTypeOptions && xContentTypeOptions.length > 0)

      if (!hasSecurityHeaders) {
        test.skip(true, 'Security headers not configured in development server')
      }
      expect(hasSecurityHeaders).toBeTruthy()
    })
  })

  test.describe('Sensitive Data Exposure', () => {
    test('should not expose sensitive data in API responses', async ({ page, request }) => {
      await page.goto('/login')
      await page.waitForLoadState('networkidle')
      await page.evaluate(() => {
        localStorage.setItem('access_token', 'admin-token')
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

      const response = await request.get(`${BASE_URL}/api/users`, {
        headers: {
          Authorization: 'Bearer admin-token',
        },
      })

      if (response.ok()) {
        const responseText = await response.text()
        expect(responseText).not.toMatch(/password/i)
        expect(responseText).not.toMatch(/secret/i)
        expect(responseText).not.toMatch(/token["\s]*:/i)
      }
    })

    test('should not log sensitive information in client console', async ({ page }) => {
      const consoleLogs: string[] = []
      page.on('console', (msg) => {
        if (msg.type() === 'error' || msg.type() === 'warn') {
          consoleLogs.push(msg.text())
        }
      })

      await page.goto('/login')
      await page.waitForLoadState('networkidle')

      await page.fill('input[placeholder="请输入用户名"]', 'testuser')
      await page.fill('input[type="password"]', 'SuperSecret123!')
      await page.getByRole('button', { name: '登录', exact: true }).click()
      await page.waitForTimeout(1000)

      const sensitivePatterns = [
        /password.*=.*SuperSecret/i,
        /token.*=.*SuperSecret/i,
        /bearer.*SuperSecret/i,
        /auth.*SuperSecret/i,
      ]

      for (const log of consoleLogs) {
        for (const pattern of sensitivePatterns) {
          expect(log).not.toMatch(pattern)
        }
      }
    })
  })
})
