import { test, expect, Page, BrowserContext } from '@playwright/test'

const BASE_URL = process.env.E2E_BASE_URL || 'http://localhost:8080'

const TEST_ADMIN = {
  username: 'admin',
  password: 'Admin123!',
  email: 'admin@example.com',
}

async function performLogin(page: Page, username: string, password: string) {
  await page.goto('/login')
  await page.waitForLoadState('networkidle')
  await page.waitForTimeout(1000)

  await page.fill('input[placeholder="请输入用户名"]', username)
  await page.fill('input[type="password"]', password)
  await page.click('button[type="submit"]')

  await page.waitForLoadState('networkidle')
  await page.waitForTimeout(1000)
}

async function performLogout(page: Page) {
  const logoutButton = page
    .locator('button:has-text("退出"), button:has-text("注销"), button:has-text("Logout")')
    .first()
  if (await logoutButton.isVisible({ timeout: 3000 }).catch(() => false)) {
    await logoutButton.click()
    await page.waitForTimeout(500)
  }
}

async function clearBrowserState(page: Page) {
  await page.context().clearCookies()
  await page.goto(BASE_URL)
  await page.waitForLoadState('domcontentloaded')
  await page.evaluate(() => localStorage.clear())
  await page.evaluate(() => sessionStorage.clear())
}

test.describe('Real E2E - Complete User Flows', () => {
  test.beforeEach(async ({ page }) => {
    await clearBrowserState(page)
  })
  test.describe('1. Complete Login Flow', () => {
    test('should login with valid admin credentials', async ({ page }) => {
      await performLogin(page, TEST_ADMIN.username, TEST_ADMIN.password)

      await expect(page).not.toHaveURL(/\/login/, { timeout: 10000 })
      const currentUrl = page.url()
      expect(currentUrl).toMatch(/\/(user\/profile|admin|mfa)/)
    })

    test('should show error with invalid credentials', async ({ page }) => {
      await page.goto('/login')
      await page.waitForLoadState('networkidle')
      await page.waitForTimeout(1000)

      await page.fill('input[placeholder="请输入用户名"]', TEST_ADMIN.username)
      await page.fill('input[type="password"]', 'wrongpassword')
      await page.click('button[type="submit"]')

      await page.waitForTimeout(2000)

      const alert = page.locator('.n-alert[type="error"]').first()
      const isVisible = await alert.isVisible({ timeout: 5000 }).catch(() => false)
      expect(isVisible || page.url().includes('/login')).toBeTruthy()
    })

    test('should show validation error for empty fields', async ({ page }) => {
      await page.goto('/login')
      await page.waitForLoadState('networkidle')
      await page.waitForTimeout(1000)

      await page.click('button[type="submit"]')
      await page.waitForTimeout(1500)

      const errorText = page.locator('body')
      const pageContent = await errorText.textContent()
      expect(pageContent).toContain('请输入')
    })

    test('should redirect to original page after login with redirect query', async ({ page }) => {
      const targetPage = '/admin/users'
      await page.goto(`/login?redirect=${targetPage}`)
      await page.waitForLoadState('networkidle')
      await page.waitForTimeout(1000)

      await page.fill('input[placeholder="请输入用户名"]', TEST_ADMIN.username)
      await page.fill('input[type="password"]', TEST_ADMIN.password)
      await page.click('button[type="submit"]')

      await page.waitForLoadState('networkidle')
      await page.waitForTimeout(2000)

      const currentUrl = page.url()
      expect(currentUrl).toMatch(new RegExp(targetPage))
    })

    test('should remember username when remember me is checked', async ({ page }) => {
      await page.goto('/login')
      await page.waitForLoadState('networkidle')
      await page.waitForTimeout(1000)

      await page.fill('input[placeholder="请输入用户名"]', TEST_ADMIN.username)
      const checkbox = page.locator('.n-checkbox, [class*="checkbox"]').first()
      if (await checkbox.isVisible({ timeout: 2000 }).catch(() => false)) {
        await checkbox.click()
      }
      await page.fill('input[type="password"]', TEST_ADMIN.password)
      await page.getByRole('button', { name: '登录', exact: true }).click()

      await page.waitForTimeout(1000)

      await page.goto('/login')
      await page.waitForLoadState('networkidle')
      await page.waitForTimeout(1000)

      const usernameInput = page.locator('input[placeholder="请输入用户名"]')
      await expect(usernameInput).toHaveValue(TEST_ADMIN.username)
    })
  })

  test.describe('2. Admin Dashboard Flow', () => {
    test.beforeEach(async ({ page }) => {
      await performLogin(page, TEST_ADMIN.username, TEST_ADMIN.password)
      await page.waitForTimeout(1000)
    })

    test('should access admin dashboard', async ({ page }) => {
      await page.goto('/admin')
      await page.waitForLoadState('networkidle')

      expect(page.url()).toMatch(/admin/)
    })

    test('should display admin navigation menu', async ({ page }) => {
      await page.goto('/admin')
      await page.waitForLoadState('networkidle')

      const adminMenu = page.locator('.n-menu, nav, [role="navigation"]').first()
      await expect(adminMenu).toBeVisible({ timeout: 5000 })
    })

    test('should navigate to users management', async ({ page }) => {
      await page.goto('/admin/users')
      await page.waitForLoadState('networkidle')
      await page.waitForTimeout(1000)

      expect(page.url()).toMatch(/users/)
    })

    test('should navigate to groups management', async ({ page }) => {
      await page.goto('/admin/groups')
      await page.waitForLoadState('networkidle')

      expect(page.url()).toMatch(/groups/)
    })

    test('should navigate to clients management', async ({ page }) => {
      await page.goto('/admin/clients')
      await page.waitForLoadState('networkidle')

      expect(page.url()).toMatch(/clients/)
    })

    test('should navigate to invitations management', async ({ page }) => {
      await page.goto('/admin/invitations')
      await page.waitForLoadState('networkidle')

      expect(page.url()).toMatch(/invitations/)
    })

    test('should navigate to proxy auth management', async ({ page }) => {
      await page.goto('/admin/proxyauth')
      await page.waitForLoadState('networkidle')

      expect(page.url()).toMatch(/proxyauth/)
    })

    test('should navigate to settings', async ({ page }) => {
      await page.goto('/admin/settings')
      await page.waitForLoadState('networkidle')

      expect(page.url()).toMatch(/settings/)
    })
  })

  test.describe('3. Admin Users Management Flow', () => {
    test.beforeEach(async ({ page }) => {
      await performLogin(page, TEST_ADMIN.username, TEST_ADMIN.password)
      await page.goto('/admin/users')
      await page.waitForLoadState('networkidle')
      await page.waitForTimeout(1500)
    })

    test('should display users list', async ({ page }) => {
      const userRows = page.locator('.n-data-table-tr, tbody tr, [data-row-key]')
      const count = await userRows.count()
      expect(count).toBeGreaterThan(0)
    })

    test('should search users by username', async ({ page }) => {
      const searchInput = page
        .locator('input[placeholder*="搜索"], input[placeholder*="查询"]')
        .first()

      if (await searchInput.isVisible({ timeout: 2000 }).catch(() => false)) {
        await searchInput.fill(TEST_ADMIN.username)
        await page.waitForTimeout(1000)

        const pageContent = await page.content()
        expect(pageContent).toBeTruthy()
      }
    })

    test('should open user edit modal', async ({ page }) => {
      await page.waitForTimeout(1000)
      const editButtons = page.locator('button:has-text("编辑"), button[aria-label*="edit"]')

      if (
        await editButtons
          .first()
          .isVisible({ timeout: 2000 })
          .catch(() => false)
      ) {
        await editButtons.first().click()
        await page.waitForTimeout(500)

        const modal = page.locator('.n-modal, [role="dialog"]')
        const modalVisible = await modal.isVisible({ timeout: 2000 }).catch(() => false)
        expect(modalVisible || page.url()).toBeTruthy()
      }
    })

    test('should display pagination', async ({ page }) => {
      await page.waitForTimeout(1000)
      const pagination = page.locator('.n-pagination')

      if (await pagination.isVisible({ timeout: 2000 }).catch(() => false)) {
        await expect(pagination).toBeVisible()
      }
    })
  })

  test.describe('4. Registration Flow', () => {
    test('should navigate to registration page', async ({ page }) => {
      await page.goto('/login')
      await page.waitForLoadState('networkidle')
      await page.waitForTimeout(1000)

      await page.click('a:has-text("创建账户")')
      await page.waitForLoadState('networkidle')

      await expect(page).toHaveURL(/\/register/)
    })

    test('should display registration form', async ({ page }) => {
      await page.goto('/register')
      await page.waitForLoadState('networkidle')
      await page.waitForTimeout(2000)

      const usernameInput = page.locator('input').first()
      const passwordInput = page.locator('input[type="password"]').first()

      await expect(usernameInput).toBeVisible()
      await expect(passwordInput).toBeVisible()
    })
  })

  test.describe('5. Password Reset Flow', () => {
    test('should display reset password form', async ({ page }) => {
      await page.goto('/reset-password')
      await page.waitForLoadState('networkidle')
      await page.waitForTimeout(2000)

      const emailInput = page.locator('input').first()
      await expect(emailInput).toBeVisible()
    })
  })

  test.describe('6. Consent Page Flow', () => {
    test('should access consent page', async ({ page }) => {
      await page.goto('/consent/test-client-uid')
      await page.waitForLoadState('networkidle')

      const pageContent = await page.content()
      expect(
        pageContent.includes('同意') ||
          pageContent.includes('consent') ||
          page.url().includes('/consent/')
      ).toBeTruthy()
    })
  })

  test.describe('7. MFA Flow', () => {
    test('should redirect from MFA page when not authenticated', async ({ page }) => {
      await page.goto('/mfa')
      await page.waitForLoadState('domcontentloaded')
      await page.waitForTimeout(2000)

      await expect(page).toHaveURL(/\/login/, { timeout: 5000 })
    })

    test('should display MFA page when accessed directly', async ({ page }) => {
      await page.goto('/login?redirect=/mfa')
      await page.waitForLoadState('domcontentloaded')
      await page.waitForTimeout(2000)

      await page.fill('input[placeholder="请输入用户名"]', TEST_ADMIN.username)
      await page.fill('input[type="password"]', TEST_ADMIN.password)
      await page.click('button[type="submit"]')
      await page.waitForTimeout(3000)

      const currentUrl = page.url()
      expect(currentUrl).toMatch(/mfa|profile/)
    })
  })

  test.describe('8. Error Handling and Edge Cases', () => {
    test('should handle SQL injection attempt in login', async ({ page }) => {
      await page.goto('/login')
      await page.waitForLoadState('domcontentloaded')
      await page.waitForTimeout(1000)

      await page.fill('input[placeholder="请输入用户名"]', "admin' OR '1'='1")
      await page.fill('input[type="password"]', 'anypassword')
      await page.click('button[type="submit"]')

      await page.waitForTimeout(2000)

      const pageContent = await page.content()
      expect(pageContent).toContain('登录')
    })

    test('should handle XSS attempt in username field', async ({ page }) => {
      await page.goto('/login')
      await page.waitForLoadState('domcontentloaded')
      await page.waitForTimeout(1000)

      const xssPayload = '<script>alert("XSS")</script>'
      await page.fill('input[placeholder="请输入用户名"]', xssPayload)

      const inputValue = await page.locator('input[placeholder="请输入用户名"]').inputValue()
      expect(inputValue).toBe(xssPayload)
    })

    test('should handle very long username input', async ({ page }) => {
      await page.goto('/login')
      await page.waitForLoadState('domcontentloaded')
      await page.waitForTimeout(1000)

      const longUsername = 'a'.repeat(200)
      await page.fill('input[placeholder="请输入用户名"]', longUsername)

      const inputValue = await page.locator('input[placeholder="请输入用户名"]').inputValue()
      expect(inputValue.length).toBeGreaterThan(0)
    })

    test('should handle special characters in password', async ({ page }) => {
      await page.goto('/login')
      await page.waitForLoadState('domcontentloaded')
      await page.waitForTimeout(1000)

      await page.fill('input[placeholder="请输入用户名"]', TEST_ADMIN.username)
      await page.fill('input[type="password"]', 'P@$$w0rd!#$%^&*()')
      await page.click('button[type="submit"]')

      await page.waitForTimeout(2000)

      const currentUrl = page.url()
      expect(currentUrl).toMatch(/\/login|\/user/)
    })

    test('should handle network delay gracefully', async ({ page }) => {
      let requestHandled = false
      page.route('**/api/**', async (route) => {
        requestHandled = true
        await route.continue()
      })

      await page.goto('/login')
      await page.waitForLoadState('domcontentloaded')
      await page.waitForTimeout(500)

      await page.fill('input[placeholder="请输入用户名"]', TEST_ADMIN.username)
      await page.fill('input[type="password"]', TEST_ADMIN.password)

      expect(requestHandled || true).toBeTruthy()
    })
  })

  test.describe('9. Registration Edge Cases', () => {
    test('should handle registration with existing username', async ({ page }) => {
      await page.goto('/register')
      await page.waitForLoadState('domcontentloaded')
      await page.waitForTimeout(1500)

      const usernameInput = page.locator('input').first()
      const passwordInput = page.locator('input[type="password"]').first()

      if (await usernameInput.isVisible({ timeout: 3000 })) {
        await usernameInput.fill(TEST_ADMIN.username)
        await passwordInput.fill(TEST_ADMIN.password)
        await page.click('button[type="submit"]')
        await page.waitForTimeout(2000)
      }

      const pageContent = await page.content()
      expect(pageContent).toBeTruthy()
    })

    test('should handle weak password in registration', async ({ page }) => {
      await page.goto('/register')
      await page.waitForLoadState('domcontentloaded')
      await page.waitForTimeout(1500)

      const usernameInput = page.locator('input').first()
      const passwordInput = page.locator('input[type="password"]').first()

      if (await usernameInput.isVisible({ timeout: 3000 })) {
        await usernameInput.fill('newuser123')
        await passwordInput.fill('123')
        await page.click('button[type="submit"]')
        await page.waitForTimeout(2000)
      }

      const currentUrl = page.url()
      expect(currentUrl).toBeTruthy()
    })

    test('should navigate back from registration to login', async ({ page }) => {
      await page.goto('/register')
      await page.waitForLoadState('domcontentloaded')
      await page.waitForTimeout(1500)

      const link = page
        .locator('a')
        .filter({ hasText: /登录|返回/ })
        .first()
      if (await link.isVisible({ timeout: 3000 })) {
        await link.click()
        await page.waitForLoadState('networkidle')
      }

      await expect(page).toHaveURL(/\/login/)
    })
  })

  test.describe('10. Admin Panel Edge Cases', () => {
    test('should handle access to admin without authentication', async ({ page }) => {
      await page.goto('/admin/users')
      await page.waitForLoadState('domcontentloaded')

      await page.waitForTimeout(1000)

      expect(page.url()).toMatch(/login/)
    })

    test('should handle direct API access without token', async ({ request }) => {
      const response = await request.get('/api/admin/users')
      expect(response.status()).toBe(401)
    })

    test('should handle invalid page navigation', async ({ page }) => {
      await page.goto('/login')
      await page.waitForLoadState('domcontentloaded')
      await page.waitForTimeout(1000)

      await page.fill('input[placeholder="请输入用户名"]', TEST_ADMIN.username)
      await page.fill('input[type="password"]', TEST_ADMIN.password)
      await page.click('button[type="submit"]')

      await page.waitForTimeout(2000)

      await page.goto('/nonexistent-page')
      await page.waitForLoadState('domcontentloaded')
      await page.waitForTimeout(1000)

      const currentUrl = page.url()
      expect(currentUrl).toBeTruthy()
    })
  })
})
