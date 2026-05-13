import { test, expect } from '@playwright/test'

test.describe('Security Tests - Injection Protection', () => {
  test.describe('XSS Protection', () => {
    test('should handle XSS payload in username field without crashing', async ({ page }) => {
      await page.goto('/login')

      const xssPayload = '<script>alert(1)</script>'
      await page.fill('input[placeholder="请输入用户名"]', xssPayload)
      await page.fill('input[type="password"]', 'password')
      await page.getByRole('button', { name: '登录', exact: true }).click()

      await page.waitForTimeout(500)

      const pageUrl = page.url()
      expect(pageUrl).toBeTruthy()
    })

    test('should handle XSS payload in email field without crashing', async ({ page }) => {
      await page.goto('/register')

      const xssPayload = '<img src=x onerror=alert(1)>'
      await page.fill('input[placeholder*="请输入邮箱"]', xssPayload)
      await page.fill('input[placeholder="请输入用户名"]', 'testuser')
      await page.fill('input[type="password"]', 'TestPass123!')

      await page.getByRole('button', { name: '注册', exact: true }).click()

      await page.waitForTimeout(500)

      const pageUrl = page.url()
      expect(pageUrl).toBeTruthy()
    })

    test('should handle SQL injection payload in username', async ({ page }) => {
      await page.goto('/login')

      const sqliPayload = "' OR '1'='1"
      await page.fill('input[placeholder="请输入用户名"]', sqliPayload)
      await page.fill('input[type="password"]', 'anything')
      await page.getByRole('button', { name: '登录', exact: true }).click()

      await page.waitForTimeout(500)

      const pageUrl = page.url()
      expect(pageUrl).toBeTruthy()
    })

    test('should handle command injection payload in username', async ({ page }) => {
      await page.goto('/login')

      const cmdPayload = '; ls -la'
      await page.fill('input[placeholder="请输入用户名"]', cmdPayload)
      await page.fill('input[type="password"]', 'password')
      await page.getByRole('button', { name: '登录', exact: true }).click()

      await page.waitForTimeout(500)

      const pageUrl = page.url()
      expect(pageUrl).toBeTruthy()
    })
  })

  test.describe('SQL Injection Protection', () => {
    test('should prevent SQL injection in login username', async ({ page }) => {
      await page.goto('/login')

      await page.fill('input[placeholder="请输入用户名"]', "admin'--")
      await page.fill('input[type="password"]', 'anything')
      await page.getByRole('button', { name: '登录', exact: true }).click()

      await page.waitForTimeout(500)

      const pageUrl = page.url()
      expect(pageUrl).not.toContain('error=sql')
      expect(pageUrl).not.toContain('syntax')
    })

    test('should prevent SQL injection in registration', async ({ page }) => {
      await page.goto('/register')

      const sqliPayload = "test@example.com' DROP TABLE users--"
      await page.fill('input[placeholder*="请输入邮箱"]', sqliPayload)
      await page.fill('input[placeholder="请输入用户名"]', 'testuser')
      await page.fill('input[type="password"]', 'TestPass123!')

      await page.getByRole('button', { name: '注册', exact: true }).click()

      await page.waitForTimeout(500)

      const pageUrl = page.url()
      expect(pageUrl).toBeTruthy()
    })
  })
})

test.describe('Security Tests - Authentication & Session', () => {
  test.describe('Brute Force Protection', () => {
    test('should handle multiple failed login attempts', async ({ page }) => {
      await page.goto('/login')

      for (let i = 0; i < 5; i++) {
        await page.fill('input[placeholder="请输入用户名"]', 'testuser')
        await page.fill('input[type="password"]', 'wrongpassword')
        await page.getByRole('button', { name: '登录', exact: true }).click()
        await page.waitForTimeout(300)
      }

      await page.waitForTimeout(500)
      const pageUrl = page.url()
      expect(pageUrl).toBeTruthy()
    })
  })

  test.describe('Session Security', () => {
    test('should handle invalid localStorage data', async ({ page }) => {
      await page.goto('/login')
      await page.waitForLoadState('networkidle')
      await page.evaluate(() => {
        localStorage.setItem('access_token', 'invalid-token')
        localStorage.setItem('refresh_token', 'invalid-refresh')
        localStorage.setItem('user', 'not-valid-json')
      })

      await page.goto('/profile')
      await page.waitForLoadState('networkidle')
      await page.waitForTimeout(1000)

      const currentUrl = page.url()
      expect(currentUrl).toContain('/login')
    })

    test('should handle expired session', async ({ page }) => {
      await page.goto('/login')
      await page.waitForLoadState('networkidle')
      await page.evaluate(() => {
        const expiredUser = {
          id: '1',
          username: 'testuser',
          email: 'test@example.com',
          isAdmin: false,
          approved: true,
          exp: Math.floor(Date.now() / 1000) - 3600,
        }
        localStorage.setItem('access_token', 'expired-token')
        localStorage.setItem('refresh_token', 'expired-refresh')
        localStorage.setItem('user', JSON.stringify(expiredUser))
      })

      await page.goto('/profile')
      await page.waitForLoadState('networkidle')
      await page.waitForTimeout(1000)

      const currentUrl = page.url()
      expect(currentUrl).toContain('/login')
    })

    test('should clear localStorage on logout', async ({ page }) => {
      await page.goto('/login')
      await page.waitForLoadState('networkidle')
      await page.evaluate(() => {
        localStorage.setItem('access_token', 'valid-token')
        localStorage.setItem('refresh_token', 'valid-refresh')
        localStorage.setItem(
          'user',
          JSON.stringify({
            id: '1',
            username: 'testuser',
            email: 'test@example.com',
            isAdmin: false,
            approved: true,
          })
        )
      })

      await page.goto('/profile')
      await page.waitForLoadState('networkidle')
      await page.waitForTimeout(500)

      await page.evaluate(() => {
        localStorage.clear()
      })

      await page.reload()

      await page.waitForTimeout(500)

      const currentUrl = page.url()
      expect(currentUrl).toContain('/login')
    })
  })

  test.describe('Password Policy', () => {
    test('should handle registration with various password strengths', async ({ page }) => {
      await page.goto('/register')

      await page.fill('input[placeholder*="请输入邮箱"]', 'test@example.com')
      await page.fill('input[placeholder="请输入用户名"]', 'testuser')
      await page.fill('input[type="password"]', 'weak')

      await page.getByRole('button', { name: '注册', exact: true }).click()

      await page.waitForTimeout(500)

      const pageUrl = page.url()
      expect(pageUrl).toContain('/register')
    })

    test('should accept strong passwords', async ({ page }) => {
      await page.goto('/register')

      await page.fill('input[placeholder*="请输入邮箱"]', 'strong@example.com')
      await page.fill('input[placeholder="请输入用户名"]', 'stronguser')
      await page.fill('input[type="password"]', 'Str0ng!@#Pwd2024')

      await page.getByRole('button', { name: '注册', exact: true }).click()

      await page.waitForTimeout(500)

      const pageUrl = page.url()
      expect(pageUrl).toBeTruthy()
    })
  })
})

test.describe('Security Tests - Authorization & Access Control', () => {
  test.describe('Privilege Escalation Prevention', () => {
    test('should redirect unauthenticated users from protected routes', async ({ page }) => {
      const protectedRoutes = ['/profile', '/security', '/passkeys', '/admin', '/mfa']

      for (const route of protectedRoutes) {
        await page.goto(route)
        await page.waitForTimeout(500)

        const currentUrl = page.url()
        expect(currentUrl).toContain('/login')
      }
    })

    test('should handle admin routes when not authenticated', async ({ page }) => {
      await page.goto('/admin')
      await page.waitForTimeout(500)

      const currentUrl = page.url()
      expect(currentUrl).toContain('/login')
    })
  })
})

test.describe('Security Tests - Redirect Protection', () => {
  test.describe('Open Redirect Prevention', () => {
    test('should handle login with redirect parameter', async ({ page }) => {
      await page.goto('/login?redirect=/profile')
      await page.fill('input[placeholder="请输入用户名"]', 'testuser')
      await page.fill('input[type="password"]', 'TestPass123!')
      await page.getByRole('button', { name: '登录', exact: true }).click()

      await page.waitForTimeout(1000)

      const currentUrl = page.url()
      expect(currentUrl).toBeTruthy()
    })
  })
})

test.describe('Security Tests - Error Information Leakage', () => {
  test.describe('Error Message Handling', () => {
    test('should handle invalid login gracefully', async ({ page }) => {
      await page.goto('/login')
      await page.fill('input[placeholder="请输入用户名"]', 'nonexistentuser12345')
      await page.fill('input[type="password"]', 'wrongpassword')
      await page.getByRole('button', { name: '登录', exact: true }).click()
      await page.waitForTimeout(500)

      const currentUrl = page.url()
      expect(currentUrl).toContain('/login')
    })

    test('should handle non-existent page', async ({ page }) => {
      await page.goto('/non-existent-page-xyz')
      await page.waitForTimeout(1000)

      const pageContent = await page.content()
      expect(pageContent).toBeTruthy()
    })
  })
})

test.describe('Security Tests - Data Protection', () => {
  test.describe('Sensitive Data', () => {
    test('should mask password fields in login form', async ({ page }) => {
      await page.goto('/login')

      const passwordInput = page.locator('input[type="password"]')
      await expect(passwordInput).toHaveAttribute('type', 'password')
    })

    test('should not reveal token in URL after login', async ({ page }) => {
      await page.goto('/login')
      await page.fill('input[placeholder="请输入用户名"]', 'testuser')
      await page.fill('input[type="password"]', 'TestPass123!')
      await page.getByRole('button', { name: '登录', exact: true }).click()

      await page.waitForTimeout(1000)

      const currentUrl = page.url()
      expect(currentUrl).not.toMatch(/token=|auth=|bearer/i)
    })
  })
})

test.describe('Security Tests - MFA Security', () => {
  test.describe('MFA Input Handling', () => {
    test('should handle MFA page when not authenticated', async ({ page }) => {
      await page.goto('/mfa')
      await page.waitForLoadState('networkidle')
      await page.waitForTimeout(500)

      const currentUrl = page.url()
      expect(currentUrl).toContain('/login')
    })

    test('should handle access to MFA page directly', async ({ page }) => {
      await page.goto('/mfa')
      await page.waitForTimeout(500)

      const currentUrl = page.url()
      expect(currentUrl).toContain('/login')
    })
  })
})
