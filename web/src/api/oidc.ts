import apiClient from './client'
import type { ApiResponse } from '@/types'

export interface OIDCInteraction {
  uid: string
  client: {
    clientId: string
    name: string
    logoUri?: string
  }
  scopes: string[]
  claims: Record<string, any>
  requestUrl: string
}

export const oidcApi = {
  getInteraction: (uid: string) =>
    apiClient.get<ApiResponse<OIDCInteraction>>(`/oidc/interaction/${uid}`),

  confirmInteraction: (uid: string) =>
    apiClient.post<ApiResponse<{ redirectTo: string }>>(`/oidc/interaction/${uid}/confirm`),

  cancelInteraction: (uid: string) =>
    apiClient.delete<ApiResponse<{ redirectTo: string }>>(`/oidc/interaction/${uid}/cancel`),

  refreshToken: (refreshToken: string) =>
    apiClient.post<
      ApiResponse<{
        accessToken: string
        refreshToken: string
        expiresIn: number
        expiresAt?: string
      }>
    >(`/oidc/token`, {
      grant_type: 'refresh_token',
      refresh_token: refreshToken,
    }),

  logout: (idTokenHint?: string, postLogoutRedirectURI?: string, state?: string) => {
    const params = new URLSearchParams()
    if (idTokenHint) params.append('id_token_hint', idTokenHint)
    if (postLogoutRedirectURI) params.append('post_logout_redirect_uri', postLogoutRedirectURI)
    if (state) params.append('state', state)
    const queryString = params.toString()
    return apiClient.get<ApiResponse<{ redirectTo: string }>>(
      `/oidc/logout${queryString ? `?${queryString}` : ''}`
    )
  },
}
