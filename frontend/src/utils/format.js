export function formatBytes(bytes, decimals = 2) {
  if (bytes === 0) return '0 B'
  if (!bytes) return '-'

  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB']
  const i = Math.floor(Math.log(Math.abs(bytes)) / Math.log(k))

  return parseFloat((bytes / Math.pow(k, i)).toFixed(decimals)) + ' ' + sizes[i]
}

export function formatBytesRate(bytesPerSecond, decimals = 2) {
  if (!bytesPerSecond || bytesPerSecond === 0) return '0 B/s'

  const k = 1024
  const sizes = ['B/s', 'KB/s', 'MB/s', 'GB/s']
  const i = Math.floor(Math.log(Math.abs(bytesPerSecond)) / Math.log(k))

  return parseFloat((bytesPerSecond / Math.pow(k, i)).toFixed(decimals)) + ' ' + sizes[i]
}

export function formatUptime(seconds) {
  if (!seconds) return '0s'

  const days = Math.floor(seconds / 86400)
  const hours = Math.floor((seconds % 86400) / 3600)
  const minutes = Math.floor((seconds % 3600) / 60)
  const secs = Math.floor(seconds % 60)

  const parts = []
  if (days > 0) parts.push(`${days}d`)
  if (hours > 0) parts.push(`${hours}h`)
  if (minutes > 0) parts.push(`${minutes}m`)
  if (secs > 0 || parts.length === 0) parts.push(`${secs}s`)

  return parts.join(' ')
}

export function formatPercent(value, decimals = 1) {
  if (value === undefined || value === null) return '-'
  return value.toFixed(decimals) + '%'
}

export function formatLatency(ms) {
  if (ms === undefined || ms === null || ms < 0) return '-'
  return ms + ' ms'
}
