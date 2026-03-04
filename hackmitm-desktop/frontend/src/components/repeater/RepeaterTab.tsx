import { useState, useMemo } from 'react'
import { Plus, X, Send, Copy, Clock, HardDrive, RotateCcw } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { Badge } from '@/components/ui/badge'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { HttpEditor } from '@/components/common/HttpEditor'
import { MessageViewer, ViewMode } from '@/components/proxy/MessageViewer'
import { useRepeaterStore } from '@/store'
import { SendRequest } from '../../../wailsjs/go/main/App'
import { models } from '../../../wailsjs/go/models'
import { cn } from '@/lib/utils'
import { buildRequestMessage, buildResponseMessage } from '@/lib/formatters'

const methods = ['GET', 'POST', 'PUT', 'DELETE', 'PATCH', 'HEAD', 'OPTIONS']

export function RepeaterTab() {
  const {
    tabs,
    activeTabId,
    addTab,
    removeTab,
    setActiveTab,
    updateRequest,
    setResponse,
    setLoading,
  } = useRepeaterStore()

  const activeTab = tabs.find(t => t.id === activeTabId) || tabs[0]

  // View modes
  const [requestViewMode, setRequestViewMode] = useState<ViewMode>('pretty')

  // Build headers text from object
  const headersText = useMemo(() => {
    if (!activeTab?.request.headers) return ''
    return Object.entries(activeTab.request.headers)
      .map(([key, value]) => `${key}: ${value}`)
      .join('\n')
  }, [activeTab?.request.headers])

  const handleSend = async () => {
    if (!activeTab) return

    setLoading(activeTab.id, true)
    try {
      // Parse headers from current headers state
      const headers: Record<string, string> = { ...activeTab.request.headers }

      const req = new models.RepeaterRequest({
        id: activeTab.id,
        name: activeTab.name,
        method: activeTab.request.method,
        url: activeTab.request.url,
        headers,
        body: activeTab.request.body,
        createdAt: new Date().toISOString(),
        updatedAt: new Date().toISOString(),
      })
      const response = await SendRequest(req)
      setResponse(activeTab.id, response)
    } catch (error) {
      console.error('Request failed:', error)
      setResponse(activeTab.id, {
        statusCode: 0,
        statusText: 'Error',
        headers: {},
        body: String(error),
        responseTime: 0,
        contentLength: 0
      })
    } finally {
      setLoading(activeTab.id, false)
    }
  }

  const handleAddTab = () => {
    const newId = addTab()
    setActiveTab(newId)
  }

  // Build request message for display (Raw view)
  const buildDisplayRequest = () => {
    if (!activeTab) return ''
    try {
      return buildRequestMessage(
        activeTab.request.method,
        activeTab.request.url ? new URL(activeTab.request.url).pathname : '/',
        activeTab.request.url ? new URL(activeTab.request.url).host : '',
        activeTab.request.headers,
        activeTab.request.body
      )
    } catch {
      return ''
    }
  }

  // Build response message for display
  const buildDisplayResponse = () => {
    if (!activeTab?.response) return ''
    return buildResponseMessage(
      activeTab.response.statusCode,
      activeTab.response.statusText,
      activeTab.response.headers,
      activeTab.response.body,
      activeTab.response.headers['Content-Type']
    )
  }

  // Handle headers change
  const handleHeadersChange = (value: string) => {
    const headers: Record<string, string> = {}
    value.split('\n').forEach(line => {
      const colonIndex = line.indexOf(':')
      if (colonIndex > 0) {
        const key = line.substring(0, colonIndex).trim()
        const val = line.substring(colonIndex + 1).trim()
        if (key && val) {
          headers[key] = val
        }
      }
    })
    updateRequest(activeTab.id, { headers })
  }

  return (
    <div className="flex flex-col h-full">
      {/* Tab bar */}
      <div className="flex items-center gap-1 px-2 py-1 border-b border-border bg-muted/20 overflow-x-auto">
        {tabs.map((tab) => (
          <div
            key={tab.id}
            className={cn(
              'flex items-center gap-1 px-3 py-1.5 rounded-t text-sm cursor-pointer group min-w-[100px] max-w-[150px] transition-colors',
              activeTabId === tab.id
                ? 'bg-background border-b-2 border-primary'
                : tab.hasNewContent
                  ? 'bg-blue-100 hover:bg-blue-200 text-blue-800 font-medium border-l-2 border-l-blue-500'
                  : 'bg-muted/50 hover:bg-muted'
            )}
            onClick={() => setActiveTab(tab.id)}
          >
            {tab.hasNewContent && activeTabId !== tab.id && (
              <span className="w-2 h-2 rounded-full bg-blue-500 animate-pulse flex-shrink-0" />
            )}
            <span className="truncate flex-1">{tab.name}</span>
            <button
              className="opacity-0 group-hover:opacity-100 hover:text-destructive"
              onClick={(e) => {
                e.stopPropagation()
                removeTab(tab.id)
              }}
            >
              <X className="w-3 h-3" />
            </button>
          </div>
        ))}
        <Button
          variant="ghost"
          size="icon"
          className="w-6 h-6"
          onClick={handleAddTab}
        >
          <Plus className="w-4 h-4" />
        </Button>
      </div>

      {/* Request/Response panels */}
      <div className="flex-1 flex overflow-hidden">
        {/* Request panel */}
        <div className="flex-1 flex flex-col border-r border-border">
          {/* Request toolbar */}
          <div className="flex items-center gap-2 px-4 py-2 border-b border-border bg-gray-50 dark:bg-gray-800">
            <Select
              value={activeTab?.request.method}
              onValueChange={(method) => updateRequest(activeTab.id, { method })}
            >
              <SelectTrigger className="w-24 h-8">
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                {methods.map((method) => (
                  <SelectItem key={method} value={method}>
                    <span className={cn(
                      method === 'GET' && 'text-blue-500',
                      method === 'POST' && 'text-green-500',
                      method === 'PUT' && 'text-yellow-500',
                      method === 'DELETE' && 'text-red-500'
                    )}>
                      {method}
                    </span>
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>

            <Input
              placeholder="https://example.com/api/endpoint"
              value={activeTab?.request.url || ''}
              onChange={(e) => updateRequest(activeTab.id, { url: e.target.value })}
              className="flex-1 h-8"
            />

            <Button
              onClick={handleSend}
              disabled={!activeTab?.request.url || activeTab?.loading}
              className="h-8"
            >
              <Send className="w-4 h-4 mr-1" />
              {activeTab?.loading ? 'Sending...' : 'Send'}
            </Button>
          </div>

          {/* Request content */}
          <Tabs defaultValue="body" className="flex-1 flex flex-col">
            <TabsList className="mx-4 mt-2">
              <TabsTrigger value="body">Body</TabsTrigger>
              <TabsTrigger value="headers">Headers</TabsTrigger>
              <TabsTrigger value="raw">Raw</TabsTrigger>
            </TabsList>

            <TabsContent value="body" className="flex-1 mt-2 overflow-hidden">
              <HttpEditor
                value={activeTab?.request.body || ''}
                onChange={(value) => updateRequest(activeTab.id, { body: value })}
                language="plaintext"
                placeholder="Request body..."
                autoDetectContentType={activeTab?.request.headers?.['Content-Type']}
                height="100%"
              />
            </TabsContent>

            <TabsContent value="headers" className="flex-1 mt-2 overflow-hidden">
              <HttpEditor
                value={headersText}
                onChange={handleHeadersChange}
                language="http-request"
                placeholder="Content-Type: application/json&#10;Authorization: Bearer token"
                height="100%"
              />
            </TabsContent>

            <TabsContent value="raw" className="flex-1 mt-2">
              <MessageViewer
                title="REQUEST"
                content={buildDisplayRequest()}
                viewMode={requestViewMode}
                onViewModeChange={setRequestViewMode}
                className="h-full"
              />
            </TabsContent>
          </Tabs>
        </div>

        {/* Response panel */}
        <div className="flex-1 flex flex-col">
          {/* Response toolbar */}
          <div className="flex items-center gap-2 px-4 py-2 border-b border-border bg-gray-50 dark:bg-gray-800">
            {activeTab?.response ? (
              <>
                <Badge
                  variant={activeTab.response.statusCode < 400 ? 'success' : 'destructive'}
                  className="font-mono"
                >
                  {activeTab.response.statusCode} {activeTab.response.statusText}
                </Badge>
                <div className="flex items-center gap-1 text-xs text-gray-500">
                  <Clock className="w-3 h-3" />
                  {activeTab.response.responseTime}ms
                </div>
                <div className="flex items-center gap-1 text-xs text-gray-500">
                  <HardDrive className="w-3 h-3" />
                  {activeTab.response.contentLength} bytes
                </div>
                <div className="flex-1" />
                <Button
                  variant="ghost"
                  size="sm"
                  onClick={() => updateRequest(activeTab.id, { body: '' })}
                  className="h-7"
                >
                  <RotateCcw className="w-3 h-3 mr-1" />
                  Reset
                </Button>
                <Button
                  variant="ghost"
                  size="sm"
                  className="h-7"
                  onClick={() => navigator.clipboard.writeText(activeTab.response?.body || '')}
                >
                  <Copy className="w-3 h-3 mr-1" />
                  Copy
                </Button>
              </>
            ) : (
              <span className="text-sm text-muted-foreground">No response yet</span>
            )}
          </div>

          {/* Response content */}
          {activeTab?.response ? (
            <div className="flex-1 overflow-hidden">
              <HttpEditor
                value={buildDisplayResponse()}
                language="http-response"
                height="100%"
                autoDetectContentType={activeTab.response.headers?.['Content-Type']}
              />
            </div>
          ) : (
            <div className="flex items-center justify-center h-full text-muted-foreground bg-gray-50 dark:bg-gray-900">
              <div className="text-center">
                <Send className="w-10 h-10 mx-auto mb-3 text-gray-300" />
                <p className="text-sm">Send a request to see the response</p>
              </div>
            </div>
          )}
        </div>
      </div>
    </div>
  )
}
