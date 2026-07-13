import { showFailToast, showSuccessToast, showToast } from 'vant'

export type MessageType = 'success' | 'danger' | 'warning' | 'info'

export function showMessage(msg: string, type: MessageType = 'success') {
  const className = `app-toast app-toast--${type}`
  const message = msg.trim() || '操作完成'

  if (type === 'danger') {
    return showFailToast({ message, className })
  }
  if (type === 'success') {
    return showSuccessToast({ message, className })
  }
  if (type === 'warning') {
    return showToast({ message, icon: 'warning-o', className })
  }
  return showToast({ message, className })
}
