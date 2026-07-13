<script setup lang="ts">
import { ref, watch } from 'vue'
import { useRouter } from 'vue-router'
import type { OrderListItem } from '@/types'
import {
  ORDER_STATUS_TABS,
  buildOrderRatingQuery,
  normalizeOrderList,
  orderStatusTagType,
  pickOrderListNo,
} from '@/utils/orderList'
import { shortenOrderNo } from '@/utils/order'

const props = defineProps<{
  title?: string
  backPath: string
  ratingDetailPath: string
  fetchOrders: (params: {
    page: number
    page_size: number
    order_no?: string
    status?: number
  }) => Promise<{ list?: OrderListItem[]; total?: number }>
}>()

const router = useRouter()

const orders = ref<OrderListItem[]>([])
const loading = ref(false)
const finished = ref(false)
const page = ref(1)
const total = ref(0)
const orderNo = ref('')
const searchNo = ref('')
const status = ref(0)
const pageSize = 10

async function loadPage() {
  if (finished.value) return
  loading.value = true
  try {
    const res = await props.fetchOrders({
      page: page.value,
      page_size: pageSize,
      order_no: orderNo.value || undefined,
      status: status.value || undefined,
    })
    const batch = normalizeOrderList(res.list || [])
    total.value = Number(res.total ?? 0)
    orders.value.push(...batch)
    if (orders.value.length >= total.value || batch.length < pageSize) {
      finished.value = true
    } else {
      page.value += 1
    }
  } catch {
    finished.value = true
  } finally {
    loading.value = false
  }
}

function resetAndReload() {
  page.value = 1
  orders.value = []
  finished.value = false
  void loadPage()
}

function onSearch() {
  orderNo.value = searchNo.value.trim()
  resetAndReload()
}

function onClearSearch() {
  searchNo.value = ''
  orderNo.value = ''
  resetAndReload()
}

watch(status, (val, oldVal) => {
  if (val !== oldVal) resetAndReload()
})

function goRatingDetail(item: OrderListItem) {
  router.push({
    path: props.ratingDetailPath,
    query: buildOrderRatingQuery(item),
  })
}
</script>

<template>
  <div class="page">
    <van-nav-bar :title="title || '历史订单'" left-arrow @click-left="$router.push(backPath)" />

    <van-tabs v-model:active="status" shrink sticky offset-top="46">
      <van-tab v-for="tab in ORDER_STATUS_TABS" :key="tab.value" :title="tab.label" :name="tab.value" />
    </van-tabs>

    <div class="search-bar">
      <van-search v-model="searchNo" placeholder="搜索订单号" show-action @search="onSearch">
        <template #action>
          <div class="search-actions">
            <span @click="onSearch">搜索</span>
            <span v-if="orderNo" class="clear" @click="onClearSearch">清除</span>
          </div>
        </template>
      </van-search>
    </div>

    <div v-if="total > 0" class="page-summary">共 {{ total }} 条，已加载 {{ orders.length }} 条</div>

    <van-list
      v-model:loading="loading"
      :finished="finished"
      finished-text="没有更多了"
      @load="loadPage"
    >
      <van-empty v-if="!loading && !orders.length" description="暂无订单记录" />
      <div v-for="item in orders" :key="item.id" class="order-card card">
        <div class="order-head">
          <div>
            <p class="order-no">{{ shortenOrderNo(pickOrderListNo(item)) }}</p>
            <p class="order-time">{{ item.createdAt }}</p>
          </div>
          <van-tag :type="orderStatusTagType(item.status)" plain>{{ item.statusName }}</van-tag>
        </div>

        <div class="order-route">
          <p><van-icon name="location-o" /> {{ item.startAddress }}</p>
          <p><van-icon name="aim" /> {{ item.endAddress }}</p>
        </div>

        <div class="order-meta">
          <span v-if="item.distance"> {{ item.distance }} km</span>
          <span v-if="item.duration"> · {{ item.duration }} 分钟</span>
          <span class="price">¥{{ item.price?.toFixed(2) }}</span>
        </div>

        <div class="order-actions">
          <van-button size="small" type="primary" plain round @click="goRatingDetail(item)">
            评价记录
          </van-button>
        </div>
      </div>
    </van-list>
  </div>
</template>

<style scoped>
.search-bar {
  background: var(--van-background-2, #f7f8fa);
}
.page-summary {
  padding: 8px 16px 0;
  font-size: 12px;
  color: var(--van-text-color-3);
}
.search-actions {
  display: flex;
  gap: 12px;
  color: var(--van-primary-color);
}
.search-actions .clear {
  color: var(--van-text-color-2);
}
.order-card {
  margin: 12px 16px;
  padding: 14px 16px;
}
.order-head {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  gap: 12px;
}
.order-no {
  font-size: 15px;
  font-weight: 600;
}
.order-time {
  margin-top: 4px;
  font-size: 12px;
  color: var(--van-text-color-3);
}
.order-route {
  margin-top: 12px;
  font-size: 13px;
  color: var(--van-text-color-2);
  line-height: 1.7;
}
.order-route .van-icon {
  margin-right: 4px;
  color: var(--van-primary-color);
}
.order-meta {
  margin-top: 10px;
  font-size: 12px;
  color: var(--van-text-color-3);
}
.order-meta .price {
  margin-left: 8px;
  font-size: 16px;
  font-weight: 700;
  color: #ff6b00;
}
.order-actions {
  margin-top: 12px;
  display: flex;
  justify-content: flex-end;
}
</style>
