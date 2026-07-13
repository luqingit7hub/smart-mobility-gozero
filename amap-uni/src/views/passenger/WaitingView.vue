<script setup lang="ts">
import { computed, onActivated, onMounted, onUnmounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { showSuccessToast, showToast } from 'vant'
import { showAppConfirm, showAppNotice } from '@/utils/dialog'
import { orderApi } from '@/api/user'
import type { ActiveOrder, OrderNotifyEvent } from '@/types'
import { formatPrice, shortenOrderNo } from '@/utils/order'
import { isSameOrder } from '@/utils/orderEvent'
import {
  applyDriverAcceptedToOrder,
  applyTripStartedToOrder,
  ensurePassengerOrderWs,
  getActivePassengerOrder,
  isFreshPassengerWaitingOrder,
  passengerPhaseFromStatus,
  resolveActivePassengerOrder,
  setActivePassengerOrder,
  subscribePassengerOrderChange,
  subscribePassengerOrderWs,
} from '@/utils/passengerTrip'
import { saveCompletedTrip, buildCompletedTripFromNotify, getCompletedTrip } from '@/utils/trip'
import { syncPassengerLocationToOrderEnd } from '@/utils/tripLocation'

const router = useRouter()

const orderInfo = ref<ActiveOrder | null>(null)
const driverInfo = ref<OrderNotifyEvent | null>(null)
const cancelling = ref(false)
const waitSeconds = ref(0)
const loading = ref(true)
const pushNotice = ref('')
const phase = ref<'waiting' | 'accepted' | 'inTrip' | 'cancelled' | 'completed'>('waiting')

const orderNo = computed(() => orderInfo.value?.orderNo?.trim() || '')
const priceParts = computed(() => formatPrice(orderInfo.value?.price))
const inTrip = computed(() => phase.value === 'inTrip')
const accepted = computed(() => phase.value === 'accepted')
const canCancel = computed(() => phase.value === 'waiting')
const pageTitle = computed(() => {
  if (inTrip.value) return '行程进行中'
  if (accepted.value) return '等待上车'
  return '等待接单'
})

let waitTimer: ReturnType<typeof setInterval> | null = null
let statusPollTimer: ReturnType<typeof setInterval> | null = null
let unsubWs: (() => void) | null = null
let unsubChange: (() => void) | null = null
let acceptedNotified = false
let tripStartedNotified = false
let cancelledNotified = false
let completedNotified = false

function goToCompletedPage(evt?: OrderNotifyEvent) {
  if (completedNotified) return
  completedNotified = true
  const trip = orderInfo.value
    ? {
        ...orderInfo.value,
        orderNo: orderNo.value || orderInfo.value.orderNo,
        driverName:
          driverInfo.value?.driver_name || evt?.driver_name || orderInfo.value.driverName,
        carNumber:
          driverInfo.value?.car_number || evt?.car_number || orderInfo.value.carNumber,
      }
    : null
  void syncPassengerLocationToOrderEnd({
    endAddress: trip?.end,
    endLng: trip?.endLng,
    endLat: trip?.endLat,
  })
  saveCompletedTrip(
    buildCompletedTripFromNotify(
      evt ?? { event: 'order_completed', order_no: orderNo.value, msg: '行程已完成' },
      trip,
    ),
  )
  setActivePassengerOrder(null)
  phase.value = 'completed'
  void showAppNotice('行程已完成', evt?.msg || '感谢使用，欢迎再次叫车', {
    confirmButtonText: '查看详情',
  }).finally(() => {
    router.replace('/passenger/completed')
  })
}

function notifyDriverAccepted(evt?: OrderNotifyEvent) {
  if (acceptedNotified) return
  acceptedNotified = true
  const name = evt?.driver_name || driverInfo.value?.driver_name || '司机师傅'
  const car = evt?.car_number || driverInfo.value?.car_number
  void showAppNotice(
    '司机已接单',
    car ? `${name}（${car}）正在赶来，请在上车点等候` : `${name} 正在赶来，请在上车点等候`,
    { confirmButtonText: '知道了' },
  )
}

function notifyTripStarted(evt?: OrderNotifyEvent) {
  if (tripStartedNotified) return
  tripStartedNotified = true
  void showAppNotice(
    '行程已开始',
    evt?.msg || '请系好安全带，祝您旅途愉快',
    { confirmButtonText: '知道了' },
  )
}

function syncFromActiveOrder() {
  const order = getActivePassengerOrder()
  if (!order?.orderNo || order.orderNo !== orderNo.value) return
  if (order.status === 5 && phase.value !== 'inTrip') {
    applyOrderToView(order)
    notifyTripStarted()
  } else if (order.status === 2 && phase.value === 'waiting') {
    applyOrderToView(order)
    notifyDriverAccepted()
  }
}

function applyOrderToView(order: ActiveOrder) {
  orderInfo.value = order
  setActivePassengerOrder(order)
  phase.value = passengerPhaseFromStatus(order.status)
  if ((order.status === 2 || order.status === 5) && (order.driverName || order.carNumber)) {
    driverInfo.value = {
      event: order.status === 5 ? 'trip_started' : 'driver_accepted',
      order_no: order.orderNo,
      driver_name: order.driverName,
      car_number: order.carNumber,
      car_type: order.carType,
      rating: order.driverRating,
    }
  }
}

function restoreWaitSeconds() {
  const startedAt = orderInfo.value?.waitStartedAt
  if (startedAt && startedAt > 0) {
    waitSeconds.value = Math.max(0, Math.floor((Date.now() - startedAt) / 1000))
    return
  }
  waitSeconds.value = 0
}

function handleWs(evt: OrderNotifyEvent) {
  if (!isSameOrder(evt, orderNo.value)) return

  switch (evt.event) {
    case 'driver_accepted':
      if (orderInfo.value) {
        applyOrderToView(applyDriverAcceptedToOrder(evt, orderInfo.value))
      }
      driverInfo.value = evt
      notifyDriverAccepted(evt)
      break
    case 'order_cancelled':
      if (cancelledNotified) break
      cancelledNotified = true
      phase.value = 'cancelled'
      setActivePassengerOrder(null)
      showToast(evt.msg || '订单已取消')
      setTimeout(() => router.replace('/passenger'), 800)
      break
    case 'order_completed':
      goToCompletedPage(evt)
      break
    case 'trip_started':
      if (orderInfo.value) {
        applyOrderToView(applyTripStartedToOrder(orderInfo.value))
      }
      if (evt.driver_name || evt.car_number) {
        driverInfo.value = evt
      }
      notifyTripStarted(evt)
      break
    case 'order_pushed_drivers': {
      const msg =
        evt.msg ||
        (evt.pushed_driver_count != null
          ? `已向附近 ${evt.push_radius_km ?? 20} 公里内 ${evt.pushed_driver_count} 位司机推送订单`
          : '系统已扩大搜索范围，继续为您寻找司机')
      pushNotice.value = msg
      showToast(msg)
      break
    }
  }
}

async function pollOrderStatus() {
  if (phase.value !== 'waiting' && phase.value !== 'accepted' && phase.value !== 'inTrip') return
  try {
    const res = await orderApi.ongoingOrder({ silent: true })
    const raw = res?.order as Record<string, unknown> | undefined
    const hasOrder = Boolean(res?.hasOrder ?? raw?.orderNo ?? raw?.order_no)
    if (!hasOrder) {
      if ((phase.value === 'accepted' || phase.value === 'inTrip') && orderNo.value) {
        goToCompletedPage()
        return
      }
      if (isFreshPassengerWaitingOrder(getActivePassengerOrder())) {
        return
      }
      if (phase.value === 'waiting' && orderNo.value && !cancelledNotified) {
        cancelledNotified = true
        phase.value = 'cancelled'
        setActivePassengerOrder(null)
        showToast('订单已取消（可能超时无人接单）')
        setTimeout(() => router.replace('/passenger'), 800)
      }
      return
    }
    if (!raw) return
    const order = await resolveActivePassengerOrder()
    if (!order?.orderNo) return
    if (order.status === 5 && phase.value !== 'inTrip') {
      applyOrderToView(order)
      notifyTripStarted()
    } else if (order.status === 2 && phase.value === 'waiting') {
      applyOrderToView(order)
      if (!driverInfo.value && (order.driverName || order.carNumber)) {
        driverInfo.value = {
          event: 'driver_accepted',
          order_no: order.orderNo,
          driver_name: order.driverName,
          car_number: order.carNumber,
          car_type: order.carType,
          rating: order.driverRating,
        }
      }
      notifyDriverAccepted()
    }
  } catch {
    // 轮询失败静默，避免打扰用户
  }
}

async function loadActiveOrder() {
  const localOrder = getActivePassengerOrder()
  if (localOrder?.orderNo) {
    if (!localOrder.waitStartedAt) localOrder.waitStartedAt = Date.now()
    applyOrderToView(localOrder)
    restoreWaitSeconds()
    loading.value = false
  } else {
    loading.value = true
  }

  try {
    const order = await resolveActivePassengerOrder()
    if (!order?.orderNo) {
      if (!localOrder?.orderNo) {
        const completed = getCompletedTrip()
        if (completed?.orderNo) {
          router.replace('/passenger/completed')
          return
        }
        router.replace('/passenger')
      }
      return
    }
    if (!order.waitStartedAt) order.waitStartedAt = Date.now()
    applyOrderToView(order)
    restoreWaitSeconds()
    if (order.status === 5) {
      notifyTripStarted()
    } else if (order.status === 2) {
      notifyDriverAccepted()
    }
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  ensurePassengerOrderWs()
  unsubWs = subscribePassengerOrderWs(handleWs)
  unsubChange = subscribePassengerOrderChange(syncFromActiveOrder)
  void loadActiveOrder()
  waitTimer = setInterval(() => waitSeconds.value++, 1000)
  statusPollTimer = setInterval(() => void pollOrderStatus(), 5000)
})

onActivated(() => {
  syncFromActiveOrder()
  void loadActiveOrder()
  void pollOrderStatus()
})

onUnmounted(() => {
  if (waitTimer) clearInterval(waitTimer)
  if (statusPollTimer) clearInterval(statusPollTimer)
  unsubWs?.()
  unsubChange?.()
})

function formatWaitTime(sec: number): string {
  const m = Math.floor(sec / 60)
  const s = sec % 60
  if (m === 0) return `${s} 秒`
  return `${m} 分 ${s.toString().padStart(2, '0')} 秒`
}

async function onCancel() {
  if (!orderNo.value) {
    await showAppNotice('暂无法取消', '未能获取订单号，请返回叫车页重新下单', {
      confirmButtonText: '返回叫车',
    })
    router.replace('/passenger')
    return
  }

  try {
    await showAppConfirm({
      title: '确认取消订单？',
      message: `取消后需重新叫车\n订单号：${orderNo.value}`,
      confirmButtonText: '确认取消',
      cancelButtonText: '继续等待',
    })
  } catch {
    return
  }

  cancelling.value = true
  try {
    await orderApi.cancelOrder(orderNo.value, '乘客主动取消')
    setActivePassengerOrder(null)
    showSuccessToast('订单已取消')
    router.replace('/passenger')
  } finally {
    cancelling.value = false
  }
}
</script>

<template>
  <div class="waiting-page">
    <van-nav-bar :title="pageTitle" left-arrow @click-left="router.push('/passenger')" />

    <van-loading v-if="loading" class="page-loading" />

    <div v-else class="waiting-body">
      <div class="search-hero">
        <div v-if="!accepted && !inTrip" class="radar">
          <span class="radar-ring ring-1" />
          <span class="radar-ring ring-2" />
          <span class="radar-ring ring-3" />
          <span class="radar-core">
            <van-icon name="logistics" size="28" color="#1677ff" />
          </span>
        </div>
        <div v-else-if="inTrip" class="trip-icon">
          <van-icon name="shield-o" size="48" color="#1677ff" />
        </div>
        <div v-else class="accepted-icon">
          <van-icon name="checked" size="48" color="#52c41a" />
        </div>
        <h2 class="hero-title">
          {{
            inTrip
              ? '行程已开始'
              : accepted
                ? '司机已接单，正在赶来'
                : '正在为您寻找司机'
          }}
        </h2>
        <p class="hero-sub">
          {{
            inTrip
              ? '请系好安全带，祝您旅途愉快'
              : accepted
                ? '请在上车点等候'
                : `已等待 ${formatWaitTime(waitSeconds)}`
          }}
        </p>
        <span class="status-badge" :class="{ ok: accepted, trip: inTrip }">
          {{ inTrip ? '行程中' : accepted ? '已接单' : '实时监听中' }}
        </span>
        <p v-if="!accepted && !inTrip" class="wait-tip">
          <span>订单 6 分钟后没司机接单，将主动通知附近司机进行接单</span>
          <span>十分钟后仍未接单，将自动取消订单</span>
        </p>
      </div>

      <div v-if="inTrip" class="safety-card">
        <van-icon name="info-o" class="safety-icon" />
        <div>
          <p class="safety-title">安全提示</p>
          <p class="safety-text">请全程系好安全带，注意保管随身物品，如有问题可联系司机或客服。</p>
        </div>
      </div>

      <div v-if="pushNotice && !accepted && !inTrip" class="push-notice card">
        <van-icon name="volume-o" class="push-notice-icon" />
        <p>{{ pushNotice }}</p>
      </div>

      <div v-if="driverInfo" class="driver-card">
        <div class="driver-avatar">
          <van-icon name="manager-o" size="32" color="#1677ff" />
        </div>
        <div class="driver-meta">
          <p class="driver-name">{{ driverInfo.driver_name || '司机师傅' }}</p>
          <p class="driver-car">
            {{ driverInfo.car_number || '车牌获取中' }}
            <span v-if="driverInfo.car_type"> · {{ driverInfo.car_type }}</span>
          </p>
          <p v-if="driverInfo.rating" class="driver-rating">评分 {{ driverInfo.rating.toFixed(1) }}</p>
        </div>
      </div>

      <div v-if="orderInfo" class="trip-card">
        <div class="trip-header">
          <span class="trip-label">订单号</span>
          <span v-if="orderNo" class="trip-order-no">{{ shortenOrderNo(orderNo) }}</span>
          <span v-else class="trip-order-no pending">获取中...</span>
        </div>

        <div class="route-section">
          <div class="route-line">
            <span class="route-dot start" />
            <span class="route-bar" />
            <span class="route-dot end" />
          </div>
          <div class="route-text">
            <div class="route-item">
              <span class="route-tag start-tag">起</span>
              <p>{{ orderInfo.start }}</p>
            </div>
            <div class="route-item">
              <span class="route-tag end-tag">终</span>
              <p>{{ orderInfo.end }}</p>
            </div>
          </div>
        </div>

        <div v-if="orderInfo.distance || orderInfo.duration" class="trip-chips">
          <span v-if="orderInfo.distance" class="chip">
            <van-icon name="location-o" /> {{ orderInfo.distance }} km
          </span>
          <span v-if="orderInfo.duration" class="chip">
            <van-icon name="clock-o" /> 约 {{ orderInfo.duration }} 分钟
          </span>
        </div>

        <div class="price-section">
          <div class="price-left">
            <p class="price-title">{{ inTrip ? '行程计费中' : '预计支付' }}</p>
            <p class="price-tip">
              {{ inTrip ? '到达目的地后司机将结束行程' : '司机接单后按实际行程结算' }}
            </p>
          </div>
          <div class="price-right">
            <span class="currency">{{ priceParts.symbol }}</span>
            <span class="amount-int">{{ priceParts.integer }}</span>
            <span class="amount-dec">.{{ priceParts.decimal }}</span>
          </div>
        </div>
      </div>

      <van-empty v-else description="暂无订单信息" />
    </div>

    <div v-if="canCancel && !loading" class="bottom-bar">
      <van-button round block plain type="danger" :loading="cancelling" @click="onCancel">
        取消订单
      </van-button>
    </div>
  </div>
</template>

<style scoped>
.page-loading {
  display: flex;
  justify-content: center;
  padding: 80px 0;
}
.waiting-page {
  min-height: 100vh;
  display: flex;
  flex-direction: column;
  background: #f5f6f8;
}
.waiting-body {
  flex: 1;
  padding: 8px 16px 100px;
}
.search-hero {
  text-align: center;
  padding: 24px 0 20px;
}
.push-notice {
  display: flex;
  align-items: flex-start;
  gap: 10px;
  margin-bottom: 12px;
  padding: 12px 14px;
  background: #fff7e6;
  border: 1px solid #ffe7ba;
}
.push-notice-icon {
  font-size: 18px;
  color: #fa8c16;
  margin-top: 2px;
  flex-shrink: 0;
}
.push-notice p {
  font-size: 13px;
  line-height: 1.5;
  color: #614700;
}
.radar {
  position: relative;
  width: 120px;
  height: 120px;
  margin: 0 auto 20px;
}
.radar-ring {
  position: absolute;
  inset: 0;
  border: 2px solid rgba(22, 119, 255, 0.25);
  border-radius: 50%;
  animation: radar-pulse 2.4s ease-out infinite;
}
.ring-2 {
  animation-delay: 0.8s;
}
.ring-3 {
  animation-delay: 1.6s;
}
.radar-core {
  position: absolute;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
  width: 56px;
  height: 56px;
  background: #fff;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  box-shadow: 0 4px 16px rgba(22, 119, 255, 0.15);
}
@keyframes radar-pulse {
  0% {
    transform: scale(0.5);
    opacity: 0.8;
  }
  100% {
    transform: scale(1.2);
    opacity: 0;
  }
}
.hero-title {
  font-size: 20px;
  font-weight: 600;
  color: #1a1a1a;
}
.hero-sub {
  margin-top: 6px;
  font-size: 13px;
  color: #8c8c8c;
}
.status-badge {
  display: inline-block;
  margin-top: 12px;
  padding: 4px 14px;
  font-size: 12px;
  color: #1677ff;
  background: #e6f4ff;
  border-radius: 20px;
}
.status-badge.ok {
  color: #52c41a;
  background: #f6ffed;
}
.status-badge.trip {
  color: #1677ff;
  background: #e6f4ff;
}
.wait-tip {
  margin: 14px 12px 0;
  font-size: 12px;
  line-height: 1.6;
  color: #8c8c8c;
  text-align: center;
}
.wait-tip span {
  display: block;
}
.trip-icon {
  width: 120px;
  height: 120px;
  margin: 0 auto 20px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: #fff;
  border-radius: 50%;
  box-shadow: 0 4px 16px rgba(22, 119, 255, 0.2);
}
.safety-card {
  display: flex;
  gap: 12px;
  align-items: flex-start;
  background: linear-gradient(135deg, #e6f4ff 0%, #f0f5ff 100%);
  border: 1px solid #91caff;
  border-radius: 14px;
  padding: 14px 16px;
  margin-bottom: 12px;
}
.safety-icon {
  font-size: 22px;
  color: #1677ff;
  margin-top: 2px;
  flex-shrink: 0;
}
.safety-title {
  font-size: 14px;
  font-weight: 600;
  color: #1677ff;
}
.safety-text {
  margin-top: 6px;
  font-size: 13px;
  line-height: 1.55;
  color: #595959;
}
.accepted-icon {
  width: 120px;
  height: 120px;
  margin: 0 auto 20px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: #fff;
  border-radius: 50%;
  box-shadow: 0 4px 16px rgba(82, 196, 26, 0.2);
}
.driver-card {
  display: flex;
  gap: 14px;
  align-items: center;
  background: #fff;
  border-radius: 16px;
  padding: 16px 18px;
  margin-bottom: 12px;
  box-shadow: 0 2px 12px rgba(0, 0, 0, 0.06);
}
.driver-avatar {
  width: 56px;
  height: 56px;
  border-radius: 50%;
  background: #e6f4ff;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}
.driver-name {
  font-size: 17px;
  font-weight: 600;
}
.driver-car {
  margin-top: 4px;
  font-size: 13px;
  color: #666;
}
.driver-rating {
  margin-top: 4px;
  font-size: 12px;
  color: #fa8c16;
}
.trip-card {
  background: #fff;
  border-radius: 16px;
  padding: 18px;
  box-shadow: 0 2px 12px rgba(0, 0, 0, 0.06);
}
.trip-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding-bottom: 14px;
  margin-bottom: 14px;
  border-bottom: 1px dashed #f0f0f0;
}
.trip-label {
  font-size: 13px;
  color: #8c8c8c;
}
.trip-order-no {
  font-size: 13px;
  font-family: ui-monospace, monospace;
  color: #333;
  background: #f5f5f5;
  padding: 2px 8px;
  border-radius: 4px;
}
.trip-order-no.pending {
  color: #1677ff;
  background: #e6f4ff;
}
.route-section {
  display: flex;
  gap: 14px;
}
.route-line {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding-top: 4px;
}
.route-dot {
  width: 10px;
  height: 10px;
  border-radius: 50%;
  flex-shrink: 0;
}
.route-dot.start {
  background: #52c41a;
  box-shadow: 0 0 0 3px rgba(82, 196, 26, 0.2);
}
.route-dot.end {
  background: #ff4d4f;
  box-shadow: 0 0 0 3px rgba(255, 77, 79, 0.2);
}
.route-bar {
  width: 2px;
  flex: 1;
  min-height: 32px;
  margin: 6px 0;
  background: linear-gradient(to bottom, #52c41a, #ff4d4f);
  opacity: 0.3;
}
.route-text {
  flex: 1;
  min-width: 0;
}
.route-item {
  display: flex;
  gap: 10px;
  align-items: flex-start;
}
.route-item + .route-item {
  margin-top: 18px;
}
.route-tag {
  flex-shrink: 0;
  width: 22px;
  height: 22px;
  line-height: 22px;
  text-align: center;
  font-size: 11px;
  font-weight: 600;
  border-radius: 6px;
  color: #fff;
}
.start-tag {
  background: #52c41a;
}
.end-tag {
  background: #ff4d4f;
}
.route-item p {
  font-size: 15px;
  line-height: 1.5;
  word-break: break-all;
}
.trip-chips {
  display: flex;
  gap: 10px;
  margin-top: 16px;
  flex-wrap: wrap;
}
.chip {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  padding: 6px 12px;
  font-size: 12px;
  color: #666;
  background: #f7f8fa;
  border-radius: 20px;
}
.price-section {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-top: 18px;
  padding: 16px;
  background: linear-gradient(135deg, #fff7e6 0%, #fff 100%);
  border-radius: 12px;
  border: 1px solid #ffe7ba;
}
.price-title {
  font-size: 14px;
  font-weight: 500;
}
.price-tip {
  margin-top: 4px;
  font-size: 11px;
  color: #999;
}
.price-right {
  display: flex;
  align-items: baseline;
  color: #ff6b00;
}
.currency {
  font-size: 16px;
  font-weight: 600;
}
.amount-int {
  font-size: 32px;
  font-weight: 700;
  line-height: 1;
}
.amount-dec {
  font-size: 16px;
  font-weight: 600;
}
.bottom-bar {
  position: fixed;
  left: 0;
  right: 0;
  bottom: 0;
  padding: 12px 16px calc(12px + env(safe-area-inset-bottom));
  background: #fff;
  box-shadow: 0 -2px 12px rgba(0, 0, 0, 0.06);
}
</style>
