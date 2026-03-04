import { useState, useEffect } from 'react'
import {
  Play,
  Pause,
  Square,
  Plus,
  X,
  Target,
  AlertTriangle,
  CheckCircle,
  ChevronDown,
  ChevronUp,
  Filter,
  RefreshCw,
} from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Badge } from '@/components/ui/badge'
import { Checkbox } from '@/components/ui/checkbox'
import { Label } from '@/components/ui/label'
import { Progress } from '@/components/ui/progress'
import { cn } from '@/lib/utils'
import { useActiveScanStore } from '@/store/activeScanStore'

// Severity colors
const severityColors: Record<string, string> = {
  critical: 'bg-purple-500 text-white',
  high: 'bg-red-500 text-white',
  medium: 'bg-orange-500 text-white',
  low: 'bg-yellow-500 text-black',
  info: 'bg-blue-500 text-white',
}

const severityBorderColors: Record<string, string> = {
  critical: 'border-purple-500',
  high: 'border-red-500',
  medium: 'border-orange-500',
  low: 'border-yellow-500',
  info: 'border-blue-500',
}

export function ActiveScanTab() {
  const {
    targets,
    plugins,
    activeScanId,
    progress,
    findings,
    isLoading,
    selectedFinding,
    createScan,
    startScan,
    pauseScan,
    resumeScan,
    stopScan,
    addTarget,
    removeTarget,
    clearTargets,
    togglePlugin,
    loadProgress,
    loadFindings,
    selectFinding,
    setActiveScan,
  } = useActiveScanStore()

  // Local state
  const [newTargetUrl, setNewTargetUrl] = useState('')
  const [newTargetMethod, setNewTargetMethod] = useState('GET')
  const [showAdvanced, setShowAdvanced] = useState(false)
  const [concurrency, setConcurrency] = useState(5)
  const [rateLimit, setRateLimit] = useState(10)
  const [timeout, setTimeout_] = useState(30)
  const [severityFilter, setSeverityFilter] = useState<string>('all')

  // Current scan state
  const currentProgress = activeScanId ? progress[activeScanId] : null
  const currentFindings = activeScanId ? findings[activeScanId] || [] : []
  const scanStatus = currentProgress?.status || 'idle'

  // Poll progress when scan is running
  useEffect(() => {
    if (!activeScanId || scanStatus !== 'running') return

    const interval = setInterval(() => {
      loadProgress(activeScanId)
      loadFindings(activeScanId)
    }, 1000)

    return () => clearInterval(interval)
  }, [activeScanId, scanStatus])

  // Add target handler
  const handleAddTarget = async () => {
    if (!newTargetUrl.trim()) return

    try {
      const url = new URL(newTargetUrl)
      await addTarget({
        url: url.toString(),
        method: newTargetMethod,
        headers: {},
        body: '',
      })
      setNewTargetUrl('')
    } catch {
      // Invalid URL, still add it
      await addTarget({
        url: newTargetUrl,
        method: newTargetMethod,
        headers: {},
        body: '',
      })
      setNewTargetUrl('')
    }
  }

  // Start scan handler
  const handleStartScan = async () => {
    if (targets.length === 0) return

    const enabledPlugins = plugins.filter(p => p.enabled).map(p => p.id)
    const scanId = await createScan({
      id: `scan-${Date.now()}`,
      name: `Scan ${Object.keys(progress).length + 1}`,
      concurrency,
      rateLimit,
      timeout,
      followRedirects: true,
      enabledPlugins,
    })

    setActiveScan(scanId)
    await startScan(scanId)
  }

  // Pause/Resume handler
  const handlePauseResume = async () => {
    if (!activeScanId) return
    if (scanStatus === 'running') {
      await pauseScan(activeScanId)
    } else if (scanStatus === 'paused') {
      await resumeScan(activeScanId)
    }
  }

  // Stop scan handler
  const handleStopScan = async () => {
    if (!activeScanId) return
    await stopScan(activeScanId)
  }

  // Filter findings
  const filteredFindings = currentFindings.filter(f => {
    if (severityFilter === 'all') return true
    return f.severity === severityFilter
  })

  // Calculate progress percentage
  const progressPercent = currentProgress && currentProgress.totalRequests > 0
    ? Math.round((currentProgress.completedReqs / currentProgress.totalRequests) * 100)
    : 0

  return (
    <div className="flex h-full">
      {/* Left Panel - Configuration */}
      <div className="w-1/3 border-r border-gray-200 flex flex-col overflow-hidden">
        {/* Header */}
        <div className="p-4 border-b border-gray-200 bg-gray-50">
          <h3 className="font-semibold flex items-center gap-2">
            <Target className="w-4 h-4" />
            Active Scan
          </h3>
          <p className="text-xs text-gray-500 mt-1">
            Configure and run active vulnerability scans
          </p>
        </div>

        {/* Target Input */}
        <div className="p-4 border-b border-gray-200">
          <Label className="text-xs font-medium mb-2 block">Add Target URL</Label>
          <div className="flex gap-2">
            <select
              value={newTargetMethod}
              onChange={(e) => setNewTargetMethod(e.target.value)}
              className="h-8 px-2 border rounded text-xs"
            >
              <option>GET</option>
              <option>POST</option>
              <option>PUT</option>
              <option>DELETE</option>
              <option>PATCH</option>
            </select>
            <Input
              value={newTargetUrl}
              onChange={(e) => setNewTargetUrl(e.target.value)}
              placeholder="https://example.com/api"
              className="h-8 text-xs flex-1"
              onKeyDown={(e) => e.key === 'Enter' && handleAddTarget()}
            />
            <Button size="sm" onClick={handleAddTarget} className="h-8">
              <Plus className="w-4 h-4" />
            </Button>
          </div>
        </div>

        {/* Target List */}
        <div className="flex-1 overflow-auto p-4">
          <div className="flex items-center justify-between mb-2">
            <Label className="text-xs font-medium">Targets ({targets.length})</Label>
            {targets.length > 0 && (
              <Button variant="ghost" size="sm" onClick={clearTargets} className="h-6 text-xs">
                Clear All
              </Button>
            )}
          </div>

          {targets.length === 0 ? (
            <div className="text-center py-8 text-gray-400">
              <Target className="w-8 h-8 mx-auto mb-2" />
              <p className="text-xs">No targets added</p>
            </div>
          ) : (
            <div className="space-y-2">
              {targets.map((target) => (
                <div
                  key={target.id}
                  className="p-2 border rounded-lg bg-white text-xs group"
                >
                  <div className="flex items-center justify-between">
                    <div className="flex items-center gap-2">
                      <Badge variant="outline" className="text-xs">{target.method}</Badge>
                      <span className="truncate max-w-[180px]">{target.url}</span>
                    </div>
                    <Button
                      variant="ghost"
                      size="sm"
                      onClick={() => removeTarget(target.id)}
                      className="h-6 w-6 p-0 opacity-0 group-hover:opacity-100"
                    >
                      <X className="w-3 h-3" />
                    </Button>
                  </div>
                </div>
              ))}
            </div>
          )}
        </div>

        {/* Plugin Selection */}
        <div className="p-4 border-t border-gray-200">
          <Label className="text-xs font-medium mb-2 block">Scan Plugins</Label>
          <div className="space-y-2 max-h-40 overflow-auto">
            {plugins.map((plugin) => (
              <div key={plugin.id} className="flex items-center gap-2">
                <Checkbox
                  id={plugin.id}
                  checked={plugin.enabled}
                  onCheckedChange={() => togglePlugin(plugin.id)}
                />
                <Label htmlFor={plugin.id} className="text-xs cursor-pointer flex-1">
                  {plugin.name}
                </Label>
                <Badge className={cn('text-xs', severityColors[plugin.severity])}>
                  {plugin.severity}
                </Badge>
              </div>
            ))}
          </div>
        </div>

        {/* Advanced Settings */}
        <div className="p-4 border-t border-gray-200">
          <button
            onClick={() => setShowAdvanced(!showAdvanced)}
            className="flex items-center gap-1 text-xs text-gray-500 hover:text-gray-700"
          >
            {showAdvanced ? <ChevronUp className="w-3 h-3" /> : <ChevronDown className="w-3 h-3" />}
            Advanced Settings
          </button>

          {showAdvanced && (
            <div className="mt-3 space-y-3">
              <div className="grid grid-cols-3 gap-2">
                <div>
                  <Label className="text-xs text-gray-500">Concurrency</Label>
                  <Input
                    type="number"
                    value={concurrency}
                    onChange={(e) => setConcurrency(parseInt(e.target.value) || 5)}
                    className="h-7 text-xs"
                  />
                </div>
                <div>
                  <Label className="text-xs text-gray-500">Rate Limit</Label>
                  <Input
                    type="number"
                    value={rateLimit}
                    onChange={(e) => setRateLimit(parseInt(e.target.value) || 10)}
                    className="h-7 text-xs"
                  />
                </div>
                <div>
                  <Label className="text-xs text-gray-500">Timeout</Label>
                  <Input
                    type="number"
                    value={timeout}
                    onChange={(e) => setTimeout_(parseInt(e.target.value) || 30)}
                    className="h-7 text-xs"
                  />
                </div>
              </div>
            </div>
          )}
        </div>

        {/* Action Buttons */}
        <div className="p-4 border-t border-gray-200 bg-white">
          <div className="flex gap-2">
            {scanStatus === 'idle' || scanStatus === 'completed' || scanStatus === 'cancelled' ? (
              <Button
                onClick={handleStartScan}
                disabled={targets.length === 0 || isLoading}
                className="flex-1"
              >
                <Play className="w-4 h-4 mr-1" />
                Start Scan
              </Button>
            ) : (
              <>
                <Button
                  variant="outline"
                  onClick={handlePauseResume}
                  disabled={isLoading}
                  className="flex-1"
                >
                  {scanStatus === 'running' ? (
                    <>
                      <Pause className="w-4 h-4 mr-1" />
                      Pause
                    </>
                  ) : (
                    <>
                      <Play className="w-4 h-4 mr-1" />
                      Resume
                    </>
                  )}
                </Button>
                <Button
                  variant="destructive"
                  onClick={handleStopScan}
                  disabled={isLoading}
                >
                  <Square className="w-4 h-4" />
                </Button>
              </>
            )}
          </div>
        </div>
      </div>

      {/* Right Panel - Results */}
      <div className="flex-1 flex flex-col overflow-hidden">
        {/* Progress Bar */}
        {currentProgress && scanStatus === 'running' && (
          <div className="p-4 border-b border-gray-200 bg-gray-50">
            <div className="flex items-center justify-between mb-2">
              <div className="flex items-center gap-2">
                <RefreshCw className="w-4 h-4 animate-spin text-blue-500" />
                <span className="text-sm font-medium">Scanning...</span>
              </div>
              <div className="text-xs text-gray-500">
                {currentProgress.completedReqs} / {currentProgress.totalRequests} requests
              </div>
            </div>
            <Progress value={progressPercent} className="h-2" />
            <div className="flex items-center justify-between mt-2 text-xs text-gray-500">
              <span>Target: {currentProgress.currentTarget || 'N/A'}</span>
              <span>{currentProgress.requestsPerSec.toFixed(1)} req/s</span>
              <span>Findings: {currentProgress.findingsCount}</span>
            </div>
          </div>
        )}

        {/* Status Badge */}
        {currentProgress && (
          <div className="px-4 py-2 border-b border-gray-200">
            <Badge
              className={cn(
                scanStatus === 'running' && 'bg-blue-500',
                scanStatus === 'paused' && 'bg-yellow-500',
                scanStatus === 'completed' && 'bg-green-500',
                scanStatus === 'error' && 'bg-red-500',
                scanStatus === 'cancelled' && 'bg-gray-500',
              )}
            >
              {scanStatus.toUpperCase()}
            </Badge>
          </div>
        )}

        {/* Findings Filter */}
        {currentFindings.length > 0 && (
          <div className="px-4 py-2 border-b border-gray-200 flex items-center gap-2">
            <Filter className="w-4 h-4 text-gray-400" />
            <div className="flex gap-1">
              <Button
                variant={severityFilter === 'all' ? 'default' : 'outline'}
                size="sm"
                onClick={() => setSeverityFilter('all')}
                className="h-6 text-xs"
              >
                All
              </Button>
              {['critical', 'high', 'medium', 'low'].map((sev) => (
                <Button
                  key={sev}
                  variant={severityFilter === sev ? 'default' : 'outline'}
                  size="sm"
                  onClick={() => setSeverityFilter(sev)}
                  className="h-6 text-xs"
                >
                  {sev}
                </Button>
              ))}
            </div>
          </div>
        )}

        {/* Findings List */}
        <div className="flex-1 overflow-auto">
          {filteredFindings.length === 0 ? (
            <div className="flex flex-col items-center justify-center h-full text-gray-400">
              {scanStatus === 'running' ? (
                <>
                  <RefreshCw className="w-8 h-8 mb-2 animate-spin" />
                  <p className="text-sm">Scanning for vulnerabilities...</p>
                </>
              ) : (
                <>
                  <CheckCircle className="w-8 h-8 mb-2" />
                  <p className="text-sm">No findings to display</p>
                </>
              )}
            </div>
          ) : (
            <div className="divide-y divide-gray-100">
              {filteredFindings.map((finding) => (
                <div
                  key={finding.id}
                  onClick={() => selectFinding(finding)}
                  className={cn(
                    'p-3 cursor-pointer hover:bg-gray-50 transition-colors',
                    selectedFinding?.id === finding.id && 'bg-blue-50',
                    'border-l-4',
                    severityBorderColors[finding.severity]
                  )}
                >
                  <div className="flex items-start gap-2">
                    <AlertTriangle className={cn(
                      'w-4 h-4 mt-0.5',
                      finding.severity === 'critical' && 'text-purple-500',
                      finding.severity === 'high' && 'text-red-500',
                      finding.severity === 'medium' && 'text-orange-500',
                      finding.severity === 'low' && 'text-yellow-500',
                    )} />
                    <div className="flex-1 min-w-0">
                      <div className="flex items-center gap-2">
                        <span className="font-medium text-sm truncate">{finding.title}</span>
                        <Badge className={cn('text-xs', severityColors[finding.severity])}>
                          {finding.severity}
                        </Badge>
                      </div>
                      <p className="text-xs text-gray-500 truncate">{finding.url}</p>
                      <div className="flex items-center gap-2 mt-1 text-xs text-gray-400">
                        <span>{finding.pluginName}</span>
                        <span>•</span>
                        <span>Confidence: {finding.confidence}%</span>
                      </div>
                    </div>
                  </div>
                </div>
              ))}
            </div>
          )}
        </div>

        {/* Finding Detail Panel */}
        {selectedFinding && (
          <div className="h-64 border-t border-gray-200 overflow-auto">
            <div className="p-4">
              <div className="flex items-center justify-between mb-3">
                <h4 className="font-medium">{selectedFinding.title}</h4>
                <Button
                  variant="ghost"
                  size="sm"
                  onClick={() => selectFinding(null)}
                  className="h-6 w-6 p-0"
                >
                  <X className="w-4 h-4" />
                </Button>
              </div>

              <div className="space-y-3 text-sm">
                <div>
                  <Label className="text-xs text-gray-500">Description</Label>
                  <p className="mt-1">{selectedFinding.description}</p>
                </div>

                <div>
                  <Label className="text-xs text-gray-500">Evidence</Label>
                  <code className="block mt-1 p-2 bg-gray-100 rounded text-xs overflow-x-auto">
                    {selectedFinding.evidence}
                  </code>
                </div>

                <div>
                  <Label className="text-xs text-gray-500">Payload</Label>
                  <code className="block mt-1 p-2 bg-gray-100 rounded text-xs overflow-x-auto">
                    {selectedFinding.payload}
                  </code>
                </div>
              </div>
            </div>
          </div>
        )}
      </div>
    </div>
  )
}

export default ActiveScanTab
