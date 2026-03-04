import { useState, useEffect, useMemo, useRef } from 'react'
import {
  Network,
  RefreshCw,
  Forward,
  Trash2,
  Edit3,
  X,
  Pause,
  Info,
  Globe,
  Zap,
  ZapOff,
  Eye,
} from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Input } from '@/components/ui/input'
import { Textarea } from '@/components/ui/textarea'
import { Switch } from '@/components/ui/switch'
import { Label } from '@/components/ui/label'
import { useProxyStore, useTrafficStore, useRepeaterStore, useIntruderStore, InterceptedRequest } from '@/store'
import {
  GetProxyStatus,
  SetInterceptMode,
  GetTraffic,
  Connect,
  Disconnect,
  ForwardIntercepted,
  DropIntercepted,
} from '../../../wailsjs/go/main/App'
import type { TrafficItem } from '@/types'
import { formatBytes, formatDuration } from '@/lib/utils'
import { cn } from '@/lib/utils'
import { MessageViewer, ViewMode } from './MessageViewer'
import { buildRequestMessage, buildResponseMessage, getStatusText } from '@/lib/formatters'
import { ResizablePanelGroup } from '@/components/ui/resizable'
import { WebSocketTab } from './WebSocketTab'
import { toast } from '@/components/ui/toast'
import { TrafficContextMenu } from './TrafficContextMenu'
import {
  TrafficFilter,
  TrafficFilterState,
  applyTrafficFilters,
  extractHosts,
} from './TrafficFilter'

// 子标签页类型
type ProxySubTab = 'intercept' | 'http-history' | 'websocket-history' | 'options'

// 默认筛选状态
const defaultFilters: TrafficFilterState = {
  searchQuery: '',
  methodFilter: 'all',
  statusFilter: 'all',
  hostFilter: 'all',
  contentTypeFilter: 'all',
  minSize: '',
  maxSize: '',
  minTime: '',
  maxTime: '',
}

export function ProxyTab() {
  const { connected, status, setStatus, setConnected, apiEndpoint } = useProxyStore()
  const {
    items,
    addItem,
    interceptMode,
    setInterceptMode,
    clearItems,
    interceptQueue,
    selectedInterceptItem,
    selectInterceptItem,
    removeFromInterceptQueue,
    setInterceptEnabled,
  } = useTrafficStore()

  // Repeater and Intruder stores
  const { addTab: addRepeaterTab, setActiveTab: setRepeaterActiveTab } = useRepeaterStore()
  const { addTab: addIntruderTab, setActiveTab: setIntruderActiveTab } = useIntruderStore()

  // 子标签页状态
  const [activeSubTab, setActiveSubTab] = useState<ProxySubTab>('http-history')

  // 筛选状态
  const [filters, setFilters] = useState<TrafficFilterState>(defaultFilters)
  const [isLoading, setIsLoading] = useState(false)

  // 选中的流量
  const [selectedTraffic, setSelectedTraffic] = useState<TrafficItem | null>(null)

  // 编辑状态
  const [editingRequest, setEditingRequest] = useState<TrafficItem | null>(null)
  const [editHeaders, setEditHeaders] = useState('')
  const [editBody, setEditBody] = useState('')

  // 消息视图模式
  const [requestViewMode, setRequestViewMode] = useState<ViewMode>('pretty')
  const [responseViewMode, setResponseViewMode] = useState<ViewMode>('pretty')
  const [interceptViewMode, setInterceptViewMode] = useState<ViewMode>('pretty')

  // 右键菜单
  const [contextMenu, setContextMenu] = useState<{ x: number; y: number; item: TrafficItem } | null>(null)

  const listRef = useRef<HTMLDivElement>(null)

  // 从流量列表提取 Host 列表
  const hostList = useMemo(() => extractHosts(items), [items])

  // 加载状态
  const loadStatus = async () => {
    try {
      const s = await GetProxyStatus()
      setStatus(s)
    } catch (e) {
      console.error('Failed to load status:', e)
    }
  }

  // 加载流量
  const loadTraffic = async () => {
    try {
      const traffic = await GetTraffic(1000)
      if (traffic && traffic.length > 0) {
        clearItems()
        traffic.forEach((item: TrafficItem) => addItem(item))
      }
    } catch (e) {
      console.error('Failed to load traffic:', e)
    }
  }

  useEffect(() => {
    if (connected) {
      loadStatus()
      loadTraffic()
    }
  }, [connected])

  useEffect(() => {
    if (!connected) return
    const interval = setInterval(() => loadStatus(), 2000)
    return () => clearInterval(interval)
  }, [connected])

  useEffect(() => {
    const handleClick = () => setContextMenu(null)
    window.addEventListener('click', handleClick)
    return () => window.removeEventListener('click', handleClick)
  }, [])

  const handleStartProxy = async () => {
    setIsLoading(true)
    try {
      await Connect()
      setConnected(true)
      loadStatus()
    } catch (e) {
      console.error('Failed to start proxy:', e)
    }
    setIsLoading(false)
  }

  const handleStopProxy = async () => {
    setIsLoading(true)
    try {
      await Disconnect()
      setConnected(false)
    } catch (e) {
      console.error('Failed to stop proxy:', e)
    }
    setIsLoading(false)
  }

  const handleToggleIntercept = async (enabled: boolean) => {
    try {
      await SetInterceptMode(enabled)
      setInterceptMode(enabled)
      setInterceptEnabled(enabled)
    } catch (e) {
      console.error('Failed to toggle intercept mode:', e)
    }
  }

  // Intercept 操作
  const handleInterceptForward = async (item?: InterceptedRequest) => {
    const target = item || selectedInterceptItem
    if (!target) return
    try {
      await ForwardIntercepted(target.id || '')
      removeFromInterceptQueue(target.id || '')
      if (selectedInterceptItem?.id === target.id) {
        // Select next item in interceptQueue
        const remaining = interceptQueue.filter(q => q.id !== target.id)
        selectInterceptItem(remaining.length > 0 ? remaining[0] : null)
      }
    } catch (e) {
      console.error('Failed to forward request:', e)
    }
  }

  const handleInterceptDrop = async (item?: InterceptedRequest) => {
    const target = item || selectedInterceptItem
    if (!target) return
    try {
      await DropIntercepted(target.id || '')
      removeFromInterceptQueue(target.id || '')
      if (selectedInterceptItem?.id === target.id) {
        // Select next item in interceptQueue
        const remaining = interceptQueue.filter(q => q.id !== target.id)
        selectInterceptItem(remaining.length > 0 ? remaining[0] : null)
      }
    } catch (e) {
      console.error('Failed to drop request:', e)
    }
  }

  const handleInterceptEdit = (item?: InterceptedRequest) => {
    const target = item || selectedInterceptItem
    if (!target) return
    // Convert InterceptedRequest to TrafficItem format for editing
    const trafficItem: TrafficItem = {
      id: target.id,
      timestamp: target.timestamp,
      method: target.method,
      url: target.url,
      host: target.host,
      path: target.path,
      statusCode: target.statusCode || 0,
      contentType: target.contentType,
      requestSize: target.requestSize,
      responseSize: 0,
      duration: 0,
      requestHeaders: target.requestHeaders || target.headers,
      responseHeaders: {},
      requestBody: target.requestBody || target.body,
      responseBody: '',
      clientIP: target.clientIP,
      protocol: target.protocol,
      intercepted: true,
    }
    setEditingRequest(trafficItem)
    const headersStr = trafficItem.requestHeaders
      ? Object.entries(trafficItem.requestHeaders).map(([k, v]) => `${k}: ${v}`).join('\n')
      : ''
    setEditHeaders(headersStr)
    setEditBody(trafficItem.requestBody || '')
  }

  const handleModifiedForward = async () => {
    if (!editingRequest) return

    // Parse headers from edit text
    const headers: Record<string, string> = {}
    editHeaders.split('\n').forEach(line => {
      const colonIndex = line.indexOf(':')
      if (colonIndex > 0) {
        const key = line.substring(0, colonIndex).trim()
        const value = line.substring(colonIndex + 1).trim()
        if (key && value) {
          headers[key] = value
        }
      }
    })

    try {
      // Dynamically import the new method
      const { ModifyAndForwardRequest } = await import('../../../wailsjs/go/main/App')

      // Send modified request
      const result = await ModifyAndForwardRequest(editingRequest.id || '', {
        method: editingRequest.method,
        url: editingRequest.url,
        headers: headers,
        body: editBody,
        originalUrl: editingRequest.url,
      })

      // Remove from intercept interceptQueue
      removeFromInterceptQueue(editingRequest.id || '')

      // Show result
      if (result.error) {
        console.error('Modified request error:', result.error)
      }
    } catch (e) {
      console.error('Failed to forward modified request:', e)
      // Fallback: try simple forward
      try {
        await ForwardIntercepted(editingRequest.id || '')
        removeFromInterceptQueue(editingRequest.id || '')
      } catch (fallbackError) {
        console.error('Fallback forward also failed:', fallbackError)
      }
    }
    setEditingRequest(null)
  }

  const handleModify = (item?: TrafficItem) => {
    const target = item || selectedTraffic
    if (!target) return
    setEditingRequest(target)
    const headersStr = target.requestHeaders
      ? Object.entries(target.requestHeaders).map(([k, v]) => `${k}: ${v}`).join('\n')
      : ''
    setEditHeaders(headersStr)
    setEditBody(target.requestBody || '')
    setActiveSubTab('intercept')
  }

  // 发送到 Repeater
  const handleSendToRepeater = (item?: TrafficItem) => {
    const target = item || selectedTraffic
    if (!target) return

    const tabId = addRepeaterTab({
      name: `${target.method} ${target.host}`,
      request: {
        id: '',
        name: '',
        method: target.method || 'GET',
        url: target.url || `https://${target.host}${target.path}`,
        headers: target.requestHeaders || {},
        body: target.requestBody || '',
        createdAt: new Date().toISOString(),
        updatedAt: new Date().toISOString()
      }
    })
    setRepeaterActiveTab(tabId)
    setContextMenu(null)
    toast.success(`已发送到 Repeater: ${target.method} ${target.host}`)
  }

  // 发送到 Intruder
  const handleSendToIntruder = (item?: TrafficItem) => {
    const target = item || selectedTraffic
    if (!target) return

    const tabId = addIntruderTab({
      name: `${target.method} ${target.host}`,
      request: {
        method: target.method || 'GET',
        url: target.url || `https://${target.host}${target.path}`,
        headers: target.requestHeaders || {},
        body: target.requestBody || ''
      }
    })
    setIntruderActiveTab(tabId)
    setContextMenu(null)
    toast.success(`已发送到 Intruder: ${target.method} ${target.host}`)
  }

  // 发送到 Scanner
  const handleSendToScanner = () => {
    if (!selectedTraffic) return
    toast.info(`已发送到 Scanner: ${selectedTraffic.method} ${selectedTraffic.host}`)
    setContextMenu(null)
  }

  // 复制成功提示
  const handleCopySuccess = () => {
    toast.success('已复制到剪贴板')
  }

  const handleContextMenu = (e: React.MouseEvent, item: TrafficItem) => {
    e.preventDefault()
    setContextMenu({ x: e.clientX, y: e.clientY, item })
  }

  // 应用筛选
  const filteredItems = useMemo(() => {
    return applyTrafficFilters(items, filters)
  }, [items, filters])

  const getStatusColor = (code: number) => {
    if (!code) return 'text-gray-400'
    if (code >= 200 && code < 300) return 'text-green-600'
    if (code >= 300 && code < 400) return 'text-blue-600'
    if (code >= 400 && code < 500) return 'text-orange-500'
    return 'text-red-500'
  }

  const getMethodColor = (method: string) => {
    switch (method) {
      case 'GET': return 'bg-blue-100 text-blue-700 border-blue-200'
      case 'POST': return 'bg-green-100 text-green-700 border-green-200'
      case 'PUT': return 'bg-yellow-100 text-yellow-700 border-yellow-200'
      case 'DELETE': return 'bg-red-100 text-red-700 border-red-200'
      case 'PATCH': return 'bg-purple-100 text-purple-700 border-purple-200'
      case 'OPTIONS': return 'bg-gray-100 text-gray-700 border-gray-200'
      case 'HEAD': return 'bg-cyan-100 text-cyan-700 border-cyan-200'
      default: return 'bg-gray-100 text-gray-700 border-gray-200'
    }
  }

  // 构建完整的请求消息
  const buildFullRequestMessage = (item: TrafficItem): string => {
    return buildRequestMessage(
      item.method || 'GET',
      item.path || '/',
      item.host || 'unknown',
      item.requestHeaders,
      item.requestBody
    )
  }

  // 构建拦截请求的消息
  const buildInterceptRequestMessage = (item: InterceptedRequest): string => {
    return buildRequestMessage(
      item.method || 'GET',
      item.path || '/',
      item.host || 'unknown',
      item.requestHeaders || item.headers,
      item.requestBody || item.body
    )
  }

  // 构建完整的响应消息
  const buildFullResponseMessage = (item: TrafficItem): string => {
    return buildResponseMessage(
      item.statusCode || 200,
      getStatusText(item.statusCode || 200),
      item.responseHeaders,
      item.responseBody,
      item.contentType
    )
  }

  // 渲染子标签页内容
  const renderSubTabContent = () => {
    switch (activeSubTab) {
      case 'intercept':
        return renderInterceptTab()
      case 'http-history':
        return renderHttpHistoryTab()
      case 'websocket-history':
        return <WebSocketTab />
      case 'options':
        return renderOptionsTab()
      default:
        return renderHttpHistoryTab()
    }
  }

  // Intercept 标签页 - Burp 风格
  const renderInterceptTab = () => (
    <div className="flex h-full flex-col">
      {/* Intercept 控制栏 - Burp 风格 */}
      <div className="h-9 border-b border-gray-200 bg-white flex items-center px-3 gap-2 flex-shrink-0">
        {/* Intercept on/off 按钮 */}
        <button
          onClick={() => handleToggleIntercept(!interceptMode)}
          className={cn(
            'h-7 px-3 text-xs font-medium rounded transition-colors',
            interceptMode
              ? 'bg-blue-500 text-white hover:bg-blue-600'
              : 'bg-gray-200 text-gray-600 hover:bg-gray-300'
          )}
        >
          Intercept is {interceptMode ? 'on' : 'off'}
        </button>

        {interceptQueue.length > 0 && (
          <>
            {/* Forward 按钮 */}
            <Button
              size="sm"
              onClick={() => handleInterceptForward()}
              disabled={!selectedInterceptItem}
              className="h-7 px-3 text-xs bg-orange-500 hover:bg-orange-600 text-white"
            >
              <Forward className="w-3.5 h-3.5 mr-1" />
              Forward
            </Button>

            {/* Drop 按钮 */}
            <Button
              size="sm"
              variant="outline"
              onClick={() => handleInterceptDrop()}
              disabled={!selectedInterceptItem}
              className="h-7 px-3 text-xs text-red-600 hover:bg-red-50"
            >
              <Trash2 className="w-3.5 h-3.5 mr-1" />
              Drop
            </Button>

            {/* Edit 按钮 */}
            <Button
              size="sm"
              variant="outline"
              onClick={() => handleInterceptEdit()}
              disabled={!selectedInterceptItem}
              className="h-7 px-3 text-xs"
            >
              <Edit3 className="w-3.5 h-3.5 mr-1" />
              Edit
            </Button>

            {/* 队列计数 */}
            <Badge className="bg-orange-500 h-5 px-2 text-[11px]">
              {interceptQueue.length} 个请求等待
            </Badge>
          </>
        )}
      </div>

      {editingRequest ? (
        // 编辑模式
        <div className="flex-1 flex flex-col overflow-hidden">
          <div className="h-8 bg-blue-50 border-b border-blue-200 flex items-center px-3 flex-shrink-0">
            <Edit3 className="w-3.5 h-3.5 text-blue-500 mr-2" />
            <Badge variant="outline" className="text-xs">{editingRequest.method}</Badge>
            <span className="text-xs text-gray-600 ml-2 truncate flex-1">
              {editingRequest.host}{editingRequest.path}
            </span>
            <div className="flex gap-1">
              <Button size="sm" variant="ghost" onClick={() => setEditingRequest(null)} className="h-6 text-xs">取消</Button>
              <Button size="sm" onClick={handleModifiedForward} className="h-6 text-xs bg-orange-500 hover:bg-orange-600 text-white">
                <Forward className="w-3 h-3 mr-1" />Forward
              </Button>
            </div>
          </div>

          <div className="flex-1 flex min-h-0">
            <div className="flex-1 flex flex-col border-r border-gray-200">
              <div className="h-6 bg-gray-100 border-b border-gray-200 flex items-center px-2 flex-shrink-0">
                <span className="text-[10px] font-medium text-gray-600">HEADERS</span>
              </div>
              <Textarea
                value={editHeaders}
                onChange={(e) => setEditHeaders(e.target.value)}
                className="flex-1 font-mono text-xs border-0 rounded-none focus-visible:ring-0 resize-none"
                placeholder="Request headers..."
              />
            </div>
            <div className="flex-1 flex flex-col">
              <div className="h-6 bg-gray-100 border-b border-gray-200 flex items-center px-2 flex-shrink-0">
                <span className="text-[10px] font-medium text-gray-600">BODY</span>
              </div>
              <Textarea
                value={editBody}
                onChange={(e) => setEditBody(e.target.value)}
                className="flex-1 font-mono text-xs border-0 rounded-none focus-visible:ring-0 resize-none"
                placeholder="Request body..."
              />
            </div>
          </div>
        </div>
      ) : (
        // 正常拦截模式 - 可拖拽面板
        <ResizablePanelGroup direction="vertical" defaultSizes={[30, 70]} minSizes={[60, 100]}>
          {/* 拦截请求列表 */}
          <div className="flex flex-col h-full bg-white">
            {/* 表头 */}
            <div className="h-6 border-b border-gray-200 bg-gray-50 flex items-center px-2 text-[10px] font-medium text-gray-500 flex-shrink-0">
              <div className="w-16">Time</div>
              <div className="w-10">Method</div>
              <div className="flex-1 min-w-0">URL</div>
              <div className="w-14 text-center">Status</div>
              <div className="w-12 text-center">Length</div>
            </div>

            {/* 请求列表 */}
            <div className="flex-1 overflow-auto">
              {interceptQueue.length === 0 ? (
                <div className="flex flex-col items-center justify-center h-full text-gray-400 py-10">
                  <Pause className="w-10 h-10 mb-2" />
                  <p className="text-sm font-medium">拦截模式</p>
                  <p className="text-xs mt-1 text-center max-w-xs">
                    {interceptMode ? '拦截已开启，等待请求...' : '开启拦截模式后，请求将在此暂停等待操作'}
                  </p>
                </div>
              ) : (
                interceptQueue.map((item, index) => (
                  <div
                    key={item.id || index}
                    onClick={() => selectInterceptItem(item)}
                    className={cn(
                      'flex items-center px-2 h-6 text-[11px] cursor-pointer border-b border-gray-50',
                      selectedInterceptItem?.id === item.id
                        ? 'bg-blue-50 border-l-2 border-l-blue-500 pl-[6px]'
                        : 'hover:bg-gray-50 border-l-2 border-l-transparent'
                    )}
                  >
                    <div className="w-16 text-gray-500 font-mono">
                      {item.timestamp ? new Date(item.timestamp).toLocaleTimeString('zh-CN', { hour: '2-digit', minute: '2-digit', second: '2-digit' }) : '-'}
                    </div>
                    <div className="w-10">
                      <span className={cn('px-1 py-0.5 rounded text-[10px] font-medium', getMethodColor(item.method))}>
                        {item.method || 'GET'}
                      </span>
                    </div>
                    <div className="flex-1 min-w-0 truncate">
                      <span className="text-gray-700">{item.host}</span>
                      <span className="text-gray-400 ml-1">{item.path}</span>
                    </div>
                    <div className={cn('w-14 text-center font-medium', getStatusColor(item.statusCode || 0))}>
                      {item.statusCode || '...'}
                    </div>
                    <div className="w-12 text-center text-gray-500">
                      {formatBytes(item.requestSize)}
                    </div>
                  </div>
                ))
              )}
            </div>
          </div>

          {/* 请求详情区 - 可拖拽 */}
          <div className="flex flex-col h-full bg-white border-t border-gray-300">
            {selectedInterceptItem ? (
              <>
                {/* 请求信息头 */}
                <div className="h-7 bg-gray-100 border-b border-gray-200 flex items-center px-3 flex-shrink-0">
                  <Badge variant="outline" className="text-[10px]">{selectedInterceptItem.method}</Badge>
                  <span className="text-[11px] text-gray-600 ml-2 truncate flex-1">
                    {selectedInterceptItem.host}{selectedInterceptItem.path}
                  </span>
                  <div className="flex items-center gap-1">
                    <Badge variant="secondary" className="text-[10px]">Request</Badge>
                  </div>
                </div>

                {/* 请求内容 */}
                <div className="flex-1 min-h-0">
                  <MessageViewer
                    title="REQUEST"
                    content={buildInterceptRequestMessage(selectedInterceptItem)}
                    viewMode={interceptViewMode}
                    contentType={selectedInterceptItem.contentType}
                    onViewModeChange={setInterceptViewMode}
                    className="h-full"
                  />
                </div>
              </>
            ) : (
              <div className="flex-1 flex items-center justify-center text-gray-400">
                <div className="text-center">
                  <Eye className="w-8 h-8 mx-auto mb-2" />
                  <p className="text-xs">选择请求查看详情</p>
                </div>
              </div>
            )}
          </div>
        </ResizablePanelGroup>
      )}
    </div>
  )

  // HTTP History 标签页 - Burp 风格，可拖拽面板
  const renderHttpHistoryTab = () => (
    <div className="flex h-full flex-col">
      {/* 增强筛选工具栏 */}
      <TrafficFilter
        filters={filters}
        onFiltersChange={setFilters}
        hostList={hostList}
        resultCount={filteredItems.length}
        totalCount={items.length}
      />

      {/* 主内容区 - 可拖拽面板 */}
      <div className="flex-1 overflow-hidden">
        {selectedTraffic ? (
          <ResizablePanelGroup direction="vertical" defaultSizes={[40, 60]} minSizes={[80, 100]}>
            {/* 流量列表 */}
            <div className="flex flex-col h-full">
              {/* 表头 */}
              <div className="h-6 border-b border-gray-200 bg-gray-50 flex items-center px-2 text-[10px] font-medium text-gray-500 flex-shrink-0">
                <div className="w-8 text-center">#</div>
                <div className="w-12">Method</div>
                <div className="flex-1 min-w-0">Host / Path</div>
                <div className="w-12 text-center">Status</div>
                <div className="w-14 text-center">Size</div>
                <div className="w-12 text-center">Time</div>
              </div>

              {/* 流量行 */}
              <div className="flex-1 overflow-auto" ref={listRef}>
                {filteredItems.length === 0 ? (
                  <div className="flex flex-col items-center justify-center h-full text-gray-400 py-20">
                    <Network className="w-10 h-10 mb-2" />
                    <p className="text-sm">暂无流量数据</p>
                    <p className="text-xs mt-1">连接到 HackMITM 服务后开始捕获流量</p>
                  </div>
                ) : (
                  filteredItems.slice(0, 500).map((item, index) => (
                    <div
                      key={item.id || index}
                      onClick={() => setSelectedTraffic(item)}
                      onContextMenu={(e) => handleContextMenu(e, item)}
                      className={cn(
                        'flex items-center px-2 h-5 text-[11px] cursor-pointer border-b border-gray-50',
                        selectedTraffic?.id === item.id
                          ? 'bg-blue-50 border-l-2 border-l-blue-500 pl-[6px]'
                          : 'hover:bg-gray-50 border-l-2 border-l-transparent'
                      )}
                    >
                      <div className="w-8 text-center text-gray-400 font-mono">{index + 1}</div>
                      <div className="w-12">
                        <span className={cn('px-1 py-0.5 rounded text-[10px] font-medium', getMethodColor(item.method))}>
                          {item.method || 'GET'}
                        </span>
                      </div>
                      <div className="flex-1 min-w-0 flex items-center">
                        <span className="text-gray-700 truncate">{item.host}</span>
                        <span className="text-gray-400 truncate ml-1">{item.path}</span>
                      </div>
                      <div className={cn('w-12 text-center font-medium', getStatusColor(item.statusCode))}>
                        {item.statusCode || '-'}
                      </div>
                      <div className="w-14 text-center text-gray-500">{formatBytes(item.responseSize)}</div>
                      <div className="w-12 text-center text-gray-400">{formatDuration(item.duration)}</div>
                    </div>
                  ))
                )}
              </div>
            </div>

            {/* 请求/响应详情面板 - 可拖拽调整 */}
            <div className="flex flex-col h-full bg-white border-t border-gray-300">
              {/* 面板头部 */}
              <div className="h-7 bg-gray-100 border-b border-gray-200 flex items-center px-3 flex-shrink-0">
                <Badge variant="outline" className="text-[10px]">{selectedTraffic.method}</Badge>
                <span className="text-[11px] text-gray-600 ml-2 truncate flex-1">
                  {selectedTraffic.host}{selectedTraffic.path}
                </span>
                <div className="flex items-center gap-2">
                  <span className={cn('text-[11px] font-medium', getStatusColor(selectedTraffic.statusCode))}>
                    {selectedTraffic.statusCode}
                  </span>
                  <span className="text-[10px] text-gray-400">{formatBytes(selectedTraffic.responseSize)}</span>
                  <Button size="sm" variant="ghost" onClick={() => setSelectedTraffic(null)} className="h-5 w-5 p-0">
                    <X className="w-3 h-3" />
                  </Button>
                </div>
              </div>

              {/* 请求/响应分栏 - 可拖拽调整 */}
              <div className="flex-1 min-h-0">
                <ResizablePanelGroup direction="horizontal" defaultSizes={[40, 60]} minSizes={[200, 200]}>
                  {/* 请求面板 */}
                  <div className="h-full">
                    <MessageViewer
                      title="REQUEST"
                      content={buildFullRequestMessage(selectedTraffic)}
                      viewMode={requestViewMode}
                      contentType={selectedTraffic.contentType}
                      onViewModeChange={setRequestViewMode}
                      onSendToRepeater={() => handleSendToRepeater(selectedTraffic)}
                      onSendToIntruder={() => handleSendToIntruder(selectedTraffic)}
                      onSendToScanner={handleSendToScanner}
                      onCopy={handleCopySuccess}
                      className="h-full border-r border-gray-200"
                    />
                  </div>

                  {/* 响应面板 */}
                  <div className="h-full">
                    <MessageViewer
                      title="RESPONSE"
                      content={buildFullResponseMessage(selectedTraffic)}
                      viewMode={responseViewMode}
                      contentType={selectedTraffic.contentType}
                      statusCode={selectedTraffic.statusCode}
                      onViewModeChange={setResponseViewMode}
                      onCopy={handleCopySuccess}
                      showRender={true}
                      className="h-full"
                    />
                  </div>
                </ResizablePanelGroup>
              </div>
            </div>
          </ResizablePanelGroup>
        ) : (
          // 没有选中时只显示列表
          <div className="flex flex-col h-full">
            <div className="h-6 border-b border-gray-200 bg-gray-50 flex items-center px-2 text-[10px] font-medium text-gray-500 flex-shrink-0">
              <div className="w-8 text-center">#</div>
              <div className="w-12">Method</div>
              <div className="flex-1 min-w-0">Host / Path</div>
              <div className="w-12 text-center">Status</div>
              <div className="w-14 text-center">Size</div>
              <div className="w-12 text-center">Time</div>
            </div>

            <div className="flex-1 overflow-auto" ref={listRef}>
              {filteredItems.length === 0 ? (
                <div className="flex flex-col items-center justify-center h-full text-gray-400 py-20">
                  <Network className="w-10 h-10 mb-2" />
                  <p className="text-sm">暂无流量数据</p>
                  <p className="text-xs mt-1">连接到 HackMITM 服务后开始捕获流量</p>
                </div>
              ) : (
                filteredItems.slice(0, 500).map((item, index) => (
                  <div
                    key={item.id || index}
                    onClick={() => setSelectedTraffic(item)}
                    onContextMenu={(e) => handleContextMenu(e, item)}
                    className={cn(
                      'flex items-center px-2 h-5 text-[11px] cursor-pointer border-b border-gray-50',
                      'hover:bg-gray-50 border-l-2 border-l-transparent'
                    )}
                  >
                    <div className="w-8 text-center text-gray-400 font-mono">{index + 1}</div>
                    <div className="w-12">
                      <span className={cn('px-1 py-0.5 rounded text-[10px] font-medium', getMethodColor(item.method))}>
                        {item.method || 'GET'}
                      </span>
                    </div>
                    <div className="flex-1 min-w-0 flex items-center">
                      <span className="text-gray-700 truncate">{item.host}</span>
                      <span className="text-gray-400 truncate ml-1">{item.path}</span>
                    </div>
                    <div className={cn('w-12 text-center font-medium', getStatusColor(item.statusCode))}>
                      {item.statusCode || '-'}
                    </div>
                    <div className="w-14 text-center text-gray-500">{formatBytes(item.responseSize)}</div>
                    <div className="w-12 text-center text-gray-400">{formatDuration(item.duration)}</div>
                  </div>
                ))
              )}
            </div>
          </div>
        )}
      </div>

      {/* 右键菜单 */}
      {contextMenu && (
        <TrafficContextMenu
          item={contextMenu.item}
          x={contextMenu.x}
          y={contextMenu.y}
          onClose={() => setContextMenu(null)}
          onSendToRepeater={handleSendToRepeater}
          onSendToIntruder={handleSendToIntruder}
          onEdit={handleModify}
        />
      )}
    </div>
  )

  // Options 标签页
  const renderOptionsTab = () => (
    <div className="flex h-full flex-col overflow-auto p-4">
      <div className="max-w-2xl mx-auto space-y-6">
        {/* API 连接设置 */}
        <div className="bg-white rounded-lg border border-gray-200 p-4">
          <h3 className="text-sm font-semibold text-gray-800 mb-4 flex items-center gap-2">
            <Globe className="w-4 h-4" />
            API 连接设置
          </h3>
          <p className="text-xs text-gray-500 mb-4">
            用于连接 HackMITM 服务的 API 端点配置。请确保 HackMITM 服务正在运行。
          </p>

          <div className="space-y-4">
            <div className="grid grid-cols-3 items-center gap-4">
              <Label className="text-xs text-gray-600">监控 API 地址</Label>
              <Input
                value={apiEndpoint}
                onChange={(e) => useProxyStore.getState().setApiEndpoint(e.target.value)}
                placeholder="http://localhost:9090"
                className="col-span-2 h-8 text-xs"
                disabled={connected}
              />
            </div>

            <div className="grid grid-cols-3 items-center gap-4">
              <Label className="text-xs text-gray-600">连接状态</Label>
              <div className="col-span-2 flex items-center gap-2">
                <Badge variant={connected ? 'success' : 'secondary'} className="text-xs">
                  {connected ? '已连接' : '未连接'}
                </Badge>
                <Button
                  size="sm"
                  variant={connected ? 'destructive' : 'default'}
                  onClick={connected ? handleStopProxy : handleStartProxy}
                  disabled={isLoading}
                  className="h-7 text-xs"
                >
                  {isLoading && <RefreshCw className="w-3 h-3 animate-spin mr-1" />}
                  {connected ? '断开连接' : '连接服务'}
                </Button>
              </div>
            </div>

            {connected && status && (
              <>
                <div className="grid grid-cols-3 items-center gap-4">
                  <Label className="text-xs text-gray-600">代理端口</Label>
                  <span className="col-span-2 text-xs text-gray-700">{status.port || 4443}</span>
                </div>
                <div className="grid grid-cols-3 items-center gap-4">
                  <Label className="text-xs text-gray-600">活跃连接</Label>
                  <span className="col-span-2 text-xs text-gray-700">{status.activeConnections || 0}</span>
                </div>
                <div className="grid grid-cols-3 items-center gap-4">
                  <Label className="text-xs text-gray-600">总请求数</Label>
                  <span className="col-span-2 text-xs text-gray-700">{status.totalRequests || 0}</span>
                </div>
              </>
            )}
          </div>
        </div>

        {/* 拦截设置 */}
        <div className="bg-white rounded-lg border border-gray-200 p-4">
          <h3 className="text-sm font-semibold text-gray-800 mb-4 flex items-center gap-2">
            <Pause className="w-4 h-4" />
            拦截设置
          </h3>

          <div className="space-y-4">
            <div className="grid grid-cols-3 items-center gap-4">
              <Label className="text-xs text-gray-600">拦截模式</Label>
              <div className="col-span-2 flex items-center gap-3">
                <Switch checked={interceptMode} onCheckedChange={handleToggleIntercept} />
                <span className="text-xs text-gray-600">
                  {interceptMode ? '已开启 - 请求将被暂停' : '已关闭 - 请求自动转发'}
                </span>
              </div>
            </div>
          </div>
        </div>

        {/* 代理设置 */}
        <div className="bg-white rounded-lg border border-gray-200 p-4">
          <h3 className="text-sm font-semibold text-gray-800 mb-4 flex items-center gap-2">
            <Network className="w-4 h-4" />
            代理设置
          </h3>

          <div className="space-y-4">
            <div className="grid grid-cols-3 items-center gap-4">
              <Label className="text-xs text-gray-600">监听地址</Label>
              <Input defaultValue="127.0.0.1" className="col-span-2 h-8 text-xs" placeholder="127.0.0.1" />
            </div>
            <div className="grid grid-cols-3 items-center gap-4">
              <Label className="text-xs text-gray-600">监听端口</Label>
              <Input defaultValue="4443" className="col-span-2 h-8 text-xs" placeholder="4443" />
            </div>
            <div className="grid grid-cols-3 items-center gap-4">
              <Label className="text-xs text-gray-600">HTTPS 支持</Label>
              <div className="col-span-2 flex items-center gap-3">
                <Switch defaultChecked />
                <span className="text-xs text-gray-600">启用 HTTPS 拦截</span>
              </div>
            </div>
          </div>
        </div>

        {/* 证书信息 */}
        <div className="bg-white rounded-lg border border-gray-200 p-4">
          <h3 className="text-sm font-semibold text-gray-800 mb-4 flex items-center gap-2">
            <Info className="w-4 h-4" />
            证书信息
          </h3>
          <p className="text-xs text-gray-500 mb-4">
            HTTPS 拦截需要安装 CA 证书。证书位置: <code className="bg-gray-100 px-1 rounded">./certs/ca-cert.pem</code>
          </p>
          <Button size="sm" variant="outline" className="h-8 text-xs">打开证书目录</Button>
        </div>
      </div>
    </div>
  )

  return (
    <div className="flex h-full flex-col">
      {/* 子标签栏 - Burp 风格 */}
      <div className="sub-tabs-bar">
        {/* 拦截开关按钮 */}
        <button
          onClick={() => handleToggleIntercept(!interceptMode)}
          className={cn(
            'sub-tab flex items-center gap-1.5',
            interceptMode && 'text-orange-500'
          )}
        >
          {interceptMode ? (
            <Zap className="w-3.5 h-3.5" />
          ) : (
            <ZapOff className="w-3.5 h-3.5" />
          )}
          Intercept is {interceptMode ? 'on' : 'off'}
        </button>

        <div className="w-px h-4 bg-gray-300 mx-2" />

        <button
          onClick={() => setActiveSubTab('http-history')}
          className={cn('sub-tab', activeSubTab === 'http-history' && 'sub-tab-active')}
        >
          HTTP history
        </button>

        {/* WebSockets history */}
        <button
          onClick={() => setActiveSubTab('websocket-history')}
          className={cn('sub-tab', activeSubTab === 'websocket-history' && 'sub-tab-active')}
        >
          WebSockets history
        </button>

        <div className="flex-1" />

        <button
          onClick={() => setActiveSubTab('options')}
          className={cn('sub-tab', activeSubTab === 'options' && 'sub-tab-active')}
        >
          Options
        </button>

        {/* 拦截队列指示器 */}
        {interceptMode && interceptQueue.length > 0 && (
          <Badge className="ml-2 h-5 px-1.5 text-[10px] bg-orange-500 animate-pulse">
            {interceptQueue.length}
          </Badge>
        )}
      </div>

      {/* 子标签内容 */}
      <div className="flex-1 overflow-hidden">
        {renderSubTabContent()}
      </div>
    </div>
  )
}
