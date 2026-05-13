import apiClient from '../client'
import type { ApiResponse, ListResponse } from '@/types'

export interface CreateInvitationRequest {
  email: string
  username?: string
  scopes?: string
  groupId?: string
  maxUses?: number
  expiresAt?: string
}

export interface InvitationDetailItem {
  id: string
  email: string
  username?: string
  code: string
  expiresAt: string
  createdAt: string
}

export const adminInvitationsApi = {
  list: () => apiClient.get<ListResponse<InvitationDetailItem>>('/admin/invitations'),
  get: (id: string) => apiClient.get<ApiResponse<InvitationDetailItem>>(`/admin/invitations/${id}`),
  create: (data: CreateInvitationRequest) =>
    apiClient.post<ApiResponse<InvitationDetailItem>>('/admin/invitations', data),
  delete: (id: string) => apiClient.delete<ApiResponse<void>>(`/admin/invitations/${id}`),
}
