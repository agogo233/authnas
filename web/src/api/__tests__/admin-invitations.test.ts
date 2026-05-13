import { describe, it, expect, vi, beforeEach } from 'vitest'
import { adminInvitationsApi } from '@/api/admin/invitations'

vi.mock('@/api/client', () => ({
  default: {
    get: vi.fn(),
    post: vi.fn(),
    delete: vi.fn(),
  },
}))

describe('adminInvitationsApi', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  describe('list', () => {
    it('should call GET /admin/invitations', async () => {
      const { default: apiClient } = await import('@/api/client')
      vi.mocked(apiClient.get).mockResolvedValue({
        data: {
          success: true,
          data: [
            {
              id: 'inv-1',
              email: 'user1@example.com',
              code: 'CODE1',
              expiresAt: '2024-12-31',
              createdAt: '2024-01-01',
            },
            {
              id: 'inv-2',
              email: 'user2@example.com',
              code: 'CODE2',
              expiresAt: '2024-12-31',
              createdAt: '2024-01-02',
            },
          ],
          total: 2,
        },
      })

      const result = await adminInvitationsApi.list()

      expect(apiClient.get).toHaveBeenCalledWith('/admin/invitations')
      expect(result.data.data.length).toBe(2)
      expect(result.data.data[0].email).toBe('user1@example.com')
    })

    it('should return empty list when no invitations', async () => {
      const { default: apiClient } = await import('@/api/client')
      vi.mocked(apiClient.get).mockResolvedValue({
        data: { success: true, data: [], total: 0 },
      })

      const result = await adminInvitationsApi.list()

      expect(result.data.data).toEqual([])
      expect(result.data.total).toBe(0)
    })
  })

  describe('get', () => {
    it('should call GET /admin/invitations/:id', async () => {
      const { default: apiClient } = await import('@/api/client')
      vi.mocked(apiClient.get).mockResolvedValue({
        data: {
          success: true,
          data: {
            id: 'invitation-123',
            email: 'test@example.com',
            username: 'testuser',
            code: 'TESTCODE123',
            expiresAt: '2024-12-31T23:59:59Z',
            createdAt: '2024-01-01T00:00:00Z',
          },
        },
      })

      const result = await adminInvitationsApi.get('invitation-123')

      expect(apiClient.get).toHaveBeenCalledWith('/admin/invitations/invitation-123')
      expect(result.data.data.email).toBe('test@example.com')
      expect(result.data.data.code).toBe('TESTCODE123')
    })

    it('should return invitation without username', async () => {
      const { default: apiClient } = await import('@/api/client')
      vi.mocked(apiClient.get).mockResolvedValue({
        data: {
          success: true,
          data: {
            id: 'invitation-456',
            email: 'user@example.com',
            code: 'CODE456',
            expiresAt: '2024-12-31',
            createdAt: '2024-01-01',
          },
        },
      })

      const result = await adminInvitationsApi.get('invitation-456')

      expect(result.data.data.username).toBeUndefined()
    })
  })

  describe('create', () => {
    it('should call POST /admin/invitations with all fields', async () => {
      const { default: apiClient } = await import('@/api/client')
      vi.mocked(apiClient.post).mockResolvedValue({
        data: {
          success: true,
          data: {
            id: 'new-inv-123',
            email: 'newuser@example.com',
            username: 'newuser',
            code: 'NEWCODE123',
            expiresAt: '2024-12-31',
            createdAt: '2024-01-01',
          },
        },
      })

      const result = await adminInvitationsApi.create({
        email: 'newuser@example.com',
        username: 'newuser',
        scopes: 'openid profile',
        groupId: 'group-123',
        maxUses: 5,
        expiresAt: '2024-12-31',
      })

      expect(apiClient.post).toHaveBeenCalledWith('/admin/invitations', {
        email: 'newuser@example.com',
        username: 'newuser',
        scopes: 'openid profile',
        groupId: 'group-123',
        maxUses: 5,
        expiresAt: '2024-12-31',
      })
      expect(result.data.data.id).toBe('new-inv-123')
    })

    it('should call POST /admin/invitations with only required fields', async () => {
      const { default: apiClient } = await import('@/api/client')
      vi.mocked(apiClient.post).mockResolvedValue({
        data: {
          success: true,
          data: {
            id: 'inv-789',
            email: 'user@example.com',
            code: 'CODE789',
            expiresAt: '2024-12-31',
            createdAt: '2024-01-01',
          },
        },
      })

      const result = await adminInvitationsApi.create({
        email: 'user@example.com',
      })

      expect(apiClient.post).toHaveBeenCalledWith('/admin/invitations', {
        email: 'user@example.com',
      })
      expect(result.data.data.email).toBe('user@example.com')
    })

    it('should create invitation with scopes only', async () => {
      const { default: apiClient } = await import('@/api/client')
      vi.mocked(apiClient.post).mockResolvedValue({
        data: {
          success: true,
          data: {
            id: 'inv-scope',
            email: 'scoped@example.com',
            scopes: 'openid email',
            code: 'SCOPECODE',
            expiresAt: '2024-12-31',
            createdAt: '2024-01-01',
          },
        },
      })

      await adminInvitationsApi.create({
        email: 'scoped@example.com',
        scopes: 'openid email',
      })

      expect(apiClient.post).toHaveBeenCalledWith('/admin/invitations', {
        email: 'scoped@example.com',
        scopes: 'openid email',
      })
    })
  })

  describe('delete', () => {
    it('should call DELETE /admin/invitations/:id', async () => {
      const { default: apiClient } = await import('@/api/client')
      vi.mocked(apiClient.delete).mockResolvedValue({
        data: { success: true },
      })

      const result = await adminInvitationsApi.delete('invitation-123')

      expect(apiClient.delete).toHaveBeenCalledWith('/admin/invitations/invitation-123')
      expect(result.data.success).toBe(true)
    })
  })
})
