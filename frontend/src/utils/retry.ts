/**
 * 通用请求重试与加载状态机工具库
 * 提供指数退避、网络超时熔断、非幂等错误识别与重试生命周期回调
 */

export type LoadingState = 'idle' | 'loading' | 'success' | 'empty' | 'error' | 'retrying' | 'stale'

export interface RetryOptions {
  /** 最大重试次数（默认 3 次） */
  maxRetries?: number
  /** 初始重试延迟（毫秒，默认 600ms） */
  baseDelayMs?: number
  /** 最大重试延迟上限（毫秒，默认 4000ms） */
  maxDelayMs?: number
  /** 指数退避乘数（默认 2.0） */
  factor?: number
  /** 单次请求超时时间（毫秒，默认 8000ms，0 表示不设超时） */
  timeoutMs?: number
  /** 自定义判断当前错误是否可重试（不可重试如 404/403/权限拒绝等应立即中断） */
  isRetryable?: (err: any) => boolean
  /** 每次触发重试前的回调通知 */
  onRetry?: (attempt: number, delayMs: number, err: any) => void
}

/**
 * 判断错误是否属于可重试的临时网络波动或服务端错误
 */
export function isRetryableError(err: any): boolean {
  if (!err) return false
  const msg = String(err.message || err).toLowerCase()

  // 明确不可重试的客户端错误 (404 Not Found, 403 Forbidden, 401 Unauthorized, 权限错误)
  if (msg.includes('404') || msg.includes('not found')) return false
  if (msg.includes('403') || msg.includes('forbidden')) return false
  if (msg.includes('401') || msg.includes('unauthorized')) return false
  if (msg.includes('abort') || msg.includes('cancelled')) return false

  // 临时性故障：网络不可达、超时、连接重置、5xx 服务端异常、429 限流等
  if (
    msg.includes('timeout') ||
    msg.includes('timed out') ||
    msg.includes('network') ||
    msg.includes('failed to fetch') ||
    msg.includes('connection refused') ||
    msg.includes('connection reset') ||
    msg.includes('500') ||
    msg.includes('502') ||
    msg.includes('503') ||
    msg.includes('504') ||
    msg.includes('429') ||
    msg.includes('busy')
  ) {
    return true
  }

  // 默认对未知错误在限制次数内允许重试
  return true
}

/**
 * 具有指数退避与随机抖动的异步请求重试封装器
 */
export async function fetchWithRetry<T>(
  fn: () => Promise<T>,
  options: RetryOptions = {}
): Promise<T> {
  const {
    maxRetries = 3,
    baseDelayMs = 600,
    maxDelayMs = 4000,
    factor = 2,
    timeoutMs = 8000,
    isRetryable = isRetryableError,
    onRetry
  } = options

  let attempt = 0

  while (true) {
    try {
      if (timeoutMs > 0) {
        let timer: any = null
        const timeoutPromise = new Promise<never>((_, reject) => {
          timer = setTimeout(() => {
            reject(new Error(`Request timed out after ${timeoutMs}ms`))
          }, timeoutMs)
        })

        const res = await Promise.race([fn(), timeoutPromise])
        if (timer) clearTimeout(timer)
        return res
      } else {
        return await fn()
      }
    } catch (err: any) {
      attempt++

      // 如果已达最大重试次数，或者错误属于不可恢复错误，直接向上抛出
      if (attempt > maxRetries || !isRetryable(err)) {
        throw err
      }

      // 指数退避算法 + 随机抖动 (避免并发请求风暴)
      const exponentialDelay = baseDelayMs * Math.pow(factor, attempt - 1)
      const jitter = Math.floor(Math.random() * 200)
      const delayMs = Math.min(maxDelayMs, exponentialDelay + jitter)

      if (onRetry) {
        onRetry(attempt, delayMs, err)
      }

      await new Promise((resolve) => setTimeout(resolve, delayMs))
    }
  }
}
