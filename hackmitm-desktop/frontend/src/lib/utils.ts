import { type ClassValue, clsx } from "clsx"
import { twMerge } from "tailwind-merge"

export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs))
}

export function formatBytes(bytes: number | undefined | null, decimals = 2): string {
  // Handle undefined, null, NaN, and invalid values
  if (bytes === undefined || bytes === null || isNaN(bytes) || !isFinite(bytes)) {
    return '0 B'
  }
  if (bytes === 0) return '0 B'
  if (bytes < 0) return '0 B'

  const k = 1024
  const dm = decimals < 0 ? 0 : decimals
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  const safeI = Math.min(i, sizes.length - 1)
  return `${parseFloat((bytes / Math.pow(k, safeI)).toFixed(dm))} ${sizes[safeI]}`
}

export function formatDuration(ms: number | undefined | null): string {
  // Handle undefined, null, NaN
  if (ms === undefined || ms === null || isNaN(ms) || !isFinite(ms)) {
    return '-'
  }
  if (ms < 1000) return `${Math.round(ms)}ms`
  if (ms < 60000) return `${(ms / 1000).toFixed(2)}s`
  return `${(ms / 60000).toFixed(2)}m`
}

export function formatTimestamp(date: Date | string): string {
  const d = typeof date === 'string' ? new Date(date) : date
  return d.toLocaleTimeString('en-US', {
    hour12: false,
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit'
  })
}

export function getStatusCodeClass(statusCode: number): string {
  if (statusCode >= 200 && statusCode < 300) return 'status-2xx'
  if (statusCode >= 300 && statusCode < 400) return 'status-3xx'
  if (statusCode >= 400 && statusCode < 500) return 'status-4xx'
  if (statusCode >= 500) return 'status-5xx'
  return ''
}

export function getMethodClass(method: string): string {
  const m = method.toUpperCase()
  switch (m) {
    case 'GET': return 'row-get'
    case 'POST': return 'row-post'
    case 'PUT': return 'row-put'
    case 'DELETE': return 'row-delete'
    default: return ''
  }
}

export function parseHeaders(headerString: string): Record<string, string> {
  const headers: Record<string, string> = {}
  const lines = headerString.split('\n')
  for (const line of lines) {
    const colonIndex = line.indexOf(':')
    if (colonIndex > 0) {
      const key = line.substring(0, colonIndex).trim()
      const value = line.substring(colonIndex + 1).trim()
      headers[key] = value
    }
  }
  return headers
}

export function headersToString(headers: Record<string, string>): string {
  return Object.entries(headers)
    .map(([key, value]) => `${key}: ${value}`)
    .join('\n')
}
