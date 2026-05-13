import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import ResetPassword from '../ResetPassword.vue'

const mockMessage = {
  success: vi.fn(),
  error: vi.fn(),
}

vi.mock('@/api/auth', () => ({
  authApi: {
    forgotPassword: vi.fn(),
  },
}))

vi.mock('vue-router', () => ({
  useRouter: () => ({
    push: vi.fn(),
  }),
}))

vi.mock('naive-ui', () => ({
  NCard: {
    name: 'NCard',
    props: ['title', 'bordered', 'style'],
    template:
      '<div class="n-card"><div class="n-card-header"><slot name="header" /></div><slot /></div>',
  },
  NForm: {
    name: 'NForm',
    emits: ['submit'],
    template: '<form @submit.prevent="$emit(\'submit\')"><slot /></form>',
  },
  NFormItem: {
    name: 'NFormItem',
    props: ['label', 'path'],
    template: '<div class="n-form-item"><label v-if="label">{{ label }}</label><slot /></div>',
  },
  NInput: {
    name: 'NInput',
    props: ['modelValue', 'placeholder', 'size'],
    emits: ['update:modelValue'],
    template:
      '<input :value="modelValue" :placeholder="placeholder" @input="$emit(\'update:modelValue\', $event.target.value)" />',
  },
  NButton: {
    name: 'NButton',
    props: ['type', 'loading', 'attrType', 'block', 'size'],
    emits: ['click'],
    template:
      '<button :type="attrType" :disabled="loading" @click="$emit(\'click\')"><slot /></button>',
  },
  NAlert: {
    name: 'NAlert',
    props: ['type'],
    template: '<div class="n-alert"><slot /></div>',
  },
  useMessage: () => mockMessage,
}))

describe('ResetPassword.vue', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    localStorage.clear()
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
    it('renders reset password form correctly', () => {
      const wrapper = mount(ResetPassword, mountOptions)
      expect(wrapper.find('.auth-container').exists()).toBe(true)
    })

    it('renders email input', () => {
      const wrapper = mount(ResetPassword, mountOptions)
      const emailInput = wrapper.find('input[placeholder="请输入您的邮箱"]')
      expect(emailInput.exists()).toBe(true)
    })

    it('has a submit button', () => {
      const wrapper = mount(ResetPassword, mountOptions)
      const submitButton = wrapper.find('button[type="submit"]')
      expect(submitButton.exists()).toBe(true)
      expect(submitButton.text()).toContain('发送重置邮件')
    })

    it('has link to login page', () => {
      const wrapper = mount(ResetPassword, mountOptions)
      const loginLink = wrapper.find('a[href="/login"]')
      expect(loginLink.exists()).toBe(true)
    })

    it('renders page header', () => {
      const wrapper = mount(ResetPassword, mountOptions)
      expect(wrapper.find('.page-header h1').text()).toBe('重置密码')
      expect(wrapper.find('.page-header p').text()).toBe('请输入您的注册邮箱')
    })
  })

  describe('component state', () => {
    it('initializes with empty email', () => {
      const wrapper = mount(ResetPassword, mountOptions)
      const vm = wrapper.vm as any
      expect(vm.email).toBe('')
    })

    it('initializes with no error', () => {
      const wrapper = mount(ResetPassword, mountOptions)
      const vm = wrapper.vm as any
      expect(vm.error).toBe('')
    })

    it('initializes with success as false', () => {
      const wrapper = mount(ResetPassword, mountOptions)
      const vm = wrapper.vm as any
      expect(vm.success).toBe(false)
    })

    it('initializes with loading as false', () => {
      const wrapper = mount(ResetPassword, mountOptions)
      const vm = wrapper.vm as any
      expect(vm.loading).toBe(false)
    })
  })

  describe('form validation', () => {
    it('shows error when email is empty', async () => {
      const wrapper = mount(ResetPassword, mountOptions)
      const vm = wrapper.vm as any

      vm.email = ''
      await vm.handleSubmit()
      await flushPromises()

      expect(vm.error).toBe('请输入您的邮箱')
    })

    it('does not call API when email is empty', async () => {
      const { authApi } = await import('@/api/auth')
      const wrapper = mount(ResetPassword, mountOptions)
      const vm = wrapper.vm as any

      vm.email = ''
      await vm.handleSubmit()
      await flushPromises()

      expect(authApi.forgotPassword).not.toHaveBeenCalled()
    })
  })

  describe('form submission', () => {
    it('calls forgotPassword API with email', async () => {
      const { authApi } = await import('@/api/auth')
      vi.mocked(authApi.forgotPassword).mockResolvedValue({
        data: { success: true },
      } as any)

      const wrapper = mount(ResetPassword, mountOptions)
      const vm = wrapper.vm as any

      vm.email = 'test@example.com'
      await vm.handleSubmit()
      await flushPromises()

      expect(authApi.forgotPassword).toHaveBeenCalledWith({ email: 'test@example.com' })
    })

    it('shows success message on successful request', async () => {
      const { authApi } = await import('@/api/auth')
      vi.mocked(authApi.forgotPassword).mockResolvedValue({
        data: { success: true },
      } as any)

      const wrapper = mount(ResetPassword, mountOptions)
      const vm = wrapper.vm as any

      vm.email = 'test@example.com'
      await vm.handleSubmit()
      await flushPromises()

      expect(vm.success).toBe(true)
      expect(mockMessage.success).toHaveBeenCalledWith('密码重置邮件已发送')
    })

    it('shows success alert on successful request', async () => {
      const { authApi } = await import('@/api/auth')
      vi.mocked(authApi.forgotPassword).mockResolvedValue({
        data: { success: true },
      } as any)

      const wrapper = mount(ResetPassword, mountOptions)
      const vm = wrapper.vm as any

      vm.email = 'test@example.com'
      await vm.handleSubmit()
      await flushPromises()

      const alert = wrapper.find('.n-alert')
      expect(alert.exists()).toBe(true)
      expect(alert.text()).toContain('密码重置邮件已发送')
    })

    it('shows error message on API failure', async () => {
      const { authApi } = await import('@/api/auth')
      vi.mocked(authApi.forgotPassword).mockRejectedValue({
        response: { data: { message: 'Email not found' } },
      })

      const wrapper = mount(ResetPassword, mountOptions)
      const vm = wrapper.vm as any

      vm.email = 'nonexistent@example.com'
      await vm.handleSubmit()
      await flushPromises()

      expect(vm.error).toContain('Email not found')
    })

    it('shows generic error on network failure', async () => {
      const { authApi } = await import('@/api/auth')
      vi.mocked(authApi.forgotPassword).mockRejectedValue(new Error('Network error'))

      const wrapper = mount(ResetPassword, mountOptions)
      const vm = wrapper.vm as any

      vm.email = 'test@example.com'
      await vm.handleSubmit()
      await flushPromises()

      expect(vm.error).toBe('发送密码重置邮件失败')
    })

    it('clears previous error on new submission', async () => {
      const { authApi } = await import('@/api/auth')
      vi.mocked(authApi.forgotPassword)
        .mockRejectedValueOnce({ response: { data: { message: 'First error' } } })
        .mockResolvedValueOnce({ data: { success: true } })

      const wrapper = mount(ResetPassword, mountOptions)
      const vm = wrapper.vm as any

      vm.email = 'test@example.com'
      await vm.handleSubmit()
      await flushPromises()
      expect(vm.error).toContain('First error')

      vm.email = 'another@example.com'
      await vm.handleSubmit()
      await flushPromises()
      expect(vm.error).toBe('')
    })
  })

  describe('loading state', () => {
    it('sets loading to true during submission', async () => {
      const { authApi } = await import('@/api/auth')
      vi.mocked(authApi.forgotPassword).mockImplementation(
        () => new Promise((resolve) => setTimeout(() => resolve({ data: { success: true } }), 100))
      )

      const wrapper = mount(ResetPassword, mountOptions)
      const vm = wrapper.vm as any

      vm.email = 'test@example.com'
      const submitPromise = vm.handleSubmit()

      await flushPromises()
      expect(vm.loading).toBe(true)

      await submitPromise
      await flushPromises()
      expect(vm.loading).toBe(false)
    })

    it('disables button during loading', async () => {
      const { authApi } = await import('@/api/auth')
      vi.mocked(authApi.forgotPassword).mockImplementation(
        () => new Promise((resolve) => setTimeout(() => resolve({ data: { success: true } }), 100))
      )

      const wrapper = mount(ResetPassword, mountOptions)
      const vm = wrapper.vm as any

      vm.email = 'test@example.com'
      const submitPromise = vm.handleSubmit()
      await flushPromises()

      const button = wrapper.find('button[type="submit"]')
      expect((button.element as HTMLButtonElement).disabled).toBe(true)

      await submitPromise
    })
  })
})
