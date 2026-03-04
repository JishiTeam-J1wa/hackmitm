/**
 * 格式化工具函数 - 用于 HTTP 消息的 Pretty 视图
 */

/**
 * 格式化 JSON 字符串
 */
export function formatJSON(text: string): string {
  if (!text || !text.trim()) return ''

  try {
    const parsed = JSON.parse(text)
    return JSON.stringify(parsed, null, 2)
  } catch {
    return text
  }
}

/**
 * 判断字符串是否为有效 JSON
 */
export function isValidJSON(text: string): boolean {
  if (!text || !text.trim()) return false

  try {
    JSON.parse(text)
    return true
  } catch {
    return false
  }
}

/**
 * 格式化 HTML/XML
 * 简单的缩进格式化
 */
export function formatHTML(text: string): string {
  if (!text || !text.trim()) return ''

  try {
    // 简单的 HTML 格式化：在标签前后添加换行和缩进
    let formatted = text
      .replace(/>\s*</g, '>\n<')  // 在标签之间添加换行
      .replace(/(<[^/][^>]*>)(.*?)(<\/[^>]*>)/g, (match, open, content, close) => {
        if (content.trim() && !content.includes('<')) {
          return `${open}${content}${close}`
        }
        return match
      })

    // 计算缩进
    const lines = formatted.split('\n')
    let indent = 0
    const indentStr = '  '

    return lines.map(line => {
      const trimmed = line.trim()
      if (!trimmed) return ''

      // 闭合标签减少缩进
      if (trimmed.startsWith('</')) {
        indent = Math.max(0, indent - 1)
      }

      const result = indentStr.repeat(indent) + trimmed

      // 开始标签增加缩进（非自闭合）
      if (trimmed.startsWith('<') &&
          !trimmed.startsWith('</') &&
          !trimmed.startsWith('<!') &&
          !trimmed.startsWith('<?') &&
          !trimmed.endsWith('/>') &&
          !trimmed.match(/^<[^>]+\/>$/)) {
        indent++
      }

      return result
    }).join('\n')
  } catch {
    return text
  }
}

/**
 * 判断 Content-Type 是否为 JSON
 */
export function isJSONContentType(contentType: string | undefined): boolean {
  if (!contentType) return false
  return contentType.includes('application/json') ||
         contentType.includes('text/json') ||
         contentType.includes('+json')
}

/**
 * 判断 Content-Type 是否为 HTML/XML
 */
export function isHTMLContentType(contentType: string | undefined): boolean {
  if (!contentType) return false
  return contentType.includes('text/html') ||
         contentType.includes('application/xhtml') ||
         contentType.includes('text/xml') ||
         contentType.includes('application/xml')
}

/**
 * 判断 Content-Type 是否为 JavaScript
 */
export function isJSContentType(contentType: string | undefined): boolean {
  if (!contentType) return false
  return contentType.includes('javascript') ||
         contentType.includes('ecmascript')
}

/**
 * 自动检测并格式化内容
 */
export function autoFormat(text: string, contentType?: string): string {
  if (!text || !text.trim()) return ''

  // 根据 Content-Type 判断
  if (contentType) {
    if (isJSONContentType(contentType)) {
      return formatJSON(text)
    }
    if (isHTMLContentType(contentType)) {
      // 尝试 JSON 格式化（可能是 JSONP）
      if (text.trim().startsWith('{') || text.trim().startsWith('[')) {
        const formatted = formatJSON(text)
        if (formatted !== text) return formatted
      }
      return formatHTML(text)
    }
  }

  // 尝试自动检测
  const trimmed = text.trim()

  // JSON
  if (trimmed.startsWith('{') || trimmed.startsWith('[')) {
    const formatted = formatJSON(text)
    if (formatted !== text) return formatted
  }

  // HTML/XML
  if (trimmed.startsWith('<')) {
    return formatHTML(text)
  }

  return text
}

/**
 * 语法高亮 - 为 JSON 添加 HTML span 标签
 */
export function highlightJSON(text: string): string {
  return text
    .replace(/("(?:[^"\\]|\\.)*")(\s*:)/g, '<span class="json-key">$1</span>$2')  // key
    .replace(/:(\s*)("(?:[^"\\]|\\.)*")/g, ':$1<span class="json-string">$2</span>')  // string value
    .replace(/:(\s*)(-?\d+\.?\d*)/g, ':$1<span class="json-number">$2</span>')  // number
    .replace(/:(\s*)(true|false)/g, ':$1<span class="json-boolean">$2</span>')  // boolean
    .replace(/:(\s*)(null)/g, ':$1<span class="json-null">$2</span>')  // null
}

/**
 * 构建 HTTP 消息（请求行/状态行 + Headers + Body）
 */
export function buildHTTPMessage(
  startLine: string,
  headers: string | Record<string, string> | undefined,
  body: string | undefined
): string {
  let message = startLine + '\n'

  if (headers) {
    if (typeof headers === 'string') {
      message += headers
      if (!headers.endsWith('\n')) {
        message += '\n'
      }
    } else {
      message += Object.entries(headers)
        .map(([key, value]) => `${key}: ${value}`)
        .join('\n')
      message += '\n'
    }
  }

  message += '\n'

  if (body) {
    message += body
  }

  return message
}

/**
 * 构建完整的 HTTP 请求消息
 */
export function buildRequestMessage(
  method: string,
  path: string,
  host: string,
  headers?: string | Record<string, string>,
  body?: string
): string {
  const startLine = `${method} ${path || '/'} HTTP/1.1`
  let message = startLine + '\n'

  // 添加 Host header
  message += `Host: ${host}\n`

  if (headers) {
    if (typeof headers === 'string') {
      // 如果是字符串，过滤掉已有的 Host header
      const headerLines = headers.split('\n').filter(line =>
        !line.toLowerCase().startsWith('host:')
      )
      message += headerLines.join('\n')
    } else {
      const { Host, host: _, ...rest } = headers as Record<string, string> & { Host?: string }
      message += Object.entries(rest)
        .map(([key, value]) => `${key}: ${value}`)
        .join('\n')
    }
    message += '\n'
  }

  message += '\n'

  if (body) {
    message += body
  }

  return message
}

/**
 * 构建完整的 HTTP 响应消息
 */
export function buildResponseMessage(
  statusCode: number,
  statusText: string,
  headers?: string | Record<string, string>,
  body?: string,
  contentType?: string
): string {
  const startLine = `HTTP/1.1 ${statusCode} ${statusText || getStatusText(statusCode)}`
  let message = startLine + '\n'

  if (contentType) {
    message += `Content-Type: ${contentType}\n`
  }

  if (headers) {
    if (typeof headers === 'string') {
      // 如果是字符串，过滤掉已有的 Content-Type header
      const headerLines = headers.split('\n').filter(line =>
        !line.toLowerCase().startsWith('content-type:')
      )
      message += headerLines.join('\n')
    } else {
      const { 'Content-Type': ct, ...rest } = headers as Record<string, string> & { 'Content-Type'?: string }
      message += Object.entries(rest)
        .map(([key, value]) => `${key}: ${value}`)
        .join('\n')
    }
    message += '\n'
  }

  message += '\n'

  if (body) {
    message += body
  }

  return message
}

/**
 * 根据状态码获取状态文本
 */
export function getStatusText(statusCode: number): string {
  const statusTexts: Record<number, string> = {
    200: 'OK',
    201: 'Created',
    204: 'No Content',
    301: 'Moved Permanently',
    302: 'Found',
    304: 'Not Modified',
    400: 'Bad Request',
    401: 'Unauthorized',
    403: 'Forbidden',
    404: 'Not Found',
    405: 'Method Not Allowed',
    500: 'Internal Server Error',
    502: 'Bad Gateway',
    503: 'Service Unavailable',
  }
  return statusTexts[statusCode] || 'Unknown'
}
