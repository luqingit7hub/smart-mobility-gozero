<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { clearCompletedTrip, getCompletedTrip, type CompletedTrip } from '@/utils/trip'
import { formatPrice } from '@/utils/order'

const router = useRouter()
const trip = ref<CompletedTrip | null>(null)

const priceParts = computed(() => formatPrice(trip.value?.price))

onMounted(() => {
  const data = getCompletedTrip()
  if (!data?.orderNo) {
    router.replace('/passenger')
    return
  }
  trip.value = data
})

function goHome() {
  clearCompletedTrip()
  router.replace('/passenger')
}
</script>

<template>
  <div class="page completed-page">
    <van-nav-bar title="行程结束" />

    <div v-if="trip" class="page-content">
      <div class="hero">
        <van-icon name="checked" size="64" color="#52c41a" />
        <h2>行程已完成</h2>
        <p class="subtitle">{{ trip.msg || '感谢使用高德网约车' }}</p>
      </div>

      <div class="card">
        <p class="order-no">订单号 {{ trip.orderNo }}</p>
        <div class="route">
          <p><span class="tag start">起</span>{{ trip.start }}</p>
          <p><span class="tag end">终</span>{{ trip.end }}</p>
        </div>
        <div v-if="trip.driverName || trip.carNumber" class="driver-line">
          <van-icon name="manager-o" />
          {{ trip.driverName || '司机师傅' }}
          <span v-if="trip.carNumber"> · {{ trip.carNumber }}</span>
        </div>
        <div class="price-row">
          <span>实付金额</span>
          <span class="price">
            {{ priceParts.symbol }}{{ priceParts.integer }}.{{ priceParts.decimal }}
          </span>
        </div>
      </div>

      <van-button type="primary" round block class="home-btn" @click="goHome">
        返回首页
      </van-button>
    </div>
  </div>
</template>

<style scoped>
.completed-page {
  min-height: 100vh;
  background: #f5f6f8;
}
.hero {
  text-align: center;
  padding: 32px 16px 24px;
}
.hero h2 {
  margin-top: 16px;
  font-size: 22px;
  font-weight: 600;
}
.subtitle {
  margin-top: 8px;
  font-size: 14px;
  color: #8c8c8c;
}
.order-no {
  font-size: 12px;
  color: #8c8c8c;
  margin-bottom: 12px;
}
.route p {
  display: flex;
  align-items: flex-start;
  gap: 10px;
  margin: 10px 0;
  font-size: 15px;
  line-height: 1.5;
}
.tag {
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
.tag.start {
  background: #52c41a;
}
.tag.end {
  background: #ff4d4f;
}
.driver-line {
  display: flex;
  align-items: center;
  gap: 6px;
  margin-top: 14px;
  padding-top: 14px;
  border-top: 1px dashed #f0f0f0;
  font-size: 14px;
  color: #666;
}
.price-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-top: 16px;
  padding: 14px;
  background: #fff7e6;
  border-radius: 10px;
  font-size: 14px;
}
.price {
  font-size: 22px;
  font-weight: 700;
  color: #ff6b00;
}
.home-btn {
  margin-top: 24px;
}
</style>
