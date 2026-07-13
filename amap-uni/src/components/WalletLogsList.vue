<script setup lang="ts">
import { ref } from 'vue'
import type { WalletLogItem } from '@/types'
import { formatWalletAmount, normalizeWalletLogs } from '@/utils/wallet'

const props = defineProps<{
  title?: string
  backPath: string
  fetchLogs: (params: { page: number; page_size: number; order_no?: string }) => Promise<{
    list?: WalletLogItem[]
    total?: number
  }>
}>()

const logs = ref<WalletLogItem[]>([])
const loading = ref(false)
const finished = ref(false)
const page = ref(1)
const total = ref(0)
const orderNo = ref('')
const searchNo = ref('')
const pageSize = 10

async function loadPage() {
  if (finished.value) return
  loading.value = true
  try {
    const res = await props.fetchLogs({
      page: page.value,
      page_size: pageSize,
      order_no: orderNo.value || undefined,
    })
    const batch = normalizeWalletLogs(res.list || [])
    total.value = Number(res.total ?? 0)
    logs.value.push(...batch)
    if (logs.value.length >= total.value || batch.length < pageSize) {
      finished.value = true
    } else {
      page.value += 1
    }
  } catch {
    // 请求失败时结束 loading，避免 van-list 一直转圈
    finished.value = true
  } finally {
    loading.value = false
  }
}

function resetAndReload() {
  page.value = 1
  logs.value = []
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
</script>

<template>
  <div class="page">
    <van-nav-bar :title="title || '流水明细'" left-arrow @click-left="$router.push(backPath)" />
    <div class="search-bar">
      <van-search
        v-model="searchNo"
        placeholder="订单号 / 流水单号"
        show-action
        @search="onSearch"
      >
        <template #action>
          <div class="search-actions">
            <span @click="onSearch">搜索</span>
            <span v-if="orderNo" class="clear" @click="onClearSearch">清除</span>
          </div>
        </template>
      </van-search>
    </div>

    <div v-if="total > 0" class="page-summary">共 {{ total }} 条，已加载 {{ logs.length }} 条</div>

    <van-list
      v-model:loading="loading"
      :finished="finished"
      finished-text="没有更多了"
      @load="loadPage"
    >
      <van-empty v-if="!loading && !logs.length" description="暂无流水记录" />
      <div v-for="item in logs" :key="item.id" class="log-card card">
        <div class="log-head">
          <div>
            <p class="log-type">{{ item.typeName }}</p>
            <p class="log-time">{{ item.createdAt }}</p>
          </div>
          <p class="log-amount" :class="item.direction === 'out' ? 'out' : 'in'">
            {{ formatWalletAmount(item.signedAmount ?? 0) }}
          </p>
        </div>
        <div class="log-meta">
          <p v-if="item.orderNo">单号: {{ item.orderNo }}</p>
          <p v-if="item.balanceAfter != null">余额: ¥{{ item.balanceAfter.toFixed(2) }}</p>
          <p v-if="item.status === 2" class="pending">待支付</p>
          <p v-if="item.remark" class="remark">{{ item.remark }}</p>
        </div>
      </div>
    </van-list>
  </div>
</template>

<style scoped>
.search-bar {
  position: sticky;
  top: 0;
  z-index: 2;
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
.log-card {
  margin: 12px 16px;
  padding: 14px 16px;
}
.log-head {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  gap: 12px;
}
.log-type {
  font-size: 15px;
  font-weight: 600;
}
.log-time {
  margin-top: 4px;
  font-size: 12px;
  color: var(--van-text-color-3);
}
.log-amount {
  font-size: 18px;
  font-weight: 700;
  white-space: nowrap;
}
.log-amount.in {
  color: #07c160;
}
.log-amount.out {
  color: #ee0a24;
}
.log-meta {
  margin-top: 10px;
  font-size: 12px;
  color: var(--van-text-color-2);
  line-height: 1.6;
}
.pending {
  color: #ff976a;
}
.remark {
  color: var(--van-text-color-3);
}
</style>
