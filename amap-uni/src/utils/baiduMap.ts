import { BAIDU_MAP_AK } from '@/config/env'

let loadPromise: Promise<void> | null = null

function wrapBaiduAlert(onBaiduError: (message: string) => void) {
  const originalAlert = window.alert.bind(window)
  window.alert = (message?: unknown) => {
    const text = String(message ?? '')
    if (
      text.includes('APP服务被禁用') ||
      text.includes('APP 服务被禁用') ||
      text.includes('百度未授权') ||
      text.includes('ak') ||
      text.includes('AK')
    ) {
      onBaiduError(text)
      return
    }
    originalAlert(message)
  }
  return () => {
    window.alert = originalAlert
  }
}

function waitForBMapGL(timeoutMs = 8000): Promise<void> {
  return new Promise((resolve, reject) => {
    if (window.BMapGL) {
      resolve()
      return
    }
    const started = Date.now()
    const timer = window.setInterval(() => {
      if (window.BMapGL) {
        window.clearInterval(timer)
        resolve()
        return
      }
      if (Date.now() - started >= timeoutMs) {
        window.clearInterval(timer)
        reject(new Error('BMapGL 初始化超时'))
      }
    }, 50)
  })
}

export function loadBaiduMap(): Promise<void> {
  if (typeof window !== 'undefined' && window.BMapGL) {
    return Promise.resolve()
  }
  if (loadPromise) {
    return loadPromise
  }
  const ak = BAIDU_MAP_AK
  if (!ak) {
    return Promise.reject(
      new Error('未配置百度地图 AK：复制 .env.example 为 .env.development 并填写 VITE_BAIDU_MAP_AK'),
    )
  }

  loadPromise = new Promise((resolve, reject) => {
    const callbackName = `__bmap_gl_init_${Date.now()}`
    let restoreAlert: (() => void) | null = null
    const cleanup = () => {
      delete (window as unknown as Record<string, unknown>)[callbackName]
      restoreAlert?.()
      restoreAlert = null
    }

    restoreAlert = wrapBaiduAlert((message) => {
      cleanup()
      loadPromise = null
      reject(
        new Error(
          '百度地图浏览器端 AK 无效：请在控制台新建「浏览器端」应用，开启 JavaScript API，并将 127.0.0.1 加入 Referer 白名单',
        ),
      )
      console.error('[BaiduMap]', message)
    })

    ;(window as unknown as Record<string, unknown>)[callbackName] = () => {
      cleanup()
      void waitForBMapGL()
        .then(resolve)
        .catch(reject)
    }

    const script = document.createElement('script')
    script.type = 'text/javascript'
    script.src = `https://api.map.baidu.com/api?v=1.0&type=webgl&ak=${ak}&callback=${callbackName}`
    script.onerror = () => {
      cleanup()
      loadPromise = null
      reject(new Error('百度地图脚本加载失败'))
    }
    document.head.appendChild(script)
  })

  return loadPromise
}
