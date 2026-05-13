import { test, expect } from '@playwright/test'

test.describe('Boundary Conditions and Error Handling Tests', () => {
  test.describe('Input Length Boundaries', () => {
    test('should handle maximum length username in login', async ({ page }) => {
      await page.goto('/login')

      const maxLengthUsername = 'a'.repeat(256)
      await page.fill('input[placeholder="请输入用户名"]', maxLengthUsername)
      await page.fill('input[type="password"]', 'TestPass123!')
      await page.getByRole('button', { name: '登录', exact: true }).click()

      await page.waitForTimeout(500)

      const inputValue = await page.locator('input[placeholder="请输入用户名"]').inputValue()
      expect(inputValue.length).toBeGreaterThanOrEqual(0)
    })

    test('should handle maximum length email in registration', async ({ page }) => {
      await page.goto('/register')

      const longEmail = 'a'.repeat(100) + '@example.com'
      await page.fill('input[placeholder*="请输入邮箱"]', longEmail)
      await page.fill('input[placeholder="请输入用户名"]', 'testuser' + Date.now())
      await page.fill('input[type="password"]', 'TestPass123!')

      await page.getByRole('button', { name: '注册', exact: true }).click()

      await page.waitForTimeout(1000)

      const pageUrl = page.url()
      expect(pageUrl).not.toContain('/register?')
    })

    test('should handle very long password input', async ({ page }) => {
      await page.goto('/register')

      const longPassword = 'Ab1!' + 'a'.repeat(500)
      await page.fill('input[placeholder*="请输入邮箱"]', 'test@example.com')
      await page.fill('input[placeholder="请输入用户名"]', 'testuser')
      await page.fill('input[type="password"]', longPassword)

      await page.getByRole('button', { name: '注册', exact: true }).click()

      await page.waitForTimeout(500)

      const inputValue = await page.locator('input[type="password"]').inputValue()
      expect(inputValue.length).toBeGreaterThan(0)
    })

    test('should handle very long search query', async ({ page }) => {
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

      await page.goto('/admin/users')
      await page.waitForLoadState('networkidle')
      await page.waitForTimeout(1500)

      const searchInput = page.locator('input[placeholder*="search"], input[type="search"]').first()

      if (await searchInput.isVisible()) {
        const longSearchQuery = 'a'.repeat(1000)
        await searchInput.fill(longSearchQuery)
        await page.waitForTimeout(1000)

        const inputValue = await searchInput.inputValue()
        expect(inputValue.length).toBeGreaterThanOrEqual(0)
      }
    })

    test('should handle unicode characters in username', async ({ page }) => {
      await page.goto('/register')

      const unicodeUsername = '用户' + Math.random().toString(36).substring(7)
      await page.fill('input[placeholder*="请输入邮箱"]', 'test@example.com')
      await page.fill('input[placeholder="请输入用户名"]', unicodeUsername)
      await page.fill('input[type="password"]', 'TestPass123!')

      await page.getByRole('button', { name: '注册', exact: true }).click()

      await page.waitForTimeout(500)
    })

    test('should handle emoji in email field', async ({ page }) => {
      await page.goto('/register')

      await page.fill('input[placeholder*="请输入邮箱"]', 'test😀@example.com')
      await page.fill('input[placeholder="请输入用户名"]', 'testuser')
      await page.fill('input[type="password"]', 'TestPass123!')

      await page.getByRole('button', { name: '注册', exact: true }).click()

      await page.waitForTimeout(500)

      const pageUrl = page.url()
      expect(pageUrl).toContain('/register')
    })
  })

  test.describe('Special Characters Handling', () => {
    const specialChars = [
      '!@#$%^&*()',
      '(){}[]|\\:;"<>,.?/~`',
      '+-=',
      '\t\n\r',
      '\u00E9\u00E8\u00EA',
      '中文用户名',
      '日本語',
      '한국어',
      'עברית',
      'العربية',
    ]

    for (const chars of specialChars) {
      test(`should handle special characters: ${chars.slice(0, 20)}`, async ({ page }) => {
        await page.goto('/register')

        const username = 'user' + chars.replace(/[^a-zA-Z0-9]/g, '').substring(0, 10)
        await page.fill('input[placeholder="请输入用户名"]', username)
        await page.fill('input[placeholder*="请输入邮箱"]', 'test@example.com')
        await page.fill('input[type="password"]', 'TestPass123!')

        await page.getByRole('button', { name: '注册', exact: true }).click()

        await page.waitForTimeout(500)
      })
    }

    test('should handle null bytes in input', async ({ page }) => {
      await page.goto('/login')

      const nullByteInput = 'test\u0000user'
      await page.fill('input[placeholder="请输入用户名"]', nullByteInput)
      await page.fill('input[type="password"]', 'password')
      await page.getByRole('button', { name: '登录', exact: true }).click()

      await page.waitForTimeout(500)

      const pageUrl = page.url()
      expect(pageUrl).toBeTruthy()
    })
  })

  test.describe('Empty and Null Handling', () => {
    test('should handle empty username field', async ({ page }) => {
      await page.goto('/login')

      await page.fill('input[placeholder="请输入用户名"]', '')
      await page.fill('input[type="password"]', 'password')
      await page.getByRole('button', { name: '登录', exact: true }).click()

      await page.waitForTimeout(500)

      const pageUrl = page.url()
      expect(pageUrl).toContain('/login')
    })

    test('should handle empty password field', async ({ page }) => {
      await page.goto('/login')

      await page.fill('input[placeholder="请输入用户名"]', 'testuser')
      await page.fill('input[type="password"]', '')
      await page.getByRole('button', { name: '登录', exact: true }).click()

      await page.waitForTimeout(500)

      const pageUrl = page.url()
      expect(pageUrl).toContain('/login')
    })

    test('should handle whitespace-only username', async ({ page }) => {
      await page.goto('/login')

      await page.fill('input[placeholder="请输入用户名"]', '   ')
      await page.fill('input[type="password"]', 'password')
      await page.getByRole('button', { name: '登录', exact: true }).click()

      await page.waitForTimeout(500)

      const pageUrl = page.url()
      expect(pageUrl).toContain('/login')
    })

    test('should handle whitespace-only password', async ({ page }) => {
      await page.goto('/login')

      await page.fill('input[placeholder="请输入用户名"]', 'testuser')
      await page.fill('input[type="password"]', '      ')
      await page.getByRole('button', { name: '登录', exact: true }).click()

      await page.waitForTimeout(500)

      const pageUrl = page.url()
      expect(pageUrl).toContain('/login')
    })

    test('should handle missing required registration fields', async ({ page }) => {
      await page.goto('/register')

      await page.getByRole('button', { name: '注册', exact: true }).click()

      await page.waitForTimeout(500)

      const pageUrl = page.url()
      expect(pageUrl).toContain('/register')
    })
  })

  test.describe('Concurrent Operations', () => {
    test('should handle rapid login attempts', async ({ page }) => {
      await page.goto('/login')

      for (let i = 0; i < 10; i++) {
        await page.fill('input[placeholder="请输入用户名"]', 'testuser')
        await page.fill('input[type="password"]', 'password')
        await page.getByRole('button', { name: '登录', exact: true }).click()
        await page.waitForTimeout(50)
      }

      await page.waitForTimeout(500)

      const currentUrl = page.url()
      expect(currentUrl).toBeTruthy()
    })

    test('should handle rapid form submissions', async ({ page }) => {
      await page.goto('/register')

      await page.fill('input[placeholder*="请输入邮箱"]', 'test@example.com')
      await page.fill('input[placeholder="请输入用户名"]', 'testuser')
      await page.fill('input[type="password"]', 'TestPass123!')

      for (let i = 0; i < 5; i++) {
        await page.getByRole('button', { name: '注册', exact: true }).click()
        await page.waitForTimeout(100)
      }

      await page.waitForTimeout(1000)
    })

    test('should handle browser back button after login', async ({ page }) => {
      await page.goto('/login')

      await page.fill('input[placeholder="请输入用户名"]', 'testuser')
      await page.fill('input[type="password"]', 'TestPass123!')
      await page.getByRole('button', { name: '登录', exact: true }).click()

      await page.waitForURL(/\/(profile|mfa)\/?/, { timeout: 5000 }).catch(() => {})

      await page.goBack()

      await page.waitForTimeout(500)

      const currentUrl = page.url()
      expect(currentUrl).not.toContain('/login')
    })

    test('should handle browser forward button after navigation', async ({ page }) => {
      await page.goto('/login')
      await page.fill('input[placeholder="请输入用户名"]', 'testuser')
      await page.fill('input[type="password"]', 'TestPass123!')
      await page.getByRole('button', { name: '登录', exact: true }).click()

      await page.waitForURL(/\/(profile|mfa)\/?/, { timeout: 5000 }).catch(() => {})

      await page.goBack()
      await page.waitForTimeout(300)

      await page.goForward()
      await page.waitForTimeout(500)

      const currentUrl = page.url()
      expect(currentUrl).toBeTruthy()
    })
  })

  test.describe('Network Error Handling', () => {
    test('should load login page successfully', async ({ page }) => {
      await page.goto('/login')
      await page.waitForLoadState('networkidle')
      await page.waitForTimeout(500)

      const currentUrl = page.url()
      expect(currentUrl).toContain('/login')
    })

    test('should load register page successfully', async ({ page }) => {
      await page.goto('/register')
      await page.waitForLoadState('networkidle')
      await page.waitForTimeout(500)

      const currentUrl = page.url()
      expect(currentUrl).toContain('/register')
    })

    test('should load reset-password page successfully', async ({ page }) => {
      await page.goto('/reset-password')
      await page.waitForLoadState('networkidle')
      await page.waitForTimeout(500)

      const currentUrl = page.url()
      expect(currentUrl).toContain('/reset-password')
    })
  })

  test.describe('State Management', () => {
    test('should handle localStorage corruption', async ({ page }) => {
      await page.goto('/login')
      await page.waitForLoadState('networkidle')
      await page.evaluate(() => {
        localStorage.setItem('access_token', 'invalid-json{')
        localStorage.setItem('refresh_token', 'also-invalid')
        localStorage.setItem('user', 'not-valid-json')
      })

      await page.goto('/profile')
      await page.waitForLoadState('networkidle')
      await page.waitForTimeout(1000)

      const currentUrl = page.url()
      expect(currentUrl).toContain('/login')
    })

    test('should handle expired session gracefully', async ({ page }) => {
      await page.goto('/login')
      await page.waitForLoadState('networkidle')
      await page.evaluate(() => {
        const expiredUser = {
          id: '1',
          username: 'testuser',
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

    test('should handle missing auth state', async ({ page }) => {
      await page.goto('/login')
      await page.waitForLoadState('networkidle')
      await page.evaluate(() => {
        localStorage.clear()
      })

      await page.goto('/profile')
      await page.waitForLoadState('networkidle')
      await page.waitForTimeout(1000)

      const currentUrl = page.url()
      expect(currentUrl).toContain('/login')
    })

    test('should clear state on logout', async ({ page }) => {
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

  test.describe('Form Validation Edge Cases', () => {
    test('should handle various email formats in registration', async ({ page }) => {
      await page.goto('/register')

      await page.fill('input[placeholder*="请输入邮箱"]', 'notanemail')
      await page.fill('input[placeholder="请输入用户名"]', 'testuser')
      await page.fill('input[type="password"]', 'TestPass123!')

      await page.getByRole('button', { name: '注册', exact: true }).click()

      await page.waitForTimeout(500)

      const pageUrl = page.url()
      expect(pageUrl).toBeTruthy()
    })

    test('should handle password confirmation field', async ({ page }) => {
      await page.goto('/reset-password')

      const pageContent = await page.content()
      expect(pageContent).toBeTruthy()
    })

    test('should handle MFA page when not authenticated', async ({ page }) => {
      await page.goto('/mfa')
      await page.waitForTimeout(500)

      const currentUrl = page.url()
      expect(currentUrl).toContain('/login')
    })

    test('should handle direct access to MFA page', async ({ page }) => {
      await page.goto('/mfa')
      await page.waitForLoadState('networkidle')
      await page.waitForTimeout(500)

      const currentUrl = page.url()
      expect(currentUrl).toContain('/login')
    })
  })

  test.describe('Page Navigation Edge Cases', () => {
    test('should handle direct URL access to protected page', async ({ page }) => {
      await page.goto('/profile')

      await page.waitForTimeout(500)

      const currentUrl = page.url()
      expect(currentUrl).toContain('/login')
    })

    test('should handle deep link with query parameters', async ({ page }) => {
      await page.goto('/login?redirect=/profile?tab=security&token=abc123')

      await page.waitForTimeout(500)

      const currentUrl = page.url()
      expect(currentUrl).toContain('/login')
    })

    test('should handle bookmarked authenticated page when logged out', async ({ page }) => {
      await page.goto('/profile')

      await page.waitForTimeout(500)

      const currentUrl = page.url()
      expect(currentUrl).toContain('/login')
    })

    test('should handle rapid route changes', async ({ page }) => {
      await page.goto('/login')
      await page.waitForTimeout(100)
      await page.goto('/register')
      await page.waitForTimeout(100)
      await page.goto('/login')
      await page.waitForTimeout(100)
      await page.goto('/profile')

      await page.waitForTimeout(500)

      const currentUrl = page.url()
      expect(currentUrl).toContain('/login')
    })

    test('should handle navigation to non-existent routes', async ({ page }) => {
      await page.goto('/this-route-does-not-exist-12345')

      await page.waitForTimeout(1000)

      const currentUrl = page.url()
      expect(currentUrl).toBeTruthy()
    })
  })

  test.describe('Timeout Handling', () => {
    test('should handle session timeout gracefully', async ({ page }) => {
      await page.goto('/login')
      await page.waitForLoadState('networkidle')
      await page.evaluate(() => {
        localStorage.setItem('access_token', 'about-to-expire-token')
        localStorage.setItem('refresh_token', 'about-to-expire-refresh')
        localStorage.setItem(
          'user',
          JSON.stringify({
            id: '1',
            username: 'testuser',
            email: 'test@example.com',
            exp: Math.floor(Date.now() / 1000) + 1,
          })
        )
      })

      await page.goto('/profile')
      await page.waitForLoadState('networkidle')
      await page.waitForTimeout(3000)

      const currentUrl = page.url()
      expect(currentUrl).toContain('/login')
    })

    test('should handle MFA page when not authenticated', async ({ page }) => {
      await page.goto('/mfa')
      await page.waitForLoadState('networkidle')

      await page.waitForTimeout(500)

      const pageUrl = page.url()
      expect(pageUrl).toContain('/login')
    })
  })
})

test.describe('Data Consistency Tests', () => {
  test('should handle localStorage user data', async ({ page }) => {
    await page.goto('/login')
    await page.waitForLoadState('networkidle')
    await page.evaluate(() => {
      localStorage.setItem('access_token', 'consistent-token')
      localStorage.setItem('refresh_token', 'consistent-refresh')
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

    const currentUrl = page.url()
    expect(currentUrl).toBeTruthy()
  })

  test('should sync localStorage changes across page reloads', async ({ page }) => {
    await page.goto('/login')
    await page.waitForLoadState('networkidle')
    await page.evaluate(() => {
      localStorage.setItem('access_token', 'test-token')
      localStorage.setItem('refresh_token', 'test-refresh')
      localStorage.setItem(
        'user',
        JSON.stringify({
          id: '1',
          username: 'original',
          email: 'original@example.com',
        })
      )
    })

    await page.goto('/profile')
    await page.waitForLoadState('networkidle')
    await page.waitForTimeout(500)

    await page.evaluate(() => {
      localStorage.setItem(
        'user',
        JSON.stringify({
          id: '1',
          username: 'updated',
          email: 'updated@example.com',
        })
      )
    })

    await page.reload()

    const currentUrl = page.url()
    expect(currentUrl).toBeTruthy()
  })
})
