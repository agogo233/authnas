import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'
import AdminUsers from '../admin/Users.vue'
import type { User } from '@/types'

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
    users: {
      list: vi.fn(),
      create: vi.fn(),
      update: vi.fn(),
      delete: vi.fn(),
      approve: vi.fn(),
      resetPassword: vi.fn(),
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
    props: ['columns', 'data', 'loading', 'bordered', 'pagination'],
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
    emits: ['update:modelValue', 'keyup'],
    template:
      '<input :type="type || \'text\'" :value="modelValue" :placeholder="placeholder" :maxlength="maxlength" :disabled="disabled" :readonly="readonly" @input="$emit(\'update:modelValue\', $event.target.value)" @keyup="$emit(\'keyup\', $event)" />',
  },
  NSwitch: {
    name: 'NSwitch',
    props: ['modelValue'],
    emits: ['update:modelValue'],
    template: '<div class="n-switch" @click="$emit(\'update:modelValue\', !modelValue)"></div>',
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
  NInputGroup: {
    name: 'NInputGroup',
    template: '<div class="n-input-group"><slot /></div>',
  },
  NPagination: {
    name: 'NPagination',
    props: ['page', 'pageSize', 'pageSizes', 'itemCount', 'showSizePicker'],
    emits: ['update:page', 'update:pageSize'],
    template: '<div class="n-pagination"></div>',
  },
  useMessage: () => mockMessage,
}))

describe('AdminUsers.vue', () => {
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

  const mockUsers: User[] = [
    {
      id: '1',
      username: 'admin',
      email: 'admin@example.com',
      name: 'Admin User',
      isAdmin: true,
      approved: true,
      mfaRequired: false,
      emailVerified: true,
      createdAt: '2024-01-01T00:00:00Z',
    },
    {
      id: '2',
      username: 'testuser',
      email: 'test@example.com',
      name: 'Test User',
      isAdmin: false,
      approved: false,
      mfaRequired: false,
      emailVerified: true,
      createdAt: '2024-01-02T00:00:00Z',
    },
  ]

  describe('rendering', () => {
    it('renders user management page correctly', async () => {
      const { adminApi } = await import('@/api/admin')
      vi.mocked(adminApi.users.list).mockResolvedValue({
        data: { success: true, data: mockUsers, total: 2 },
      } as any)

      const wrapper = mount(AdminUsers, mountOptions)
      await flushPromises()

      expect(wrapper.find('.page-container').exists()).toBe(true)
      expect(wrapper.find('.n-card').exists()).toBe(true)
    })

    it('renders data table when users exist', async () => {
      const { adminApi } = await import('@/api/admin')
      vi.mocked(adminApi.users.list).mockResolvedValue({
        data: { success: true, data: mockUsers, total: 2 },
      } as any)

      const wrapper = mount(AdminUsers, mountOptions)
      await flushPromises()

      expect(wrapper.find('.n-data-table').exists()).toBe(true)
    })

    it('renders empty state when no users', async () => {
      const { adminApi } = await import('@/api/admin')
      vi.mocked(adminApi.users.list).mockResolvedValue({
        data: { success: true, data: [], total: 0 },
      } as any)

      const wrapper = mount(AdminUsers, mountOptions)
      await flushPromises()

      expect(wrapper.find('.n-empty').exists()).toBe(true)
    })

    it('renders create user button', async () => {
      const { adminApi } = await import('@/api/admin')
      vi.mocked(adminApi.users.list).mockResolvedValue({
        data: { success: true, data: mockUsers, total: 2 },
      } as any)

      const wrapper = mount(AdminUsers, mountOptions)
      await flushPromises()

      const buttons = wrapper.findAll('button')
      const createButton = buttons.find((b) => b.text().includes('创建用户'))
      expect(createButton?.exists()).toBe(true)
    })

    it('renders search input', async () => {
      const { adminApi } = await import('@/api/admin')
      vi.mocked(adminApi.users.list).mockResolvedValue({
        data: { success: true, data: mockUsers, total: 2 },
      } as any)

      const wrapper = mount(AdminUsers, mountOptions)
      await flushPromises()

      const searchInput = wrapper.find('input')
      expect(searchInput.exists()).toBe(true)
    })

    it('renders pagination controls', async () => {
      const { adminApi } = await import('@/api/admin')
      vi.mocked(adminApi.users.list).mockResolvedValue({
        data: { success: true, data: mockUsers, total: 2 },
      } as any)

      const wrapper = mount(AdminUsers, mountOptions)
      await flushPromises()

      const pagination = wrapper.findAll('.n-pagination')
      expect(pagination.length).toBeGreaterThan(0)
    })
  })

  describe('component state', () => {
    it('initializes with empty search query', async () => {
      const wrapper = mount(AdminUsers, mountOptions)
      const vm = wrapper.vm as any
      expect(vm.searchQuery).toBe('')
    })

    it('initializes with default pagination', async () => {
      const wrapper = mount(AdminUsers, mountOptions)
      const vm = wrapper.vm as any
      expect(vm.page).toBe(1)
      expect(vm.pageSize).toBe(10)
    })

    it('initializes with modal closed', async () => {
      const wrapper = mount(AdminUsers, mountOptions)
      const vm = wrapper.vm as any
      expect(vm.showUserModal).toBe(false)
      expect(vm.showPasswordModal).toBe(false)
    })
  })

  describe('component logic', () => {
    it('fetches users on mount', async () => {
      const { adminApi } = await import('@/api/admin')
      vi.mocked(adminApi.users.list).mockResolvedValue({
        data: { success: true, data: mockUsers, total: 2 },
      } as any)

      mount(AdminUsers, mountOptions)
      await flushPromises()

      expect(adminApi.users.list).toHaveBeenCalledWith({
        page: 1,
        pageSize: 10,
        search: '',
      })
    })

    it('shows error on fetch failure', async () => {
      const { adminApi } = await import('@/api/admin')
      vi.mocked(adminApi.users.list).mockRejectedValue({
        response: { data: { message: 'Failed to fetch users' } },
      })

      const wrapper = mount(AdminUsers, mountOptions)
      await flushPromises()

      expect(mockMessage.error).toHaveBeenCalledWith('Failed to fetch users')
    })
  })

  describe('user modal operations', () => {
    it('opens create modal when clicking create button', async () => {
      const { adminApi } = await import('@/api/admin')
      vi.mocked(adminApi.users.list).mockResolvedValue({
        data: { success: true, data: mockUsers, total: 2 },
      } as any)

      const wrapper = mount(AdminUsers, mountOptions)
      await flushPromises()

      const createButton = wrapper.findAll('button').find((b) => b.text() === '创建用户')
      expect(createButton).toBeDefined()

      await createButton!.trigger('click')
      await flushPromises()

      const vm = wrapper.vm as any
      expect(vm.showUserModal).toBe(true)
      expect(vm.editingUser).toBeNull()
    })

    it('opens edit modal with user data', async () => {
      const { adminApi } = await import('@/api/admin')
      vi.mocked(adminApi.users.list).mockResolvedValue({
        data: { success: true, data: mockUsers, total: 2 },
      } as any)

      const wrapper = mount(AdminUsers, mountOptions)
      await flushPromises()

      // Trigger openEditModal through component method
      const vm = wrapper.vm as any
      vm.openEditModal(mockUsers[0])
      await flushPromises()

      expect(vm.showUserModal).toBe(true)
      expect(vm.editingUser).toEqual(mockUsers[0])
    })

    it('closes modal when cancel is clicked', async () => {
      const { adminApi } = await import('@/api/admin')
      vi.mocked(adminApi.users.list).mockResolvedValue({
        data: { success: true, data: mockUsers, total: 2 },
      } as any)

      const wrapper = mount(AdminUsers, mountOptions)
      await flushPromises()

      // Open modal first
      const vm = wrapper.vm as any
      vm.openCreateModal()
      await flushPromises()

      // Close modal
      vm.showUserModal = false
      await flushPromises()

      expect(vm.showUserModal).toBe(false)
    })
  })

  describe('user creation', () => {
    it('calls create API when saving new user', async () => {
      const { adminApi } = await import('@/api/admin')
      vi.mocked(adminApi.users.list).mockResolvedValue({
        data: { success: true, data: mockUsers, total: 2 },
      } as any)
      vi.mocked(adminApi.users.create).mockResolvedValue({
        data: { success: true },
      } as any)

      const wrapper = mount(AdminUsers, mountOptions)
      await flushPromises()

      const vm = wrapper.vm as any
      vm.openCreateModal()
      await flushPromises()

      vm.userForm.email = 'new@example.com'
      vm.userForm.username = 'newuser'
      vm.userForm.password = 'Password123!'
      vm.userForm.name = 'New User'
      vm.userForm.isAdmin = false
      vm.userForm.approved = true
      vm.userForm.mfaRequired = false

      await vm.handleSaveUser()
      await flushPromises()

      expect(adminApi.users.create).toHaveBeenCalled()
      expect(mockMessage.success).toHaveBeenCalledWith('用户创建成功')
      expect(vm.showUserModal).toBe(false)
    })

    it('shows error when create fails', async () => {
      const { adminApi } = await import('@/api/admin')
      vi.mocked(adminApi.users.list).mockResolvedValue({
        data: { success: true, data: mockUsers, total: 2 },
      } as any)
      vi.mocked(adminApi.users.create).mockRejectedValue({
        response: { data: { message: 'Create failed' } },
      })

      const wrapper = mount(AdminUsers, mountOptions)
      await flushPromises()

      const vm = wrapper.vm as any
      vm.openCreateModal()
      await flushPromises()

      vm.userForm.email = 'new@example.com'
      vm.userForm.username = 'newuser'
      vm.userForm.password = 'Password123!'

      await vm.handleSaveUser()
      await flushPromises()

      expect(mockMessage.error).toHaveBeenCalledWith('Create failed')
    })
  })

  describe('user update', () => {
    it('calls update API when saving edited user', async () => {
      const { adminApi } = await import('@/api/admin')
      vi.mocked(adminApi.users.list).mockResolvedValue({
        data: { success: true, data: mockUsers, total: 2 },
      } as any)
      vi.mocked(adminApi.users.update).mockResolvedValue({
        data: { success: true },
      } as any)

      const wrapper = mount(AdminUsers, mountOptions)
      await flushPromises()

      const vm = wrapper.vm as any
      vm.openEditModal(mockUsers[0])
      await flushPromises()

      vm.userForm.name = 'Updated Name'

      await vm.handleSaveUser()
      await flushPromises()

      expect(adminApi.users.update).toHaveBeenCalled()
      expect(mockMessage.success).toHaveBeenCalledWith('用户更新成功')
    })
  })

  describe('user deletion', () => {
    it('calls delete API when confirming delete', async () => {
      const { adminApi } = await import('@/api/admin')
      vi.mocked(adminApi.users.list).mockResolvedValue({
        data: { success: true, data: mockUsers, total: 2 },
      } as any)
      vi.mocked(adminApi.users.delete).mockResolvedValue({
        data: { success: true },
      } as any)

      const wrapper = mount(AdminUsers, mountOptions)
      await flushPromises()

      const vm = wrapper.vm as any
      await vm.handleDelete('1')
      await flushPromises()

      expect(adminApi.users.delete).toHaveBeenCalledWith('1')
      expect(mockMessage.success).toHaveBeenCalledWith('用户删除成功')
    })

    it('shows error when delete fails', async () => {
      const { adminApi } = await import('@/api/admin')
      vi.mocked(adminApi.users.list).mockResolvedValue({
        data: { success: true, data: mockUsers, total: 2 },
      } as any)
      vi.mocked(adminApi.users.delete).mockRejectedValue({
        response: { data: { message: 'Delete failed' } },
      })

      const wrapper = mount(AdminUsers, mountOptions)
      await flushPromises()

      const vm = wrapper.vm as any
      await vm.handleDelete('1')
      await flushPromises()

      expect(mockMessage.error).toHaveBeenCalledWith('Delete failed')
    })
  })

  describe('user approval', () => {
    it('calls approve API when approving user', async () => {
      const { adminApi } = await import('@/api/admin')
      vi.mocked(adminApi.users.list).mockResolvedValue({
        data: { success: true, data: mockUsers, total: 2 },
      } as any)
      vi.mocked(adminApi.users.approve).mockResolvedValue({
        data: { success: true },
      } as any)

      const wrapper = mount(AdminUsers, mountOptions)
      await flushPromises()

      const vm = wrapper.vm as any
      await vm.handleApprove('2')
      await flushPromises()

      expect(adminApi.users.approve).toHaveBeenCalledWith('2', { approved: true })
      expect(mockMessage.success).toHaveBeenCalledWith('用户批准成功')
    })
  })

  describe('password reset', () => {
    it('opens password modal when clicking reset password', async () => {
      const wrapper = mount(AdminUsers, mountOptions)
      await flushPromises()

      const vm = wrapper.vm as any
      vm.openPasswordModal('1')
      await flushPromises()

      expect(vm.showPasswordModal).toBe(true)
      expect(vm.resetPasswordUserId).toBe('1')
      expect(vm.newPassword).toBe('')
    })

    it('calls resetPassword API when resetting password', async () => {
      const { adminApi } = await import('@/api/admin')
      vi.mocked(adminApi.users.list).mockResolvedValue({
        data: { success: true, data: mockUsers, total: 2 },
      } as any)
      vi.mocked(adminApi.users.resetPassword).mockResolvedValue({
        data: { success: true },
      } as any)

      const wrapper = mount(AdminUsers, mountOptions)
      await flushPromises()

      const vm = wrapper.vm as any
      vm.openPasswordModal('1')
      vm.newPassword = 'NewPassword123!'

      await vm.handleResetPassword()
      await flushPromises()

      expect(adminApi.users.resetPassword).toHaveBeenCalledWith('1', {
        newPassword: 'NewPassword123!',
      })
      expect(mockMessage.success).toHaveBeenCalledWith('密码重置成功')
      expect(vm.showPasswordModal).toBe(false)
    })
  })

  describe('search functionality', () => {
    it('debounces search on input', async () => {
      const { adminApi } = await import('@/api/admin')
      vi.mocked(adminApi.users.list).mockResolvedValue({
        data: { success: true, data: mockUsers, total: 2 },
      } as any)

      const wrapper = mount(AdminUsers, mountOptions)
      await flushPromises()

      const vm = wrapper.vm as any
      vm.searchQuery = 'test'

      // Wait for debounce
      await new Promise((resolve) => setTimeout(resolve, 400))
      await flushPromises()

      expect(adminApi.users.list).toHaveBeenCalledWith(expect.objectContaining({ search: 'test' }))
    })

    it('resets page when search changes', async () => {
      const wrapper = mount(AdminUsers, mountOptions)
      await flushPromises()

      const vm = wrapper.vm as any
      vm.page = 5
      vm.searchQuery = 'test'

      // Trigger watch
      await new Promise((resolve) => setTimeout(resolve, 400))
      await flushPromises()

      expect(vm.page).toBe(1)
    })
  })

  describe('pagination', () => {
    it('calls API with new page when page changes', async () => {
      const { adminApi } = await import('@/api/admin')
      vi.mocked(adminApi.users.list).mockResolvedValue({
        data: { success: true, data: mockUsers, total: 2 },
      } as any)

      const wrapper = mount(AdminUsers, mountOptions)
      await flushPromises()

      const vm = wrapper.vm as any
      vm.handlePageChange(3)
      await flushPromises()

      expect(adminApi.users.list).toHaveBeenCalledWith(expect.objectContaining({ page: 3 }))
    })

    it('resets page when page size changes', async () => {
      const wrapper = mount(AdminUsers, mountOptions)
      await flushPromises()

      const vm = wrapper.vm as any
      vm.handlePageSizeChange(20)
      await flushPromises()

      expect(vm.page).toBe(1)
      expect(vm.pageSize).toBe(20)
    })
  })

  describe('reset password modal', () => {
    it('shows error when reset password fails', async () => {
      const { adminApi } = await import('@/api/admin')
      vi.mocked(adminApi.users.list).mockResolvedValue({
        data: { success: true, data: mockUsers, total: 2 },
      } as any)
      vi.mocked(adminApi.users.resetPassword).mockRejectedValue({
        response: { data: { message: 'Reset password failed' } },
      })

      const wrapper = mount(AdminUsers, mountOptions)
      await flushPromises()

      const vm = wrapper.vm as any
      vm.openPasswordModal('1')
      vm.newPassword = 'NewPassword123!'
      await vm.handleResetPassword()
      await flushPromises()

      expect(mockMessage.error).toHaveBeenCalledWith('Reset password failed')
    })
  })

  describe('approve error handling', () => {
    it('shows error when approve fails', async () => {
      const { adminApi } = await import('@/api/admin')
      vi.mocked(adminApi.users.list).mockResolvedValue({
        data: { success: true, data: mockUsers, total: 2 },
      } as any)
      vi.mocked(adminApi.users.approve).mockRejectedValue({
        response: { data: { message: 'Approve failed' } },
      })

      const wrapper = mount(AdminUsers, mountOptions)
      await flushPromises()

      const vm = wrapper.vm as any
      await vm.handleApprove('2')
      await flushPromises()

      expect(mockMessage.error).toHaveBeenCalledWith('Approve failed')
    })
  })

  describe('update error handling', () => {
    it('shows error when update fails', async () => {
      const { adminApi } = await import('@/api/admin')
      vi.mocked(adminApi.users.list).mockResolvedValue({
        data: { success: true, data: mockUsers, total: 2 },
      } as any)
      vi.mocked(adminApi.users.update).mockRejectedValue({
        response: { data: { message: 'Update failed' } },
      })

      const wrapper = mount(AdminUsers, mountOptions)
      await flushPromises()

      const vm = wrapper.vm as any
      vm.openEditModal(mockUsers[0])
      await flushPromises()
      vm.userForm.name = 'Updated Name'
      await vm.handleSaveUser()
      await flushPromises()

      expect(mockMessage.error).toHaveBeenCalledWith('Update failed')
    })
  })
})
