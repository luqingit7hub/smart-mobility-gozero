<script setup lang="ts">
import { computed, onActivated, onMounted, onUnmounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { showSuccessToast, showToast } from 'vant'
import { driverApi, mapApi, GRAB_LIST_RADIUS_M } from '@/api/driver'
import type { GrabOrderItem, GrabOrderResp } from '@/types'
import { connectOrderWs } from '@/utils/orderWs'
import {
  getBd09LngLat,
  loadDriverLocation,
  saveDriverLocation,
} from '@/utils/geo'
import { showAppConfirm } from '@/utils/dialog'
import BookTripMap, { type TripMapPoint } from '@/components/BookTripMap.vue'
import {
  getActiveDriverTrip,
  isGrabSuccess,
  normalizeGrabOrder,
  resolveActiveDriverTrip,
  setActiveDriverTrip,
} from '@/utils/driverTrip'

import {
  DRIVER_LOCATION_ADDRESS_KEY,
  DRIVER_LOCATION_UPDATED,
} from '@/utils/tripLocation'

const LOCATION_KEY = DRIVER_LOCATION_ADDRESS_KEY

const router = useRouter()

const orders = ref<GrabOrderItem[]>([])
const loading = ref(false)
const grabbing = ref('')
const locationAddress = ref(localStorage.getItem(LOCATION_KEY) || '')
const driverLocation = ref<TripMapPoint | null>(null)
const locating = ref(false)
const showManualLocateSheet = ref(false)
const manualAddressInput = ref('')
const manualLocating = ref(false)
const wsConnected = ref(false)
const activeTrip = ref<GrabOrderItem | null>(null)

let pollTimer: ReturnType<typeof setInterval> | null = null
let closeWs: (() => void) | null = null

const locationSummary = computed(() =>
  locationAddress.value ? locationAddress.value : '未设置接单位置，请点击右侧定位',
)

function goActiveTrip() {
  if (activeTrip.value) router.push('/driver/active')
}

function applyDriverLocationCache(address: string, point: TripMapPoint) {
  locationAddress.value = address
  driverLocation.value = point
  saveDriverLocation({ lng: point.lng, lat: point.lat, address })
  localStorage.setItem(LOCATION_KEY, address)
  if (!pollTimer) {
    pollTimer = setInterval(loadOrders, 8000)
  }
}

async function applyDriverLocation(address: string) {
  const res = await mapApi.getCoordinates(address, 2)
  if (!Number.isFinite(res.lng) || !Number.isFinite(res.lat)) {
    throw new Error('invalid coordinates')
  }
  applyDriverLocationCache(res.address || address, { lng: res.lng, lat: res.lat })
  await loadOrders()
}

async function loadCachedDriverLocation() {
  let cached = loadDriverLocation()
  if (!cached && locationAddress.value) {
    try {
      const res = await mapApi.getCoordinates(locationAddress.value, 2)
      if (Number.isFinite(res.lng) && Number.isFinite(res.lat)) {
        cached = {
          lng: res.lng,
          lat: res.lat,
          address: res.address || locationAddress.value,
        }
        saveDriverLocation(cached)
      }
    } catch {
      return
    }
  }
  if (!cached) return

  let { lng, lat, address } = cached
  if (!address.trim()) {
    try {
      const res = await mapApi.reverseGeocode(lng, lat)
      address = res.address
      lng = res.lng || lng
      lat = res.lat || lat
      saveDriverLocation({ lng, lat, address })
    } catch {
      return
    }
  }

  applyDriverLocationCache(address, { lng, lat })
}

function onLocateMe() {
  if (locating.value) return
  showAppConfirm({
    title: '位置提示',
    message: '资质原因目前无法拿到您的准确位置',
    confirmButtonText: '继续自动定位',
    cancelButtonText: '选择手动定位',
  })
    .then(() => runAutoLocate())
    .catch(() => startManualLocate())
}

function startManualLocate() {
  manualAddressInput.value = locationAddress.value
  showManualLocateSheet.value = true
}

async function confirmManualLocate() {
  const address = manualAddressInput.value.trim()
  if (!address) {
    showToast('请输入接单位置')
    return
  }
  manualLocating.value = true
  try {
    await applyDriverLocation(address)
    showManualLocateSheet.value = false
    showSuccessToast('接单位置已更新')
  } catch {
    showToast('地址解析失败，请检查输入')
  } finally {
    manualLocating.value = false
  }
}

async function runAutoLocate() {
  if (locating.value) return
  locating.value = true
  try {
    const pos = await getBd09LngLat()
    if (!pos) {
      showToast('定位失败，请允许浏览器使用位置权限')
      return
    }
    const res = await mapApi.reverseGeocode(pos.lng, pos.lat)
    const point = { lng: res.lng || pos.lng, lat: res.lat || pos.lat }
    applyDriverLocationCache(res.address, point)
    await loadOrders()
    showSuccessToast('已定位到当前位置')
  } catch {
    showToast('定位失败，请稍后重试')
  } finally {
    locating.value = false
  }
}

async function loadOrders() {
  if (!locationAddress.value) {
    orders.value = []
    return
  }
  loading.value = true
  try {
    const res = await driverApi.grabList(GRAB_LIST_RADIUS_M, 20)
    orders.value = (res.orders || []).map((o) =>
      normalizeGrabOrder(o as unknown as Record<string, unknown>),
    )
  } catch {
    orders.value = []
  } finally {
    loading.value = false
  }
}

function refreshActiveTripLocal() {
  activeTrip.value = getActiveDriverTrip()
}

async function refreshActiveTrip() {
  activeTrip.value = await resolveActiveDriverTrip()
}

async function onGrab(order: GrabOrderItem) {
  const normalized = normalizeGrabOrder(order as unknown as Record<string, unknown>)
  if (!normalized.orderNo) {
    showToast('订单号无效')
    return
  }
  grabbing.value = normalized.orderNo
  try {
    const res = await driverApi.grabOrder(normalized.orderNo)
    if (isGrabSuccess(res as GrabOrderResp & Record<string, unknown>)) {
      const active = { ...normalized, status: 2 }
      setActiveDriverTrip(active)
      activeTrip.value = active
      showSuccessToast('抢单成功')
      await router.push('/driver/active')
    } else {
      showToast(res.msg || '抢单失败')
    }
  } finally {
    grabbing.value = ''
    loadOrders()
  }
}

function startDriverWs() {
  const token = localStorage.getItem('driver_token')
  if (!token) return
  wsConnected.value = false
  closeWs = connectOrderWs('driver', token, (evt) => {
    if (evt.event === 'new_order_nearby') {
      const tip = evt.msg || `附近有新订单：${evt.start_address || '请打开抢单列表'}`
      showToast(tip)
      loadOrders()
    }
  }, {
    onOpen: () => {
      wsConnected.value = true
    },
    onClose: () => {
      wsConnected.value = false
    },
  })
}

function onDriverLocationUpdated(evt: Event) {
  const detail = (evt as CustomEvent<{ lng: number; lat: number; address: string }>).detail
  if (!detail || !Number.isFinite(detail.lng) || !Number.isFinite(detail.lat)) return
  applyDriverLocationCache(detail.address, { lng: detail.lng, lat: detail.lat })
}

onMounted(async () => {
  refreshActiveTripLocal()
  void refreshActiveTrip()
  window.addEventListener('activeDriverOrderChange', refreshActiveTripLocal)
  await loadCachedDriverLocation()
  if (!locationAddress.value) {
    onLocateMe()
  } else {
    await loadOrders()
    if (!pollTimer) {
      pollTimer = setInterval(loadOrders, 8000)
    }
  }
  startDriverWs()
  window.addEventListener(DRIVER_LOCATION_UPDATED, onDriverLocationUpdated)
})

onActivated(async () => {
  await loadCachedDriverLocation()
})

onUnmounted(() => {
  window.removeEventListener(DRIVER_LOCATION_UPDATED, onDriverLocationUpdated)
  window.removeEventListener('activeDriverOrderChange', refreshActiveTripLocal)
  if (pollTimer) clearInterval(pollTimer)
  closeWs?.()
})
</script>

<template>
  <div class="page orders-page">
    <van-nav-bar title="抢单大厅">
      <template #right>
        <van-button size="small" type="primary" :loading="loading" @click="loadOrders">
          刷新
        </van-button>
      </template>
    </van-nav-bar>

    <div class="orders-body">
      <div class="status-row">
        <span class="online">● 已上线</span>
        <span class="ws">{{ wsConnected ? '推送已连接' : '推送连接中' }}</span>
      </div>

      <div v-if="activeTrip" class="active-trip card" @click="goActiveTrip">
        <div class="active-trip-main">
          <van-icon name="logistics" class="active-trip-icon" />
          <div>
            <p class="active-trip-title">进行中的行程</p>
            <p class="active-trip-addr">{{ activeTrip.startAddress }} → {{ activeTrip.endAddress }}</p>
          </div>
        </div>
        <van-icon name="arrow" />
      </div>

      <div class="location-hero primary-gradient">
        <div class="hero-icon-wrap">
          <van-icon name="location-o" />
        </div>
        <div class="hero-text">
          <p class="hero-title">当前接单位置</p>
          <p class="hero-hint">{{ locationSummary }}</p>
        </div>
        <button
          type="button"
          class="hero-locate-btn"
          :class="{ loading: locating }"
          :disabled="locating"
          @click="onLocateMe"
        >
          <van-icon :name="locating ? 'loading' : 'aim'" />
          <span>{{ locating ? '定位中' : '我的位置' }}</span>
        </button>
      </div>

      <BookTripMap
        class="orders-map-slot"
        :user-location="driverLocation"
        :locating="locating"
        @locate="onLocateMe"
      />

      <van-pull-refresh v-model="loading" @refresh="loadOrders">
        <van-empty
          v-if="!loading && !orders.length"
          :description="locationAddress ? '附近暂无待接订单' : '请先设置接单位置'"
        />
        <div v-else class="page-content">
          <div v-for="order in orders" :key="order.orderNo" class="card order-item">
            <div class="order-header">
              <span class="order-no">{{ order.orderNo }}</span>
              <div class="price-block">
                <span class="price">¥{{ order.price?.toFixed(2) }}</span>
                <span class="price-hint">订单金额</span>
              </div>
            </div>
            <p class="addr"><van-icon name="location-o" /> {{ order.startAddress }}</p>
            <p class="addr"><van-icon name="aim" /> {{ order.endAddress }}</p>
            <div class="meta">
              <span>{{ order.distance }} km</span>
              <span>{{ order.duration }} 分钟</span>
              <span>距您 {{ order.distanceToDriver?.toFixed(1) }} km</span>
            </div>
            <van-button
              type="primary"
              round
              block
              class="mt"
              :loading="grabbing === order.orderNo"
              @click="onGrab(order)"
            >
              立即抢单
            </van-button>
          </div>
        </div>
      </van-pull-refresh>
    </div>

    <van-popup
      v-model:show="showManualLocateSheet"
      position="bottom"
      round
      :style="{ maxHeight: '50vh' }"
    >
      <div class="manual-locate-sheet">
        <div class="sheet-header">
          <span>手动定位</span>
          <van-icon name="cross" @click="showManualLocateSheet = false" />
        </div>
        <p class="sheet-tip">输入接单位置，系统将转换为坐标并用于附近抢单</p>
        <van-field
          v-model="manualAddressInput"
          label="接单位置"
          placeholder="如：宿迁职业技术学院"
          clearable
        />
        <van-button
          type="primary"
          round
          block
          class="manual-locate-btn"
          :loading="manualLocating"
          @click="confirmManualLocate"
        >
          确认位置
        </van-button>
      </div>
    </van-popup>
  </div>
</template>

<style scoped>
.orders-page {
  background: var(--app-bg, #eef2f8);
}

.orders-body {
  padding-bottom: 12px;
}

.status-row {
  display: flex;
  justify-content: space-between;
  padding: 8px 16px 0;
  font-size: 12px;
}
.online {
  color: #1677ff;
}
.ws {
  color: #8c8c8c;
}
.active-trip {
  margin: 12px 16px 0;
  padding: 12px 14px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  background: linear-gradient(135deg, #e6f4ff 0%, #f0f5ff 100%);
  border: 1px solid #91caff;
}
.active-trip-main {
  display: flex;
  align-items: flex-start;
  gap: 10px;
  flex: 1;
  min-width: 0;
}
.active-trip-icon {
  font-size: 22px;
  color: #1677ff;
  margin-top: 2px;
}
.active-trip-title {
  font-size: 14px;
  font-weight: 600;
  color: #1677ff;
}
.active-trip-addr {
  font-size: 12px;
  color: #595959;
  margin-top: 4px;
  word-break: break-all;
}

.location-hero {
  display: flex;
  align-items: center;
  gap: 12px;
  margin: 12px 16px 0;
  padding: 16px 14px;
  color: #fff;
  border-radius: var(--app-radius-md, 14px);
  box-shadow: var(--app-shadow-sm, 0 2px 10px rgba(15, 23, 42, 0.06));
}

.orders-map-slot {
  margin: 12px 16px 0;
  border-radius: var(--app-radius-md, 14px);
  overflow: hidden;
}

.hero-icon-wrap {
  width: 44px;
  height: 44px;
  border-radius: 14px;
  background: rgba(255, 255, 255, 0.2);
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 24px;
  flex-shrink: 0;
}

.hero-text {
  flex: 1;
  min-width: 0;
}

.hero-title {
  font-size: 15px;
  font-weight: 600;
  letter-spacing: 0.3px;
}

.hero-hint {
  margin-top: 4px;
  font-size: 12px;
  opacity: 0.92;
  line-height: 1.45;
  word-break: break-all;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}

.hero-locate-btn {
  flex-shrink: 0;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 4px;
  height: 34px;
  padding: 0 12px;
  border: 1px solid rgba(255, 255, 255, 0.45);
  border-radius: 999px;
  background: rgba(255, 255, 255, 0.16);
  color: #fff;
  font-size: 12px;
  white-space: nowrap;
  cursor: pointer;
  backdrop-filter: blur(4px);
}

.hero-locate-btn:disabled {
  opacity: 0.75;
  cursor: wait;
}

.hero-locate-btn .van-icon {
  font-size: 15px;
}

.hero-locate-btn.loading .van-icon {
  animation: hero-locate-spin 0.8s linear infinite;
}

@keyframes hero-locate-spin {
  from {
    transform: rotate(0deg);
  }
  to {
    transform: rotate(360deg);
  }
}

.manual-locate-sheet {
  padding: 16px 16px 24px;
}

.sheet-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  font-size: 16px;
  font-weight: 600;
  margin-bottom: 8px;
}

.sheet-header .van-icon {
  font-size: 18px;
  color: #999;
  padding: 4px;
}

.sheet-tip {
  font-size: 13px;
  color: #8c8c8c;
  margin: 0 0 12px;
  line-height: 1.5;
}

.manual-locate-sheet :deep(.van-cell) {
  margin-bottom: 16px;
  border-radius: 10px;
  background: var(--app-bg, #eef2f8);
}

.manual-locate-btn {
  height: 44px;
  font-weight: 600;
}

.orders-body .page-content {
  padding: 0 16px;
}

.order-item .order-header {
  display: flex;
  justify-content: space-between;
  margin-bottom: 8px;
}
.order-no {
  font-size: 12px;
  color: #8c8c8c;
}
.price-block {
  text-align: right;
}
.price-block .price {
  display: block;
  font-size: 18px;
  font-weight: 600;
  color: #ff4d4f;
}
.price-hint {
  font-size: 11px;
  color: #8c8c8c;
}
.addr {
  font-size: 14px;
  margin: 6px 0;
  display: flex;
  align-items: center;
  gap: 4px;
}
.meta {
  display: flex;
  gap: 12px;
  font-size: 12px;
  color: #8c8c8c;
  margin-top: 8px;
}
.mt {
  margin-top: 12px;
}
</style>
