<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { showSuccessToast, showToast } from 'vant'
import type { ActiveOrder, OrderNotifyEvent } from '@/types'
import { isSameOrder } from '@/utils/orderEvent'
import { saveCompletedTrip, buildCompletedTripFromNotify } from '@/utils/trip'
import { syncPassengerLocationToOrderEnd } from '@/utils/tripLocation'
import {
  applyDriverAcceptedToOrder,
  applyTripStartedToOrder,
  ensurePassengerOrderWs,
  getActivePassengerOrder,
  orderStatusLabel,
  setActivePassengerOrder,
  stopPassengerOrderWs,
  subscribePassengerOrderChange,
  subscribePassengerOrderWs,
  syncActivePassengerOrderFromServer,
} from '@/utils/passengerTrip'

const route = useRoute()
const router = useRouter()

const activeTrip = ref<ActiveOrder | null>(null)
let unsubWs: (() => void) | null = null
let unsubChange: (() => void) | null = null
let pollTimer: ReturnType<typeof setInterval> | null = null

const hideTabbar = computed(() =>
  ['/passenger/waiting', '/passenger/wallet', '/passenger/coupons', '/passenger/realname', '/passenger/issue', '/passenger/rate', '/passenger/completed'].some(
    (p) => route.path.startsWith(p),
  ) || route.path.startsWith('/passenger/wallet/'),
)

const isChatTab = computed(() => route.path.startsWith('/passenger/ai-chat'))

const showActiveBar = computed(
  () => !!activeTrip.value && !route.path.startsWith('/passenger/waiting') && !isChatTab.value,
)

const activeTab = computed(() => {
  if (route.path.startsWith('/passenger/mine')) return 2
  if (isChatTab.value) return 1
  return 0
})

function syncLocalTrip() {
  activeTrip.value = getActivePassengerOrder()
}

async function refreshTrip() {
  activeTrip.value = await syncActivePassengerOrderFromServer()
}

function goActiveTrip() {
  router.push('/passenger/waiting')
}

function onTabChange(index: number) {
  if (index === 0) router.replace('/passenger')
  else if (index === 1) router.replace('/passenger/ai-chat')
  else router.replace('/passenger/mine')
}

function handlePassengerWs(evt: OrderNotifyEvent) {
  const trip = activeTrip.value || getActivePassengerOrder()
  if (trip && !isSameOrder(evt, trip.orderNo)) return

  switch (evt.event) {
    case 'driver_accepted':
      if (trip) {
        const updated = applyDriverAcceptedToOrder(evt, trip)
        setActivePassengerOrder(updated)
        activeTrip.value = updated
      }
      if (!route.path.startsWith('/passenger/waiting')) {
        showSuccessToast('司机已接单，点击查看行程')
      }
      break
    case 'order_cancelled':
      setActivePassengerOrder(null)
      activeTrip.value = null
      showToast(evt.msg || '订单已取消')
      if (!route.path.startsWith('/passenger')) {
        router.replace('/passenger')
      }
      break
    case 'order_completed':
      if (route.path.startsWith('/passenger/waiting')) {
        // 等待页由 WaitingView 弹窗并跳转
        break
      }
      void syncPassengerLocationToOrderEnd({
        endAddress: trip?.end,
        endLng: trip?.endLng,
        endLat: trip?.endLat,
      })
      saveCompletedTrip(buildCompletedTripFromNotify(evt, trip || getActivePassengerOrder()))
      setActivePassengerOrder(null)
      activeTrip.value = null
      showSuccessToast(evt.msg || '行程已完成')
      router.replace('/passenger/completed')
      break
    case 'trip_started':
      if (trip) {
        const updated = applyTripStartedToOrder(trip)
        setActivePassengerOrder(updated)
        activeTrip.value = updated
      }
      if (!route.path.startsWith('/passenger/waiting')) {
        showSuccessToast(evt.msg || '行程已开始，请系好安全带')
      }
      break
    case 'order_pushed_drivers': {
      const msg =
        evt.msg ||
        (evt.pushed_driver_count != null
          ? `已向附近 ${evt.push_radius_km ?? 20} 公里内 ${evt.pushed_driver_count} 位司机推送`
          : '正在扩大搜索范围')
      if (route.path.startsWith('/passenger/waiting')) {
        break
      }
      showToast(msg)
      break
    }
  }
}

onMounted(() => {
  syncLocalTrip()
  void refreshTrip()
  ensurePassengerOrderWs()
  unsubWs = subscribePassengerOrderWs(handlePassengerWs)
  unsubChange = subscribePassengerOrderChange(syncLocalTrip)
  pollTimer = setInterval(() => void refreshTrip(), 20000)
})

onUnmounted(() => {
  unsubWs?.()
  unsubChange?.()
  if (pollTimer) clearInterval(pollTimer)
  stopPassengerOrderWs()
})

watch(() => route.fullPath, () => {
  syncLocalTrip()
  if (!route.path.startsWith('/passenger/waiting')) {
    void refreshTrip()
  }
})
</script>

<template>
  <div class="layout" :class="{ 'has-active-bar': showActiveBar, 'is-chat-tab': isChatTab }">
    <div class="layout-body">
      <router-view />
    </div>

    <div v-if="showActiveBar" class="active-trip-bar" @click="goActiveTrip">
      <div class="active-trip-bar-main">
        <van-icon name="logistics" class="active-trip-bar-icon" />
        <div>
          <p class="active-trip-bar-title">{{ orderStatusLabel(activeTrip?.status) }}</p>
          <p class="active-trip-bar-addr">
            {{ activeTrip?.start }} → {{ activeTrip?.end }}
          </p>
        </div>
      </div>
      <van-button size="small" type="primary" round>查看</van-button>
    </div>

    <van-tabbar v-if="!hideTabbar" :model-value="activeTab" @update:model-value="onTabChange">
      <van-tabbar-item icon="home-o">叫车</van-tabbar-item>
      <van-tabbar-item icon="chat-o">助手</van-tabbar-item>
      <van-tabbar-item icon="user-o">我的</van-tabbar-item>
    </van-tabbar>
  </div>
</template>

<style scoped>
.layout {
  min-height: 100vh;
  padding-bottom: 50px;
  display: flex;
  flex-direction: column;
}
.layout.is-chat-tab {
  height: 100vh;
  min-height: 100vh;
}
.layout.is-chat-tab .layout-body {
  overflow: hidden;
}
.layout-body {
  flex: 1;
  min-height: 0;
  display: flex;
  flex-direction: column;
}
.layout.has-active-bar {
  padding-bottom: 118px;
}
.active-trip-bar {
  position: fixed;
  left: 12px;
  right: 12px;
  bottom: 58px;
  z-index: 1000;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  padding: 12px 14px;
  border-radius: 12px;
  background: linear-gradient(135deg, #e6f4ff 0%, #f0f5ff 100%);
  border: 1px solid #91caff;
  box-shadow: 0 4px 12px rgba(22, 119, 255, 0.15);
}
.active-trip-bar-main {
  display: flex;
  align-items: flex-start;
  gap: 10px;
  flex: 1;
  min-width: 0;
}
.active-trip-bar-icon {
  font-size: 22px;
  color: #1677ff;
  margin-top: 2px;
}
.active-trip-bar-title {
  font-size: 14px;
  font-weight: 600;
  color: #1677ff;
}
.active-trip-bar-addr {
  font-size: 12px;
  color: #595959;
  margin-top: 4px;
  word-break: break-all;
}
</style>
