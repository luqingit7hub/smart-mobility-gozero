<script setup lang="ts">
import { onMounted, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { showAppConfirm } from '@/utils/dialog'
import { useAuthStore } from '@/stores/auth'
import {
  getActivePassengerOrder,
  orderStatusLabel,
  resolveActivePassengerOrder,
  setActivePassengerOrder,
  stopPassengerOrderWs,
} from '@/utils/passengerTrip'
import { openRealNameFromMine } from '@/utils/realName'
import type { ActiveOrder } from '@/types'

const router = useRouter()
const route = useRoute()
const auth = useAuthStore()

const activeTrip = ref<ActiveOrder | null>(null)

async function refreshTrip() {
  activeTrip.value = await resolveActivePassengerOrder()
}

onMounted(() => {
  activeTrip.value = getActivePassengerOrder()
  void refreshTrip()
})

watch(() => route.fullPath, () => {
  void refreshTrip()
})

function goActiveTrip() {
  router.push('/passenger/waiting')
}

function goRealName() {
  void openRealNameFromMine('passenger', router)
}

async function logout() {
  await showAppConfirm({ title: '确认退出登录？', message: '退出后需重新登录才能使用' })
  setActivePassengerOrder(null)
  stopPassengerOrderWs()
  auth.logout()
  router.replace('/passenger/login')
}
</script>

<template>
  <div class="page">
    <van-nav-bar title="我的" />

    <div class="profile-card primary-gradient">
      <van-icon name="user-circle-o" size="48" />
      <div class="profile-info">
        <p class="phone">{{ auth.phone || '乘客用户' }}</p>
        <p class="hint">高德网约车乘客端</p>
      </div>
    </div>

    <div v-if="activeTrip" class="active-trip card" @click="goActiveTrip">
      <div class="active-trip-main">
        <van-icon name="logistics" class="active-trip-icon" />
        <div>
          <p class="active-trip-title">{{ orderStatusLabel(activeTrip.status) }}</p>
          <p class="active-trip-addr">{{ activeTrip.start }} → {{ activeTrip.end }}</p>
          <p class="active-trip-price">¥{{ activeTrip.price?.toFixed(2) }}</p>
        </div>
      </div>
      <van-button size="small" type="primary" round>查看</van-button>
    </div>

    <van-cell-group inset class="menu-group">
      <van-cell
        title="进行中订单"
        icon="logistics"
        is-link
        :value="activeTrip ? orderStatusLabel(activeTrip.status) : '暂无'"
        to="/passenger/waiting"
      />
      <van-cell title="我的钱包" icon="balance-o" is-link to="/passenger/wallet" />
      <van-cell title="历史订单" icon="orders-o" is-link to="/passenger/order/list" />
      <van-cell title="我的优惠券" icon="coupon-o" is-link to="/passenger/coupons" />
      <van-cell title="实名认证" icon="certificate" is-link @click="goRealName" />
      <van-cell
        v-if="auth.isCompanyUser"
        title="发放优惠券"
        icon="gift-o"
        is-link
        to="/passenger/issue"
      />
    </van-cell-group>

    <van-cell-group inset class="menu-group">
      <van-cell title="退出登录" icon="revoke" is-link @click="logout" />
    </van-cell-group>
  </div>
</template>

<style scoped>
.profile-card {
  margin: 16px;
  padding: 24px;
  border-radius: 16px;
  display: flex;
  align-items: center;
  gap: 16px;
}
.profile-info .phone {
  font-size: 18px;
  font-weight: 600;
}
.profile-info .hint {
  font-size: 13px;
  opacity: 0.85;
  margin-top: 4px;
}
.active-trip {
  margin: 0 16px 12px;
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
.active-trip-price {
  font-size: 16px;
  font-weight: 600;
  color: #ff4d4f;
  margin-top: 6px;
}
.menu-group {
  margin-top: 12px;
}
</style>
