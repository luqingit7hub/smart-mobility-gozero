<script setup lang="ts">
import OrderList from '@/components/OrderList.vue'
import { driverApi } from '@/api/driver'
</script>

<template>
  <OrderList
    title="历史订单"
    back-path="/driver/mine"
    rating-detail-path="/driver/order/rating"
    :fetch-orders="driverApi.orderList"
  />
</template>
