import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import UserPasskeys from '../user/Passkeys.vue'
import type { Passkey } from '@/types'

const mockMessage = {
  success: vi.fn(),
  error: vi.fn(),
  warning: vi.fn(),
  info: vi.fn(),
}

vi.mock('@/api/auth', () => ({
  passkeyApi: {
    getPasskeys: vi.fn(),
    registrationStart: vi.fn(),
    registrationEnd: vi.fn(),
    deletePasskey: vi.fn(),
  },
}))

vi.mock('naive-ui', () => ({
  NCard: {
    name: 'NCard',
    props: ['title', 'bordered'],
    template: '<div class="n-card"><slot /><slot name="header-extra" /></div>',
  },
  NDataTable: {
    name: 'NDataTable',
    props: ['columns', 'data', 'loading', 'bordered'],
    template: '<div class="n-data-table"><slot /></div>',
  },
  NButton: {
    name: 'NButton',
    props: ['type', 'size', 'loading', 'disabled'],
    emits: ['click'],
    template: '<button :disabled="disabled || loading" @click="$emit(\'click\')"><slot /></button>',
  },
  NAlert: {
    name: 'NAlert',
    props: ['type'],
    template: '<div class="n-alert"><slot /></div>',
  },
  NModal: {
    name: 'NModal',
    props: ['show', 'preset', 'title', 'style'],
    emits: ['update:show'],
    template: '<div class="n-modal" v-if="show"><slot /></div>',
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
    props: ['modelValue', 'placeholder'],
    emits: ['update:modelValue'],
    template:
      '<input :value="modelValue" :placeholder="placeholder" @input="$emit(\'update:modelValue\', $event.target.value)" />',
  },
  NSpace: {
    name: 'NSpace',
    props: ['justify', 'size'],
    template: '<div class="n-space"><slot /></div>',
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

describe('UserPasskeys.vue', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  const mockPasskeys: Passkey[] = [
    {
      id: '1',
      userId: 'user1',
      name: 'My Laptop',
      credentialId: 'cred-1-abc123',
      publicKey: 'pub-key-1',
      counter: 1,
      deviceType: 'usb',
      lastUsedAt: '2024-01-15T00:00:00Z',
      createdAt: '2024-01-01T00:00:00Z',
    },
    {
      id: '2',
      userId: 'user1',
      name: 'Office Key',
      credentialId: 'cred-2-def456',
      publicKey: 'pub-key-2',
      counter: 2,
      deviceType: 'usb',
      lastUsedAt: null,
      createdAt: '2024-01-05T00:00:00Z',
    },
  ]

  const mountOptions = {
    global: {
      stubs: {},
    },
  }

  describe('rendering', () => {
    it('renders passkeys page', async () => {
      const { passkeyApi } = await import('@/api/auth')
      vi.mocked(passkeyApi.getPasskeys).mockResolvedValue({
        data: { success: true, data: mockPasskeys },
      } as any)

      const wrapper = mount(UserPasskeys, mountOptions)
      await flushPromises()

      expect(wrapper.find('.page-container').exists()).toBe(true)
      expect(wrapper.find('.n-card').exists()).toBe(true)
    })

    it('renders data table', async () => {
      const { passkeyApi } = await import('@/api/auth')
      vi.mocked(passkeyApi.getPasskeys).mockResolvedValue({
        data: { success: true, data: mockPasskeys },
      } as any)

      const wrapper = mount(UserPasskeys, mountOptions)
      await flushPromises()

      expect(wrapper.find('.n-data-table').exists()).toBe(true)
    })

    it('renders add button', async () => {
      const { passkeyApi } = await import('@/api/auth')
      vi.mocked(passkeyApi.getPasskeys).mockResolvedValue({
        data: { success: true, data: mockPasskeys },
      } as any)

      const wrapper = mount(UserPasskeys, mountOptions)
      await flushPromises()

      const buttons = wrapper.findAll('button')
      const addButton = buttons.find((b) => b.text().includes('添加'))
      expect(addButton?.exists()).toBe(true)
    })

    it('renders info alert', async () => {
      const { passkeyApi } = await import('@/api/auth')
      vi.mocked(passkeyApi.getPasskeys).mockResolvedValue({
        data: { success: true, data: mockPasskeys },
      } as any)

      const wrapper = mount(UserPasskeys, mountOptions)
      await flushPromises()

      expect(wrapper.find('.n-alert').exists()).toBe(true)
      expect(wrapper.text()).toContain('通行密钥允许您使用 WebAuthn 安全地无密码登录')
    })
  })

  describe('data fetching', () => {
    it('fetches passkeys on mount', async () => {
      const { passkeyApi } = await import('@/api/auth')
      vi.mocked(passkeyApi.getPasskeys).mockResolvedValue({
        data: { success: true, data: mockPasskeys },
      } as any)

      mount(UserPasskeys, mountOptions)
      await flushPromises()

      expect(passkeyApi.getPasskeys).toHaveBeenCalled()
    })

    it('stores passkeys in state after fetch', async () => {
      const { passkeyApi } = await import('@/api/auth')
      vi.mocked(passkeyApi.getPasskeys).mockResolvedValue({
        data: { success: true, data: mockPasskeys },
      } as any)

      const wrapper = mount(UserPasskeys, mountOptions)
      await flushPromises()

      const vm = wrapper.vm as any
      expect(vm.passkeys).toHaveLength(2)
      expect(vm.passkeys[0].name).toBe('My Laptop')
    })
  })

  describe('modal operations', () => {
    it('opens setup modal when clicking add button', async () => {
      const { passkeyApi } = await import('@/api/auth')
      vi.mocked(passkeyApi.getPasskeys).mockResolvedValue({
        data: { success: true, data: mockPasskeys },
      } as any)

      const wrapper = mount(UserPasskeys, mountOptions)
      await flushPromises()

      const addButton = wrapper.findAll('button').find((b) => b.text().includes('添加'))
      await addButton!.trigger('click')
      await flushPromises()

      const vm = wrapper.vm as any
      expect(vm.showPasskeySetup).toBe(true)
    })

    it('closes modal on cancel', async () => {
      const { passkeyApi } = await import('@/api/auth')
      vi.mocked(passkeyApi.getPasskeys).mockResolvedValue({
        data: { success: true, data: mockPasskeys },
      } as any)

      const wrapper = mount(UserPasskeys, mountOptions)
      await flushPromises()

      const vm = wrapper.vm as any
      vm.showPasskeySetup = false
      await flushPromises()

      expect(vm.showPasskeySetup).toBe(false)
    })
  })

  describe('passkey deletion', () => {
    it('calls delete API when confirming delete', async () => {
      const { passkeyApi } = await import('@/api/auth')
      vi.mocked(passkeyApi.getPasskeys).mockResolvedValue({
        data: { success: true, data: mockPasskeys },
      } as any)
      vi.mocked(passkeyApi.deletePasskey).mockResolvedValue({ data: { success: true } } as any)

      const wrapper = mount(UserPasskeys, mountOptions)
      await flushPromises()

      const vm = wrapper.vm as any
      await vm.handleDeletePasskey('1')
      await flushPromises()

      expect(passkeyApi.deletePasskey).toHaveBeenCalledWith('1')
      expect(mockMessage.success).toHaveBeenCalledWith('通行密钥已删除')
    })

    it('shows error when delete fails', async () => {
      const { passkeyApi } = await import('@/api/auth')
      vi.mocked(passkeyApi.getPasskeys).mockResolvedValue({
        data: { success: true, data: mockPasskeys },
      } as any)
      vi.mocked(passkeyApi.deletePasskey).mockRejectedValue({
        response: { data: { message: 'Delete failed' } },
      })

      const wrapper = mount(UserPasskeys, mountOptions)
      await flushPromises()

      const vm = wrapper.vm as any
      await vm.handleDeletePasskey('1')
      await flushPromises()

      expect(mockMessage.error).toHaveBeenCalledWith('Delete failed')
    })
  })

  describe('error handling', () => {
    it('shows error on fetch failure', async () => {
      const { passkeyApi } = await import('@/api/auth')
      vi.mocked(passkeyApi.getPasskeys).mockRejectedValue({
        response: { data: { message: 'Failed to load' } },
      })

      const wrapper = mount(UserPasskeys, mountOptions)
      await flushPromises()

      expect(mockMessage.error).toHaveBeenCalledWith('Failed to load')
    })

    it('renders without crashing on error', async () => {
      const { passkeyApi } = await import('@/api/auth')
      vi.mocked(passkeyApi.getPasskeys).mockRejectedValue(new Error('Error'))

      const wrapper = mount(UserPasskeys, mountOptions)
      await flushPromises()

      expect(wrapper.find('.page-container').exists()).toBe(true)
    })
  })

  describe('passkey registration', () => {
    it('starts registration flow when clicking create', async () => {
      const { passkeyApi } = await import('@/api/auth')
      vi.mocked(passkeyApi.getPasskeys).mockResolvedValue({
        data: { success: true, data: mockPasskeys },
      } as any)
      vi.mocked(passkeyApi.registrationStart).mockResolvedValue({
        data: {
          success: true,
          data: {
            challenge: 'dGVzdC1jaGFsbGVuZ2U=',
            options: JSON.stringify({
              rp: { name: 'Test App', id: 'localhost' },
              user: { id: 'dXNlci1pZA==', name: 'user@test.com', displayName: 'User' },
              pubKeyCredParams: [{ type: 'public-key', alg: -7 }],
              timeout: 60000,
              attestation: 'none',
              authenticatorSelection: {},
            }),
          },
        },
      })

      const wrapper = mount(UserPasskeys, mountOptions)
      await flushPromises()

      const vm = wrapper.vm as any
      vm.showPasskeySetup = true
      vm.passkeyName = 'New Key'

      const mockCredential = {
        response: {
          attestationObject: new ArrayBuffer(10),
          clientDataJSON: new ArrayBuffer(10),
        },
        rawId: new Uint8Array([116, 101, 115, 116]),
      }
      vi.stubGlobal('navigator', {
        credentials: {
          create: vi.fn().mockResolvedValue(mockCredential),
        },
      })

      await vm.handleRegisterPasskey()
      await flushPromises()
    })

    it('handles registration NotAllowedError', async () => {
      const { passkeyApi } = await import('@/api/auth')
      vi.mocked(passkeyApi.getPasskeys).mockResolvedValue({
        data: { success: true, data: mockPasskeys },
      } as any)
      vi.mocked(passkeyApi.registrationStart).mockResolvedValue({
        data: {
          success: true,
          data: {
            challenge: 'dGVzdC1jaGFsbGVuZ2U=',
            options: JSON.stringify({
              rp: { name: 'Test App', id: 'localhost' },
              user: { id: 'dXNlci1pZA==', name: 'user@test.com', displayName: 'User' },
              pubKeyCredParams: [{ type: 'public-key', alg: -7 }],
              timeout: 60000,
              attestation: 'none',
              authenticatorSelection: {},
            }),
          },
        },
      })

      const wrapper = mount(UserPasskeys, mountOptions)
      await flushPromises()

      const vm = wrapper.vm as any
      vm.showPasskeySetup = true
      vm.passkeyName = 'New Key'

      const notAllowedError = new Error('NotAllowedError')
      notAllowedError.name = 'NotAllowedError'

      vi.stubGlobal('navigator', {
        credentials: {
          create: vi.fn().mockRejectedValue(notAllowedError),
        },
      })

      await vm.handleRegisterPasskey()
      await flushPromises()

      expect(mockMessage.error).toHaveBeenCalledWith('通行密钥注册已取消')
    })

    it('handles registration failure with generic error', async () => {
      const { passkeyApi } = await import('@/api/auth')
      vi.mocked(passkeyApi.getPasskeys).mockResolvedValue({
        data: { success: true, data: mockPasskeys },
      } as any)
      vi.mocked(passkeyApi.registrationStart).mockResolvedValue({
        data: {
          success: true,
          data: {
            challenge: 'dGVzdC1jaGFsbGVuZ2U=',
            options: JSON.stringify({
              rp: { name: 'Test App', id: 'localhost' },
              user: { id: 'dXNlci1pZA==', name: 'user@test.com', displayName: 'User' },
              pubKeyCredParams: [{ type: 'public-key', alg: -7 }],
              timeout: 60000,
              attestation: 'none',
              authenticatorSelection: {},
            }),
          },
        },
      })

      const wrapper = mount(UserPasskeys, mountOptions)
      await flushPromises()

      const vm = wrapper.vm as any
      vm.showPasskeySetup = true
      vm.passkeyName = 'New Key'

      vi.stubGlobal('navigator', {
        credentials: {
          create: vi.fn().mockRejectedValue(new Error('Registration failed')),
        },
      })

      await vm.handleRegisterPasskey()
      await flushPromises()

      expect(mockMessage.error).toHaveBeenCalledWith('Registration failed')
    })

    it('handles registration start failure', async () => {
      const { passkeyApi } = await import('@/api/auth')
      vi.mocked(passkeyApi.getPasskeys).mockResolvedValue({
        data: { success: true, data: mockPasskeys },
      } as any)
      vi.mocked(passkeyApi.registrationStart).mockResolvedValue({
        data: {
          success: false,
        },
      } as any)

      const wrapper = mount(UserPasskeys, mountOptions)
      await flushPromises()

      const vm = wrapper.vm as any
      vm.showPasskeySetup = true
      vm.passkeyName = 'New Key'

      await vm.handleRegisterPasskey()
      await flushPromises()

      expect(mockMessage.error).toHaveBeenCalledWith('无法开始通行密钥注册')
    })

    it('handles registration start with null data', async () => {
      const { passkeyApi } = await import('@/api/auth')
      vi.mocked(passkeyApi.getPasskeys).mockResolvedValue({
        data: { success: true, data: mockPasskeys },
      } as any)
      vi.mocked(passkeyApi.registrationStart).mockResolvedValue({
        data: {
          success: true,
          data: null,
        },
      } as any)

      const wrapper = mount(UserPasskeys, mountOptions)
      await flushPromises()

      const vm = wrapper.vm as any
      vm.showPasskeySetup = true
      vm.passkeyName = 'New Key'

      await vm.handleRegisterPasskey()
      await flushPromises()

      expect(mockMessage.error).toHaveBeenCalledWith('无法开始通行密钥注册')
    })
  })

  describe('empty state', () => {
    it('shows empty table when no passkeys', async () => {
      const { passkeyApi } = await import('@/api/auth')
      vi.mocked(passkeyApi.getPasskeys).mockResolvedValue({
        data: { success: true, data: [] },
      } as any)

      const wrapper = mount(UserPasskeys, mountOptions)
      await flushPromises()

      expect(wrapper.find('.n-data-table').exists()).toBe(true)
    })
  })

  describe('component state', () => {
    it('initializes with correct default values', async () => {
      const { passkeyApi } = await import('@/api/auth')
      vi.mocked(passkeyApi.getPasskeys).mockResolvedValue({
        data: { success: true, data: mockPasskeys },
      } as any)

      const wrapper = mount(UserPasskeys, mountOptions)
      const vm = wrapper.vm as any

      expect(vm.showPasskeySetup).toBe(false)
      expect(vm.passkeyName).toBe('')
      expect(vm.registeringPasskey).toBe(false)
    })
  })
})
