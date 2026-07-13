<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { closeToast, showToast } from 'vant'
import { driverApi } from '@/api/driver'
import { useAuthStore } from '@/stores/auth'
import { getBd09LngLat, saveDriverLocation } from '@/utils/geo'
import { syncActiveDriverTripFromServer } from '@/utils/driverTrip'
import { formatSmsError, isValidChinaMobile, notifySmsSentSuccess } from '@/utils/phone'
import { checkRealNameAfterLogin } from '@/utils/realName'
import { showMessage } from '@/utils/toast'

const router = useRouter()
const auth = useAuthStore()

const phone = ref('')
const password = ref('')
const code = ref('')
const loginType = ref(2)
const loading = ref(false)
const countdown = ref(0)

onMounted(async () => {
  if (localStorage.getItem('role') === 'driver' && localStorage.getItem('driver_token')) {
    closeToast()
    await router.replace('/driver')
    void checkRealNameAfterLogin('driver', router)
  }
})

async function sendSms() {
  const tel = phone.value.trim()
  if (!isValidChinaMobile(tel)) {
    showMessage('请输入正确的11位手机号', 'warning')
    return
  }
  try {
    await driverApi.sendSms(tel, { silent: true })
    notifySmsSentSuccess()
    countdown.value = 60
    const timer = setInterval(() => {
      countdown.value--
      if (countdown.value <= 0) clearInterval(timer)
    }, 1000)
  } catch (err) {
    showMessage(formatSmsError(err), 'danger')
  }
}

async function onSubmit() {
  const type = Number(loginType.value)
  if (!phone.value.trim()) {
    showToast('请输入手机号')
    return
  }
  if (type === 1 && !code.value.trim()) {
    showToast('请输入验证码')
    return
  }
  if (type === 2 && !password.value) {
    showToast('请输入密码')
    return
  }
  loading.value = true
  try {
    const pos = await getBd09LngLat()
    if (!pos) showToast('未获取到定位，请登录后在抢单页设置接单位置')
    const token = await driverApi.login({
      phone: phone.value.trim(),
      type,
      password: type === 2 ? password.value : undefined,
      code: type === 1 ? code.value.trim() : undefined,
      lng: pos?.lng,
      lat: pos?.lat,
    })
    if (!token?.trim()) {
      showToast('登录失败，未获取到 token')
      return
    }
    if (pos) {
      saveDriverLocation({ lng: pos.lng, lat: pos.lat, address: '' })
    }
    auth.setDriverAuth(token, phone.value.trim())
    closeToast()
    await syncActiveDriverTripFromServer()
    await router.replace('/driver')
    void checkRealNameAfterLogin('driver', router)
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div class="page auth-page">
    <div class="auth-hero driver-gradient">
      <span class="auth-back" @click="router.push('/')">
        <van-icon name="arrow-left" />
      </span>
      <h1>司机登录</h1>
      <p>登录即上线，接收附近订单</p>
    </div>
    <div class="page-content auth-content">
      <div class="auth-panel">
        <div class="auth-segment">
          <button
            type="button"
            class="auth-segment-btn"
            :class="{ active: loginType === 1 }"
            @click="loginType = 1"
          >
            验证码登录
          </button>
          <button
            type="button"
            class="auth-segment-btn"
            :class="{ active: loginType === 2 }"
            @click="loginType = 2"
          >
            密码登录
          </button>
        </div>

        <van-cell-group inset class="form-group">
          <van-field v-model="phone" label="手机号" placeholder="请输入手机号" type="tel" />
          <van-field v-if="loginType === 1" v-model="code" label="验证码" placeholder="请输入验证码">
            <template #button>
              <van-button size="small" type="primary" :disabled="countdown > 0" @click="sendSms">
                {{ countdown > 0 ? `${countdown}s` : '获取验证码' }}
              </van-button>
            </template>
          </van-field>
          <van-field v-else v-model="password" label="密码" type="password" placeholder="请输入密码" />
        </van-cell-group>

        <p class="auth-tip">登录即上线，请允许浏览器定位以便接收附近订单</p>

        <div class="form-actions">
          <van-button type="primary" round block :loading="loading" @click="onSubmit">登录并上线</van-button>
          <van-button round block plain @click="router.push('/driver/register')">
            注册成为司机
          </van-button>
        </div>
      </div>
    </div>
  </div>
</template>
