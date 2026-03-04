import { useState, useEffect, useCallback } from 'react'
import {
  Search,
  Shield,
  RefreshCw,
  ToggleLeft,
  ToggleRight,
  Plus,
  Edit,
  Filter,
} from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Badge } from '@/components/ui/badge'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { Card, CardContent } from '@/components/ui/card'
import { cn } from '@/lib/utils'

interface ScanRule {
  id: string
  name: string
  description: string
  severity: 'critical' | 'high' | 'medium' | 'low' | 'info'
  enabled: boolean
  priority: number
  tags: string[]
}

const severityConfig = {
  critical: { color: 'bg-red-500 text-white', label: 'Critical' },
  high: { color: 'bg-orange-500 text-white', label: 'High' },
  medium: { color: 'bg-yellow-500 text-white', label: 'Medium' },
  low: { color: 'bg-blue-500 text-white', label: 'Low' },
  info: { color: 'bg-gray-500 text-white', label: 'Info' },
}

export function RuleConfig() {
  const [rules, setRules] = useState<ScanRule[]>([])
  const [loading, setLoading] = useState(true)
  const [search, setSearch] = useState('')
  const [severityFilter, setSeverityFilter] = useState('all')
  const [statusFilter, setStatusFilter] = useState('all')

  // Fetch rules from API
  const fetchRules = useCallback(async () => {
    setLoading(true)
    try {
      const response = await fetch('/api/scanner/rules')
      if (response.ok) {
        const data = await response.json()
        setRules(data.data || [])
      }
    } catch (error) {
      console.error('Failed to fetch rules:', error)
      // Use mock data for demo
      setRules([
        {
          id: 'sql-injection-basic',
          name: 'SQL Injection Detection',
          description: 'Detects basic SQL injection patterns in request parameters',
          severity: 'critical',
          enabled: true,
          priority: 100,
          tags: ['injection', 'sql', 'owasp'],
        },
        {
          id: 'xss-reflected',
          name: 'Reflected XSS Detection',
          description: 'Detects reflected cross-site scripting patterns',
          severity: 'high',
          enabled: true,
          priority: 90,
          tags: ['xss', 'injection', 'owasp'],
        },
        {
          id: 'sensitive-info-exposure',
          name: 'Sensitive Information Exposure',
          description: 'Detects sensitive data in responses like API keys, passwords',
          severity: 'high',
          enabled: true,
          priority: 85,
          tags: ['info-exposure', 'credentials'],
        },
        {
          id: 'path-traversal',
          name: 'Path Traversal Detection',
          description: 'Detects directory traversal attack patterns',
          severity: 'high',
          enabled: true,
          priority: 88,
          tags: ['path-traversal', 'lfi', 'owasp'],
        },
        {
          id: 'ssrf-detection',
          name: 'SSRF Detection',
          description: 'Detects server-side request forgery patterns',
          severity: 'high',
          enabled: true,
          priority: 85,
          tags: ['ssrf', 'owasp'],
        },
        {
          id: 'auth-bypass',
          name: 'Authentication Bypass',
          description: 'Detects potential authentication bypass patterns',
          severity: 'critical',
          enabled: true,
          priority: 95,
          tags: ['auth', 'bypass', 'owasp'],
        },
      ])
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => {
    fetchRules()
  }, [fetchRules])

  // Toggle rule enabled state
  const toggleRule = async (ruleId: string, enabled: boolean) => {
    try {
      const response = await fetch(`/api/scanner/rules/${ruleId}`, {
        method: 'PATCH',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ enabled }),
      })

      if (response.ok) {
        setRules((prev) =>
          prev.map((rule) => (rule.id === ruleId ? { ...rule, enabled } : rule))
        )
      }
    } catch (error) {
      console.error('Failed to toggle rule:', error)
      // Update locally for demo
      setRules((prev) =>
        prev.map((rule) => (rule.id === ruleId ? { ...rule, enabled } : rule))
      )
    }
  }

  // Reload rules
  const reloadRules = async () => {
    try {
      await fetch('/api/scanner/reload', { method: 'POST' })
      fetchRules()
    } catch (error) {
      console.error('Failed to reload rules:', error)
    }
  }

  // Filter rules
  const filteredRules = rules.filter((rule) => {
    if (search) {
      const searchLower = search.toLowerCase()
      if (
        !rule.name.toLowerCase().includes(searchLower) &&
        !rule.description.toLowerCase().includes(searchLower) &&
        !rule.tags.some((tag) => tag.toLowerCase().includes(searchLower))
      ) {
        return false
      }
    }
    if (severityFilter !== 'all' && rule.severity !== severityFilter) return false
    if (statusFilter === 'enabled' && !rule.enabled) return false
    if (statusFilter === 'disabled' && rule.enabled) return false
    return true
  })

  // Stats
  const stats = {
    total: rules.length,
    enabled: rules.filter((r) => r.enabled).length,
    disabled: rules.filter((r) => !r.enabled).length,
    critical: rules.filter((r) => r.severity === 'critical').length,
    high: rules.filter((r) => r.severity === 'high').length,
    medium: rules.filter((r) => r.severity === 'medium').length,
    low: rules.filter((r) => r.severity === 'low').length,
  }

  return (
    <div className="flex flex-col h-full bg-gray-50">
      {/* Header */}
      <div className="px-4 py-3 bg-white border-b border-gray-200 flex items-center justify-between">
        <div className="flex items-center gap-3">
          <Shield className="w-5 h-5 text-blue-500" />
          <h2 className="text-lg font-semibold text-gray-800">Scan Rules</h2>
          <Badge variant="secondary" className="text-xs">
            {stats.enabled}/{stats.total} enabled
          </Badge>
        </div>
        <div className="flex items-center gap-2">
          <Button variant="outline" size="sm" onClick={reloadRules}>
            <RefreshCw className="w-4 h-4 mr-1" />
            Reload
          </Button>
          <Button size="sm">
            <Plus className="w-4 h-4 mr-1" />
            Custom Rule
          </Button>
        </div>
      </div>

      {/* Stats Bar */}
      <div className="px-4 py-2 bg-white border-b border-gray-200 flex items-center gap-4 text-xs">
        <div className="flex items-center gap-1.5">
          <div className="w-2 h-2 rounded-full bg-red-500" />
          <span className="text-gray-600">Critical: {stats.critical}</span>
        </div>
        <div className="flex items-center gap-1.5">
          <div className="w-2 h-2 rounded-full bg-orange-500" />
          <span className="text-gray-600">High: {stats.high}</span>
        </div>
        <div className="flex items-center gap-1.5">
          <div className="w-2 h-2 rounded-full bg-yellow-500" />
          <span className="text-gray-600">Medium: {stats.medium}</span>
        </div>
        <div className="flex items-center gap-1.5">
          <div className="w-2 h-2 rounded-full bg-blue-500" />
          <span className="text-gray-600">Low: {stats.low}</span>
        </div>
      </div>

      {/* Filters */}
      <div className="px-4 py-2 bg-white border-b border-gray-200 flex items-center gap-2">
        <div className="relative flex-1 max-w-xs">
          <Search className="absolute left-2 top-1/2 -translate-y-1/2 w-3.5 h-3.5 text-gray-400" />
          <Input
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            placeholder="Search rules..."
            className="h-8 pl-7 text-xs"
          />
        </div>

        <Select value={severityFilter} onValueChange={setSeverityFilter}>
          <SelectTrigger className="w-28 h-8 text-xs">
            <SelectValue placeholder="Severity" />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="all">All Severity</SelectItem>
            <SelectItem value="critical">Critical</SelectItem>
            <SelectItem value="high">High</SelectItem>
            <SelectItem value="medium">Medium</SelectItem>
            <SelectItem value="low">Low</SelectItem>
          </SelectContent>
        </Select>

        <Select value={statusFilter} onValueChange={setStatusFilter}>
          <SelectTrigger className="w-28 h-8 text-xs">
            <SelectValue placeholder="Status" />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="all">All Status</SelectItem>
            <SelectItem value="enabled">Enabled</SelectItem>
            <SelectItem value="disabled">Disabled</SelectItem>
          </SelectContent>
        </Select>
      </div>

      {/* Rules List */}
      <div className="flex-1 overflow-auto p-4">
        {loading ? (
          <div className="flex items-center justify-center h-64">
            <RefreshCw className="w-6 h-6 animate-spin text-gray-400" />
          </div>
        ) : filteredRules.length === 0 ? (
          <div className="flex flex-col items-center justify-center h-64 text-gray-400">
            <Filter className="w-12 h-12 mb-3" />
            <p className="text-sm font-medium">No rules found</p>
            <p className="text-xs mt-1">Try adjusting your filters</p>
          </div>
        ) : (
          <div className="space-y-2">
            {filteredRules.map((rule) => {
              const severity = severityConfig[rule.severity]
              return (
                <Card key={rule.id} className="bg-white shadow-sm">
                  <CardContent className="p-4">
                    <div className="flex items-start justify-between">
                      <div className="flex-1">
                        <div className="flex items-center gap-2 mb-1">
                          <h3 className="text-sm font-medium text-gray-800">{rule.name}</h3>
                          <Badge className={cn('text-[10px]', severity.color)}>
                            {severity.label}
                          </Badge>
                          {!rule.enabled && (
                            <Badge variant="outline" className="text-[10px] text-gray-500">
                              Disabled
                            </Badge>
                          )}
                        </div>
                        <p className="text-xs text-gray-500 mb-2">{rule.description}</p>
                        <div className="flex items-center gap-2">
                          {rule.tags.map((tag) => (
                            <Badge key={tag} variant="secondary" className="text-[10px]">
                              {tag}
                            </Badge>
                          ))}
                        </div>
                      </div>

                      <div className="flex items-center gap-2">
                        <Button
                          variant="ghost"
                          size="icon"
                          className="w-8 h-8"
                          title="Edit rule"
                        >
                          <Edit className="w-4 h-4 text-gray-500" />
                        </Button>
                        <Button
                          variant="ghost"
                          size="icon"
                          className="w-8 h-8"
                          onClick={() => toggleRule(rule.id, !rule.enabled)}
                          title={rule.enabled ? 'Disable rule' : 'Enable rule'}
                        >
                          {rule.enabled ? (
                            <ToggleRight className="w-5 h-5 text-green-500" />
                          ) : (
                            <ToggleLeft className="w-5 h-5 text-gray-400" />
                          )}
                        </Button>
                      </div>
                    </div>
                  </CardContent>
                </Card>
              )
            })}
          </div>
        )}
      </div>
    </div>
  )
}

export default RuleConfig
