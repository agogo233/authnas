import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'
import UserProfile from '../user/Profile.vue'

const mockMessage = {
  success: vi.fn(),
  error: vi.fn(),
  warning: vi.fn(),
  info: vi.fn(),
}

vi.mock('@/api/auth', () => ({
  userApi: {
    getMe: vi.fn(),
    updateMe: vi.fn(),
  },
  authApi: {
    sendVerifyEmail: vi.fn(),
  },
}))

vi.mock('@/stores/auth', () => ({
  useAuthStore: () => ({
    setUser: vi.fn(),
  }),
}))

vi.mock('naive-ui', () => ({
  NCard: {
    name: 'NCard',
    props: ['title', 'bordered'],
    template: '<div class="n-card"><slot /></div>',
  },
  NDescriptions: {
    name: 'NDescriptions',
    props: ['column', 'bordered'],
    template: '<div class="n-descriptions"><slot /></div>',
  },
  NDescriptionsItem: {
    name: 'NDescriptionsItem',
    props: ['label'],
    template:
      '<div class="n-descriptions-item"><span class="label">{{ label }}</span><slot /></div>',
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
    props: ['modelValue', 'disabled', 'placeholder'],
    emits: ['update:modelValue'],
    template:
      '<input :value="modelValue" :disabled="disabled" :placeholder="placeholder" @input="$emit(\'update:modelValue\', $event.target.value)" />',
  },
  useMessage: () => mockMessage,
}))

describe('UserProfile.vue', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    setActivePinia(createPinia())
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
    createdAt: '2024-01-15T00:00:00Z',
  }

  const mountOptions = {
    global: {
      stubs: {},
    },
  }

  describe('rendering', () => {
    it('renders profile page', async () => {
      const { userApi } = await import('@/api/auth')
      vi.mocked(userApi.getMe).mockResolvedValue({ data: { success: true, data: mockUser } } as any)

      const wrapper = mount(UserProfile, mountOptions)
      await flushPromises()

      expect(wrapper.find('.page-container').exists()).toBe(true)
      expect(wrapper.find('.n-card').exists()).toBe(true)
    })

    it('displays user profile information', async () => {
      const { userApi } = await import('@/api/auth')
      vi.mocked(userApi.getMe).mockResolvedValue({ data: { success: true, data: mockUser } } as any)

      const wrapper = mount(UserProfile, mountOptions)
      await flushPromises()

      expect(wrapper.text()).toContain('testuser')
      expect(wrapper.text()).toContain('test@example.com')
      expect(wrapper.text()).toContain('Test User')
    })

    it('shows edit button', async () => {
      const { userApi } = await import('@/api/auth')
      vi.mocked(userApi.getMe).mockResolvedValue({ data: { success: true, data: mockUser } } as any)

      const wrapper = mount(UserProfile, mountOptions)
      await flushPromises()

      const buttons = wrapper.findAll('button')
      const editButton = buttons.find((b) => b.text().includes('编辑资料'))
      expect(editButton?.exists()).toBe(true)
    })
  })

  describe('profile loading', () => {
    it('fetches user profile on mount', async () => {
      const { userApi } = await import('@/api/auth')
      vi.mocked(userApi.getMe).mockResolvedValue({ data: { success: true, data: mockUser } } as any)

      mount(UserProfile, mountOptions)
      await flushPromises()

      expect(userApi.getMe).toHaveBeenCalled()
    })

    it('stores user data after fetch', async () => {
      const { userApi } = await import('@/api/auth')
      vi.mocked(userApi.getMe).mockResolvedValue({ data: { success: true, data: mockUser } } as any)

      const wrapper = mount(UserProfile, mountOptions)
      await flushPromises()

      const vm = wrapper.vm as any
      expect(vm.user).toBeTruthy()
      expect(vm.user.username).toBe('testuser')
    })
  })

  describe('edit mode', () => {
    it('opens edit form when clicking edit button', async () => {
      const { userApi } = await import('@/api/auth')
      vi.mocked(userApi.getMe).mockResolvedValue({ data: { success: true, data: mockUser } } as any)

      const wrapper = mount(UserProfile, mountOptions)
      await flushPromises()

      const editButton = wrapper.findAll('button').find((b) => b.text().includes('编辑资料'))
      await editButton!.trigger('click')
      await flushPromises()

      const vm = wrapper.vm as any
      expect(vm.editing).toBe(true)
    })

    it('prefills edit form with current user data', async () => {
      const { userApi } = await import('@/api/auth')
      vi.mocked(userApi.getMe).mockResolvedValue({ data: { success: true, data: mockUser } } as any)

      const wrapper = mount(UserProfile, mountOptions)
      await flushPromises()

      const vm = wrapper.vm as any
      vm.startEdit()
      await flushPromises()

      expect(vm.editForm.name).toBe('Test User')
      expect(vm.editForm.email).toBe('test@example.com')
    })

    it('closes edit form on cancel', async () => {
      const { userApi } = await import('@/api/auth')
      vi.mocked(userApi.getMe).mockResolvedValue({ data: { success: true, data: mockUser } } as any)

      const wrapper = mount(UserProfile, mountOptions)
      await flushPromises()

      const vm = wrapper.vm as any
      vm.startEdit()
      await flushPromises()

      vm.cancelEdit()
      await flushPromises()

      expect(vm.editing).toBe(false)
    })
  })

  describe('profile update', () => {
    it('calls update API when saving changes', async () => {
      const { userApi } = await import('@/api/auth')
      vi.mocked(userApi.getMe).mockResolvedValue({ data: { success: true, data: mockUser } } as any)
      vi.mocked(userApi.updateMe).mockResolvedValue({
        data: { success: true, data: { ...mockUser, name: 'Updated Name' } },
      } as any)

      const wrapper = mount(UserProfile, mountOptions)
      await flushPromises()

      const vm = wrapper.vm as any
      vm.startEdit()
      await flushPromises()

      vm.editForm.name = 'Updated Name'
      await vm.saveEdit()
      await flushPromises()

      expect(userApi.updateMe).toHaveBeenCalled()
      expect(mockMessage.success).toHaveBeenCalledWith('个人信息更新成功')
    })

    it('shows error when email is empty', async () => {
      const { userApi } = await import('@/api/auth')
      vi.mocked(userApi.getMe).mockResolvedValue({ data: { success: true, data: mockUser } } as any)

      const wrapper = mount(UserProfile, mountOptions)
      await flushPromises()

      const vm = wrapper.vm as any
      vm.startEdit()
      await flushPromises()

      vm.editForm.email = ''
      await vm.saveEdit()
      await flushPromises()

      expect(mockMessage.error).toHaveBeenCalledWith('邮箱不能为空')
    })

    it('shows error when update fails', async () => {
      const { userApi } = await import('@/api/auth')
      vi.mocked(userApi.getMe).mockResolvedValue({ data: { success: true, data: mockUser } } as any)
      vi.mocked(userApi.updateMe).mockRejectedValue({
        response: { data: { message: 'Update failed' } },
      })

      const wrapper = mount(UserProfile, mountOptions)
      await flushPromises()

      const vm = wrapper.vm as any
      vm.startEdit()
      await flushPromises()

      await vm.saveEdit()
      await flushPromises()

      expect(mockMessage.error).toHaveBeenCalledWith('Update failed')
    })
  })

  describe('email verification', () => {
    it('shows verification button for unverified email', async () => {
      const { userApi } = await import('@/api/auth')
      vi.mocked(userApi.getMe).mockResolvedValue({
        data: { success: true, data: { ...mockUser, emailVerified: false } },
      } as any)

      const wrapper = mount(UserProfile, mountOptions)
      await flushPromises()

      const buttons = wrapper.findAll('button')
      const verifyButton = buttons.find((b) => b.text().includes('立即验证'))
      expect(verifyButton?.exists()).toBe(true)
    })

    it('shows verified badge for verified email', async () => {
      const { userApi } = await import('@/api/auth')
      vi.mocked(userApi.getMe).mockResolvedValue({ data: { success: true, data: mockUser } } as any)

      const wrapper = mount(UserProfile, mountOptions)
      await flushPromises()

      expect(wrapper.text()).toContain('已验证')
    })

    it('calls send verification email when clicking verify', async () => {
      const { userApi, authApi } = await import('@/api/auth')
      vi.mocked(userApi.getMe).mockResolvedValue({
        data: { success: true, data: { ...mockUser, emailVerified: false } },
      } as any)
      vi.mocked(authApi.sendVerifyEmail).mockResolvedValue({ data: { success: true } } as any)

      const wrapper = mount(UserProfile, mountOptions)
      await flushPromises()

      const vm = wrapper.vm as any
      await vm.handleResendVerification()
      await flushPromises()

      expect(authApi.sendVerifyEmail).toHaveBeenCalled()
      expect(mockMessage.success).toHaveBeenCalledWith('验证邮件已发送')
    })
  })

  describe('error handling', () => {
    it('shows error on profile load failure', async () => {
      const { userApi } = await import('@/api/auth')
      vi.mocked(userApi.getMe).mockRejectedValue({
        response: { data: { message: 'Failed to load' } },
      })

      mount(UserProfile, mountOptions)
      await flushPromises()

      expect(mockMessage.error).toHaveBeenCalledWith('加载用户信息失败')
    })

    it('renders without crashing on error', async () => {
      const { userApi } = await import('@/api/auth')
      vi.mocked(userApi.getMe).mockRejectedValue(new Error('Error'))

      const wrapper = mount(UserProfile, mountOptions)
      await flushPromises()

      expect(wrapper.find('.page-container').exists()).toBe(true)
    })
  })

  describe('initial state', () => {
    it('initializes with empty user', async () => {
      const wrapper = mount(UserProfile, mountOptions)
      const vm = wrapper.vm as any

      expect(vm.user).toBeNull()
      expect(vm.editing).toBe(false)
      expect(vm.loading).toBe(true)
    })

    it('initializes with empty edit form', async () => {
      const wrapper = mount(UserProfile, mountOptions)
      const vm = wrapper.vm as any

      expect(vm.editForm.name).toBe('')
      expect(vm.editForm.email).toBe('')
    })
  })
})
