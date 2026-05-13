import { test, expect } from '@playwright/test'

const TEST_ADMIN = {
  username: 'admin',
  password: 'Admin123!',
}

async function performLogin(page: Page, username: string, password: string) {
  await page.goto('/login')
  await page.waitForLoadState('networkidle')
  await page.waitForTimeout(500)

  await page.fill('input[placeholder="请输入用户名"]', username)
  await page.fill('input[type="password"]', password)
  await page.click('button[type="submit"]')

  await page.waitForLoadState('networkidle')
  await page.waitForTimeout(1000)
}

type Page = import('@playwright/test').Page

test.describe('Security Settings Page E2E', () => {
  test.beforeEach(async ({ page }) => {
    await performLogin(page, TEST_ADMIN.username, TEST_ADMIN.password)
    await page.goto('/user/security')
    await page.waitForLoadState('networkidle')
    await page.waitForTimeout(1000)
  })

  test('should display security settings page', async ({ page }) => {
    await expect(page.getByText('安全设置').first()).toBeVisible({ timeout: 5000 })
  })

  test('should have password change tab', async ({ page }) => {
    await expect(page.locator('.n-tabs-tab').filter({ hasText: '修改密码' }).first()).toBeVisible()
  })

  test('should have MFA tab', async ({ page }) => {
    await expect(page.locator('.n-tabs-tab').filter({ hasText: '两步验证' }).first()).toBeVisible()
  })

  test('should have Passkeys tab', async ({ page }) => {
    await expect(page.locator('.n-tabs-tab').filter({ hasText: '通行密钥' }).first()).toBeVisible()
  })

  test('should have Sessions tab', async ({ page }) => {
    await expect(page.locator('.n-tabs-tab').filter({ hasText: '会话管理' }).first()).toBeVisible()
  })

  test('should show password change form in first tab', async ({ page }) => {
    await expect(page.locator('input[placeholder*="请输入当前密码"]')).toBeVisible()
    await expect(page.locator('input[placeholder*="请输入新密码"]')).toBeVisible()
    await expect(page.locator('input[placeholder*="请再次输入"]')).toBeVisible()
  })

  test('should show validation error for empty password fields', async ({ page }) => {
    await page.getByRole('button', { name: '更新密码', exact: true }).click()
    await expect(page.getByText('请填写所有字段')).toBeVisible({ timeout: 5000 })
  })

  test('should show validation error for password mismatch', async ({ page }) => {
    await page.fill('input[placeholder*="请输入当前密码"]', 'oldpass123')
    await page.fill('input[placeholder*="请输入新密码"]', 'NewPass123!')
    await page.fill('input[placeholder*="请再次输入"]', 'DifferentPass123!')

    await page.getByRole('button', { name: '更新密码', exact: true }).click()
    await expect(page.getByText('两次输入的密码不一致')).toBeVisible({ timeout: 5000 })
  })

  test('should show password strength indicator when typing new password', async ({ page }) => {
    await page.fill('input[placeholder*="请输入新密码"]', 'abc')

    const strengthIndicator = page.locator('.password-strength .strength-label')
    await expect(strengthIndicator).toBeVisible()
  })

  test('should switch to MFA tab', async ({ page }) => {
    await page.locator('.n-tabs-tab').filter({ hasText: '两步验证' }).click()
    await page.waitForTimeout(500)

    await expect(page.getByText('TOTP 目前未启用，启用后可添加额外的安全保护')).toBeVisible()
  })

  test('should switch to Passkeys tab', async ({ page }) => {
    await page.locator('.n-tabs-tab').filter({ hasText: '通行密钥' }).click()
    await page.waitForTimeout(500)

    await expect(page.getByText('通行密钥允许您使用 WebAuthn 安全地无密码登录')).toBeVisible()
  })

  test('should switch to Sessions tab', async ({ page }) => {
    await page.locator('.n-tabs-tab').filter({ hasText: '会话管理' }).click()
    await page.waitForTimeout(500)

    await expect(page.getByText('以下是您的所有活动会话列表')).toBeVisible()
  })

  test('should show revoke all sessions button', async ({ page }) => {
    await page.locator('.n-tabs-tab').filter({ hasText: '会话管理' }).click()
    await page.waitForTimeout(500)

    await expect(page.getByRole('button', { name: '撤销所有会话' })).toBeVisible()
  })

  test('should show warning before revoking sessions', async ({ page }) => {
    await page.locator('.n-tabs-tab').filter({ hasText: '会话管理' }).click()
    await page.waitForTimeout(500)

    await expect(page.getByRole('alert').first()).toBeVisible()
  })
})
