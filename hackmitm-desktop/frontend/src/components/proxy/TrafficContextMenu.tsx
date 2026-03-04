import { Forward, Zap, Edit3, Copy } from 'lucide-react'
import type { TrafficItem } from '@/types'

/**
 * TrafficContextMenu Props
 */
export interface TrafficContextMenuProps {
  /** The traffic item to show context menu for */
  item: TrafficItem
  /** X position of the menu */
  x: number
  /** Y position of the menu */
  y: number
  /** Callback when menu should close */
  onClose: () => void
  /** Callback when user selects "Send to Repeater" */
  onSendToRepeater?: (item: TrafficItem) => void
  /** Callback when user selects "Send to Intruder" */
  onSendToIntruder?: (item: TrafficItem) => void
  /** Callback when user selects "Edit Request" */
  onEdit?: (item: TrafficItem) => void
}

/**
 * TrafficContextMenu - Reusable right-click context menu for traffic items
 *
 * Provides standard actions: Send to Repeater, Send to Intruder, Edit, Copy URL, Copy as cURL
 */
export function TrafficContextMenu({
  item,
  x,
  y,
  onClose,
  onSendToRepeater,
  onSendToIntruder,
  onEdit,
}: TrafficContextMenuProps) {
  const handleSendToRepeater = () => {
    onSendToRepeater?.(item)
    onClose()
  }

  const handleSendToIntruder = () => {
    onSendToIntruder?.(item)
    onClose()
  }

  const handleEdit = () => {
    onEdit?.(item)
    onClose()
  }

  const handleCopyUrl = () => {
    navigator.clipboard.writeText(item.url || `https://${item.host}${item.path}`)
    onClose()
  }

  const handleCopyAsCurl = () => {
    const curlCmd = `curl -X ${item.method} '${item.url || `https://${item.host}${item.path}`}'${
      item.requestHeaders
        ? ' ' + Object.entries(item.requestHeaders)
            .map(([k, v]) => `-H '${k}: ${v}'`)
            .join(' ')
        : ''
    }${item.requestBody ? ` -d '${item.requestBody}'` : ''}`
    navigator.clipboard.writeText(curlCmd)
    onClose()
  }

  return (
    <div
      className="fixed bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-lg shadow-lg py-1 z-50 min-w-[150px]"
      style={{ left: x, top: y }}
      onClick={(e) => e.stopPropagation()}
    >
      <button
        onClick={handleSendToRepeater}
        className="w-full px-3 py-1.5 text-left text-xs hover:bg-gray-100 dark:hover:bg-gray-700 flex items-center gap-2"
      >
        <Forward className="w-3.5 h-3.5 text-green-600" />
        发送到 Repeater
      </button>
      <button
        onClick={handleSendToIntruder}
        className="w-full px-3 py-1.5 text-left text-xs hover:bg-gray-100 dark:hover:bg-gray-700 flex items-center gap-2"
      >
        <Zap className="w-3.5 h-3.5 text-orange-600" />
        发送到 Intruder
      </button>
      {onEdit && (
        <button
          onClick={handleEdit}
          className="w-full px-3 py-1.5 text-left text-xs hover:bg-gray-100 dark:hover:bg-gray-700 flex items-center gap-2"
        >
          <Edit3 className="w-3.5 h-3.5 text-blue-600" />
          编辑请求
        </button>
      )}
      <div className="border-t border-gray-200 dark:border-gray-700 my-1" />
      <button
        onClick={handleCopyUrl}
        className="w-full px-3 py-1.5 text-left text-xs hover:bg-gray-100 dark:hover:bg-gray-700 flex items-center gap-2"
      >
        <Copy className="w-3.5 h-3.5" />
        复制 URL
      </button>
      <button
        onClick={handleCopyAsCurl}
        className="w-full px-3 py-1.5 text-left text-xs hover:bg-gray-100 dark:hover:bg-gray-700 flex items-center gap-2"
      >
        <Copy className="w-3.5 h-3.5" />
        复制为 cURL
      </button>
    </div>
  )
}

export default TrafficContextMenu
