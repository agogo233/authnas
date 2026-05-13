import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import AdminDashboard from '../admin/Dashboard.vue'
import * as adminModule from '@/api/admin'

const mockMessage = {
  success: vi.fn(),
  error: vi.fn(),
  warning: vi.fn(),
  info: vi.fn(),
}

vi.mock('@/api/admin', () => ({
  adminApi: {
    users: {
      count: vi.fn(),
    },
    groups: {
      list: vi.fn(),
    },
    clients: {
      list: vi.fn(),
    },
    invitations: {
      list: vi.fn(),
    },
  },
}))

vi.mock('naive-ui', () => ({
  NCard: {
    name: 'NCard',
    props: ['title', 'hoverable', 'style'],
    template: '<div class="n-card"><slot /><slot name="header-extra" /></div>',
  },
  NGrid: {
    name: 'NGrid',
    props: ['cols', 'xGap', 'yGap', 'responsive', 'itemResponsive'],
    template: '<div class="n-grid"><slot /></div>',
  },
  NGi: {
    name: 'NGi',
    template: '<div class="n-gi"><slot /></div>',
  },
  NStatistic: {
    name: 'NStatistic',
    props: ['label', 'value'],
    template:
      '<div class="n-statistic"><span class="label">{{ label }}</span><span class="value">{{ value }}</span></div>',
  },
  NSpin: {
    name: 'NSpin',
    props: ['show'],
    template: '<div class="n-spin" v-if="show"><slot /></div>',
  },
  useMessage: () => mockMessage,
}))

vi.mock('vue-router', () => ({
  useRouter: () => ({ push: vi.fn() }),
}))

describe('AdminDashboard.vue', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  const mockUsersRes = { data: { success: true, data: { total: 100 } } }
  const mockGroupsRes = { data: { success: true, data: [{ id: '1', name: 'Group A' }] } }
  const mockClientsRes = { data: { success: true, data: [{ id: '1', name: 'Client A' }] } }
  const mockInvitationsRes = {
    data: {
      success: true,
      data: [
        { id: '1', expiresAt: new Date(Date.now() + 86400000).toISOString() },
        { id: '2', expiresAt: new Date(Date.now() - 86400000).toISOString() },
      ],
    },
  }

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
    it('renders dashboard title', async () => {
      vi.mocked(adminModule.adminApi.users.count).mockResolvedValue(mockUsersRes as any)
      vi.mocked(adminModule.adminApi.groups.list).mockResolvedValue(mockGroupsRes as any)
      vi.mocked(adminModule.adminApi.clients.list).mockResolvedValue(mockClientsRes as any)
      vi.mocked(adminModule.adminApi.invitations.list).mockResolvedValue(mockInvitationsRes as any)

      const wrapper = mount(AdminDashboard, mountOptions)
      await flushPromises()

      expect(wrapper.find('h1').text()).toBe('管理后台')
    })

    it('renders page container', async () => {
      vi.mocked(adminModule.adminApi.users.count).mockResolvedValue(mockUsersRes as any)
      vi.mocked(adminModule.adminApi.groups.list).mockResolvedValue(mockGroupsRes as any)
      vi.mocked(adminModule.adminApi.clients.list).mockResolvedValue(mockClientsRes as any)
      vi.mocked(adminModule.adminApi.invitations.list).mockResolvedValue(mockInvitationsRes as any)

      const wrapper = mount(AdminDashboard, mountOptions)
      await flushPromises()

      expect(wrapper.find('.page-container').exists()).toBe(true)
    })
  })

  describe('data fetching', () => {
    it('fetches all data on mount', async () => {
      vi.mocked(adminModule.adminApi.users.count).mockResolvedValue(mockUsersRes as any)
      vi.mocked(adminModule.adminApi.groups.list).mockResolvedValue(mockGroupsRes as any)
      vi.mocked(adminModule.adminApi.clients.list).mockResolvedValue(mockClientsRes as any)
      vi.mocked(adminModule.adminApi.invitations.list).mockResolvedValue(mockInvitationsRes as any)

      mount(AdminDashboard, mountOptions)
      await flushPromises()

      expect(adminModule.adminApi.users.count).toHaveBeenCalled()
      expect(adminModule.adminApi.groups.list).toHaveBeenCalled()
      expect(adminModule.adminApi.clients.list).toHaveBeenCalled()
      expect(adminModule.adminApi.invitations.list).toHaveBeenCalled()
    })

    it('stores user count in state', async () => {
      vi.mocked(adminModule.adminApi.users.count).mockResolvedValue({
        data: { success: true, data: { total: 42 } },
      } as any)
      vi.mocked(adminModule.adminApi.groups.list).mockResolvedValue(mockGroupsRes as any)
      vi.mocked(adminModule.adminApi.clients.list).mockResolvedValue(mockClientsRes as any)
      vi.mocked(adminModule.adminApi.invitations.list).mockResolvedValue(mockInvitationsRes as any)

      const wrapper = mount(AdminDashboard, mountOptions)
      await flushPromises()

      const vm = wrapper.vm as any
      expect(vm.stats.users).toBe(42)
    })

    it('stores groups count in state', async () => {
      vi.mocked(adminModule.adminApi.users.count).mockResolvedValue(mockUsersRes as any)
      vi.mocked(adminModule.adminApi.groups.list).mockResolvedValue({
        data: { success: true, data: [{ id: '1' }, { id: '2' }] },
      } as any)
      vi.mocked(adminModule.adminApi.clients.list).mockResolvedValue(mockClientsRes as any)
      vi.mocked(adminModule.adminApi.invitations.list).mockResolvedValue(mockInvitationsRes as any)

      const wrapper = mount(AdminDashboard, mountOptions)
      await flushPromises()

      const vm = wrapper.vm as any
      expect(vm.stats.groups).toBe(2)
    })

    it('stores clients count in state', async () => {
      vi.mocked(adminModule.adminApi.users.count).mockResolvedValue(mockUsersRes as any)
      vi.mocked(adminModule.adminApi.groups.list).mockResolvedValue(mockGroupsRes as any)
      vi.mocked(adminModule.adminApi.clients.list).mockResolvedValue({
        data: { success: true, data: [{ id: '1' }, { id: '2' }, { id: '3' }] },
      } as any)
      vi.mocked(adminModule.adminApi.invitations.list).mockResolvedValue(mockInvitationsRes as any)

      const wrapper = mount(AdminDashboard, mountOptions)
      await flushPromises()

      const vm = wrapper.vm as any
      expect(vm.stats.clients).toBe(3)
    })

    it('calculates active invitations correctly', async () => {
      const futureDate = new Date(Date.now() + 86400000).toISOString()
      const pastDate = new Date(Date.now() - 86400000).toISOString()
      const customInvitationsRes = {
        data: {
          success: true,
          data: [
            { id: '1', expiresAt: futureDate },
            { id: '2', expiresAt: futureDate },
            { id: '3', expiresAt: pastDate },
          ],
        },
      }

      vi.mocked(adminModule.adminApi.users.count).mockResolvedValue(mockUsersRes as any)
      vi.mocked(adminModule.adminApi.groups.list).mockResolvedValue(mockGroupsRes as any)
      vi.mocked(adminModule.adminApi.clients.list).mockResolvedValue(mockClientsRes as any)
      vi.mocked(adminModule.adminApi.invitations.list).mockResolvedValue(
        customInvitationsRes as any
      )

      const wrapper = mount(AdminDashboard, mountOptions)
      await flushPromises()

      const vm = wrapper.vm as any
      expect(vm.stats.activeInvitations).toBe(2)
    })

    it('stores total invitations count', async () => {
      vi.mocked(adminModule.adminApi.users.count).mockResolvedValue(mockUsersRes as any)
      vi.mocked(adminModule.adminApi.groups.list).mockResolvedValue(mockGroupsRes as any)
      vi.mocked(adminModule.adminApi.clients.list).mockResolvedValue(mockClientsRes as any)
      vi.mocked(adminModule.adminApi.invitations.list).mockResolvedValue({
        data: { success: true, data: [{ id: '1' }, { id: '2' }, { id: '3' }] },
      } as any)

      const wrapper = mount(AdminDashboard, mountOptions)
      await flushPromises()

      const vm = wrapper.vm as any
      expect(vm.stats.invitations).toBe(3)
    })
  })

  describe('error handling', () => {
    it('shows error message on API failure', async () => {
      vi.mocked(adminModule.adminApi.users.count).mockRejectedValue({
        response: { data: { message: 'Failed to fetch users' } },
      })
      vi.mocked(adminModule.adminApi.groups.list).mockRejectedValue(new Error('Network error'))
      vi.mocked(adminModule.adminApi.clients.list).mockRejectedValue(new Error('Network error'))
      vi.mocked(adminModule.adminApi.invitations.list).mockRejectedValue(new Error('Network error'))

      mount(AdminDashboard, mountOptions)
      await flushPromises()

      expect(mockMessage.error).toHaveBeenCalled()
    })

    it('renders without crashing on error', async () => {
      vi.mocked(adminModule.adminApi.users.count).mockRejectedValue(new Error('Error'))
      vi.mocked(adminModule.adminApi.groups.list).mockRejectedValue(new Error('Error'))
      vi.mocked(adminModule.adminApi.clients.list).mockRejectedValue(new Error('Error'))
      vi.mocked(adminModule.adminApi.invitations.list).mockRejectedValue(new Error('Error'))

      const wrapper = mount(AdminDashboard, mountOptions)
      await flushPromises()

      expect(wrapper.find('h1').text()).toBe('管理后台')
    })
  })

  describe('loading state', () => {
    it('shows loading state initially', async () => {
      vi.mocked(adminModule.adminApi.users.count).mockImplementation(
        () => new Promise(() => {}) as any
      )
      vi.mocked(adminModule.adminApi.groups.list).mockImplementation(
        () => new Promise(() => {}) as any
      )
      vi.mocked(adminModule.adminApi.clients.list).mockImplementation(
        () => new Promise(() => {}) as any
      )
      vi.mocked(adminModule.adminApi.invitations.list).mockImplementation(
        () => new Promise(() => {}) as any
      )

      const wrapper = mount(AdminDashboard, mountOptions)
      await flushPromises()

      expect(wrapper.findComponent({ name: 'NSpin' }).exists()).toBe(true)
    })

    it('renders dashboard after loading completes', async () => {
      vi.mocked(adminModule.adminApi.users.count).mockResolvedValue(mockUsersRes as any)
      vi.mocked(adminModule.adminApi.groups.list).mockResolvedValue(mockGroupsRes as any)
      vi.mocked(adminModule.adminApi.clients.list).mockResolvedValue(mockClientsRes as any)
      vi.mocked(adminModule.adminApi.invitations.list).mockResolvedValue(mockInvitationsRes as any)

      const wrapper = mount(AdminDashboard, mountOptions)
      await flushPromises()

      expect(wrapper.find('h1').text()).toBe('管理后台')
    })
  })

  describe('initial state', () => {
    it('initializes with zero counts', async () => {
      vi.mocked(adminModule.adminApi.users.count).mockResolvedValue({
        data: { success: true, data: { total: 0 } },
      } as any)
      vi.mocked(adminModule.adminApi.groups.list).mockResolvedValue({
        data: { success: true, data: [] },
      } as any)
      vi.mocked(adminModule.adminApi.clients.list).mockResolvedValue({
        data: { success: true, data: [] },
      } as any)
      vi.mocked(adminModule.adminApi.invitations.list).mockResolvedValue({
        data: { success: true, data: [] },
      } as any)

      const wrapper = mount(AdminDashboard, mountOptions)
      const vm = wrapper.vm as any

      expect(vm.stats.users).toBe(0)
      expect(vm.stats.groups).toBe(0)
      expect(vm.stats.clients).toBe(0)
      expect(vm.stats.invitations).toBe(0)
      expect(vm.stats.activeInvitations).toBe(0)
    })
  })
})
