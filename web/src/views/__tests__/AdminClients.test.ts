import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'
import AdminClients from '../admin/Clients.vue'

const mockMessage = {
  success: vi.fn(),
  error: vi.fn(),
  warning: vi.fn(),
  info: vi.fn(),
}

const mockRouter = {
  push: vi.fn(),
  replace: vi.fn(),
}

vi.mock('@/api/admin', () => ({
  adminApi: {
    clients: {
      list: vi.fn(),
      create: vi.fn(),
      update: vi.fn(),
      delete: vi.fn(),
    },
  },
}))

vi.mock('vue-router', () => ({
  useRouter: () => mockRouter,
  useRoute: () => ({
    query: {},
    params: {},
  }),
}))

vi.mock('naive-ui', () => ({
  NCard: {
    name: 'NCard',
    props: ['title', 'style'],
    template: '<div class="n-card"><slot /><slot name="header-extra" /></div>',
  },
  NDataTable: {
    name: 'NDataTable',
    props: ['columns', 'data', 'loading', 'bordered'],
    template: '<div class="n-data-table"><slot /></div>',
  },
  NButton: {
    name: 'NButton',
    props: ['type', 'size', 'loading', 'disabled', 'attrType', 'block'],
    emits: ['click'],
    template:
      '<button :type="attrType" :disabled="disabled || loading" @click="$emit(\'click\')"><slot /></button>',
  },
  NSpace: {
    name: 'NSpace',
    props: ['justify', 'align', 'vertical', 'size'],
    template: '<div class="n-space"><slot /></div>',
  },
  NModal: {
    name: 'NModal',
    props: ['show', 'preset', 'title', 'style'],
    emits: ['update:show'],
    template: '<div class="n-modal" v-if="show"><slot /></div>',
  },
  NForm: {
    name: 'NForm',
    props: ['model', 'labelPlacement'],
    template: '<form class="n-form"><slot /></form>',
  },
  NFormItem: {
    name: 'NFormItem',
    props: ['label'],
    template: '<div class="n-form-item"><label v-if="label">{{ label }}</label><slot /></div>',
  },
  NInput: {
    name: 'NInput',
    props: ['modelValue', 'type', 'placeholder', 'size', 'maxlength', 'disabled', 'readonly'],
    emits: ['update:modelValue'],
    template:
      '<input :type="type || \'text\'" :value="modelValue" :placeholder="placeholder" :maxlength="maxlength" :disabled="disabled" :readonly="readonly" @input="$emit(\'update:modelValue\', $event.target.value)" />',
  },
  NAlert: {
    name: 'NAlert',
    props: ['type'],
    template: '<div class="n-alert"><slot /></div>',
  },
  NPopconfirm: {
    name: 'NPopconfirm',
    props: [],
    emits: ['positive-click'],
    template:
      '<div class="n-popconfirm" @click="$emit(\'positive-click\')"><slot name="trigger" /></div>',
  },
  NInputGroup: {
    name: 'NInputGroup',
    template: '<div class="n-input-group"><slot /></div>',
  },
  useMessage: () => mockMessage,
}))

describe('AdminClients.vue', () => {
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

  interface MockClient {
    id: string
    clientId: string
    name: string
    logoUri?: string
    createdAt: string
  }

  const mockClients: MockClient[] = [
    {
      id: '1',
      clientId: 'my-app-1',
      name: 'My Application 1',
      logoUri: 'https://example.com/logo1.png',
      createdAt: '2024-01-01T00:00:00Z',
    },
    {
      id: '2',
      clientId: 'my-app-2',
      name: 'My Application 2',
      logoUri: undefined,
      createdAt: '2024-01-02T00:00:00Z',
    },
  ]

  describe('rendering', () => {
    it('renders client management page correctly', async () => {
      const { adminApi } = await import('@/api/admin')
      vi.mocked(adminApi.clients.list).mockResolvedValue({
        data: { success: true, data: mockClients },
      } as any)

      const wrapper = mount(AdminClients, mountOptions)
      await flushPromises()

      expect(wrapper.find('.page-container').exists()).toBe(true)
      expect(wrapper.find('.n-card').exists()).toBe(true)
    })

    it('renders data table with clients', async () => {
      const { adminApi } = await import('@/api/admin')
      vi.mocked(adminApi.clients.list).mockResolvedValue({
        data: { success: true, data: mockClients },
      } as any)

      const wrapper = mount(AdminClients, mountOptions)
      await flushPromises()

      expect(wrapper.find('.n-data-table').exists()).toBe(true)
    })

    it('renders create client button', async () => {
      const { adminApi } = await import('@/api/admin')
      vi.mocked(adminApi.clients.list).mockResolvedValue({
        data: { success: true, data: mockClients },
      } as any)

      const wrapper = mount(AdminClients, mountOptions)
      await flushPromises()

      const buttons = wrapper.findAll('button')
      const createButton = buttons.find((b) => b.text().includes('创建客户端'))
      expect(createButton?.exists()).toBe(true)
    })
  })

  describe('component state', () => {
    it('initializes with empty clients list', async () => {
      const wrapper = mount(AdminClients, mountOptions)
      const vm = wrapper.vm as any
      expect(vm.clients).toEqual([])
    })

    it('initializes with modal closed', async () => {
      const wrapper = mount(AdminClients, mountOptions)
      const vm = wrapper.vm as any
      expect(vm.showClientModal).toBe(false)
      expect(vm.showSecretModal).toBe(false)
    })
  })

  describe('component logic', () => {
    it('fetches clients on mount', async () => {
      const { adminApi } = await import('@/api/admin')
      vi.mocked(adminApi.clients.list).mockResolvedValue({
        data: { success: true, data: mockClients },
      } as any)

      mount(AdminClients, mountOptions)
      await flushPromises()

      expect(adminApi.clients.list).toHaveBeenCalled()
    })

    it('shows error on fetch failure', async () => {
      const { adminApi } = await import('@/api/admin')
      vi.mocked(adminApi.clients.list).mockRejectedValue({
        response: { data: { message: 'Failed to fetch clients' } },
      })

      const wrapper = mount(AdminClients, mountOptions)
      await flushPromises()

      expect(mockMessage.error).toHaveBeenCalledWith('Failed to fetch clients')
    })
  })

  describe('client modal operations', () => {
    it('opens create modal when clicking create button', async () => {
      const { adminApi } = await import('@/api/admin')
      vi.mocked(adminApi.clients.list).mockResolvedValue({
        data: { success: true, data: mockClients },
      } as any)

      const wrapper = mount(AdminClients, mountOptions)
      await flushPromises()

      const createButton = wrapper.findAll('button').find((b) => b.text() === '创建客户端')
      expect(createButton).toBeDefined()

      await createButton!.trigger('click')
      await flushPromises()

      const vm = wrapper.vm as any
      expect(vm.showClientModal).toBe(true)
      expect(vm.editingClient).toBeNull()
    })

    it('opens edit modal with client data', async () => {
      const { adminApi } = await import('@/api/admin')
      vi.mocked(adminApi.clients.list).mockResolvedValue({
        data: { success: true, data: mockClients },
      } as any)

      const wrapper = mount(AdminClients, mountOptions)
      await flushPromises()

      const vm = wrapper.vm as any
      const client = {
        ...mockClients[0],
        clientSecret: 'secret',
        redirectUris: 'https://example.com',
      }
      vm.openEditModal(client)
      await flushPromises()

      expect(vm.showClientModal).toBe(true)
      expect(vm.editingClient).toEqual(client)
    })

    it('resets form when opening create modal', async () => {
      const wrapper = mount(AdminClients, mountOptions)
      await flushPromises()

      const vm = wrapper.vm as any
      vm.openCreateModal()
      await flushPromises()

      expect(vm.clientForm.clientId).toBe('')
      expect(vm.clientForm.name).toBe('')
      expect(vm.clientForm.redirectUris).toBe('')
    })
  })

  describe('client creation', () => {
    it('calls create API when saving new client', async () => {
      const { adminApi } = await import('@/api/admin')
      vi.mocked(adminApi.clients.list).mockResolvedValue({
        data: { success: true, data: mockClients },
      } as any)
      vi.mocked(adminApi.clients.create).mockResolvedValue({
        data: { success: true },
      } as any)

      const wrapper = mount(AdminClients, mountOptions)
      await flushPromises()

      const vm = wrapper.vm as any
      vm.openCreateModal()
      await flushPromises()

      vm.clientForm.clientId = 'new-client'
      vm.clientForm.name = 'New Client'
      vm.clientForm.redirectUris = 'https://example.com/callback'

      await vm.handleSaveClient()
      await flushPromises()

      expect(adminApi.clients.create).toHaveBeenCalled()
      expect(mockMessage.success).toHaveBeenCalledWith('客户端创建成功')
    })

    it('shows secret modal when new client has secret', async () => {
      const { adminApi } = await import('@/api/admin')
      vi.mocked(adminApi.clients.list).mockResolvedValue({
        data: { success: true, data: mockClients },
      } as any)
      vi.mocked(adminApi.clients.create).mockResolvedValue({
        data: { success: true, data: { clientSecret: 'generated-secret' } },
      } as any)

      const wrapper = mount(AdminClients, mountOptions)
      await flushPromises()

      const vm = wrapper.vm as any
      vm.openCreateModal()
      await flushPromises()

      vm.clientForm.clientId = 'new-client'
      vm.clientForm.name = 'New Client'
      vm.clientForm.redirectUris = 'https://example.com/callback'

      await vm.handleSaveClient()
      await flushPromises()

      expect(vm.showSecretModal).toBe(true)
      expect(vm.newClientSecret).toBe('generated-secret')
    })
  })

  describe('client update', () => {
    it('calls update API when saving edited client', async () => {
      const { adminApi } = await import('@/api/admin')
      vi.mocked(adminApi.clients.list).mockResolvedValue({
        data: { success: true, data: mockClients },
      } as any)
      vi.mocked(adminApi.clients.update).mockResolvedValue({
        data: { success: true },
      } as any)

      const wrapper = mount(AdminClients, mountOptions)
      await flushPromises()

      const vm = wrapper.vm as any
      const client = {
        ...mockClients[0],
        clientSecret: 'secret',
        redirectUris: 'https://example.com',
      }
      vm.openEditModal(client)
      await flushPromises()

      vm.clientForm.name = 'Updated Client'

      await vm.handleSaveClient()
      await flushPromises()

      expect(adminApi.clients.update).toHaveBeenCalled()
      expect(mockMessage.success).toHaveBeenCalledWith('客户端更新成功')
    })
  })

  describe('client deletion', () => {
    it('calls delete API when confirming delete', async () => {
      const { adminApi } = await import('@/api/admin')
      vi.mocked(adminApi.clients.list).mockResolvedValue({
        data: { success: true, data: mockClients },
      } as any)
      vi.mocked(adminApi.clients.delete).mockResolvedValue({
        data: { success: true },
      } as any)

      const wrapper = mount(AdminClients, mountOptions)
      await flushPromises()

      const vm = wrapper.vm as any
      await vm.handleDelete('1')
      await flushPromises()

      expect(adminApi.clients.delete).toHaveBeenCalledWith('1')
      expect(mockMessage.success).toHaveBeenCalledWith('客户端删除成功')
    })

    it('shows error when delete fails', async () => {
      const { adminApi } = await import('@/api/admin')
      vi.mocked(adminApi.clients.list).mockResolvedValue({
        data: { success: true, data: mockClients },
      } as any)
      vi.mocked(adminApi.clients.delete).mockRejectedValue({
        response: { data: { message: 'Delete failed' } },
      })

      const wrapper = mount(AdminClients, mountOptions)
      await flushPromises()

      const vm = wrapper.vm as any
      await vm.handleDelete('1')
      await flushPromises()

      expect(mockMessage.error).toHaveBeenCalledWith('Delete failed')
    })
  })

  describe('clipboard functionality', () => {
    it('copies secret to clipboard', async () => {
      // Mock the clipboard API before component access
      const mockWriteText = vi.fn().mockResolvedValue(undefined)
      vi.stubGlobal('navigator', {
        ...navigator,
        clipboard: {
          writeText: mockWriteText,
        },
      })

      const wrapper = mount(AdminClients, mountOptions)
      await flushPromises()

      const vm = wrapper.vm as any
      vm.newClientSecret = 'test-secret'
      vm.copyToClipboard('test-secret')
      await flushPromises()

      expect(mockWriteText).toHaveBeenCalledWith('test-secret')
      expect(mockMessage.success).toHaveBeenCalledWith('已复制到剪贴板')
    })
  })
})
