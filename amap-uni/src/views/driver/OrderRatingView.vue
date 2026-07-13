<script setup lang="ts">
import OrderRatingDetail from '@/components/OrderRatingDetail.vue'
import { driverApi } from '@/api/driver'
</script>

<template>
  <OrderRatingDetail
    role="driver"
    back-path="/driver/order/list"
    :fetch-rating="(orderNo) => driverApi.orderRating(orderNo, { silent: true })"
  />
</template>
