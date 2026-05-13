import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'
import AdminInvitations from '../admin/Invitations.vue'
import type { Group } from '@/types'

const mockMessage = {
  success: vi.fn(),
  error: vi.fn(),
  warning: vi.fn(),
  info: vi.fn(),
}

vi.mock('@/api/admin', () => ({
  adminApi: {
    invitations: {
      list: vi.fn(),
      create: vi.fn(),
      delete: vi.fn(),
    },
    groups: {
      list: vi.fn(),
    },
  },
}))

vi.mock('vue-router', () => ({
  useRouter: () => ({
    push: vi.fn(),
    replace: vi.fn(),
  }),
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
    props: ['modelValue', 'type', 'placeholder', 'size', 'maxlength', 'disabled'],
    emits: ['update:modelValue'],
    template:
      '<input :type="type || \'text\'" :value="modelValue" :placeholder="placeholder" :maxlength="maxlength" :disabled="disabled" @input="$emit(\'update:modelValue\', $event.target.value)" />',
  },
  NSelect: {
    name: 'NSelect',
    props: ['modelValue', 'options', 'placeholder', 'clearable'],
    emits: ['update:modelValue'],
    template: '<div class="n-select">{{ modelValue }}</div>',
  },
  NDatePicker: {
    name: 'NDatePicker',
    props: ['modelValue', 'type'],
    emits: ['update:modelValue'],
    template: '<div class="n-date-picker">{{ modelValue }}</div>',
  },
  NEmpty: {
    name: 'NEmpty',
    props: ['description'],
    template:
      '<div class="n-empty"><span v-if="description">{{ description }}</span><slot name="extra" /></div>',
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

interface InvitationListItem {
  id: string
  email: string
  username?: string
  expiresAt: string
  createdAt: string
}

describe('AdminInvitations.vue', () => {
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

  const mockInvitations: InvitationListItem[] = [
    {
      id: '1',
      email: 'user1@example.com',
      username: 'user1',
      expiresAt: '2030-01-01T00:00:00Z',
      createdAt: '2024-01-01T00:00:00Z',
    },
    {
      id: '2',
      email: 'user2@example.com',
      expiresAt: '2020-01-01T00:00:00Z',
      createdAt: '2024-01-02T00:00:00Z',
    },
  ]

  const mockGroups: Group[] = [
    { id: '1', name: 'Group 1', createdAt: '2024-01-01T00:00:00Z' },
    { id: '2', name: 'Group 2', createdAt: '2024-01-02T00:00:00Z' },
  ]

  describe('rendering', () => {
    it('renders invitation management page correctly', async () => {
      const { adminApi } = await import('@/api/admin')
      vi.mocked(adminApi.invitations.list).mockResolvedValue({
        data: { success: true, data: mockInvitations },
      } as any)
      vi.mocked(adminApi.groups.list).mockResolvedValue({
        data: { success: true, data: mockGroups },
      } as any)

      const wrapper = mount(AdminInvitations, mountOptions)
      await flushPromises()

      expect(wrapper.find('.page-container').exists()).toBe(true)
      expect(wrapper.find('.n-card').exists()).toBe(true)
    })

    it('renders data table when invitations exist', async () => {
      const { adminApi } = await import('@/api/admin')
      vi.mocked(adminApi.invitations.list).mockResolvedValue({
        data: { success: true, data: mockInvitations },
      } as any)
      vi.mocked(adminApi.groups.list).mockResolvedValue({
        data: { success: true, data: mockGroups },
      } as any)

      const wrapper = mount(AdminInvitations, mountOptions)
      await flushPromises()

      expect(wrapper.find('.n-data-table').exists()).toBe(true)
    })

    it('renders create invitation button', async () => {
      const { adminApi } = await import('@/api/admin')
      vi.mocked(adminApi.invitations.list).mockResolvedValue({
        data: { success: true, data: mockInvitations },
      } as any)
      vi.mocked(adminApi.groups.list).mockResolvedValue({
        data: { success: true, data: mockGroups },
      } as any)

      const wrapper = mount(AdminInvitations, mountOptions)
      await flushPromises()

      const buttons = wrapper.findAll('button')
      const createButton = buttons.find((b) => b.text().includes('创建邀请'))
      expect(createButton?.exists()).toBe(true)
    })
  })

  describe('component state', () => {
    it('initializes with empty invitations array', async () => {
      const wrapper = mount(AdminInvitations, mountOptions)
      const vm = wrapper.vm as any
      expect(vm.invitations).toEqual([])
    })

    it('initializes with modal closed', async () => {
      const wrapper = mount(AdminInvitations, mountOptions)
      const vm = wrapper.vm as any
      expect(vm.showInviteModal).toBe(false)
    })

    it('initializes with default invite form', async () => {
      const wrapper = mount(AdminInvitations, mountOptions)
      const vm = wrapper.vm as any
      expect(vm.inviteForm.email).toBe('')
      expect(vm.inviteForm.username).toBe('')
      expect(vm.inviteForm.scopes).toBe('')
      expect(vm.inviteForm.groupId).toBeUndefined()
      expect(vm.inviteForm.maxUses).toBe('1')
      expect(vm.inviteForm.expiresAt).toBeNull()
    })
  })

  describe('fetchInvitations', () => {
    it('fetches invitations on mount', async () => {
      const { adminApi } = await import('@/api/admin')
      vi.mocked(adminApi.invitations.list).mockResolvedValue({
        data: { success: true, data: mockInvitations },
      } as any)
      vi.mocked(adminApi.groups.list).mockResolvedValue({
        data: { success: true, data: mockGroups },
      } as any)

      mount(AdminInvitations, mountOptions)
      await flushPromises()

      expect(adminApi.invitations.list).toHaveBeenCalled()
    })

    it('shows error on fetch failure', async () => {
      const { adminApi } = await import('@/api/admin')
      vi.mocked(adminApi.invitations.list).mockRejectedValue({
        response: { data: { message: 'Failed to fetch invitations' } },
      })
      vi.mocked(adminApi.groups.list).mockResolvedValue({
        data: { success: true, data: mockGroups },
      } as any)

      const wrapper = mount(AdminInvitations, mountOptions)
      await flushPromises()

      expect(mockMessage.error).toHaveBeenCalledWith('Failed to fetch invitations')
    })
  })

  describe('fetchGroups', () => {
    it('fetches groups on mount', async () => {
      const { adminApi } = await import('@/api/admin')
      vi.mocked(adminApi.invitations.list).mockResolvedValue({
        data: { success: true, data: mockInvitations },
      } as any)
      vi.mocked(adminApi.groups.list).mockResolvedValue({
        data: { success: true, data: mockGroups },
      } as any)

      mount(AdminInvitations, mountOptions)
      await flushPromises()

      expect(adminApi.groups.list).toHaveBeenCalled()
    })

    it('populates group options computed property', async () => {
      const { adminApi } = await import('@/api/admin')
      vi.mocked(adminApi.invitations.list).mockResolvedValue({
        data: { success: true, data: mockInvitations },
      } as any)
      vi.mocked(adminApi.groups.list).mockResolvedValue({
        data: { success: true, data: mockGroups },
      } as any)

      const wrapper = mount(AdminInvitations, mountOptions)
      await flushPromises()

      const vm = wrapper.vm as any
      expect(vm.groupOptions).toEqual([
        { label: 'Group 1', value: '1' },
        { label: 'Group 2', value: '2' },
      ])
    })
  })

  describe('getInviteRequest', () => {
    it('returns correct request object with all fields', async () => {
      const wrapper = mount(AdminInvitations, mountOptions)
      await flushPromises()

      const vm = wrapper.vm as any
      vm.inviteForm.email = 'test@example.com'
      vm.inviteForm.username = 'testuser'
      vm.inviteForm.scopes = 'openid profile'
      vm.inviteForm.groupId = '1'
      vm.inviteForm.maxUses = '5'
      vm.inviteForm.expiresAt = 1893456000000

      const request = vm.getInviteRequest()

      expect(request.email).toBe('test@example.com')
      expect(request.username).toBe('testuser')
      expect(request.scopes).toBe('openid profile')
      expect(request.groupId).toBe('1')
      expect(request.maxUses).toBe(5)
      expect(request.expiresAt).toBeDefined()
    })

    it('omits undefined fields', async () => {
      const wrapper = mount(AdminInvitations, mountOptions)
      await flushPromises()

      const vm = wrapper.vm as any
      vm.inviteForm.email = 'test@example.com'
      vm.inviteForm.username = ''
      vm.inviteForm.scopes = ''
      vm.inviteForm.groupId = undefined
      vm.inviteForm.maxUses = ''
      vm.inviteForm.expiresAt = null

      const request = vm.getInviteRequest()

      expect(request.email).toBe('test@example.com')
      expect(request.username).toBeUndefined()
      expect(request.scopes).toBeUndefined()
      expect(request.groupId).toBeUndefined()
      expect(request.maxUses).toBeUndefined()
      expect(request.expiresAt).toBeUndefined()
    })
  })

  describe('openCreateModal', () => {
    it('resets form and opens modal', async () => {
      const wrapper = mount(AdminInvitations, mountOptions)
      await flushPromises()

      const vm = wrapper.vm as any
      vm.openCreateModal()
      await flushPromises()

      expect(vm.showInviteModal).toBe(true)
      expect(vm.inviteForm.email).toBe('')
      expect(vm.inviteForm.username).toBe('')
      expect(vm.inviteForm.scopes).toBe('')
      expect(vm.inviteForm.groupId).toBeUndefined()
      expect(vm.inviteForm.maxUses).toBe('1')
      expect(vm.inviteForm.expiresAt).toBeNull()
    })
  })

  describe('handleCreateInvite', () => {
    it('calls create API with correct request', async () => {
      const { adminApi } = await import('@/api/admin')
      vi.mocked(adminApi.invitations.list).mockResolvedValue({
        data: { success: true, data: mockInvitations },
      } as any)
      vi.mocked(adminApi.groups.list).mockResolvedValue({
        data: { success: true, data: mockGroups },
      } as any)
      vi.mocked(adminApi.invitations.create).mockResolvedValue({
        data: { success: true },
      } as any)

      const wrapper = mount(AdminInvitations, mountOptions)
      await flushPromises()

      const vm = wrapper.vm as any
      vm.inviteForm.email = 'new@example.com'
      vm.inviteForm.maxUses = '3'

      await vm.handleCreateInvite()
      await flushPromises()

      expect(adminApi.invitations.create).toHaveBeenCalled()
      expect(mockMessage.success).toHaveBeenCalledWith('邀请创建成功')
      expect(vm.showInviteModal).toBe(false)
    })

    it('shows error when create fails', async () => {
      const { adminApi } = await import('@/api/admin')
      vi.mocked(adminApi.invitations.list).mockResolvedValue({
        data: { success: true, data: mockInvitations },
      } as any)
      vi.mocked(adminApi.groups.list).mockResolvedValue({
        data: { success: true, data: mockGroups },
      } as any)
      vi.mocked(adminApi.invitations.create).mockRejectedValue({
        response: { data: { message: 'Create failed' } },
      })

      const wrapper = mount(AdminInvitations, mountOptions)
      await flushPromises()

      const vm = wrapper.vm as any
      vm.inviteForm.email = 'new@example.com'

      await vm.handleCreateInvite()
      await flushPromises()

      expect(mockMessage.error).toHaveBeenCalledWith('Create failed')
    })
  })

  describe('handleDelete', () => {
    it('calls delete API and refreshes list', async () => {
      const { adminApi } = await import('@/api/admin')
      vi.mocked(adminApi.invitations.list).mockResolvedValue({
        data: { success: true, data: mockInvitations },
      } as any)
      vi.mocked(adminApi.groups.list).mockResolvedValue({
        data: { success: true, data: mockGroups },
      } as any)
      vi.mocked(adminApi.invitations.delete).mockResolvedValue({
        data: { success: true },
      } as any)

      const wrapper = mount(AdminInvitations, mountOptions)
      await flushPromises()

      const vm = wrapper.vm as any
      await vm.handleDelete('1')
      await flushPromises()

      expect(adminApi.invitations.delete).toHaveBeenCalledWith('1')
      expect(mockMessage.success).toHaveBeenCalledWith('邀请删除成功')
    })

    it('shows error when delete fails', async () => {
      const { adminApi } = await import('@/api/admin')
      vi.mocked(adminApi.invitations.list).mockResolvedValue({
        data: { success: true, data: mockInvitations },
      } as any)
      vi.mocked(adminApi.groups.list).mockResolvedValue({
        data: { success: true, data: mockGroups },
      } as any)
      vi.mocked(adminApi.invitations.delete).mockRejectedValue({
        response: { data: { message: 'Delete failed' } },
      })

      const wrapper = mount(AdminInvitations, mountOptions)
      await flushPromises()

      const vm = wrapper.vm as any
      await vm.handleDelete('1')
      await flushPromises()

      expect(mockMessage.error).toHaveBeenCalledWith('Delete failed')
    })
  })
})
