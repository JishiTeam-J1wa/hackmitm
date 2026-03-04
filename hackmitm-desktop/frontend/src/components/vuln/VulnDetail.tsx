import {
  X,
  Copy,
  ExternalLink,
  CheckCircle,
  XCircle,
  AlertTriangle,
  AlertCircle,
  Info,
  Globe,
  Clock,
  Shield,
  FileText,
  Wrench,
  Link as LinkIcon,
} from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { Textarea } from '@/components/ui/textarea'
import { cn } from '@/lib/utils'
import type { Vulnerability } from '@/types'

interface VulnDetailProps {
  vuln: Vulnerability
  onClose: () => void
  onStatusChange: (status: 'open' | 'fixed' | 'ignored') => void
  onDelete: () => void
}

const severityConfig = {
  critical: {
    icon: AlertTriangle,
    color: 'text-red-600',
    bgColor: 'bg-red-50',
    borderColor: 'border-red-200',
    badgeColor: 'bg-red-500 text-white',
    label: 'Critical',
  },
  high: {
    icon: AlertCircle,
    color: 'text-orange-600',
    bgColor: 'bg-orange-50',
    borderColor: 'border-orange-200',
    badgeColor: 'bg-orange-500 text-white',
    label: 'High',
  },
  medium: {
    icon: AlertCircle,
    color: 'text-yellow-600',
    bgColor: 'bg-yellow-50',
    borderColor: 'border-yellow-200',
    badgeColor: 'bg-yellow-500 text-white',
    label: 'Medium',
  },
  low: {
    icon: Info,
    color: 'text-blue-600',
    bgColor: 'bg-blue-50',
    borderColor: 'border-blue-200',
    badgeColor: 'bg-blue-500 text-white',
    label: 'Low',
  },
}

export function VulnDetail({ vuln, onClose, onStatusChange, onDelete }: VulnDetailProps) {
  const severity = severityConfig[vuln.severity]
  const SeverityIcon = severity.icon

  const handleCopy = (content: string) => {
    navigator.clipboard.writeText(content)
  }

  return (
    <div className="flex flex-col h-full bg-white">
      {/* Header */}
      <div className={cn('border-b p-4', severity.borderColor, severity.bgColor)}>
        <div className="flex items-start gap-3">
          <SeverityIcon className={cn('w-6 h-6 mt-0.5 flex-shrink-0', severity.color)} />
          <div className="flex-1 min-w-0">
            <div className="flex items-center gap-2 mb-1">
              <h3 className="text-sm font-semibold text-gray-900">{vuln.title}</h3>
              <Badge className={cn('text-[10px]', severity.badgeColor)}>{severity.label}</Badge>
              <Badge variant="outline" className="text-[10px]">{vuln.type}</Badge>
            </div>
            <div className="flex items-center gap-4 text-xs text-gray-600">
              <div className="flex items-center gap-1">
                <Globe className="w-3 h-3" />
                <span className="truncate max-w-xs">{vuln.url}</span>
              </div>
              <div className="flex items-center gap-1">
                <Clock className="w-3 h-3" />
                <span>{new Date(vuln.createdAt).toLocaleString('zh-CN')}</span>
              </div>
            </div>
          </div>
          <Button variant="ghost" size="icon" onClick={onClose} className="w-8 h-8">
            <X className="w-4 h-4" />
          </Button>
        </div>

        {/* Action buttons */}
        <div className="flex items-center gap-2 mt-3">
          <Button
            size="sm"
            variant="outline"
            onClick={() => onStatusChange('fixed')}
            className="h-7 text-xs"
            disabled={vuln.status === 'fixed'}
          >
            <CheckCircle className="w-3 h-3 mr-1 text-green-500" />
            Mark Fixed
          </Button>
          <Button
            size="sm"
            variant="outline"
            onClick={() => onStatusChange('ignored')}
            className="h-7 text-xs"
            disabled={vuln.status === 'ignored'}
          >
            <XCircle className="w-3 h-3 mr-1 text-gray-400" />
            Ignore
          </Button>
          <Button
            size="sm"
            variant="outline"
            onClick={() => onStatusChange('open')}
            className="h-7 text-xs"
            disabled={vuln.status === 'open'}
          >
            <AlertCircle className="w-3 h-3 mr-1 text-orange-500" />
            Reopen
          </Button>
          <div className="flex-1" />
          <Button
            size="sm"
            variant="destructive"
            onClick={onDelete}
            className="h-7 text-xs"
          >
            Delete
          </Button>
        </div>
      </div>

      {/* Content */}
      <div className="flex-1 overflow-auto">
        <Tabs defaultValue="description" className="h-full flex flex-col">
          <TabsList className="px-4 pt-3 flex-shrink-0">
            <TabsTrigger value="description" className="text-xs">
              <FileText className="w-3 h-3 mr-1" />
              Description
            </TabsTrigger>
            <TabsTrigger value="request" className="text-xs">
              Request
            </TabsTrigger>
            <TabsTrigger value="response" className="text-xs">
              Response
            </TabsTrigger>
            <TabsTrigger value="remediation" className="text-xs">
              <Wrench className="w-3 h-3 mr-1" />
              Remediation
            </TabsTrigger>
          </TabsList>

          <TabsContent value="description" className="flex-1 p-4 overflow-auto">
            <div className="space-y-4">
              {/* Details */}
              <div>
                <h4 className="text-xs font-medium text-gray-700 mb-2">Description</h4>
                <p className="text-sm text-gray-600">{vuln.description}</p>
              </div>

              {/* Meta info */}
              <div className="grid grid-cols-2 gap-4">
                <div className="bg-gray-50 rounded p-3">
                  <span className="text-[10px] text-gray-500 block mb-1">Method</span>
                  <Badge variant="outline" className="text-xs">{vuln.method}</Badge>
                </div>
                <div className="bg-gray-50 rounded p-3">
                  <span className="text-[10px] text-gray-500 block mb-1">Source</span>
                  <Badge variant="outline" className="text-xs">
                    <Shield className="w-3 h-3 mr-1" />
                    {vuln.source === 'passive' ? 'Passive Scan' : 'Active Scan'}
                  </Badge>
                </div>
                {vuln.cwe && (
                  <div className="bg-gray-50 rounded p-3">
                    <span className="text-[10px] text-gray-500 block mb-1">CWE</span>
                    <a
                      href={`https://cwe.mitre.org/data/definitions/${vuln.cwe.replace('CWE-', '')}.html`}
                      target="_blank"
                      rel="noopener noreferrer"
                      className="text-xs text-blue-600 hover:underline flex items-center gap-1"
                    >
                      {vuln.cwe}
                      <ExternalLink className="w-3 h-3" />
                    </a>
                  </div>
                )}
                {vuln.cvss && (
                  <div className="bg-gray-50 rounded p-3">
                    <span className="text-[10px] text-gray-500 block mb-1">CVSS Score</span>
                    <span className={cn('text-sm font-medium', severity.color)}>
                      {vuln.cvss.toFixed(1)}
                    </span>
                  </div>
                )}
              </div>

              {/* References */}
              {vuln.references.length > 0 && (
                <div>
                  <h4 className="text-xs font-medium text-gray-700 mb-2">References</h4>
                  <div className="space-y-1">
                    {vuln.references.map((ref, index) => (
                      <a
                        key={index}
                        href={ref}
                        target="_blank"
                        rel="noopener noreferrer"
                        className="flex items-center gap-1 text-xs text-blue-600 hover:underline"
                      >
                        <LinkIcon className="w-3 h-3" />
                        {ref}
                      </a>
                    ))}
                  </div>
                </div>
              )}
            </div>
          </TabsContent>

          <TabsContent value="request" className="flex-1 overflow-hidden flex flex-col">
            <div className="flex items-center justify-between px-4 py-2 border-b bg-gray-50">
              <span className="text-xs text-gray-500">Request</span>
              <Button
                variant="ghost"
                size="sm"
                onClick={() => handleCopy(vuln.request)}
                className="h-6 text-xs"
              >
                <Copy className="w-3 h-3 mr-1" />
                Copy
              </Button>
            </div>
            <Textarea
              readOnly
              value={vuln.request}
              className="flex-1 font-mono text-xs border-0 rounded-none resize-none focus-visible:ring-0"
            />
          </TabsContent>

          <TabsContent value="response" className="flex-1 overflow-hidden flex flex-col">
            <div className="flex items-center justify-between px-4 py-2 border-b bg-gray-50">
              <span className="text-xs text-gray-500">Response</span>
              <Button
                variant="ghost"
                size="sm"
                onClick={() => handleCopy(vuln.response)}
                className="h-6 text-xs"
              >
                <Copy className="w-3 h-3 mr-1" />
                Copy
              </Button>
            </div>
            <Textarea
              readOnly
              value={vuln.response}
              className="flex-1 font-mono text-xs border-0 rounded-none resize-none focus-visible:ring-0"
            />
          </TabsContent>

          <TabsContent value="remediation" className="flex-1 p-4 overflow-auto">
            <div className="space-y-4">
              <div>
                <h4 className="text-xs font-medium text-gray-700 mb-2">How to Fix</h4>
                <p className="text-sm text-gray-600">{vuln.remediation}</p>
              </div>

              {/* CWE-based remediation tips */}
              {vuln.cwe && (
                <div className="bg-blue-50 border border-blue-200 rounded p-3">
                  <h4 className="text-xs font-medium text-blue-800 mb-1">
                    Additional Resources for {vuln.cwe}
                  </h4>
                  <p className="text-xs text-blue-600">
                    Consult the CWE database and OWASP guidelines for detailed remediation steps.
                  </p>
                </div>
              )}
            </div>
          </TabsContent>
        </Tabs>
      </div>
    </div>
  )
}
