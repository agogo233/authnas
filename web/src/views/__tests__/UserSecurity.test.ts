import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import UserSecurity from '../user/Security.vue'

const mockMessage = {
  success: vi.fn(),
  error: vi.fn(),
  warning: vi.fn(),
  info: vi.fn(),
}

vi.mock('@/api/auth', () => ({
  userApi: {
    getMe: vi.fn(),
    updatePassword: vi.fn(),
    getSessions: vi.fn(),
    deleteSession: vi.fn(),
    deleteAllSessions: vi.fn(),
  },
  passkeyApi: {
    getPasskeys: vi.fn(),
    deletePasskey: vi.fn(),
    registrationStart: vi.fn(),
    registrationEnd: vi.fn(),
  },
  totpApi: {
    setup: vi.fn(),
    verify: vi.fn(),
    delete: vi.fn(),
    register: vi.fn(),
  },
}))

vi.mock('naive-ui', () => ({
  NCard: {
    name: 'NCard',
    props: ['title', 'bordered'],
    template: '<div class="n-card"><slot /></div>',
  },
  NTabs: {
    name: 'NTabs',
    props: ['value'],
    template: '<div class="n-tabs"><slot /></div>',
  },
  NTabPane: {
    name: 'NTabPane',
    props: ['name', 'tab'],
    template: '<div class="n-tab-pane" :data-name="name"><slot /></div>',
  },
  NForm: {
    name: 'NForm',
    props: ['labelWidth'],
    template: '<form class="n-form"><slot /></form>',
  },
  NFormItem: {
    name: 'NFormItem',
    props: ['label'],
    template: '<div class="n-form-item"><label v-if="label">{{ label }}</label><slot /></div>',
  },
  NInput: {
    name: 'NInput',
    props: ['modelValue', 'type', 'placeholder'],
    emits: ['update:modelValue'],
    template:
      '<input :type="type || \'text\'" :value="modelValue" :placeholder="placeholder" @input="$emit(\'update:modelValue\', $event.target.value)" />',
  },
  NButton: {
    name: 'NButton',
    props: ['type', 'size', 'loading', 'disabled'],
    emits: ['click'],
    template: '<button :disabled="disabled || loading" @click="$emit(\'click\')"><slot /></button>',
  },
  NSpace: {
    name: 'NSpace',
    props: ['justify'],
    template: '<div class="n-space"><slot /></div>',
  },
  NAlert: {
    name: 'NAlert',
    props: ['type'],
    template: '<div class="n-alert"><slot /></div>',
  },
  NModal: {
    name: 'NModal',
    props: ['show', 'preset', 'title'],
    template: '<div class="n-modal" v-if="show"><slot /></div>',
  },
  NImage: {
    name: 'NImage',
    props: ['src', 'width', 'height'],
    template: '<img :src="src" :style="{ width, height }" />',
  },
  NProgress: {
    name: 'NProgress',
    props: ['type', 'percentage', 'color'],
    template: '<div class="n-progress" :style="{ color }">{{ percentage }}%</div>',
  },
  NPopconfirm: {
    name: 'NPopconfirm',
    props: [],
    emits: ['positive-click'],
    template:
      '<div class="n-popconfirm" @click="$emit(\'positive-click\')"><slot name="trigger" /></div>',
  },
  useMessage: () => mockMessage,
}))

describe('UserSecurity.vue', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  const mockUser = {
    id: '1',
    username: 'testuser',
    email: 'test@example.com',
    name: 'Test User',
    isAdmin: false,
    approved: true,
    mfaRequired: false,
    emailVerified: true,
    hasTotp: false,
    createdAt: '2024-01-15T00:00:00Z',
  }

  const mountOptions = {
    global: {
      stubs: {},
    },
  }

  describe('rendering', () => {
    it('renders security page', async () => {
      const { userApi } = await import('@/api/auth')
      vi.mocked(userApi.getMe).mockResolvedValue({ data: { success: true, data: mockUser } } as any)

      const wrapper = mount(UserSecurity, mountOptions)
      await flushPromises()

      expect(wrapper.find('.page-container').exists()).toBe(true)
      expect(wrapper.find('.n-card').exists()).toBe(true)
    })

    it('renders password change form', async () => {
      const { userApi } = await import('@/api/auth')
      vi.mocked(userApi.getMe).mockResolvedValue({ data: { success: true, data: mockUser } } as any)

      const wrapper = mount(UserSecurity, mountOptions)
      await flushPromises()

      expect(wrapper.find('.n-form').exists()).toBe(true)
    })

    it('renders all security sections', async () => {
      const { userApi } = await import('@/api/auth')
      vi.mocked(userApi.getMe).mockResolvedValue({ data: { success: true, data: mockUser } } as any)

      const wrapper = mount(UserSecurity, mountOptions)
      await flushPromises()

      // Should render tabs for different security sections
      expect(wrapper.find('.n-tabs').exists()).toBe(true)
    })
  })

  describe('password strength calculation', () => {
    it('calculates password strength for weak passwords', async () => {
      const { userApi } = await import('@/api/auth')
      vi.mocked(userApi.getMe).mockResolvedValue({ data: { success: true, data: mockUser } } as any)

      const wrapper = mount(UserSecurity, mountOptions)
      const vm = wrapper.vm as any

      // Test the function directly through the component
      const strength = vm.calculatePasswordStrength('abc')
      expect(strength.score).toBeLessThan(3)
    })

    it('calculates password strength for strong passwords', async () => {
      const { userApi } = await import('@/api/auth')
      vi.mocked(userApi.getMe).mockResolvedValue({ data: { success: true, data: mockUser } } as any)

      const wrapper = mount(UserSecurity, mountOptions)
      const vm = wrapper.vm as any

      // Strong password: 12+ chars, mixed case, numbers, special chars
      const strength = vm.calculatePasswordStrength('StrongPass123!')
      expect(strength.score).toBeGreaterThanOrEqual(4)
    })

    it('returns zero score for empty password', async () => {
      const { userApi } = await import('@/api/auth')
      vi.mocked(userApi.getMe).mockResolvedValue({ data: { success: true, data: mockUser } } as any)

      const wrapper = mount(UserSecurity, mountOptions)
      const vm = wrapper.vm as any

      const strength = vm.calculatePasswordStrength('')
      expect(strength.score).toBe(0)
    })
  })

  describe('password change validation', () => {
    it('validates empty fields', async () => {
      const { userApi } = await import('@/api/auth')
      vi.mocked(userApi.getMe).mockResolvedValue({ data: { success: true, data: mockUser } } as any)

      const wrapper = mount(UserSecurity, mountOptions)
      await flushPromises()

      const vm = wrapper.vm as any
      vm.oldPassword = ''
      vm.newPassword = ''
      vm.confirmPassword = ''

      await vm.handlePasswordChange()
      await flushPromises()

      expect(mockMessage.error).toHaveBeenCalledWith('请填写所有字段')
    })

    it('validates password mismatch', async () => {
      const { userApi } = await import('@/api/auth')
      vi.mocked(userApi.getMe).mockResolvedValue({ data: { success: true, data: mockUser } } as any)

      const wrapper = mount(UserSecurity, mountOptions)
      await flushPromises()

      const vm = wrapper.vm as any
      vm.oldPassword = 'old123'
      vm.newPassword = 'new123456'
      vm.confirmPassword = 'different'

      await vm.handlePasswordChange()
      await flushPromises()

      expect(mockMessage.error).toHaveBeenCalledWith('两次输入的密码不一致')
    })

    it('validates weak password strength', async () => {
      const { userApi } = await import('@/api/auth')
      vi.mocked(userApi.getMe).mockResolvedValue({ data: { success: true, data: mockUser } } as any)

      const wrapper = mount(UserSecurity, mountOptions)
      await flushPromises()

      const vm = wrapper.vm as any
      vm.oldPassword = 'old123'
      vm.newPassword = 'abc' // Too weak
      vm.confirmPassword = 'abc'

      await vm.handlePasswordChange()
      await flushPromises()

      expect(mockMessage.error).toHaveBeenCalledWith('密码强度太弱')
    })
  })

  describe('password change submission', () => {
    it('calls updatePassword API on valid submission', async () => {
      const { userApi } = await import('@/api/auth')
      vi.mocked(userApi.getMe).mockResolvedValue({ data: { success: true, data: mockUser } } as any)
      vi.mocked(userApi.updatePassword).mockResolvedValue({ data: { success: true } } as any)

      const wrapper = mount(UserSecurity, mountOptions)
      await flushPromises()

      const vm = wrapper.vm as any
      vm.oldPassword = 'old123'
      vm.newPassword = 'NewStrongPass123!'
      vm.confirmPassword = 'NewStrongPass123!'

      await vm.handlePasswordChange()
      await flushPromises()

      expect(userApi.updatePassword).toHaveBeenCalled()
      expect(mockMessage.success).toHaveBeenCalledWith('密码修改成功')
    })

    it('clears form on success', async () => {
      const { userApi } = await import('@/api/auth')
      vi.mocked(userApi.getMe).mockResolvedValue({ data: { success: true, data: mockUser } } as any)
      vi.mocked(userApi.updatePassword).mockResolvedValue({ data: { success: true } } as any)

      const wrapper = mount(UserSecurity, mountOptions)
      await flushPromises()

      const vm = wrapper.vm as any
      vm.oldPassword = 'old123'
      vm.newPassword = 'NewStrongPass123!'
      vm.confirmPassword = 'NewStrongPass123!'

      await vm.handlePasswordChange()
      await flushPromises()

      expect(vm.oldPassword).toBe('')
      expect(vm.newPassword).toBe('')
      expect(vm.confirmPassword).toBe('')
    })

    it('shows error on API failure', async () => {
      const { userApi } = await import('@/api/auth')
      vi.mocked(userApi.getMe).mockResolvedValue({ data: { success: true, data: mockUser } } as any)
      vi.mocked(userApi.updatePassword).mockRejectedValue({
        response: { data: { message: 'Wrong password' } },
      })

      const wrapper = mount(UserSecurity, mountOptions)
      await flushPromises()

      const vm = wrapper.vm as any
      vm.oldPassword = 'wrong'
      vm.newPassword = 'NewStrongPass123!'
      vm.confirmPassword = 'NewStrongPass123!'

      await vm.handlePasswordChange()
      await flushPromises()

      expect(mockMessage.error).toHaveBeenCalledWith('Wrong password')
    })
  })

  describe('totp status', () => {
    it('fetches TOTP status on mount', async () => {
      const { userApi } = await import('@/api/auth')
      vi.mocked(userApi.getMe).mockResolvedValue({ data: { success: true, data: mockUser } } as any)

      mount(UserSecurity, mountOptions)
      await flushPromises()

      expect(userApi.getMe).toHaveBeenCalled()
    })

    it('stores TOTP enabled status from user data', async () => {
      const { userApi } = await import('@/api/auth')
      vi.mocked(userApi.getMe).mockResolvedValue({
        data: { success: true, data: { ...mockUser, hasTotp: true } },
      } as any)

      const wrapper = mount(UserSecurity, mountOptions)
      await flushPromises()

      const vm = wrapper.vm as any
      expect(vm.totpEnabled).toBe(true)
    })
  })

  describe('TOTP enable flow', () => {
    it('enables TOTP and shows setup modal on success', async () => {
      const { userApi, totpApi } = await import('@/api/auth')
      vi.mocked(userApi.getMe).mockResolvedValue({ data: { success: true, data: mockUser } } as any)
      vi.mocked(totpApi.register).mockResolvedValue({
        data: {
          success: true,
          data: { secret: 'JBSWY3DPEHPK3PXP', qr_code_uri: 'otpauth://test' },
        },
      } as any)

      const wrapper = mount(UserSecurity, mountOptions)
      await flushPromises()

      const vm = wrapper.vm as any
      await vm.handleEnableTotp()
      await flushPromises()

      expect(vm.showTotpSetup).toBe(true)
      expect(vm.totpSecret).toBe('JBSWY3DPEHPK3PXP')
      expect(vm.totpLoading).toBe(false)
    })

    it('shows error message on TOTP enable failure', async () => {
      const { userApi, totpApi } = await import('@/api/auth')
      vi.mocked(userApi.getMe).mockResolvedValue({ data: { success: true, data: mockUser } } as any)
      vi.mocked(totpApi.register).mockRejectedValue({
        response: { data: { message: 'TOTP registration failed' } },
      })

      const wrapper = mount(UserSecurity, mountOptions)
      await flushPromises()

      const vm = wrapper.vm as any
      await vm.handleEnableTotp()
      await flushPromises()

      expect(mockMessage.error).toHaveBeenCalledWith('TOTP registration failed')
      expect(vm.totpLoading).toBe(false)
    })
  })

  describe('TOTP verify flow', () => {
    it('verifies TOTP code successfully', async () => {
      const { userApi, totpApi, passkeyApi } = await import('@/api/auth')
      vi.mocked(userApi.getMe).mockResolvedValue({ data: { success: true, data: mockUser } } as any)
      vi.mocked(totpApi.verify).mockResolvedValue({ data: { success: true } } as any)
      vi.mocked(passkeyApi.getPasskeys).mockResolvedValue({
        data: { success: true, data: [] },
      } as any)
      vi.mocked(userApi.getSessions).mockResolvedValue({ data: { success: true, data: [] } } as any)

      const wrapper = mount(UserSecurity, mountOptions)
      await flushPromises()

      const vm = wrapper.vm as any
      vm.totpVerifyCode = '123456'
      vm.showTotpSetup = true

      await vm.handleVerifyTotp()
      await flushPromises()

      expect(mockMessage.success).toHaveBeenCalledWith('TOTP 启用成功')
      expect(vm.totpEnabled).toBe(true)
      expect(vm.showTotpSetup).toBe(false)
      expect(vm.totpVerifyCode).toBe('')
    })

    it('validates 6-digit TOTP code', async () => {
      const { userApi } = await import('@/api/auth')
      vi.mocked(userApi.getMe).mockResolvedValue({ data: { success: true, data: mockUser } } as any)

      const wrapper = mount(UserSecurity, mountOptions)
      await flushPromises()

      const vm = wrapper.vm as any
      vm.totpVerifyCode = '123' // too short

      await vm.handleVerifyTotp()
      await flushPromises()

      expect(mockMessage.error).toHaveBeenCalledWith('请输入 6 位数字验证码')
    })

    it('shows error on TOTP verify failure', async () => {
      const { userApi, totpApi } = await import('@/api/auth')
      vi.mocked(userApi.getMe).mockResolvedValue({ data: { success: true, data: mockUser } } as any)
      vi.mocked(totpApi.verify).mockRejectedValue({
        response: { data: { message: 'Invalid code' } },
      })

      const wrapper = mount(UserSecurity, mountOptions)
      await flushPromises()

      const vm = wrapper.vm as any
      vm.totpVerifyCode = '000000'

      await vm.handleVerifyTotp()
      await flushPromises()

      expect(mockMessage.error).toHaveBeenCalledWith('Invalid code')
      expect(vm.totpVerifyLoading).toBe(false)
    })
  })

  describe('TOTP disable flow', () => {
    it('opens disable TOTP modal', async () => {
      const { userApi } = await import('@/api/auth')
      vi.mocked(userApi.getMe).mockResolvedValue({ data: { success: true, data: mockUser } } as any)

      const wrapper = mount(UserSecurity, mountOptions)
      await flushPromises()

      const vm = wrapper.vm as any
      vm.totpEnabled = true

      await vm.handleDisableTotp()
      await flushPromises()

      expect(vm.showDisableTotpModal).toBe(true)
      expect(vm.disableTotpCode).toBe('')
    })

    it('validates disable code length', async () => {
      const { userApi } = await import('@/api/auth')
      vi.mocked(userApi.getMe).mockResolvedValue({ data: { success: true, data: mockUser } } as any)

      const wrapper = mount(UserSecurity, mountOptions)
      await flushPromises()

      const vm = wrapper.vm as any
      vm.disableTotpCode = '123'

      await vm.confirmDisableTotp()
      await flushPromises()

      expect(mockMessage.error).toHaveBeenCalledWith('请输入 6 位数字验证码')
    })

    it('disables TOTP successfully', async () => {
      const { userApi, totpApi, passkeyApi } = await import('@/api/auth')
      vi.mocked(userApi.getMe).mockResolvedValue({ data: { success: true, data: mockUser } } as any)
      vi.mocked(totpApi.delete).mockResolvedValue({ data: { success: true } } as any)
      vi.mocked(passkeyApi.getPasskeys).mockResolvedValue({
        data: { success: true, data: [] },
      } as any)
      vi.mocked(userApi.getSessions).mockResolvedValue({ data: { success: true, data: [] } } as any)

      const wrapper = mount(UserSecurity, mountOptions)
      await flushPromises()

      const vm = wrapper.vm as any
      vm.totpEnabled = true
      vm.disableTotpCode = '123456'

      await vm.confirmDisableTotp()
      await flushPromises()

      expect(mockMessage.success).toHaveBeenCalledWith('TOTP 已禁用')
      expect(vm.totpEnabled).toBe(false)
      expect(vm.showDisableTotpModal).toBe(false)
    })

    it('shows error on TOTP disable failure', async () => {
      const { userApi, totpApi } = await import('@/api/auth')
      vi.mocked(userApi.getMe).mockResolvedValue({ data: { success: true, data: mockUser } } as any)
      vi.mocked(totpApi.delete).mockRejectedValue({
        response: { data: { message: 'Disable failed' } },
      })

      const wrapper = mount(UserSecurity, mountOptions)
      await flushPromises()

      const vm = wrapper.vm as any
      vm.disableTotpCode = '123456'

      await vm.confirmDisableTotp()
      await flushPromises()

      expect(mockMessage.error).toHaveBeenCalledWith('Disable failed')
    })
  })

  describe('passkeys loading', () => {
    it('loads passkeys successfully', async () => {
      const { userApi, passkeyApi } = await import('@/api/auth')
      vi.mocked(userApi.getMe).mockResolvedValue({ data: { success: true, data: mockUser } } as any)
      const mockPasskeys = [
        { id: '1', name: 'Test Passkey', createdAt: '2024-01-01', lastUsedAt: null },
      ]
      vi.mocked(passkeyApi.getPasskeys).mockResolvedValue({
        data: { success: true, data: mockPasskeys },
      } as any)
      vi.mocked(userApi.getSessions).mockResolvedValue({ data: { success: true, data: [] } } as any)

      const wrapper = mount(UserSecurity, mountOptions)
      await flushPromises()

      const vm = wrapper.vm as any
      await vm.loadPasskeys()
      await flushPromises()

      expect(vm.passkeys).toEqual(mockPasskeys)
      expect(vm.passkeyLoading).toBe(false)
    })

    it('shows error on passkeys load failure', async () => {
      const { userApi, passkeyApi } = await import('@/api/auth')
      vi.mocked(userApi.getMe).mockResolvedValue({ data: { success: true, data: mockUser } } as any)
      vi.mocked(passkeyApi.getPasskeys).mockRejectedValue({
        response: { data: { message: 'Load failed' } },
      })
      vi.mocked(userApi.getSessions).mockResolvedValue({ data: { success: true, data: [] } } as any)

      const wrapper = mount(UserSecurity, mountOptions)
      await flushPromises()

      const vm = wrapper.vm as any
      await vm.loadPasskeys()
      await flushPromises()

      expect(mockMessage.error).toHaveBeenCalledWith('Load failed')
      expect(vm.passkeyLoading).toBe(false)
    })
  })

  describe('passkey deletion', () => {
    it('deletes passkey successfully', async () => {
      const { userApi, passkeyApi } = await import('@/api/auth')
      vi.mocked(userApi.getMe).mockResolvedValue({ data: { success: true, data: mockUser } } as any)
      vi.mocked(passkeyApi.deletePasskey).mockResolvedValue({ data: { success: true } } as any)
      vi.mocked(passkeyApi.getPasskeys).mockResolvedValue({
        data: { success: true, data: [] },
      } as any)
      vi.mocked(userApi.getSessions).mockResolvedValue({ data: { success: true, data: [] } } as any)

      const wrapper = mount(UserSecurity, mountOptions)
      await flushPromises()

      const vm = wrapper.vm as any
      await vm.handleDeletePasskey('passkey-id')
      await flushPromises()

      expect(mockMessage.success).toHaveBeenCalledWith('通行密钥已删除')
      expect(passkeyApi.deletePasskey).toHaveBeenCalledWith('passkey-id')
    })

    it('shows error on passkey deletion failure', async () => {
      const { userApi, passkeyApi } = await import('@/api/auth')
      vi.mocked(userApi.getMe).mockResolvedValue({ data: { success: true, data: mockUser } } as any)
      vi.mocked(passkeyApi.deletePasskey).mockRejectedValue({
        response: { data: { message: 'Delete failed' } },
      })
      vi.mocked(userApi.getSessions).mockResolvedValue({ data: { success: true, data: [] } } as any)
      vi.mocked(passkeyApi.getPasskeys).mockResolvedValue({
        data: { success: true, data: [{ id: '1', name: 'Test', createdAt: '2024-01-01' }] },
      } as any)

      const wrapper = mount(UserSecurity, mountOptions)
      await flushPromises()

      const vm = wrapper.vm as any
      await vm.handleDeletePasskey('passkey-id')
      await flushPromises()

      expect(mockMessage.error).toHaveBeenCalledWith('Delete failed')
    })
  })

  describe('sessions management', () => {
    it('loads sessions successfully', async () => {
      const { userApi, passkeyApi } = await import('@/api/auth')
      vi.mocked(userApi.getMe).mockResolvedValue({ data: { success: true, data: mockUser } } as any)
      const mockSessions = [
        {
          id: 'session-1',
          createdAt: '2024-01-01T00:00:00Z',
          expiresAt: '2024-01-02T00:00:00Z',
        },
      ]
      vi.mocked(userApi.getSessions).mockResolvedValue({
        data: { success: true, data: mockSessions },
      } as any)
      vi.mocked(passkeyApi.getPasskeys).mockResolvedValue({
        data: { success: true, data: [] },
      } as any)

      const wrapper = mount(UserSecurity, mountOptions)
      await flushPromises()

      const vm = wrapper.vm as any
      await vm.loadSessions()
      await flushPromises()

      expect(vm.sessions).toEqual(mockSessions)
    })

    it('handles sessions load error', async () => {
      const { userApi, passkeyApi } = await import('@/api/auth')
      vi.mocked(userApi.getMe).mockResolvedValue({ data: { success: true, data: mockUser } } as any)
      vi.mocked(userApi.getSessions).mockRejectedValue(new Error('Error'))
      vi.mocked(passkeyApi.getPasskeys).mockResolvedValue({
        data: { success: true, data: [] },
      } as any)

      const wrapper = mount(UserSecurity, mountOptions)
      await flushPromises()

      const vm = wrapper.vm as any
      await vm.loadSessions()
      await flushPromises()

      expect(vm.sessions).toEqual([])
    })

    it('revokes single session successfully', async () => {
      const { userApi, passkeyApi } = await import('@/api/auth')
      vi.mocked(userApi.getMe).mockResolvedValue({ data: { success: true, data: mockUser } } as any)
      vi.mocked(userApi.deleteSession).mockResolvedValue({ data: { success: true } } as any)
      vi.mocked(userApi.getSessions).mockResolvedValue({ data: { success: true, data: [] } } as any)
      vi.mocked(passkeyApi.getPasskeys).mockResolvedValue({
        data: { success: true, data: [] },
      } as any)

      const wrapper = mount(UserSecurity, mountOptions)
      await flushPromises()

      const vm = wrapper.vm as any
      await vm.handleRevokeSession('session-id')
      await flushPromises()

      expect(mockMessage.success).toHaveBeenCalledWith('会话已撤销')
      expect(userApi.deleteSession).toHaveBeenCalledWith('session-id')
    })

    it('shows error on session revoke failure', async () => {
      const { userApi, passkeyApi } = await import('@/api/auth')
      vi.mocked(userApi.getMe).mockResolvedValue({ data: { success: true, data: mockUser } } as any)
      vi.mocked(userApi.deleteSession).mockRejectedValue({
        response: { data: { message: 'Revoke failed' } },
      })
      vi.mocked(userApi.getSessions).mockResolvedValue({
        data: {
          success: true,
          data: [{ id: 'session-1', createdAt: '2024-01-01', expiresAt: '2024-01-02' }],
        },
      } as any)
      vi.mocked(passkeyApi.getPasskeys).mockResolvedValue({
        data: { success: true, data: [] },
      } as any)

      const wrapper = mount(UserSecurity, mountOptions)
      await flushPromises()

      const vm = wrapper.vm as any
      await vm.handleRevokeSession('session-id')
      await flushPromises()

      expect(mockMessage.error).toHaveBeenCalledWith('Revoke failed')
    })

    it('revokes all sessions successfully', async () => {
      const { userApi, passkeyApi } = await import('@/api/auth')
      vi.mocked(userApi.getMe).mockResolvedValue({ data: { success: true, data: mockUser } } as any)
      vi.mocked(userApi.deleteAllSessions).mockResolvedValue({ data: { success: true } } as any)
      vi.mocked(userApi.getSessions).mockResolvedValue({ data: { success: true, data: [] } } as any)
      vi.mocked(passkeyApi.getPasskeys).mockResolvedValue({
        data: { success: true, data: [] },
      } as any)

      const wrapper = mount(UserSecurity, mountOptions)
      await flushPromises()

      const vm = wrapper.vm as any
      await vm.handleRevokeAllSessions()
      await flushPromises()

      expect(mockMessage.success).toHaveBeenCalledWith('所有会话已撤销')
      expect(userApi.deleteAllSessions).toHaveBeenCalled()
      expect(vm.sessionLoading).toBe(false)
    })

    it('shows error on revoke all sessions failure', async () => {
      const { userApi, passkeyApi } = await import('@/api/auth')
      vi.mocked(userApi.getMe).mockResolvedValue({ data: { success: true, data: mockUser } } as any)
      vi.mocked(userApi.deleteAllSessions).mockRejectedValue({
        response: { data: { message: 'Revoke all failed' } },
      })
      vi.mocked(userApi.getSessions).mockResolvedValue({
        data: {
          success: true,
          data: [{ id: 'session-1', createdAt: '2024-01-01', expiresAt: '2024-01-02' }],
        },
      } as any)
      vi.mocked(passkeyApi.getPasskeys).mockResolvedValue({
        data: { success: true, data: [] },
      } as any)

      const wrapper = mount(UserSecurity, mountOptions)
      await flushPromises()

      const vm = wrapper.vm as any
      await vm.handleRevokeAllSessions()
      await flushPromises()

      expect(mockMessage.error).toHaveBeenCalledWith('Revoke all failed')
      expect(vm.sessionLoading).toBe(false)
    })
  })

  describe('error handling', () => {
    it('renders without crashing on error', async () => {
      const { userApi } = await import('@/api/auth')
      vi.mocked(userApi.getMe).mockRejectedValue(new Error('Error'))

      const wrapper = mount(UserSecurity, mountOptions)
      await flushPromises()

      expect(wrapper.find('.page-container').exists()).toBe(true)
    })
  })

  describe('initial state', () => {
    it('initializes with empty password fields', async () => {
      const wrapper = mount(UserSecurity, mountOptions)
      const vm = wrapper.vm as any

      expect(vm.oldPassword).toBe('')
      expect(vm.newPassword).toBe('')
      expect(vm.confirmPassword).toBe('')
    })

    it('initializes with loading false for password', async () => {
      const wrapper = mount(UserSecurity, mountOptions)
      const vm = wrapper.vm as any

      expect(vm.passwordLoading).toBe(false)
    })
  })
})
