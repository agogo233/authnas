import { createRouter, createWebHistory, type RouteRecordRaw } from 'vue-router'
import { useAuthStore } from '@/stores/auth'

const routes: RouteRecordRaw[] = [
  {
    path: '/login',
    name: 'Login',
    component: () => import('@/views/Login.vue'),
    meta: { requiresAuth: false },
  },
  {
    path: '/register',
    name: 'Register',
    component: () => import('@/views/Register.vue'),
    meta: { requiresAuth: false },
  },
  {
    path: '/reset-password',
    name: 'ResetPassword',
    component: () => import('@/views/ResetPassword.vue'),
    meta: { requiresAuth: false },
  },
  {
    path: '/verify-email',
    name: 'VerifyEmail',
    component: () => import('@/views/VerifyEmail.vue'),
    meta: { requiresAuth: false },
  },
  {
    path: '/consent/:uid',
    name: 'Consent',
    component: () => import('@/views/Consent.vue'),
    meta: { requiresAuth: false },
  },
  {
    path: '/mfa',
    name: 'Mfa',
    component: () => import('@/views/Mfa.vue'),
    meta: { requiresAuth: true },
  },
  {
    path: '/user',
    component: () => import('@/layouts/UserLayout.vue'),
    meta: { requiresAuth: true },
    children: [
      {
        path: 'profile',
        name: 'Profile',
        component: () => import('@/views/user/Profile.vue'),
      },
      {
        path: 'security',
        name: 'Security',
        component: () => import('@/views/user/Security.vue'),
      },
      {
        path: 'passkeys',
        name: 'Passkeys',
        component: () => import('@/views/user/Passkeys.vue'),
      },
    ],
  },
  {
    path: '/profile',
    redirect: '/user/profile',
  },
  {
    path: '/security',
    redirect: '/user/security',
  },
  {
    path: '/passkeys',
    redirect: '/user/passkeys',
  },
  {
    path: '/admin',
    component: () => import('@/layouts/AdminLayout.vue'),
    meta: { requiresAuth: true, requiresAdmin: true },
    children: [
      {
        path: '',
        name: 'AdminDashboard',
        component: () => import('@/views/admin/Dashboard.vue'),
      },
      {
        path: 'users',
        name: 'AdminUsers',
        component: () => import('@/views/admin/Users.vue'),
      },
      {
        path: 'groups',
        name: 'AdminGroups',
        component: () => import('@/views/admin/Groups.vue'),
      },
      {
        path: 'clients',
        name: 'AdminClients',
        component: () => import('@/views/admin/Clients.vue'),
      },
      {
        path: 'invitations',
        name: 'AdminInvitations',
        component: () => import('@/views/admin/Invitations.vue'),
      },
      {
        path: 'proxyauth',
        name: 'AdminProxyAuth',
        component: () => import('@/views/admin/ProxyAuth.vue'),
      },
      {
        path: 'settings',
        name: 'AdminSettings',
        component: () => import('@/views/admin/Settings.vue'),
      },
    ],
  },
  {
    path: '/',
    redirect: () => {
      const authStore = useAuthStore()
      if (authStore.isAuthenticated) {
        return '/user/profile'
      }
      return '/login'
    },
  },
  {
    path: '/error',
    name: 'Error',
    component: () => import('@/views/Error.vue'),
    meta: { requiresAuth: false },
  },
  {
    path: '/:pathMatch(.*)*',
    name: 'NotFound',
    redirect: '/error',
  },
]

const router = createRouter({
  history: createWebHistory(),
  routes,
})

router.beforeEach(async (to, _from, next) => {
  const authStore = useAuthStore()
  authStore.initFromStorage()

  if (to.meta.requiresAuth) {
    if (!authStore.isAuthenticated) {
      next({ name: 'Login', query: { redirect: to.fullPath } })
      return
    }
    if (to.meta.requiresAdmin && !authStore.isAdmin) {
      next({ name: 'Login' })
      return
    }
    next()
  } else {
    next()
  }
})

export default router
