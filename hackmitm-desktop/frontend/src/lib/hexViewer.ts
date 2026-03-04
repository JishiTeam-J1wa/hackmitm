/**
 * 十六进制视图工具
 */

const MAX_HEX_VIEW_SIZE = 5 * 1024 * 1024 // 5MB 最大限制
const BYTES_PER_LINE = 16

/**
 * 格式化字节数为人类可读格式
 */
function formatBytes(bytes: number): string {
  if (bytes < 1024) return bytes + ' B'
  if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(2) + ' KB'
  return (bytes / (1024 * 1024)).toFixed(2) + ' MB'
}

/**
 * 将字符串转换为十六进制视图格式
 * 格式: 00000000  48 54 54 50 2f 31 2e 31  20 32 30 30 20 4f 4b  HTTP/1.1 200 OK
 */
export function toHexView(data: string): string {
  if (!data) return ''

  const encoder = new TextEncoder()
  const bytes = encoder.encode(data)

  // 检查大小限制
  if (bytes.length > MAX_HEX_VIEW_SIZE) {
    return `[Content too large for hex view: ${formatBytes(bytes.length)}. Maximum allowed: ${formatBytes(MAX_HEX_VIEW_SIZE)}]`
  }

  return bytesToHexLines(bytes).join('\n')
}

/**
 * 将字节数组转换为十六进制行
 */
function bytesToHexLines(bytes: Uint8Array): string[] {
  const lines: string[] = []

  for (let offset = 0; offset < bytes.length; offset += BYTES_PER_LINE) {
    const slice = bytes.slice(offset, offset + BYTES_PER_LINE)

    // 偏移量
    const offsetStr = offset.toString(16).padStart(8, '0')

    // 十六进制部分（前8个字节 + 后8个字节）
    const hexParts: string[] = []
    for (let i = 0; i < BYTES_PER_LINE; i++) {
      if (i === 8) hexParts.push(' ') // 中间加空格
      if (i < slice.length) {
        hexParts.push(slice[i].toString(16).padStart(2, '0'))
      } else {
        hexParts.push('  ')
      }
    }
    const hexStr = hexParts.join(' ')

    // ASCII 部分
    let asciiStr = ''
    for (let i = 0; i < slice.length; i++) {
      const byte = slice[i]
      // 可打印 ASCII 字符 (32-126)
      if (byte >= 32 && byte <= 126) {
        asciiStr += String.fromCharCode(byte)
      } else {
        asciiStr += '.'
      }
    }

    lines.push(`${offsetStr}  ${hexStr}  ${asciiStr}`)
  }

  return lines
}

/**
 * 将 ArrayBuffer 转换为十六进制视图
 */
export function bufferToHexView(buffer: ArrayBuffer): string {
  const bytes = new Uint8Array(buffer)

  // 检查大小限制
  if (bytes.length > MAX_HEX_VIEW_SIZE) {
    return `[Content too large for hex view: ${formatBytes(bytes.length)}. Maximum allowed: ${formatBytes(MAX_HEX_VIEW_SIZE)}]`
  }

  return bytesToHexLines(bytes).join('\n')
}

/**
 * 获取十六进制视图统计信息
 */
export function getHexStats(data: string): { size: number; lines: number } {
  if (!data) return { size: 0, lines: 0 }

  const encoder = new TextEncoder()
  const bytes = encoder.encode(data)

  return {
    size: bytes.length,
    lines: Math.ceil(bytes.length / BYTES_PER_LINE)
  }
}
