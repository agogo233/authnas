import { describe, it, expect, vi, beforeEach } from 'vitest'
import { authApi, userApi, passkeyApi, totpApi } from '../auth'

vi.mock('@/api/client', () => ({
  default: {
    get: vi.fn(),
    post: vi.fn(),
    put: vi.fn(),
    delete: vi.fn(),
  },
}))

describe('authApi', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  describe('getCsrfToken', () => {
    it('should call GET /auth/csrf', async () => {
      const { default: apiClient } = await import('@/api/client')
      vi.mocked(apiClient.get).mockResolvedValue({
        data: { success: true, data: { token: 'csrf-token', expiresAt: '2024-12-31' } },
      })

      const result = await authApi.getCsrfToken()

      expect(apiClient.get).toHaveBeenCalledWith('/auth/csrf')
      expect(result.data.data.token).toBe('csrf-token')
    })
  })

  describe('login', () => {
    it('should call POST /auth/login with correct data', async () => {
      const { default: apiClient } = await import('@/api/client')
      vi.mocked(apiClient.post).mockResolvedValue({
        data: {
          success: true,
          data: {
            accessToken: 'access-token',
            refreshToken: 'refresh-token',
            expiresAt: '2024-12-31',
            user: {
              id: '1',
              username: 'test',
              email: 'test@test.com',
              emailVerified: true,
              approved: true,
              createdAt: '2024-01-01',
            },
          },
        },
      })

      const result = await authApi.login({
        input: 'testuser',
        password: 'password123',
        remember: true,
      })

      expect(apiClient.post).toHaveBeenCalledWith('/auth/login', {
        input: 'testuser',
        password: 'password123',
        remember: true,
      })
      expect(result.data.data.accessToken).toBe('access-token')
    })
  })

  describe('register', () => {
    it('should call POST /auth/register with correct data', async () => {
      const { default: apiClient } = await import('@/api/client')
      vi.mocked(apiClient.post).mockResolvedValue({
        data: { success: true, data: { accessToken: 'token', refreshToken: 'refresh' } },
      })

      const result = await authApi.register({
        email: 'test@example.com',
        username: 'testuser',
        password: 'Password123!',
        inviteId: 'invite-123',
      })

      expect(apiClient.post).toHaveBeenCalledWith('/auth/register', {
        email: 'test@example.com',
        username: 'testuser',
        password: 'Password123!',
        inviteId: 'invite-123',
      })
      expect(result.data.success).toBe(true)
    })

    it('should register without inviteId', async () => {
      const { default: apiClient } = await import('@/api/client')
      vi.mocked(apiClient.post).mockResolvedValue({
        data: { success: true, data: { accessToken: 'token', refreshToken: 'refresh' } },
      })

      await authApi.register({
        email: 'test@example.com',
        username: 'testuser',
        password: 'Password123!',
      })

      expect(apiClient.post).toHaveBeenCalledWith('/auth/register', {
        email: 'test@example.com',
        username: 'testuser',
        password: 'Password123!',
      })
    })
  })

  describe('passkeyStart', () => {
    it('should call POST /auth/passkey/start', async () => {
      const { default: apiClient } = await import('@/api/client')
      vi.mocked(apiClient.post).mockResolvedValue({
        data: { success: true, data: { challenge: 'challenge', options: '{}' } },
      })

      const result = await authApi.passkeyStart({ username: 'testuser' })

      expect(apiClient.post).toHaveBeenCalledWith('/auth/passkey/start', { username: 'testuser' })
      expect(result.data.data.challenge).toBe('challenge')
    })
  })

  describe('passkeyEnd', () => {
    it('should call POST /auth/passkey/end', async () => {
      const { default: apiClient } = await import('@/api/client')
      vi.mocked(apiClient.post).mockResolvedValue({
        data: { success: true, data: { accessToken: 'token', refreshToken: 'refresh' } },
      })

      const result = await authApi.passkeyEnd({
        credentialId: 'cred-123',
        challenge: 'challenge',
        response: 'response',
      })

      expect(apiClient.post).toHaveBeenCalledWith('/auth/passkey/end', {
        credentialId: 'cred-123',
        challenge: 'challenge',
        response: 'response',
      })
      expect(result.data.success).toBe(true)
    })
  })

  describe('totpVerify', () => {
    it('should call POST /auth/totp', async () => {
      const { default: apiClient } = await import('@/api/client')
      vi.mocked(apiClient.post).mockResolvedValue({
        data: { success: true, data: { accessToken: 'token', refreshToken: 'refresh' } },
      })

      const result = await authApi.totpVerify({ token: '123456', mfaToken: 'mfa-token' })

      expect(apiClient.post).toHaveBeenCalledWith('/auth/totp', {
        token: '123456',
        mfaToken: 'mfa-token',
      })
      expect(result.data.success).toBe(true)
    })
  })

  describe('verifyEmail', () => {
    it('should call POST /auth/verify_email', async () => {
      const { default: apiClient } = await import('@/api/client')
      vi.mocked(apiClient.post).mockResolvedValue({
        data: { success: true },
      })

      const result = await authApi.verifyEmail({ userId: 'user-123', challenge: 'challenge' })

      expect(apiClient.post).toHaveBeenCalledWith('/auth/verify_email', {
        userId: 'user-123',
        challenge: 'challenge',
      })
      expect(result.data.success).toBe(true)
    })
  })

  describe('sendVerifyEmail', () => {
    it('should call POST /auth/send_verify_email', async () => {
      const { default: apiClient } = await import('@/api/client')
      vi.mocked(apiClient.post).mockResolvedValue({
        data: { success: true },
      })

      const result = await authApi.sendVerifyEmail({ email: 'test@example.com' })

      expect(apiClient.post).toHaveBeenCalledWith('/auth/send_verify_email', {
        email: 'test@example.com',
      })
      expect(result.data.success).toBe(true)
    })
  })

  describe('getInvitation', () => {
    it('should call GET /auth/invitation/:id/:challenge', async () => {
      const { default: apiClient } = await import('@/api/client')
      vi.mocked(apiClient.get).mockResolvedValue({
        data: { success: true, data: { email: 'test@example.com', username: 'testuser' } },
      })

      const result = await authApi.getInvitation('invite-123', 'challenge')

      expect(apiClient.get).toHaveBeenCalledWith('/auth/invitation/invite-123/challenge')
      expect(result.data.data.email).toBe('test@example.com')
    })
  })

  describe('forgotPassword', () => {
    it('should call POST /auth/forgot_password', async () => {
      const { default: apiClient } = await import('@/api/client')
      vi.mocked(apiClient.post).mockResolvedValue({
        data: { success: true },
      })

      const result = await authApi.forgotPassword({ email: 'test@example.com' })

      expect(apiClient.post).toHaveBeenCalledWith('/auth/forgot_password', {
        email: 'test@example.com',
      })
      expect(result.data.success).toBe(true)
    })
  })

  describe('resetPassword', () => {
    it('should call POST /auth/reset_password', async () => {
      const { default: apiClient } = await import('@/api/client')
      vi.mocked(apiClient.post).mockResolvedValue({
        data: { success: true },
      })

      const result = await authApi.resetPassword({
        code: 'reset-code',
        newPassword: 'NewPassword123!',
      })

      expect(apiClient.post).toHaveBeenCalledWith('/auth/reset_password', {
        code: 'reset-code',
        newPassword: 'NewPassword123!',
      })
      expect(result.data.success).toBe(true)
    })
  })
})

describe('userApi', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  describe('getMe', () => {
    it('should call GET /user/me', async () => {
      const { default: apiClient } = await import('@/api/client')
      vi.mocked(apiClient.get).mockResolvedValue({
        data: {
          success: true,
          data: {
            id: '1',
            username: 'test',
            email: 'test@test.com',
            emailVerified: true,
            approved: true,
            createdAt: '2024-01-01',
          },
        },
      })

      const result = await userApi.getMe()

      expect(apiClient.get).toHaveBeenCalledWith('/user/me')
      expect(result.data.data.username).toBe('test')
    })
  })

  describe('updateMe', () => {
    it('should call PUT /user/me with partial user data', async () => {
      const { default: apiClient } = await import('@/api/client')
      vi.mocked(apiClient.put).mockResolvedValue({
        data: {
          success: true,
          data: {
            id: '1',
            username: 'updated',
            email: 'test@test.com',
            emailVerified: true,
            approved: true,
            createdAt: '2024-01-01',
          },
        },
      })

      const result = await userApi.updateMe({ name: 'Updated Name' })

      expect(apiClient.put).toHaveBeenCalledWith('/user/me', { name: 'Updated Name' })
      expect(result.data.data.username).toBe('updated')
    })
  })

  describe('updatePassword', () => {
    it('should call PUT /user/me/password', async () => {
      const { default: apiClient } = await import('@/api/client')
      vi.mocked(apiClient.put).mockResolvedValue({
        data: { success: true },
      })

      const result = await userApi.updatePassword({ oldPassword: 'old', newPassword: 'new' })

      expect(apiClient.put).toHaveBeenCalledWith('/user/me/password', {
        oldPassword: 'old',
        newPassword: 'new',
      })
      expect(result.data.success).toBe(true)
    })
  })

  describe('getSessions', () => {
    it('should call GET /user/me/sessions', async () => {
      const { default: apiClient } = await import('@/api/client')
      vi.mocked(apiClient.get).mockResolvedValue({
        data: {
          success: true,
          data: [
            { id: '1', userId: 'user-1', createdAt: '2024-01-01', expiresAt: '2024-12-31' },
            { id: '2', userId: 'user-1', createdAt: '2024-01-02', expiresAt: '2024-12-31' },
          ],
        },
      })

      const result = await userApi.getSessions()

      expect(apiClient.get).toHaveBeenCalledWith('/user/me/sessions')
      expect(result.data.data.length).toBe(2)
    })
  })

  describe('deleteSession', () => {
    it('should call DELETE /user/me/sessions/:id', async () => {
      const { default: apiClient } = await import('@/api/client')
      vi.mocked(apiClient.delete).mockResolvedValue({
        data: { success: true },
      })

      const result = await userApi.deleteSession('session-1')

      expect(apiClient.delete).toHaveBeenCalledWith('/user/me/sessions/session-1')
      expect(result.data.success).toBe(true)
    })
  })

  describe('deleteAllSessions', () => {
    it('should call DELETE /user/me/sessions', async () => {
      const { default: apiClient } = await import('@/api/client')
      vi.mocked(apiClient.delete).mockResolvedValue({
        data: { success: true },
      })

      const result = await userApi.deleteAllSessions()

      expect(apiClient.delete).toHaveBeenCalledWith('/user/me/sessions')
      expect(result.data.success).toBe(true)
    })
  })
})

describe('passkeyApi', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  describe('registrationStart', () => {
    it('should call POST /passkey/registration/start', async () => {
      const { default: apiClient } = await import('@/api/client')
      vi.mocked(apiClient.post).mockResolvedValue({
        data: { success: true, data: { challenge: 'challenge', options: '{}' } },
      })

      const result = await passkeyApi.registrationStart()

      expect(apiClient.post).toHaveBeenCalledWith('/passkey/registration/start')
      expect(result.data.data.challenge).toBe('challenge')
    })
  })

  describe('registrationEnd', () => {
    it('should call POST /passkey/registration/end', async () => {
      const { default: apiClient } = await import('@/api/client')
      vi.mocked(apiClient.post).mockResolvedValue({
        data: {
          success: true,
          data: {
            id: 'pk-1',
            credentialId: 'cred-1',
            name: 'My Passkey',
            createdAt: '2024-01-01',
            updatedAt: '2024-01-01',
          },
        },
      })

      const result = await passkeyApi.registrationEnd({
        challenge: 'challenge',
        options: '{}',
        name: 'My Passkey',
      })

      expect(apiClient.post).toHaveBeenCalledWith('/passkey/registration/end', {
        challenge: 'challenge',
        options: '{}',
        name: 'My Passkey',
      })
      expect(result.data.data.name).toBe('My Passkey')
    })
  })

  describe('getPasskeys', () => {
    it('should call GET /passkey', async () => {
      const { default: apiClient } = await import('@/api/client')
      vi.mocked(apiClient.get).mockResolvedValue({
        data: {
          success: true,
          data: [
            {
              id: 'pk-1',
              credentialId: 'cred-1',
              createdAt: '2024-01-01',
              updatedAt: '2024-01-01',
            },
            {
              id: 'pk-2',
              credentialId: 'cred-2',
              createdAt: '2024-01-02',
              updatedAt: '2024-01-02',
            },
          ],
        },
      })

      const result = await passkeyApi.getPasskeys()

      expect(apiClient.get).toHaveBeenCalledWith('/passkey')
      expect(result.data.data.length).toBe(2)
    })
  })

  describe('deletePasskey', () => {
    it('should call DELETE /passkey/:id', async () => {
      const { default: apiClient } = await import('@/api/client')
      vi.mocked(apiClient.delete).mockResolvedValue({
        data: { success: true },
      })

      const result = await passkeyApi.deletePasskey('pk-1')

      expect(apiClient.delete).toHaveBeenCalledWith('/passkey/pk-1')
      expect(result.data.success).toBe(true)
    })
  })
})

describe('totpApi', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  describe('register', () => {
    it('should call POST /totp/registration', async () => {
      const { default: apiClient } = await import('@/api/client')
      vi.mocked(apiClient.post).mockResolvedValue({
        data: { success: true, data: { qr_code_uri: 'otpauth://totp/Test', secret: 'secret' } },
      })

      const result = await totpApi.register()

      expect(apiClient.post).toHaveBeenCalledWith('/totp/registration')
      expect(result.data.data.qr_code_uri).toBe('otpauth://totp/Test')
    })
  })

  describe('verify', () => {
    it('should call POST /totp/verify', async () => {
      const { default: apiClient } = await import('@/api/client')
      vi.mocked(apiClient.post).mockResolvedValue({
        data: { success: true },
      })

      const result = await totpApi.verify({ token: '123456' })

      expect(apiClient.post).toHaveBeenCalledWith('/totp/verify', { token: '123456' })
      expect(result.data.success).toBe(true)
    })
  })

  describe('delete', () => {
    it('should call DELETE /totp', async () => {
      const { default: apiClient } = await import('@/api/client')
      vi.mocked(apiClient.delete).mockResolvedValue({
        data: { success: true },
      })

      const result = await totpApi.delete({ token: '123456' })

      expect(apiClient.delete).toHaveBeenCalledWith('/totp', { data: { token: '123456' } })
      expect(result.data.success).toBe(true)
    })
  })
})
