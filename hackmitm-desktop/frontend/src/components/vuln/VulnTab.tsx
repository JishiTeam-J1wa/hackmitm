import { useState, useMemo, useCallback } from 'react'
import {
  Search,
  LayoutGrid,
  List,
  AlertTriangle,
  AlertCircle,
  Info,
  CheckCircle,
  XCircle,
  Download,
  Shield,
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
import { useContextMenu, createVulnMenuItems } from '@/components/ui/ContextMenu'
import { cn } from '@/lib/utils'
import { useVulnStore } from '@/store'
import { VulnCard } from './VulnCard'
import { VulnDetail } from './VulnDetail'
import { ResizablePanelGroup } from '@/components/ui/resizable'
import type { Vulnerability } from '@/types'

export function VulnTab() {
  const {
    vulnerabilities,
    selectedVuln,
    filters,
    viewMode,
    selectVuln,
    setFilters,
    setViewMode,
    updateVulnerability,
    deleteVulnerability,
    getStats,
  } = useVulnStore()

  const { show } = useContextMenu()
  const [localSearch, setLocalSearch] = useState(filters.search)

  const stats = getStats()

  // Filter vulnerabilities
  const filteredVulns = useMemo(() => {
    let result = [...vulnerabilities]

    if (filters.severity !== 'all') {
      result = result.filter((v) => v.severity === filters.severity)
    }

    if (filters.type !== 'all') {
      result = result.filter((v) => v.type === filters.type)
    }

    if (filters.status !== 'all') {
      result = result.filter((v) => v.status === filters.status)
    }

    if (filters.search) {
      const search = filters.search.toLowerCase()
      result = result.filter(
        (v) =>
          v.title.toLowerCase().includes(search) ||
          v.url.toLowerCase().includes(search) ||
          v.description.toLowerCase().includes(search)
      )
    }

    return result
  }, [vulnerabilities, filters])

  // Get unique types for filter
  const uniqueTypes = useMemo(() => {
    const types = new Set(vulnerabilities.map((v) => v.type))
    return Array.from(types).sort()
  }, [vulnerabilities])

  const handleSearchChange = (value: string) => {
    setLocalSearch(value)
    setFilters({ search: value })
  }

  const handleStatusChange = (id: string, status: 'open' | 'fixed' | 'ignored') => {
    updateVulnerability(id, { status })
  }

  const handleDelete = (id: string) => {
    if (confirm('Are you sure you want to delete this vulnerability?')) {
      deleteVulnerability(id)
    }
  }

  const handleExport = () => {
    const data = JSON.stringify(filteredVulns, null, 2)
    const blob = new Blob([data], { type: 'application/json' })
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = `vulnerabilities-${new Date().toISOString().split('T')[0]}.json`
    a.click()
    URL.revokeObjectURL(url)
  }

  // Context menu handler for vulnerability items
  const handleVulnContextMenu = useCallback((e: React.MouseEvent, vuln: Vulnerability) => {
    e.preventDefault()
    e.stopPropagation()

    const menuItems = createVulnMenuItems(vuln, {
      markConfirmed: () => handleStatusChange(vuln.id, 'open'),
      markFalsePositive: () => handleStatusChange(vuln.id, 'ignored'),
      markFixed: () => handleStatusChange(vuln.id, 'fixed'),
      copyDetails: () => {
        const details = `${vuln.title}\n\nSeverity: ${vuln.severity}\nType: ${vuln.type}\nURL: ${vuln.url}\n\nDescription:\n${vuln.description}\n\nRemediation:\n${vuln.remediation}`
        navigator.clipboard.writeText(details)
      },
      viewRequest: () => {
        selectVuln(vuln)
      },
      generateReport: () => {
        // TODO: Generate report for this specific vulnerability
        const report = `# ${vuln.title}\n\n**Severity:** ${vuln.severity}\n**Type:** ${vuln.type}\n**URL:** ${vuln.url}\n\n## Description\n${vuln.description}\n\n## Remediation\n${vuln.remediation}`
        const blob = new Blob([report], { type: 'text/markdown' })
        const url = URL.createObjectURL(blob)
        const a = document.createElement('a')
        a.href = url
        a.download = `vuln-${vuln.id}.md`
        a.click()
        URL.revokeObjectURL(url)
      },
      delete: () => handleDelete(vuln.id),
    })

    show(e.clientX, e.clientY, menuItems)
  }, [show, handleStatusChange, handleDelete, selectVuln])

  const StatCard = ({
    icon: Icon,
    label,
    count,
    color,
    bgColor,
  }: {
    icon: typeof AlertTriangle
    label: string
    count: number
    color: string
    bgColor: string
  }) => (
    <div className={cn('flex items-center gap-2 px-3 py-2 rounded-lg', bgColor)}>
      <Icon className={cn('w-4 h-4', color)} />
      <div>
        <div className={cn('text-lg font-semibold', color)}>{count}</div>
        <div className="text-[10px] text-gray-500">{label}</div>
      </div>
    </div>
  )

  return (
    <div className="flex h-full flex-col">
      {/* Stats bar */}
      <div className="h-14 border-b border-gray-200 bg-white px-4 flex items-center gap-3 flex-shrink-0">
        <Shield className="w-5 h-5 text-blue-500" />
        <span className="text-sm font-medium text-gray-700">Vulnerabilities</span>

        <div className="w-px h-6 bg-gray-200 mx-2" />

        <StatCard
          icon={AlertTriangle}
          label="Critical"
          count={stats.critical}
          color="text-red-600"
          bgColor="bg-red-50"
        />
        <StatCard
          icon={AlertCircle}
          label="High"
          count={stats.high}
          color="text-orange-600"
          bgColor="bg-orange-50"
        />
        <StatCard
          icon={Info}
          label="Medium"
          count={stats.medium}
          color="text-yellow-600"
          bgColor="bg-yellow-50"
        />
        <StatCard
          icon={Info}
          label="Low"
          count={stats.low}
          color="text-blue-600"
          bgColor="bg-blue-50"
        />

        <div className="flex-1" />

        <div className="flex items-center gap-2 text-xs text-gray-500">
          <CheckCircle className="w-4 h-4 text-green-500" />
          <span>{stats.fixed} Fixed</span>
          <XCircle className="w-4 h-4 text-gray-400 ml-2" />
          <span>{stats.ignored} Ignored</span>
        </div>
      </div>

      {/* Toolbar */}
      <div className="h-10 border-b border-gray-200 bg-white px-4 flex items-center gap-2 flex-shrink-0">
        {/* Search */}
        <div className="relative flex-1 max-w-xs">
          <Search className="absolute left-2 top-1/2 -translate-y-1/2 w-3.5 h-3.5 text-gray-400" />
          <Input
            value={localSearch}
            onChange={(e) => handleSearchChange(e.target.value)}
            placeholder="Search vulnerabilities..."
            className="h-7 pl-7 text-xs"
          />
        </div>

        {/* Severity filter */}
        <Select
          value={filters.severity}
          onValueChange={(v) => setFilters({ severity: v })}
        >
          <SelectTrigger className="w-28 h-7 text-xs">
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

        {/* Type filter */}
        <Select value={filters.type} onValueChange={(v) => setFilters({ type: v })}>
          <SelectTrigger className="w-28 h-7 text-xs">
            <SelectValue placeholder="Type" />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="all">All Types</SelectItem>
            {uniqueTypes.map((type) => (
              <SelectItem key={type} value={type}>
                {type}
              </SelectItem>
            ))}
          </SelectContent>
        </Select>

        {/* Status filter */}
        <Select
          value={filters.status}
          onValueChange={(v) => setFilters({ status: v })}
        >
          <SelectTrigger className="w-24 h-7 text-xs">
            <SelectValue placeholder="Status" />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="all">All Status</SelectItem>
            <SelectItem value="open">Open</SelectItem>
            <SelectItem value="fixed">Fixed</SelectItem>
            <SelectItem value="ignored">Ignored</SelectItem>
          </SelectContent>
        </Select>

        <div className="flex-1" />

        <Badge variant="secondary" className="text-xs">
          {filteredVulns.length} / {vulnerabilities.length}
        </Badge>

        {/* View mode toggle */}
        <div className="flex items-center border rounded-md">
          <Button
            variant={viewMode === 'cards' ? 'secondary' : 'ghost'}
            size="icon"
            onClick={() => setViewMode('cards')}
            className="w-7 h-7"
          >
            <LayoutGrid className="w-3.5 h-3.5" />
          </Button>
          <Button
            variant={viewMode === 'list' ? 'secondary' : 'ghost'}
            size="icon"
            onClick={() => setViewMode('list')}
            className="w-7 h-7"
          >
            <List className="w-3.5 h-3.5" />
          </Button>
        </div>

        <Button variant="outline" size="sm" onClick={handleExport} className="h-7 text-xs">
          <Download className="w-3.5 h-3.5 mr-1" />
          Export
        </Button>
      </div>

      {/* Main content */}
      <div className="flex-1 overflow-hidden">
        {selectedVuln ? (
          <ResizablePanelGroup direction="horizontal" defaultSizes={[50, 50]} minSizes={[200, 300]}>
            {/* List panel */}
            <div className="h-full overflow-auto p-3 bg-gray-50">
              {viewMode === 'cards' ? (
                <div className="grid grid-cols-1 gap-2">
                  {filteredVulns.map((vuln) => (
                    <VulnCard
                      key={vuln.id}
                      vuln={vuln}
                      isSelected={selectedVuln.id === vuln.id}
                      onClick={() => selectVuln(vuln)}
                      onContextMenu={(e) => handleVulnContextMenu(e, vuln)}
                    />
                  ))}
                </div>
              ) : (
                <div className="space-y-1">
                  {filteredVulns.map((vuln) => (
                    <div
                      key={vuln.id}
                      onClick={() => selectVuln(vuln)}
                      onContextMenu={(e) => handleVulnContextMenu(e, vuln)}
                      className={cn(
                        'px-3 py-2 rounded cursor-pointer text-xs',
                        selectedVuln.id === vuln.id
                          ? 'bg-blue-100 text-blue-800'
                          : 'hover:bg-gray-100'
                      )}
                    >
                      <div className="font-medium truncate">{vuln.title}</div>
                      <div className="text-gray-500 truncate">{vuln.url}</div>
                    </div>
                  ))}
                </div>
              )}
            </div>

            {/* Detail panel */}
            <VulnDetail
              vuln={selectedVuln}
              onClose={() => selectVuln(null)}
              onStatusChange={(status) => handleStatusChange(selectedVuln.id, status)}
              onDelete={() => handleDelete(selectedVuln.id)}
            />
          </ResizablePanelGroup>
        ) : (
          // No selection - full list view
          <div className="h-full overflow-auto p-3 bg-gray-50">
            {filteredVulns.length === 0 ? (
              <div className="flex flex-col items-center justify-center h-full text-gray-400 py-20">
                <Shield className="w-12 h-12 mb-3" />
                <p className="text-sm font-medium">No vulnerabilities found</p>
                <p className="text-xs mt-1">
                  {filters.search || filters.severity !== 'all' || filters.status !== 'all'
                    ? 'Try adjusting your filters'
                    : 'Run passive or active scans to find vulnerabilities'}
                </p>
              </div>
            ) : viewMode === 'cards' ? (
              <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-3">
                {filteredVulns.map((vuln) => (
                  <VulnCard
                    key={vuln.id}
                    vuln={vuln}
                    isSelected={false}
                    onClick={() => selectVuln(vuln)}
                    onContextMenu={(e) => handleVulnContextMenu(e, vuln)}
                  />
                ))}
              </div>
            ) : (
              <div className="bg-white rounded-lg border">
                <table className="w-full text-xs">
                  <thead className="bg-gray-50 text-gray-500">
                    <tr>
                      <th className="px-3 py-2 text-left font-medium">Severity</th>
                      <th className="px-3 py-2 text-left font-medium">Title</th>
                      <th className="px-3 py-2 text-left font-medium">Type</th>
                      <th className="px-3 py-2 text-left font-medium">URL</th>
                      <th className="px-3 py-2 text-left font-medium">Status</th>
                      <th className="px-3 py-2 text-left font-medium">Date</th>
                    </tr>
                  </thead>
                  <tbody className="divide-y">
                    {filteredVulns.map((vuln) => (
                      <tr
                        key={vuln.id}
                        onClick={() => selectVuln(vuln)}
                        onContextMenu={(e) => handleVulnContextMenu(e, vuln)}
                        className="cursor-pointer hover:bg-gray-50"
                      >
                        <td className="px-3 py-2">
                          <Badge
                            className={cn(
                              'text-[10px]',
                              vuln.severity === 'critical' && 'bg-red-500 text-white',
                              vuln.severity === 'high' && 'bg-orange-500 text-white',
                              vuln.severity === 'medium' && 'bg-yellow-500 text-white',
                              vuln.severity === 'low' && 'bg-blue-500 text-white'
                            )}
                          >
                            {vuln.severity}
                          </Badge>
                        </td>
                        <td className="px-3 py-2 font-medium truncate max-w-xs">
                          {vuln.title}
                        </td>
                        <td className="px-3 py-2 text-gray-500">{vuln.type}</td>
                        <td className="px-3 py-2 text-gray-500 truncate max-w-xs">{vuln.url}</td>
                        <td className="px-3 py-2">
                          <Badge variant="outline" className="text-[10px]">
                            {vuln.status}
                          </Badge>
                        </td>
                        <td className="px-3 py-2 text-gray-500">
                          {new Date(vuln.createdAt).toLocaleDateString('zh-CN')}
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            )}
          </div>
        )}
      </div>
    </div>
  )
}
