/** 克隆即用：未配 VITE_* 时用下列默认值；需要覆盖时再建 .env.production */

export const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || '/api'

/** 团队浏览器端 AK（百度控制台 Referer 需含 127.0.0.1、localhost、部署域名或 IP） */
export const BAIDU_MAP_AK =
  import.meta.env.VITE_BAIDU_MAP_AK || 'cn9OV8XucgMFpJSB8FoVwfcpJk755PBB'
