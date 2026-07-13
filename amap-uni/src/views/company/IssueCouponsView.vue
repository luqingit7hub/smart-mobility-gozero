<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { showSuccessToast, showToast } from 'vant'
import { mapApi } from '@/api/driver'

const router = useRouter()

const address = ref('')
const couponType = ref(1)
const moneyQuan = ref('5')
const discount = ref('8.5')
const outTime = ref('2029-12-31 23:59:59')
const loading = ref(false)

const typeOptions = [
  { label: '现金券', value: 1 },
  { label: '折扣券', value: 2 },
  { label: '免费乘车', value: 3 },
]

async function onSubmit() {
  if (!address.value.trim() || !outTime.value.trim()) {
    showToast('请填写地区和过期时间')
    return
  }
  loading.value = true
  try {
    const payload: Parameters<typeof mapApi.issueCoupons>[0] = {
      address: address.value.trim(),
      type: couponType.value,
      out_time: outTime.value.trim(),
    }
    if (couponType.value === 1) payload.money_quan = Number(moneyQuan.value)
    if (couponType.value === 2) payload.discount = Number(discount.value)
    const res = await mapApi.issueCoupons(payload)
    const count = res.issued_count ?? res.issuedCount ?? 0
    showSuccessToast(`成功发放 ${count} 张`)
    router.back()
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div class="page">
    <van-nav-bar title="发放优惠券" left-arrow @click-left="router.back()" />
    <div class="page-content">
      <div class="tip card">公司账户专用 · 按地区向同城用户发券</div>

      <van-cell-group inset>
        <van-field v-model="address" label="目标地区" placeholder="如：宿豫区" />
        <van-field label="券类型">
          <template #input>
            <van-radio-group v-model="couponType" direction="horizontal">
              <van-radio v-for="o in typeOptions" :key="o.value" :name="o.value">
                {{ o.label }}
              </van-radio>
            </van-radio-group>
          </template>
        </van-field>
        <van-field
          v-if="couponType === 1"
          v-model="moneyQuan"
          type="number"
          label="面额(元)"
        />
        <van-field
          v-if="couponType === 2"
          v-model="discount"
          type="number"
          label="折扣"
          placeholder="如 8.5"
        />
        <van-field v-model="outTime" label="过期时间" placeholder="2029-01-02 15:04:05" />
      </van-cell-group>

      <div class="actions">
        <van-button type="primary" round block :loading="loading" @click="onSubmit">
          发放优惠券
        </van-button>
      </div>
    </div>
  </div>
</template>

<style scoped>
.tip {
  background: #fff7e6;
  color: #d46b08;
  font-size: 13px;
  margin-bottom: 12px;
}
.actions {
  padding: 24px 16px;
}
</style>
