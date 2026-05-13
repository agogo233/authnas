import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'
import AdminGroups from '../admin/Groups.vue'
import type { Group } from '@/types'

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
    groups: {
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
    template:
      '<div class="n-data-table"><div v-if="!data || data.length === 0" class="n-empty">No data</div><slot v-else /></div>',
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
    props: ['modelValue', 'type', 'placeholder', 'size', 'maxlength', 'disabled', 'readonly'],
    emits: ['update:modelValue'],
    template:
      '<input :type="type || \'text\'" :value="modelValue" :placeholder="placeholder" :maxlength="maxlength" :disabled="disabled" :readonly="readonly" @input="$emit(\'update:modelValue\', $event.target.value)" />',
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
  NDrawer: {
    name: 'NDrawer',
    props: ['show', 'width', 'placement'],
    emits: ['update:show'],
    template: '<div class="n-drawer" v-if="show"><slot /></div>',
  },
  NDrawerContent: {
    name: 'NDrawerContent',
    props: ['title'],
    template: '<div class="n-drawer-content"><slot name="header" /><slot /></div>',
  },
  NAlert: {
    name: 'NAlert',
    props: ['type'],
    template: '<div class="n-alert"><slot /></div>',
  },
  useMessage: () => mockMessage,
}))

describe('AdminGroups.vue', () => {
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

  const mockGroups: Group[] = [
    {
      id: '1',
      name: 'Admin Group',
      description: 'Administrators',
      createdAt: '2024-01-01T00:00:00Z',
    },
    {
      id: '2',
      name: 'User Group',
      description: 'Regular users',
      createdAt: '2024-01-02T00:00:00Z',
    },
  ]

  describe('rendering', () => {
    it('renders group management page correctly', async () => {
      const { adminApi } = await import('@/api/admin')
      vi.mocked(adminApi.groups.list).mockResolvedValue({
        data: { success: true, data: mockGroups },
      } as any)

      const wrapper = mount(AdminGroups, mountOptions)
      await flushPromises()

      expect(wrapper.find('.page-container').exists()).toBe(true)
      expect(wrapper.find('.n-card').exists()).toBe(true)
    })

    it('renders data table when groups exist', async () => {
      const { adminApi } = await import('@/api/admin')
      vi.mocked(adminApi.groups.list).mockResolvedValue({
        data: { success: true, data: mockGroups },
      } as any)

      const wrapper = mount(AdminGroups, mountOptions)
      await flushPromises()

      expect(wrapper.find('.n-data-table').exists()).toBe(true)
    })

    it('renders empty state when no groups', async () => {
      const { adminApi } = await import('@/api/admin')
      vi.mocked(adminApi.groups.list).mockResolvedValue({
        data: { success: true, data: [], total: 0 },
      } as any)

      const wrapper = mount(AdminGroups, mountOptions)
      await flushPromises()

      expect(wrapper.find('.n-empty').exists()).toBe(true)
    })

    it('renders create group button', async () => {
      const { adminApi } = await import('@/api/admin')
      vi.mocked(adminApi.groups.list).mockResolvedValue({
        data: { success: true, data: mockGroups },
      } as any)

      const wrapper = mount(AdminGroups, mountOptions)
      await flushPromises()

      const buttons = wrapper.findAll('button')
      const createButton = buttons.find((b) => b.text().includes('创建用户组'))
      expect(createButton?.exists()).toBe(true)
    })
  })

  describe('component state', () => {
    it('initializes with empty groups array', async () => {
      const wrapper = mount(AdminGroups, mountOptions)
      const vm = wrapper.vm as any
      expect(vm.groups).toEqual([])
    })

    it('initializes with modal closed', async () => {
      const wrapper = mount(AdminGroups, mountOptions)
      const vm = wrapper.vm as any
      expect(vm.showGroupModal).toBe(false)
      expect(vm.showMembersDrawer).toBe(false)
    })

    it('initializes with editingGroup as null', async () => {
      const wrapper = mount(AdminGroups, mountOptions)
      const vm = wrapper.vm as any
      expect(vm.editingGroup).toBeNull()
    })
  })

  describe('fetchGroups', () => {
    it('fetches groups on mount', async () => {
      const { adminApi } = await import('@/api/admin')
      vi.mocked(adminApi.groups.list).mockResolvedValue({
        data: { success: true, data: mockGroups },
      } as any)

      mount(AdminGroups, mountOptions)
      await flushPromises()

      expect(adminApi.groups.list).toHaveBeenCalled()
    })

    it('shows error on fetch failure', async () => {
      const { adminApi } = await import('@/api/admin')
      vi.mocked(adminApi.groups.list).mockRejectedValue({
        response: { data: { message: 'Failed to fetch groups' } },
      })

      mount(AdminGroups, mountOptions)
      await flushPromises()

      expect(mockMessage.error).toHaveBeenCalledWith('Failed to fetch groups')
    })

    it('sets loading state during fetch', async () => {
      const { adminApi } = await import('@/api/admin')
      let resolve: any
      vi.mocked(adminApi.groups.list).mockReturnValue(
        new Promise((r) => {
          resolve = r
        })
      )

      const wrapper = mount(AdminGroups, mountOptions)
      const vm = wrapper.vm as any

      expect(vm.loading).toBe(true)

      resolve({ data: { success: true, data: mockGroups } })
      await flushPromises()

      expect(vm.loading).toBe(false)
    })
  })

  describe('openCreateModal', () => {
    it('opens modal with empty form for new group', async () => {
      const wrapper = mount(AdminGroups, mountOptions)
      await flushPromises()

      const vm = wrapper.vm as any
      vm.openCreateModal()
      await flushPromises()

      expect(vm.showGroupModal).toBe(true)
      expect(vm.editingGroup).toBeNull()
      expect(vm.groupForm.name).toBe('')
      expect(vm.groupForm.description).toBe('')
    })
  })

  describe('openEditModal', () => {
    it('opens modal with group data for editing', async () => {
      const wrapper = mount(AdminGroups, mountOptions)
      await flushPromises()

      const vm = wrapper.vm as any
      vm.openEditModal(mockGroups[0])
      await flushPromises()

      expect(vm.showGroupModal).toBe(true)
      expect(vm.editingGroup).toEqual(mockGroups[0])
      expect(vm.groupForm.name).toBe(mockGroups[0].name)
      expect(vm.groupForm.description).toBe(mockGroups[0].description)
    })
  })

  describe('handleSaveGroup - create', () => {
    it('calls create API when creating new group', async () => {
      const { adminApi } = await import('@/api/admin')
      vi.mocked(adminApi.groups.list).mockResolvedValue({
        data: { success: true, data: mockGroups },
      } as any)
      vi.mocked(adminApi.groups.create).mockResolvedValue({
        data: { success: true },
      } as any)

      const wrapper = mount(AdminGroups, mountOptions)
      await flushPromises()

      const vm = wrapper.vm as any
      vm.openCreateModal()
      await flushPromises()

      vm.groupForm.name = 'New Group'
      vm.groupForm.description = 'New Description'

      await vm.handleSaveGroup()
      await flushPromises()

      expect(adminApi.groups.create).toHaveBeenCalledWith({
        name: 'New Group',
        description: 'New Description',
      })
      expect(mockMessage.success).toHaveBeenCalledWith('用户组创建成功')
      expect(vm.showGroupModal).toBe(false)
    })

    it('shows error when create fails', async () => {
      const { adminApi } = await import('@/api/admin')
      vi.mocked(adminApi.groups.list).mockResolvedValue({
        data: { success: true, data: mockGroups },
      } as any)
      vi.mocked(adminApi.groups.create).mockRejectedValue({
        response: { data: { message: 'Create failed' } },
      })

      const wrapper = mount(AdminGroups, mountOptions)
      await flushPromises()

      const vm = wrapper.vm as any
      vm.openCreateModal()
      await flushPromises()

      vm.groupForm.name = 'New Group'

      await vm.handleSaveGroup()
      await flushPromises()

      expect(mockMessage.error).toHaveBeenCalledWith('Create failed')
    })
  })

  describe('handleSaveGroup - update', () => {
    it('calls update API when updating existing group', async () => {
      const { adminApi } = await import('@/api/admin')
      vi.mocked(adminApi.groups.list).mockResolvedValue({
        data: { success: true, data: mockGroups },
      } as any)
      vi.mocked(adminApi.groups.update).mockResolvedValue({
        data: { success: true },
      } as any)

      const wrapper = mount(AdminGroups, mountOptions)
      await flushPromises()

      const vm = wrapper.vm as any
      vm.openEditModal(mockGroups[0])
      await flushPromises()

      vm.groupForm.name = 'Updated Name'

      await vm.handleSaveGroup()
      await flushPromises()

      expect(adminApi.groups.update).toHaveBeenCalledWith(mockGroups[0].id, {
        name: 'Updated Name',
        description: mockGroups[0].description || '',
      })
      expect(mockMessage.success).toHaveBeenCalledWith('用户组更新成功')
    })

    it('shows error when update fails', async () => {
      const { adminApi } = await import('@/api/admin')
      vi.mocked(adminApi.groups.list).mockResolvedValue({
        data: { success: true, data: mockGroups },
      } as any)
      vi.mocked(adminApi.groups.update).mockRejectedValue({
        response: { data: { message: 'Update failed' } },
      })

      const wrapper = mount(AdminGroups, mountOptions)
      await flushPromises()

      const vm = wrapper.vm as any
      vm.openEditModal(mockGroups[0])
      await flushPromises()

      await vm.handleSaveGroup()
      await flushPromises()

      expect(mockMessage.error).toHaveBeenCalledWith('Update failed')
    })
  })

  describe('handleDelete', () => {
    it('calls delete API and refreshes list', async () => {
      const { adminApi } = await import('@/api/admin')
      vi.mocked(adminApi.groups.list).mockResolvedValue({
        data: { success: true, data: mockGroups },
      } as any)
      vi.mocked(adminApi.groups.delete).mockResolvedValue({
        data: { success: true },
      } as any)

      const wrapper = mount(AdminGroups, mountOptions)
      await flushPromises()

      const vm = wrapper.vm as any
      await vm.handleDelete('1')
      await flushPromises()

      expect(adminApi.groups.delete).toHaveBeenCalledWith('1')
      expect(mockMessage.success).toHaveBeenCalledWith('用户组删除成功')
    })

    it('shows error when delete fails', async () => {
      const { adminApi } = await import('@/api/admin')
      vi.mocked(adminApi.groups.list).mockResolvedValue({
        data: { success: true, data: mockGroups },
      } as any)
      vi.mocked(adminApi.groups.delete).mockRejectedValue({
        response: { data: { message: 'Delete failed' } },
      })

      const wrapper = mount(AdminGroups, mountOptions)
      await flushPromises()

      const vm = wrapper.vm as any
      await vm.handleDelete('1')
      await flushPromises()

      expect(mockMessage.error).toHaveBeenCalledWith('Delete failed')
    })
  })

  describe('openMembersDrawer', () => {
    it('opens drawer with selected group', async () => {
      const wrapper = mount(AdminGroups, mountOptions)
      await flushPromises()

      const vm = wrapper.vm as any
      vm.openMembersDrawer(mockGroups[0])
      await flushPromises()

      expect(vm.showMembersDrawer).toBe(true)
      expect(vm.selectedGroup).toEqual(mockGroups[0])
      expect(vm.groupMembers).toEqual([])
    })
  })
})
