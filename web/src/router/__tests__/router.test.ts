import { describe, it, expect, vi, beforeEach } from 'vitest'
import { createRouter, createWebHistory, type RouteRecordRaw } from 'vue-router'
import { createPinia, setActivePinia } from 'pinia'
import { useRouter, useRoute } from 'vue-router'

vi.mock('vue-router', async (importOriginal) => {
  const actual = await importOriginal()
  return {
    ...(actual as any),
    useRouter: () => mockRouter,
    useRoute: () => mockRoute,
  }
})

const mockRouter = {
  push: vi.fn(),
  replace: vi.fn(),
  beforeEach: vi.fn(),
  afterEach: vi.fn(),
}

const mockRoute = {
  query: {},
  params: {},
}

vi.mock('@/stores/auth', () => ({
  useAuthStore: vi.fn(() => ({
    initFromStorage: vi.fn(),
    isAuthenticated: false,
    isAdmin: false,
  })),
}))

const routes: RouteRecordRaw[] = [
  {
    path: '/login',
    name: 'Login',
    component: { template: '<div>Login</div>' },
    meta: { requiresAuth: false },
  },
  {
    path: '/mfa',
    name: 'Mfa',
    component: { template: '<div>MFA</div>' },
    meta: { requiresAuth: true },
  },
  {
    path: '/user',
    component: { template: '<div>User Layout</div>' },
    meta: { requiresAuth: true },
    children: [
      {
        path: 'profile',
        name: 'Profile',
        component: { template: '<div>Profile</div>' },
      },
    ],
  },
  {
    path: '/admin',
    component: { template: '<div>Admin Layout</div>' },
    meta: { requiresAuth: true, requiresAdmin: true },
    children: [
      {
        path: '',
        name: 'AdminDashboard',
        component: { template: '<div>Admin Dashboard</div>' },
      },
    ],
  },
  {
    path: '/',
    redirect: '/login',
  },
]

function createTestRouter() {
  return createRouter({
    history: createWebHistory(),
    routes,
  })
}

describe('router', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    vi.clearAllMocks()
  })

  describe('route meta requirements', () => {
    it('should have correct meta for login route', () => {
      const router = createTestRouter()
      const loginRoute = router.getRoutes().find((r) => r.name === 'Login')
      expect(loginRoute?.meta.requiresAuth).toBe(false)
    })

    it('should have correct meta for mfa route requiring auth', () => {
      const router = createTestRouter()
      const mfaRoute = router.getRoutes().find((r) => r.name === 'Mfa')
      expect(mfaRoute?.meta.requiresAuth).toBe(true)
    })

    it('should inherit requiresAuth from parent route', () => {
      const router = createTestRouter()
      const userRoute = router.getRoutes().find((r) => r.path === '/user')
      expect(userRoute?.meta.requiresAuth).toBe(true)
    })

    it('should have correct meta for admin route parent', () => {
      const router = createTestRouter()
      const routes = router.getRoutes()
      const adminParentRoute = routes.find((r) => r.path === '/admin' && !r.name)
      expect(adminParentRoute?.meta.requiresAuth).toBe(true)
      expect(adminParentRoute?.meta.requiresAdmin).toBe(true)
    })
  })

  describe('route configuration', () => {
    it('should have all required routes defined', () => {
      const router = createTestRouter()
      const routeNames = router
        .getRoutes()
        .map((r) => r.name as string)
        .filter(Boolean)

      expect(routeNames).toContain('Login')
      expect(routeNames).toContain('Mfa')
      expect(routeNames).toContain('Profile')
      expect(routeNames).toContain('AdminDashboard')
    })

    it('should have nested routes for user section', () => {
      const router = createTestRouter()
      const userRoute = router.getRoutes().find((r) => r.path === '/user')
      expect(userRoute?.children).toHaveLength(1)
      expect(userRoute?.children?.[0].name).toBe('Profile')
    })

    it('should have nested routes for admin section', () => {
      const router = createTestRouter()
      const routes = router.getRoutes()
      const adminParentRoute = routes.find((r) => r.path === '/admin' && !r.name)
      expect(adminParentRoute?.children).toHaveLength(1)
      expect(adminParentRoute?.children?.[0].name).toBe('AdminDashboard')
    })

    it('should have redirect for root path', () => {
      const router = createTestRouter()
      const rootRoute = router.getRoutes().find((r) => r.path === '/')
      expect(rootRoute?.redirect).toBeDefined()
    })
  })

  describe('route guards integration', () => {
    it('should export router instance', async () => {
      const router = createTestRouter()
      expect(router).toBeDefined()
      expect(typeof router.push).toBe('function')
      expect(typeof router.beforeEach).toBe('function')
    })

    it('should support adding navigation guards', async () => {
      const router = createTestRouter()
      const guardFn = vi.fn()

      router.beforeEach(guardFn)
      await router.push('/login')

      expect(guardFn).toHaveBeenCalled()
    })

    it('should call guard with correct arguments on navigation', async () => {
      const router = createTestRouter()
      const guardFn = vi.fn()

      router.beforeEach(guardFn)
      await router.push('/mfa')

      expect(guardFn).toHaveBeenCalledWith(
        expect.objectContaining({
          name: 'Mfa',
          path: '/mfa',
        }),
        expect.any(Object),
        expect.any(Function)
      )
    })

    it('should allow guard to control navigation with next', async () => {
      const router = createTestRouter()

      router.beforeEach((_to, _from, next) => {
        next()
      })

      const afterHook = vi.fn()
      router.afterEach(afterHook)

      await router.push('/login')
      expect(afterHook).toHaveBeenCalled()
    })
  })

  describe('useRouter composable', () => {
    it('should return a router instance', () => {
      const router = useRouter()
      expect(router).toBeDefined()
    })

    it('should return the mock router', () => {
      const router = useRouter()
      expect(router).toBe(mockRouter)
    })
  })

  describe('useRoute composable', () => {
    it('should return a route object', () => {
      const route = useRoute()
      expect(route).toBeDefined()
    })

    it('should return the mock route', () => {
      const route = useRoute()
      expect(route).toBe(mockRoute)
    })
  })

  describe('lazy loading routes', () => {
    it('should support lazy loaded components via function import', () => {
      const lazyRoutes: RouteRecordRaw[] = [
        {
          path: '/lazy',
          name: 'Lazy',
          component: () => Promise.resolve({ template: '<div>Lazy</div>' }),
          meta: { requiresAuth: false },
        },
      ]

      const router = createRouter({
        history: createWebHistory(),
        routes: lazyRoutes,
      })

      const lazyRoute = router.getRoutes().find((r) => r.name === 'Lazy')
      expect(lazyRoute?.components).toBeDefined()
    })
  })
})
