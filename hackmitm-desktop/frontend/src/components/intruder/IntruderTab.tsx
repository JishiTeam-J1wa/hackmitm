import { useState, useRef } from 'react'
import {
  Plus,
  X,
  Play,
  Pause,
  Trash2,
  Settings,
  Target,
  List,
  Zap
} from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Badge } from '@/components/ui/badge'
import { Textarea } from '@/components/ui/textarea'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { useIntruderStore, AttackResult } from '@/store/intruderStore'
import { MessageViewer } from '@/components/proxy/MessageViewer'
import { cn } from '@/lib/utils'
import { buildRequestMessage, buildResponseMessage } from '@/lib/formatters'
import { SendRequest } from '../../../wailsjs/go/main/App'
import { models } from '../../../wailsjs/go/models'

const methods = ['GET', 'POST', 'PUT', 'DELETE', 'PATCH', 'HEAD', 'OPTIONS']

const attackTypes = [
  { value: 'sniper', label: 'Sniper', desc: '使用单个payload集合，逐个位置攻击' },
  { value: 'battering_ram', label: 'Battering ram', desc: '使用单个payload集合，同时在所有位置注入' },
  { value: 'pitchfork', label: 'Pitchfork', desc: '多个payload集合，并行迭代' },
  { value: 'cluster_bomb', label: 'Cluster bomb', desc: '多个payload集合，组合所有可能' },
]

// Default simple payloads for quick start
const defaultPayloads = [
  'admin',
  'test',
  'root',
  'user',
  'password',
  '123456',
  "' OR '1'='1",
  '<script>alert(1)</script>',
  '../../../etc/passwd',
  '${7*7}',
]

export function IntruderTab() {
  const {
    tabs,
    activeTabId,
    addTab,
    removeTab,
    setActiveTab,
    setRequest,
    addPosition,
    removePosition,
    clearPositions,
    setPayloadConfig,
    setAttackType,
    addResult,
    clearResults,
    setStatus,
    setProgress,
  } = useIntruderStore()

  const activeTab = tabs.find(t => t.id === activeTabId) || tabs[0]

  // UI state
  const [activeSubTab, setActiveSubTab] = useState<'positions' | 'payloads' | 'results'>('positions')
  const [selectedResult, setSelectedResult] = useState<AttackResult | null>(null)

  // For text selection
  const [selectionStart, setSelectionStart] = useState<number | null>(null)
  const [selectionEnd, setSelectionEnd] = useState<number | null>(null)

  // Attack control
  const attackStoppedRef = useRef(false)

  // Add payload position button
  const handleAddPosition = () => {
    if (selectionStart !== null && selectionEnd !== null && selectionStart !== selectionEnd) {
      const body = activeTab.request.body
      const originalValue = body.substring(selectionStart, selectionEnd)

      addPosition(activeTab.id, {
        startIndex: selectionStart,
        endIndex: selectionEnd,
        originalValue
      })

      setSelectionStart(null)
      setSelectionEnd(null)
    }
  }

  // Clear all positions
  const handleClearPositions = () => {
    clearPositions(activeTab.id)
  }

  // Start attack - send real HTTP requests
  const handleStartAttack = async () => {
    if (activeTab.positions.length === 0 || !activeTab.request.url) return

    attackStoppedRef.current = false
    setStatus(activeTab.id, 'running')
    clearResults(activeTab.id)

    // Get payloads for each position
    const payloadsByPosition: Record<string, string[]> = {}
    activeTab.positions.forEach(pos => {
      const config = activeTab.payloadConfigs[pos.id]
      payloadsByPosition[pos.id] = config?.items?.length ? config.items : defaultPayloads
    })

    // Generate all attack combinations based on attack type
    const attackCombinations: Record<string, string>[] = []

    if (activeTab.attackType === 'sniper') {
      // Sniper: iterate through each position with all payloads
      activeTab.positions.forEach(pos => {
        const payloads = payloadsByPosition[pos.id] || []
        payloads.forEach(payload => {
          attackCombinations.push({ [pos.id]: payload })
        })
      })
    } else if (activeTab.attackType === 'battering_ram') {
      // Battering ram: same payload in all positions
      const maxLen = Math.max(...Object.values(payloadsByPosition).map(arr => arr.length))
      for (let i = 0; i < maxLen; i++) {
        const combo: Record<string, string> = {}
        activeTab.positions.forEach(pos => {
          const payloads = payloadsByPosition[pos.id] || []
          combo[pos.id] = payloads[i % payloads.length] || ''
        })
        attackCombinations.push(combo)
      }
    } else if (activeTab.attackType === 'pitchfork') {
      // Pitchfork: parallel iteration
      const maxLen = Math.max(...Object.values(payloadsByPosition).map(arr => arr.length))
      for (let i = 0; i < maxLen; i++) {
        const combo: Record<string, string> = {}
        activeTab.positions.forEach(pos => {
          const payloads = payloadsByPosition[pos.id] || []
          combo[pos.id] = payloads[i] || payloads[payloads.length - 1] || ''
        })
        attackCombinations.push(combo)
      }
    } else {
      // Cluster bomb: all combinations
      const generateCombinations = (positions: typeof activeTab.positions, idx: number, current: Record<string, string>) => {
        if (idx >= positions.length) {
          attackCombinations.push({ ...current })
          return
        }
        const pos = positions[idx]
        const payloads = payloadsByPosition[pos.id] || []
        payloads.forEach(payload => {
          current[pos.id] = payload
          generateCombinations(positions, idx + 1, current)
        })
      }
      generateCombinations(activeTab.positions, 0, {})
    }

    const totalRequests = attackCombinations.length
    setProgress(activeTab.id, 0, totalRequests)

    // Execute attack
    for (let i = 0; i < attackCombinations.length; i++) {
      if (attackStoppedRef.current) {
        setStatus(activeTab.id, 'paused')
        return
      }

      const combo = attackCombinations[i]

      // Build modified request body
      let modifiedBody = activeTab.request.body
      const sortedPositions = [...activeTab.positions].sort((a, b) => b.startIndex - a.startIndex)
      sortedPositions.forEach(pos => {
        const payload = combo[pos.id] || ''
        modifiedBody = modifiedBody.substring(0, pos.startIndex) +
          payload +
          modifiedBody.substring(pos.endIndex)
      })

      const startTime = Date.now()

      try {
        // Send actual request using backend
        const req = new models.RepeaterRequest({
          id: `intruder-${Date.now()}`,
          name: `Attack ${i + 1}`,
          method: activeTab.request.method,
          url: activeTab.request.url,
          headers: activeTab.request.headers,
          body: modifiedBody,
          createdAt: new Date().toISOString(),
          updatedAt: new Date().toISOString(),
        })

        const response = await SendRequest(req)
        const responseTime = Date.now() - startTime

        const result: AttackResult = {
          id: `result-${Date.now()}-${i}`,
          positionValues: combo,
          request: buildRequestMessage(
            activeTab.request.method,
            new URL(activeTab.request.url).pathname,
            new URL(activeTab.request.url).host,
            activeTab.request.headers,
            modifiedBody
          ),
          response: buildResponseMessage(
            response.statusCode,
            response.statusText,
            response.headers,
            response.body,
            response.headers['Content-Type']
          ),
          statusCode: response.statusCode,
          responseTime: responseTime,
          responseLength: response.contentLength,
          timestamp: new Date().toISOString()
        }

        addResult(activeTab.id, result)
      } catch (error) {
        const result: AttackResult = {
          id: `result-${Date.now()}-${i}`,
          positionValues: combo,
          request: buildRequestMessage(
            activeTab.request.method,
            new URL(activeTab.request.url).pathname,
            new URL(activeTab.request.url).host,
            activeTab.request.headers,
            modifiedBody
          ),
          response: `Error: ${error}`,
          statusCode: 0,
          responseTime: Date.now() - startTime,
          responseLength: 0,
          timestamp: new Date().toISOString(),
          error: String(error)
        }
        addResult(activeTab.id, result)
      }

      setProgress(activeTab.id, i + 1, totalRequests)

      // Small delay between requests to avoid overwhelming
      await new Promise(resolve => setTimeout(resolve, 100))
    }

    setStatus(activeTab.id, 'completed')
  }

  // Stop attack
  const handleStopAttack = () => {
    attackStoppedRef.current = true
    setStatus(activeTab.id, 'paused')
  }

  // Status badge
  const getStatusBadge = () => {
    switch (activeTab.status) {
      case 'running':
        return <Badge className="bg-green-500 animate-pulse">Running</Badge>
      case 'paused':
        return <Badge className="bg-yellow-500">Paused</Badge>
      case 'completed':
        return <Badge className="bg-blue-500">Completed</Badge>
      default:
        return <Badge variant="secondary">Idle</Badge>
    }
  }

  return (
    <div className="flex flex-col h-full">
      {/* Tab bar */}
      <div className="flex items-center gap-1 px-2 py-1 border-b border-border bg-muted/20 overflow-x-auto">
        {tabs.map((tab) => (
          <div
            key={tab.id}
            className={cn(
              'flex items-center gap-1 px-3 py-1.5 rounded-t text-sm cursor-pointer group min-w-[100px] max-w-[150px]',
              activeTabId === tab.id
                ? 'bg-background border-b-2 border-primary'
                : 'bg-muted/50 hover:bg-muted'
            )}
            onClick={() => setActiveTab(tab.id)}
          >
            <Zap className="w-3 h-3 text-orange-500" />
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
          onClick={() => addTab()}
        >
          <Plus className="w-4 h-4" />
        </Button>
      </div>

      {/* Request configuration */}
      <div className="flex-1 flex flex-col overflow-hidden">
        {/* Target configuration bar */}
        <div className="flex items-center gap-2 px-4 py-2 border-b border-border bg-gray-50">
          <Select
            value={activeTab?.request.method}
            onValueChange={(method) => setRequest(activeTab.id, { ...activeTab.request, method })}
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
            onChange={(e) => setRequest(activeTab.id, { ...activeTab.request, url: e.target.value })}
            className="flex-1 h-8"
          />

          {getStatusBadge()}

          {activeTab.status === 'running' ? (
            <Button onClick={handleStopAttack} variant="destructive" className="h-8">
              <Pause className="w-4 h-4 mr-1" />
              Pause
            </Button>
          ) : (
            <Button
              onClick={handleStartAttack}
              disabled={!activeTab?.request.url || activeTab.positions.length === 0}
              className="h-8"
            >
              <Play className="w-4 h-4 mr-1" />
              Start attack
            </Button>
          )}
        </div>

        {/* Attack type selector */}
        <div className="flex items-center gap-2 px-4 py-1.5 border-b border-border bg-white">
          <span className="text-xs text-gray-500">Attack type:</span>
          <Select
            value={activeTab.attackType}
            onValueChange={(type) => setAttackType(activeTab.id, type as any)}
          >
            <SelectTrigger className="w-40 h-7 text-xs">
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              {attackTypes.map((type) => (
                <SelectItem key={type.value} value={type.value}>
                  {type.label}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
          <span className="text-xs text-gray-400">
            {attackTypes.find(t => t.value === activeTab.attackType)?.desc}
          </span>
        </div>

        {/* Main content area with sub-tabs */}
        <div className="flex-1 flex flex-col overflow-hidden">
          {/* Sub-tabs */}
          <div className="flex items-center border-b border-border">
            <button
              onClick={() => setActiveSubTab('positions')}
              className={cn(
                'px-4 py-2 text-sm font-medium border-b-2 transition-colors',
                activeSubTab === 'positions'
                  ? 'border-orange-500 text-orange-600'
                  : 'border-transparent text-gray-500 hover:text-gray-700'
              )}
            >
              <Target className="w-4 h-4 inline mr-1" />
              Positions ({activeTab.positions.length})
            </button>
            <button
              onClick={() => setActiveSubTab('payloads')}
              className={cn(
                'px-4 py-2 text-sm font-medium border-b-2 transition-colors',
                activeSubTab === 'payloads'
                  ? 'border-orange-500 text-orange-600'
                  : 'border-transparent text-gray-500 hover:text-gray-700'
              )}
            >
              <List className="w-4 h-4 inline mr-1" />
              Payloads
            </button>
            <button
              onClick={() => setActiveSubTab('results')}
              className={cn(
                'px-4 py-2 text-sm font-medium border-b-2 transition-colors',
                activeSubTab === 'results'
                  ? 'border-orange-500 text-orange-600'
                  : 'border-transparent text-gray-500 hover:text-gray-700'
              )}
            >
              <Settings className="w-4 h-4 inline mr-1" />
              Results ({activeTab.results.length})
            </button>
          </div>

          {/* Positions tab */}
          {activeSubTab === 'positions' && (
            <div className="flex-1 flex flex-col overflow-hidden">
              {/* Position toolbar */}
              <div className="flex items-center gap-2 px-3 py-2 border-b border-gray-200 bg-gray-50">
                <Button
                  size="sm"
                  variant="outline"
                  onClick={handleAddPosition}
                  disabled={selectionStart === null || selectionEnd === null || selectionStart === selectionEnd}
                  className="h-7 text-xs"
                >
                  <Target className="w-3 h-3 mr-1" />
                  Add §
                </Button>
                <Button
                  size="sm"
                  variant="outline"
                  onClick={handleClearPositions}
                  disabled={activeTab.positions.length === 0}
                  className="h-7 text-xs"
                >
                  <Trash2 className="w-3 h-3 mr-1" />
                  Clear
                </Button>
                <span className="text-xs text-gray-500">
                  提示: 在下方编辑器中选择文本后点击 "Add §" 添加payload位置
                </span>
              </div>

              {/* Positions list */}
              {activeTab.positions.length > 0 && (
                <div className="border-b border-gray-200 bg-orange-50 p-2">
                  <div className="text-xs font-medium text-orange-700 mb-1">Payload positions:</div>
                  <div className="flex flex-wrap gap-1">
                    {activeTab.positions.map((pos) => (
                      <Badge
                        key={pos.id}
                        variant="outline"
                        className="bg-white border-orange-300 text-orange-700"
                      >
                        §{pos.originalValue}§
                        <button
                          className="ml-1 hover:text-red-500"
                          onClick={() => removePosition(activeTab.id, pos.id)}
                        >
                          <X className="w-3 h-3" />
                        </button>
                      </Badge>
                    ))}
                  </div>
                </div>
              )}

              {/* Request body editor */}
              <div className="flex-1 relative">
                <Textarea
                  value={activeTab.request.body}
                  onChange={(e) => setRequest(activeTab.id, { ...activeTab.request, body: e.target.value })}
                  onSelect={(e) => {
                    const target = e.target as HTMLTextAreaElement
                    setSelectionStart(target.selectionStart)
                    setSelectionEnd(target.selectionEnd)
                  }}
                  className="h-full font-mono text-xs border-0 rounded-none focus-visible:ring-0 resize-none"
                  placeholder="Request body...&#10;&#10;选择文本后点击 Add § 添加payload位置"
                />
              </div>
            </div>
          )}

          {/* Payloads tab */}
          {activeSubTab === 'payloads' && (
            <div className="flex-1 overflow-auto p-4">
              <div className="space-y-4">
                <div className="text-sm font-medium">Payload Settings</div>

                {activeTab.positions.length === 0 ? (
                  <div className="text-sm text-gray-500">
                    请先在 Positions 标签中添加 payload 位置
                  </div>
                ) : (
                  activeTab.positions.map((pos, idx) => (
                    <div key={pos.id} className="border border-gray-200 rounded-lg p-3">
                      <div className="flex items-center justify-between mb-2">
                        <span className="text-sm font-medium">
                          Position {idx + 1}: <code className="text-orange-600">§{pos.originalValue}§</code>
                        </span>
                      </div>
                      <Textarea
                        value={activeTab.payloadConfigs[pos.id]?.items?.join('\n') || defaultPayloads.join('\n')}
                        onChange={(e) => {
                          const items = e.target.value.split('\n').filter(s => s.trim())
                          setPayloadConfig(activeTab.id, pos.id, {
                            type: 'list',
                            items
                          })
                        }}
                        className="font-mono text-xs h-32"
                        placeholder="每行一个payload..."
                      />
                    </div>
                  ))
                )}
              </div>
            </div>
          )}

          {/* Results tab */}
          {activeSubTab === 'results' && (
            <div className="flex-1 flex overflow-hidden">
              {/* Results list */}
              <div className="w-1/2 flex flex-col border-r border-gray-200">
                {/* Results header */}
                <div className="h-6 bg-gray-100 border-b border-gray-200 flex items-center px-2 text-[10px] font-medium text-gray-500">
                  <div className="w-12">#</div>
                  <div className="flex-1">Payload</div>
                  <div className="w-12 text-center">Status</div>
                  <div className="w-16 text-center">Time</div>
                  <div className="w-16 text-center">Length</div>
                </div>

                {/* Results list */}
                <div className="flex-1 overflow-auto">
                  {activeTab.results.length === 0 ? (
                    <div className="flex items-center justify-center h-full text-gray-400">
                      <div className="text-center">
                        <Target className="w-8 h-8 mx-auto mb-2" />
                        <p className="text-xs">点击 "Start attack" 开始攻击</p>
                      </div>
                    </div>
                  ) : (
                    activeTab.results.map((result, idx) => (
                      <div
                        key={result.id}
                        onClick={() => setSelectedResult(result)}
                        className={cn(
                          'flex items-center px-2 h-6 text-[11px] cursor-pointer border-b border-gray-50',
                          selectedResult?.id === result.id
                            ? 'bg-blue-50 border-l-2 border-l-blue-500 pl-[6px]'
                            : 'hover:bg-gray-50 border-l-2 border-l-transparent'
                        )}
                      >
                        <div className="w-12 text-gray-400 font-mono">{idx + 1}</div>
                        <div className="flex-1 truncate">
                          {Object.values(result.positionValues).join(', ')}
                        </div>
                        <div className={cn(
                          'w-12 text-center font-medium',
                          result.statusCode >= 200 && result.statusCode < 300 ? 'text-green-600' :
                          result.statusCode >= 400 ? 'text-red-500' : 'text-orange-500'
                        )}>
                          {result.statusCode}
                        </div>
                        <div className="w-16 text-center text-gray-500">{result.responseTime}ms</div>
                        <div className="w-16 text-center text-gray-500">{result.responseLength}</div>
                      </div>
                    ))
                  )}
                </div>

                {/* Progress bar */}
                {activeTab.status === 'running' && (
                  <div className="h-1 bg-gray-200">
                    <div
                      className="h-full bg-green-500 transition-all duration-300"
                      style={{ width: `${(activeTab.progress.current / activeTab.progress.total) * 100}%` }}
                    />
                  </div>
                )}
              </div>

              {/* Result detail */}
              <div className="w-1/2 flex flex-col">
                {selectedResult ? (
                  <MessageViewer
                    title="RESPONSE"
                    content={selectedResult.response}
                    viewMode="pretty"
                    statusCode={selectedResult.statusCode}
                    onViewModeChange={() => {}}
                    className="h-full"
                  />
                ) : (
                  <div className="flex items-center justify-center h-full text-gray-400">
                    <div className="text-center">
                      <Target className="w-8 h-8 mx-auto mb-2" />
                      <p className="text-xs">选择结果查看详情</p>
                    </div>
                  </div>
                )}
              </div>
            </div>
          )}
        </div>
      </div>
    </div>
  )
}
