import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { mount } from '@vue/test-utils'
import AdminLayout from '../AdminLayout.vue'

// Mock vue-router
const mockPush = vi.fn()
vi.mock('vue-router', () => ({
  useRouter: () => ({
    push: mockPush,
  }),
  useRoute: () => ({
    path: '/admin',
  }),
}))

// Mock naive-ui
vi.mock('naive-ui', () => ({
  NIcon: {
    name: 'NIcon',
    props: ['component', 'size'],
    template: '<span class="n-icon"><slot /></span>',
  },
}))

// Mock icons
vi.mock('../icons', () => ({
  DashboardIcon: 'DashboardIcon',
  PeopleOutlineIcon: 'PeopleOutlineIcon',
  FolderOutlineIcon: 'FolderOutlineIcon',
  ColorPaletteOutlineIcon: 'ColorPaletteOutlineIcon',
  MailOutlineIcon: 'MailOutlineIcon',
  ShieldIcon: 'ShieldIcon',
  SettingsOutlineIcon: 'SettingsOutlineIcon',
  ReturnUpBackOutlineIcon: 'ReturnUpBackOutlineIcon',
}))

describe('AdminLayout', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  afterEach(() => {
    vi.restoreAllMocks()
  })

  it('renders the admin layout correctly', () => {
    const wrapper = mount(AdminLayout, {
      global: {
        stubs: {
          'router-view': { template: '<div class="router-view"></div>' },
          transition: { template: '<div class="transition"><slot /></div>' },
        },
      },
    })
    expect(wrapper.find('.admin-layout').exists()).toBe(true)
    expect(wrapper.find('.admin-sidebar').exists()).toBe(true)
    expect(wrapper.find('.admin-main').exists()).toBe(true)
  })

  it('displays the admin title in sidebar header', () => {
    const wrapper = mount(AdminLayout, {
      global: {
        stubs: {
          'router-view': { template: '<div class="router-view"></div>' },
          transition: { template: '<div class="transition"><slot /></div>' },
        },
      },
    })
    expect(wrapper.find('.logo').text()).toContain('管理后台')
  })

  it('renders all navigation items', () => {
    const wrapper = mount(AdminLayout, {
      global: {
        stubs: {
          'router-view': { template: '<div class="router-view"></div>' },
          transition: { template: '<div class="transition"><slot /></div>' },
        },
      },
    })
    const navItems = wrapper.findAll('.nav-item')
    expect(navItems.length).toBe(8) // 7 main items + 1 footer item
  })

  it('contains overview navigation item', () => {
    const wrapper = mount(AdminLayout, {
      global: {
        stubs: {
          'router-view': { template: '<div class="router-view"></div>' },
          transition: { template: '<div class="transition"><slot /></div>' },
        },
      },
    })
    const navItems = wrapper.findAll('.nav-item')
    const overviewItem = navItems.find((item) => item.text().includes('概览'))
    expect(overviewItem).toBeDefined()
  })

  it('contains users navigation item', () => {
    const wrapper = mount(AdminLayout, {
      global: {
        stubs: {
          'router-view': { template: '<div class="router-view"></div>' },
          transition: { template: '<div class="transition"><slot /></div>' },
        },
      },
    })
    const navItems = wrapper.findAll('.nav-item')
    const usersItem = navItems.find((item) => item.text().includes('用户'))
    expect(usersItem).toBeDefined()
  })

  it('contains groups navigation item', () => {
    const wrapper = mount(AdminLayout, {
      global: {
        stubs: {
          'router-view': { template: '<div class="router-view"></div>' },
          transition: { template: '<div class="transition"><slot /></div>' },
        },
      },
    })
    const navItems = wrapper.findAll('.nav-item')
    const groupsItem = navItems.find((item) => item.text().includes('用户组'))
    expect(groupsItem).toBeDefined()
  })

  it('contains clients navigation item', () => {
    const wrapper = mount(AdminLayout, {
      global: {
        stubs: {
          'router-view': { template: '<div class="router-view"></div>' },
          transition: { template: '<div class="transition"><slot /></div>' },
        },
      },
    })
    const navItems = wrapper.findAll('.nav-item')
    const clientsItem = navItems.find((item) => item.text().includes('客户端'))
    expect(clientsItem).toBeDefined()
  })

  it('contains invitations navigation item', () => {
    const wrapper = mount(AdminLayout, {
      global: {
        stubs: {
          'router-view': { template: '<div class="router-view"></div>' },
          transition: { template: '<div class="transition"><slot /></div>' },
        },
      },
    })
    const navItems = wrapper.findAll('.nav-item')
    const invitationsItem = navItems.find((item) => item.text().includes('邀请'))
    expect(invitationsItem).toBeDefined()
  })

  it('contains proxy auth navigation item', () => {
    const wrapper = mount(AdminLayout, {
      global: {
        stubs: {
          'router-view': { template: '<div class="router-view"></div>' },
          transition: { template: '<div class="transition"><slot /></div>' },
        },
      },
    })
    const navItems = wrapper.findAll('.nav-item')
    const proxyAuthItem = navItems.find((item) => item.text().includes('代理认证'))
    expect(proxyAuthItem).toBeDefined()
  })

  it('contains settings navigation item', () => {
    const wrapper = mount(AdminLayout, {
      global: {
        stubs: {
          'router-view': { template: '<div class="router-view"></div>' },
          transition: { template: '<div class="transition"><slot /></div>' },
        },
      },
    })
    const navItems = wrapper.findAll('.nav-item')
    const settingsItem = navItems.find((item) => item.text().includes('系统设置'))
    expect(settingsItem).toBeDefined()
  })

  it('contains footer with back to user center link', () => {
    const wrapper = mount(AdminLayout, {
      global: {
        stubs: {
          'router-view': { template: '<div class="router-view"></div>' },
          transition: { template: '<div class="transition"><slot /></div>' },
        },
      },
    })
    const footerItems = wrapper.find('.sidebar-footer').findAll('.nav-item')
    const backLink = footerItems.find((item) => item.text().includes('返回用户中心'))
    expect(backLink).toBeDefined()
  })

  it('navigates to correct path when clicking nav item', async () => {
    const wrapper = mount(AdminLayout, {
      global: {
        stubs: {
          'router-view': { template: '<div class="router-view"></div>' },
          transition: { template: '<div class="transition"><slot /></div>' },
        },
      },
    })
    const navItems = wrapper.findAll('.nav-item')
    const usersItem = navItems[1] // Users nav item
    await usersItem.trigger('click')
    expect(mockPush).toHaveBeenCalledWith('/admin/users')
  })

  it('marks active nav item correctly for overview', async () => {
    const wrapper = mount(AdminLayout, {
      global: {
        stubs: {
          'router-view': { template: '<div class="router-view"></div>' },
          transition: { template: '<div class="transition"><slot /></div>' },
        },
      },
    })
    const navItems = wrapper.findAll('.nav-item')
    const overviewItem = navItems[0]
    expect(overviewItem.classes()).toContain('active')
  })

  it('shows content wrapper', () => {
    const wrapper = mount(AdminLayout, {
      global: {
        stubs: {
          'router-view': { template: '<div class="router-view"></div>' },
          transition: { template: '<div class="transition"><slot /></div>' },
        },
      },
    })
    expect(wrapper.find('.content-wrapper').exists()).toBe(true)
  })

  it('renders router-view', () => {
    const wrapper = mount(AdminLayout, {
      global: {
        stubs: {
          'router-view': { template: '<div class="router-view"></div>' },
          transition: true,
        },
      },
    })
    expect(wrapper.find('.router-view').exists()).toBe(true)
  })
})
