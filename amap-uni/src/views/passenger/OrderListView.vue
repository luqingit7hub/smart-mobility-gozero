<script setup lang="ts">
import OrderList from '@/components/OrderList.vue'
import { userApi } from '@/api/user'
</script>

<template>
  <OrderList
    title="历史订单"
    back-path="/passenger/mine"
    rating-detail-path="/passenger/order/rating"
    :fetch-orders="userApi.orderList"
  />
</template>
