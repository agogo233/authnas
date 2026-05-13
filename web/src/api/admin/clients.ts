import apiClient from '../client'
import type { ApiResponse, ListResponse } from '@/types'

export interface CreateClientRequest {
  clientId: string
  name: string
  logoUri?: string
  redirectUris?: string
  postLogoutRedirectUris?: string
  grantTypes?: string
  responseTypes?: string
  scopes?: string
}

export interface UpdateClientRequest {
  name?: string
  logoUri?: string
  redirectUris?: string
  postLogoutRedirectUris?: string
  grantTypes?: string
  responseTypes?: string
  scopes?: string
}

export interface ClientListItem {
  id: string
  clientId: string
  name: string
  logoUri?: string
  createdAt: string
}

export const adminClientsApi = {
  list: () => apiClient.get<ListResponse<ClientListItem>>('/admin/clients'),
  get: (id: string) => apiClient.get<ApiResponse<ClientListItem>>(`/admin/clients/${id}`),
  create: (data: CreateClientRequest) =>
    apiClient.post<ApiResponse<ClientListItem>>('/admin/clients', data),
  update: (id: string, data: UpdateClientRequest) =>
    apiClient.put<ApiResponse<void>>(`/admin/clients/${id}`, data),
  delete: (id: string) => apiClient.delete<ApiResponse<void>>(`/admin/clients/${id}`),
}
