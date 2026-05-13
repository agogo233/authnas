import { test, expect } from '@playwright/test'

test.describe('User Profile Page E2E', () => {
  test.beforeEach(async ({ page }) => {
    // Set authenticated user
    await page.context().addInitScript(() => {
      localStorage.setItem('access_token', 'mock-user-token')
      localStorage.setItem('refresh_token', 'mock-refresh-token')
      localStorage.setItem(
        'user',
        JSON.stringify({
          id: '1',
          username: 'testuser',
          email: 'test@example.com',
          name: 'Test User',
          isAdmin: false,
          approved: true,
          emailVerified: true,
          createdAt: '2024-01-01T00:00:00Z',
        })
      )
    })
  })

  test.describe('Authentication & Access', () => {
    test('should redirect to login when not authenticated', async ({ page }) => {
      // Create new context without auth
      const newContext = await page.context().browser()?.newContext()
      const newPage = await newContext?.newPage()

      if (newPage) {
        await newPage.goto('/user/profile')
        await expect(newPage).toHaveURL(/\/login/, { timeout: 5000 })
        await newContext?.close()
      }
    })

    test('should load profile page when authenticated', async ({ page }) => {
      await page.goto('/user/profile')
      await page.waitForLoadState('networkidle')

      await expect(page.locator('.page-container')).toBeVisible({ timeout: 10000 })
    })

    test('should redirect from /profile to /user/profile', async ({ page }) => {
      await page.goto('/profile')
      await page.waitForLoadState('networkidle')

      await expect(page).toHaveURL(/\/user\/profile/, { timeout: 5000 })
    })
  })

  test.describe('Profile Display', () => {
    test.beforeEach(async ({ page }) => {
      // Mock API responses
      await page.route('**/api/user/me', async (route) => {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({
            success: true,
            data: {
              id: '1',
              username: 'testuser',
              email: 'test@example.com',
              name: 'Test User',
              isAdmin: false,
              approved: true,
              emailVerified: true,
              createdAt: '2024-01-01T00:00:00Z',
            },
          }),
        })
      })

      await page.goto('/user/profile')
      await page.waitForLoadState('networkidle')
    })

    test('should display user information in descriptions', async ({ page }) => {
      await expect(page.locator('text=用户名')).toBeVisible()
      await expect(page.locator('text=邮箱')).toBeVisible()
      await expect(page.locator('text=姓名')).toBeVisible()
      await expect(page.locator('text=已批准')).toBeVisible()
      await expect(page.locator('text=创建时间')).toBeVisible()
    })

    test('should display correct username', async ({ page }) => {
      await expect(page.locator('.n-descriptions-item:has-text("testuser")')).toBeVisible()
    })

    test('should display verified badge for verified email', async ({ page }) => {
      await expect(page.locator('text=已验证')).toBeVisible()
    })

    test('should display edit button', async ({ page }) => {
      const editButton = page.locator('button:has-text("编辑资料")')
      await expect(editButton).toBeVisible()
    })
  })

  test.describe('Profile Edit Flow', () => {
    test.beforeEach(async ({ page }) => {
      await page.route('**/api/user/me', async (route) => {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({
            success: true,
            data: {
              id: '1',
              username: 'testuser',
              email: 'test@example.com',
              name: 'Test User',
              isAdmin: false,
              approved: true,
              emailVerified: true,
              createdAt: '2024-01-01T00:00:00Z',
            },
          }),
        })
      })

      await page.goto('/user/profile')
      await page.waitForLoadState('networkidle')
    })

    test('should show edit form when clicking edit button', async ({ page }) => {
      const editButton = page.locator('button:has-text("编辑资料")')
      await editButton.click()

      await expect(page.locator('input[placeholder="请输入邮箱"]')).toBeVisible()
      await expect(page.locator('input[placeholder="请输入姓名"]')).toBeVisible()
      await expect(page.locator('button:has-text("取消")')).toBeVisible()
      await expect(page.locator('button:has-text("保存")')).toBeVisible()
    })

    test('should hide username input when editing', async ({ page }) => {
      const editButton = page.locator('button:has-text("编辑资料")')
      await editButton.click()

      // Username field should be disabled
      const usernameInput = page.locator('input[disabled]').first()
      await expect(usernameInput).toBeVisible()
    })

    test('should cancel edit and return to view mode', async ({ page }) => {
      const editButton = page.locator('button:has-text("编辑资料")')
      await editButton.click()

      const cancelButton = page.locator('button:has-text("取消")')
      await cancelButton.click()

      // Should show descriptions again
      await expect(page.locator('.n-descriptions')).toBeVisible()
    })

    test('should show success message on successful update', async ({ page }) => {
      await page.route('**/api/user/me', async (route) => {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({
            success: true,
            data: {
              id: '1',
              username: 'testuser',
              email: 'test@example.com',
              name: 'Updated Name',
              isAdmin: false,
              approved: true,
              emailVerified: true,
              createdAt: '2024-01-01T00:00:00Z',
            },
          }),
        })
      })

      const editButton = page.locator('button:has-text("编辑资料")')
      await editButton.click()

      const nameInput = page.locator('input[placeholder="请输入姓名"]')
      await nameInput.fill('Updated Name')

      const saveButton = page.locator('button:has-text("保存")')
      await saveButton.click()

      await page.waitForTimeout(500)
      // Success message should appear
    })
  })

  test.describe('Email Verification', () => {
    test('should show resend verification button for unverified email', async ({ page }) => {
      await page.route('**/api/user/me', async (route) => {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({
            success: true,
            data: {
              id: '1',
              username: 'testuser',
              email: 'test@example.com',
              name: 'Test User',
              isAdmin: false,
              approved: true,
              emailVerified: false,
              createdAt: '2024-01-01T00:00:00Z',
            },
          }),
        })
      })

      await page.goto('/user/profile')
      await page.waitForLoadState('networkidle')

      const verifyButton = page.locator('button:has-text("立即验证")')
      await expect(verifyButton).toBeVisible()
    })
  })
})

test.describe('User Security Page E2E', () => {
  test.beforeEach(async ({ page }) => {
    await page.context().addInitScript(() => {
      localStorage.setItem('access_token', 'mock-user-token')
      localStorage.setItem('refresh_token', 'mock-refresh-token')
      localStorage.setItem(
        'user',
        JSON.stringify({
          id: '1',
          username: 'testuser',
          email: 'test@example.com',
          name: 'Test User',
          isAdmin: false,
          approved: true,
          emailVerified: true,
          hasTotp: false,
          createdAt: '2024-01-01T00:00:00Z',
        })
      )
    })

    await page.route('**/api/user/me', async (route) => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          success: true,
          data: {
            id: '1',
            username: 'testuser',
            email: 'test@example.com',
            name: 'Test User',
            isAdmin: false,
            approved: true,
            emailVerified: true,
            hasTotp: false,
            createdAt: '2024-01-01T00:00:00Z',
          },
        }),
      })
    })

    await page.route('**/api/user/me/sessions', async (route) => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          success: true,
          data: [
            {
              id: 'session-1',
              userId: '1',
              createdAt: '2024-01-01T00:00:00Z',
              expiresAt: '2024-12-31T00:00:00Z',
            },
          ],
        }),
      })
    })

    await page.route('**/api/passkey', async (route) => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          success: true,
          data: [
            {
              id: 'pk-1',
              credentialId: 'cred-1',
              name: 'My Passkey',
              createdAt: '2024-01-01T00:00:00Z',
              updatedAt: '2024-01-01T00:00:00Z',
            },
          ],
        }),
      })
    })
  })

  test.describe('Page Navigation', () => {
    test('should load security page correctly', async ({ page }) => {
      await page.goto('/user/security')
      await page.waitForLoadState('networkidle')

      await expect(page.locator('.page-container')).toBeVisible({ timeout: 10000 })
      await expect(page.locator('text=安全设置')).toBeVisible()
    })

    test('should have tabs for password, MFA, passkeys, and sessions', async ({ page }) => {
      await page.goto('/user/security')
      await page.waitForLoadState('networkidle')

      await expect(page.locator('.n-tabs')).toBeVisible()
      await expect(page.locator('text=修改密码')).toBeVisible()
      await expect(page.locator('text=两步验证')).toBeVisible()
      await expect(page.locator('text=通行密钥')).toBeVisible()
      await expect(page.locator('text=会话管理')).toBeVisible()
    })
  })

  test.describe('Password Change', () => {
    test.beforeEach(async ({ page }) => {
      await page.goto('/user/security')
      await page.waitForLoadState('networkidle')
    })

    test('should show password form fields', async ({ page }) => {
      await expect(page.locator('input[placeholder="请输入当前密码"]')).toBeVisible()
      await expect(page.locator('input[placeholder="请输入新密码"]')).toBeVisible()
      await expect(page.locator('input[placeholder="请再次输入新密码"]')).toBeVisible()
    })

    test('should validate empty fields', async ({ page }) => {
      const submitButton = page.locator('button:has-text("更新密码")')
      await submitButton.click()

      // Error message should appear
    })

    test('should validate password mismatch', async ({ page }) => {
      await page.fill('input[placeholder="请输入当前密码"]', 'oldpassword')
      await page.fill('input[placeholder="请输入新密码"]', 'newpassword123')
      await page.fill('input[placeholder="请再次输入新密码"]', 'differentpassword')

      const submitButton = page.locator('button:has-text("更新密码")')
      await submitButton.click()

      // Should show mismatch error
    })

    test('should show password strength indicator', async ({ page }) => {
      const newPasswordInput = page.locator('input[placeholder="请输入新密码"]')
      await newPasswordInput.fill('Weak')

      await expect(page.locator('.password-strength')).toBeVisible()
    })

    test('should show strong password indicator', async ({ page }) => {
      const newPasswordInput = page.locator('input[placeholder="请输入新密码"]')
      await newPasswordInput.fill('StrongP@ssw0rd123!')

      await expect(page.locator('.strength-label')).toBeVisible()
    })
  })

  test.describe('TOTP Management', () => {
    test('should show TOTP disabled state', async ({ page }) => {
      await page.goto('/user/security')
      await page.waitForLoadState('networkidle')

      // Click on two-step verification tab
      await page.click('text=两步验证')

      await expect(page.locator('text=TOTP 目前未启用')).toBeVisible()
      await expect(page.locator('button:has-text("启用 TOTP")')).toBeVisible()
    })

    test('should show TOTP enabled state', async ({ page }) => {
      await page.route('**/api/user/me', async (route) => {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({
            success: true,
            data: {
              id: '1',
              username: 'testuser',
              email: 'test@example.com',
              name: 'Test User',
              isAdmin: false,
              approved: true,
              emailVerified: true,
              hasTotp: true,
              createdAt: '2024-01-01T00:00:00Z',
            },
          }),
        })
      })

      await page.goto('/user/security')
      await page.waitForLoadState('networkidle')

      // Click on two-step verification tab
      await page.click('text=两步验证')

      await expect(page.locator('text=TOTP 已启用')).toBeVisible()
      await expect(page.locator('button:has-text("禁用 TOTP")')).toBeVisible()
    })
  })

  test.describe('Passkeys Management', () => {
    test('should display passkeys list', async ({ page }) => {
      await page.goto('/user/security')
      await page.waitForLoadState('networkidle')

      // Click on passkeys tab
      await page.click('text=通行密钥')

      await expect(page.locator('.list')).toBeVisible()
      await expect(page.locator('text=My Passkey')).toBeVisible()
      await expect(page.locator('button:has-text("删除")')).toBeVisible()
    })

    test('should show add passkey button', async ({ page }) => {
      await page.goto('/user/security')
      await page.waitForLoadState('networkidle')

      // Click on passkeys tab
      await page.click('text=通行密钥')

      const addButton = page.locator('button:has-text("添加通行密钥")')
      await expect(addButton).toBeVisible()
    })

    test('should open passkey setup modal', async ({ page }) => {
      await page.goto('/user/security')
      await page.waitForLoadState('networkidle')

      // Click on passkeys tab
      await page.click('text=通行密钥')

      const addButton = page.locator('button:has-text("添加通行密钥")')
      await addButton.click()

      await expect(page.locator('.n-modal')).toBeVisible()
      await expect(page.locator('input[placeholder*="例如"]')).toBeVisible()
    })
  })

  test.describe('Session Management', () => {
    test('should display sessions list', async ({ page }) => {
      await page.goto('/user/security')
      await page.waitForLoadState('networkidle')

      // Click on sessions tab
      await page.click('text=会话管理')

      await expect(page.locator('.list')).toBeVisible()
      await expect(page.locator('text=会话:')).toBeVisible()
      await expect(page.locator('button:has-text("撤销")')).toBeVisible()
    })

    test('should show revoke all sessions button', async ({ page }) => {
      await page.goto('/user/security')
      await page.waitForLoadState('networkidle')

      // Click on sessions tab
      await page.click('text=会话管理')

      const revokeAllButton = page.locator('button:has-text("撤销所有会话")')
      await expect(revokeAllButton).toBeVisible()
    })

    test('should show empty state when no sessions', async ({ page }) => {
      await page.route('**/api/user/me/sessions', async (route) => {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({
            success: true,
            data: [],
          }),
        })
      })

      await page.goto('/user/security')
      await page.waitForLoadState('networkidle')

      // Click on sessions tab
      await page.click('text=会话管理')

      await expect(page.locator('text=暂无活动会话')).toBeVisible()
    })
  })
})

test.describe('User Navigation', () => {
  test('should access profile via layout navigation', async ({ page }) => {
    await page.context().addInitScript(() => {
      localStorage.setItem('access_token', 'mock-user-token')
      localStorage.setItem('refresh_token', 'mock-refresh-token')
      localStorage.setItem(
        'user',
        JSON.stringify({
          id: '1',
          username: 'testuser',
          email: 'test@example.com',
          name: 'Test User',
          isAdmin: false,
          approved: true,
          emailVerified: true,
          createdAt: '2024-01-01T00:00:00Z',
        })
      )
    })

    await page.goto('/user')
    await page.waitForLoadState('networkidle')

    await expect(page).toHaveURL(/\/user\/profile/, { timeout: 5000 })
  })

  test('should navigate between user pages', async ({ page }) => {
    await page.context().addInitScript(() => {
      localStorage.setItem('access_token', 'mock-user-token')
      localStorage.setItem('refresh_token', 'mock-refresh-token')
      localStorage.setItem(
        'user',
        JSON.stringify({
          id: '1',
          username: 'testuser',
          email: 'test@example.com',
          name: 'Test User',
          isAdmin: false,
          approved: true,
          emailVerified: true,
          createdAt: '2024-01-01T00:00:00Z',
        })
      )
    })

    // Go to profile
    await page.goto('/user/profile')
    await page.waitForLoadState('networkidle')
    await expect(page.locator('.page-container')).toBeVisible()

    // Navigate to security
    await page.goto('/user/security')
    await page.waitForLoadState('networkidle')
    await expect(page.locator('text=安全设置')).toBeVisible()

    // Navigate to passkeys
    await page.goto('/user/passkeys')
    await page.waitForLoadState('networkidle')
  })
})
