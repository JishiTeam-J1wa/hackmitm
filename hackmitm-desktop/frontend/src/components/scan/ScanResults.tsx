import { useState, useMemo } from 'react'
import {
  Search,
  AlertTriangle,
  AlertCircle,
  Info,
  CheckCircle,
  XCircle,
  Copy,
  ExternalLink,
  Trash2,
  Eye,
} from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Badge } from '@/components/ui/badge'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { Textarea } from '@/components/ui/textarea'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { cn } from '@/lib/utils'
import { useScanStore } from '@/store'
import type { ScanResult } from '@/types'
import { ResizablePanelGroup } from '@/components/ui/resizable'

const severityConfig = {
  critical: {
    icon: AlertTriangle,
    color: 'text-red-600',
    bgColor: 'bg-red-50',
    badgeColor: 'bg-red-500 text-white',
  },
  high: {
    icon: AlertCircle,
    color: 'text-orange-600',
    bgColor: 'bg-orange-50',
    badgeColor: 'bg-orange-500 text-white',
  },
  medium: {
    icon: AlertCircle,
    color: 'text-yellow-600',
    bgColor: 'bg-yellow-50',
    badgeColor: 'bg-yellow-500 text-white',
  },
  low: {
    icon: Info,
    color: 'text-blue-600',
    bgColor: 'bg-blue-50',
    badgeColor: 'bg-blue-500 text-white',
  },
  info: {
    icon: Info,
    color: 'text-gray-600',
    bgColor: 'bg-gray-50',
    badgeColor: 'bg-gray-500 text-white',
  },
}

export function ScanResults() {
  const {
    selectedResult,
    filters,
    selectResult,
    setFilters,
    markFalsePositive,
    deleteResult,
    clearResults,
    getFilteredResults,
  } = useScanStore()

  const [localSearch, setLocalSearch] = useState(filters.search)

  const filteredResults = useMemo(() => {
    let result = getFilteredResults()

    if (localSearch) {
      const search = localSearch.toLowerCase()
      result = result.filter(
        (r) =>
          r.title.toLowerCase().includes(search) ||
          r.url.toLowerCase().includes(search) ||
          r.pluginName.toLowerCase().includes(search)
      )
    }

    return result
  }, [getFilteredResults, localSearch])

  const handleMarkFalsePositive = (id: string, isFalsePositive: boolean) => {
    markFalsePositive(id, isFalsePositive)
  }

  const handleDelete = (id: string) => {
    deleteResult(id)
    if (selectedResult?.id === id) {
      selectResult(null)
    }
  }

  const handleExportToVuln = (result: ScanResult) => {
    // TODO: Implement export to vulnerability
    console.log('Export to vuln:', result)
  }

  const getSeverityBadge = (severity: string) => {
    const config = severityConfig[severity as keyof typeof severityConfig] || severityConfig.info
    return (
      <Badge className={cn('text-[10px]', config.badgeColor)}>
        {severity}
      </Badge>
    )
  }

  const ResultDetail = ({ result }: { result: ScanResult }) => {
    const severity = severityConfig[result.severity as keyof typeof severityConfig] || severityConfig.info
    const SeverityIcon = severity.icon

    return (
      <div className="flex flex-col h-full bg-white">
        {/* Header */}
        <div className={cn('border-b p-3', severity.bgColor)}>
          <div className="flex items-start gap-2">
            <SeverityIcon className={cn('w-5 h-5 mt-0.5', severity.color)} />
            <div className="flex-1 min-w-0">
              <h4 className="text-sm font-medium text-gray-900">{result.title}</h4>
              <div className="flex items-center gap-2 mt-1">
                {getSeverityBadge(result.severity)}
                <Badge variant="outline" className="text-[10px]">{result.pluginName}</Badge>
                {result.falsePositive && (
                  <Badge variant="outline" className="text-[10px] text-gray-500">
                    False Positive
                  </Badge>
                )}
              </div>
            </div>
            <Button variant="ghost" size="icon" onClick={() => selectResult(null)} className="w-7 h-7">
              <XCircle className="w-4 h-4" />
            </Button>
          </div>

          {/* Actions */}
          <div className="flex items-center gap-2 mt-3">
            <Button
              size="sm"
              variant="outline"
              onClick={() => handleMarkFalsePositive(result.id, !result.falsePositive)}
              className="h-6 text-xs"
            >
              {result.falsePositive ? (
                <>
                  <CheckCircle className="w-3 h-3 mr-1 text-green-500" />
                  Unmark FP
                </>
              ) : (
                <>
                  <XCircle className="w-3 h-3 mr-1" />
                  Mark FP
                </>
              )}
            </Button>
            <Button
              size="sm"
              variant="outline"
              onClick={() => handleExportToVuln(result)}
              className="h-6 text-xs"
            >
              <ExternalLink className="w-3 h-3 mr-1" />
              Export to Vuln
            </Button>
            <Button
              size="sm"
              variant="destructive"
              onClick={() => handleDelete(result.id)}
              className="h-6 text-xs"
            >
              <Trash2 className="w-3 h-3 mr-1" />
              Delete
            </Button>
          </div>
        </div>

        {/* Content */}
        <div className="flex-1 overflow-hidden">
          <Tabs defaultValue="description" className="h-full flex flex-col">
            <TabsList className="px-3 pt-2 flex-shrink-0">
              <TabsTrigger value="description" className="text-xs">Description</TabsTrigger>
              <TabsTrigger value="evidence" className="text-xs">Evidence</TabsTrigger>
              <TabsTrigger value="request" className="text-xs">Request</TabsTrigger>
              <TabsTrigger value="response" className="text-xs">Response</TabsTrigger>
            </TabsList>

            <TabsContent value="description" className="flex-1 p-3 overflow-auto">
              <p className="text-sm text-gray-600">{result.description}</p>
              <div className="mt-3 text-xs text-gray-500">
                <div><strong>URL:</strong> {result.url}</div>
                <div><strong>Method:</strong> {result.method}</div>
                <div><strong>Time:</strong> {new Date(result.timestamp).toLocaleString('zh-CN')}</div>
              </div>
            </TabsContent>

            <TabsContent value="evidence" className="flex-1 p-3 overflow-auto">
              <div className="bg-yellow-50 border border-yellow-200 rounded p-3">
                <pre className="text-xs font-mono whitespace-pre-wrap break-all">
                  {result.evidence}
                </pre>
              </div>
            </TabsContent>

            <TabsContent value="request" className="flex-1 overflow-hidden flex flex-col">
              <div className="flex items-center justify-between px-3 py-1.5 border-b bg-gray-50">
                <span className="text-xs text-gray-500">Request</span>
                <Button
                  variant="ghost"
                  size="sm"
                  onClick={() => navigator.clipboard.writeText(result.request)}
                  className="h-5 text-xs"
                >
                  <Copy className="w-3 h-3 mr-1" />
                  Copy
                </Button>
              </div>
              <Textarea
                readOnly
                value={result.request}
                className="flex-1 font-mono text-xs border-0 rounded-none resize-none focus-visible:ring-0"
              />
            </TabsContent>

            <TabsContent value="response" className="flex-1 overflow-hidden flex flex-col">
              <div className="flex items-center justify-between px-3 py-1.5 border-b bg-gray-50">
                <span className="text-xs text-gray-500">Response</span>
                <Button
                  variant="ghost"
                  size="sm"
                  onClick={() => navigator.clipboard.writeText(result.response)}
                  className="h-5 text-xs"
                >
                  <Copy className="w-3 h-3 mr-1" />
                  Copy
                </Button>
              </div>
              <Textarea
                readOnly
                value={result.response}
                className="flex-1 font-mono text-xs border-0 rounded-none resize-none focus-visible:ring-0"
              />
            </TabsContent>
          </Tabs>
        </div>
      </div>
    )
  }

  return (
    <div className="flex h-full flex-col">
      {/* Toolbar */}
      <div className="h-9 border-b border-gray-200 bg-white px-3 flex items-center gap-2 flex-shrink-0">
        {/* Search */}
        <div className="relative flex-1 max-w-xs">
          <Search className="absolute left-2 top-1/2 -translate-y-1/2 w-3.5 h-3.5 text-gray-400" />
          <Input
            value={localSearch}
            onChange={(e) => setLocalSearch(e.target.value)}
            placeholder="Search results..."
            className="h-7 pl-7 text-xs"
          />
        </div>

        {/* Severity filter */}
        <Select value={filters.severity} onValueChange={(v) => setFilters({ severity: v })}>
          <SelectTrigger className="w-24 h-7 text-xs">
            <SelectValue />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="all">All</SelectItem>
            <SelectItem value="critical">Critical</SelectItem>
            <SelectItem value="high">High</SelectItem>
            <SelectItem value="medium">Medium</SelectItem>
            <SelectItem value="low">Low</SelectItem>
            <SelectItem value="info">Info</SelectItem>
          </SelectContent>
        </Select>

        {/* False positive filter */}
        <Select value={filters.falsePositive} onValueChange={(v) => setFilters({ falsePositive: v })}>
          <SelectTrigger className="w-24 h-7 text-xs">
            <SelectValue />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="all">All</SelectItem>
            <SelectItem value="false">Not FP</SelectItem>
            <SelectItem value="true">FP Only</SelectItem>
          </SelectContent>
        </Select>

        <div className="flex-1" />

        <Badge variant="secondary" className="text-xs">
          {filteredResults.length} results
        </Badge>

        <Button
          variant="ghost"
          size="sm"
          onClick={() => { if (confirm('Clear all results?')) clearResults() }}
          className="h-7 text-xs text-red-600 hover:text-red-700"
        >
          <Trash2 className="w-3.5 h-3.5 mr-1" />
          Clear All
        </Button>
      </div>

      {/* Main content */}
      <div className="flex-1 overflow-hidden">
        {selectedResult ? (
          <ResizablePanelGroup direction="vertical" defaultSizes={[40, 60]} minSizes={[80, 100]}>
            {/* Results list */}
            <div className="flex flex-col h-full">
              <div className="h-6 border-b border-gray-200 bg-gray-50 flex items-center px-2 text-[10px] font-medium text-gray-500 flex-shrink-0">
                <div className="w-12">Severity</div>
                <div className="w-32">Plugin</div>
                <div className="flex-1 min-w-0">Title</div>
                <div className="w-20 text-center">Time</div>
              </div>

              <div className="flex-1 overflow-auto">
                {filteredResults.map((result) => (
                  <div
                    key={result.id}
                    onClick={() => selectResult(result)}
                    className={cn(
                        'flex items-center px-2 h-6 text-[11px] cursor-pointer border-b border-gray-50',
                        selectedResult?.id === result.id
                          ? 'bg-blue-50 border-l-2 border-l-blue-500 pl-[6px]'
                          : 'hover:bg-gray-50 border-l-2 border-l-transparent',
                        result.falsePositive && 'opacity-50'
                      )}
                    >
                      <div className="w-12">
                        {getSeverityBadge(result.severity)}
                      </div>
                      <div className="w-32 truncate text-gray-600">{result.pluginName}</div>
                      <div className="flex-1 min-w-0 truncate font-medium">{result.title}</div>
                      <div className="w-20 text-center text-gray-400 text-[10px]">
                        {new Date(result.timestamp).toLocaleTimeString('zh-CN', {
                          hour: '2-digit',
                          minute: '2-digit',
                        })}
                      </div>
                    </div>
                  ))}
              </div>
            </div>

            {/* Detail panel */}
            <ResultDetail result={selectedResult} />
          </ResizablePanelGroup>
        ) : (
          // No selection - full list
          <div className="flex flex-col h-full">
            <div className="h-6 border-b border-gray-200 bg-gray-50 flex items-center px-2 text-[10px] font-medium text-gray-500 flex-shrink-0">
              <div className="w-12">Severity</div>
              <div className="w-32">Plugin</div>
              <div className="flex-1 min-w-0">Title / URL</div>
              <div className="w-20 text-center">Time</div>
            </div>

            <div className="flex-1 overflow-auto">
              {filteredResults.length === 0 ? (
                <div className="flex flex-col items-center justify-center h-full text-gray-400 py-20">
                  <Eye className="w-10 h-10 mb-2" />
                  <p className="text-sm">No scan results</p>
                  <p className="text-xs mt-1">Results will appear here when vulnerabilities are found</p>
                </div>
              ) : (
                filteredResults.map((result) => (
                  <div
                    key={result.id}
                    onClick={() => selectResult(result)}
                    className={cn(
                      'flex items-center px-2 h-6 text-[11px] cursor-pointer border-b border-gray-50',
                      'hover:bg-gray-50 border-l-2 border-l-transparent',
                      result.falsePositive && 'opacity-50'
                    )}
                  >
                      <div className="w-12">
                        {getSeverityBadge(result.severity)}
                      </div>
                      <div className="w-32 truncate text-gray-600">{result.pluginName}</div>
                      <div className="flex-1 min-w-0">
                        <span className="font-medium truncate">{result.title}</span>
                        <span className="text-gray-400 ml-2 text-[10px] truncate">{result.url}</span>
                      </div>
                      <div className="w-20 text-center text-gray-400 text-[10px]">
                        {new Date(result.timestamp).toLocaleTimeString('zh-CN', {
                          hour: '2-digit',
                          minute: '2-digit',
                        })}
                      </div>
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
