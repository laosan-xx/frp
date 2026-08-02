// 跨组件实例缓存：远程命令标签页（passwall / 系统更新 等）之间切换会导致
// ClientDetail 组件整体重新挂载，而顶部“当前节点信息”面板依赖 frpc 实时探测
// 出口 IP（起临时代理 + 公网探测，耗时数秒）。IP 信息（出口IP + 归属 + 分类）
// 在几分钟内基本不变，因此把最近一次结果按 client.key + 上下文 缓存到模块级，
// 切回 passwall 时先秒显缓存，避免每次都重新跑那条昂贵的探测链路。
export interface CachedCurrentNode {
  code: string
  latency: string
  error: string
  ip: string
  location: string
  isp: string
  ip_country: string
  ip_type: string
  is_isp: string
}

const cache = new Map<string, { value: CachedCurrentNode; ts: number }>()

// 缓存有效期：3 分钟内视为新鲜，切回标签页直接秒显；超过则重新探测一次。
// 选 3 分钟是权衡：既覆盖“来回切几个标签页”的体感，又保证节点出口 IP 真变了
// 之后（如用户手动切节点）不至于长期显示陈旧数据。
export const CURRENT_NODE_TTL = 3 * 60 * 1000

export function getCachedCurrentNode(key: string): CachedCurrentNode | null {
  const entry = cache.get(key)
  if (entry && Date.now() - entry.ts < CURRENT_NODE_TTL) {
    return entry.value
  }
  return null
}

export function setCachedCurrentNode(key: string, value: CachedCurrentNode): void {
  cache.set(key, { value, ts: Date.now() })
}
