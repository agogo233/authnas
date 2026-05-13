import apiClient from '../client'
import type { ApiResponse, PaginatedResponse } from '@/types'

export interface CreateUserRequest {
  email: string
  username: string
  password?: string
  name?: string
  isAdmin?: boolean
  approved?: boolean
  mfaRequired?: boolean
}

export interface UpdateUserRequest {
  email?: string
  username?: string
  name?: string
  isAdmin?: boolean
  approved?: boolean
  mfaRequired?: boolean
}

export interface ResetPasswordRequest {
  newPassword: string
}

export interface UserListItem {
  id: string
  email?: string
  username: string
  name?: string
  emailVerified: boolean
  approved: boolean
  mfaRequired: boolean
  isAdmin: boolean
  createdAt: string
}

export const adminUsersApi = {
  count: () => apiClient.get<ApiResponse<{ total: number }>>('/admin/users/count'),
  list: (params?: { page?: number; pageSize?: number; search?: string }) => {
    const searchParams = new URLSearchParams()
    if (params?.page) searchParams.set('page', String(params.page))
    if (params?.pageSize) searchParams.set('pageSize', String(params.pageSize))
    if (params?.search) searchParams.set('search', params.search)
    const query = searchParams.toString()
    return apiClient.get<PaginatedResponse<UserListItem>>(`/admin/users${query ? `?${query}` : ''}`)
  },
  get: (id: string) => apiClient.get<ApiResponse<UserListItem>>(`/admin/users/${id}`),
  create: (data: CreateUserRequest) =>
    apiClient.post<ApiResponse<UserListItem>>('/admin/users', data),
  update: (id: string, data: UpdateUserRequest) =>
    apiClient.put<ApiResponse<void>>(`/admin/users/${id}`, data),
  delete: (id: string) => apiClient.delete<ApiResponse<void>>(`/admin/users/${id}`),
  approve: (id: string, data: { approved: boolean }) =>
    apiClient.post<ApiResponse<void>>(`/admin/users/${id}/approve`, data),
  resetPassword: (id: string, data: ResetPasswordRequest) =>
    apiClient.post<ApiResponse<void>>(`/admin/users/${id}/reset-password`, data),
}
