<script setup lang="ts">
import { useRouter } from 'vue-router'

const router = useRouter()

function enter(role: 'passenger' | 'driver') {
  const storedRole = localStorage.getItem('role')
  const token = localStorage.getItem(role === 'driver' ? 'driver_token' : 'passenger_token')
  if (storedRole === role && token) {
    router.push(role === 'driver' ? '/driver' : '/passenger')
  } else {
    router.push(role === 'driver' ? '/driver/login' : '/passenger/login')
  }
}
</script>

<template>
  <div class="home-page">
    <div class="hero primary-gradient">
      <div class="hero-badge">高德网约车</div>
      <h1>安全出行 · 智能调度</h1>
      <p>便捷支付 · AI 助手全程陪伴</p>
    </div>

    <div class="role-cards">
      <div class="role-card" @click="enter('passenger')">
        <div class="role-card-top">
          <div class="icon passenger">🚗</div>
          <div>
            <h2>我是乘客</h2>
            <p>快速叫车、优惠券、余额充值</p>
          </div>
        </div>
        <van-button type="primary" round block>进入乘客端</van-button>
      </div>

      <div class="role-card" @click="enter('driver')">
        <div class="role-card-top">
          <div class="icon driver">🚕</div>
          <div>
            <h2>我是司机</h2>
            <p>抢单接单、实名认证、完成订单</p>
          </div>
        </div>
        <van-button type="primary" round block plain>进入司机端</van-button>
      </div>
    </div>

    <p class="footer text-muted">路钦个人测试开发 · 网关 127.0.0.1:8888</p>
  </div>
</template>

<style scoped>
.home-page {
  min-height: 100vh;
  background: var(--app-bg);
}
.hero {
  padding: 52px 24px 64px;
  text-align: center;
  border-radius: 0 0 28px 28px;
}
.hero-badge {
  display: inline-block;
  padding: 4px 12px;
  border-radius: 999px;
  background: rgba(255, 255, 255, 0.2);
  font-size: 12px;
  margin-bottom: 14px;
}
.hero h1 {
  font-size: 26px;
  font-weight: 700;
  margin-bottom: 8px;
}
.hero p {
  opacity: 0.92;
  font-size: 14px;
}
.role-cards {
  padding: 0 16px 24px;
  display: flex;
  flex-direction: column;
  gap: 14px;
  margin-top: -36px;
}
.role-card {
  background: #fff;
  border-radius: var(--app-radius-lg);
  padding: 20px;
  border: 1px solid var(--app-border);
  box-shadow: var(--app-shadow-md);
  cursor: pointer;
  transition: transform 0.15s ease;
}
.role-card:active {
  transform: scale(0.99);
}
.role-card-top {
  display: flex;
  gap: 14px;
  align-items: flex-start;
  margin-bottom: 16px;
}
.role-card h2 {
  font-size: 18px;
  margin-bottom: 4px;
}
.role-card p {
  color: var(--app-text-muted);
  font-size: 13px;
}
.icon {
  width: 52px;
  height: 52px;
  border-radius: 14px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 26px;
  flex-shrink: 0;
}
.icon.passenger {
  background: var(--app-primary-bg);
}
.icon.driver {
  background: #fff7e6;
}
.footer {
  text-align: center;
  padding: 8px 24px 28px;
  font-size: 12px;
}
</style>
