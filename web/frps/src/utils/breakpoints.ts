import { ref, onBeforeUnmount, watchEffect } from 'vue'

// 全局共享的媒体查询与状态，避免重复监听
let sharedMqlSm: MediaQueryList | null = null
let sharedMqlMdUp: MediaQueryList | null = null
const sharedIsSmOrBelow = ref(false)
const sharedIsMdAndUp = ref(false)
let listenerCount = 0

const syncAll = () => {
  if (sharedMqlSm) {
    sharedIsSmOrBelow.value = sharedMqlSm.matches
  }
  if (sharedMqlMdUp) {
    sharedIsMdAndUp.value = sharedMqlMdUp.matches
  }
}

function ensureInit() {
  if (typeof window === 'undefined') return
  if (!sharedMqlSm) {
    sharedMqlSm = window.matchMedia('(max-width: 768px)')
  }
  if (!sharedMqlMdUp) {
    // md 断点采用 992px，与常见 UI 库一致
    sharedMqlMdUp = window.matchMedia('(min-width: 992px)')
  }
  syncAll()
  if ('addEventListener' in sharedMqlSm) {
    sharedMqlSm.addEventListener('change', syncAll)
  } else if ('addListener' in sharedMqlSm) {
    // @ts-expect-error: legacy API
    sharedMqlSm.addListener(syncAll)
  }
  if ('addEventListener' in sharedMqlMdUp) {
    sharedMqlMdUp.addEventListener('change', syncAll)
  } else if ('addListener' in sharedMqlMdUp) {
    // @ts-expect-error: legacy API
    sharedMqlMdUp.addListener(syncAll)
  }
}

export function useBreakpoints() {
  ensureInit()
  listenerCount++

  onBeforeUnmount(() => {
    listenerCount--
    // 当没有任何组件使用时，移除监听以避免泄漏
    if (listenerCount <= 0) {
      if (sharedMqlSm) {
        if ('removeEventListener' in sharedMqlSm) {
          sharedMqlSm.removeEventListener('change', syncAll)
        } else if ('removeListener' in sharedMqlSm) {
          // @ts-expect-error: legacy API
          sharedMqlSm.removeListener(syncAll)
        }
        sharedMqlSm = null
      }
      if (sharedMqlMdUp) {
        if ('removeEventListener' in sharedMqlMdUp) {
          sharedMqlMdUp.removeEventListener('change', syncAll)
        } else if ('removeListener' in sharedMqlMdUp) {
          // @ts-expect-error: legacy API
          sharedMqlMdUp.removeListener(syncAll)
        }
        sharedMqlMdUp = null
      }
    }
  })

  return {
    isSmOrBelow: sharedIsSmOrBelow,
    isMdAndUp: sharedIsMdAndUp,
    isMdOrBelow: refComputedMdOrBelow(),
  }
}

function refComputedMdOrBelow() {
  const r = ref(false)
  // 初始化一次
  r.value = !sharedIsMdAndUp.value
  // 订阅 mdUp 变化
  watchEffect(() => {
    r.value = !sharedIsMdAndUp.value
  })
  // 这里不返回 stop，由调用方统一通过 useBreakpoints 的 onBeforeUnmount 回收监听器
  return r
}
