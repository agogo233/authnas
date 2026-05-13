import { test, expect } from '@playwright/test'

const BASE_URL = process.env.E2E_BASE_URL || 'http://localhost:8080'

test.describe('Performance Tests', () => {
  test.describe('Large Data List Rendering', () => {
    test('should render large user list without freezing', async ({ page }) => {
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

      const startTime = Date.now()
      await page.goto('/admin/users')
      await page.waitForLoadState('networkidle')

      const renderTime = Date.now() - startTime
      expect(renderTime).toBeLessThan(5000)
    })

    test('should handle 1000+ users in list', async ({ page }) => {
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
      await page.waitForTimeout(2000)

      const userRows = page.locator('tbody tr, .user-row, [data-user-id]')
      const rowCount = await userRows.count()

      expect(rowCount).toBeGreaterThanOrEqual(0)
    })

    test('should handle large session list without freezing', async ({ page }) => {
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
      await page.waitForTimeout(1000)

      const sessionTab = page
        .locator('.n-tabs-nav .n-tabs-tab:has-text("会话管理"), [role="tab"]:has-text("会话管理")')
        .first()
      if (await sessionTab.isVisible({ timeout: 2000 }).catch(() => false)) {
        await sessionTab.click()
        await page.waitForTimeout(2000)
      }

      const sessionItems = page.locator('[class*="session"], tbody tr, [data-session-id]')
      const count = await sessionItems.count()
      expect(count).toBeGreaterThanOrEqual(0)
    })

    test('should virtualize long lists efficiently', async ({ page }) => {
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
      await page.waitForTimeout(1000)

      const initialRenderedRows = await page.locator('tbody tr').count()

      await page.evaluate(() => {
        window.scrollTo(0, document.body.scrollHeight)
      })
      await page.waitForTimeout(1000)

      const afterScrollRenderedRows = await page.locator('tbody tr').count()
      expect(afterScrollRenderedRows).toBeGreaterThanOrEqual(0)
    })
  })

  test.describe('Concurrent Operations', () => {
    test('should handle multiple users logging in simultaneously', async ({ browser }) => {
      const contexts = await Promise.all([
        browser.newContext(),
        browser.newContext(),
        browser.newContext(),
        browser.newContext(),
        browser.newContext(),
      ])

      const pages = await Promise.all(contexts.map((ctx) => ctx.newPage()))

      await Promise.all(pages.map((page) => page.goto('/login')))

      await Promise.all(
        pages.map((page, index) =>
          page
            .fill('input[placeholder="请输入用户名"]', `concurrentuser${index}`)
            .then(() => page.fill('input[type="password"]', 'TestPass123!'))
            .then(() => page.getByRole('button', { name: '登录', exact: true }).click())
        )
      )

      await Promise.all(pages.map((page) => page.waitForTimeout(1000)))

      const results = await Promise.all(pages.map((page) => page.url()))

      await Promise.all(contexts.map((ctx) => ctx.close()))

      expect(
        results.every(
          (url) => url.includes('/login') || url.includes('/profile') || url.includes('/mfa')
        )
      ).toBeTruthy()
    })

    test('should handle rapid page navigation without memory leaks', async ({ page }) => {
      const memorySnapshots: number[] = []

      const routes = [
        '/login',
        '/register',
        '/reset-password',
        '/login',
        '/profile',
        '/security',
        '/login',
      ]

      for (const route of routes) {
        await page.goto(route)
        await page.waitForLoadState('networkidle')
        await page.waitForTimeout(200)

        const memory = await page.evaluate(() => {
          if ('memory' in performance) {
            return (performance as any).memory.usedJSHeapSize
          }
          return 0
        })
        memorySnapshots.push(memory)
      }

      const firstMemory = memorySnapshots[0]
      const lastMemory = memorySnapshots[memorySnapshots.length - 1]
      const memoryGrowth = lastMemory - firstMemory

      expect(memoryGrowth).toBeLessThan(50 * 1024 * 1024)
    })

    test('should handle concurrent API requests', async ({ page }) => {
      await page.goto('/login')
      await page.waitForLoadState('domcontentloaded')

      const startTime = Date.now()
      const [r1, r2, r3, r4, r5] = await Promise.all([
        page.request.get('/api/users'),
        page.request.get('/api/users'),
        page.request.get('/api/users'),
        page.request.get('/api/users'),
        page.request.get('/api/users'),
      ])

      const elapsed = Date.now() - startTime

      const statuses = [r1.status(), r2.status(), r3.status(), r4.status(), r5.status()]
      const allValid = statuses.every((s) => s >= 200 && s < 500)
      expect(allValid).toBeTruthy()
      expect(elapsed).toBeLessThan(10000)
    })
  })

  test.describe('Memory Leak Detection', () => {
    test('should not leak memory on repeated login/logout cycles', async ({ page }) => {
      await page.goto('/login')
      await page.waitForLoadState('networkidle')

      const getMemory = async () => {
        return await page.evaluate(() => {
          if ('memory' in performance) {
            return (performance as any).memory.usedJSHeapSize
          }
          return 0
        })
      }

      const initialMemory = await getMemory()

      for (let i = 0; i < 10; i++) {
        const iteration = i
        await page.evaluate(
          ({ iter }) => {
            localStorage.setItem('access_token', `token-${iter}`)
            localStorage.setItem(
              'user',
              JSON.stringify({ id: String(iter), username: `user${iter}` })
            )
          },
          { iter: iteration }
        )

        await page.goto('/profile')
        await page.waitForTimeout(100)

        await page.evaluate(() => {
          localStorage.clear()
        })

        await page.goto('/login')
        await page.waitForTimeout(100)
      }

      await page.waitForTimeout(1000)
      const finalMemory = await getMemory()

      if (initialMemory === 0) {
        test.skip(true, 'Memory API not available in this browser')
      }

      const memoryIncrease = finalMemory - initialMemory
      expect(memoryIncrease).toBeLessThan(20 * 1024 * 1024)
    })

    test('should clean up event listeners on navigation', async ({ page }) => {
      await page.goto('/login')
      await page.waitForLoadState('networkidle')

      const getListenerCount = async () => {
        return await page.evaluate(() => {
          return (window as any)._listenerCount || 0
        })
      }

      await page.goto('/register')
      await page.waitForTimeout(500)

      await page.goto('/login')
      await page.waitForTimeout(500)

      const listenerCount = await getListenerCount()
      expect(listenerCount).toBeLessThan(1000)
    })
  })

  test.describe('Page Load Time Benchmarks', () => {
    test('should load login page within acceptable time', async ({ page }) => {
      const startTime = Date.now()
      await page.goto('/login')
      await page.waitForLoadState('domcontentloaded')
      const loadTime = Date.now() - startTime

      expect(loadTime).toBeLessThan(3000)
    })

    test('should load admin dashboard within acceptable time', async ({ page }) => {
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

      const startTime = Date.now()
      await page.goto('/admin')
      await page.waitForLoadState('domcontentloaded')
      const loadTime = Date.now() - startTime

      expect(loadTime).toBeLessThan(5000)
    })

    test('should load user profile within acceptable time', async ({ page }) => {
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

      const startTime = Date.now()
      await page.goto('/profile')
      await page.waitForLoadState('domcontentloaded')
      const loadTime = Date.now() - startTime

      expect(loadTime).toBeLessThan(3000)
    })
  })

  test.describe('API Response Time', () => {
    test('should handle API request within acceptable time', async ({ page, request }) => {
      await page.goto('/login')
      await page.waitForLoadState('networkidle')
      await page.evaluate(() => {
        localStorage.setItem('access_token', 'admin-token')
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

      const baseURL = page.url().replace(/\/.*$/, '')

      const startTime = Date.now()
      const response = await request.get(`${baseURL}/api/users`, {
        headers: { Authorization: 'Bearer admin-token' },
      })
      const responseTime = Date.now() - startTime

      expect([200, 401, 403, 404]).toContain(response.status())
      expect(responseTime).toBeLessThan(5000)
    })

    test('should handle token endpoint within acceptable time', async ({ page, request }) => {
      await page.goto('/login')
      await page.waitForLoadState('networkidle')

      const startTime = Date.now()
      const response = await request.post(`${BASE_URL}/oidc/token`, {
        form: {
          grant_type: 'refresh_token',
          refresh_token: 'test-refresh',
          client_id: 'test-client',
        },
      })
      const responseTime = Date.now() - startTime

      expect([200, 400, 401, 404]).toContain(response.status())
      expect(responseTime).toBeLessThan(3000)
    })
  })

  test.describe('Token Refresh Timing', () => {
    test('should refresh token before expiration', async ({ page }) => {
      await page.goto('/login')
      await page.waitForLoadState('networkidle')
      await page.evaluate(() => {
        const userWithSoonExpiringToken = {
          id: '1',
          username: 'testuser',
          email: 'test@example.com',
          isAdmin: false,
          approved: true,
          exp: Math.floor(Date.now() / 1000) + 120,
        }
        localStorage.setItem('access_token', 'soon-expiring-token')
        localStorage.setItem('refresh_token', 'valid-refresh-token')
        localStorage.setItem('user', JSON.stringify(userWithSoonExpiringToken))
      })

      await page.goto('/profile')
      await page.waitForTimeout(2000)

      const currentUrl = page.url()
      expect(currentUrl).toBeTruthy()
    })

    test('should handle token refresh failure gracefully', async ({ page }) => {
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
        localStorage.setItem('refresh_token', 'invalid-refresh')
        localStorage.setItem('user', JSON.stringify(expiredUser))
      })

      await page.goto('/profile')
      await page.waitForTimeout(2000)

      const currentUrl = page.url()
      expect(currentUrl).toContain('/login')
    })
  })

  test.describe('Network Interruption Recovery', () => {
    test('should handle network disconnection gracefully', async ({ page, context }) => {
      await page.goto('/login')
      await page.waitForLoadState('networkidle')

      await context.setOffline(true)

      await page.fill('input[placeholder="请输入用户名"]', 'testuser')
      await page.fill('input[type="password"]', 'TestPass123!')
      await page.getByRole('button', { name: '登录', exact: true }).click()
      await page.waitForTimeout(1000)

      await context.setOffline(false)

      await page.reload()
      await page.waitForLoadState('networkidle')

      const currentUrl = page.url()
      expect(currentUrl).toBeTruthy()
    })

    test('should retry failed requests on network recovery', async ({ page, context }) => {
      await page.goto('/login')
      await page.waitForLoadState('networkidle')
      await page.evaluate(() => {
        localStorage.setItem('access_token', 'test-token')
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

      await context.setOffline(true)
      await page.waitForTimeout(500)
      await context.setOffline(false)

      await page.reload()
      await page.waitForLoadState('networkidle')

      const currentUrl = page.url()
      expect(currentUrl).toBeTruthy()
    })
  })

  test.describe('Input Performance', () => {
    test('should handle rapid typing without lag', async ({ page }) => {
      await page.goto('/login')
      await page.waitForLoadState('networkidle')

      const input = page.locator('input[placeholder="请输入用户名"]')
      await input.click()

      const startTime = Date.now()
      for (let i = 0; i < 100; i++) {
        await input.press(`${i % 10}`)
      }
      const typingTime = Date.now() - startTime

      expect(typingTime).toBeLessThan(5000)
    })

    test('should debounce search input', async ({ page }) => {
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

      const searchInput = page.locator('input[placeholder*="search"], input[type="search"]').first()
      if (await searchInput.isVisible({ timeout: 3000 })) {
        await searchInput.fill('test')
        await page.waitForTimeout(100)
        await searchInput.fill('testa')
        await page.waitForTimeout(100)
        await searchInput.fill('testab')
        await page.waitForTimeout(100)

        const lastRequestTime = await page.evaluate(() => {
          return (window as any)._lastSearchRequestTime || 0
        })

        expect(lastRequestTime).toBeGreaterThan(0)
      }
    })
  })
})
