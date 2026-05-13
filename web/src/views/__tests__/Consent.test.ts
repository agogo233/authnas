import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'
import Consent from '../Consent.vue'

vi.mock('@/api/oidc', () => ({
  oidcApi: {
    getInteraction: vi.fn(),
    confirmInteraction: vi.fn(),
    cancelInteraction: vi.fn(),
  },
}))

vi.mock('vue-router', () => ({
  useRouter: () => ({
    push: vi.fn(),
  }),
  useRoute: () => ({
    params: { uid: 'test-uid' },
    query: {},
  }),
}))

vi.mock('naive-ui', () => ({
  NCard: {
    name: 'NCard',
    props: ['title', 'style'],
    template: '<div class="n-card"><slot /></div>',
  },
  NButton: {
    name: 'NButton',
    props: ['type', 'loading', 'disabled'],
    emits: ['click'],
    template: '<button :disabled="disabled || loading" @click="$emit(\'click\')"><slot /></button>',
  },
  NSpace: {
    name: 'NSpace',
    props: ['justify', 'align'],
    template: '<div class="n-space"><slot /></div>',
  },
  NAlert: {
    name: 'NAlert',
    props: ['type'],
    template: '<div class="n-alert"><slot /></div>',
  },
  NResult: {
    name: 'NResult',
    props: ['status', 'title', 'description'],
    template: '<div class="n-result"><slot name="footer" /></div>',
  },
  NSpin: {
    name: 'NSpin',
    template: '<div class="n-spin" />',
  },
  useMessage: () => ({
    success: vi.fn(),
    error: vi.fn(),
    warning: vi.fn(),
  }),
}))

describe('Consent.vue', () => {
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

  const mockInteractionData = {
    uid: 'test-uid',
    client: {
      client_id: 'test-client',
      name: 'Test App',
      logo_uri: 'https://example.com/logo.png',
    },
    scopes: ['openid', 'profile', 'email'],
    claims: { sub: 'user123' },
    request_url: 'https://example.com/auth',
  }

  describe('rendering', () => {
    it('renders consent container correctly', () => {
      const wrapper = mount(Consent, mountOptions)
      expect(wrapper.find('.auth-container').exists()).toBe(true)
    })
  })

  describe('loading interaction data', () => {
    it('calls getInteraction API on mount', async () => {
      const { oidcApi } = await import('@/api/oidc')
      vi.mocked(oidcApi.getInteraction).mockResolvedValue({
        data: { success: true, data: mockInteractionData },
      } as any)
      
      mount(Consent, mountOptions)
      await flushPromises()
      
      expect(oidcApi.getInteraction).toHaveBeenCalledWith('test-uid')
    })

    it('displays client name when loaded', async () => {
      const { oidcApi } = await import('@/api/oidc')
      vi.mocked(oidcApi.getInteraction).mockResolvedValue({
        data: { success: true, data: mockInteractionData },
      } as any)
      
      const wrapper = mount(Consent, mountOptions)
      await flushPromises()
      
      expect(wrapper.find('.client-details h3').text()).toBe('Test App')
    })

    it('displays scope list', async () => {
      const { oidcApi } = await import('@/api/oidc')
      vi.mocked(oidcApi.getInteraction).mockResolvedValue({
        data: { success: true, data: mockInteractionData },
      } as any)
      
      const wrapper = mount(Consent, mountOptions)
      await flushPromises()
      
      const scopesSection = wrapper.find('.scopes')
      expect(scopesSection.exists()).toBe(true)
      expect(scopesSection.findAll('li').length).toBe(3)
    })
  })

  describe('scope descriptions', () => {
    it('shows correct description for openid scope', async () => {
      const { oidcApi } = await import('@/api/oidc')
      vi.mocked(oidcApi.getInteraction).mockResolvedValue({
        data: { success: true, data: mockInteractionData },
      } as any)
      
      const wrapper = mount(Consent, mountOptions)
      await flushPromises()
      
      const openidScope = wrapper.findAll('li').find(li => li.text().includes('openid'))
      expect(openidScope?.text()).toContain('验证您的身份')
    })

    it('shows correct description for profile scope', async () => {
      const { oidcApi } = await import('@/api/oidc')
      vi.mocked(oidcApi.getInteraction).mockResolvedValue({
        data: { success: true, data: mockInteractionData },
      } as any)
      
      const wrapper = mount(Consent, mountOptions)
      await flushPromises()
      
      const profileScope = wrapper.findAll('li').find(li => li.text().includes('profile'))
      expect(profileScope?.text()).toContain('访问您的基本 profile 信息')
    })

    it('shows correct description for email scope', async () => {
      const { oidcApi } = await import('@/api/oidc')
      vi.mocked(oidcApi.getInteraction).mockResolvedValue({
        data: { success: true, data: mockInteractionData },
      } as any)
      
      const wrapper = mount(Consent, mountOptions)
      await flushPromises()
      
      const emailScope = wrapper.findAll('li').find(li => li.text().includes('email'))
      expect(emailScope?.text()).toContain('访问您的邮箱地址')
    })
  })

  describe('authorization buttons', () => {
    it('has authorize button', async () => {
      const { oidcApi } = await import('@/api/oidc')
      vi.mocked(oidcApi.getInteraction).mockResolvedValue({
        data: { success: true, data: mockInteractionData },
      } as any)
      
      const wrapper = mount(Consent, mountOptions)
      await flushPromises()
      
      const authorizeBtn = wrapper.findAll('button').find(btn => btn.text().includes('授权'))
      expect(authorizeBtn).toBeDefined()
    })

    it('has decline button', async () => {
      const { oidcApi } = await import('@/api/oidc')
      vi.mocked(oidcApi.getInteraction).mockResolvedValue({
        data: { success: true, data: mockInteractionData },
      } as any)
      
      const wrapper = mount(Consent, mountOptions)
      await flushPromises()
      
      const declineBtn = wrapper.findAll('button').find(btn => btn.text().includes('拒绝'))
      expect(declineBtn).toBeDefined()
    })

    it('calls confirmInteraction when authorize is clicked', async () => {
      const { oidcApi } = await import('@/api/oidc')
      vi.mocked(oidcApi.getInteraction).mockResolvedValue({
        data: { success: true, data: mockInteractionData },
      } as any)
      vi.mocked(oidcApi.confirmInteraction).mockResolvedValue({
        data: { success: true, data: { redirect_to: 'https://example.com/callback' } },
      } as any)
      
      const wrapper = mount(Consent, mountOptions)
      await flushPromises()
      
      const authorizeBtn = wrapper.findAll('button').find(btn => btn.text().includes('授权'))
      await authorizeBtn?.trigger('click')
      
      await flushPromises()
      
      expect(oidcApi.confirmInteraction).toHaveBeenCalledWith('test-uid')
    })

    it('calls cancelInteraction when decline is clicked', async () => {
      const { oidcApi } = await import('@/api/oidc')
      vi.mocked(oidcApi.getInteraction).mockResolvedValue({
        data: { success: true, data: mockInteractionData },
      } as any)
      vi.mocked(oidcApi.cancelInteraction).mockResolvedValue({
        data: { success: true, data: { redirect_to: 'https://example.com/cancel' } },
      } as any)
      
      const wrapper = mount(Consent, mountOptions)
      await flushPromises()
      
      const declineBtn = wrapper.findAll('button').find(btn => btn.text().includes('拒绝'))
      await declineBtn?.trigger('click')
      
      await flushPromises()
      
      expect(oidcApi.cancelInteraction).toHaveBeenCalledWith('test-uid')
    })
  })

  describe('error handling', () => {
    it('shows error result when API call fails', async () => {
      const { oidcApi } = await import('@/api/oidc')
      vi.mocked(oidcApi.getInteraction).mockRejectedValue({
        response: { data: { message: 'Failed to load' } },
      })
      
      const wrapper = mount(Consent, mountOptions)
      await flushPromises()
      
      const result = wrapper.find('.n-result')
      expect(result.exists()).toBe(true)
    })
  })
})
