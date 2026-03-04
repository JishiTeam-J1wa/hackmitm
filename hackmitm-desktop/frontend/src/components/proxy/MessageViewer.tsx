import { useState, useMemo, useRef, useEffect, useCallback } from 'react'
import { Search, ChevronUp, ChevronDown, X, Copy, Send, Repeat, Crosshair, Shield } from 'lucide-react'
import DOMPurify from 'dompurify'
import { cn } from '@/lib/utils'
import { autoFormat } from '@/lib/formatters'
import { toHexView } from '@/lib/hexViewer'
import { useContextMenu } from '@/components/ui/ContextMenu'
import type { MenuItem } from '@/components/ui/ContextMenu'

export type ViewMode = 'pretty' | 'raw' | 'hex' | 'render'

interface MessageViewerProps {
  title: 'REQUEST' | 'RESPONSE'
  content: string
  viewMode: ViewMode
  contentType?: string
  statusCode?: number
  onViewModeChange: (mode: ViewMode) => void
  showRender?: boolean
  className?: string
  // 右键菜单回调
  onSendToRepeater?: () => void
  onSendToIntruder?: () => void
  onSendToScanner?: () => void
  onCopy?: () => void
}

/**
 * 消息查看器组件 - 类似 Burp Suite 的完整 HTTP 消息显示
 */
export function MessageViewer({
  title,
  content,
  viewMode,
  contentType,
  statusCode,
  onViewModeChange,
  showRender = false,
  className,
  onSendToRepeater,
  onSendToIntruder,
  onSendToScanner,
  onCopy,
}: MessageViewerProps) {
  const [searchInput, setSearchInput] = useState('')
  const [searchQuery, setSearchQuery] = useState('')
  const [currentMatchIndex, setCurrentMatchIndex] = useState(0)
  const contentRef = useRef<HTMLDivElement>(null)
  const { show } = useContextMenu()

  // 搜索防抖 - 延迟 150ms 后才执行搜索
  useEffect(() => {
    const timer = setTimeout(() => {
      setSearchQuery(searchInput)
    }, 150)
    return () => clearTimeout(timer)
  }, [searchInput])

  // 根据视图模式格式化内容
  const formattedContent = useMemo(() => {
    if (!content) return ''

    switch (viewMode) {
      case 'pretty':
        return autoFormat(content, contentType)
      case 'raw':
        return content
      case 'hex':
        return toHexView(content)
      case 'render':
        return content
      default:
        return content
    }
  }, [content, viewMode, contentType])

  // 搜索匹配
  const searchMatches = useMemo(() => {
    if (!searchQuery || !formattedContent) return []

    const matches: number[] = []
    const regex = new RegExp(searchQuery.replace(/[.*+?^${}()|[\]\\]/g, '\\$&'), 'gi')
    let match

    while ((match = regex.exec(formattedContent)) !== null) {
      matches.push(match.index)
    }

    return matches
  }, [searchQuery, formattedContent])

  // 重置匹配索引
  useEffect(() => {
    setCurrentMatchIndex(0)
  }, [searchQuery])

  // 高亮显示搜索结果
  const highlightedContent = useMemo(() => {
    if (!searchQuery || !formattedContent || searchMatches.length === 0) {
      return formattedContent
    }

    let result = ''
    let lastIndex = 0

    searchMatches.forEach((matchIndex, idx) => {
      result += formattedContent.slice(lastIndex, matchIndex)
      const matchText = formattedContent.slice(matchIndex, matchIndex + searchQuery.length)
      const isCurrentMatch = idx === currentMatchIndex
      result += `<mark class="${isCurrentMatch ? 'search-match-current' : 'search-match'}">${escapeHtml(matchText)}</mark>`
      lastIndex = matchIndex + searchQuery.length
    })

    result += formattedContent.slice(lastIndex)
    return result
  }, [formattedContent, searchQuery, searchMatches, currentMatchIndex])

  // 导航匹配
  const navigateMatch = (direction: 'up' | 'down') => {
    if (searchMatches.length === 0) return

    if (direction === 'down') {
      setCurrentMatchIndex((prev) => (prev + 1) % searchMatches.length)
    } else {
      setCurrentMatchIndex((prev) => (prev - 1 + searchMatches.length) % searchMatches.length)
    }
  }

  // 可用的视图选项
  const availableViews: ViewMode[] = useMemo(() => {
    const views: ViewMode[] = ['pretty', 'raw', 'hex']
    if (showRender) {
      views.push('render')
    }
    return views
  }, [showRender])

  // 状态码颜色
  const statusColor = useMemo(() => {
    if (!statusCode) return ''
    if (statusCode >= 200 && statusCode < 300) return 'text-green-600'
    if (statusCode >= 300 && statusCode < 400) return 'text-blue-600'
    if (statusCode >= 400 && statusCode < 500) return 'text-orange-500'
    return 'text-red-500'
  }, [statusCode])

  // 处理右键菜单
  const handleContextMenu = useCallback((e: React.MouseEvent) => {
    e.preventDefault()

    const menuItems: MenuItem[] = [
      {
        id: 'copy',
        label: '复制内容',
        icon: <Copy className="w-4 h-4" />,
        shortcut: 'Ctrl+C',
        action: () => {
          navigator.clipboard.writeText(content)
          onCopy?.()
        },
      },
    ]

    // 只有 REQUEST 才显示发送选项
    if (title === 'REQUEST') {
      menuItems.push(
        {
          id: 'divider-1',
          label: '',
          divider: true,
        },
        {
          id: 'send-to',
          label: '发送到',
          icon: <Send className="w-4 h-4" />,
          children: [
            {
              id: 'repeater',
              label: 'Repeater',
              icon: <Repeat className="w-4 h-4" />,
              shortcut: 'Ctrl+R',
              action: onSendToRepeater,
            },
            {
              id: 'intruder',
              label: 'Intruder',
              icon: <Crosshair className="w-4 h-4" />,
              shortcut: 'Ctrl+I',
              action: onSendToIntruder,
            },
            {
              id: 'scanner',
              label: 'Scanner',
              icon: <Shield className="w-4 h-4" />,
              shortcut: 'Ctrl+S',
              action: onSendToScanner,
            },
          ],
        }
      )
    }

    show(e.clientX, e.clientY, menuItems)
  }, [content, title, onSendToRepeater, onSendToIntruder, onSendToScanner, onCopy, show])

  return (
    <div className={cn('message-viewer flex flex-col h-full', className)}>
      {/* 头部 */}
      <div className="message-viewer-header h-6 bg-gray-200 flex items-center px-2 flex-shrink-0">
        <span className="text-[10px] font-semibold text-gray-600 mr-2">{title}</span>
        {statusCode !== undefined && (
          <span className={cn('text-[10px] px-1.5 py-0.5 rounded bg-gray-300', statusColor)}>
            {statusCode}
          </span>
        )}
        <div className="flex-1" />
        {/* 视图选项 */}
        <div className="flex items-center gap-0.5">
          {availableViews.map((view) => (
            <button
              key={view}
              onClick={() => onViewModeChange(view)}
              className={cn(
                'px-2 py-0.5 text-[10px] rounded transition-colors',
                viewMode === view
                  ? 'bg-white text-gray-800 shadow-sm'
                  : 'text-gray-500 hover:text-gray-700'
              )}
            >
              {view.charAt(0).toUpperCase() + view.slice(1)}
            </button>
          ))}
        </div>
      </div>

      {/* 内容区域 */}
      <div
        ref={contentRef}
        className="message-viewer-content flex-1 overflow-auto p-2 font-mono text-[11px] bg-gray-50 text-gray-800 leading-relaxed cursor-text"
        onContextMenu={handleContextMenu}
      >
        {viewMode === 'render' && showRender ? (
          <RenderView content={content} />
        ) : (
          <pre
            className="whitespace-pre-wrap break-all select-text"
            dangerouslySetInnerHTML={{ __html: highlightedContent || escapeHtml(formattedContent) }}
          />
        )}
        {!content && (
          <span className="text-gray-400">No content</span>
        )}
      </div>

      {/* 搜索栏 */}
      <div className="message-viewer-search h-6 bg-gray-50 border-t border-gray-200 flex items-center px-2 flex-shrink-0">
        <Search className="w-3 h-3 text-gray-400 mr-1.5 flex-shrink-0" />
        <input
          type="text"
          value={searchInput}
          onChange={(e) => setSearchInput(e.target.value)}
          placeholder="搜索..."
          className="h-4 text-[10px] border-0 bg-transparent shadow-none focus:outline-none px-0 flex-1 min-w-0"
        />
        {searchInput && (
          <>
            <span className="text-[10px] text-gray-400 mx-1 flex-shrink-0">
              {searchMatches.length > 0 ? `${currentMatchIndex + 1}/${searchMatches.length}` : '0'}
            </span>
            <button
              onClick={() => navigateMatch('up')}
              disabled={searchMatches.length === 0}
              className="p-0.5 text-gray-400 hover:text-gray-600 disabled:opacity-30 flex-shrink-0"
            >
              <ChevronUp className="w-3 h-3" />
            </button>
            <button
              onClick={() => navigateMatch('down')}
              disabled={searchMatches.length === 0}
              className="p-0.5 text-gray-400 hover:text-gray-600 disabled:opacity-30 flex-shrink-0"
            >
              <ChevronDown className="w-3 h-3" />
            </button>
            <button
              onClick={() => {
                setSearchInput('')
                setSearchQuery('')
              }}
              className="p-0.5 text-gray-400 hover:text-gray-600 ml-1 flex-shrink-0"
            >
              <X className="w-3 h-3" />
            </button>
          </>
        )}
      </div>
    </div>
  )
}

/**
 * HTML 渲染视图组件
 */
function RenderView({ content }: { content: string }) {
  // 使用 DOMPurify 进行安全的 HTML 清理
  const sanitizedContent = useMemo(() => {
    if (!content) return ''

    // 使用 DOMPurify 清理 HTML，移除危险的标签和属性
    return DOMPurify.sanitize(content, {
      FORBID_TAGS: ['script', 'iframe', 'object', 'embed', 'form', 'input', 'button'],
      FORBID_ATTR: ['onerror', 'onload', 'onclick', 'onmouseover', 'onfocus', 'onblur', 'formaction', 'action'],
      ALLOW_DATA_ATTR: false,
      ADD_ATTR: ['target'],
      FORCE_BODY: true,
    })
  }, [content])

  // 使用 srcdoc 而不是 src + blob URL
  const srcDoc = useMemo(() => {
    // 添加基础样式和安全策略
    return `
      <!DOCTYPE html>
      <html>
        <head>
          <meta charset="UTF-8">
          <style>
            body {
              font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
              font-size: 12px;
              line-height: 1.5;
              margin: 0;
              padding: 8px;
              color: #333;
            }
            img { max-width: 100%; height: auto; }
          </style>
        </head>
        <body>
          ${sanitizedContent}
        </body>
      </html>
    `
  }, [sanitizedContent])

  if (!content) {
    return (
      <div className="text-gray-400 text-center py-8">
        <p className="text-xs">No content to render</p>
      </div>
    )
  }

  return (
    <iframe
      srcDoc={srcDoc}
      className="w-full h-full border-0 bg-white"
      sandbox="allow-same-origin"
      title="Render preview"
    />
  )
}

/**
 * HTML 转义
 */
function escapeHtml(text: string): string {
  const div = document.createElement('div')
  div.textContent = text
  return div.innerHTML
}

export default MessageViewer
