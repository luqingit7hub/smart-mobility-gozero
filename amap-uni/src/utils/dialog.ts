import { showConfirmDialog, showDialog, type DialogOptions } from 'vant'

const CONFIRM_COLOR = '#1677ff'

type ConfirmOptions = {
  title?: string
  message: string
  confirmButtonText?: string
  cancelButtonText?: string
}

export function showAppConfirm(options: ConfirmOptions) {
  return showConfirmDialog({
    className: 'app-dialog',
    confirmButtonColor: CONFIRM_COLOR,
    cancelButtonColor: '#8c8c8c',
    closeOnClickOverlay: false,
    ...options,
  })
}

export function showAppNotice(title: string, message: string, options?: Partial<DialogOptions>) {
  return showDialog({
    className: 'app-dialog',
    title,
    message,
    confirmButtonText: '我知道了',
    confirmButtonColor: CONFIRM_COLOR,
    ...options,
  })
}
