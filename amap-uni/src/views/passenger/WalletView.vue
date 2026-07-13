<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { showSuccessToast, showToast } from 'vant'
import { showAppNotice } from '@/utils/dialog'
import { userApi } from '@/api/user'

const amount = ref('50')
const balance = ref(0)
const loading = ref(false)
const balanceLoading = ref(true)
const balanceError = ref('')
const presets = [20, 50, 100, 200]

const balanceDisplay = computed(() => balance.value.toFixed(2))

async function loadBalance() {
  balanceLoading.value = true
  balanceError.value = ''
  try {
    const res = await userApi.walletBalance({ silent: true })
    balance.value = res.balance ?? 0
  } catch (e) {
    balanceError.value = e instanceof Error ? e.message : '余额加载失败'
  } finally {
    balanceLoading.value = false
  }
}

function onWithdraw() {
  showToast('请联系管理员进行提现')
}

async function onRecharge() {
  const money = parseFloat(amount.value)
  if (!money || money <= 0) {
    showToast('请输入有效金额')
    return
  }
  loading.value = true
  try {
    const res = await userApi.recharge(money)
    const url = res.alipayUrl || res.alipay_url
    if (url) {
      window.open(url, '_blank')
      showSuccessToast('已跳转支付宝')
    }
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  void loadBalance()
  void showAppNotice('温馨提示', '由于测试通道不稳定，可联系管理员进行余额充值')
})
</script>

<template>
  <div class="page wallet-page">
    <van-nav-bar title="我的钱包" left-arrow @click-left="$router.push('/passenger/mine')" />

    <div class="page-content">
      <section class="balance-card primary-gradient">
        <div class="balance-head">
          <van-icon name="balance-o" class="balance-icon" />
          <span>账户余额（元）</span>
        </div>

        <div class="balance-amount-row">
          <van-loading v-if="balanceLoading" size="28" color="#fff" />
          <template v-else>
            <span class="currency">¥</span>
            <span class="amount">{{ balanceDisplay }}</span>
          </template>
        </div>

        <p v-if="balanceError" class="balance-tip error">{{ balanceError }}，请点击刷新</p>
        <p v-else class="balance-tip">余额可用于叫车支付，充值后刷新即可到账</p>

        <div class="balance-actions">
          <button type="button" class="action-btn ghost" :disabled="balanceLoading" @click="loadBalance">
            刷新余额
          </button>
          <button type="button" class="action-btn solid" @click="onWithdraw">金额提现</button>
        </div>
      </section>

      <section class="recharge-card">
        <div class="section-head">
          <h3>账户充值</h3>
          <router-link class="logs-link" to="/passenger/wallet/logs">流水明细</router-link>
        </div>

        <div class="amount-field">
          <label for="recharge-amount">充值金额</label>
          <div class="amount-input-wrap">
            <span class="prefix">¥</span>
            <input
              id="recharge-amount"
              v-model="amount"
              type="number"
              inputmode="decimal"
              placeholder="请输入充值金额"
            />
          </div>
        </div>

        <div class="presets">
          <button
            v-for="p in presets"
            :key="p"
            type="button"
            class="preset-btn"
            :class="{ active: amount === String(p) }"
            @click="amount = String(p)"
          >
            ¥{{ p }}
          </button>
        </div>

        <van-button type="primary" round block class="recharge-btn" :loading="loading" @click="onRecharge">
          支付宝充值
        </van-button>
      </section>
    </div>
  </div>
</template>

<style scoped>
.wallet-page .page-content {
  padding-top: 8px;
}

.balance-card {
  padding: 22px 18px 18px;
  border-radius: var(--app-radius-md, 14px);
  margin-bottom: 12px;
  box-shadow: 0 10px 28px rgba(22, 119, 255, 0.28);
}

.balance-head {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 6px;
  font-size: 14px;
  opacity: 0.92;
}

.balance-icon {
  font-size: 18px;
}

.balance-amount-row {
  display: flex;
  align-items: baseline;
  justify-content: center;
  gap: 4px;
  min-height: 56px;
  margin: 10px 0 6px;
}

.currency {
  font-size: 22px;
  font-weight: 600;
  opacity: 0.95;
}

.amount {
  font-size: 42px;
  font-weight: 700;
  letter-spacing: 0.5px;
  line-height: 1;
}

.balance-tip {
  text-align: center;
  font-size: 12px;
  opacity: 0.88;
  line-height: 1.5;
  margin-bottom: 16px;
}

.balance-tip.error {
  opacity: 1;
  color: #fff7e6;
}

.balance-actions {
  display: flex;
  gap: 10px;
  justify-content: center;
}

.action-btn {
  min-width: 108px;
  height: 34px;
  padding: 0 14px;
  border-radius: 999px;
  font-size: 13px;
  cursor: pointer;
}

.action-btn.ghost {
  border: 1px solid rgba(255, 255, 255, 0.55);
  background: rgba(255, 255, 255, 0.14);
  color: #fff;
}

.action-btn.solid {
  border: 1px solid rgba(255, 255, 255, 0.65);
  background: rgba(255, 255, 255, 0.28);
  color: #fff;
  font-weight: 600;
}

.action-btn:disabled {
  opacity: 0.7;
  cursor: not-allowed;
}

.recharge-card {
  background: var(--app-card, #fff);
  border-radius: var(--app-radius-md, 14px);
  padding: 16px;
  border: 1px solid var(--app-border, #e8edf5);
  box-shadow: var(--app-shadow-sm, 0 2px 10px rgba(15, 23, 42, 0.06));
}

.section-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 14px;
}

.section-head h3 {
  font-size: 16px;
  font-weight: 600;
}

.logs-link {
  font-size: 13px;
  color: var(--app-primary, #1677ff);
  text-decoration: none;
}

.amount-field label {
  display: block;
  font-size: 13px;
  color: var(--app-text-secondary, #6b7280);
  margin-bottom: 8px;
}

.amount-input-wrap {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 0 14px;
  height: 46px;
  border-radius: 12px;
  background: var(--app-bg, #eef2f8);
  border: 1px solid transparent;
}

.amount-input-wrap:focus-within {
  border-color: rgba(22, 119, 255, 0.35);
  background: #fff;
}

.prefix {
  font-size: 18px;
  font-weight: 600;
  color: var(--app-primary, #1677ff);
}

.amount-input-wrap input {
  flex: 1;
  border: none;
  outline: none;
  background: transparent;
  font-size: 18px;
  font-weight: 600;
  color: var(--app-text, #1f2937);
}

.amount-input-wrap input::placeholder {
  font-size: 14px;
  font-weight: 400;
  color: var(--app-text-muted, #9ca3af);
}

.presets {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
  margin: 14px 0 16px;
}

.preset-btn {
  flex: 1;
  min-width: 72px;
  height: 36px;
  border-radius: 10px;
  border: 1px solid rgba(22, 119, 255, 0.25);
  background: #fff;
  color: var(--app-primary, #1677ff);
  font-size: 14px;
  cursor: pointer;
}

.preset-btn.active {
  background: var(--app-primary-bg, #e6f4ff);
  border-color: var(--app-primary, #1677ff);
  font-weight: 600;
}

.recharge-btn {
  height: 44px;
  font-weight: 600;
  box-shadow: 0 4px 12px rgba(22, 119, 255, 0.28);
}
</style>
