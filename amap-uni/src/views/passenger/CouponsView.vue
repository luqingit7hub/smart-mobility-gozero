<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { userApi } from '@/api/user'
import type { CouponItem } from '@/types'
import { couponTypeLabels, formatCouponValue, normalizeCoupons } from '@/utils/geo'

const coupons = ref<CouponItem[]>([])
const loading = ref(true)

onMounted(async () => {
  try {
    const res = await userApi.listCoupons()
    coupons.value = normalizeCoupons(res.list || [])
  } finally {
    loading.value = false
  }
})
</script>

<template>
  <div class="page">
    <van-nav-bar title="我的优惠券" left-arrow @click-left="$router.push('/passenger/mine')" />
    <van-loading v-if="loading" class="loading" />
    <van-empty v-else-if="!coupons.length" description="暂无优惠券" />
    <div v-else class="page-content">
      <div v-for="c in coupons" :key="c.id" class="card coupon-card">
        <div class="coupon-left" :class="`type-${c.type}`">
          <span class="value">{{ formatCouponValue(c) }}</span>
          <span class="type-label">{{ couponTypeLabels[c.type] }}</span>
        </div>
        <div class="coupon-right">
          <h4>{{ couponTypeLabels[c.type] || '优惠券' }}</h4>
          <p class="text-muted">城市: {{ c.cityCode || '通用' }}</p>
          <p class="text-muted">有效期: {{ c.outTime || '-' }}</p>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.loading {
  display: flex;
  justify-content: center;
  padding: 48px;
}
.coupon-card {
  display: flex;
  overflow: hidden;
  padding: 0;
}
.coupon-left {
  width: 100px;
  color: #fff;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 16px 8px;
}
.coupon-left.type-1 {
  background: linear-gradient(135deg, #ff6b00, #ff9500);
}
.coupon-left.type-2 {
  background: linear-gradient(135deg, #1677ff, #4096ff);
}
.coupon-left.type-3 {
  background: linear-gradient(135deg, #52c41a, #73d13d);
}
.value {
  font-size: 20px;
  font-weight: 700;
}
.type-label {
  font-size: 11px;
  margin-top: 4px;
}
.coupon-right {
  flex: 1;
  padding: 16px;
}
.coupon-right h4 {
  margin-bottom: 6px;
}
</style>
