<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import type { GrabOrderItem } from '@/types'
import { getActiveDriverTrip, resolveActiveDriverTrip } from '@/utils/driverTrip'

const route = useRoute()
const router = useRouter()

const activeTrip = ref<GrabOrderItem | null>(null)

const hideTabbar = computed(() =>
  route.path.startsWith('/driver/verify') || route.path.startsWith('/driver/wallet'),
)

const isChatTab = computed(() => route.path.startsWith('/driver/ai-chat'))

const showActiveBar = computed(
  () => !!activeTrip.value && !route.path.startsWith('/driver/active') && !isChatTab.value,
)

const activeTab = computed(() => {
  if (route.path.startsWith('/driver/mine')) return 2
  if (isChatTab.value) return 1
  return 0
})

function syncActiveTrip() {
  activeTrip.value = getActiveDriverTrip()
}

async function refreshActiveTripFromServer() {
  activeTrip.value = await resolveActiveDriverTrip()
}

function goActiveTrip() {
  router.push('/driver/active')
}

function onTabChange(index: number) {
  if (index === 0) router.replace('/driver')
  else if (index === 1) router.replace('/driver/ai-chat')
  else router.replace('/driver/mine')
}

function onStorage(e: StorageEvent) {
  if (e.key === 'activeDriverOrder') syncActiveTrip()
}

function onActiveTripChange() {
  syncActiveTrip()
}

onMounted(() => {
  syncActiveTrip()
  void refreshActiveTripFromServer()
  window.addEventListener('storage', onStorage)
  window.addEventListener('activeDriverOrderChange', onActiveTripChange)
})

onUnmounted(() => {
  window.removeEventListener('storage', onStorage)
  window.removeEventListener('activeDriverOrderChange', onActiveTripChange)
})

watch(() => route.fullPath, () => {
  syncActiveTrip()
  void refreshActiveTripFromServer()
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
          <p class="active-trip-bar-title">进行中的行程</p>
          <p class="active-trip-bar-addr">
            {{ activeTrip?.startAddress }} → {{ activeTrip?.endAddress }}
          </p>
        </div>
      </div>
      <van-button size="small" type="primary" round>查看</van-button>
    </div>

    <van-tabbar v-if="!hideTabbar" :model-value="activeTab" @update:model-value="onTabChange">
      <van-tabbar-item icon="orders-o">抢单</van-tabbar-item>
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
