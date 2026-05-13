import { describe, it, expect, vi, beforeEach } from 'vitest'
import {
  adminUsersApi,
  type CreateUserRequest,
  type UpdateUserRequest,
  type ResetPasswordRequest,
} from '../users'
import apiClient from '@/api/client'

vi.mock('@/api/client', () => ({
  default: {
    get: vi.fn(),
    post: vi.fn(),
    put: vi.fn(),
    delete: vi.fn(),
  },
}))

describe('adminUsersApi', () => {
  const mockedApiClient = apiClient as unknown as {
    get: ReturnType<typeof vi.fn>
    post: ReturnType<typeof vi.fn>
    put: ReturnType<typeof vi.fn>
    delete: ReturnType<typeof vi.fn>
  }

  beforeEach(() => {
    vi.clearAllMocks()
  })

  describe('count', () => {
    it('should call get with correct path', async () => {
      mockedApiClient.get.mockResolvedValue({
        data: {
          success: true,
          data: { total: 42 },
        },
      })

      const result = await adminUsersApi.count()

      expect(mockedApiClient.get).toHaveBeenCalledWith('/admin/users/count')
      expect(result.data?.data).toEqual({ total: 42 })
    })
  })

  describe('list', () => {
    it('should call get with correct path without params', async () => {
      mockedApiClient.get.mockResolvedValue({
        data: {
          success: true,
          data: [
            {
              id: '1',
              username: 'user1',
              emailVerified: true,
              approved: true,
              mfaRequired: false,
              isAdmin: false,
            },
          ],
          total: 1,
          page: 1,
          pageSize: 10,
        },
      })

      const result = await adminUsersApi.list()

      expect(mockedApiClient.get).toHaveBeenCalledWith('/admin/users')
      expect(result.data?.data).toHaveLength(1)
    })

    it('should call get with page parameter', async () => {
      mockedApiClient.get.mockResolvedValue({
        data: { success: true, data: [], total: 0, page: 2, pageSize: 10 },
      })

      await adminUsersApi.list({ page: 2 })

      expect(mockedApiClient.get).toHaveBeenCalledWith('/admin/users?page=2')
    })

    it('should call get with pageSize parameter', async () => {
      mockedApiClient.get.mockResolvedValue({
        data: { success: true, data: [], total: 0, page: 1, pageSize: 20 },
      })

      await adminUsersApi.list({ pageSize: 20 })

      expect(mockedApiClient.get).toHaveBeenCalledWith('/admin/users?pageSize=20')
    })

    it('should call get with search parameter', async () => {
      mockedApiClient.get.mockResolvedValue({
        data: { success: true, data: [], total: 0, page: 1, pageSize: 10 },
      })

      await adminUsersApi.list({ search: 'john' })

      expect(mockedApiClient.get).toHaveBeenCalledWith('/admin/users?search=john')
    })

    it('should call get with all parameters combined', async () => {
      mockedApiClient.get.mockResolvedValue({
        data: { success: true, data: [], total: 0, page: 2, pageSize: 20 },
      })

      await adminUsersApi.list({ page: 2, pageSize: 20, search: 'john' })

      expect(mockedApiClient.get).toHaveBeenCalledWith(
        '/admin/users?page=2&pageSize=20&search=john'
      )
    })
  })

  describe('get', () => {
    it('should call get with correct id', async () => {
      mockedApiClient.get.mockResolvedValue({
        data: {
          success: true,
          data: {
            id: '1',
            username: 'testuser',
            emailVerified: true,
            approved: true,
            mfaRequired: false,
            isAdmin: false,
          },
        },
      })

      const result = await adminUsersApi.get('1')

      expect(mockedApiClient.get).toHaveBeenCalledWith('/admin/users/1')
      expect(result.data?.data).toMatchObject({ id: '1', username: 'testuser' })
    })
  })

  describe('create', () => {
    it('should call post with create request data', async () => {
      const createData: CreateUserRequest = {
        email: 'new@example.com',
        username: 'newuser',
        password: 'securepassword123',
        name: 'New User',
        isAdmin: false,
        approved: true,
        mfaRequired: false,
      }
      mockedApiClient.post.mockResolvedValue({
        data: {
          success: true,
          data: { id: '2', ...createData, emailVerified: false, createdAt: '2024-01-01T00:00:00Z' },
        },
      })

      const result = await adminUsersApi.create(createData)

      expect(mockedApiClient.post).toHaveBeenCalledWith('/admin/users', createData)
      expect(result.data?.data).toMatchObject({
        id: '2',
        email: 'new@example.com',
        username: 'newuser',
      })
    })

    it('should send minimal create request with only required fields', async () => {
      const createData: CreateUserRequest = {
        email: 'minimal@example.com',
        username: 'minimaluser',
      }
      mockedApiClient.post.mockResolvedValue({
        data: {
          success: true,
          data: {
            id: '3',
            ...createData,
            emailVerified: false,
            approved: false,
            mfaRequired: false,
            isAdmin: false,
            createdAt: '2024-01-01T00:00:00Z',
          },
        },
      })

      await adminUsersApi.create(createData)

      expect(mockedApiClient.post).toHaveBeenCalledWith('/admin/users', createData)
    })
  })

  describe('update', () => {
    it('should call put with id and update data', async () => {
      const updateData: UpdateUserRequest = {
        name: 'Updated Name',
        isAdmin: true,
      }
      mockedApiClient.put.mockResolvedValue({
        data: { success: true },
      })

      await adminUsersApi.update('1', updateData)

      expect(mockedApiClient.put).toHaveBeenCalledWith('/admin/users/1', updateData)
    })

    it('should allow partial update with only approval status changed', async () => {
      const updateData: UpdateUserRequest = {
        approved: false,
      }
      mockedApiClient.put.mockResolvedValue({
        data: { success: true },
      })

      await adminUsersApi.update('1', updateData)

      expect(mockedApiClient.put).toHaveBeenCalledWith('/admin/users/1', updateData)
    })

    it('should allow update of mfaRequired flag', async () => {
      const updateData: UpdateUserRequest = {
        mfaRequired: true,
      }
      mockedApiClient.put.mockResolvedValue({
        data: { success: true },
      })

      await adminUsersApi.update('1', updateData)

      expect(mockedApiClient.put).toHaveBeenCalledWith('/admin/users/1', updateData)
    })
  })

  describe('delete', () => {
    it('should call delete with correct id', async () => {
      mockedApiClient.delete.mockResolvedValue({
        data: { success: true },
      })

      await adminUsersApi.delete('1')

      expect(mockedApiClient.delete).toHaveBeenCalledWith('/admin/users/1')
    })
  })

  describe('approve', () => {
    it('should call post with approval data', async () => {
      mockedApiClient.post.mockResolvedValue({
        data: { success: true },
      })

      await adminUsersApi.approve('1', { approved: true })

      expect(mockedApiClient.post).toHaveBeenCalledWith('/admin/users/1/approve', {
        approved: true,
      })
    })

    it('should allow disapproval', async () => {
      mockedApiClient.post.mockResolvedValue({
        data: { success: true },
      })

      await adminUsersApi.approve('1', { approved: false })

      expect(mockedApiClient.post).toHaveBeenCalledWith('/admin/users/1/approve', {
        approved: false,
      })
    })
  })

  describe('resetPassword', () => {
    it('should call post with reset password data', async () => {
      const resetData: ResetPasswordRequest = {
        newPassword: 'newsecurepassword123',
      }
      mockedApiClient.post.mockResolvedValue({
        data: { success: true },
      })

      await adminUsersApi.resetPassword('1', resetData)

      expect(mockedApiClient.post).toHaveBeenCalledWith('/admin/users/1/reset-password', resetData)
    })
  })
})
