import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import type { User } from '@/types'
import { oidcApi } from '@/api/oidc'
import apiClient from '@/api/client'

export const useAuthStore = defineStore('auth', () => {
  const user = ref<User | null>(null)
  const accessToken = ref<string | null>(null)
  const tokenExpiresAt = ref<string | null>(null)
  const initialized = ref(false)

  const isExpired = computed(() => {
    if (!user.value) return true
    const exp = (user.value as any).exp
    if (exp) {
      return Date.now() > exp * 1000
    }
    if (user.value.expiresAt) {
      return new Date(user.value.expiresAt) < new Date()
    }
    if (tokenExpiresAt.value) {
      return new Date(tokenExpiresAt.value) < new Date()
    }
    return false
  })

  const isAuthenticated = computed(() => !!accessToken.value && !!user.value && !isExpired.value)
  const isAdmin = computed(() => user.value?.isAdmin ?? false)

  const expiresAtTimestamp = computed(() => {
    if (tokenExpiresAt.value) {
      return new Date(tokenExpiresAt.value).getTime()
    }
    if (!user.value) return 0
    const exp = (user.value as any).exp
    if (exp) {
      return exp * 1000
    }
    if (user.value.expiresAt) {
      return new Date(user.value.expiresAt).getTime()
    }
    return 0
  })

  const timeUntilExpiry = computed(() => {
    if (!expiresAtTimestamp.value) return 0
    return expiresAtTimestamp.value - Date.now()
  })

  let refreshTimeoutId: ReturnType<typeof setTimeout> | null = null

  function scheduleRefreshIfNeeded() {
    if (refreshTimeoutId) {
      clearTimeout(refreshTimeoutId)
      refreshTimeoutId = null
    }

    if (!expiresAtTimestamp.value) return

    const timeLeft = timeUntilExpiry.value
    const refreshBuffer = 5 * 60 * 1000

    if (timeLeft <= 0) {
      return
    }

    if (timeLeft <= refreshBuffer) {
      refreshAccessToken()
      return
    }

    const timeoutMs = timeLeft - refreshBuffer
    refreshTimeoutId = setTimeout(() => {
      refreshAccessToken()
    }, timeoutMs)
  }

  function setTokens(access: string, expiresAt?: string) {
    accessToken.value = access
    if (expiresAt) {
      tokenExpiresAt.value = expiresAt
    }
    scheduleRefreshIfNeeded()
  }

  function setUser(userData: User) {
    user.value = userData
    scheduleRefreshIfNeeded()
  }

  async function refreshAccessToken(): Promise<boolean> {
    try {
      const response = await apiClient.post<{
        data: { accessToken: string; expiresAt: string }
      }>('/api/auth/refresh')
      if (response.data.data?.accessToken) {
        const newAccessToken = response.data.data.accessToken
        const newExpiresAt = response.data.data.expiresAt
        setTokens(newAccessToken, newExpiresAt)
        return true
      }
      return false
    } catch {
      logout()
      return false
    }
  }

  async function logout(): Promise<void> {
    const idTokenHint = accessToken.value || undefined
    const postLogoutRedirectURI = window.location.origin + '/login'

    try {
      await oidcApi.logout(idTokenHint, postLogoutRedirectURI)
    } catch {}

    user.value = null
    accessToken.value = null
    tokenExpiresAt.value = null

    if (refreshTimeoutId) {
      clearTimeout(refreshTimeoutId)
      refreshTimeoutId = null
    }
  }

  async function initFromStorage() {
    if (initialized.value) {
      scheduleRefreshIfNeeded()
      return
    }

    try {
      const response = await apiClient.get<{
        data: {
          id: string
          email: string | null
          username: string
          name: string | null
          emailVerified: boolean
          approved: boolean
          isAdmin: boolean
          mfaRequired: boolean
          hasTotp: boolean
          hasPasskeys: boolean
          hasPassword: boolean
          tokenVersion: number
          createdAt: string
          updatedAt: string | null
          expiresAt: string | null
        }
      }>('/api/auth/me')

      const userData = response.data.data
      user.value = {
        id: userData.id,
        email: userData.email ?? undefined,
        username: userData.username,
        name: userData.name ?? undefined,
        emailVerified: userData.emailVerified,
        approved: userData.approved,
        isAdmin: userData.isAdmin,
        mfaRequired: userData.mfaRequired,
        hasTotp: userData.hasTotp,
        hasPasskeys: userData.hasPasskeys,
        hasPassword: userData.hasPassword,
        tokenVersion: userData.tokenVersion,
        createdAt: userData.createdAt,
        updatedAt: userData.updatedAt ?? undefined,
        expiresAt: userData.expiresAt ?? undefined,
      }
    } catch {
      user.value = null
      accessToken.value = null
    }

    initialized.value = true
    scheduleRefreshIfNeeded()
  }

  function cleanup() {
    if (refreshTimeoutId) {
      clearTimeout(refreshTimeoutId)
      refreshTimeoutId = null
    }
  }

  return {
    user,
    accessToken,
    tokenExpiresAt,
    isAuthenticated,
    isAdmin,
    isExpired,
    initialized,
    timeUntilExpiry,
    setTokens,
    setUser,
    logout,
    initFromStorage,
    refreshAccessToken,
    scheduleRefreshIfNeeded,
    cleanup,
  }
})
