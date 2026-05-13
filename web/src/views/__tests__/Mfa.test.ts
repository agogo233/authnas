import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'
import Mfa from '../Mfa.vue'

const mockAuthStore = {
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
    totpVerify: vi.fn(),
  },
}))

vi.mock('@/stores/auth', () => ({
  useAuthStore: () => mockAuthStore,
}))

vi.mock('vue-router', () => ({
  useRouter: () => mockRouter,
  useRoute: () => ({
    query: { token: 'test-token' },
    params: {},
  }),
}))

vi.mock('naive-ui', () => ({
  NCard: {
    name: 'NCard',
    props: ['title', 'style'],
    template: '<div class="n-card"><slot /></div>',
  },
  NInput: {
    name: 'NInput',
    props: ['modelValue', 'placeholder', 'size', 'maxlength', 'disabled'],
    emits: ['update:modelValue'],
    template:
      '<input :value="modelValue" :placeholder="placeholder" :maxlength="maxlength" :disabled="disabled" @input="$emit(\'update:modelValue\', $event.target.value)" />',
  },
  NButton: {
    name: 'NButton',
    props: ['type', 'loading', 'disabled'],
    emits: ['click'],
    template: '<button :disabled="disabled || loading" @click="$emit(\'click\')"><slot /></button>',
  },
  NSpace: {
    name: 'NSpace',
    props: ['justify', 'align', 'vertical'],
    template: '<div class="n-space"><slot /></div>',
  },
  NAlert: {
    name: 'NAlert',
    props: ['type'],
    template: '<div class="n-alert"><slot /></div>',
  },
  useMessage: () => mockMessage,
}))

describe('Mfa.vue', () => {
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
    it('renders MFA form correctly', () => {
      const wrapper = mount(Mfa, mountOptions)
      expect(wrapper.find('.auth-container').exists()).toBe(true)
    })

    it('renders code input', () => {
      const wrapper = mount(Mfa, mountOptions)
      const codeInput = wrapper.find('input[placeholder*="请输入 6 位验证码"]')
      expect(codeInput.exists()).toBe(true)
    })

    it('has verify button', () => {
      const wrapper = mount(Mfa, mountOptions)
      const verifyButton = wrapper.find('button')
      expect(verifyButton.text()).toContain('验证')
    })

    it('has link to login page', () => {
      const wrapper = mount(Mfa, mountOptions)
      const loginLink = wrapper.find('a[href="/login"]')
      expect(loginLink.exists()).toBe(true)
    })
  })

  describe('component state', () => {
    it('initializes with empty code', () => {
      const wrapper = mount(Mfa, mountOptions)
      const vm = wrapper.vm as any
      expect(vm.code).toBe('')
    })

    it('has mfaToken from route query', () => {
      const wrapper = mount(Mfa, mountOptions)
      const vm = wrapper.vm as any
      expect(vm.mfaToken).toBe('test-token')
    })
  })

  describe('component logic', () => {
    it('validates empty code', async () => {
      const wrapper = mount(Mfa, mountOptions)
      const vm = wrapper.vm as any

      vm.code = ''
      await vm.handleSubmit()
      await flushPromises()

      expect(vm.error).toContain('请输入验证码')
    })

    it('validates code length', async () => {
      const wrapper = mount(Mfa, mountOptions)
      const vm = wrapper.vm as any

      vm.code = '123'
      await vm.handleSubmit()
      await flushPromises()

      expect(vm.error).toContain('6 位数字')
    })

    it('calls totpVerify API with code', async () => {
      const { authApi } = await import('@/api/auth')
      vi.mocked(authApi.totpVerify).mockResolvedValue({
        data: { success: true },
      } as any)

      const wrapper = mount(Mfa, mountOptions)
      const vm = wrapper.vm as any

      vm.code = '123456'
      await vm.handleSubmit()
      await flushPromises()

      expect(authApi.totpVerify).toHaveBeenCalledWith({ token: '123456', mfaToken: 'test-token' })
    })

    it('redirects to profile on successful verification', async () => {
      const { authApi } = await import('@/api/auth')

      vi.mocked(authApi.totpVerify).mockResolvedValue({
        data: {
          success: true,
          data: {
            access_token: 'token',
            refresh_token: 'refresh',
            user: {
              id: '1',
              username: 'test',
              email: 'test@test.com',
              email_verified: true,
              approved: true,
              is_admin: false,
              created_at: '2024-01-01',
            },
          },
        },
      } as any)

      const wrapper = mount(Mfa, mountOptions)
      const vm = wrapper.vm as any

      vm.code = '123456'
      await vm.handleSubmit()
      await flushPromises()

      expect(mockRouter.push).toHaveBeenCalledWith('/profile')
    })

    it('handles verification error', async () => {
      const { authApi } = await import('@/api/auth')

      vi.mocked(authApi.totpVerify).mockRejectedValue({
        response: { data: { message: 'Invalid code' } },
      })

      const wrapper = mount(Mfa, mountOptions)
      const vm = wrapper.vm as any

      vm.code = '123456'
      await vm.handleSubmit()
      await flushPromises()

      expect(vm.error).toContain('Invalid code')
    })
  })
})
