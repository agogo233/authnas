import apiClient from '../client'
import type { ApiResponse, ListResponse } from '@/types'

export interface CreateProxyAuthRequest {
  name: string
  proxyUrl: string
  headerName: string
  scopes?: string
  groupId?: string
  enabled?: boolean
}

export interface UpdateProxyAuthRequest {
  name?: string
  proxyUrl?: string
  headerName?: string
  scopes?: string
  groupId?: string
  enabled?: boolean
}

export interface ProxyAuthListItem {
  id: string
  name: string
  proxyUrl: string
  enabled: boolean
  createdAt: string
}

export const adminProxyAuthApi = {
  list: () => apiClient.get<ListResponse<ProxyAuthListItem>>('/admin/proxyauth'),
  get: (id: string) => apiClient.get<ApiResponse<ProxyAuthListItem>>(`/admin/proxyauth/${id}`),
  create: (data: CreateProxyAuthRequest) =>
    apiClient.post<ApiResponse<ProxyAuthListItem>>('/admin/proxyauth', data),
  update: (id: string, data: UpdateProxyAuthRequest) =>
    apiClient.put<ApiResponse<void>>(`/admin/proxyauth/${id}`, data),
  delete: (id: string) => apiClient.delete<ApiResponse<void>>(`/admin/proxyauth/${id}`),
}
