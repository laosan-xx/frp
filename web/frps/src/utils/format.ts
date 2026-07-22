import i18n from '../locales'

const t = (key: string, params: Record<string, unknown> = {}) =>
  i18n.global.t(key, params)

export function formatDistanceToNow(date: Date): string {
  const seconds = Math.floor((new Date().getTime() - date.getTime()) / 1000)

  let interval = seconds / 31536000
  if (interval > 1) return t('timeAgo.years', { n: Math.floor(interval) })

  interval = seconds / 2592000
  if (interval > 1) return t('timeAgo.months', { n: Math.floor(interval) })

  interval = seconds / 86400
  if (interval > 1) return t('timeAgo.days', { n: Math.floor(interval) })

  interval = seconds / 3600
  if (interval > 1) return t('timeAgo.hours', { n: Math.floor(interval) })

  interval = seconds / 60
  if (interval > 1) return t('timeAgo.minutes', { n: Math.floor(interval) })

  return t('timeAgo.seconds', { n: Math.floor(seconds) })
}

export function formatUnixSeconds(seconds?: number): string {
  if (seconds == null || !Number.isFinite(seconds) || seconds <= 0) return ''

  const date = new Date(seconds * 1000)
  const pad = (value: number) => value.toString().padStart(2, '0')
  return `${pad(date.getMonth() + 1)}-${pad(date.getDate())} ${pad(date.getHours())}:${pad(date.getMinutes())}:${pad(date.getSeconds())}`
}

export function formatFileSize(bytes: number): string {
  if (!Number.isFinite(bytes) || bytes < 0) return '0 B'
  if (bytes === 0) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB', 'PB', 'EB', 'ZB', 'YB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  // Prevent index out of bounds for extremely large numbers
  const unit = sizes[i] || sizes[sizes.length - 1]
  const val = bytes / Math.pow(k, i)

  return parseFloat(val.toFixed(2)) + ' ' + unit
}
