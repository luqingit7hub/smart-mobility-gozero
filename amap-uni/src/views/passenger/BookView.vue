<script setup lang="ts">
import { computed, onActivated, onMounted, onUnmounted, ref, watch } from 'vue'
import { useRouter } from 'vue-router'
import { closeToast, showToast, showSuccessToast } from 'vant'
import { mapApi } from '@/api/map'
import { orderApi, userApi } from '@/api/user'
import BookTripMap, { type TripMapPoint } from '@/components/BookTripMap.vue'
import type { CouponItem, JourneyResult, RoutePoint } from '@/types'
import {
  trafficLabels,
  formatCouponValue,
  formatExpireTime,
  calcDiscountedPrice,
  normalizeCoupons,
  couponTypeLabels,
  annotateCouponsForOrder,
  getBd09LngLat,
  loadPassengerLocation,
  savePassengerLocation,
} from '@/utils/geo'
import type { CouponWithStatus } from '@/utils/coupon'
import { pickOrderNo, parseOrderNoFromText } from '@/utils/order'
import { buildActiveOrderFromTakeCar, setActivePassengerOrder } from '@/utils/passengerTrip'
import { showAppConfirm } from '@/utils/dialog'
import { PASSENGER_LOCATION_UPDATED } from '@/utils/tripLocation'

const router = useRouter()

const startPoint = ref('')
const endPoint = ref('')
const estimate = ref<JourneyResult | null>(null)
const coupons = ref<CouponWithStatus[]>([])
const selectedCoupon = ref(0)
const showCouponSheet = ref(false)
const estimating = ref(false)
const booking = ref(false)
const couponsLoadError = ref('')
const startCoord = ref<TripMapPoint | null>(null)
const endCoord = ref<TripMapPoint | null>(null)
const routePoints = ref<RoutePoint[]>([])
const pickMode = ref<'start' | 'end' | null>(null)
const userLocation = ref<TripMapPoint | null>(null)
const locating = ref(false)
const showManualLocateSheet = ref(false)
const manualAddressInput = ref('')
const manualLocating = ref(false)

let startGeocodeTimer: ReturnType<typeof setTimeout> | null = null
let endGeocodeTimer: ReturnType<typeof setTimeout> | null = null
let skipStartGeocode = false

const usableCoupons = computed(() => coupons.value.filter((c) => c.usable))
const unavailableCoupons = computed(() => coupons.value.filter((c) => !c.usable))

onMounted(() => {
  closeToast()
  estimate.value = null
  void loadCachedStartLocation()
  window.addEventListener(PASSENGER_LOCATION_UPDATED, onPassengerLocationUpdated)
})

onUnmounted(() => {
  window.removeEventListener(PASSENGER_LOCATION_UPDATED, onPassengerLocationUpdated)
})

onActivated(() => {
  void loadCachedStartLocation(true)
})

function onPassengerLocationUpdated(evt: Event) {
  const detail = (evt as CustomEvent<{ lng: number; lat: number; address: string }>).detail
  if (!detail || !Number.isFinite(detail.lng) || !Number.isFinite(detail.lat)) return
  applyStartLocation(detail.address, { lng: detail.lng, lat: detail.lat })
}

function applyStartLocation(address: string, point: TripMapPoint, options?: { userMarker?: boolean }) {
  skipStartGeocode = true
  startPoint.value = address
  startCoord.value = point
  userLocation.value = options?.userMarker === false ? null : point
  routePoints.value = []
  estimate.value = null
  savePassengerLocation({ lng: point.lng, lat: point.lat, address })
  skipStartGeocode = false
}

async function loadCachedStartLocation(force = false) {
  if (!force && startPoint.value.trim()) return
  const cached = loadPassengerLocation()
  if (!cached) return

  let { lng, lat, address } = cached
  if (!address.trim()) {
    try {
      const res = await mapApi.reverseGeocode(lng, lat)
      address = res.address
      lng = res.lng || lng
      lat = res.lat || lat
      savePassengerLocation({ lng, lat, address })
    } catch {
      return
    }
  }

  applyStartLocation(address, { lng, lat })
}

const selectedCouponItem = computed(() =>
  usableCoupons.value.find((c) => c.id === selectedCoupon.value) ?? null,
)

const finalPrice = computed(() => {
  if (!estimate.value) return 0
  return calcDiscountedPrice(estimate.value.price, selectedCouponItem.value)
})

const savedAmount = computed(() => {
  if (!estimate.value || !selectedCouponItem.value) return 0
  return Number((estimate.value.price - finalPrice.value).toFixed(2))
})

function pickStartAdcode(journey: JourneyResult): string {
  return journey.startAdcode ?? journey.start_adcode ?? ''
}

function coordFromJourney(journey: JourneyResult, which: 'start' | 'end'): TripMapPoint | null {
  if (which === 'start') {
    const lng = journey.startLng ?? journey.start_lng
    const lat = journey.startLat ?? journey.start_lat
    if (lng != null && lat != null) return { lng, lat }
  } else {
    const lng = journey.endLng ?? journey.end_lng
    const lat = journey.endLat ?? journey.end_lat
    if (lng != null && lat != null) return { lng, lat }
  }
  return null
}

function routePointsFromJourney(journey: JourneyResult): RoutePoint[] {
  return (journey.routePoints ?? journey.route_points ?? []).filter(
    (p) => p.lng != null && p.lat != null,
  )
}

function scheduleGeocode(
  address: string,
  target: typeof startCoord,
  clearTimer: () => void,
  setTimer: (t: ReturnType<typeof setTimeout>) => void,
) {
  clearTimer()
  const trimmed = address.trim()
  if (!trimmed) {
    target.value = null
    routePoints.value = []
    return
  }
  const timer = setTimeout(async () => {
    try {
      const res = await mapApi.getCoordinates(trimmed, 1)
      target.value = { lng: res.lng, lat: res.lat }
      routePoints.value = []
      estimate.value = null
    } catch {
      // 输入过程中可能尚未形成有效地址，静默忽略
    }
  }, 700)
  setTimer(timer)
}

watch(startPoint, (val) => {
  if (skipStartGeocode) return
  scheduleGeocode(
    val,
    startCoord,
    () => {
      if (startGeocodeTimer) clearTimeout(startGeocodeTimer)
    },
    (t) => {
      startGeocodeTimer = t
    },
  )
})

watch(endPoint, (val) => {
  scheduleGeocode(
    val,
    endCoord,
    () => {
      if (endGeocodeTimer) clearTimeout(endGeocodeTimer)
    },
    (t) => {
      endGeocodeTimer = t
    },
  )
})

function togglePickMode(mode: 'start' | 'end') {
  pickMode.value = pickMode.value === mode ? null : mode
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
  pickMode.value = null
  manualAddressInput.value = startPoint.value
  showManualLocateSheet.value = true
}

async function confirmManualLocate() {
  const address = manualAddressInput.value.trim()
  if (!address) {
    showToast('请输入上车地址')
    return
  }
  manualLocating.value = true
  try {
    const res = await mapApi.getCoordinates(address, 1)
    if (!Number.isFinite(res.lng) || !Number.isFinite(res.lat)) {
      showToast('未解析到有效坐标，请换个地址试试')
      return
    }
    applyStartLocation(res.address || address, { lng: res.lng, lat: res.lat }, { userMarker: false })
    showManualLocateSheet.value = false
    showSuccessToast('已设置上车位置')
  } catch {
    showToast('地址解析失败，请检查输入')
  } finally {
    manualLocating.value = false
  }
}

async function runAutoLocate() {
  if (locating.value) return
  locating.value = true
  pickMode.value = null
  try {
    const pos = await getBd09LngLat()
    if (!pos) {
      showToast('定位失败，请允许浏览器使用位置权限')
      return
    }
    const res = await mapApi.reverseGeocode(pos.lng, pos.lat)
    const point = { lng: res.lng || pos.lng, lat: res.lat || pos.lat }
    applyStartLocation(res.address, point)
    showSuccessToast('已定位到当前位置')
  } catch {
    showToast('定位失败，请稍后重试')
  } finally {
    locating.value = false
  }
}

async function onMapPick(lng: number, lat: number) {
  if (!pickMode.value) return
  try {
    const res = await mapApi.reverseGeocode(lng, lat)
    const point = { lng: res.lng || lng, lat: res.lat || lat }
    if (pickMode.value === 'start') {
      applyStartLocation(res.address, point)
    } else {
      endPoint.value = res.address
      endCoord.value = point
    }
    routePoints.value = []
    estimate.value = null
    pickMode.value = null
  } catch {
    showToast('地址解析失败')
  }
}

async function loadCoupons(startAdcode?: string) {
  couponsLoadError.value = ''
  try {
    const res = await userApi.listCoupons({ silent: true })
    const list = normalizeCoupons(res.list || []) as CouponItem[]
    coupons.value = annotateCouponsForOrder(list, startAdcode)
    if (selectedCoupon.value > 0 && !usableCoupons.value.some((c) => c.id === selectedCoupon.value)) {
      selectedCoupon.value = 0
    }
  } catch {
    coupons.value = []
    couponsLoadError.value = '优惠券加载失败'
  }
}

async function openCouponSheet() {
  if (couponsLoadError.value) {
    if (estimate.value) {
      await loadCoupons(pickStartAdcode(estimate.value))
    }
    if (couponsLoadError.value) {
      showToast(couponsLoadError.value)
      return
    }
  }
  if (!coupons.value.length) {
    showToast('暂无优惠券')
    return
  }
  if (!usableCoupons.value.length) {
    showToast('本单暂无可用优惠券')
  }
  showCouponSheet.value = true
}

function pickCoupon(c: CouponWithStatus | null) {
  if (c && !c.usable) {
    showToast(c.unusableReason || '该优惠券本单不可用')
    return
  }
  selectedCoupon.value = c?.id ?? 0
  showCouponSheet.value = false
}

async function onEstimate() {
  if (!startPoint.value || !endPoint.value) {
    showToast('请填写起点和终点')
    return
  }
  estimating.value = true
  try {
    estimate.value = await orderApi.journey(startPoint.value, endPoint.value)
    startCoord.value = coordFromJourney(estimate.value, 'start') ?? startCoord.value
    endCoord.value = coordFromJourney(estimate.value, 'end') ?? endCoord.value
    routePoints.value = routePointsFromJourney(estimate.value)
    selectedCoupon.value = 0
    await loadCoupons(pickStartAdcode(estimate.value))
  } finally {
    estimating.value = false
  }
}

async function onBook() {
  if (!startPoint.value || !endPoint.value) {
    showToast('请填写起点和终点')
    return
  }
  booking.value = true
  try {
    const payload: { starting_point: string; destination: string; tid?: number } = {
      starting_point: startPoint.value,
      destination: endPoint.value,
    }
    if (selectedCoupon.value > 0) {
      payload.tid = selectedCoupon.value
    }
    const res = await orderApi.takeCar(payload)
    let orderNo = res.orderNo || pickOrderNo(res as Record<string, unknown>)
    if (!orderNo && res.text) {
      orderNo = parseOrderNoFromText(res.text)
    }
    if (!orderNo) {
      try {
        const ongoing = await orderApi.ongoingOrder({ silent: true })
        const raw = ongoing?.order as Record<string, unknown> | undefined
        const hasOrder = Boolean(
          ongoing?.hasOrder ?? (ongoing as { has_order?: boolean }).has_order ?? raw?.orderNo ?? raw?.order_no,
        )
        if (hasOrder && raw) {
          orderNo = String(raw.orderNo ?? raw.order_no ?? '')
        }
      } catch {
        // ignore
      }
    }
    if (!orderNo) {
      showToast('下单成功，但未获取到订单号，请稍后在订单中查看')
      return
    }
    sessionStorage.setItem('takeCarRaw', JSON.stringify(res))
    setActivePassengerOrder(
      buildActiveOrderFromTakeCar(orderNo, startPoint.value, endPoint.value, res, {
        startLng: startCoord.value?.lng,
        startLat: startCoord.value?.lat,
        endLng: endCoord.value?.lng,
        endLat: endCoord.value?.lat,
      }),
    )
    estimate.value = null
    showSuccessToast(res.text?.trim() || '下单成功，正在为您寻找司机')
    await router.replace('/passenger/waiting')
  } finally {
    booking.value = false
  }
}
</script>

<template>
  <div class="page book-page">
    <van-nav-bar title="叫车出行" />

    <div class="page-content">
      <div class="trip-hero primary-gradient">
        <div class="hero-icon-wrap">
          <van-icon name="logistics" />
        </div>
        <div class="hero-text">
          <p class="hero-title">快速叫车</p>
          <p class="hero-hint">输入起终点，获取实时估价</p>
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

      <!-- 地图：起终点标记、路线、选点 -->
      <BookTripMap
        class="book-map-slot"
        :start="startCoord"
        :end="endCoord"
        :user-location="userLocation"
        :route-points="routePoints"
        :pick-mode="pickMode"
        :locating="locating"
        @pick="onMapPick"
        @locate="onLocateMe"
      />

      <div class="trip-form card">
          <div class="route-fields">
            <div class="route-line" aria-hidden="true">
              <span class="route-dot start" />
              <span class="route-connector" />
              <span class="route-dot end" />
            </div>
            <div class="route-inputs">
              <van-field
                v-model="startPoint"
                label="起点"
                left-icon="location-o"
                placeholder="请输入上车地点"
                :border="false"
              />
              <van-field
                v-model="endPoint"
                label="终点"
                left-icon="aim"
                placeholder="请输入目的地"
                :border="false"
              />
            </div>
          </div>
          <div class="map-pick-actions">
            <button
              type="button"
              class="map-pick-btn"
              :class="{ active: pickMode === 'start' }"
              @click="togglePickMode('start')"
            >
              <van-icon name="location-o" />
              地图选起点
            </button>
            <button
              type="button"
              class="map-pick-btn"
              :class="{ active: pickMode === 'end' }"
              @click="togglePickMode('end')"
            >
              <van-icon name="aim" />
              地图选终点
            </button>
          </div>
          <van-button
            type="primary"
            round
            block
            class="estimate-btn"
            :loading="estimating"
            @click="onEstimate"
          >
            预估行程
          </van-button>
      </div>

      <div v-if="estimate" class="card estimate-card">
        <h3>行程预估</h3>
        <div class="estimate-row">
          <span>距离</span>
          <span>{{ estimate.distance }} km</span>
        </div>
        <div class="estimate-row">
          <span>预计时长</span>
          <span>{{ estimate.duration }} 分钟</span>
        </div>
        <div class="estimate-row">
          <span>路况</span>
          <span class="tag-traffic">{{ trafficLabels[estimate.status] || '未知' }}</span>
        </div>

        <div class="price-block">
          <div class="price-main">
            <span class="label">预估费用</span>
            <div class="price-values">
              <span v-if="selectedCouponItem" class="origin-price">
                ¥{{ estimate.price?.toFixed(2) }}
              </span>
              <span class="price">¥{{ finalPrice.toFixed(2) }}</span>
            </div>
          </div>
          <p v-if="selectedCouponItem && savedAmount > 0" class="saved-tip">
            已优惠 ¥{{ savedAmount.toFixed(2) }}
          </p>
        </div>

        <div
          v-if="coupons.length || couponsLoadError"
          class="coupon-entry"
          @click="openCouponSheet"
        >
          <div class="coupon-entry-left">
            <van-icon name="coupon-o" class="coupon-icon" />
            <div>
              <p class="coupon-entry-title">优惠券</p>
              <p v-if="couponsLoadError" class="coupon-entry-desc warn">
                {{ couponsLoadError }}，点击重试
              </p>
              <p v-else-if="selectedCouponItem" class="coupon-entry-desc">
                {{ couponTypeLabels[selectedCouponItem.type] }} ·
                {{ formatCouponValue(selectedCouponItem) }}
              </p>
              <p v-else-if="usableCoupons.length" class="coupon-entry-desc muted">
                {{ usableCoupons.length }} 张可用，点击选择
              </p>
              <p v-else class="coupon-entry-desc warn">
                {{ coupons.length }} 张券本单不可用，点击查看原因
              </p>
            </div>
          </div>
          <van-icon name="arrow" class="arrow" />
        </div>

        <van-button
          type="primary"
          round
          block
          class="mt"
          :loading="booking"
          @click="onBook"
        >
          立即叫车 · ¥{{ finalPrice.toFixed(2) }}
        </van-button>
      </div>
    </div>

    <van-popup
      v-if="showCouponSheet"
      v-model:show="showCouponSheet"
      position="bottom"
      round
      :style="{ maxHeight: '70vh' }"
    >
      <div class="coupon-sheet">
        <div class="sheet-header">
          <span>选择优惠券</span>
          <van-icon name="cross" @click="showCouponSheet = false" />
        </div>

        <div class="sheet-list">
          <div
            class="coupon-card"
            :class="{ active: selectedCoupon === 0 }"
            @click="pickCoupon(null)"
          >
            <div class="coupon-card-body no-coupon">
              <p class="coupon-name">不使用优惠券</p>
              <p class="coupon-meta">按原价 ¥{{ estimate?.price?.toFixed(2) }} 支付</p>
            </div>
            <van-icon v-if="selectedCoupon === 0" name="success" class="check-icon" />
          </div>

          <div
            v-for="c in usableCoupons"
            :key="c.id"
            class="coupon-card"
            :class="{ active: selectedCoupon === c.id }"
            @click="pickCoupon(c)"
          >
            <div class="coupon-card-left" :class="`type-${c.type}`">
              <span class="coupon-value">{{ formatCouponValue(c) }}</span>
              <span class="coupon-type">{{ couponTypeLabels[c.type] }}</span>
            </div>
            <div class="coupon-card-body">
              <p class="coupon-name">{{ couponTypeLabels[c.type] || '优惠券' }}</p>
              <p class="coupon-meta">{{ formatExpireTime(c.outTime) }}</p>
              <p v-if="estimate" class="coupon-price">
                券后约
                <strong>¥{{ calcDiscountedPrice(estimate.price, c).toFixed(2) }}</strong>
              </p>
            </div>
            <van-icon v-if="selectedCoupon === c.id" name="success" class="check-icon" />
          </div>

          <template v-if="unavailableCoupons.length">
            <p class="sheet-section-title">本单不可用</p>
            <div
              v-for="c in unavailableCoupons"
              :key="`disabled-${c.id}`"
              class="coupon-card disabled"
              @click="pickCoupon(c)"
            >
              <div class="coupon-card-left" :class="`type-${c.type}`">
                <span class="coupon-value">{{ formatCouponValue(c) }}</span>
                <span class="coupon-type">{{ couponTypeLabels[c.type] }}</span>
              </div>
              <div class="coupon-card-body">
                <p class="coupon-name">{{ couponTypeLabels[c.type] || '优惠券' }}</p>
                <p class="coupon-meta warn">{{ c.unusableReason }}</p>
                <p class="coupon-meta">{{ formatExpireTime(c.outTime) }}</p>
              </div>
            </div>
          </template>
        </div>
      </div>
    </van-popup>

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
        <p class="sheet-tip">输入上车地址，系统将转换为地图坐标并显示</p>
        <van-field
          v-model="manualAddressInput"
          label="上车地址"
          placeholder="如：上海市浦东新区惠南镇盐大路"
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
.book-page {
  background: var(--app-bg, #eef2f8);
}

.book-page .page-content {
  padding-top: 12px;
}

.trip-hero {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 16px 14px;
  color: #fff;
  border-radius: var(--app-radius-md, 14px);
  box-shadow: var(--app-shadow-sm, 0 2px 10px rgba(15, 23, 42, 0.06));
  margin-bottom: 12px;
}

.hero-text {
  flex: 1;
  min-width: 0;
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

.book-map-slot {
  margin-bottom: 12px;
  border-radius: var(--app-radius-md, 14px);
  overflow: hidden;
}

.map-pick-actions {
  display: flex;
  gap: 8px;
  margin: 4px 2px 8px;
}

.map-pick-btn {
  flex: 1;
  min-width: 0;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 4px;
  height: 34px;
  border: 1px solid rgba(22, 119, 255, 0.25);
  border-radius: 999px;
  background: #fff;
  color: var(--app-primary, #1677ff);
  font-size: 12px;
  cursor: pointer;
}

.map-pick-btn.active {
  background: var(--app-primary, #1677ff);
  border-color: var(--app-primary, #1677ff);
  color: #fff;
}

.trip-form {
  padding: 12px 14px 16px;
  margin-bottom: 12px;
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

.hero-title {
  font-size: 17px;
  font-weight: 600;
  letter-spacing: 0.3px;
}

.hero-hint {
  margin-top: 4px;
  font-size: 12px;
  opacity: 0.9;
  line-height: 1.5;
}

.route-fields {
  display: flex;
  gap: 10px;
  align-items: stretch;
  padding: 4px 2px 8px;
}

.route-line {
  display: flex;
  flex-direction: column;
  align-items: center;
  width: 8px;
  padding: 16px 0;
  flex-shrink: 0;
}

.route-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  flex-shrink: 0;
}

.route-dot.start {
  background: #52c41a;
}

.route-dot.end {
  background: #ff4d4f;
}

.route-connector {
  flex: 1;
  width: 2px;
  min-height: 20px;
  margin: 4px 0;
  border-radius: 1px;
  background: linear-gradient(180deg, #52c41a 0%, #d9d9d9 50%, #ff4d4f 100%);
}

.route-inputs {
  flex: 1;
  min-width: 0;
  border-radius: 12px;
  background: var(--app-bg, #eef2f8);
  overflow: hidden;
}

.route-inputs :deep(.van-cell) {
  background: transparent;
  padding: 12px 14px;
}

.route-inputs :deep(.van-cell:after) {
  left: 14px;
  right: 14px;
  border-color: rgba(15, 23, 42, 0.06);
}

.route-inputs :deep(.van-field__left-icon) {
  color: var(--app-primary, #1677ff);
}

.estimate-btn {
  margin-top: 4px;
  height: 44px;
  font-weight: 600;
  box-shadow: 0 4px 12px rgba(22, 119, 255, 0.28);
}

.estimate-card {
  margin-top: 0;
}

.estimate-row {
  display: flex;
  justify-content: space-between;
  padding: 6px 0;
  font-size: 14px;
  color: #666;
}
h3 {
  margin-bottom: 8px;
  font-size: 16px;
}
.mt {
  margin-top: 16px;
}

.price-block {
  margin-top: 12px;
  padding: 14px;
  background: #fafafa;
  border-radius: 10px;
}
.price-main {
  display: flex;
  justify-content: space-between;
  align-items: flex-end;
}
.price-main .label {
  font-size: 14px;
  color: #666;
}
.price-values {
  display: flex;
  align-items: baseline;
  gap: 8px;
}
.origin-price {
  font-size: 14px;
  color: #999;
  text-decoration: line-through;
}
.saved-tip {
  margin-top: 6px;
  font-size: 12px;
  color: #ff6b00;
  text-align: right;
}

.coupon-entry {
  margin-top: 12px;
  padding: 12px 14px;
  background: #fff7e6;
  border: 1px solid #ffe7ba;
  border-radius: 10px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  cursor: pointer;
}
.coupon-entry-left {
  display: flex;
  align-items: center;
  gap: 10px;
}
.coupon-icon {
  font-size: 22px;
  color: #ff6b00;
}
.coupon-entry-title {
  font-size: 14px;
  font-weight: 500;
}
.coupon-entry-desc {
  font-size: 12px;
  color: #ff6b00;
  margin-top: 2px;
}
.coupon-entry-desc.muted {
  color: #8c8c8c;
}
.coupon-entry-desc.warn {
  color: #cf1322;
}
.arrow {
  color: #ccc;
}

.coupon-sheet {
  padding: 16px 16px 24px;
}
.sheet-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  font-size: 16px;
  font-weight: 600;
  margin-bottom: 16px;
}
.sheet-header .van-icon {
  font-size: 18px;
  color: #999;
  padding: 4px;
}
.sheet-section-title {
  margin: 8px 0 12px;
  font-size: 13px;
  color: #8c8c8c;
}
.sheet-list {
  max-height: 55vh;
  overflow-y: auto;
}

.coupon-card {
  display: flex;
  align-items: stretch;
  margin-bottom: 12px;
  border-radius: 12px;
  overflow: hidden;
  border: 2px solid transparent;
  background: #f7f8fa;
  position: relative;
  cursor: pointer;
}
.coupon-card.active {
  border-color: #1677ff;
  background: #f0f7ff;
}
.coupon-card.disabled {
  opacity: 0.55;
  cursor: not-allowed;
}
.coupon-card-left {
  width: 96px;
  flex-shrink: 0;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 12px 8px;
  color: #fff;
}
.coupon-card-left.type-1 {
  background: linear-gradient(135deg, #ff6b00, #ff9500);
}
.coupon-card-left.type-2 {
  background: linear-gradient(135deg, #1677ff, #4096ff);
}
.coupon-card-left.type-3 {
  background: linear-gradient(135deg, #52c41a, #73d13d);
}
.coupon-value {
  font-size: 20px;
  font-weight: 700;
  line-height: 1.2;
}
.coupon-type {
  font-size: 11px;
  margin-top: 4px;
  opacity: 0.9;
}
.coupon-card-body {
  flex: 1;
  padding: 12px 36px 12px 14px;
  display: flex;
  flex-direction: column;
  justify-content: center;
}
.coupon-card-body.no-coupon {
  padding-left: 16px;
}
.coupon-name {
  font-size: 14px;
  font-weight: 500;
}
.coupon-meta {
  font-size: 12px;
  color: #8c8c8c;
  margin-top: 4px;
}
.coupon-meta.warn {
  color: #cf1322;
}
.coupon-price {
  font-size: 12px;
  color: #666;
  margin-top: 6px;
}
.coupon-price strong {
  color: #ff6b00;
  font-size: 15px;
}
.check-icon {
  position: absolute;
  right: 12px;
  top: 50%;
  transform: translateY(-50%);
  color: #1677ff;
  font-size: 20px;
}

.manual-locate-sheet {
  padding: 16px 16px 24px;
}

.manual-locate-sheet .sheet-tip {
  margin: 0 0 12px;
  font-size: 13px;
  color: #8c8c8c;
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
</style>
