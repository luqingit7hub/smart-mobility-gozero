<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { showSuccessToast, showToast } from 'vant'
import { showAppConfirm } from '@/utils/dialog'
import type { OrderRatingItem } from '@/types'
import {
  isNoRatingError,
  isOrderCompleted,
  normalizeOrderRating,
  parseRatingTags,
} from '@/utils/orderList'
import { saveCompletedTrip } from '@/utils/trip'

const props = defineProps<{
  role: 'passenger' | 'driver'
  backPath: string
  fetchRating: (orderNo: string) => Promise<OrderRatingItem>
  deleteRating?: (orderNo: string) => Promise<void>
}>()

const route = useRoute()
const router = useRouter()

const loading = ref(true)
const deleting = ref(false)
const rating = ref<OrderRatingItem | null>(null)
const empty = ref(false)

const orderNo = computed(() => String(route.query.orderNo || '').trim())
const orderStatus = computed(() => Number(route.query.status || 0))
const orderStart = computed(() => String(route.query.start || ''))
const orderEnd = computed(() => String(route.query.end || ''))
const orderPrice = computed(() => {
  const raw = route.query.price
  const n = typeof raw === 'string' ? parseFloat(raw) : NaN
  return Number.isFinite(n) ? n : undefined
})

const tagList = computed(() => parseRatingTags(rating.value?.tags))

const emptyText = computed(() => {
  if (props.role === 'driver') return '乘客尚未评价该订单'
  if (isOrderCompleted(orderStatus.value)) return '您尚未评价该订单'
  return '仅已完成订单可查看或提交评价'
})

async function loadRating() {
  if (!orderNo.value) {
    showToast('订单号缺失')
    router.replace(props.backPath)
    return
  }

  loading.value = true
  empty.value = false
  rating.value = null
  try {
    const data = await props.fetchRating(orderNo.value)
    rating.value = normalizeOrderRating(data)
  } catch (err) {
    if (isNoRatingError(err)) {
      empty.value = true
      if (props.role === 'passenger' && isOrderCompleted(orderStatus.value)) {
        await promptPassengerRate()
      }
      return
    }
    const msg = err instanceof Error ? err.message : '评价加载失败'
    showToast(msg)
  } finally {
    loading.value = false
  }
}

async function promptPassengerRate() {
  try {
    await showAppConfirm({
      title: '暂无评价',
      message: '该订单尚未评价，是否前往评价？',
      confirmButtonText: '去评价',
      cancelButtonText: '暂不评价',
    })
    saveCompletedTrip({
      orderNo: orderNo.value,
      start: orderStart.value || undefined,
      end: orderEnd.value || undefined,
      price: orderPrice.value,
    })
    router.push('/passenger/rate')
  } catch {
    // 用户取消
  }
}

async function handleDeleteRating() {
  if (!props.deleteRating || !orderNo.value || deleting.value) return
  try {
    await showAppConfirm({
      title: '删除评价',
      message: '确定要删除该评价吗？删除后可重新评价。',
      confirmButtonText: '删除',
      cancelButtonText: '取消',
    })
  } catch {
    return
  }

  deleting.value = true
  try {
    await props.deleteRating(orderNo.value)
    rating.value = null
    empty.value = true
    showSuccessToast('评价已删除')
  } catch (err) {
    const msg = err instanceof Error ? err.message : '删除失败'
    showToast(msg)
  } finally {
    deleting.value = false
  }
}

function goRateFromEmpty() {
  if (props.role !== 'passenger' || !isOrderCompleted(orderStatus.value)) return
  saveCompletedTrip({
    orderNo: orderNo.value,
    start: orderStart.value || undefined,
    end: orderEnd.value || undefined,
    price: orderPrice.value,
  })
  router.push('/passenger/rate')
}

onMounted(() => {
  void loadRating()
})
</script>

<template>
  <div class="page">
    <van-nav-bar title="评价详情" left-arrow @click-left="$router.push(backPath)" />

    <div v-if="loading" class="loading-wrap">
      <van-loading size="24">加载中...</van-loading>
    </div>

    <div v-else-if="rating" class="page-content">
      <div class="card summary-card">
        <p class="label">订单号</p>
        <p class="value">{{ rating.orderNo || orderNo }}</p>
      </div>

      <div class="card">
        <p class="section-title">服务评分</p>
        <div class="rate-row">
          <van-rate
            :model-value="rating.rating || 0"
            readonly
            size="24"
            color="#ffd21e"
            void-icon="star"
            void-color="#eee"
          />
          <span class="rate-text">{{ rating.rating || 0 }} 星</span>
        </div>

        <p v-if="tagList.length" class="section-title">标签</p>
        <div v-if="tagList.length" class="tags">
          <span v-for="tag in tagList" :key="tag" class="tag">{{ tag }}</span>
        </div>

        <p class="section-title">评价内容</p>
        <p class="comment">{{ rating.comment?.trim() || '暂无文字评价' }}</p>

        <p v-if="rating.createdAt" class="created-at">评价时间：{{ rating.createdAt }}</p>
      </div>

      <van-button
        v-if="deleteRating"
        round
        block
        plain
        type="danger"
        class="action-btn"
        :loading="deleting"
        @click="handleDeleteRating"
      >
        删除评价
      </van-button>
    </div>

    <div v-else-if="empty" class="page-content">
      <van-empty :description="emptyText" />
      <van-button
        v-if="role === 'passenger' && isOrderCompleted(orderStatus)"
        type="primary"
        round
        block
        class="action-btn"
        @click="goRateFromEmpty"
      >
        去评价
      </van-button>
      <van-button round block plain class="action-btn" @click="$router.push(backPath)">
        返回订单列表
      </van-button>
    </div>
  </div>
</template>

<style scoped>
.loading-wrap {
  display: flex;
  justify-content: center;
  padding: 48px 16px;
}
.summary-card .label {
  font-size: 13px;
  color: var(--van-text-color-3);
}
.summary-card .value {
  margin-top: 6px;
  font-size: 15px;
  font-weight: 600;
  word-break: break-all;
}
.section-title {
  font-size: 14px;
  font-weight: 600;
  margin-bottom: 10px;
}
.rate-row {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 16px;
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
  padding: 4px 12px;
  font-size: 12px;
  border-radius: 14px;
  background: #e6f4ff;
  color: #1677ff;
}
.comment {
  font-size: 14px;
  line-height: 1.7;
  color: var(--van-text-color-2);
  white-space: pre-wrap;
}
.created-at {
  margin-top: 16px;
  font-size: 12px;
  color: var(--van-text-color-3);
}
.action-btn {
  margin: 12px 16px 0;
}
</style>
