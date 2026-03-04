import { useState, useMemo } from 'react'
import { AlertTriangle, Bug, CheckCircle, Search, Trash2, X } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Badge } from '@/components/ui/badge'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { cn } from '@/lib/utils'
import { HttpEditor } from '@/components/common/HttpEditor'

// Severity type
type Severity = 'high' | 'medium' | 'low' | 'info'

// Vulnerability interface
interface Vulnerability {
  id: string
  ruleId: string
  name: string
  severity: Severity
  description: string
  remediation: string
  url: string
  method: string
  evidence: string
  location: string
  request: string
  response: string
  timestamp: string
  falsePositive: boolean
}

// Mock data for demonstration
const mockVulnerabilities: Vulnerability[] = [
  {
    id: '1',
    ruleId: 'SQL-INJECTION-001',
    name: 'Potential SQL Injection',
    severity: 'high',
    description: 'Request parameter may be vulnerable to SQL injection attacks.',
    remediation: 'Use parameterized queries or prepared statements.',
    url: 'https://example.com/api/users?id=1',
    method: 'GET',
    evidence: "id=1' OR '1'='1",
    location: 'query parameter',
    request: 'GET /api/users?id=1%27%20OR%20%271%27=%271 HTTP/1.1\nHost: example.com',
    response: 'HTTP/1.1 200 OK\n\n{"users": [...]}',
    timestamp: '2026-03-03T10:00:00Z',
    falsePositive: false,
  },
  {
    id: '2',
    ruleId: 'XSS-001',
    name: 'Cross-Site Scripting (XSS)',
    severity: 'high',
    description: 'Reflected XSS vulnerability detected in search parameter.',
    remediation: 'Implement proper output encoding and Content-Security-Policy.',
    url: 'https://example.com/search?q=test',
    method: 'GET',
    evidence: '<script>alert(1)</script>',
    location: 'query parameter',
    request: 'GET /search?q=%3Cscript%3Ealert(1)%3C/script%3E HTTP/1.1\nHost: example.com',
    response: 'HTTP/1.1 200 OK\n\n<div>Results for: <script>alert(1)</script></div>',
    timestamp: '2026-03-03T09:30:00Z',
    falsePositive: false,
  },
  {
    id: '3',
    ruleId: 'SENSITIVE-001',
    name: 'Credit Card Number Exposure',
    severity: 'high',
    description: 'Response contains what appears to be a credit card number.',
    remediation: 'Mask or tokenize sensitive payment data.',
    url: 'https://api.example.com/orders/123',
    method: 'GET',
    evidence: '4111-1111-1111-1111',
    location: 'response body',
    request: 'GET /orders/123 HTTP/1.1\nHost: api.example.com',
    response: 'HTTP/1.1 200 OK\n\n{"card": "4111-1111-1111-1111"}',
    timestamp: '2026-03-03T09:00:00Z',
    falsePositive: false,
  },
  {
    id: '4',
    ruleId: 'HEADERS-001',
    name: 'Missing X-Frame-Options Header',
    severity: 'medium',
    description: 'Response does not include X-Frame-Options security header.',
    remediation: 'Add X-Frame-Options: DENY or SAMEORIGIN header.',
    url: 'https://example.com/page',
    method: 'GET',
    evidence: 'Missing: X-Frame-Options',
    location: 'headers',
    request: 'GET /page HTTP/1.1\nHost: example.com',
    response: 'HTTP/1.1 200 OK\nContent-Type: text/html\n\n<html>...</html>',
    timestamp: '2026-03-03T08:30:00Z',
    falsePositive: false,
  },
  {
    id: '5',
    ruleId: 'INFO-001',
    name: 'Server Version Disclosure',
    severity: 'low',
    description: 'Server header reveals technology version information.',
    remediation: 'Remove or obfuscate server version headers.',
    url: 'https://example.com/',
    method: 'GET',
    evidence: 'Server: Apache/2.4.41 (Ubuntu)',
    location: 'headers',
    request: 'GET / HTTP/1.1\nHost: example.com',
    response: 'HTTP/1.1 200 OK\nServer: Apache/2.4.41 (Ubuntu)\n\n<html>...</html>',
    timestamp: '2026-03-03T08:00:00Z',
    falsePositive: false,
  },
]

// Severity config
const severityConfig: Record<Severity, { color: string; bgColor: string; borderColor: string }> = {
  high: { color: 'text-red-600', bgColor: 'bg-red-50', borderColor: 'border-red-200' },
  medium: { color: 'text-orange-600', bgColor: 'bg-orange-50', borderColor: 'border-orange-200' },
  low: { color: 'text-yellow-600', bgColor: 'bg-yellow-50', borderColor: 'border-yellow-200' },
  info: { color: 'text-blue-600', bgColor: 'bg-blue-50', borderColor: 'border-blue-200' },
}

export function VulnsTab() {
  const [vulnerabilities, setVulnerabilities] = useState<Vulnerability[]>(mockVulnerabilities)
  const [selectedVuln, setSelectedVuln] = useState<Vulnerability | null>(null)
  const [searchQuery, setSearchQuery] = useState('')
  const [severityFilter, setSeverityFilter] = useState<Severity | 'all'>('all')
  const [showFalsePositives] = useState(false)

  // Filtered vulnerabilities
  const filteredVulns = useMemo(() => {
    return vulnerabilities.filter(vuln => {
      // Filter out false positives unless showing them
      if (vuln.falsePositive && !showFalsePositives) return false

      // Severity filter
      if (severityFilter !== 'all' && vuln.severity !== severityFilter) return false

      // Search filter
      if (searchQuery) {
        const query = searchQuery.toLowerCase()
        return (
          vuln.name.toLowerCase().includes(query) ||
          vuln.url.toLowerCase().includes(query) ||
          vuln.description.toLowerCase().includes(query)
        )
      }

      return true
    })
  }, [vulnerabilities, searchQuery, severityFilter, showFalsePositives])

  // Count by severity
  const counts = useMemo(() => {
    const active = vulnerabilities.filter(v => !v.falsePositive)
    return {
      high: active.filter(v => v.severity === 'high').length,
      medium: active.filter(v => v.severity === 'medium').length,
      low: active.filter(v => v.severity === 'low').length,
      info: active.filter(v => v.severity === 'info').length,
      total: active.length,
    }
  }, [vulnerabilities])

  // Mark as false positive
  const markFalsePositive = (id: string) => {
    setVulnerabilities(prev =>
      prev.map(v => v.id === id ? { ...v, falsePositive: true } : v)
    )
    if (selectedVuln?.id === id) {
      setSelectedVuln(null)
    }
  }

  // Restore false positive
  const restoreVuln = (id: string) => {
    setVulnerabilities(prev =>
      prev.map(v => v.id === id ? { ...v, falsePositive: false } : v)
    )
  }

  // Delete vulnerability
  const deleteVuln = (id: string) => {
    setVulnerabilities(prev => prev.filter(v => v.id !== id))
    if (selectedVuln?.id === id) {
      setSelectedVuln(null)
    }
  }

  return (
    <div className="flex h-full">
      {/* Left panel - Vulnerability list */}
      <div className="w-1/2 flex flex-col border-r border-border">
        {/* Header with filters */}
        <div className="p-4 border-b border-border bg-gray-50 dark:bg-gray-800 space-y-3">
          {/* Search */}
          <div className="relative">
            <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 w-4 h-4 text-gray-400" />
            <Input
              placeholder="Search vulnerabilities..."
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              className="pl-10"
            />
          </div>

          {/* Severity filter buttons */}
          <div className="flex items-center gap-2 flex-wrap">
            <Button
              variant={severityFilter === 'all' ? 'default' : 'outline'}
              size="sm"
              onClick={() => setSeverityFilter('all')}
            >
              All ({counts.total})
            </Button>
            <Button
              variant={severityFilter === 'high' ? 'destructive' : 'outline'}
              size="sm"
              onClick={() => setSeverityFilter('high')}
            >
              High ({counts.high})
            </Button>
            <Button
              variant={severityFilter === 'medium' ? 'default' : 'outline'}
              size="sm"
              onClick={() => setSeverityFilter('medium')}
              className="border-orange-500 text-orange-600"
            >
              Medium ({counts.medium})
            </Button>
            <Button
              variant={severityFilter === 'low' ? 'default' : 'outline'}
              size="sm"
              onClick={() => setSeverityFilter('low')}
              className="border-yellow-500 text-yellow-600"
            >
              Low ({counts.low})
            </Button>
            <Button
              variant={severityFilter === 'info' ? 'default' : 'outline'}
              size="sm"
              onClick={() => setSeverityFilter('info')}
              className="border-blue-500 text-blue-600"
            >
              Info ({counts.info})
            </Button>
          </div>
        </div>

        {/* Vulnerability list */}
        <div className="flex-1 overflow-auto">
          {filteredVulns.length === 0 ? (
            <div className="flex flex-col items-center justify-center h-full text-gray-400">
              <Bug className="w-12 h-12 mb-3" />
              <p className="text-sm">No vulnerabilities found</p>
            </div>
          ) : (
            filteredVulns.map((vuln) => {
              const config = severityConfig[vuln.severity]
              return (
                <div
                  key={vuln.id}
                  onClick={() => setSelectedVuln(vuln)}
                  className={cn(
                    'p-3 border-b cursor-pointer transition-colors',
                    selectedVuln?.id === vuln.id
                      ? 'bg-blue-50 dark:bg-blue-950'
                      : 'hover:bg-gray-50 dark:hover:bg-gray-800',
                    vuln.falsePositive && 'opacity-50'
                  )}
                >
                  <div className="flex items-start gap-3">
                    <div className={cn('mt-1', config.color)}>
                      <AlertTriangle className="w-4 h-4" />
                    </div>
                    <div className="flex-1 min-w-0">
                      <div className="flex items-center gap-2">
                        <span className="font-medium text-sm truncate">{vuln.name}</span>
                        <Badge
                          variant="outline"
                          className={cn(
                            'text-xs',
                            config.color,
                            config.bgColor,
                            config.borderColor
                          )}
                        >
                          {vuln.severity.toUpperCase()}
                        </Badge>
                        {vuln.falsePositive && (
                          <Badge variant="secondary" className="text-xs">
                            False Positive
                          </Badge>
                        )}
                      </div>
                      <p className="text-xs text-gray-500 truncate mt-1">{vuln.url}</p>
                      <p className="text-xs text-gray-400 mt-1">
                        {new Date(vuln.timestamp).toLocaleString()}
                      </p>
                    </div>
                  </div>
                </div>
              )
            })
          )}
        </div>
      </div>

      {/* Right panel - Vulnerability details */}
      <div className="w-1/2 flex flex-col">
        {selectedVuln ? (
          <>
            {/* Detail header */}
            <div className="p-4 border-b border-border bg-gray-50 dark:bg-gray-800">
              <div className="flex items-start justify-between">
                <div>
                  <div className="flex items-center gap-2">
                    <h3 className="font-semibold">{selectedVuln.name}</h3>
                    <Badge
                      className={cn(
                        severityConfig[selectedVuln.severity].color,
                        severityConfig[selectedVuln.severity].bgColor
                      )}
                    >
                      {selectedVuln.severity.toUpperCase()}
                    </Badge>
                  </div>
                  <p className="text-sm text-gray-500 mt-1">{selectedVuln.url}</p>
                </div>
                <div className="flex gap-2">
                  {selectedVuln.falsePositive ? (
                    <Button
                      variant="outline"
                      size="sm"
                      onClick={() => restoreVuln(selectedVuln.id)}
                    >
                      <CheckCircle className="w-4 h-4 mr-1" />
                      Restore
                    </Button>
                  ) : (
                    <Button
                      variant="outline"
                      size="sm"
                      onClick={() => markFalsePositive(selectedVuln.id)}
                    >
                      <X className="w-4 h-4 mr-1" />
                      False Positive
                    </Button>
                  )}
                  <Button
                    variant="outline"
                    size="sm"
                    onClick={() => deleteVuln(selectedVuln.id)}
                  >
                    <Trash2 className="w-4 h-4" />
                  </Button>
                </div>
              </div>
            </div>

            {/* Detail content */}
            <div className="flex-1 overflow-auto p-4 space-y-4">
              {/* Description */}
              <Card>
                <CardHeader className="py-2">
                  <CardTitle className="text-sm">Description</CardTitle>
                </CardHeader>
                <CardContent className="text-sm">{selectedVuln.description}</CardContent>
              </Card>

              {/* Evidence */}
              <Card>
                <CardHeader className="py-2">
                  <CardTitle className="text-sm">Evidence</CardTitle>
                </CardHeader>
                <CardContent>
                  <code className="text-xs bg-gray-100 dark:bg-gray-800 p-2 rounded block overflow-x-auto">
                    {selectedVuln.evidence}
                  </code>
                </CardContent>
              </Card>

              {/* Remediation */}
              <Card>
                <CardHeader className="py-2">
                  <CardTitle className="text-sm">Remediation</CardTitle>
                </CardHeader>
                <CardContent className="text-sm">{selectedVuln.remediation}</CardContent>
              </Card>

              {/* Request/Response */}
              <Tabs defaultValue="request" className="w-full">
                <TabsList className="w-full">
                  <TabsTrigger value="request" className="flex-1">Request</TabsTrigger>
                  <TabsTrigger value="response" className="flex-1">Response</TabsTrigger>
                </TabsList>
                <TabsContent value="request" className="mt-2">
                  <div className="h-64 border rounded-lg overflow-hidden">
                    <HttpEditor
                      value={selectedVuln.request}
                      language="http-request"
                      readOnly
                      height="100%"
                    />
                  </div>
                </TabsContent>
                <TabsContent value="response" className="mt-2">
                  <div className="h-64 border rounded-lg overflow-hidden">
                    <HttpEditor
                      value={selectedVuln.response}
                      language="http-response"
                      readOnly
                      height="100%"
                    />
                  </div>
                </TabsContent>
              </Tabs>
            </div>
          </>
        ) : (
          <div className="flex flex-col items-center justify-center h-full text-gray-400">
            <Bug className="w-12 h-12 mb-3" />
            <p className="text-sm">Select a vulnerability to view details</p>
          </div>
        )}
      </div>
    </div>
  )
}

export default VulnsTab
