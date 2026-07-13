import type { Router } from 'vue-router'
import { showToast } from 'vant'
import { showAppConfirm } from '@/utils/dialog'
import { driverApi } from '@/api/driver'
import { userApi } from '@/api/user'
import type { RealNameStatus, UserRole } from '@/types'

export function normalizeRealNameStatus(data: RealNameStatus): RealNameStatus {
  return {
    ...data,
    verified: Boolean(data.verified),
    status: data.status ?? (data.verified ? 1 : 2),
    statusName: data.statusName || data.status_name || '',
    realName: data.realName || data.real_name || '',
    carNumber: data.carNumber || data.car_number || '',
    carType: data.carType || data.car_type || '',
    carColor: data.carColor || data.car_color || '',
  }
}

export function isRealNameVerified(data: RealNameStatus): boolean {
  const row = normalizeRealNameStatus(data)
  return row.verified === true || row.status === 1
}

function realNameTargetPath(role: UserRole): string {
  return role === 'passenger' ? '/passenger/realname' : '/driver/verify'
}

async function fetchRealNameStatus(role: UserRole): Promise<RealNameStatus> {
  if (role === 'passenger') {
    return userApi.realNameStatus({ silent: true })
  }
  return driverApi.realNameStatus({ silent: true })
}

/** 登录成功后检查实名状态，未认证则弹窗引导 */
export async function checkRealNameAfterLogin(role: UserRole, router: Router) {
  try {
    const data = normalizeRealNameStatus(await fetchRealNameStatus(role))
    if (isRealNameVerified(data)) return

    const targetPath = realNameTargetPath(role)
    if (router.currentRoute.value.path.startsWith(targetPath)) return

    if (data.status === 3) {
      showToast('账号已禁用，请联系客服')
      return
    }

    const isPassenger = role === 'passenger'
    await showAppConfirm({
      title: isPassenger ? '未完成实名认证' : '未完成资质认证',
      message: isPassenger
        ? '您尚未完成实名认证，请尽快完成认证以正常使用服务。'
        : '您尚未完成资质认证，完成后方可正常接单。是否前往认证？',
      confirmButtonText: '去认证',
      cancelButtonText: '稍后再说',
    })
    await router.push(targetPath)
  } catch {
    // 查询失败不阻断登录主流程
  }
}

/** 「我的」页点击认证入口：已认证则提示，未认证才跳转 */
export async function openRealNameFromMine(role: UserRole, router: Router) {
  try {
    const data = normalizeRealNameStatus(await fetchRealNameStatus(role))
    const isPassenger = role === 'passenger'

    if (data.status === 3) {
      showToast('账号已禁用，请联系客服')
      return
    }

    if (isRealNameVerified(data)) {
      const nameHint = data.realName ? `（${data.realName}）` : ''
      showToast(
        isPassenger
          ? `您已完成实名认证${nameHint}，无需重复认证`
          : `您已完成资质认证${nameHint}，无需重复认证`,
      )
      return
    }

    await router.push(realNameTargetPath(role))
  } catch (err) {
    const msg = err instanceof Error ? err.message : '认证状态查询失败'
    showToast(msg)
  }
}
