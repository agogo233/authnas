import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import Error from '../Error.vue'

const mockRouter = {
  push: vi.fn(),
}

const mockRoute = {
  query: {},
  params: {},
}

vi.mock('vue-router', () => ({
  useRouter: () => mockRouter,
  useRoute: () => mockRoute,
}))

vi.mock('naive-ui', () => ({
  NResult: {
    name: 'NResult',
    props: ['status', 'title', 'description'],
    template: '<div class="n-result"><slot name="footer" /></div>',
  },
  NButton: {
    name: 'NButton',
    props: ['type'],
    emits: ['click'],
    template: '<button @click="$emit(\'click\')"><slot /></button>',
  },
  NSpace: {
    name: 'NSpace',
    props: ['justify'],
    template: '<div class="n-space"><slot /></div>',
  },
}))

describe('Error.vue', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    mockRoute.query = {}
    mockRoute.params = {}
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
    it('renders error container', () => {
      const wrapper = mount(Error, mountOptions)
      expect(wrapper.find('.error-container').exists()).toBe(true)
    })

    it('renders NResult component', () => {
      const wrapper = mount(Error, mountOptions)
      expect(wrapper.find('.n-result').exists()).toBe(true)
    })

    it('renders footer with buttons', () => {
      const wrapper = mount(Error, mountOptions)
      const footer = wrapper.find('.n-space')
      expect(footer.exists()).toBe(true)
      expect(footer.findAll('button').length).toBe(2)
    })

    it('has go home button', () => {
      const wrapper = mount(Error, mountOptions)
      const buttons = wrapper.findAll('button')
      expect(buttons[0].text()).toBe('返回首页')
    })

    it('has go login button', () => {
      const wrapper = mount(Error, mountOptions)
      const buttons = wrapper.findAll('button')
      expect(buttons[1].text()).toBe('前往登录')
    })
  })

  describe('status handling', () => {
    it('defaults to 500 status when no query param', async () => {
      const wrapper = mount(Error, mountOptions)
      await flushPromises()
      const vm = wrapper.vm as any
      expect(vm.status).toBe(500)
    })

    it('parses status from query param', async () => {
      mockRoute.query = { status: '404' }
      const wrapper = mount(Error, mountOptions)
      await flushPromises()
      const vm = wrapper.vm as any
      expect(vm.status).toBe(404)
    })

    it('handles invalid status gracefully', async () => {
      mockRoute.query = { status: 'invalid' }
      const wrapper = mount(Error, mountOptions)
      await flushPromises()
      const vm = wrapper.vm as any
      expect(vm.status).toBe(NaN)
    })
  })

  describe('title generation', () => {
    it('shows "请求错误" for 400 status', async () => {
      mockRoute.query = { status: '400' }
      const wrapper = mount(Error, mountOptions)
      await flushPromises()
      const vm = wrapper.vm as any
      expect(vm.title).toBe('请求错误')
    })

    it('shows "未授权" for 401 status', async () => {
      mockRoute.query = { status: '401' }
      const wrapper = mount(Error, mountOptions)
      await flushPromises()
      const vm = wrapper.vm as any
      expect(vm.title).toBe('未授权')
    })

    it('shows "禁止访问" for 403 status', async () => {
      mockRoute.query = { status: '403' }
      const wrapper = mount(Error, mountOptions)
      await flushPromises()
      const vm = wrapper.vm as any
      expect(vm.title).toBe('禁止访问')
    })

    it('shows "页面不存在" for 404 status', async () => {
      mockRoute.query = { status: '404' }
      const wrapper = mount(Error, mountOptions)
      await flushPromises()
      const vm = wrapper.vm as any
      expect(vm.title).toBe('页面不存在')
    })

    it('shows "服务器内部错误" for 500 status', async () => {
      mockRoute.query = { status: '500' }
      const wrapper = mount(Error, mountOptions)
      await flushPromises()
      const vm = wrapper.vm as any
      expect(vm.title).toBe('服务器内部错误')
    })

    it('shows "错误" for unknown status', async () => {
      mockRoute.query = { status: '999' }
      const wrapper = mount(Error, mountOptions)
      await flushPromises()
      const vm = wrapper.vm as any
      expect(vm.title).toBe('错误')
    })
  })

  describe('description handling', () => {
    it('shows custom message from query param', async () => {
      mockRoute.query = { message: 'Custom error message' }
      const wrapper = mount(Error, mountOptions)
      await flushPromises()
      const vm = wrapper.vm as any
      expect(vm.description).toBe('Custom error message')
    })

    it('shows default message when no query param', async () => {
      const wrapper = mount(Error, mountOptions)
      await flushPromises()
      const vm = wrapper.vm as any
      expect(vm.description).toBe('发生了意外的错误')
    })

    it('handles URL encoded messages', async () => {
      mockRoute.query = { message: 'Error%20with%20spaces' }
      const wrapper = mount(Error, mountOptions)
      await flushPromises()
      const vm = wrapper.vm as any
      expect(vm.description).toBe('Error%20with%20spaces')
    })
  })

  describe('statusString computation', () => {
    it('returns "error" for 400', async () => {
      mockRoute.query = { status: '400' }
      const wrapper = mount(Error, mountOptions)
      await flushPromises()
      const vm = wrapper.vm as any
      expect(vm.statusString).toBe('error')
    })

    it('returns "error" for 401', async () => {
      mockRoute.query = { status: '401' }
      const wrapper = mount(Error, mountOptions)
      await flushPromises()
      const vm = wrapper.vm as any
      expect(vm.statusString).toBe('error')
    })

    it('returns "500" for 500', async () => {
      mockRoute.query = { status: '500' }
      const wrapper = mount(Error, mountOptions)
      await flushPromises()
      const vm = wrapper.vm as any
      expect(vm.statusString).toBe('500')
    })

    it('returns "404" for 404', async () => {
      mockRoute.query = { status: '404' }
      const wrapper = mount(Error, mountOptions)
      await flushPromises()
      const vm = wrapper.vm as any
      expect(vm.statusString).toBe('404')
    })
  })

  describe('navigation', () => {
    it('navigates to home when goHome is called', async () => {
      const wrapper = mount(Error, mountOptions)
      await flushPromises()
      const vm = wrapper.vm as any

      vm.goHome()

      expect(mockRouter.push).toHaveBeenCalledWith('/')
    })

    it('navigates to login when goLogin is called', async () => {
      const wrapper = mount(Error, mountOptions)
      await flushPromises()
      const vm = wrapper.vm as any

      vm.goLogin()

      expect(mockRouter.push).toHaveBeenCalledWith('/login')
    })

    it('home button triggers goHome', async () => {
      const wrapper = mount(Error, mountOptions)
      await flushPromises()
      const buttons = wrapper.findAll('button')

      await buttons[0].trigger('click')

      expect(mockRouter.push).toHaveBeenCalledWith('/')
    })

    it('login button triggers goLogin', async () => {
      const wrapper = mount(Error, mountOptions)
      await flushPromises()
      const buttons = wrapper.findAll('button')

      await buttons[1].trigger('click')

      expect(mockRouter.push).toHaveBeenCalledWith('/login')
    })
  })
})
