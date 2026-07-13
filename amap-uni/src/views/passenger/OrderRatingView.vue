<script setup lang="ts">
import OrderRatingDetail from '@/components/OrderRatingDetail.vue'
import { userApi } from '@/api/user'

async function deleteRating(orderNo: string) {
  await userApi.deleteOrderRating(orderNo)
}
</script>

<template>
  <OrderRatingDetail
    role="passenger"
    back-path="/passenger/order/list"
    :fetch-rating="(orderNo) => userApi.orderRating(orderNo, { silent: true })"
    :delete-rating="deleteRating"
  />
</template>
