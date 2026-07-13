import { showMessage } from '@/utils/toast'

export const SMS_SENT_TIP = '验证码已经发送至手机,请自行查收'

export function isValidChinaMobile(phone: string): boolean {
  return /^1[3-9]\d{9}$/.test(phone.trim())
}

export function formatSmsError(err: unknown): string {
  if (err instanceof Error) return err.message
  return '验证码发送失败'
}

export function notifySmsSentSuccess() {
  showMessage(SMS_SENT_TIP, 'success')
}
