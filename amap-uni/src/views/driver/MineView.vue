<script setup lang="ts">
import { onMounted, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { showAppConfirm } from '@/utils/dialog'
import { useAuthStore } from '@/stores/auth'
import type { GrabOrderItem } from '@/types'
import { resolveActiveDriverTrip } from '@/utils/driverTrip'
import { openRealNameFromMine } from '@/utils/realName'
import { calcDriverIncome } from '@/utils/driverIncome'

const router = useRouter()
const route = useRoute()
const auth = useAuthStore()
const activeTrip = ref<GrabOrderItem | null>(null)

async function refreshTrip() {
  activeTrip.value = await resolveActiveDriverTrip()
}

onMounted(() => {
  void refreshTrip()
})

watch(() => route.fullPath, () => {
  void refreshTrip()
})

function goActiveTrip() {
  router.push('/driver/active')
}

function goVerify() {
  void openRealNameFromMine('driver', router)
}

async function logout() {
  try {
    await showAppConfirm({
      title: '确认退出登录？',
      message: '有进行中订单时请先完成订单；退出不会取消订单，重新登录后可继续',
      confirmButtonText: '退出',
      cancelButtonText: '取消',
    })
  } catch {
    return
  }
  auth.logout()
  router.replace('/driver/login')
}
</script>

<template>
  <div class="page">
    <van-nav-bar title="我的" />

    <div class="profile-card primary-gradient">
      <van-icon name="user-circle-o" size="48" />
      <div class="profile-info">
        <p class="phone">{{ auth.phone || '司机用户' }}</p>
        <p class="hint">高德网约车司机端</p>
      </div>
    </div>

    <div v-if="activeTrip" class="active-trip card" @click="goActiveTrip">
      <div class="active-trip-main">
        <van-icon name="logistics" class="active-trip-icon" />
        <div>
          <p class="active-trip-title">进行中的行程</p>
          <p class="active-trip-addr">{{ activeTrip.startAddress }} → {{ activeTrip.endAddress }}</p>
          <p class="active-trip-price">
            订单 ¥{{ activeTrip.price?.toFixed(2) }}
            <span class="active-trip-income"> · 预计收入 ¥{{ calcDriverIncome(activeTrip.price ?? 0).income.toFixed(2) }}</span>
          </p>
        </div>
      </div>
      <van-button size="small" type="primary" round>查看</van-button>
    </div>

    <van-cell-group inset class="menu-group">
      <van-cell
        title="进行中订单"
        icon="logistics"
        is-link
        :value="activeTrip ? '1 单进行中' : '暂无'"
        to="/driver/active"
      />
      <van-cell title="我的钱包" icon="balance-o" is-link to="/driver/wallet" />
      <van-cell title="历史订单" icon="orders-o" is-link to="/driver/order/list" />
      <van-cell title="资质认证" icon="certificate" is-link @click="goVerify" />
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
  padding: 14px 16px;
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
.active-trip-income {
  font-size: 13px;
  font-weight: 500;
  color: #1677ff;
}
.menu-group {
  margin-top: 12px;
}
</style>
