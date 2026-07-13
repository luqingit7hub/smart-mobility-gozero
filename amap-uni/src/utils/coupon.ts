import type { CouponItem } from '@/types'

export const couponTypeLabels: Record<number, string> = {
  1: '现金券',
  2: '折扣券',
  3: '免费乘车券',
}

export function normalizeCoupon(raw: Record<string, unknown>): CouponItem {
  return {
    id: Number(raw.id ?? 0),
    type: Number(raw.type ?? 0),
    moneyQuan: Number(raw.moneyQuan ?? raw.money_quan ?? raw.quan_money ?? 0),
    discount: Number(raw.discount ?? 0),
    cityCode: String(raw.cityCode ?? raw.city_code ?? ''),
    outTime: String(raw.outTime ?? raw.out_time ?? ''),
  }
}

export function normalizeCoupons(list: unknown[]): CouponItem[] {
  return list.map((item) => normalizeCoupon(item as Record<string, unknown>))
}

export function isCouponExpired(outTime?: string): boolean {
  if (!outTime) return false
  return new Date(outTime.replace(/-/g, '/')).getTime() < Date.now()
}

export function cityLevelAdcode(adcode: string): string {
  if (adcode.length < 4) return adcode
  return adcode.slice(0, 4) + '00'
}

/** 与后端 common/pkg/citycode.go MatchCouponCityCode 一致 */
export function matchCouponCityCode(couponCityCode: string, userAdcode: string): boolean {
  if (!couponCityCode || !userAdcode) return false
  if (couponCityCode === userAdcode) return true
  return cityLevelAdcode(couponCityCode) === cityLevelAdcode(userAdcode)
}

export type CouponWithStatus = CouponItem & {
  usable: boolean
  unusableReason?: string
}

/** 按过期与起点城市标注本单可用性（需先完成估价拿到 startAdcode） */
export function annotateCouponsForOrder(
  list: CouponItem[],
  startAdcode?: string,
): CouponWithStatus[] {
  return list.map((c) => {
    if (isCouponExpired(c.outTime)) {
      return { ...c, usable: false, unusableReason: '已过期' }
    }
    if (!startAdcode) {
      return { ...c, usable: true }
    }
    if (!c.cityCode || !matchCouponCityCode(c.cityCode, startAdcode)) {
      return { ...c, usable: false, unusableReason: '仅限指定地区，当前起点不可用' }
    }
    return { ...c, usable: true }
  })
}

export function formatCouponValue(c: CouponItem): string {
  if (c.type === 1) return (c.moneyQuan ?? 0) > 0 ? `减¥${c.moneyQuan}` : '现金抵扣'
  if (c.type === 2) return formatDiscountLabel(c.discount)
  if (c.type === 3) return '免费乘车'
  return '优惠券'
}

export function formatDiscountLabel(discount?: number): string {
  const d = discount ?? 0
  if (d <= 0) return '折扣优惠'
  if (d > 0 && d < 1) {
    const zhe = Number((d * 10).toFixed(1))
    return `${zhe % 1 === 0 ? zhe.toFixed(0) : zhe}折`
  }
  if (d <= 10) return `${d}折`
  return `${d}折`
}

export function calcDiscountedPrice(basePrice: number, coupon: CouponItem | null): number {
  if (!coupon || basePrice <= 0) return basePrice
  let price = basePrice
  if (coupon.type === 1) price = basePrice - (coupon.moneyQuan ?? 0)
  else if (coupon.type === 2) price = basePrice * (coupon.discount ?? 1)
  else if (coupon.type === 3) price = 0
  return Math.max(0, Number(price.toFixed(2)))
}

export function formatExpireTime(outTime?: string): string {
  if (!outTime) return '长期有效'
  const date = outTime.slice(0, 10)
  const time = outTime.slice(11, 16)
  return time ? `${date} ${time} 前有效` : `${date} 前有效`
}
