<script setup lang="ts">
import { onBeforeUnmount, onMounted, ref, shallowRef, watch } from 'vue'
import { loadBaiduMap } from '@/utils/baiduMap'

export interface TripMapPoint {
  lng: number
  lat: number
}

const props = defineProps<{
  start?: TripMapPoint | null
  end?: TripMapPoint | null
  userLocation?: TripMapPoint | null
  routePoints?: TripMapPoint[]
  pickMode?: 'start' | 'end' | null
  locating?: boolean
}>()

const emit = defineEmits<{
  pick: [lng: number, lat: number]
  locate: []
}>()

const containerRef = ref<HTMLElement>()
const mapInstance = shallowRef<BMapGL.Map | null>(null)
const mapError = ref('')
let startMarker: BMapGL.Marker | null = null
let endMarker: BMapGL.Marker | null = null
let userMarker: BMapGL.Marker | null = null
let routeLine: BMapGL.Polyline | null = null

const USER_LOCATION_ICON =
  'data:image/svg+xml,' +
  encodeURIComponent(
    '<svg xmlns="http://www.w3.org/2000/svg" width="28" height="28" viewBox="0 0 28 28">' +
      '<circle cx="14" cy="14" r="11" fill="#1677ff" fill-opacity="0.25"/>' +
      '<circle cx="14" cy="14" r="6" fill="#1677ff" stroke="#fff" stroke-width="2"/>' +
      '</svg>',
  )

function sameCoord(a?: TripMapPoint | null, b?: TripMapPoint | null) {
  if (!a || !b) return false
  return Math.abs(a.lng - b.lng) < 1e-5 && Math.abs(a.lat - b.lat) < 1e-5
}

function createUserMarker(point: BMapGL.Point) {
  const icon = new BMapGL.Icon(USER_LOCATION_ICON, new BMapGL.Size(28, 28), {
    anchor: new BMapGL.Size(14, 14),
  })
  return new BMapGL.Marker(point, { icon, title: '我的位置' })
}

function toPoint(p: TripMapPoint) {
  return new BMapGL.Point(p.lng, p.lat)
}

function isValidPoint(p?: TripMapPoint | null): p is TripMapPoint {
  return !!p && (p.lng !== 0 || p.lat !== 0)
}

function clearRoute() {
  if (mapInstance.value && routeLine) {
    mapInstance.value.removeOverlay(routeLine)
    routeLine = null
  }
}

function updateMapView() {
  const map = mapInstance.value
  if (!map) return

  if (startMarker) {
    map.removeOverlay(startMarker)
    startMarker = null
  }
  if (endMarker) {
    map.removeOverlay(endMarker)
    endMarker = null
  }
  if (userMarker) {
    map.removeOverlay(userMarker)
    userMarker = null
  }
  clearRoute()

  const markerPoints: BMapGL.Point[] = []

  if (isValidPoint(props.userLocation)) {
    const pt = toPoint(props.userLocation)
    userMarker = createUserMarker(pt)
    map.addOverlay(userMarker)
    markerPoints.push(pt)
  }

  if (isValidPoint(props.start) && !sameCoord(props.start, props.userLocation)) {
    const pt = toPoint(props.start)
    startMarker = new BMapGL.Marker(pt, { title: '起点' })
    map.addOverlay(startMarker)
    markerPoints.push(pt)
  }
  if (isValidPoint(props.end)) {
    const pt = toPoint(props.end)
    endMarker = new BMapGL.Marker(pt, { title: '终点' })
    map.addOverlay(endMarker)
    markerPoints.push(pt)
  }

  const route = props.routePoints?.filter((p) => isValidPoint(p)) ?? []
  if (route.length > 1) {
    const linePts = route.map(toPoint)
    routeLine = new BMapGL.Polyline(linePts, {
      strokeColor: '#1677ff',
      strokeWeight: 5,
      strokeOpacity: 0.85,
    })
    map.addOverlay(routeLine)
    map.setViewport(linePts)
    return
  }

  if (markerPoints.length === 1) {
    map.centerAndZoom(markerPoints[0], 15)
  } else if (markerPoints.length > 1) {
    map.setViewport(markerPoints)
  }
}

onMounted(async () => {
  try {
    await loadBaiduMap()
  } catch (e) {
    mapError.value = e instanceof Error ? e.message : '地图加载失败'
    return
  }
  if (!containerRef.value) return

  const map = new BMapGL.Map(containerRef.value)
  map.enableScrollWheelZoom(true)
  map.centerAndZoom(new BMapGL.Point(116.404, 39.915), 12)
  map.addEventListener('click', (e) => {
    if (props.pickMode) {
      emit('pick', e.latlng.lng, e.latlng.lat)
    }
  })
  mapInstance.value = map
  updateMapView()
})

watch(
  () => [props.start, props.end, props.userLocation, props.routePoints],
  () => updateMapView(),
  { deep: true },
)

onBeforeUnmount(() => {
  mapInstance.value?.destroy()
  mapInstance.value = null
})
</script>

<template>
  <div class="book-trip-map" :class="{ picking: !!pickMode }">
    <div ref="containerRef" class="map-container" />
    <div v-if="mapError" class="map-fallback">
      <van-icon name="location-o" />
      <p>{{ mapError }}</p>
      <p class="map-fallback-hint">
        请登录
        <a href="https://lbsyun.baidu.com/apiconsole/key" target="_blank" rel="noreferrer">百度地图开放平台</a>
        ，新建「浏览器端」应用并配置 AK，写入 .env.development 的 VITE_BAIDU_MAP_AK
      </p>
    </div>
    <button
      v-if="!mapError"
      type="button"
      class="locate-btn"
      :class="{ loading: locating }"
      :disabled="locating"
      aria-label="定位到我的位置"
      @click="emit('locate')"
    >
      <van-icon :name="locating ? 'loading' : 'aim'" :class="{ spin: locating }" />
    </button>
    <div v-if="!mapError && pickMode" class="pick-hint">
      点击地图选择{{ pickMode === 'start' ? '起点' : '终点' }}
    </div>
  </div>
</template>

<style scoped>
.book-trip-map {
  position: relative;
  width: 100%;
  height: 220px;
  background: #e8edf5;
}

.map-container {
  width: 100%;
  height: 100%;
}

.map-fallback {
  position: absolute;
  inset: 0;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 6px;
  padding: 16px;
  text-align: center;
  color: var(--app-text-secondary, #6b7280);
  background: #e8edf5;
}

.map-fallback .van-icon {
  font-size: 28px;
  color: var(--app-primary, #1677ff);
}

.map-fallback p {
  font-size: 13px;
  line-height: 1.5;
}

.map-fallback-hint {
  font-size: 12px;
  color: var(--app-text-muted, #9ca3af);
  line-height: 1.6;
}

.map-fallback-hint a {
  color: var(--app-primary, #1677ff);
}

.book-trip-map.picking .map-container {
  cursor: crosshair;
}

.locate-btn {
  position: absolute;
  right: 10px;
  bottom: 10px;
  z-index: 2;
  width: 36px;
  height: 36px;
  border: none;
  border-radius: 10px;
  background: #fff;
  color: var(--app-primary, #1677ff);
  box-shadow: 0 2px 10px rgba(15, 23, 42, 0.12);
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
}

.locate-btn:disabled {
  cursor: wait;
  opacity: 0.85;
}

.locate-btn .van-icon {
  font-size: 20px;
}

.locate-btn .spin {
  animation: locate-spin 0.8s linear infinite;
}

@keyframes locate-spin {
  from {
    transform: rotate(0deg);
  }
  to {
    transform: rotate(360deg);
  }
}

.pick-hint {
  position: absolute;
  left: 50%;
  bottom: 12px;
  transform: translateX(-50%);
  padding: 6px 14px;
  border-radius: 20px;
  background: rgba(22, 119, 255, 0.92);
  color: #fff;
  font-size: 12px;
  pointer-events: none;
  box-shadow: 0 2px 8px rgba(22, 119, 255, 0.35);
  z-index: 1;
}
</style>
