import { createRouter, createWebHistory } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import type { UserRole } from '@/types'

function getStoredAuth() {
  const role = localStorage.getItem('role') as UserRole | null
  if (role !== 'passenger' && role !== 'driver') {
    return { role: null as UserRole | null, token: '' }
  }
  const token = localStorage.getItem(role === 'driver' ? 'driver_token' : 'passenger_token') || ''
  return { role, token }
}

const router = createRouter({
  history: createWebHistory(),
  routes: [
    {
      path: '/',
      name: 'home',
      component: () => import('@/views/HomeView.vue'),
      meta: { title: '高德网约车' },
    },
    {
      path: '/passenger/login',
      name: 'passenger-login',
      component: () => import('@/views/passenger/LoginView.vue'),
      meta: { title: '乘客登录' },
    },
    {
      path: '/passenger/register',
      name: 'passenger-register',
      component: () => import('@/views/passenger/RegisterView.vue'),
      meta: { title: '乘客注册' },
    },
    {
      path: '/passenger',
      component: () => import('@/views/passenger/LayoutView.vue'),
      meta: { requiresAuth: true, role: 'passenger' },
      children: [
        {
          path: '',
          name: 'passenger-book',
          component: () => import('@/views/passenger/BookView.vue'),
          meta: { title: '叫车' },
        },
        {
          path: 'waiting',
          name: 'passenger-waiting',
          component: () => import('@/views/passenger/WaitingView.vue'),
          meta: { title: '等待接单' },
        },
        {
          path: 'rate',
          name: 'passenger-rate',
          component: () => import('@/views/passenger/RateView.vue'),
          meta: { title: '评价司机' },
        },
        {
          path: 'completed',
          name: 'passenger-completed',
          component: () => import('@/views/passenger/CompletedView.vue'),
          meta: { title: '行程结束' },
        },
        {
          path: 'wallet',
          name: 'passenger-wallet',
          component: () => import('@/views/passenger/WalletView.vue'),
          meta: { title: '我的钱包' },
        },
        {
          path: 'wallet/logs',
          name: 'passenger-wallet-logs',
          component: () => import('@/views/passenger/WalletLogsView.vue'),
          meta: { title: '流水明细' },
        },
        {
          path: 'order/list',
          name: 'passenger-order-list',
          component: () => import('@/views/passenger/OrderListView.vue'),
          meta: { title: '历史订单' },
        },
        {
          path: 'order/rating',
          name: 'passenger-order-rating',
          component: () => import('@/views/passenger/OrderRatingView.vue'),
          meta: { title: '评价详情' },
        },
        {
          path: 'coupons',
          name: 'passenger-coupons',
          component: () => import('@/views/passenger/CouponsView.vue'),
          meta: { title: '优惠券' },
        },
        {
          path: 'realname',
          name: 'passenger-realname',
          component: () => import('@/views/passenger/RealNameView.vue'),
          meta: { title: '实名认证' },
        },
        {
          path: 'issue',
          name: 'passenger-issue',
          component: () => import('@/views/company/IssueCouponsView.vue'),
          meta: { title: '发放优惠券' },
        },
        {
          path: 'mine',
          name: 'passenger-mine',
          component: () => import('@/views/passenger/MineView.vue'),
          meta: { title: '我的' },
        },
        {
          path: 'ai-chat',
          name: 'passenger-ai-chat',
          component: () => import('@/views/passenger/AiChatView.vue'),
          meta: { title: 'AI 助手' },
        },
      ],
    },
    {
      path: '/driver/login',
      name: 'driver-login',
      component: () => import('@/views/driver/LoginView.vue'),
      meta: { title: '司机登录' },
    },
    {
      path: '/driver/register',
      name: 'driver-register',
      component: () => import('@/views/driver/RegisterView.vue'),
      meta: { title: '司机注册' },
    },
    {
      path: '/driver',
      component: () => import('@/views/driver/LayoutView.vue'),
      meta: { requiresAuth: true, role: 'driver' },
      children: [
        {
          path: '',
          name: 'driver-orders',
          component: () => import('@/views/driver/OrdersView.vue'),
          meta: { title: '抢单大厅' },
        },
        {
          path: 'active',
          name: 'driver-active',
          component: () => import('@/views/driver/ActiveOrderView.vue'),
          meta: { title: '进行中订单' },
        },
        {
          path: 'verify',
          name: 'driver-verify',
          component: () => import('@/views/driver/VerifyView.vue'),
          meta: { title: '资质认证' },
        },
        {
          path: 'wallet',
          name: 'driver-wallet',
          component: () => import('@/views/driver/WalletView.vue'),
          meta: { title: '我的钱包' },
        },
        {
          path: 'wallet/logs',
          name: 'driver-wallet-logs',
          component: () => import('@/views/driver/WalletLogsView.vue'),
          meta: { title: '流水明细' },
        },
        {
          path: 'order/list',
          name: 'driver-order-list',
          component: () => import('@/views/driver/OrderListView.vue'),
          meta: { title: '历史订单' },
        },
        {
          path: 'order/rating',
          name: 'driver-order-rating',
          component: () => import('@/views/driver/OrderRatingView.vue'),
          meta: { title: '评价详情' },
        },
        {
          path: 'mine',
          name: 'driver-mine',
          component: () => import('@/views/driver/MineView.vue'),
          meta: { title: '我的' },
        },
        {
          path: 'ai-chat',
          name: 'driver-ai-chat',
          component: () => import('@/views/driver/AiChatView.vue'),
          meta: { title: 'AI 助手' },
        },
      ],
    },
  ],
})

router.beforeEach((to) => {
  document.title = (to.meta.title as string) || '高德网约车'
  const requiresAuth = to.matched.some((r) => r.meta.requiresAuth)
  if (!requiresAuth) return true

  const requiredRole = to.matched.find((r) => r.meta.role)?.meta.role as UserRole | undefined
  if (!requiredRole) return true

  const { role, token } = getStoredAuth()
  if (!token || role !== requiredRole) {
    return requiredRole === 'driver' ? '/driver/login' : '/passenger/login'
  }

  const auth = useAuthStore()
  if (auth.role !== role || auth.token !== token) {
    const phone = localStorage.getItem('phone') || ''
    if (role === 'driver') auth.setDriverAuth(token, phone)
    else auth.setPassengerAuth(token, phone)
  }
  return true
})

export default router
