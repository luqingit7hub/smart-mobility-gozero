<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { showSuccessToast, showToast } from 'vant'
import type { UploaderFileListItem } from 'vant'
import { userApi } from '@/api/user'
import { fileFromUploader } from '@/utils/upload'

const router = useRouter()

const realName = ref('')
const cardNo = ref('')
const nickname = ref('')
const email = ref('')
const gender = ref(1)
const avatarList = ref<UploaderFileListItem[]>([])
const loading = ref(false)

async function onSubmit() {
  const avatar = avatarList.value[0] ? fileFromUploader(avatarList.value[0]) : null
  if (!realName.value.trim() || !cardNo.value.trim()) {
    showToast('请填写真实姓名和身份证号')
    return
  }
  if (!nickname.value.trim() || !email.value.trim()) {
    showToast('请填写昵称和邮箱')
    return
  }
  if (!avatar) {
    showToast('请上传头像')
    return
  }
  if (gender.value !== 1 && gender.value !== 2) {
    showToast('请选择性别')
    return
  }

  const fd = new FormData()
  fd.append('avatar', avatar)
  fd.append('real_name', realName.value.trim())
  fd.append('card_no', cardNo.value.trim())
  fd.append('nickname', nickname.value.trim())
  fd.append('email', email.value.trim())
  fd.append('gender', String(gender.value))

  loading.value = true
  try {
    await userApi.realName(fd)
    showSuccessToast('实名认证成功')
    router.back()
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div class="page">
    <van-nav-bar title="实名认证" left-arrow @click-left="router.back()" />

    <div class="page-content">
      <div class="card upload-card">
        <p class="upload-label">头像 <span class="required">*</span></p>
        <p class="text-muted tip">上传后由服务端存储，用于个人资料展示</p>
        <van-uploader v-model="avatarList" :max-count="1" preview-size="88" />
      </div>

      <van-cell-group inset>
        <van-field v-model="realName" label="真实姓名" placeholder="与身份证一致" required />
        <van-field v-model="cardNo" label="身份证号" placeholder="18位身份证号" required />
        <van-field v-model="nickname" label="昵称" placeholder="必填" required />
        <van-field v-model="email" label="邮箱" placeholder="必填" type="email" required />
        <van-field label="性别" required>
          <template #input>
            <van-radio-group v-model="gender" direction="horizontal">
              <van-radio :name="1">男</van-radio>
              <van-radio :name="2">女</van-radio>
            </van-radio-group>
          </template>
        </van-field>
      </van-cell-group>

      <p class="hint text-muted">
        账号需处于「未实名」状态方可提交；实名信息与第三方接口校验一致后生效。
      </p>

      <div class="actions">
        <van-button type="primary" round block :loading="loading" @click="onSubmit">
          提交认证
        </van-button>
      </div>
    </div>
  </div>
</template>

<style scoped>
.upload-card {
  margin-bottom: 12px;
}
.upload-label {
  font-size: 15px;
  font-weight: 500;
  margin-bottom: 4px;
}
.required {
  color: #ff4d4f;
}
.tip {
  margin-bottom: 12px;
}
.hint {
  padding: 12px 16px 0;
  font-size: 12px;
  line-height: 1.5;
}
.actions {
  padding: 24px 16px;
}
</style>
