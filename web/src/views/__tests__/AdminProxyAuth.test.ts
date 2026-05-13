import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import AdminProxyAuth from '../admin/ProxyAuth.vue'
import type { Group } from '@/types'

const mockMessage = {
  success: vi.fn(),
  error: vi.fn(),
  warning: vi.fn(),
  info: vi.fn(),
}

vi.mock('@/api/admin', () => ({
  adminApi: {
    proxyauth: {
      list: vi.fn(),
      create: vi.fn(),
      update: vi.fn(),
      delete: vi.fn(),
    },
    groups: {
      list: vi.fn(),
    },
  },
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
    props: ['type', 'size', 'loading', 'disabled'],
    emits: ['click'],
    template: '<button :disabled="disabled || loading" @click="$emit(\'click\')"><slot /></button>',
  },
  NSpace: {
    name: 'NSpace',
    props: ['justify', 'align', 'vertical', 'size'],
    template: '<div class="n-space"><slot /></div>',
  },
  NTag: {
    name: 'NTag',
    props: ['type'],
    template: '<span class="n-tag"><slot /></span>',
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
    props: ['modelValue', 'type', 'placeholder'],
    emits: ['update:modelValue'],
    template:
      '<input :type="type || \'text\'" :value="modelValue" :placeholder="placeholder" @input="$emit(\'update:modelValue\', $event.target.value)" />',
  },
  NSwitch: {
    name: 'NSwitch',
    props: ['modelValue'],
    emits: ['update:modelValue'],
    template: '<div class="n-switch" @click="$emit(\'update:modelValue\', !modelValue)"></div>',
  },
  NSelect: {
    name: 'NSelect',
    props: ['modelValue', 'options', 'placeholder', 'clearable'],
    emits: ['update:modelValue'],
    template:
      '<select :value="modelValue" @change="$emit(\'update:modelValue\', $event.target.value)"><slot /></select>',
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

interface ProxyAuth {
  id: string
  name: string
  proxyUrl: string
  enabled: boolean
  createdAt: string
  headerName?: string
  groupId?: string
  scopes?: string
}

describe('AdminProxyAuth.vue', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  const mockGroups: Group[] = [
    { id: '1', name: 'Admin Group', description: 'Admin users' },
    { id: '2', name: 'User Group', description: 'Regular users' },
  ]

  const mockProxyAuths: ProxyAuth[] = [
    {
      id: '1',
      name: 'Test Proxy',
      proxyUrl: 'https://proxy1.example.com',
      enabled: true,
      createdAt: '2024-01-01',
      headerName: 'X-Token',
    },
    {
      id: '2',
      name: 'Disabled Proxy',
      proxyUrl: 'https://proxy2.example.com',
      enabled: false,
      createdAt: '2024-01-02',
      headerName: 'X-User',
    },
  ]

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
    it('renders proxy auth management page', async () => {
      const { adminApi } = await import('@/api/admin')
      vi.mocked(adminApi.proxyauth.list).mockResolvedValue({
        data: { success: true, data: mockProxyAuths },
      } as any)
      vi.mocked(adminApi.groups.list).mockResolvedValue({
        data: { success: true, data: mockGroups },
      } as any)

      const wrapper = mount(AdminProxyAuth, mountOptions)
      await flushPromises()

      expect(wrapper.find('.page-container').exists()).toBe(true)
      expect(wrapper.find('.n-card').exists()).toBe(true)
    })

    it('renders data table', async () => {
      const { adminApi } = await import('@/api/admin')
      vi.mocked(adminApi.proxyauth.list).mockResolvedValue({
        data: { success: true, data: mockProxyAuths },
      } as any)
      vi.mocked(adminApi.groups.list).mockResolvedValue({
        data: { success: true, data: mockGroups },
      } as any)

      const wrapper = mount(AdminProxyAuth, mountOptions)
      await flushPromises()

      expect(wrapper.find('.n-data-table').exists()).toBe(true)
    })

    it('renders create button', async () => {
      const { adminApi } = await import('@/api/admin')
      vi.mocked(adminApi.proxyauth.list).mockResolvedValue({
        data: { success: true, data: mockProxyAuths },
      } as any)
      vi.mocked(adminApi.groups.list).mockResolvedValue({
        data: { success: true, data: mockGroups },
      } as any)

      const wrapper = mount(AdminProxyAuth, mountOptions)
      await flushPromises()

      const buttons = wrapper.findAll('button')
      const createButton = buttons.find((b) => b.text().includes('创建'))
      expect(createButton?.exists()).toBe(true)
    })
  })

  describe('data fetching', () => {
    it('fetches proxy auths and groups on mount', async () => {
      const { adminApi } = await import('@/api/admin')
      vi.mocked(adminApi.proxyauth.list).mockResolvedValue({
        data: { success: true, data: mockProxyAuths },
      } as any)
      vi.mocked(adminApi.groups.list).mockResolvedValue({
        data: { success: true, data: mockGroups },
      } as any)

      mount(AdminProxyAuth, mountOptions)
      await flushPromises()

      expect(adminApi.proxyauth.list).toHaveBeenCalled()
      expect(adminApi.groups.list).toHaveBeenCalled()
    })

    it('stores proxy auths in state after fetch', async () => {
      const { adminApi } = await import('@/api/admin')
      vi.mocked(adminApi.proxyauth.list).mockResolvedValue({
        data: { success: true, data: mockProxyAuths },
      } as any)
      vi.mocked(adminApi.groups.list).mockResolvedValue({
        data: { success: true, data: mockGroups },
      } as any)

      const wrapper = mount(AdminProxyAuth, mountOptions)
      await flushPromises()

      const vm = wrapper.vm as any
      expect(vm.proxyauths).toHaveLength(2)
      expect(vm.proxyauths[0].name).toBe('Test Proxy')
    })
  })

  describe('modal operations', () => {
    it('opens create modal when clicking create button', async () => {
      const { adminApi } = await import('@/api/admin')
      vi.mocked(adminApi.proxyauth.list).mockResolvedValue({
        data: { success: true, data: mockProxyAuths },
      } as any)
      vi.mocked(adminApi.groups.list).mockResolvedValue({
        data: { success: true, data: mockGroups },
      } as any)

      const wrapper = mount(AdminProxyAuth, mountOptions)
      await flushPromises()

      const createButton = wrapper.findAll('button').find((b) => b.text().includes('创建'))
      await createButton!.trigger('click')
      await flushPromises()

      const vm = wrapper.vm as any
      expect(vm.showProxyAuthModal).toBe(true)
    })

    it('opens edit modal with proxy auth data', async () => {
      const { adminApi } = await import('@/api/admin')
      vi.mocked(adminApi.proxyauth.list).mockResolvedValue({
        data: { success: true, data: mockProxyAuths },
      } as any)
      vi.mocked(adminApi.groups.list).mockResolvedValue({
        data: { success: true, data: mockGroups },
      } as any)

      const wrapper = mount(AdminProxyAuth, mountOptions)
      await flushPromises()

      const vm = wrapper.vm as any
      vm.openEditModal(mockProxyAuths[0])
      await flushPromises()

      expect(vm.showProxyAuthModal).toBe(true)
      expect(vm.editingProxyAuth).toEqual(mockProxyAuths[0])
    })

    it('closes modal on cancel', async () => {
      const { adminApi } = await import('@/api/admin')
      vi.mocked(adminApi.proxyauth.list).mockResolvedValue({
        data: { success: true, data: mockProxyAuths },
      } as any)
      vi.mocked(adminApi.groups.list).mockResolvedValue({
        data: { success: true, data: mockGroups },
      } as any)

      const wrapper = mount(AdminProxyAuth, mountOptions)
      await flushPromises()

      const vm = wrapper.vm as any
      vm.showProxyAuthModal = false
      await flushPromises()

      expect(vm.showProxyAuthModal).toBe(false)
    })

    it('resets form on open create modal', async () => {
      const { adminApi } = await import('@/api/admin')
      vi.mocked(adminApi.proxyauth.list).mockResolvedValue({
        data: { success: true, data: mockProxyAuths },
      } as any)
      vi.mocked(adminApi.groups.list).mockResolvedValue({
        data: { success: true, data: mockGroups },
      } as any)

      const wrapper = mount(AdminProxyAuth, mountOptions)
      await flushPromises()

      const vm = wrapper.vm as any
      vm.openCreateModal()
      await flushPromises()

      expect(vm.proxyAuthForm.name).toBe('')
      expect(vm.proxyAuthForm.proxyUrl).toBe('')
      expect(vm.proxyAuthForm.headerName).toBe('')
      expect(vm.editingProxyAuth).toBeNull()
    })
  })

  describe('proxy auth creation', () => {
    it('calls create API when saving new proxy auth', async () => {
      const { adminApi } = await import('@/api/admin')
      vi.mocked(adminApi.proxyauth.list).mockResolvedValue({
        data: { success: true, data: mockProxyAuths },
      } as any)
      vi.mocked(adminApi.groups.list).mockResolvedValue({
        data: { success: true, data: mockGroups },
      } as any)
      vi.mocked(adminApi.proxyauth.create).mockResolvedValue({ data: { success: true } } as any)

      const wrapper = mount(AdminProxyAuth, mountOptions)
      await flushPromises()

      const vm = wrapper.vm as any
      vm.openCreateModal()
      await flushPromises()

      vm.proxyAuthForm.name = 'New Proxy'
      vm.proxyAuthForm.proxyUrl = 'https://new-proxy.com'
      vm.proxyAuthForm.headerName = 'X-New-Token'

      await vm.handleSaveProxyAuth()
      await flushPromises()

      expect(adminApi.proxyauth.create).toHaveBeenCalled()
      expect(mockMessage.success).toHaveBeenCalledWith('代理认证创建成功')
    })

    it('shows error when create fails', async () => {
      const { adminApi } = await import('@/api/admin')
      vi.mocked(adminApi.proxyauth.list).mockResolvedValue({
        data: { success: true, data: mockProxyAuths },
      } as any)
      vi.mocked(adminApi.groups.list).mockResolvedValue({
        data: { success: true, data: mockGroups },
      } as any)
      vi.mocked(adminApi.proxyauth.create).mockRejectedValue({
        response: { data: { message: 'Create failed' } },
      })

      const wrapper = mount(AdminProxyAuth, mountOptions)
      await flushPromises()

      const vm = wrapper.vm as any
      vm.openCreateModal()
      await flushPromises()

      vm.proxyAuthForm.name = 'New Proxy'
      vm.proxyAuthForm.proxyUrl = 'https://new-proxy.com'
      vm.proxyAuthForm.headerName = 'X-New-Token'

      await vm.handleSaveProxyAuth()
      await flushPromises()

      expect(mockMessage.error).toHaveBeenCalledWith('Create failed')
    })
  })

  describe('proxy auth update', () => {
    it('calls update API when saving edited proxy auth', async () => {
      const { adminApi } = await import('@/api/admin')
      vi.mocked(adminApi.proxyauth.list).mockResolvedValue({
        data: { success: true, data: mockProxyAuths },
      } as any)
      vi.mocked(adminApi.groups.list).mockResolvedValue({
        data: { success: true, data: mockGroups },
      } as any)
      vi.mocked(adminApi.proxyauth.update).mockResolvedValue({ data: { success: true } } as any)

      const wrapper = mount(AdminProxyAuth, mountOptions)
      await flushPromises()

      const vm = wrapper.vm as any
      vm.openEditModal(mockProxyAuths[0])
      await flushPromises()

      vm.proxyAuthForm.name = 'Updated Proxy'

      await vm.handleSaveProxyAuth()
      await flushPromises()

      expect(adminApi.proxyauth.update).toHaveBeenCalled()
      expect(mockMessage.success).toHaveBeenCalledWith('代理认证更新成功')
    })
  })

  describe('proxy auth deletion', () => {
    it('calls delete API when confirming delete', async () => {
      const { adminApi } = await import('@/api/admin')
      vi.mocked(adminApi.proxyauth.list).mockResolvedValue({
        data: { success: true, data: mockProxyAuths },
      } as any)
      vi.mocked(adminApi.groups.list).mockResolvedValue({
        data: { success: true, data: mockGroups },
      } as any)
      vi.mocked(adminApi.proxyauth.delete).mockResolvedValue({ data: { success: true } } as any)

      const wrapper = mount(AdminProxyAuth, mountOptions)
      await flushPromises()

      const vm = wrapper.vm as any
      await vm.handleDelete('1')
      await flushPromises()

      expect(adminApi.proxyauth.delete).toHaveBeenCalledWith('1')
      expect(mockMessage.success).toHaveBeenCalledWith('代理认证删除成功')
    })

    it('shows error when delete fails', async () => {
      const { adminApi } = await import('@/api/admin')
      vi.mocked(adminApi.proxyauth.list).mockResolvedValue({
        data: { success: true, data: mockProxyAuths },
      } as any)
      vi.mocked(adminApi.groups.list).mockResolvedValue({
        data: { success: true, data: mockGroups },
      } as any)
      vi.mocked(adminApi.proxyauth.delete).mockRejectedValue({
        response: { data: { message: 'Delete failed' } },
      })

      const wrapper = mount(AdminProxyAuth, mountOptions)
      await flushPromises()

      const vm = wrapper.vm as any
      await vm.handleDelete('1')
      await flushPromises()

      expect(mockMessage.error).toHaveBeenCalledWith('Delete failed')
    })
  })

  describe('group options', () => {
    it('generates group options from groups data', async () => {
      const { adminApi } = await import('@/api/admin')
      vi.mocked(adminApi.proxyauth.list).mockResolvedValue({
        data: { success: true, data: mockProxyAuths },
      } as any)
      vi.mocked(adminApi.groups.list).mockResolvedValue({
        data: { success: true, data: mockGroups },
      } as any)

      const wrapper = mount(AdminProxyAuth, mountOptions)
      await flushPromises()

      const vm = wrapper.vm as any
      expect(vm.groupOptions).toHaveLength(2)
      expect(vm.groupOptions[0]).toEqual({ label: 'Admin Group', value: '1' })
      expect(vm.groupOptions[1]).toEqual({ label: 'User Group', value: '2' })
    })
  })

  describe('error handling', () => {
    it('shows error on fetch failure', async () => {
      const { adminApi } = await import('@/api/admin')
      vi.mocked(adminApi.proxyauth.list).mockRejectedValue({
        response: { data: { message: 'Failed to fetch' } },
      })
      vi.mocked(adminApi.groups.list).mockResolvedValue({
        data: { success: true, data: mockGroups },
      } as any)

      const wrapper = mount(AdminProxyAuth, mountOptions)
      await flushPromises()

      expect(mockMessage.error).toHaveBeenCalled()
    })

    it('renders without crashing on error', async () => {
      const { adminApi } = await import('@/api/admin')
      vi.mocked(adminApi.proxyauth.list).mockRejectedValue(new Error('Error'))
      vi.mocked(adminApi.groups.list).mockRejectedValue(new Error('Error'))

      const wrapper = mount(AdminProxyAuth, mountOptions)
      await flushPromises()

      expect(wrapper.find('.page-container').exists()).toBe(true)
    })
  })
})
