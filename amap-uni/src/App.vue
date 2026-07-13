<script setup lang="ts">
const themeVars = {
  primaryColor: '#1677ff',
  buttonPrimaryBackground: '#1677ff',
  buttonPrimaryBorderColor: '#1677ff',
  buttonBorderRadius: '10px',
  cellGroupInsetPadding: '16px',
  radiusLg: '14px',
}
</script>

<template>
  <van-config-provider :theme-vars="themeVars">
    <router-view />
  </van-config-provider>
</template>
