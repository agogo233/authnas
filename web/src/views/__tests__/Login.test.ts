import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'
import Login from '../Login.vue'

const mockAuthStore = {
  isAuthenticated: false,
  setTokens: vi.fn(),
  setUser: vi.fn(),
}

const mockRouter = {
  push: vi.fn(),
}

const mockMessage = {
  success: vi.fn(),
  error: vi.fn(),
}

vi.mock('@/api/auth', () => ({
  authApi: {
    login: vi.fn(),
    passkeyStart: vi.fn(),
    passkeyEnd: vi.fn(),
  },
}))

vi.mock('@/stores/auth', () => ({
  useAuthStore: () => mockAuthStore,
}))

vi.mock('vue-router', () => ({
  useRouter: () => mockRouter,
  useRoute: () => ({
    query: {},
  }),
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
    props: ['type', 'loading', 'disabled', 'attrType', 'block'],
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
  NCheckbox: {
    name: 'NCheckbox',
    props: ['checked'],
    emits: ['update:checked'],
    template:
      '<input type="checkbox" :checked="checked" @change="$emit(\'update:checked\', $event.target.checked)" />',
  },
  useMessage: () => mockMessage,
}))

describe('Login.vue', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    localStorage.clear()
    setActivePinia(createPinia())
  })

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
    it('renders login form correctly', () => {
      const wrapper = mount(Login, mountOptions)
      expect(wrapper.find('.auth-container').exists()).toBe(true)
    })

    it('renders username input', () => {
      const wrapper = mount(Login, mountOptions)
      const usernameInput = wrapper.find('input[placeholder*="请输入用户名"]')
      expect(usernameInput.exists()).toBe(true)
    })

    it('renders password input', () => {
      const wrapper = mount(Login, mountOptions)
      const passwordInput = wrapper.find('input[type="password"]')
      expect(passwordInput.exists()).toBe(true)
    })

    it('has a login button', () => {
      const wrapper = mount(Login, mountOptions)
      const loginButton = wrapper.find('button[type="submit"]')
      expect(loginButton.exists()).toBe(true)
    })

    it('has a link to register page', () => {
      const wrapper = mount(Login, mountOptions)
      const registerLink = wrapper.find('a[href="/register"]')
      expect(registerLink.exists()).toBe(true)
    })

    it('has a link to reset password page', () => {
      const wrapper = mount(Login, mountOptions)
      const resetLink = wrapper.find('a[href="/reset-password"]')
      expect(resetLink.exists()).toBe(true)
    })
  })

  describe('component state', () => {
    it('initializes with empty username and password', () => {
      const wrapper = mount(Login, mountOptions)
      const vm = wrapper.vm as any
      expect(vm.username).toBe('')
      expect(vm.password).toBe('')
    })

    it('has remember me unchecked by default', () => {
      const wrapper = mount(Login, mountOptions)
      const vm = wrapper.vm as any
      expect(vm.rememberMe).toBe(false)
    })

    it('has no error by default', () => {
      const wrapper = mount(Login, mountOptions)
      const vm = wrapper.vm as any
      expect(vm.error).toBe('')
    })
  })

  describe('component logic', () => {
    it('validates empty username and password', async () => {
      const wrapper = mount(Login, mountOptions)
      const vm = wrapper.vm as any

      await vm.handleLogin()
      await flushPromises()

      expect(vm.error).toContain('请输入用户名和密码')
    })

    it('validates empty password only', async () => {
      const wrapper = mount(Login, mountOptions)
      const vm = wrapper.vm as any

      vm.username = 'testuser'
      await vm.handleLogin()
      await flushPromises()

      expect(vm.error).toContain('请输入用户名和密码')
    })

    it('calls login API with correct data', async () => {
      const { authApi } = await import('@/api/auth')
      vi.mocked(authApi.login).mockResolvedValue({
        data: {
          success: true,
          data: {
            accessToken: 'token',
            refreshToken: 'refresh',
            expiresAt: '2025-01-01T00:00:00Z',
            user: {
              id: '1',
              username: 'test',
              email: 'test@test.com',
              emailVerified: true,
              approved: true,
              isAdmin: false,
              createdAt: '2024-01-01',
            },
          },
        },
      } as any)

      const wrapper = mount(Login, mountOptions)
      const vm = wrapper.vm as any

      vm.username = 'testuser'
      vm.password = 'password123'
      await vm.handleLogin()
      await flushPromises()

      expect(authApi.login).toHaveBeenCalledWith({ input: 'testuser', password: 'password123' })
    })

    it('stores tokens on successful login', async () => {
      const { authApi } = await import('@/api/auth')
      vi.mocked(authApi.login).mockResolvedValue({
        data: {
          success: true,
          data: {
            accessToken: 'token',
            refreshToken: 'refresh',
            expiresAt: '2025-01-01T00:00:00Z',
            user: {
              id: '1',
              username: 'test',
              email: 'test@test.com',
              emailVerified: true,
              approved: true,
              isAdmin: false,
              createdAt: '2024-01-01',
            },
          },
        },
      } as any)

      const wrapper = mount(Login, mountOptions)
      const vm = wrapper.vm as any

      vm.username = 'testuser'
      vm.password = 'password123'
      await vm.handleLogin()
      await flushPromises()

      expect(mockAuthStore.setTokens).toHaveBeenCalledWith('token', '2025-01-01T00:00:00Z')
      expect(mockAuthStore.setUser).toHaveBeenCalled()
    })

    it('redirects to profile on successful login', async () => {
      const { authApi } = await import('@/api/auth')
      vi.mocked(authApi.login).mockResolvedValue({
        data: {
          success: true,
          data: {
            accessToken: 'token',
            refreshToken: 'refresh',
            expiresAt: '2025-01-01T00:00:00Z',
            user: {
              id: '1',
              username: 'test',
              email: 'test@test.com',
              emailVerified: true,
              approved: true,
              isAdmin: false,
              createdAt: '2024-01-01',
            },
          },
        },
      } as any)

      const wrapper = mount(Login, mountOptions)
      const vm = wrapper.vm as any

      vm.username = 'testuser'
      vm.password = 'password123'
      await vm.handleLogin()
      await flushPromises()

      expect(mockRouter.push).toHaveBeenCalledWith('/profile')
    })

    it('shows error on login failure', async () => {
      const { authApi } = await import('@/api/auth')
      vi.mocked(authApi.login).mockRejectedValue({
        response: { data: { message: 'Invalid credentials' } },
      })

      const wrapper = mount(Login, mountOptions)
      const vm = wrapper.vm as any

      vm.username = 'testuser'
      vm.password = 'wrongpassword'
      await vm.handleLogin()
      await flushPromises()

      expect(vm.error).toContain('Invalid credentials')
    })
  })

  describe('remember me functionality', () => {
    it('checkbox is unchecked by default when no remembered username', () => {
      const wrapper = mount(Login, mountOptions)
      const checkbox = wrapper.find('input[type="checkbox"]')
      expect((checkbox.element as HTMLInputElement).checked).toBe(false)
    })

    it('loads remembered username from localStorage', async () => {
      localStorage.setItem('remembered_username', 'storeduser')

      const wrapper = mount(Login, mountOptions)
      const vm = wrapper.vm as any

      expect(vm.username).toBe('storeduser')
    })
  })

  describe('mfa required flow', () => {
    it('redirects to MFA when mfaRequired is true', async () => {
      const { authApi } = await import('@/api/auth')
      vi.mocked(authApi.login).mockResolvedValue({
        data: {
          success: true,
          data: {
            mfaRequired: true,
            mfaToken: 'mfa-token-123',
          },
        },
      } as any)

      const wrapper = mount(Login, mountOptions)
      const vm = wrapper.vm as any

      vm.username = 'testuser'
      vm.password = 'password123'
      await vm.handleLogin()
      await flushPromises()

      expect(mockRouter.push).toHaveBeenCalledWith({
        path: '/mfa',
        query: { token: 'mfa-token-123' },
      })
    })

    it('handles mfaRequired with missing mfaToken', async () => {
      const { authApi } = await import('@/api/auth')
      vi.mocked(authApi.login).mockResolvedValue({
        data: {
          success: true,
          data: { mfaRequired: true },
        },
      } as any)

      const wrapper = mount(Login, mountOptions)
      const vm = wrapper.vm as any

      vm.username = 'testuser'
      vm.password = 'password123'
      await vm.handleLogin()
      await flushPromises()

      expect(mockRouter.push).toHaveBeenCalledWith({ path: '/mfa', query: { token: '' } })
    })
  })

  describe('passkey login', () => {
    it('validates username before passkey login', async () => {
      const wrapper = mount(Login, mountOptions)
      const vm = wrapper.vm as any

      vm.username = ''
      await vm.handlePasskeyLogin()
      await flushPromises()

      expect(vm.error).toContain('请先输入用户名')
    })

    it('passkey login clears loading on error', async () => {
      const { authApi } = await import('@/api/auth')
      vi.mocked(authApi.passkeyStart).mockRejectedValue(new Error('Network error'))

      const wrapper = mount(Login, mountOptions)
      const vm = wrapper.vm as any

      vm.username = 'testuser'
      await vm.handlePasskeyLogin()
      await flushPromises()

      expect(vm.passkeyLoading).toBe(false)
    })
  })

  describe('login error handling', () => {
    it('handles network error without response', async () => {
      const { authApi } = await import('@/api/auth')
      vi.mocked(authApi.login).mockRejectedValue(new Error('Network failure'))

      const wrapper = mount(Login, mountOptions)
      const vm = wrapper.vm as any

      vm.username = 'testuser'
      vm.password = 'password123'
      await vm.handleLogin()
      await flushPromises()

      expect(vm.error).toContain('登录失败')
    })
  })
})
