/** 与 rpcOrder orderoverlogic.go 中 driverIncomeRate / platformIncomeRate 保持一致 */
export const DRIVER_INCOME_RATE = 0.85
export const PLATFORM_COMMISSION_RATE = 0.15

export interface DriverIncomeBreakdown {
  orderPrice: number
  income: number
  commission: number
}

export function calcDriverIncome(orderPrice: number): DriverIncomeBreakdown {
  const price = orderPrice > 0 ? orderPrice : 0
  return {
    orderPrice: price,
    income: Number((price * DRIVER_INCOME_RATE).toFixed(2)),
    commission: Number((price * PLATFORM_COMMISSION_RATE).toFixed(2)),
  }
}

export function formatDriverCompleteMessage(orderPrice: number): string {
  const { income, commission } = calcDriverIncome(orderPrice)
  return `订单已完成！订单金额 ¥${orderPrice.toFixed(2)}，平台抽成 15%（¥${commission.toFixed(2)}），您的收入 85%（¥${income.toFixed(2)}）已入账`
}
