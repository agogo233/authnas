export interface User {
  id: string
  email?: string
  username: string
  name?: string
  isAdmin?: boolean
  emailVerified: boolean
  approved: boolean
  mfaRequired?: boolean
  hasTotp?: boolean
  hasPasskeys?: boolean
  hasPassword?: boolean
  tokenVersion?: number
  expiresAt?: string
  createdAt: string
  updatedAt?: string
}

export interface UserUpdate {
  email?: string
  name?: string
}

export interface LoginRequest {
  input: string
  password: string
  remember?: boolean
}

export interface RegisterRequest {
  email: string
  username: string
  password: string
  inviteId?: string
  challenge?: string
}

export interface LoginData {
  accessToken: string
  refreshToken: string
  expiresAt?: string
  user: User
}

export interface PublicConfig {
  app_name: string
  signup_requires_approval: boolean
  email_verification: boolean
  mfa_required: boolean
  password_min_length: number
  password_strength: number
  default_redirect: string
  contact_email: string
}

export interface RegisterResponse {
  accessToken: string
  refreshToken: string
  expiresAt?: string
  user?: User
}

export interface PasskeyEndMFAData {
  mfaRequired: boolean
  userId: string
  mfaToken: string
}

export interface AuthResponse {
  accessToken: string
  refreshToken: string
  expiresAt?: string
  user?: User
}

export interface Group {
  id: string
  name: string
  description?: string
  createdAt: string
  updatedAt?: string
}

export interface Client {
  id: string
  clientId: string
  clientSecret?: string
  name: string
  logoUri?: string
  redirectUris: string
  postLogoutRedirectUris?: string
  grantTypes?: string
  responseTypes?: string
  scopes?: string
  createdAt: string
  updatedAt?: string
}

export interface Invitation {
  id: string
  email: string
  username?: string
  code: string
  scopes?: string
  groupId?: string
  maxUses?: number
  usedCount: number
  expiresAt: string
  createdAt: string
  createdBy?: string
}

export interface Passkey {
  id: string
  userId?: string
  name?: string
  credentialId: string
  attestationType?: string
  lastUsedAt?: string
  createdAt: string
  updatedAt: string
}

export interface TOTP {
  id: string
  userId: string
  issuer: string
  createdAt: string
  updatedAt?: string
}

export interface ProxyAuth {
  id: string
  name: string
  proxyUrl: string
  headerName: string
  scopes?: string
  groupId?: string
  enabled: boolean
  createdAt: string
  updatedAt?: string
}

export interface ApiResponse<T> {
  success: boolean
  data?: T
  message?: string
  code?: string
}

export interface PaginatedResponse<T> {
  success: boolean
  data: T[]
  total: number
  page: number
  pageSize: number
}

export interface ListResponse<T> {
  success: boolean
  data: T[]
  total: number
}

export interface Session {
  id: string
  userId: string
  clientId?: string
  createdAt: string
  expiresAt: string
  userAgent?: string
}
