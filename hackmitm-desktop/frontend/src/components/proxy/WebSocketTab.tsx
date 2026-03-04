import { useState, useMemo } from 'react'
import {
  ArrowUpCircle,
  ArrowDownCircle,
  Search,
  RefreshCw,
  Copy,
  Send,
  X,
  Database,
  Wifi,
  WifiOff,
} from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Badge } from '@/components/ui/badge'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { cn } from '@/lib/utils'
import { formatBytes } from '@/lib/utils'
import type { WebSocketMessage } from '@/types'
import { ResizablePanelGroup } from '@/components/ui/resizable'
import { useProxyStore } from '@/store'

type DirectionFilter = 'all' | 'incoming' | 'outgoing'
type TypeFilter = 'all' | 'text' | 'binary' | 'ping' | 'pong' | 'close'

export function WebSocketTab() {
  const { connected } = useProxyStore()
  const [messages] = useState<WebSocketMessage[]>([])
  const [selectedMessage, setSelectedMessage] = useState<WebSocketMessage | null>(null)
  const [searchQuery, setSearchQuery] = useState('')
  const [directionFilter, setDirectionFilter] = useState<DirectionFilter>('all')
  const [typeFilter, setTypeFilter] = useState<TypeFilter>('all')
  const [viewMode, setViewMode] = useState<'raw' | 'json' | 'hex'>('raw')

  // TODO: Load WebSocket messages from backend when API is available
  // For now, messages array is empty - WebSocket traffic will be captured
  // when the backend provides a WebSocket history API

  // Filter messages
  const filteredMessages = useMemo(() => {
    let result = [...messages]

    if (directionFilter !== 'all') {
      result = result.filter((m) => m.direction === directionFilter)
    }

    if (typeFilter !== 'all') {
      result = result.filter((m) => m.type === typeFilter)
    }

    if (searchQuery) {
      const query = searchQuery.toLowerCase()
      result = result.filter(
        (m) =>
          m.content.toLowerCase().includes(query) ||
          m.url.toLowerCase().includes(query) ||
          m.host.toLowerCase().includes(query)
      )
    }

    return result
  }, [messages, directionFilter, typeFilter, searchQuery])

  const getDirectionIcon = (direction: string) => {
    return direction === 'incoming' ? (
      <ArrowDownCircle className="w-4 h-4 text-green-500" />
    ) : (
      <ArrowUpCircle className="w-4 h-4 text-blue-500" />
    )
  }

  const getTypeColor = (type: string) => {
    switch (type) {
      case 'text':
        return 'bg-gray-100 text-gray-700'
      case 'binary':
        return 'bg-purple-100 text-purple-700'
      case 'ping':
        return 'bg-yellow-100 text-yellow-700'
      case 'pong':
        return 'bg-cyan-100 text-cyan-700'
      case 'close':
        return 'bg-red-100 text-red-700'
      default:
        return 'bg-gray-100 text-gray-700'
    }
  }

  const formatContent = (content: string, mode: 'raw' | 'json' | 'hex'): string => {
    if (mode === 'json') {
      try {
        const parsed = JSON.parse(content)
        return JSON.stringify(parsed, null, 2)
      } catch {
        return content
      }
    }
    if (mode === 'hex') {
      // Convert to hex representation
      return content
        .split('')
        .map((c) => c.charCodeAt(0).toString(16).padStart(2, '0'))
        .join(' ')
    }
    return content
  }

  const handleCopy = () => {
    if (selectedMessage) {
      navigator.clipboard.writeText(selectedMessage.content)
    }
  }

  const handleSendToRepeater = () => {
    // TODO: Implement send to repeater
    console.log('Send to repeaper:', selectedMessage)
  }

  const handleRefresh = () => {
    // TODO: Refresh WebSocket messages from backend
    console.log('Refresh WebSocket messages')
  }

  return (
    <div className="flex h-full flex-col">
      {/* Toolbar */}
      <div className="h-9 border-b border-gray-200 bg-white flex items-center px-3 gap-2 flex-shrink-0">
        <div className="flex items-center gap-1.5 text-xs text-gray-500">
          {connected ? (
            <Wifi className="w-3.5 h-3.5 text-green-500" />
          ) : (
            <WifiOff className="w-3.5 h-3.5 text-gray-400" />
          )}
          <span>WebSocket History</span>
        </div>

        <div className="w-px h-4 bg-gray-200 mx-2" />

        {/* Search */}
        <div className="relative flex-1 max-w-xs">
          <Search className="absolute left-2 top-1/2 -translate-y-1/2 w-3.5 h-3.5 text-gray-400" />
          <Input
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            placeholder="Search messages..."
            className="h-7 pl-7 text-xs"
          />
        </div>

        {/* Direction filter */}
        <Select value={directionFilter} onValueChange={(v) => setDirectionFilter(v as DirectionFilter)}>
          <SelectTrigger className="w-24 h-7 text-xs">
            <SelectValue />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="all">All</SelectItem>
            <SelectItem value="incoming">Incoming</SelectItem>
            <SelectItem value="outgoing">Outgoing</SelectItem>
          </SelectContent>
        </Select>

        {/* Type filter */}
        <Select value={typeFilter} onValueChange={(v) => setTypeFilter(v as TypeFilter)}>
          <SelectTrigger className="w-24 h-7 text-xs">
            <SelectValue />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="all">All Types</SelectItem>
            <SelectItem value="text">Text</SelectItem>
            <SelectItem value="binary">Binary</SelectItem>
            <SelectItem value="ping">Ping</SelectItem>
            <SelectItem value="pong">Pong</SelectItem>
            <SelectItem value="close">Close</SelectItem>
          </SelectContent>
        </Select>

        <div className="flex-1" />

        <Badge variant="secondary" className="text-xs">
          {filteredMessages.length} messages
        </Badge>

        <Button variant="ghost" size="icon" className="w-7 h-7" title="Refresh" onClick={handleRefresh}>
          <RefreshCw className="w-3.5 h-3.5" />
        </Button>
      </div>

      {/* Main content */}
      <div className="flex-1 overflow-hidden">
        {selectedMessage ? (
          <ResizablePanelGroup direction="vertical" defaultSizes={[40, 60]} minSizes={[80, 100]}>
            {/* Message list */}
            <div className="flex flex-col h-full">
              <div className="h-6 border-b border-gray-200 bg-gray-50 flex items-center px-2 text-[10px] font-medium text-gray-500 flex-shrink-0">
                <div className="w-12">Time</div>
                <div className="w-8 text-center">Dir</div>
                <div className="w-16">Type</div>
                <div className="flex-1 min-w-0">Host / URL</div>
                <div className="w-16 text-center">Size</div>
              </div>

              <div className="flex-1 overflow-auto">
                {filteredMessages.length === 0 ? (
                  <div className="flex flex-col items-center justify-center h-full text-gray-400 py-10">
                    <Database className="w-10 h-10 mb-2" />
                    <p className="text-sm font-medium">No WebSocket messages</p>
                    <p className="text-xs mt-1 text-center max-w-xs">
                      {connected
                        ? 'WebSocket messages will appear here when captured'
                        : 'Connect to the proxy to capture WebSocket traffic'}
                    </p>
                  </div>
                ) : (
                  filteredMessages.map((msg, index) => (
                    <div
                      key={msg.id || index}
                      onClick={() => setSelectedMessage(msg)}
                      className={cn(
                        'flex items-center px-2 h-6 text-[11px] cursor-pointer border-b border-gray-50',
                        selectedMessage?.id === msg.id
                          ? 'bg-blue-50 border-l-2 border-l-blue-500 pl-[6px]'
                          : 'hover:bg-gray-50 border-l-2 border-l-transparent'
                      )}
                    >
                      <div className="w-12 text-gray-500 font-mono">
                        {new Date(msg.timestamp).toLocaleTimeString('zh-CN', {
                          hour: '2-digit',
                          minute: '2-digit',
                          second: '2-digit',
                        })}
                      </div>
                      <div className="w-8 flex justify-center">{getDirectionIcon(msg.direction)}</div>
                      <div className="w-16">
                        <Badge variant="outline" className={cn('text-[10px] px-1', getTypeColor(msg.type))}>
                          {msg.type}
                        </Badge>
                      </div>
                      <div className="flex-1 min-w-0 truncate">
                        <span className="text-gray-700">{msg.host}</span>
                        <span className="text-gray-400 ml-1 text-[10px]">{msg.url.replace(`wss://${msg.host}`, '')}</span>
                      </div>
                      <div className="w-16 text-center text-gray-500">{formatBytes(msg.size)}</div>
                    </div>
                  ))
                )}
              </div>
            </div>

            {/* Message detail */}
            <div className="flex flex-col h-full bg-white border-t border-gray-300">
              <div className="h-7 bg-gray-100 border-b border-gray-200 flex items-center px-3 flex-shrink-0">
                {getDirectionIcon(selectedMessage.direction)}
                <Badge variant="outline" className={cn('ml-2 text-[10px]', getTypeColor(selectedMessage.type))}>
                  {selectedMessage.type}
                </Badge>
                <span className="text-[11px] text-gray-600 ml-2 truncate flex-1">
                  {selectedMessage.url}
                </span>
                <div className="flex items-center gap-1">
                  <Button
                    variant="ghost"
                    size="sm"
                    onClick={handleCopy}
                    className="h-5 px-2 text-[10px]"
                  >
                    <Copy className="w-3 h-3 mr-1" />
                    Copy
                  </Button>
                  <Button
                    variant="ghost"
                    size="sm"
                    onClick={handleSendToRepeater}
                    className="h-5 px-2 text-[10px]"
                  >
                    <Send className="w-3 h-3 mr-1" />
                    Repeater
                  </Button>
                  <Button
                    variant="ghost"
                    size="sm"
                    onClick={() => setSelectedMessage(null)}
                    className="h-5 w-5 p-0"
                  >
                    <X className="w-3 h-3" />
                  </Button>
                </div>
              </div>

              {/* Content viewer */}
              <div className="flex-1 min-h-0">
                <Tabs value={viewMode} onValueChange={(v) => setViewMode(v as any)} className="h-full flex flex-col">
                  <TabsList className="mx-3 mt-2 flex-shrink-0">
                    <TabsTrigger value="raw" className="text-xs">Raw</TabsTrigger>
                    <TabsTrigger value="json" className="text-xs">JSON</TabsTrigger>
                    <TabsTrigger value="hex" className="text-xs">Hex</TabsTrigger>
                  </TabsList>

                  <TabsContent value={viewMode} className="flex-1 mt-2 overflow-hidden">
                    <div className="h-full overflow-auto p-3 bg-gray-50">
                      <pre className="font-mono text-xs whitespace-pre-wrap break-all">
                        {formatContent(selectedMessage.content, viewMode)}
                      </pre>
                    </div>
                  </TabsContent>
                </Tabs>
              </div>
            </div>
          </ResizablePanelGroup>
        ) : (
          // No selection - full list view
          <div className="flex flex-col h-full">
            <div className="h-6 border-b border-gray-200 bg-gray-50 flex items-center px-2 text-[10px] font-medium text-gray-500 flex-shrink-0">
              <div className="w-12">Time</div>
              <div className="w-8 text-center">Dir</div>
              <div className="w-16">Type</div>
              <div className="flex-1 min-w-0">Host / URL</div>
              <div className="w-16 text-center">Size</div>
            </div>

            <div className="flex-1 overflow-auto">
              {filteredMessages.length === 0 ? (
                <div className="flex flex-col items-center justify-center h-full text-gray-400 py-20">
                  {connected ? (
                    <>
                      <Wifi className="w-10 h-10 mb-2" />
                      <p className="text-sm">No WebSocket messages captured</p>
                      <p className="text-xs mt-1">WebSocket traffic will appear here when available</p>
                    </>
                  ) : (
                    <>
                      <WifiOff className="w-10 h-10 mb-2" />
                      <p className="text-sm">Not connected to proxy</p>
                      <p className="text-xs mt-1">Connect to capture WebSocket traffic</p>
                    </>
                  )}
                </div>
              ) : (
                filteredMessages.map((msg, index) => (
                  <div
                    key={msg.id || index}
                    onClick={() => setSelectedMessage(msg)}
                    className={cn(
                      'flex items-center px-2 h-6 text-[11px] cursor-pointer border-b border-gray-50',
                      'hover:bg-gray-50 border-l-2 border-l-transparent'
                    )}
                  >
                    <div className="w-12 text-gray-500 font-mono">
                      {new Date(msg.timestamp).toLocaleTimeString('zh-CN', {
                        hour: '2-digit',
                        minute: '2-digit',
                        second: '2-digit',
                      })}
                    </div>
                    <div className="w-8 flex justify-center">{getDirectionIcon(msg.direction)}</div>
                    <div className="w-16">
                      <Badge variant="outline" className={cn('text-[10px] px-1', getTypeColor(msg.type))}>
                        {msg.type}
                      </Badge>
                    </div>
                    <div className="flex-1 min-w-0 truncate">
                      <span className="text-gray-700">{msg.host}</span>
                      <span className="text-gray-400 ml-1 text-[10px]">{msg.url.replace(`wss://${msg.host}`, '')}</span>
                    </div>
                    <div className="w-16 text-center text-gray-500">{formatBytes(msg.size)}</div>
                  </div>
                ))
              )}
            </div>
          </div>
        )}
      </div>
    </div>
  )
}
