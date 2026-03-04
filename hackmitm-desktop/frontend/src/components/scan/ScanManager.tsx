import { useState, useEffect } from 'react'
import {
  Pause,
  Plus,
  X,
  Globe,
  Zap,
  Clock,
  Activity,
  Shield,
  RefreshCw,
  AlertTriangle,
} from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Badge } from '@/components/ui/badge'
import { Switch } from '@/components/ui/switch'
import { Label } from '@/components/ui/label'
import { cn } from '@/lib/utils'
import { useScanStore } from '@/store'
import { useProxyStore } from '@/store'

// Pattern type colors
const patternColors: Record<string, string> = {
  api: 'bg-blue-500',
  webpage: 'bg-green-500',
  download: 'bg-purple-500',
  bot: 'bg-yellow-500',
  attack: 'bg-red-500',
  static: 'bg-gray-500',
  auth: 'bg-orange-500',
  admin: 'bg-pink-500',
  default: 'bg-gray-400',
}

export function ScanManager() {
  const {
    isScanning,
    isLoading,
    config,
    stats,
    trafficPatterns,
    setEnabled,
    updateConfig,
    loadTrafficPatterns,
    loadStats,
  } = useScanStore()

  const { connected } = useProxyStore()

  const [newInclude, setNewInclude] = useState('')
  const [newExclude, setNewExclude] = useState('')

  // Load data on mount and when connected
  useEffect(() => {
    if (connected) {
      loadStats()
    }
  }, [connected])

  // Refresh patterns periodically
  useEffect(() => {
    if (!connected || !isScanning) return
    const interval = setInterval(() => {
      loadTrafficPatterns()
    }, 5000)
    return () => clearInterval(interval)
  }, [connected, isScanning])

  const handleAddInclude = () => {
    if (newInclude.trim()) {
      updateConfig({
        includePatterns: [...config.includePatterns, newInclude.trim()],
      })
      setNewInclude('')
    }
  }

  const handleAddExclude = () => {
    if (newExclude.trim()) {
      updateConfig({
        excludePatterns: [...config.excludePatterns, newExclude.trim()],
      })
      setNewExclude('')
    }
  }

  const handleRemoveInclude = (index: number) => {
    updateConfig({
      includePatterns: config.includePatterns.filter((_, i) => i !== index),
    })
  }

  const handleRemoveExclude = (index: number) => {
    updateConfig({
      excludePatterns: config.excludePatterns.filter((_, i) => i !== index),
    })
  }

  const handleRefresh = () => {
    loadStats()
  }

  // Get attack count from traffic patterns
  const attackCount = trafficPatterns.find(p => p.type === 'attack')?.count || stats.highCount
  const hasAttacks = attackCount > 0

  return (
    <div className="flex h-full flex-col overflow-auto">
      {/* Main toggle */}
      <div className="p-4 border-b border-gray-200 bg-white">
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-3">
            <div
              className={cn(
                'w-10 h-10 rounded-full flex items-center justify-center',
                isScanning ? (hasAttacks ? 'bg-red-100' : 'bg-green-100') : 'bg-gray-100'
              )}
            >
              {hasAttacks && isScanning ? (
                <AlertTriangle className="w-5 h-5 text-red-600" />
              ) : isScanning ? (
                <Zap className="w-5 h-5 text-green-600" />
              ) : (
                <Pause className="w-5 h-5 text-gray-400" />
              )}
            </div>
            <div>
              <h3 className="text-sm font-medium text-gray-900">Passive Scan</h3>
              <p className="text-xs text-gray-500">
                {isScanning
                  ? hasAttacks
                    ? `Detected ${attackCount} potential threats`
                    : 'Scanning traffic for vulnerabilities'
                  : 'Scanning is paused'}
              </p>
            </div>
          </div>
          <div className="flex items-center gap-3">
            <Switch checked={isScanning} onCheckedChange={setEnabled} />
            <Badge
              className={cn(
                'text-xs',
                isScanning
                  ? hasAttacks
                    ? 'bg-red-500 text-white'
                    : 'bg-green-500 text-white'
                  : 'bg-gray-200 text-gray-600'
              )}
            >
              {isScanning ? (hasAttacks ? 'Alerts' : 'Running') : 'Stopped'}
            </Badge>
          </div>
        </div>
      </div>

      {/* Statistics */}
      <div className="p-4 border-b border-gray-200 bg-gray-50">
        <div className="flex items-center justify-between mb-3">
          <h4 className="text-xs font-medium text-gray-500 flex items-center gap-1">
            <Activity className="w-3.5 h-3.5" />
            Real-time Statistics
          </h4>
          <Button
            variant="ghost"
            size="sm"
            onClick={handleRefresh}
            disabled={isLoading}
            className="h-6 text-xs"
          >
            <RefreshCw className={cn('w-3 h-3 mr-1', isLoading && 'animate-spin')} />
            Refresh
          </Button>
        </div>

        <div className="grid grid-cols-4 gap-3">
          <div className="bg-white rounded-lg border p-3">
            <div className="text-2xl font-semibold text-gray-900">{stats.totalScanned || trafficPatterns.reduce((sum, p) => sum + p.count, 0)}</div>
            <div className="text-xs text-gray-500">Total Scanned</div>
          </div>
          <div className="bg-white rounded-lg border p-3">
            <div className={cn('text-2xl font-semibold', hasAttacks ? 'text-red-600' : 'text-orange-600')}>
              {stats.totalFindings || attackCount}
            </div>
            <div className="text-xs text-gray-500">Total Findings</div>
          </div>
          <div className="bg-white rounded-lg border p-3">
            <div className={cn('text-2xl font-semibold', hasAttacks ? 'text-red-600' : 'text-gray-600')}>
              {stats.criticalCount + stats.highCount || attackCount}
            </div>
            <div className="text-xs text-gray-500">Critical/High</div>
          </div>
          <div className="bg-white rounded-lg border p-3">
            <div className="text-2xl font-semibold text-blue-600">
              {stats.mediumCount + stats.lowCount}
            </div>
            <div className="text-xs text-gray-500">Medium/Low</div>
          </div>
        </div>

        {/* Traffic Patterns */}
        {trafficPatterns.length > 0 && (
          <div className="mt-4">
            <h4 className="text-xs font-medium text-gray-500 mb-2">Traffic Patterns</h4>
            <div className="flex flex-wrap gap-2">
              {trafficPatterns.map((pattern) => (
                <Badge
                  key={pattern.type}
                  variant="outline"
                  className={cn(
                    'text-xs',
                    pattern.type === 'attack' && 'border-red-500 text-red-700 bg-red-50'
                  )}
                >
                  <div
                    className={cn(
                      'w-2 h-2 rounded-full mr-1.5',
                      patternColors[pattern.type] || patternColors.default
                    )}
                  />
                  {pattern.type}: {pattern.count}
                </Badge>
              ))}
            </div>
          </div>
        )}
      </div>

      {/* Settings */}
      <div className="flex-1 p-4 overflow-auto">
        <div className="space-y-6 max-w-2xl">
          {/* Scan scope */}
          <div className="bg-white rounded-lg border p-4">
            <h4 className="text-sm font-medium text-gray-900 mb-4 flex items-center gap-2">
              <Globe className="w-4 h-4" />
              Scan Scope
            </h4>

            {/* Include patterns */}
            <div className="mb-4">
              <Label className="text-xs text-gray-600 mb-2 block">Include Patterns</Label>
              <div className="flex gap-2 mb-2">
                <Input
                  value={newInclude}
                  onChange={(e) => setNewInclude(e.target.value)}
                  placeholder="e.g., *.example.com"
                  className="h-8 text-xs"
                  onKeyDown={(e) => e.key === 'Enter' && handleAddInclude()}
                />
                <Button size="sm" onClick={handleAddInclude} className="h-8">
                  <Plus className="w-4 h-4" />
                </Button>
              </div>
              <div className="flex flex-wrap gap-1">
                {config.includePatterns.map((pattern, index) => (
                  <Badge key={index} variant="secondary" className="text-xs">
                    {pattern}
                    <button
                      onClick={() => handleRemoveInclude(index)}
                      className="ml-1 hover:text-red-500"
                    >
                      <X className="w-3 h-3" />
                    </button>
                  </Badge>
                ))}
              </div>
            </div>

            {/* Exclude patterns */}
            <div>
              <Label className="text-xs text-gray-600 mb-2 block">Exclude Patterns</Label>
              <div className="flex gap-2 mb-2">
                <Input
                  value={newExclude}
                  onChange={(e) => setNewExclude(e.target.value)}
                  placeholder="e.g., *.js, *.css"
                  className="h-8 text-xs"
                  onKeyDown={(e) => e.key === 'Enter' && handleAddExclude()}
                />
                <Button size="sm" onClick={handleAddExclude} className="h-8">
                  <Plus className="w-4 h-4" />
                </Button>
              </div>
              <div className="flex flex-wrap gap-1">
                {config.excludePatterns.map((pattern, index) => (
                  <Badge key={index} variant="outline" className="text-xs">
                    {pattern}
                    <button
                      onClick={() => handleRemoveExclude(index)}
                      className="ml-1 hover:text-red-500"
                    >
                      <X className="w-3 h-3" />
                    </button>
                  </Badge>
                ))}
              </div>
            </div>
          </div>

          {/* Rate limiting */}
          <div className="bg-white rounded-lg border p-4">
            <h4 className="text-sm font-medium text-gray-900 mb-4 flex items-center gap-2">
              <Clock className="w-4 h-4" />
              Rate Limiting
            </h4>

            <div className="grid grid-cols-2 gap-4">
              <div>
                <Label className="text-xs text-gray-600 mb-2 block">
                  Max Requests/Second
                </Label>
                <Input
                  type="number"
                  value={config.maxRequestsPerSecond}
                  onChange={(e) =>
                    updateConfig({ maxRequestsPerSecond: parseInt(e.target.value) || 100 })
                  }
                  className="h-8 text-xs"
                />
              </div>
              <div>
                <Label className="text-xs text-gray-600 mb-2 block">
                  Timeout (seconds)
                </Label>
                <Input
                  type="number"
                  value={config.timeout}
                  onChange={(e) =>
                    updateConfig({ timeout: parseInt(e.target.value) || 30 })
                  }
                  className="h-8 text-xs"
                />
              </div>
            </div>
          </div>

          {/* Info */}
          <div className="bg-blue-50 border border-blue-200 rounded-lg p-4">
            <div className="flex items-start gap-3">
              <Shield className="w-5 h-5 text-blue-500 mt-0.5" />
              <div>
                <h4 className="text-sm font-medium text-blue-800">About Passive Scanning</h4>
                <p className="text-xs text-blue-600 mt-1">
                  Passive scanning analyzes HTTP traffic in real-time without sending additional
                  requests. It can detect various security issues like information disclosure,
                  missing security headers, and potential injection points.
                </p>
                <p className="text-xs text-blue-600 mt-2">
                  <strong>Backend Integration:</strong> This component uses the HackMITM backend's
                  traffic pattern recognition and security plugin to detect potential threats.
                </p>
              </div>
            </div>
          </div>

          {/* Connection Status */}
          {!connected && (
            <div className="bg-yellow-50 border border-yellow-200 rounded-lg p-4">
              <div className="flex items-start gap-3">
                <AlertTriangle className="w-5 h-5 text-yellow-600 mt-0.5" />
                <div>
                  <h4 className="text-sm font-medium text-yellow-800">Not Connected</h4>
                  <p className="text-xs text-yellow-600 mt-1">
                    Connect to the HackMITM service to start scanning traffic.
                  </p>
                </div>
              </div>
            </div>
          )}
        </div>
      </div>
    </div>
  )
}
