import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'
import AdminSettings from '../admin/Settings.vue'

vi.mock('@/api/client', () => ({
  default: {
    get: vi.fn(),
    post: vi.fn(),
  },
}))

import apiClient from '@/api/client'

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
    template: '<div class="n-card"><slot /></div>',
  },
  NTabs: {
    name: 'NTabs',
    props: ['type'],
    template: '<div class="n-tabs"><slot /></div>',
  },
  NTabPane: {
    name: 'NTabPane',
    props: ['name', 'tab'],
    template: '<div class="n-tab-pane" :data-tab="name"><slot /></div>',
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
  NSwitch: {
    name: 'NSwitch',
    props: ['modelValue', 'disabled'],
    emits: ['update:modelValue'],
    template:
      '<div class="n-switch" @click="!disabled && $emit(\'update:modelValue\', !modelValue)"></div>',
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
  NAlert: {
    name: 'NAlert',
    props: ['type'],
    template: '<div class="n-alert"><slot /></div>',
  },
  NInputNumber: {
    name: 'NInputNumber',
    props: ['modelValue', 'min', 'max', 'disabled'],
    emits: ['update:modelValue'],
    template:
      '<input type="number" :value="modelValue" :min="min" :max="max" :disabled="disabled" @input="$emit(\'update:modelValue\', $event.target.value)" />',
  },
  NSelect: {
    name: 'NSelect',
    props: ['modelValue', 'options', 'disabled'],
    emits: ['update:modelValue'],
    template:
      '<div class="n-select" @click="!disabled && $emit(\'update:modelValue\', modelValue)"><slot /></div>',
  },
  NPopconfirm: {
    name: 'NPopconfirm',
    props: [],
    emits: ['positive-click'],
    template: '<div class="n-popconfirm"><slot /><slot name="trigger" /></div>',
  },
  useMessage: () => mockMessage,
}))

describe('AdminSettings.vue', () => {
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

  const createMockApiResponse = (data: any) => ({
    data: { success: true, data },
  })

  const setupDefaultMocks = () => {
    ;(apiClient.get as any).mockResolvedValue(createMockApiResponse({}))
    ;(apiClient.post as any).mockResolvedValue(createMockApiResponse({}))
  }

  describe('rendering', () => {
    it('renders settings page correctly', async () => {
      setupDefaultMocks()
      const wrapper = mount(AdminSettings, mountOptions)
      await flushPromises()

      expect(wrapper.find('.page-container').exists()).toBe(true)
      expect(wrapper.find('.n-card').exists()).toBe(true)
    })

    it('renders tabs component', async () => {
      setupDefaultMocks()
      const wrapper = mount(AdminSettings, mountOptions)
      await flushPromises()

      expect(wrapper.find('.n-tabs').exists()).toBe(true)
    })

    it('renders all tab panes', async () => {
      setupDefaultMocks()
      const wrapper = mount(AdminSettings, mountOptions)
      await flushPromises()

      const tabPanes = wrapper.findAll('.n-tab-pane')
      expect(tabPanes.length).toBe(5)
    })
  })

  describe('component state', () => {
    it('initializes with default general form values', async () => {
      setupDefaultMocks()
      const wrapper = mount(AdminSettings, mountOptions)
      const vm = wrapper.vm as any

      expect(vm.generalForm.appName).toBe('AuthNas')
      expect(vm.generalForm.appUrl).toBe('http://localhost:8080')
    })

    it('initializes with default security form values', async () => {
      setupDefaultMocks()
      const wrapper = mount(AdminSettings, mountOptions)
      const vm = wrapper.vm as any

      expect(vm.securityForm.emailVerificationRequired).toBe(false)
      expect(vm.securityForm.signupRequiresApproval).toBe(false)
      expect(vm.securityForm.mfaRequired).toBe(false)
      expect(vm.securityForm.passwordMinLength).toBe(8)
      expect(vm.securityForm.passwordStrength).toBe(3)
    })

    it('initializes with default email form values', async () => {
      setupDefaultMocks()
      const wrapper = mount(AdminSettings, mountOptions)
      const vm = wrapper.vm as any

      expect(vm.emailForm.enabled).toBe(false)
      expect(vm.emailForm.smtpPort).toBe(587)
      expect(vm.emailForm.fromName).toBe('AuthNas')
    })

    it('initializes with default session form values', async () => {
      setupDefaultMocks()
      const wrapper = mount(AdminSettings, mountOptions)
      const vm = wrapper.vm as any

      expect(vm.sessionForm.accessTokenExpiry).toBe(15)
      expect(vm.sessionForm.refreshTokenExpiry).toBe(7)
      expect(vm.sessionForm.maxSessionsPerUser).toBe(5)
    })

    it('initializes with default rate limit form values', async () => {
      setupDefaultMocks()
      const wrapper = mount(AdminSettings, mountOptions)
      const vm = wrapper.vm as any

      expect(vm.rateLimitForm.enabled).toBe(true)
      expect(vm.rateLimitForm.loginLimit).toBe(5)
      expect(vm.rateLimitForm.registerLimit).toBe(3)
      expect(vm.rateLimitForm.apiLimit).toBe(60)
    })
  })

  describe('fetchConfig', () => {
    it('fetches all settings in parallel on mount', async () => {
      setupDefaultMocks()
      mount(AdminSettings, mountOptions)
      await flushPromises()

      expect(apiClient.get).toHaveBeenCalledWith('/api/admin/settings/general')
      expect(apiClient.get).toHaveBeenCalledWith('/api/admin/settings/security')
      expect(apiClient.get).toHaveBeenCalledWith('/api/admin/settings/email')
      expect(apiClient.get).toHaveBeenCalledWith('/api/admin/settings/session')
      expect(apiClient.get).toHaveBeenCalledWith('/api/admin/settings/ratelimit')
    })

    it('updates general form when config is loaded', async () => {
      ;(apiClient.get as any).mockImplementation((url: string) => {
        if (url === '/api/admin/settings/general') {
          return Promise.resolve(
            createMockApiResponse({
              app_name: 'Custom App',
              app_url: 'https://custom.com',
            })
          )
        }
        return Promise.resolve(createMockApiResponse({}))
      })

      const wrapper = mount(AdminSettings, mountOptions)
      await flushPromises()

      const vm = wrapper.vm as any
      expect(vm.generalForm.appName).toBe('Custom App')
      expect(vm.generalForm.appUrl).toBe('https://custom.com')
    })

    it('updates security form when config is loaded', async () => {
      ;(apiClient.get as any).mockImplementation((url: string) => {
        if (url === '/api/admin/settings/security') {
          return Promise.resolve(
            createMockApiResponse({
              email_verification_required: true,
              signup_requires_approval: true,
              mfa_required: true,
            })
          )
        }
        return Promise.resolve(createMockApiResponse({}))
      })

      const wrapper = mount(AdminSettings, mountOptions)
      await flushPromises()

      const vm = wrapper.vm as any
      expect(vm.securityForm.emailVerificationRequired).toBe(true)
      expect(vm.securityForm.signupRequiresApproval).toBe(true)
      expect(vm.securityForm.mfaRequired).toBe(true)
    })

    it('updates email form when config is loaded', async () => {
      ;(apiClient.get as any).mockImplementation((url: string) => {
        if (url === '/api/admin/settings/email') {
          return Promise.resolve(
            createMockApiResponse({
              enabled: true,
              smtp_host: 'smtp.example.com',
              smtp_port: 465,
              smtp_user: 'user@example.com',
              from_email: 'noreply@example.com',
              from_name: 'Custom Sender',
            })
          )
        }
        return Promise.resolve(createMockApiResponse({}))
      })

      const wrapper = mount(AdminSettings, mountOptions)
      await flushPromises()

      const vm = wrapper.vm as any
      expect(vm.emailForm.enabled).toBe(true)
      expect(vm.emailForm.smtpHost).toBe('smtp.example.com')
      expect(vm.emailForm.smtpPort).toBe(465)
      expect(vm.emailForm.smtpUser).toBe('user@example.com')
      expect(vm.emailForm.fromAddress).toBe('noreply@example.com')
      expect(vm.emailForm.fromName).toBe('Custom Sender')
    })

    it('handles fetch error gracefully', async () => {
      const consoleSpy = vi.spyOn(console, 'error').mockImplementation(() => {})
      ;(apiClient.get as any).mockRejectedValue(new Error('Network error'))

      const wrapper = mount(AdminSettings, mountOptions)
      await flushPromises()

      expect(consoleSpy).toHaveBeenCalledWith('获取配置失败:', expect.any(Error))
      expect(mockMessage.error).toHaveBeenCalledWith('获取配置失败')

      consoleSpy.mockRestore()
    })
  })

  describe('save settings', () => {
    it('saves general settings successfully', async () => {
      setupDefaultMocks()
      const wrapper = mount(AdminSettings, mountOptions)
      await flushPromises()

      const vm = wrapper.vm as any
      vm.generalFormRef = { validate: vi.fn().mockResolvedValue(true) }
      vm.generalForm.appName = 'Updated App'
      await vm.saveGeneralSettings()
      await flushPromises()

      expect(apiClient.post).toHaveBeenCalledWith('/api/admin/settings/general', {
        app_name: 'Updated App',
        app_url: 'http://localhost:8080',
      })
      expect(mockMessage.success).toHaveBeenCalledWith('常规设置已保存')
    })

    it('saves security settings successfully', async () => {
      setupDefaultMocks()
      const wrapper = mount(AdminSettings, mountOptions)
      await flushPromises()

      const vm = wrapper.vm as any
      vm.securityForm.emailVerificationRequired = true
      await vm.saveSecuritySettings()
      await flushPromises()

      expect(apiClient.post).toHaveBeenCalledWith(
        '/api/admin/settings/security',
        expect.objectContaining({
          email_verification_required: true,
        })
      )
      expect(mockMessage.success).toHaveBeenCalledWith('安全设置已保存')
    })

    it('saves email settings successfully when enabled', async () => {
      setupDefaultMocks()
      const wrapper = mount(AdminSettings, mountOptions)
      await flushPromises()

      const vm = wrapper.vm as any
      vm.emailFormRef = { validate: vi.fn().mockResolvedValue(true) }
      vm.emailForm.enabled = true
      vm.emailForm.smtpHost = 'smtp.test.com'
      vm.emailForm.smtpPort = 587
      vm.emailForm.smtpUser = 'test@test.com'
      vm.emailForm.fromAddress = 'noreply@test.com'
      await vm.saveEmailSettings()
      await flushPromises()

      expect(apiClient.post).toHaveBeenCalledWith(
        '/api/admin/settings/email',
        expect.objectContaining({
          enabled: true,
          smtp_host: 'smtp.test.com',
        })
      )
      expect(mockMessage.success).toHaveBeenCalledWith('邮件设置已保存')
    })

    it('shows warning when saving email settings without enabling', async () => {
      setupDefaultMocks()
      const wrapper = mount(AdminSettings, mountOptions)
      await flushPromises()

      const vm = wrapper.vm as any
      vm.emailForm.enabled = false
      await vm.saveEmailSettings()
      await flushPromises()

      expect(mockMessage.warning).toHaveBeenCalledWith('请先启用邮件功能')
      expect(apiClient.post).not.toHaveBeenCalled()
    })

    it('saves session settings successfully', async () => {
      setupDefaultMocks()
      const wrapper = mount(AdminSettings, mountOptions)
      await flushPromises()

      const vm = wrapper.vm as any
      vm.sessionForm.accessTokenExpiry = 30
      await vm.saveSessionSettings()
      await flushPromises()

      expect(apiClient.post).toHaveBeenCalledWith('/api/admin/settings/session', {
        access_token_expiry: 30,
        refresh_token_expiry: 7,
        max_sessions_per_user: 5,
      })
      expect(mockMessage.success).toHaveBeenCalledWith('会话设置已保存')
    })

    it('saves rate limit settings successfully', async () => {
      setupDefaultMocks()
      const wrapper = mount(AdminSettings, mountOptions)
      await flushPromises()

      const vm = wrapper.vm as any
      vm.rateLimitForm.loginLimit = 10
      await vm.saveRateLimitSettings()
      await flushPromises()

      expect(apiClient.post).toHaveBeenCalledWith('/api/admin/settings/ratelimit', {
        enabled: true,
        login_limit: 10,
        register_limit: 3,
        api_limit: 60,
      })
      expect(mockMessage.success).toHaveBeenCalledWith('限流设置已保存')
    })

    it('handles save error gracefully', async () => {
      ;(apiClient.get as any).mockResolvedValue(createMockApiResponse({}))
      ;(apiClient.post as any).mockRejectedValue({
        response: { data: { message: 'Server error' } },
      })

      const wrapper = mount(AdminSettings, mountOptions)
      await flushPromises()

      const vm = wrapper.vm as any
      vm.generalFormRef = { validate: vi.fn().mockResolvedValue(true) }
      await vm.saveGeneralSettings()
      await flushPromises()

      expect(mockMessage.error).toHaveBeenCalledWith('Server error')
    })
  })

  describe('sendTestEmail', () => {
    it('shows warning when email is not enabled', async () => {
      setupDefaultMocks()
      const wrapper = mount(AdminSettings, mountOptions)
      await flushPromises()

      const vm = wrapper.vm as any
      vm.emailForm.enabled = false
      await vm.sendTestEmail()
      await flushPromises()

      expect(mockMessage.warning).toHaveBeenCalledWith('请先启用邮件功能')
    })

    it('sends test email when email is enabled', async () => {
      setupDefaultMocks()
      const wrapper = mount(AdminSettings, mountOptions)
      await flushPromises()

      const vm = wrapper.vm as any
      vm.emailForm.enabled = true
      await vm.sendTestEmail()
      await flushPromises()

      expect(apiClient.post).toHaveBeenCalledWith(
        '/api/admin/settings/email/test',
        expect.objectContaining({
          smtpHost: '',
          smtpPort: 587,
        })
      )
      expect(mockMessage.success).toHaveBeenCalledWith('测试邮件已发送，请检查收件箱')
    })

    it('handles test email error gracefully', async () => {
      ;(apiClient.get as any).mockResolvedValue(createMockApiResponse({}))
      ;(apiClient.post as any).mockRejectedValue({
        response: { data: { message: 'SMTP not configured' } },
      })

      const wrapper = mount(AdminSettings, mountOptions)
      await flushPromises()

      const vm = wrapper.vm as any
      vm.emailForm.enabled = true
      await vm.sendTestEmail()
      await flushPromises()

      expect(mockMessage.error).toHaveBeenCalledWith('SMTP not configured')
    })
  })

  describe('form validation', () => {
    it('has general form ref defined', async () => {
      setupDefaultMocks()
      const wrapper = mount(AdminSettings, mountOptions)
      await flushPromises()

      const vm = wrapper.vm as any
      expect(vm.generalFormRef).toBeDefined()
    })

    it('has email form ref defined', async () => {
      setupDefaultMocks()
      const wrapper = mount(AdminSettings, mountOptions)
      await flushPromises()

      const vm = wrapper.vm as any
      expect(vm.emailFormRef).toBeDefined()
    })
  })

  describe('securityOptions', () => {
    it('contains correct options', async () => {
      setupDefaultMocks()
      const wrapper = mount(AdminSettings, mountOptions)
      await flushPromises()

      const vm = wrapper.vm as any
      expect(vm.securityOptions).toHaveLength(4)
      expect(vm.securityOptions[0]).toEqual({ label: '低 (1)', value: 1 })
      expect(vm.securityOptions[1]).toEqual({ label: '中 (2)', value: 2 })
      expect(vm.securityOptions[2]).toEqual({ label: '高 (3)', value: 3 })
      expect(vm.securityOptions[3]).toEqual({ label: '非常高 (4)', value: 4 })
    })
  })

  describe('data echo on load', () => {
    it('correctly maps general settings response to form', async () => {
      ;(apiClient.get as any).mockImplementation((url: string) => {
        if (url === '/api/admin/settings/general') {
          return Promise.resolve(
            createMockApiResponse({
              app_name: 'TestApp',
              app_url: 'https://test.example.com',
            })
          )
        }
        return Promise.resolve(createMockApiResponse({}))
      })

      const wrapper = mount(AdminSettings, mountOptions)
      await flushPromises()

      const vm = wrapper.vm as any
      expect(vm.generalForm.appName).toBe('TestApp')
      expect(vm.generalForm.appUrl).toBe('https://test.example.com')
    })

    it('correctly maps session settings response to form', async () => {
      ;(apiClient.get as any).mockImplementation((url: string) => {
        if (url === '/api/admin/settings/session') {
          return Promise.resolve(
            createMockApiResponse({
              access_token_expiry: 30,
              refresh_token_expiry: 14,
              max_sessions_per_user: 10,
            })
          )
        }
        return Promise.resolve(createMockApiResponse({}))
      })

      const wrapper = mount(AdminSettings, mountOptions)
      await flushPromises()

      const vm = wrapper.vm as any
      expect(vm.sessionForm.accessTokenExpiry).toBe(30)
      expect(vm.sessionForm.refreshTokenExpiry).toBe(14)
      expect(vm.sessionForm.maxSessionsPerUser).toBe(10)
    })

    it('correctly maps rate limit settings response to form', async () => {
      ;(apiClient.get as any).mockImplementation((url: string) => {
        if (url === '/api/admin/settings/ratelimit') {
          return Promise.resolve(
            createMockApiResponse({
              login_limit: 10,
              register_limit: 5,
              api_limit: 100,
            })
          )
        }
        return Promise.resolve(createMockApiResponse({}))
      })

      const wrapper = mount(AdminSettings, mountOptions)
      await flushPromises()

      const vm = wrapper.vm as any
      expect(vm.rateLimitForm.loginLimit).toBe(10)
      expect(vm.rateLimitForm.registerLimit).toBe(5)
      expect(vm.rateLimitForm.apiLimit).toBe(100)
    })
  })

  describe('password field handling', () => {
    it('clears password field on each fetch', async () => {
      ;(apiClient.get as any).mockImplementation((url: string) => {
        if (url === '/api/admin/settings/email') {
          return Promise.resolve(
            createMockApiResponse({
              enabled: true,
              smtp_host: 'smtp.test.com',
            })
          )
        }
        return Promise.resolve(createMockApiResponse({}))
      })

      const wrapper = mount(AdminSettings, mountOptions)
      await flushPromises()

      const vm = wrapper.vm as any
      expect(vm.emailForm.smtpPassword).toBe('')
    })

    it('allows setting smtp password', async () => {
      setupDefaultMocks()
      const wrapper = mount(AdminSettings, mountOptions)
      await flushPromises()

      const vm = wrapper.vm as any
      vm.emailForm.smtpPassword = 'new-password'

      expect(vm.emailForm.smtpPassword).toBe('new-password')
    })
  })

  describe('error handling', () => {
    it('displays error when settings fetch fails', async () => {
      ;(apiClient.get as any).mockRejectedValue(new Error('Network failure'))
      const consoleSpy = vi.spyOn(console, 'error').mockImplementation(() => {})

      mount(AdminSettings, mountOptions)
      await flushPromises()

      expect(mockMessage.error).toHaveBeenCalledWith('获取配置失败')

      consoleSpy.mockRestore()
    })

    it('handles partial fetch failure gracefully', async () => {
      ;(apiClient.get as any).mockImplementation((url: string) => {
        if (url === '/api/admin/settings/general') {
          return Promise.reject(new Error('Failed'))
        }
        return Promise.resolve(createMockApiResponse({}))
      })
      const consoleSpy = vi.spyOn(console, 'error').mockImplementation(() => {})

      const wrapper = mount(AdminSettings, mountOptions)
      await flushPromises()

      expect(mockMessage.error).toHaveBeenCalledWith('获取配置失败')
      expect(wrapper.vm).toBeTruthy()

      consoleSpy.mockRestore()
    })
  })
})
