import { useMemo, useRef, useCallback } from 'react'
import { useVirtualizer } from '@tanstack/react-virtual'
import { ScrollArea } from '@/components/ui/scroll-area'
import { Badge } from '@/components/ui/badge'
import { useContextMenu, createTrafficMenuItems } from '@/components/ui/ContextMenu'
import { useTrafficStore, useRepeaterStore } from '@/store'
import { cn, formatBytes, getStatusCodeClass, getMethodClass } from '@/lib/utils'
import type { TrafficItem } from '@/types'

interface TrafficTableProps {
  onSelect: (item: TrafficItem | null) => void
  selectedId?: string
}

export function TrafficTable({ onSelect, selectedId }: TrafficTableProps) {
  const { items, filter, deleteItem } = useTrafficStore()
  const { addRequest } = useRepeaterStore()
  const { show } = useContextMenu()
  const parentRef = useRef<HTMLDivElement>(null)

  // Filter items
  const filteredItems = useMemo(() => {
    return items.filter(item => {
      if (filter.method && item.method !== filter.method) return false
      if (filter.host && !item.host.toLowerCase().includes(filter.host.toLowerCase())) return false
      if (filter.path && !item.path.toLowerCase().includes(filter.path.toLowerCase())) return false
      if (filter.statusCode && item.statusCode.toString() !== filter.statusCode) return false
      if (filter.search) {
        const search = filter.search.toLowerCase()
        return (
          item.url.toLowerCase().includes(search) ||
          item.host.toLowerCase().includes(search) ||
          item.path.toLowerCase().includes(search)
        )
      }
      return true
    })
  }, [items, filter])

  const virtualizer = useVirtualizer({
    count: filteredItems.length,
    getScrollElement: () => parentRef.current,
    estimateSize: () => 32,
    overscan: 10,
  })

  // Copy functions
  const copyToClipboard = useCallback((text: string) => {
    navigator.clipboard.writeText(text)
  }, [])

  const generateCurlCommand = useCallback((item: TrafficItem) => {
    let curl = `curl -X ${item.method} '${item.url}'`
    Object.entries(item.requestHeaders).forEach(([key, value]) => {
      curl += ` \\\n  -H '${key}: ${value}'`
    })
    if (item.requestBody) {
      curl += ` \\\n  -d '${item.requestBody.replace(/'/g, "'\\''")}'`
    }
    return curl
  }, [])

  // Context menu handler
  const handleContextMenu = useCallback((e: React.MouseEvent, item: TrafficItem) => {
    e.preventDefault()
    e.stopPropagation()

    const menuItems = createTrafficMenuItems(item, {
      sendToRepeater: () => {
        addRequest({
          id: `rep-${Date.now()}`,
          name: `${item.method} ${item.host}${item.path}`,
          method: item.method,
          url: item.url,
          headers: item.requestHeaders,
          body: item.requestBody,
          createdAt: new Date().toISOString(),
          updatedAt: new Date().toISOString(),
        })
      },
      copyUrl: () => copyToClipboard(item.url),
      copyAsCurl: () => copyToClipboard(generateCurlCommand(item)),
      copyRequest: () => copyToClipboard(
        `${item.method} ${item.url}\n\nHeaders:\n${Object.entries(item.requestHeaders).map(([k, v]) => `${k}: ${v}`).join('\n')}\n\nBody:\n${item.requestBody}`
      ),
      copyResponse: () => copyToClipboard(item.responseBody),
      delete: () => deleteItem(item.id),
      repeatRequest: () => {
        // TODO: Implement repeat request via API
        console.log('Repeat request:', item.id)
      },
      scanForVulns: () => {
        // TODO: Implement vulnerability scanning
        console.log('Scan for vulns:', item.id)
      },
      export: () => {
        const data = JSON.stringify(item, null, 2)
        const blob = new Blob([data], { type: 'application/json' })
        const url = URL.createObjectURL(blob)
        const a = document.createElement('a')
        a.href = url
        a.download = `traffic-${item.id}.json`
        a.click()
        URL.revokeObjectURL(url)
      },
    })

    show(e.clientX, e.clientY, menuItems)
  }, [show, addRequest, copyToClipboard, generateCurlCommand, deleteItem])

  return (
    <div className="flex flex-col h-full">
      {/* Table header */}
      <div className="flex items-center h-10 px-2 border-b border-border bg-muted/30 text-xs font-medium text-muted-foreground">
        <div className="w-20">Method</div>
        <div className="w-48 truncate">Host</div>
        <div className="flex-1 truncate">Path</div>
        <div className="w-16 text-center">Status</div>
        <div className="w-20 text-right">Size</div>
        <div className="w-16 text-right">Time</div>
      </div>

      {/* Table body */}
      <ScrollArea ref={parentRef} className="flex-1">
        <div
          style={{ height: `${virtualizer.getTotalSize()}px` }}
          className="relative"
        >
          {virtualizer.getVirtualItems().map((virtualRow) => {
            const item = filteredItems[virtualRow.index]
            return (
              <div
                key={item.id}
                style={{
                  position: 'absolute',
                  top: 0,
                  left: 0,
                  width: '100%',
                  height: `${virtualRow.size}px`,
                  transform: `translateY(${virtualRow.start}px)`,
                }}
                className={cn(
                  'flex items-center px-2 text-xs cursor-pointer border-b border-border/50 hover:bg-muted/50',
                  selectedId === item.id && 'bg-primary/10',
                  getMethodClass(item.method)
                )}
                onClick={() => onSelect(item)}
                onContextMenu={(e) => handleContextMenu(e, item)}
              >
                {/* Method */}
                <div className="w-20">
                  <Badge
                    variant="outline"
                    className={cn(
                      'text-xs font-mono',
                      item.method === 'GET' && 'text-blue-500',
                      item.method === 'POST' && 'text-green-500',
                      item.method === 'PUT' && 'text-yellow-500',
                      item.method === 'DELETE' && 'text-red-500'
                    )}
                  >
                    {item.method}
                  </Badge>
                </div>

                {/* Host */}
                <div className="w-48 truncate font-mono">{item.host}</div>

                {/* Path */}
                <div className="flex-1 truncate font-mono text-muted-foreground">
                  {item.path}
                </div>

                {/* Status */}
                <div className={cn('w-16 text-center font-mono', getStatusCodeClass(item.statusCode))}>
                  {item.statusCode}
                </div>

                {/* Size */}
                <div className="w-20 text-right font-mono text-muted-foreground">
                  {formatBytes(item.responseSize)}
                </div>

                {/* Time */}
                <div className="w-16 text-right font-mono text-muted-foreground">
                  {item.duration}ms
                </div>
              </div>
            )
          })}
        </div>
      </ScrollArea>

      {/* Footer */}
      <div className="flex items-center h-8 px-2 border-t border-border bg-muted/30 text-xs text-muted-foreground">
        {filteredItems.length} requests
      </div>
    </div>
  )
}
