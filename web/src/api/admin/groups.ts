import apiClient from '../client'
import type { ApiResponse, ListResponse } from '@/types'

export interface CreateGroupRequest {
  name: string
  description?: string
}

export interface UpdateGroupRequest {
  name?: string
  description?: string
}

export interface GroupListItem {
  id: string
  name: string
  description?: string
  createdAt: string
}

export const adminGroupsApi = {
  list: () => apiClient.get<ListResponse<GroupListItem>>('/admin/groups'),
  get: (id: string) => apiClient.get<ApiResponse<GroupListItem>>(`/admin/groups/${id}`),
  create: (data: CreateGroupRequest) =>
    apiClient.post<ApiResponse<GroupListItem>>('/admin/groups', data),
  update: (id: string, data: UpdateGroupRequest) =>
    apiClient.put<ApiResponse<GroupListItem>>(`/admin/groups/${id}`, data),
  delete: (id: string) => apiClient.delete<ApiResponse<void>>(`/admin/groups/${id}`),
}
