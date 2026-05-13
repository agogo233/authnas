import { describe, it, expect, vi, beforeEach } from 'vitest'
import { adminInvitationsApi, type CreateInvitationRequest } from '../invitations'
import apiClient from '@/api/client'

vi.mock('@/api/client', () => ({
  default: {
    get: vi.fn(),
    post: vi.fn(),
    delete: vi.fn(),
  },
}))

describe('adminInvitationsApi', () => {
  const mockedApiClient = apiClient as ReturnType<typeof vi.fn> & {
    get: ReturnType<typeof vi.fn>
    post: ReturnType<typeof vi.fn>
    delete: ReturnType<typeof vi.fn>
  }

  beforeEach(() => {
    vi.clearAllMocks()
  })

  describe('list', () => {
    it('should call get with correct path', async () => {
      mockedApiClient.get.mockResolvedValue({
        data: {
          success: true,
          data: [
            {
              id: '1',
              email: 'test@example.com',
              code: 'ABC123',
              expiresAt: '2024-12-31T00:00:00Z',
            },
          ],
          total: 1,
        },
      })

      const result = await adminInvitationsApi.list()

      expect(mockedApiClient.get).toHaveBeenCalledWith('/admin/invitations')
      expect(result.data?.data).toHaveLength(1)
      expect(result.data?.total).toBe(1)
    })

    it('should return empty list when no invitations exist', async () => {
      mockedApiClient.get.mockResolvedValue({
        data: {
          success: true,
          data: [],
          total: 0,
        },
      })

      const result = await adminInvitationsApi.list()

      expect(result.data?.data).toHaveLength(0)
      expect(result.data?.total).toBe(0)
    })
  })

  describe('get', () => {
    it('should call get with correct id', async () => {
      mockedApiClient.get.mockResolvedValue({
        data: {
          success: true,
          data: {
            id: '1',
            email: 'test@example.com',
            code: 'ABC123',
            expiresAt: '2024-12-31T00:00:00Z',
          },
        },
      })

      const result = await adminInvitationsApi.get('1')

      expect(mockedApiClient.get).toHaveBeenCalledWith('/admin/invitations/1')
      expect(result.data?.data).toEqual({
        id: '1',
        email: 'test@example.com',
        code: 'ABC123',
        expiresAt: '2024-12-31T00:00:00Z',
      })
    })
  })

  describe('create', () => {
    it('should call post with create request data', async () => {
      const createData: CreateInvitationRequest = {
        email: 'invite@example.com',
        username: 'inviteduser',
        scopes: 'openid profile',
        groupId: 'group1',
        maxUses: 5,
        expiresAt: '2024-12-31T00:00:00Z',
      }
      mockedApiClient.post.mockResolvedValue({
        data: {
          success: true,
          data: { id: '2', ...createData, code: 'XYZ789', createdAt: '2024-01-01T00:00:00Z' },
        },
      })

      const result = await adminInvitationsApi.create(createData)

      expect(mockedApiClient.post).toHaveBeenCalledWith('/admin/invitations', createData)
      expect(result.data?.data).toMatchObject({ id: '2', email: 'invite@example.com' })
    })

    it('should send minimal create request with only email', async () => {
      const createData: CreateInvitationRequest = {
        email: 'simple@example.com',
      }
      mockedApiClient.post.mockResolvedValue({
        data: {
          success: true,
          data: { id: '3', ...createData, code: 'SIMPLE', createdAt: '2024-01-01T00:00:00Z' },
        },
      })

      await adminInvitationsApi.create(createData)

      expect(mockedApiClient.post).toHaveBeenCalledWith('/admin/invitations', createData)
    })
  })

  describe('delete', () => {
    it('should call delete with correct id', async () => {
      mockedApiClient.delete.mockResolvedValue({
        data: { success: true },
      })

      await adminInvitationsApi.delete('1')

      expect(mockedApiClient.delete).toHaveBeenCalledWith('/admin/invitations/1')
    })
  })
})
