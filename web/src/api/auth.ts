import apiClient from './client'
import type {
  ApiResponse,
  LoginData,
  PublicConfig,
  RegisterResponse,
  User,
  UserUpdate,
  Session,
  Passkey,
  PasskeyEndMFAData,
} from '@/types'

export const authApi = {
  getPublicConfig: () => apiClient.get<ApiResponse<PublicConfig>>('/config'),
  getCsrfToken: () =>
    apiClient.get<ApiResponse<{ token: string; expiresAt: string }>>('/auth/csrf'),
  login: (data: { input: string; password: string; remember?: boolean }) =>
    apiClient.post<ApiResponse<LoginData>>('/auth/login', data),
  register: (data: {
    email: string
    username: string
    password: string
    inviteId?: string
    challenge?: string
  }) => apiClient.post<ApiResponse<RegisterResponse>>('/auth/register', data),
  passkeyStart: (data: { username?: string }) =>
    apiClient.post<ApiResponse<{ challenge: string; options: string }>>(
      '/auth/passkey/start',
      data
    ),
  passkeyEnd: (data: { credentialId: string; challenge: string; response: string }) =>
    apiClient.post<ApiResponse<RegisterResponse | PasskeyEndMFAData>>('/auth/passkey/end', data),
  totpVerify: (data: { token: string; mfaToken?: string }) =>
    apiClient.post<ApiResponse<RegisterResponse>>('/auth/totp', data),
  verifyEmail: (data: { userId: string; challenge: string }) =>
    apiClient.post<ApiResponse<void>>('/auth/verify_email', data),
  sendVerifyEmail: (data: { email: string }) =>
    apiClient.post<ApiResponse<void>>('/auth/send_verify_email', data),
  getInvitation: (id: string, challenge: string) =>
    apiClient.get<ApiResponse<{ email: string; username?: string }>>(
      `/auth/invitation/${id}/${challenge}`
    ),
  forgotPassword: (data: { email: string }) =>
    apiClient.post<ApiResponse<void>>('/auth/forgot_password', data),
  resetPassword: (data: { code: string; newPassword: string }) =>
    apiClient.post<ApiResponse<void>>('/auth/reset_password', data),
}

export const userApi = {
  getMe: () => apiClient.get<ApiResponse<User>>('/user/me'),
  updateMe: (data: UserUpdate) => apiClient.put<ApiResponse<User>>('/user/me', data),
  updatePassword: (data: { oldPassword: string; newPassword: string }) =>
    apiClient.put<ApiResponse<void>>('/user/me/password', data),
  getSessions: () => apiClient.get<ApiResponse<Session[]>>('/user/me/sessions'),
  deleteSession: (id: string) => apiClient.delete<ApiResponse<void>>(`/user/me/sessions/${id}`),
  deleteAllSessions: () => apiClient.delete<ApiResponse<void>>('/user/me/sessions'),
}

export const passkeyApi = {
  registrationStart: () =>
    apiClient.post<ApiResponse<{ challenge: string; options: string }>>(
      '/passkey/registration/start'
    ),
  registrationEnd: (data: { challenge: string; options: string; name?: string }) =>
    apiClient.post<
      ApiResponse<{
        id: string
        name?: string
        credentialId: string
        createdAt: string
        updatedAt: string
      }>
    >('/passkey/registration/end', data),
  getPasskeys: () => apiClient.get<ApiResponse<Passkey[]>>('/passkey'),
  deletePasskey: (id: string) => apiClient.delete<ApiResponse<void>>(`/passkey/${id}`),
}

export const totpApi = {
  register: () =>
    apiClient.post<ApiResponse<{ qr_code_uri: string; secret: string }>>('/totp/registration'),
  verify: (data: { token: string }) => apiClient.post<ApiResponse<void>>('/totp/verify', data),
  delete: (data: { token: string }) => apiClient.delete<ApiResponse<void>>('/totp', { data }),
}
