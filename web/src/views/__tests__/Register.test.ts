import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'
import Register from '../Register.vue'

const mockRouter = {
  push: vi.fn(),
}

vi.mock('@/api/auth', () => ({
  authApi: {
    getPublicConfig: vi.fn(),
    getCsrfToken: vi.fn(),
    register: vi.fn(),
  },
}))

vi.mock('@/stores/auth', () => ({
  useAuthStore: () => ({
    isAuthenticated: false,
    setTokens: vi.fn(),
    setUser: vi.fn(),
  }),
}))

vi.mock('vue-router', () => ({
  useRouter: () => mockRouter,
}))

vi.mock('naive-ui', () => ({
  NCard: {
    name: 'NCard',
    props: ['title', 'style'],
    template: '<div class="n-card"><slot /></div>',
  },
  NForm: {
    name: 'NForm',
    emits: ['submit'],
    template: '<form @submit.prevent="$emit(\'submit\')"><slot /></form>',
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
    props: ['type', 'loading', 'disabled', 'attrType'],
    emits: ['click'],
    template:
      '<button :type="attrType" :disabled="disabled || loading" @click="$emit(\'click\')"><slot /></button>',
  },
  NSpace: {
    name: 'NSpace',
    props: ['justify', 'align'],
    template: '<div class="n-space"><slot /></div>',
  },
  NAlert: {
    name: 'NAlert',
    props: ['type'],
    template: '<div class="n-alert"><slot /></div>',
  },
  NProgress: {
    name: 'NProgress',
    props: ['percentage', 'color', 'showIndicator', 'height'],
    template: '<div class="n-progress" />',
  },
  useMessage: () => ({
    success: vi.fn(),
    error: vi.fn(),
  }),
}))

describe('Register.vue', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    localStorage.clear()
    setActivePinia(createPinia())
  })

  async function mockDefaultPolicy() {
    const { authApi } = await import('@/api/auth')
    vi.mocked(authApi.getPublicConfig).mockResolvedValue({
      data: {
        success: true,
        data: {
          app_name: 'AuthNas',
          signup_requires_approval: false,
          email_verification: false,
          mfa_required: false,
          password_min_length: 8,
          password_strength: 3,
          default_redirect: 'http://localhost:8080',
          contact_email: '',
        },
      },
    } as any)
    vi.mocked(authApi.getCsrfToken).mockResolvedValue({
      data: { success: true, data: { token: 'csrf-token', expiresAt: '2026-01-01T00:00:00Z' } },
    } as any)
  }

  const mountOptions = {
    global: {
      stubs: {
        'router-link': {
          name: 'RouterLink',
          props: ['to'],
          template: '<a :href="to"><slot /></a>',
        },
      },
    },
  }

  describe('rendering', () => {
    it('renders registration form correctly', async () => {
      await mockDefaultPolicy()
      const wrapper = mount(Register, mountOptions)
      await flushPromises()
      expect(wrapper.find('.auth-container').exists()).toBe(true)
    })

    it('renders email input', async () => {
      await mockDefaultPolicy()
      const wrapper = mount(Register, mountOptions)
      await flushPromises()
      const emailInput = wrapper.find('input[placeholder*="请输入邮箱"]')
      expect(emailInput.exists()).toBe(true)
    })

    it('renders username input', async () => {
      await mockDefaultPolicy()
      const wrapper = mount(Register, mountOptions)
      await flushPromises()
      const usernameInput = wrapper.find('input[placeholder*="请输入用户名"]')
      expect(usernameInput.exists()).toBe(true)
    })

    it('renders password input', async () => {
      await mockDefaultPolicy()
      const wrapper = mount(Register, mountOptions)
      await flushPromises()
      const passwordInput = wrapper.find('input[type="password"]')
      expect(passwordInput.exists()).toBe(true)
    })

    it('has invitation code field', async () => {
      await mockDefaultPolicy()
      const wrapper = mount(Register, mountOptions)
      await flushPromises()
      const invitationInput = wrapper.find('input[placeholder*="请输入邀请码"]')
      expect(invitationInput.exists()).toBe(true)
    })

    it('has link to login page', async () => {
      await mockDefaultPolicy()
      const wrapper = mount(Register, mountOptions)
      await flushPromises()
      const loginLink = wrapper.find('a[href="/login"]')
      expect(loginLink.exists()).toBe(true)
    })
  })

  describe('component state', () => {
    it('initializes with empty fields', async () => {
      await mockDefaultPolicy()
      const wrapper = mount(Register, mountOptions)
      await flushPromises()
      const vm = wrapper.vm as any
      expect(vm.email).toBe('')
      expect(vm.username).toBe('')
      expect(vm.password).toBe('')
      expect(vm.invitationCode).toBe('')
    })

    it('initializes with no error', async () => {
      await mockDefaultPolicy()
      const wrapper = mount(Register, mountOptions)
      await flushPromises()
      const vm = wrapper.vm as any
      expect(vm.error).toBe('')
    })
  })

  describe('component logic', () => {
    it('validates required fields', async () => {
      await mockDefaultPolicy()
      const wrapper = mount(Register, mountOptions)
      await flushPromises()
      const vm = wrapper.vm as any

      await vm.handleRegister()
      await flushPromises()

      expect(vm.error).toContain('请填写所有必填字段')
    })

    it('validates email field', async () => {
      await mockDefaultPolicy()
      const wrapper = mount(Register, mountOptions)
      await flushPromises()
      const vm = wrapper.vm as any

      vm.email = 'test@example.com'
      vm.username = 'testuser'
      vm.password = ''
      await vm.handleRegister()
      await flushPromises()

      expect(vm.error).toContain('请填写所有必填字段')
    })

    it('rejects short passwords before strength validation', async () => {
      await mockDefaultPolicy()
      const wrapper = mount(Register, mountOptions)
      await flushPromises()
      const vm = wrapper.vm as any

      vm.email = 'test@example.com'
      vm.username = 'testuser'
      vm.password = 'weak'
      await vm.handleRegister()
      await flushPromises()

      expect(vm.error).toContain('密码长度不能少于 8 位')
    })

    it('calls register API with correct data', async () => {
      const { authApi } = await import('@/api/auth')
      await mockDefaultPolicy()
      vi.mocked(authApi.register).mockResolvedValue({
        data: { success: true },
      } as any)

      const wrapper = mount(Register, mountOptions)
      await flushPromises()
      const vm = wrapper.vm as any

      vm.email = 'test@example.com'
      vm.username = 'testuser'
      vm.password = 'Password123!'
      await vm.handleRegister()
      await flushPromises()

      expect(authApi.register).toHaveBeenCalledWith({
        email: 'test@example.com',
        username: 'testuser',
        password: 'Password123!',
        inviteId: undefined,
      })
    })

    it('redirects to login on successful registration', async () => {
      const { authApi } = await import('@/api/auth')
      await mockDefaultPolicy()
      vi.mocked(authApi.register).mockResolvedValue({
        data: { success: true },
      } as any)

      const wrapper = mount(Register, mountOptions)
      await flushPromises()
      const vm = wrapper.vm as any

      vm.email = 'test@example.com'
      vm.username = 'testuser'
      vm.password = 'Password123!'
      await vm.handleRegister()
      await flushPromises()

      expect(mockRouter.push).toHaveBeenCalledWith('/login')
    })

    it('handles registration error', async () => {
      const { authApi } = await import('@/api/auth')
      await mockDefaultPolicy()
      vi.mocked(authApi.register).mockRejectedValue({
        response: { data: { message: 'User already exists' } },
      })

      const wrapper = mount(Register, mountOptions)
      await flushPromises()
      const vm = wrapper.vm as any

      vm.email = 'test@example.com'
      vm.username = 'testuser'
      vm.password = 'Password123!'
      await vm.handleRegister()
      await flushPromises()

      expect(vm.error).toContain('User already exists')
    })

    it('passes invitation code if provided', async () => {
      const { authApi } = await import('@/api/auth')
      await mockDefaultPolicy()
      vi.mocked(authApi.register).mockResolvedValue({
        data: { success: true },
      } as any)

      const wrapper = mount(Register, mountOptions)
      await flushPromises()
      const vm = wrapper.vm as any

      vm.email = 'test@example.com'
      vm.username = 'testuser'
      vm.password = 'Password123!'
      vm.invitationCode = 'INVITE123'
      await vm.handleRegister()
      await flushPromises()

      expect(authApi.register).toHaveBeenCalledWith({
        email: 'test@example.com',
        username: 'testuser',
        password: 'Password123!',
        inviteId: 'INVITE123',
      })
    })
  })

  describe('password strength calculation', () => {
    it('calculates very weak for short password', async () => {
      await mockDefaultPolicy()
      const wrapper = mount(Register, mountOptions)
      await flushPromises()
      const vm = wrapper.vm as any

      const result = vm.calculatePasswordStrength('abc')
      expect(result.label).toContain('非常弱')
    })

    it('calculates weak for 8+ char password', async () => {
      await mockDefaultPolicy()
      const wrapper = mount(Register, mountOptions)
      await flushPromises()
      const vm = wrapper.vm as any

      const result = vm.calculatePasswordStrength('password')
      expect(result.label).toContain('弱')
    })

    it('calculates good for password with special chars', async () => {
      await mockDefaultPolicy()
      const wrapper = mount(Register, mountOptions)
      await flushPromises()
      const vm = wrapper.vm as any

      const result = vm.calculatePasswordStrength('Pas1!')
      expect(result.label).toContain('良好')
    })
  })
})
