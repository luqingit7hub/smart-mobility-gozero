<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { showToast } from 'vant'
import { showAppNotice } from '@/utils/dialog'
import { driverApi } from '@/api/driver'

const balance = ref(0)
const balanceLoading = ref(true)
const balanceError = ref('')

const balanceDisplay = computed(() => balance.value.toFixed(2))

function onWithdraw() {
  showToast('请联系管理员进行提现')
}

async function loadBalance() {
  balanceLoading.value = true
  balanceError.value = ''
  try {
    const res = await driverApi.walletBalance({ silent: true })
    balance.value = res.balance ?? 0
  } catch (e) {
    balanceError.value = e instanceof Error ? e.message : '余额加载失败'
  } finally {
    balanceLoading.value = false
  }
}

onMounted(() => {
  void loadBalance()
  void showAppNotice('温馨提示', '由于测试通道不稳定，可联系管理员进行余额提现')
})
</script>

<template>
  <div class="page wallet-page">
    <van-nav-bar title="我的钱包" left-arrow @click-left="$router.push('/driver/mine')" />

    <div class="page-content">
      <section class="balance-card driver-gradient">
        <div class="balance-head">
          <van-icon name="balance-o" class="balance-icon" />
          <span>可提现余额（元）</span>
        </div>

        <div class="balance-amount-row">
          <van-loading v-if="balanceLoading" size="28" color="#fff" />
          <template v-else>
            <span class="currency">¥</span>
            <span class="amount">{{ balanceDisplay }}</span>
          </template>
        </div>

        <p v-if="balanceError" class="balance-tip error">{{ balanceError }}，请点击刷新</p>
        <p v-else class="balance-tip">完单后收入将计入账户余额</p>

        <div class="balance-actions">
          <button type="button" class="action-btn ghost" :disabled="balanceLoading" @click="loadBalance">
            刷新余额
          </button>
          <button type="button" class="action-btn solid" @click="onWithdraw">金额提现</button>
        </div>
      </section>

      <section class="detail-card">
        <div class="section-head">
          <h3>账户明细</h3>
          <router-link class="logs-link" to="/driver/wallet/logs">流水明细</router-link>
        </div>
        <p class="detail-hint">可在流水明细中查看每笔收入与提现记录</p>
      </section>

      <section class="income-tips">
        <div class="tips-head">
          <van-icon name="info-o" class="tips-icon" />
          <h3>收入说明</h3>
        </div>
        <ul class="tips-list">
          <li>完单后，订单实际支付金额的 <strong>85%</strong> 计入您的可提现余额。</li>
          <li>平台收取 <strong>15%</strong> 作为信息服务费，用于订单撮合、系统运维等服务。</li>
          <li>乘客使用优惠券的订单，优惠差额由平台补贴，不影响您按完单金额分成。</li>
          <li>收入到账后可在「流水明细」查看每笔订单的分账记录。</li>
          <li>提现需联系管理员处理，测试环境通道可能不稳定。</li>
        </ul>
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
  box-shadow: 0 10px 28px rgba(9, 88, 217, 0.28);
  color: #fff;
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

.detail-card {
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
  margin-bottom: 8px;
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

.detail-hint {
  font-size: 13px;
  color: var(--app-text-secondary, #6b7280);
  line-height: 1.5;
}

.income-tips {
  margin-top: 12px;
  padding: 16px;
  border-radius: var(--app-radius-md, 14px);
  background: #fff;
  border: 1px solid var(--app-border, #e8edf5);
  box-shadow: var(--app-shadow-sm, 0 2px 10px rgba(15, 23, 42, 0.06));
}

.tips-head {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 12px;
}

.tips-icon {
  font-size: 18px;
  color: var(--app-primary, #1677ff);
}

.tips-head h3 {
  font-size: 15px;
  font-weight: 600;
  color: var(--app-text, #1f2937);
}

.tips-list {
  margin: 0;
  padding-left: 18px;
  color: var(--app-text-secondary, #6b7280);
  font-size: 13px;
  line-height: 1.7;
}

.tips-list li + li {
  margin-top: 8px;
}

.tips-list strong {
  color: var(--app-primary, #1677ff);
  font-weight: 600;
}
</style>
