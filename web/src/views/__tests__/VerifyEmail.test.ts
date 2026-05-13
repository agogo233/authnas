import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'
import VerifyEmail from '../VerifyEmail.vue'

const mockMessage = {
  success: vi.fn(),
  error: vi.fn(),
}

const mockRouter = {
  push: vi.fn(),
}

const mockRoute = {
  query: {} as Record<string, string>,
  params: {},
}

vi.mock('@/api/auth', () => ({
  authApi: {
    verifyEmail: vi.fn(),
    sendVerifyEmail: vi.fn(),
  },
}))

vi.mock('vue-router', () => ({
  useRouter: () => mockRouter,
  useRoute: () => mockRoute,
}))

vi.mock('naive-ui', () => ({
  NCard: {
    name: 'NCard',
    props: ['title', 'bordered', 'style'],
    template: '<div class="n-card"><slot /></div>',
  },
  NButton: {
    name: 'NButton',
    props: ['type', 'loading'],
    emits: ['click'],
    template: '<button :disabled="loading" @click="$emit(\'click\')"><slot /></button>',
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
  NResult: {
    name: 'NResult',
    props: ['status', 'title', 'description'],
    template: '<div class="n-result"><slot name="footer" /></div>',
  },
  useMessage: () => mockMessage,
}))

describe('VerifyEmail.vue', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    localStorage.clear()
    mockRoute.query = {}
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
    it('renders verify container correctly', () => {
      const wrapper = mount(VerifyEmail, mountOptions)
      expect(wrapper.find('.auth-container').exists()).toBe(true)
    })
  })

  describe('handleResend', () => {
    it('calls sendVerifyEmail API when email is available', async () => {
      const { authApi } = await import('@/api/auth')
      vi.mocked(authApi.sendVerifyEmail).mockResolvedValue({
        data: { success: true },
      } as any)
      mockRoute.query = { email: 'test@example.com' }

      const wrapper = mount(VerifyEmail, mountOptions)
      const vm = wrapper.vm as any

      await vm.handleResend()
      await flushPromises()

      expect(authApi.sendVerifyEmail).toHaveBeenCalledWith({ email: 'test@example.com' })
    })

    it('shows success message after resend', async () => {
      const { authApi } = await import('@/api/auth')
      vi.mocked(authApi.sendVerifyEmail).mockResolvedValue({
        data: { success: true },
      } as any)
      mockRoute.query = { email: 'test@example.com' }

      const wrapper = mount(VerifyEmail, mountOptions)
      const vm = wrapper.vm as any

      await vm.handleResend()
      await flushPromises()

      expect(vm.resendSuccess).toBe(true)
    })

    it('shows error when email is not available', async () => {
      mockRoute.query = {}

      const wrapper = mount(VerifyEmail, mountOptions)
      const vm = wrapper.vm as any

      await vm.handleResend()
      await flushPromises()

      expect(mockMessage.error).toHaveBeenCalledWith('无法获取邮箱地址')
    })

    it('handles resend error', async () => {
      const { authApi } = await import('@/api/auth')
      vi.mocked(authApi.sendVerifyEmail).mockRejectedValue({
        response: { data: { message: 'Failed to send' } },
      })
      mockRoute.query = { email: 'test@example.com' }

      const wrapper = mount(VerifyEmail, mountOptions)
      const vm = wrapper.vm as any

      await vm.handleResend()
      await flushPromises()

      expect(mockMessage.error).toHaveBeenCalled()
    })
  })
})
