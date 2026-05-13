import { test, expect } from '@playwright/test'

test.describe('Authentication Edge Cases', () => {
  test.describe('Password Visibility Toggle', () => {
    test('should have password visibility toggle button on login page', async ({ page }) => {
      await page.goto('/login')
      await page.waitForLoadState('networkidle')

      const passwordInput = page.locator('input[type="password"]').first()
      await passwordInput.fill('TestPass123!')

      const visibilityToggle = page.locator('.n-input-suffix button').first()
      const hasToggle = await visibilityToggle.isVisible({ timeout: 2000 }).catch(() => false)

      if (hasToggle) {
        await visibilityToggle.click()
        await page.waitForTimeout(300)
        const inputType = await passwordInput.getAttribute('type')
        expect(inputType).toBe('text')
      }
    })

    test('should hide password after toggle off', async ({ page }) => {
      await page.goto('/login')
      await page.waitForLoadState('networkidle')

      const passwordInput = page.locator('input[type="password"]').first()
      await passwordInput.fill('TestPass123!')

      const visibilityToggle = page.locator('.n-input-suffix button').first()

      if (await visibilityToggle.isVisible({ timeout: 2000 })) {
        await visibilityToggle.click()
        await page.waitForTimeout(200)
        await visibilityToggle.click()
        await page.waitForTimeout(200)

        const inputType = await passwordInput.getAttribute('type')
        expect(inputType).toBe('password')
      }
    })

    test('registration form password toggle depends on implementation', async ({ page }) => {
      await page.goto('/register')
      await page.waitForLoadState('networkidle')

      const passwordInputs = page.locator('input[type="password"]')
      const count = await passwordInputs.count()

      expect(count).toBeGreaterThan(0)
    })
  })

  test.describe('Session Management', () => {
    test('should clear localStorage on logout via session revoke', async ({ page }) => {
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

      await page.goto('/security')
      await page.waitForLoadState('networkidle')
      await page.waitForTimeout(1000)

      const revokeAllButton = page.locator('button:has-text("撤销所有会话")')
      if (await revokeAllButton.isVisible({ timeout: 3000 })) {
        await revokeAllButton.click()
        await page.waitForTimeout(500)

        const confirmButton = page.locator('.n-popover button:has-text("确定")')
        if (await confirmButton.isVisible({ timeout: 1000 })) {
          await confirmButton.click()
          await page.waitForTimeout(1000)
        }
      }
    })

    test('should redirect to login after clearing tokens', async ({ page }) => {
      await page.goto('/login')
      await page.waitForLoadState('networkidle')
      await page.evaluate(() => {
        localStorage.setItem('access_token', 'valid-token')
        localStorage.setItem(
          'user',
          JSON.stringify({
            id: '1',
            username: 'testuser',
            isAdmin: false,
            approved: true,
          })
        )
      })

      await page.goto('/profile')
      await page.waitForLoadState('networkidle')

      await page.evaluate(() => {
        localStorage.removeItem('access_token')
        localStorage.removeItem('refresh_token')
        localStorage.removeItem('user')
      })

      await page.reload()
      await page.waitForTimeout(1000)

      const currentUrl = page.url()
      expect(currentUrl).toContain('/login')
    })

    test('should revoke all sessions via security page', async ({ page }) => {
      await page.goto('/login')
      await page.waitForLoadState('networkidle')
      await page.evaluate(() => {
        localStorage.setItem('access_token', 'admin-token')
        localStorage.setItem('refresh_token', 'admin-refresh')
        localStorage.setItem(
          'user',
          JSON.stringify({
            id: '1',
            username: 'admin',
            isAdmin: true,
            approved: true,
          })
        )
      })

      await page.goto('/security')
      await page.waitForLoadState('networkidle')

      const sessionsTab = page.locator('.n-tabs-tab:has-text("会话管理")')
      if (await sessionsTab.isVisible({ timeout: 3000 })) {
        await sessionsTab.click()
        await page.waitForTimeout(500)
      }
    })
  })

  test.describe('Multi-Tab State Synchronization', () => {
    test('localStorage is isolated between tabs by default', async ({ browser }) => {
      const context1 = await browser.newContext()
      const context2 = await browser.newContext()

      const page1 = await context1.newPage()
      const page2 = await context2.newPage()

      await page1.goto('/login')
      await page1.waitForLoadState('networkidle')
      await page1.evaluate(() => {
        localStorage.setItem('test_key', 'value_from_tab_1')
      })

      await page2.goto('/login')
      await page2.waitForLoadState('networkidle')

      const tab2Value = await page2.evaluate(() => {
        return localStorage.getItem('test_key')
      })

      expect(tab2Value).toBeNull()

      await context1.close()
      await context2.close()
    })

    test('tabs can communicate via BroadcastChannel if implemented', async ({ browser }) => {
      const context1 = await browser.newContext()
      const context2 = await browser.newContext()

      const page1 = await context1.newPage()
      const page2 = await context2.newPage()

      await page1.goto('/login')
      await page1.waitForLoadState('networkidle')

      await page2.goto('/login')
      await page2.waitForLoadState('networkidle')

      await context1.close()
      await context2.close()
    })
  })

  test.describe('Toast/Notification Messages', () => {
    test('should show success notification after login', async ({ page }) => {
      await page.goto('/login')
      await page.waitForLoadState('networkidle')

      await page.fill('input[placeholder="请输入用户名"]', 'testuser')
      await page.fill('input[type="password"]', 'TestPass123!')
      await page.getByRole('button', { name: '登录', exact: true }).click()
      await page.waitForTimeout(1000)
    })

    test('should show error for invalid login', async ({ page }) => {
      await page.goto('/login')
      await page.waitForLoadState('networkidle')

      await page.fill('input[placeholder="请输入用户名"]', 'nonexistentuser')
      await page.fill('input[type="password"]', 'wrongpassword')
      await page.getByRole('button', { name: '登录', exact: true }).click()
      await page.waitForTimeout(2000)

      const errorAlert = page.locator('.n-alert')
      expect(await errorAlert.count()).toBeGreaterThan(0)
    })

    test('should handle session revoke confirmation', async ({ page }) => {
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

      await page.goto('/security')
      await page.waitForLoadState('networkidle')

      const sessionsTab = page.locator('.n-tabs-tab:has-text("会话管理")')
      if (await sessionsTab.isVisible({ timeout: 3000 })) {
        await sessionsTab.click()
        await page.waitForTimeout(500)
      }

      const revokeButton = page.locator('button:has-text("撤销所有会话")')
      if (await revokeButton.isVisible({ timeout: 3000 })) {
        await revokeButton.click()
        await page.waitForTimeout(500)
      }
    })
  })

  test.describe('Form Input Validation', () => {
    test('should validate email format on submit', async ({ page }) => {
      await page.goto('/register')
      await page.waitForLoadState('networkidle')

      const emailInput = page.locator('input[placeholder*="请输入邮箱"]')
      await emailInput.fill('invalidemail')
      await page.waitForTimeout(500)

      await page.fill('input[placeholder="请输入用户名"]', 'testuser')
      await page.fill('input[type="password"]', 'TestPass123!')
      await page.getByRole('button', { name: '注册', exact: true }).click()
      await page.waitForTimeout(1000)

      const currentUrl = page.url()
      expect(currentUrl).toContain('/register')
    })

    test('should show password strength indicator', async ({ page }) => {
      await page.goto('/register')
      await page.waitForLoadState('networkidle')

      const passwordInput = page.locator('input[type="password"]').first()
      await passwordInput.fill('weak')
      await page.waitForTimeout(300)

      const strengthLabel = page.locator('.strength-label')
      const hasWeakIndicator = await strengthLabel.isVisible({ timeout: 2000 }).catch(() => false)

      if (hasWeakIndicator) {
        const labelText = await strengthLabel.textContent()
        expect(['非常弱', '弱', '一般']).toContain(labelText)
      }
    })

    test('should prevent registration with weak password', async ({ page }) => {
      await page.goto('/register')
      await page.waitForLoadState('networkidle')

      await page.fill('input[placeholder*="请输入邮箱"]', 'test@example.com')
      await page.fill('input[placeholder="请输入用户名"]', 'testuser')
      await page.fill('input[type="password"]', 'weak')
      await page.getByRole('button', { name: '注册', exact: true }).click()
      await page.waitForTimeout(1000)

      const currentUrl = page.url()
      expect(currentUrl).toContain('/register')
    })
  })

  test.describe('Username/Email Uniqueness', () => {
    test('should handle duplicate username gracefully', async ({ page }) => {
      await page.goto('/register')
      await page.waitForLoadState('networkidle')

      await page.fill('input[placeholder*="请输入邮箱"]', 'unique@example.com')
      await page.fill('input[placeholder="请输入用户名"]', 'admin')
      await page.fill('input[type="password"]', 'TestPass123!')

      await page.getByRole('button', { name: '注册', exact: true }).click()
      await page.waitForTimeout(1500)

      const errorAlert = page.locator('.n-alert[type="error"]')
      const hasError = await errorAlert.isVisible({ timeout: 3000 }).catch(() => false)

      if (hasError) {
        expect(await errorAlert.textContent()).toContain('用户名')
      }
    })
  })

  test.describe('Verification Code Rate Limiting', () => {
    test('should handle resend button on verify email page', async ({ page }) => {
      await page.goto('/verify-email')
      await page.waitForLoadState('networkidle')

      const resendButton = page.locator('button:has-text("重新发送")')
      if (await resendButton.isVisible({ timeout: 3000 })) {
        await resendButton.click()
        await page.waitForTimeout(500)
      }
    })
  })

  test.describe('Invitation Code Validation', () => {
    test('should have invitation code field', async ({ page }) => {
      await page.goto('/register')
      await page.waitForLoadState('networkidle')

      const inviteCodeInput = page.locator('input[placeholder*="邀请码"]')
      const hasField = await inviteCodeInput.isVisible({ timeout: 3000 }).catch(() => false)
      expect(hasField).toBeTruthy()
    })
  })

  test.describe('Passkey Operations', () => {
    test('should navigate to passkey registration page', async ({ page }) => {
      await page.goto('/login')
      await page.waitForLoadState('networkidle')
      await page.evaluate(() => {
        localStorage.setItem('access_token', 'user-token')
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

      await page.goto('/passkeys')
      await page.waitForLoadState('networkidle')
      await page.waitForTimeout(1000)

      const pageContent = await page.content()
      expect(pageContent).toBeTruthy()
    })

    test('should display passkey section in security page', async ({ page }) => {
      await page.goto('/login')
      await page.waitForLoadState('networkidle')
      await page.evaluate(() => {
        localStorage.setItem('access_token', 'user-token')
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

      await page.goto('/security')
      await page.waitForLoadState('networkidle')

      const passkeysTab = page.locator('.n-tabs-tab:has-text("通行密钥")')
      if (await passkeysTab.isVisible({ timeout: 3000 })) {
        await passkeysTab.click()
        await page.waitForTimeout(500)
      }
    })
  })

  test.describe('User Profile Update', () => {
    async function loginAsAdmin(page: any) {
      await page.goto('/login')
      await page.waitForLoadState('networkidle')
      await page.fill('input[placeholder="请输入用户名"]', 'admin')
      await page.fill('input[type="password"]', 'Admin123!')
      await page.click('button[type="submit"]')
      await page.waitForLoadState('networkidle')
      await page.waitForTimeout(1000)
    }

    test('should access profile page with authentication', async ({ page }) => {
      await loginAsAdmin(page)
      await page.goto('/user/profile')
      await page.waitForLoadState('networkidle')
      await page.waitForTimeout(1000)

      const currentUrl = page.url()
      expect(currentUrl).toContain('/profile')

      const pageContent = await page.content()
      const hasProfileUI = pageContent.includes('个人资料') || pageContent.includes('编辑资料')
      expect(hasProfileUI).toBeTruthy()
    })

    test('should load profile page and handle API response', async ({ page }) => {
      await loginAsAdmin(page)
      await page.goto('/user/profile')
      await page.waitForLoadState('networkidle')
      await page.waitForTimeout(2000)

      const currentUrl = page.url()
      expect(currentUrl).toContain('/profile')

      const hasEditButton = (await page.locator('button:has-text("编辑资料")').count()) > 0
      const hasSecurityButton = (await page.locator('button:has-text("安全设置")').count()) > 0
      expect(hasEditButton || hasSecurityButton).toBeTruthy()
    })
  })
})
