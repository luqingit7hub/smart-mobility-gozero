<script setup lang="ts">
import WalletLogsList from '@/components/WalletLogsList.vue'
import { driverApi } from '@/api/driver'
</script>

<template>
  <WalletLogsList
    title="流水明细"
    back-path="/driver/wallet"
    :fetch-logs="driverApi.walletLogs"
  />
</template>
