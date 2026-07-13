<script setup lang="ts">
import { computed, onActivated, onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { showSuccessToast, showToast } from 'vant'
import { showAppConfirm } from '@/utils/dialog'
import { driverApi } from '@/api/driver'
import type { GrabOrderItem } from '@/types'
import { ORDER_STATUS_ACCEPTED, ORDER_STATUS_ON_BOARD } from '@/types'
import { calcDriverIncome } from '@/utils/driverIncome'
import {
  finishDriverTrip,
  isOrderAlreadyCompletedError,
  isRequestTimeoutError,
  recoverOrderOverAfterTimeout,
  resolveActiveDriverTrip,
  setActiveDriverTrip,
  syncActiveDriverTripFromServer,
} from '@/utils/driverTrip'
import { syncDriverLocationToOrderEnd } from '@/utils/tripLocation'

const router = useRouter()

const order = ref<GrabOrderItem | null>(null)
const completing = ref(false)
const starting = ref(false)
const syncing = ref(false)
const phoneTail = ref('')

const incomeBreakdown = computed(() => calcDriverIncome(order.value?.price ?? 0))

/** 无 status 字段时视为已接单（需先开始行程） */
const orderStatus = computed(() => {
  const s = order.value?.status
  if (s === ORDER_STATUS_ON_BOARD) return ORDER_STATUS_ON_BOARD
  if (s === ORDER_STATUS_ACCEPTED) return ORDER_STATUS_ACCEPTED
  return ORDER_STATUS_ACCEPTED
})

const needStartTrip = computed(() => orderStatus.value === ORDER_STATUS_ACCEPTED)
const canComplete = computed(() => orderStatus.value === ORDER_STATUS_ON_BOARD)

async function loadActiveOrder() {
  syncing.value = true
  try {
    order.value = await resolveActiveDriverTrip()
  } finally {
    syncing.value = false
  }
}

onMounted(loadActiveOrder)
onActivated(loadActiveOrder)

async function handleCompleteSuccess(income: number, orderPrice: number, completedOrder: GrabOrderItem | null) {
  if (completedOrder) {
    await syncDriverLocationToOrderEnd({
      endLng: completedOrder.endLng,
      endLat: completedOrder.endLat,
      endAddress: completedOrder.endAddress,
    })
  }
  showSuccessToast(`完单成功，收入 ¥${income.toFixed(2)} 已入账（订单 ¥${orderPrice.toFixed(2)} 的 85%）`)
  order.value = null
  phoneTail.value = ''
  syncActiveDriverTripFromServer()
  finishDriverTrip(router)
}

async function onStartTrip() {
  if (!order.value?.orderNo) {
    showToast('暂无进行中订单')
    return
  }
  const tail = phoneTail.value.trim()
  if (!/^\d{4}$/.test(tail)) {
    showToast('请输入乘客手机号后四位')
    return
  }

  starting.value = true
  try {
    const res = await driverApi.startOrder(order.value.orderNo, tail)
    showSuccessToast(res.msg || '行程已开始')
    const updated = { ...order.value, status: ORDER_STATUS_ON_BOARD }
    order.value = updated
    setActiveDriverTrip(updated)
    phoneTail.value = ''
  } catch (err) {
    const msg = err instanceof Error ? err.message : '开始行程失败'
    showToast(msg)
  } finally {
    starting.value = false
  }
}

async function onComplete() {
  if (!order.value) {
    showToast('暂无进行中订单')
    return
  }
  if (!canComplete.value) {
    showToast('请先确认乘客上车后再完单')
    return
  }
  const orderNo = order.value.orderNo
  const { orderPrice, income, commission } = incomeBreakdown.value
  try {
    await showAppConfirm({
      title: '确认完成订单',
      message: `订单金额 ¥${orderPrice.toFixed(2)}\n平台抽成 15%：¥${commission.toFixed(2)}\n您的收入 85%：¥${income.toFixed(2)}`,
      confirmButtonText: '确认完单',
      cancelButtonText: '取消',
    })
  } catch {
    return
  }

  completing.value = true
  const completedOrder = order.value
  try {
    await driverApi.orderOver(orderNo, { silent: true })
    await handleCompleteSuccess(income, orderPrice, completedOrder)
  } catch (err) {
    if (isOrderAlreadyCompletedError(err)) {
      await handleCompleteSuccess(income, orderPrice, completedOrder)
      return
    }
    if (isRequestTimeoutError(err)) {
      const state = await recoverOrderOverAfterTimeout(orderNo)
      if (state === 'completed') {
        await handleCompleteSuccess(income, orderPrice, completedOrder)
        return
      }
      showToast('完单请求超时，请稍后重试')
      return
    }
    const msg = err instanceof Error ? err.message : '完单失败'
    showToast(msg)
  } finally {
    completing.value = false
  }
}
</script>

<template>
  <div class="page">
    <van-nav-bar title="进行中订单" left-arrow @click-left="router.push('/driver')" />

    <van-empty v-if="!order && !syncing" description="暂无进行中订单，去抢单大厅看看吧">
      <van-button round type="primary" @click="router.push('/driver')">去抢单</van-button>
    </van-empty>

    <div v-else-if="order" class="page-content">
      <div class="card">
        <div class="status-banner" :class="{ onboard: canComplete }">
          {{ canComplete ? '行程进行中 · 送达后可完单' : '已接单 · 请确认乘客上车' }}
        </div>

        <p class="text-muted">订单号: {{ order.orderNo }}</p>
        <div class="route-info">
          <p><strong>起点</strong> {{ order.startAddress }}</p>
          <p><strong>终点</strong> {{ order.endAddress }}</p>
        </div>
        <div class="meta">
          <span>距离 {{ order.distance }} km</span>
          <span>订单金额 <span class="price">¥{{ incomeBreakdown.orderPrice.toFixed(2) }}</span></span>
        </div>

        <div class="income-card">
          <div class="income-row">
            <span>平台抽成 (15%)</span>
            <span class="commission">-¥{{ incomeBreakdown.commission.toFixed(2) }}</span>
          </div>
          <div class="income-row income-main">
            <span>预计收入 (85%)</span>
            <span class="income">¥{{ incomeBreakdown.income.toFixed(2) }}</span>
          </div>
          <p class="income-tip">完单后按订单原价结算，乘客优惠券不影响您的收入</p>
        </div>

        <div v-if="needStartTrip" class="start-section">
          <p class="start-title">确认乘客已上车</p>
          <p class="start-hint">请输入下单乘客手机号后四位，验证通过后即可开始行程</p>
          <van-field
            v-model="phoneTail"
            type="digit"
            maxlength="4"
            placeholder="手机号后四位"
            clearable
            class="phone-field"
          />
          <van-button
            type="primary"
            round
            block
            class="mt"
            :loading="starting"
            @click="onStartTrip"
          >
            确认上车 · 开始行程
          </van-button>
        </div>

        <van-button
          v-else
          type="primary"
          round
          block
          class="mt"
          :loading="completing"
          @click="onComplete"
        >
          完成订单
        </van-button>
      </div>
    </div>
  </div>
</template>

<style scoped>
.status-banner {
  margin: -4px 0 14px;
  padding: 10px 14px;
  border-radius: 10px;
  font-size: 13px;
  font-weight: 500;
  color: #1677ff;
  background: linear-gradient(135deg, #e6f4ff 0%, #f0f5ff 100%);
  border: 1px solid #91caff;
}
.status-banner.onboard {
  color: #389e0d;
  background: linear-gradient(135deg, #f6ffed 0%, #fcffe6 100%);
  border-color: #b7eb8f;
}
.route-info p {
  margin: 12px 0;
  font-size: 14px;
}
.meta {
  display: flex;
  justify-content: space-between;
  margin-top: 16px;
  font-size: 14px;
}
.income-card {
  margin-top: 16px;
  padding: 14px;
  background: linear-gradient(135deg, #e6f4ff 0%, #f0f5ff 100%);
  border: 1px solid #91caff;
  border-radius: 10px;
}
.income-row {
  display: flex;
  justify-content: space-between;
  font-size: 14px;
  color: #666;
  padding: 4px 0;
}
.income-row.income-main {
  margin-top: 6px;
  padding-top: 10px;
  border-top: 1px dashed #91caff;
  font-weight: 600;
  color: #333;
}
.commission {
  color: #8c8c8c;
}
.income {
  color: #1677ff;
  font-size: 20px;
}
.income-tip {
  margin-top: 10px;
  font-size: 12px;
  color: #8c8c8c;
  line-height: 1.5;
}
.start-section {
  margin-top: 20px;
  padding-top: 16px;
  border-top: 1px dashed #e8e8e8;
}
.start-title {
  font-size: 15px;
  font-weight: 600;
  color: #333;
}
.start-hint {
  margin-top: 6px;
  font-size: 12px;
  color: #8c8c8c;
  line-height: 1.5;
}
.phone-field {
  margin-top: 12px;
  border-radius: 10px;
  overflow: hidden;
  border: 1px solid #e8e8e8;
}
.mt {
  margin-top: 16px;
}
</style>
