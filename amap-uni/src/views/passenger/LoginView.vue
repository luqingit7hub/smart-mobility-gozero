<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { closeToast, showToast } from 'vant'
import { userApi } from '@/api/user'
import { useAuthStore } from '@/stores/auth'
import { getBd09LngLat, savePassengerLocation } from '@/utils/geo'
import { formatSmsError, isValidChinaMobile, notifySmsSentSuccess } from '@/utils/phone'
import { checkRealNameAfterLogin } from '@/utils/realName'
import { showMessage } from '@/utils/toast'

const router = useRouter()
const auth = useAuthStore()

const phone = ref('')
const password = ref('')
const code = ref('')
const loginType = ref(1)
const loading = ref(false)
const countdown = ref(0)

onMounted(async () => {
  if (localStorage.getItem('role') === 'passenger' && localStorage.getItem('passenger_token')) {
    closeToast()
    await router.replace('/passenger')
    void checkRealNameAfterLogin('passenger', router)
  }
})

async function sendSms() {
  const tel = phone.value.trim()
  if (!isValidChinaMobile(tel)) {
    showMessage('请输入正确的11位手机号', 'warning')
    return
  }
  try {
    await userApi.sendSms(tel, { silent: true })
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
    const token = await userApi.login({
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
      savePassengerLocation({ lng: pos.lng, lat: pos.lat, address: '' })
    }
    auth.setPassengerAuth(token, phone.value.trim())
    closeToast()
    await router.replace('/passenger')
    void checkRealNameAfterLogin('passenger', router)
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div class="page auth-page">
    <div class="auth-hero primary-gradient">
      <span class="auth-back" @click="router.push('/')">
        <van-icon name="arrow-left" />
      </span>
      <h1>乘客登录</h1>
      <p>欢迎回来，开启安全便捷出行</p>
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

        <div class="form-actions">
          <van-button type="primary" round block :loading="loading" @click="onSubmit">登录</van-button>
          <van-button round block plain @click="router.push('/passenger/register')">
            注册账号
          </van-button>
        </div>
      </div>
    </div>
  </div>
</template>
