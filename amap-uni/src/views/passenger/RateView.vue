<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { showSuccessToast, showToast } from 'vant'
import { userApi } from '@/api/user'
import { formatPrice } from '@/utils/order'
import { getCompletedTrip, type CompletedTrip } from '@/utils/trip'

const router = useRouter()
const route = useRoute()

const trip = ref<CompletedTrip | null>(null)
const rating = ref(5)
const comment = ref('')
const loading = ref(false)

const tagOptions = ['准时', '服务好', '车辆整洁', '驾驶平稳', '态度友好']
const selectedTags = ref<string[]>([])

const priceParts = computed(() => formatPrice(trip.value?.price))

function isAlreadyRatedError(err: unknown): boolean {
  const msg = err instanceof Error ? err.message : String(err ?? '')
  return msg.includes('已评价')
}

function toggleTag(tag: string) {
  const idx = selectedTags.value.indexOf(tag)
  if (idx >= 0) {
    selectedTags.value.splice(idx, 1)
  } else if (selectedTags.value.length < 5) {
    selectedTags.value.push(tag)
  }
}

function goCompleted() {
  router.replace('/passenger/completed')
}

function goRatingDetail() {
  if (!trip.value?.orderNo) {
    router.replace('/passenger')
    return
  }
  const query: Record<string, string> = {
    orderNo: trip.value.orderNo,
    status: '3',
  }
  if (trip.value.start) query.start = trip.value.start
  if (trip.value.end) query.end = trip.value.end
  if (trip.value.price != null) query.price = String(trip.value.price)
  router.replace({ path: '/passenger/order/rating', query })
}

onMounted(() => {
  let data = getCompletedTrip()
  if (!data?.orderNo) {
    const qOrderNo = String(route.query.orderNo || '').trim()
    if (qOrderNo) {
      data = {
        orderNo: qOrderNo,
        start: String(route.query.start || '') || undefined,
        end: String(route.query.end || '') || undefined,
        price: (() => {
          const n = parseFloat(String(route.query.price || ''))
          return Number.isFinite(n) ? n : undefined
        })(),
      }
    }
  }
  if (!data?.orderNo) {
    router.replace('/passenger')
    return
  }
  trip.value = data
})

async function onSubmit() {
  if (!trip.value?.orderNo) {
    showToast('订单信息缺失')
    return
  }
  if (rating.value < 1) {
    showToast('请选择评分')
    return
  }

  loading.value = true
  try {
    const tags =
      selectedTags.value.length > 0 ? JSON.stringify(selectedTags.value) : undefined
    await userApi.rateOrder({
      order_no: trip.value.orderNo,
      rating: rating.value,
      comment: comment.value.trim() || undefined,
      tags,
    })
    showSuccessToast('评价成功，感谢您的反馈')
    goRatingDetail()
  } catch (err) {
    if (isAlreadyRatedError(err)) {
      showSuccessToast('您已评价过该订单')
      goRatingDetail()
      return
    }
    const msg = err instanceof Error ? err.message : '评价失败'
    showToast(msg)
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div class="page rate-page">
    <van-nav-bar title="评价司机" />

    <div v-if="trip" class="page-content">
      <div class="hero">
        <van-icon name="smile-o" size="56" color="#1677ff" />
        <h2>行程已结束</h2>
        <p class="subtitle">为司机师傅打个分吧</p>
      </div>

      <div class="card driver-card">
        <div class="driver-row">
          <van-icon name="manager-o" size="28" color="#1677ff" />
          <div>
            <p class="driver-name">{{ trip.driverName || '司机师傅' }}</p>
            <p v-if="trip.carNumber" class="car-no">{{ trip.carNumber }}</p>
          </div>
        </div>
        <div class="route-brief">
          <p>{{ trip.start }}</p>
          <van-icon name="arrow-down" class="arrow" />
          <p>{{ trip.end }}</p>
        </div>
        <p class="price-line">
          实付
          <span class="price">
            {{ priceParts.symbol }}{{ priceParts.integer }}.{{ priceParts.decimal }}
          </span>
        </p>
      </div>

      <div class="card">
        <p class="section-title">服务评分</p>
        <div class="rate-row">
          <van-rate v-model="rating" size="28" color="#ffd21e" void-icon="star" void-color="#eee" />
          <span class="rate-text">{{ rating }} 星</span>
        </div>

        <p class="section-title">快捷标签</p>
        <div class="tags">
          <span
            v-for="tag in tagOptions"
            :key="tag"
            class="tag"
            :class="{ active: selectedTags.includes(tag) }"
            @click="toggleTag(tag)"
          >
            {{ tag }}
          </span>
        </div>

        <van-field
          v-model="comment"
          rows="3"
          autosize
          type="textarea"
          maxlength="500"
          show-word-limit
          placeholder="分享您的乘车体验（选填）"
          class="comment-field"
        />
      </div>

      <van-button type="primary" round block :loading="loading" @click="onSubmit">
        提交评价
      </van-button>
      <van-button round block plain class="skip-btn" @click="goCompleted">暂不评价</van-button>
    </div>
  </div>
</template>

<style scoped>
.rate-page {
  min-height: 100vh;
  background: #f5f6f8;
}
.hero {
  text-align: center;
  padding: 28px 16px 16px;
}
.hero h2 {
  margin-top: 12px;
  font-size: 20px;
  font-weight: 600;
}
.subtitle {
  margin-top: 6px;
  font-size: 14px;
  color: #8c8c8c;
}
.driver-card {
  margin-bottom: 12px;
}
.driver-row {
  display: flex;
  align-items: center;
  gap: 12px;
}
.driver-name {
  font-size: 16px;
  font-weight: 600;
}
.car-no {
  margin-top: 4px;
  font-size: 13px;
  color: #8c8c8c;
}
.route-brief {
  margin-top: 14px;
  padding-top: 14px;
  border-top: 1px dashed #f0f0f0;
  font-size: 13px;
  color: #666;
  line-height: 1.5;
}
.arrow {
  margin: 4px 0;
  color: #bfbfbf;
}
.price-line {
  margin-top: 12px;
  font-size: 14px;
  color: #666;
}
.price {
  font-size: 20px;
  font-weight: 700;
  color: #ff6b00;
  margin-left: 4px;
}
.section-title {
  font-size: 14px;
  font-weight: 600;
  margin-bottom: 12px;
}
.rate-row {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 20px;
}
.rate-text {
  font-size: 14px;
  color: #ff6b00;
  font-weight: 600;
}
.tags {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  margin-bottom: 16px;
}
.tag {
  padding: 6px 14px;
  font-size: 13px;
  border-radius: 16px;
  background: #f5f5f5;
  color: #666;
  border: 1px solid transparent;
}
.tag.active {
  background: #e6f4ff;
  color: #1677ff;
  border-color: #91caff;
}
.comment-field {
  padding: 0;
}
.skip-btn {
  margin-top: 12px;
}
</style>
