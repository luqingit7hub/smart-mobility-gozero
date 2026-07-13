<script setup lang="ts">
import WalletLogsList from '@/components/WalletLogsList.vue'
import { userApi } from '@/api/user'
</script>

<template>
  <WalletLogsList
    title="流水明细"
    back-path="/passenger/wallet"
    :fetch-logs="userApi.walletLogs"
  />
</template>
