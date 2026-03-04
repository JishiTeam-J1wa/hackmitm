import { useState } from 'react'
import {
  FileText,
  Download,
  FileJson,
  FileCode,
  File,
  Settings,
  Eye,
  Loader2,
  CheckCircle,
} from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Badge } from '@/components/ui/badge'
import { Card, CardContent } from '@/components/ui/card'
import { Separator } from '@/components/ui/separator'
import { cn } from '@/lib/utils'

interface ReportExportProps {
  sessionId?: string
  onExport?: (format: string, options: ExportOptions) => Promise<void>
}

interface ExportOptions {
  title: string
  severity: string[]
  status: string[]
}

const formatOptions = [
  {
    id: 'html',
    name: 'HTML',
    icon: FileCode,
    description: 'Interactive HTML report with styling',
    extension: '.html',
  },
  {
    id: 'json',
    name: 'JSON',
    icon: FileJson,
    description: 'Raw data in JSON format',
    extension: '.json',
  },
  {
    id: 'markdown',
    name: 'Markdown',
    icon: FileText,
    description: 'Markdown format for documentation',
    extension: '.md',
  },
]

const severityOptions = [
  { id: 'critical', label: 'Critical', color: 'bg-red-500' },
  { id: 'high', label: 'High', color: 'bg-orange-500' },
  { id: 'medium', label: 'Medium', color: 'bg-yellow-500' },
  { id: 'low', label: 'Low', color: 'bg-blue-500' },
]

const statusOptions = [
  { id: 'open', label: 'Open' },
  { id: 'fixed', label: 'Fixed' },
  { id: 'ignored', label: 'Ignored' },
]

export function ReportExport({ sessionId, onExport }: ReportExportProps) {
  const [title, setTitle] = useState('Security Assessment Report')
  const [format, setFormat] = useState('html')
  const [selectedSeverities, setSelectedSeverities] = useState<string[]>([])
  const [selectedStatuses, setSelectedStatuses] = useState<string[]>([])
  const [isExporting, setIsExporting] = useState(false)
  const [exportSuccess, setExportSuccess] = useState(false)
  const [previewData, setPreviewData] = useState<string | null>(null)

  const toggleSeverity = (severity: string) => {
    setSelectedSeverities((prev) =>
      prev.includes(severity) ? prev.filter((s) => s !== severity) : [...prev, severity]
    )
  }

  const toggleStatus = (status: string) => {
    setSelectedStatuses((prev) =>
      prev.includes(status) ? prev.filter((s) => s !== status) : [...prev, status]
    )
  }

  const handleExport = async () => {
    setIsExporting(true)
    setExportSuccess(false)

    try {
      if (onExport) {
        await onExport(format, {
          title,
          severity: selectedSeverities,
          status: selectedStatuses,
        })
      } else {
        // Default export behavior - call API
        const response = await fetch('/api/reports/generate', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({
            session_id: sessionId,
            title,
            format,
            severity: selectedSeverities,
            status: selectedStatuses,
          }),
        })

        if (!response.ok) throw new Error('Export failed')

        const blob = await response.blob()
        const url = URL.createObjectURL(blob)
        const a = document.createElement('a')
        a.href = url
        a.download = `security-report${formatOptions.find((f) => f.id === format)?.extension || ''}`
        a.click()
        URL.revokeObjectURL(url)
      }

      setExportSuccess(true)
      setTimeout(() => setExportSuccess(false), 3000)
    } catch (error) {
      console.error('Export failed:', error)
    } finally {
      setIsExporting(false)
    }
  }

  const handlePreview = async () => {
    try {
      const response = await fetch('/api/reports/generate', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          session_id: sessionId,
          title,
          format: 'json',
          severity: selectedSeverities,
          status: selectedStatuses,
        }),
      })

      if (!response.ok) throw new Error('Preview failed')

      const data = await response.json()
      setPreviewData(JSON.stringify(data, null, 2))
    } catch (error) {
      console.error('Preview failed:', error)
    }
  }

  return (
    <div className="flex h-full">
      {/* Settings Panel */}
      <div className="w-80 border-r border-gray-200 bg-white overflow-auto">
        <div className="p-4 border-b border-gray-200">
          <h2 className="text-lg font-semibold text-gray-800 flex items-center gap-2">
            <Settings className="w-5 h-5" />
            Report Settings
          </h2>
        </div>

        <div className="p-4 space-y-6">
          {/* Title */}
          <div className="space-y-2">
            <label className="text-sm font-medium text-gray-700">Report Title</label>
            <Input
              value={title}
              onChange={(e) => setTitle(e.target.value)}
              placeholder="Enter report title"
              className="text-sm"
            />
          </div>

          {/* Format Selection */}
          <div className="space-y-2">
            <label className="text-sm font-medium text-gray-700">Export Format</label>
            <div className="space-y-2">
              {formatOptions.map((option) => {
                const Icon = option.icon
                return (
                  <div
                    key={option.id}
                    onClick={() => setFormat(option.id)}
                    className={cn(
                      'flex items-center gap-3 p-3 rounded-lg border cursor-pointer transition-all',
                      format === option.id
                        ? 'border-blue-500 bg-blue-50'
                        : 'border-gray-200 hover:border-gray-300 hover:bg-gray-50'
                    )}
                  >
                    <Icon
                      className={cn(
                        'w-5 h-5',
                        format === option.id ? 'text-blue-500' : 'text-gray-400'
                      )}
                    />
                    <div className="flex-1">
                      <div
                        className={cn(
                          'text-sm font-medium',
                          format === option.id ? 'text-blue-700' : 'text-gray-700'
                        )}
                      >
                        {option.name}
                      </div>
                      <div className="text-xs text-gray-500">{option.description}</div>
                    </div>
                  </div>
                )
              })}
            </div>
          </div>

          <Separator />

          {/* Severity Filter */}
          <div className="space-y-2">
            <label className="text-sm font-medium text-gray-700">Filter by Severity</label>
            <div className="flex flex-wrap gap-2">
              {severityOptions.map((option) => (
                <Badge
                  key={option.id}
                  onClick={() => toggleSeverity(option.id)}
                  className={cn(
                    'cursor-pointer transition-all',
                    selectedSeverities.includes(option.id)
                      ? `${option.color} text-white`
                      : 'bg-gray-100 text-gray-600 hover:bg-gray-200'
                  )}
                >
                  {option.label}
                </Badge>
              ))}
            </div>
            {selectedSeverities.length === 0 && (
              <p className="text-xs text-gray-500">All severities included</p>
            )}
          </div>

          {/* Status Filter */}
          <div className="space-y-2">
            <label className="text-sm font-medium text-gray-700">Filter by Status</label>
            <div className="flex flex-wrap gap-2">
              {statusOptions.map((option) => (
                <Badge
                  key={option.id}
                  variant="outline"
                  onClick={() => toggleStatus(option.id)}
                  className={cn(
                    'cursor-pointer transition-all',
                    selectedStatuses.includes(option.id)
                      ? 'bg-blue-500 text-white border-blue-500'
                      : 'hover:bg-gray-100'
                  )}
                >
                  {option.label}
                </Badge>
              ))}
            </div>
            {selectedStatuses.length === 0 && (
              <p className="text-xs text-gray-500">All statuses included</p>
            )}
          </div>

          <Separator />

          {/* Action Buttons */}
          <div className="space-y-2">
            <Button
              onClick={handleExport}
              disabled={isExporting || !title.trim()}
              className="w-full"
            >
              {isExporting ? (
                <>
                  <Loader2 className="w-4 h-4 mr-2 animate-spin" />
                  Exporting...
                </>
              ) : exportSuccess ? (
                <>
                  <CheckCircle className="w-4 h-4 mr-2" />
                  Exported!
                </>
              ) : (
                <>
                  <Download className="w-4 h-4 mr-2" />
                  Export Report
                </>
              )}
            </Button>

            <Button variant="outline" onClick={handlePreview} className="w-full">
              <Eye className="w-4 h-4 mr-2" />
              Preview Data
            </Button>
          </div>
        </div>
      </div>

      {/* Preview Panel */}
      <div className="flex-1 flex flex-col bg-gray-50">
        <div className="px-4 py-3 border-b border-gray-200 bg-white">
          <h3 className="text-sm font-medium text-gray-700">Preview</h3>
        </div>

        <div className="flex-1 overflow-auto p-4">
          {previewData ? (
            <Card>
              <CardContent className="p-4">
                <pre className="text-xs font-mono text-gray-700 whitespace-pre-wrap overflow-auto">
                  {previewData}
                </pre>
              </CardContent>
            </Card>
          ) : (
            <div className="flex flex-col items-center justify-center h-full text-gray-400">
              <File className="w-16 h-16 mb-4" />
              <p className="text-sm font-medium">No preview available</p>
              <p className="text-xs mt-1">Click "Preview Data" to see report content</p>
            </div>
          )}
        </div>
      </div>
    </div>
  )
}

export default ReportExport
