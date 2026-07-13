<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { showSuccessToast, showToast } from 'vant'
import type { UploaderFileListItem } from 'vant'
import { driverApi } from '@/api/driver'
import { fileFromUploader } from '@/utils/upload'

const router = useRouter()

const realName = ref('')
const cardNo = ref('')
const email = ref('')
const carNumber = ref('')
const carType = ref('')
const carColor = ref('')
const avatarList = ref<UploaderFileListItem[]>([])
const licenseList = ref<UploaderFileListItem[]>([])
const vehicleList = ref<UploaderFileListItem[]>([])
const loading = ref(false)

async function onSubmit() {
  const avatar = avatarList.value[0] ? fileFromUploader(avatarList.value[0]) : null
  const license = licenseList.value[0] ? fileFromUploader(licenseList.value[0]) : null
  const vehicle = vehicleList.value[0] ? fileFromUploader(vehicleList.value[0]) : null

  if (!realName.value.trim() || !cardNo.value.trim() || !email.value.trim()) {
    showToast('请填写姓名、身份证和邮箱')
    return
  }
  if (!carNumber.value.trim() || !carType.value.trim() || !carColor.value.trim()) {
    showToast('请填写完整车辆信息')
    return
  }
  if (!avatar || !license || !vehicle) {
    showToast('请上传头像、驾驶证和行驶证照片')
    return
  }

  const fd = new FormData()
  fd.append('avatar', avatar)
  fd.append('license_photo', license)
  fd.append('vehicle_photo', vehicle)
  fd.append('real_name', realName.value.trim())
  fd.append('card_no', cardNo.value.trim())
  fd.append('email', email.value.trim())
  fd.append('car_number', carNumber.value.trim())
  fd.append('car_type', carType.value.trim())
  fd.append('car_color', carColor.value.trim())

  loading.value = true
  try {
    await driverApi.realName(fd)
    showSuccessToast('认证提交成功')
    router.back()
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div class="page">
    <van-nav-bar title="资质认证" left-arrow @click-left="$router.push('/driver/mine')" />

    <div class="page-content">
      <div class="card">
        <h3>司机实名与车辆信息</h3>
        <p class="text-muted intro">需上传头像、驾驶证、行驶证照片（multipart 提交至网关）</p>

        <div class="upload-block">
          <p class="upload-title">头像 <span class="required">*</span></p>
          <van-uploader v-model="avatarList" :max-count="1" preview-size="80" />
        </div>
        <div class="upload-block">
          <p class="upload-title">驾驶证照片 <span class="required">*</span></p>
          <van-uploader v-model="licenseList" :max-count="1" preview-size="80" />
        </div>
        <div class="upload-block">
          <p class="upload-title">行驶证照片 <span class="required">*</span></p>
          <van-uploader v-model="vehicleList" :max-count="1" preview-size="80" />
        </div>

        <van-field v-model="realName" label="真实姓名" placeholder="与身份证一致" />
        <van-field v-model="cardNo" label="身份证" placeholder="18位身份证号" />
        <van-field v-model="email" label="邮箱" placeholder="必填" type="email" />
        <van-field v-model="carNumber" label="车牌号" placeholder="苏A12345" />
        <van-field v-model="carType" label="车型" placeholder="经济型" />
        <van-field v-model="carColor" label="车身颜色" placeholder="白色" />

        <van-button type="primary" round block class="mt" :loading="loading" @click="onSubmit">
          提交认证
        </van-button>
      </div>
    </div>
  </div>
</template>

<style scoped>
h3 {
  margin-bottom: 8px;
  font-size: 16px;
}
.intro {
  margin-bottom: 16px;
}
.upload-block {
  margin-bottom: 16px;
}
.upload-title {
  font-size: 14px;
  margin-bottom: 8px;
  font-weight: 500;
}
.required {
  color: #ff4d4f;
}
.mt {
  margin-top: 20px;
}
</style>
