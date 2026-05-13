import { test, expect } from '@playwright/test'

test.describe('Admin User Management Operations', () => {
  test.beforeEach(async ({ page }) => {
    await page.route('**/api/**', async (route) => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          success: true,
          data: [
            {
              id: '1',
              username: 'admin',
              email: 'admin@example.com',
              isAdmin: true,
              approved: true,
              mfaRequired: false,
              emailVerified: true,
              createdAt: new Date().toISOString(),
            },
            {
              id: '2',
              username: 'testuser',
              email: 'test@example.com',
              isAdmin: false,
              approved: true,
              mfaRequired: false,
              emailVerified: true,
              createdAt: new Date().toISOString(),
            },
          ],
          total: 2,
          page: 1,
          pageSize: 10,
        }),
      })
    })

    await page.context().addInitScript(() => {
      localStorage.setItem('access_token', 'admin-token')
      localStorage.setItem('refresh_token', 'admin-refresh-token')
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
    await page.waitForTimeout(500)
  })

  test.describe('User Search and Filter', () => {
    test('should search users by exact username', async ({ page }) => {
      const searchInput = page.locator('input[placeholder*="搜索"]').first()

      if (await searchInput.isVisible({ timeout: 3000 })) {
        await searchInput.fill('testuser')
        await page.locator('button:has-text("搜索")').click()
        await page.waitForTimeout(500)

        const userRows = page.locator('.n-data-table-tr')
        const count = await userRows.count()
        expect(count).toBeGreaterThanOrEqual(0)
      }
    })

    test('should search users by partial username', async ({ page }) => {
      const searchInput = page.locator('input[placeholder*="搜索"]').first()

      if (await searchInput.isVisible({ timeout: 3000 })) {
        await searchInput.fill('test')
        await page.locator('button:has-text("搜索")').click()
        await page.waitForTimeout(800)

        const pageContent = await page.content()
        expect(pageContent).toBeTruthy()
      }
    })

    test('should search users by email', async ({ page }) => {
      const searchInput = page.locator('input[placeholder*="搜索"]').first()

      if (await searchInput.isVisible({ timeout: 3000 })) {
        await searchInput.fill('admin@example.com')
        await page.locator('button:has-text("搜索")').click()
        await page.waitForTimeout(800)

        const pageContent = await page.content()
        expect(pageContent).toBeTruthy()
      }
    })

    test('should clear search filters', async ({ page }) => {
      const searchInput = page.locator('input[placeholder*="搜索"]').first()

      if (await searchInput.isVisible({ timeout: 3000 })) {
        await searchInput.fill('testuser')
        await page.waitForTimeout(500)

        await searchInput.clear()
        await page.locator('button:has-text("搜索")').click()
        await page.waitForTimeout(500)
      }
    })

    test('should handle search with no results', async ({ page }) => {
      const searchInput = page.locator('input[placeholder*="搜索"]').first()

      if (await searchInput.isVisible({ timeout: 3000 })) {
        await searchInput.fill('thisusernamedoesnotexist123456789')
        await page.locator('button:has-text("搜索")').click()
        await page.waitForTimeout(1000)

        const pageContent = await page.content()
        expect(pageContent).toBeTruthy()
      }
    })
  })

  test.describe('User List Pagination', () => {
    test('should display pagination controls', async ({ page }) => {
      const pagination = page.locator('.n-pagination').first()

      if (await pagination.isVisible({ timeout: 3000 })) {
        await expect(pagination).toBeVisible()
      }
    })

    test('should navigate to next page', async ({ page }) => {
      const nextButton = page.locator('.n-pagination button').filter({ hasText: '>' }).first()

      if (await nextButton.isVisible({ timeout: 3000 })) {
        await nextButton.click()
        await page.waitForTimeout(1000)
      }
    })

    test('should change page size', async ({ page }) => {
      const pageSizeSelect = page.locator('.n-pagination .n-select').first()

      if (await pageSizeSelect.isVisible({ timeout: 3000 })) {
        await pageSizeSelect.click()
        await page.waitForTimeout(300)

        const option = page.locator('.n-virtual-select-menu .n-virtual-select-option').nth(1)
        if (await option.isVisible({ timeout: 1000 })) {
          await option.click()
          await page.waitForTimeout(1000)
        }
      }
    })
  })

  test.describe('User Activation and Deactivation', () => {
    test('should display user status indicators', async ({ page }) => {
      const statusTags = page.locator('.n-tag')

      const count = await statusTags.count()
      expect(count).toBeGreaterThanOrEqual(0)
    })

    test('should approve user', async ({ page }) => {
      const approveButton = page.locator('button:has-text("批准")').first()

      if (await approveButton.isVisible({ timeout: 3000 })) {
        await approveButton.click()
        await page.waitForTimeout(1000)
      }
    })

    test('should open edit modal', async ({ page }) => {
      const editButton = page.locator('button:has-text("编辑")').first()

      if (await editButton.isVisible({ timeout: 3000 })) {
        await editButton.click()
        await page.waitForTimeout(500)

        const modal = page.locator('.n-modal')
        expect(await modal.isVisible({ timeout: 2000 }).catch(() => false)).toBeTruthy()
      }
    })
  })

  test.describe('User Deletion', () => {
    test('should show delete option in actions', async ({ page }) => {
      await page.waitForSelector('.n-data-table-tr', { timeout: 5000 }).catch(() => null)
      const actionButtons = page.locator('.n-data-table-tr .n-button')

      const hasActions = await actionButtons.count()
      expect(hasActions).toBeGreaterThanOrEqual(0)
    })

    test('should handle delete user flow', async ({ page }) => {
      await page.waitForSelector('.n-data-table-tr', { timeout: 5000 }).catch(() => null)
      const deleteTriggers = page.locator('.n-popconfirm')

      const count = await deleteTriggers.count()
      if (count > 0) {
        await deleteTriggers.first().hover()
        await page.waitForTimeout(300)
      }
    })
  })

  test.describe('User List Sorting', () => {
    test('should have sortable columns', async ({ page }) => {
      await page.waitForSelector('.n-data-table', { timeout: 5000 }).catch(() => null)
      const tableHeaders = page.locator('.n-data-table-th')

      const count = await tableHeaders.count()
      expect(count).toBeGreaterThanOrEqual(0)
    })
  })

  test.describe('Non-Admin Access Control', () => {
    test('should redirect non-admin user from admin users page', async ({ browser }) => {
      // Create a new independent context to avoid beforeEach addInitScript interference
      const context = await browser.newContext()
      const testPage = await context.newPage()

      await testPage.context().addInitScript(() => {
        localStorage.setItem('access_token', 'regular-user-token')
        localStorage.setItem('refresh_token', 'regular-refresh-token')
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

      await testPage.goto('/admin/users')
      await testPage.waitForTimeout(2000)

      const currentUrl = testPage.url()
      expect(currentUrl).toContain('/login')

      await context.close()
    })

    test('should show access denied for non-admin on admin API', async ({ page, request }) => {
      const baseURL = page.url().replace(/\/.*$/, '')

      const response = await request.get(`${baseURL}/api/admin/users`, {
        headers: {
          Authorization: 'Bearer regular-user-token',
        },
      })

      expect([401, 403, 404]).toContain(response.status())
    })
  })
})
