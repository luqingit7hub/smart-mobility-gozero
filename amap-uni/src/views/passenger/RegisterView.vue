<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { showToast, showSuccessToast } from 'vant'
import { userApi } from '@/api/user'
import { formatSmsError, isValidChinaMobile, notifySmsSentSuccess } from '@/utils/phone'
import { showMessage } from '@/utils/toast'

const router = useRouter()

const phone = ref('')
const password = ref('')
const code = ref('')
const loading = ref(false)
const countdown = ref(0)

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
  if (!phone.value.trim() || !password.value.trim() || !code.value.trim()) {
    showToast('请填写完整信息')
    return
  }
  loading.value = true
  try {
    await userApi.register({
      phone: phone.value.trim(),
      password: password.value,
      code: code.value.trim(),
    })
    showSuccessToast('注册成功，请登录')
    await router.replace('/passenger/login')
  } catch (err) {
    const msg = err instanceof Error ? err.message : '注册失败'
    showMessage(msg, 'danger')
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div class="page auth-page">
    <div class="auth-hero primary-gradient">
      <span class="auth-back" @click="router.back()">
        <van-icon name="arrow-left" />
      </span>
      <h1>乘客注册</h1>
      <p>创建账号，享受便捷出行服务</p>
    </div>
    <div class="page-content auth-content">
      <van-cell-group inset class="form-group">
        <van-field v-model="phone" label="手机号" placeholder="请输入手机号" type="tel" />
        <van-field v-model="password" label="密码" type="password" placeholder="请输入密码" />
        <van-field v-model="code" label="验证码" placeholder="请输入验证码">
          <template #button>
            <van-button size="small" type="primary" :disabled="countdown > 0" @click="sendSms">
              {{ countdown > 0 ? `${countdown}s` : '获取验证码' }}
            </van-button>
          </template>
        </van-field>
      </van-cell-group>
      <div class="form-actions">
        <van-button type="primary" round block :loading="loading" @click="onSubmit">注册</van-button>
      </div>
    </div>
  </div>
</template>

<style scoped>
.auth-content {
  margin-top: -12px;
  padding-top: 0;
}
.form-group {
  margin-top: 8px;
}
</style>
