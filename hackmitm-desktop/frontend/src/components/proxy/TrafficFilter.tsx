import { useState } from 'react'
import {
  Search,
  Filter,
  X,
  ChevronDown,
  ChevronUp,
  Globe,
  FileCode,
  Clock,
  HardDrive,
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
import { cn } from '@/lib/utils'

export interface TrafficFilterState {
  searchQuery: string
  methodFilter: string
  statusFilter: string
  hostFilter: string
  contentTypeFilter: string
  minSize: string
  maxSize: string
  minTime: string
  maxTime: string
}

interface TrafficFilterProps {
  filters: TrafficFilterState
  onFiltersChange: (filters: TrafficFilterState) => void
  hostList?: string[]
  resultCount?: number
  totalCount?: number
}

export function TrafficFilter({
  filters,
  onFiltersChange,
  hostList = [],
  resultCount = 0,
  totalCount = 0,
}: TrafficFilterProps) {
  const [expanded, setExpanded] = useState(false)

  const updateFilter = (key: keyof TrafficFilterState, value: string) => {
    onFiltersChange({ ...filters, [key]: value })
  }

  const clearFilters = () => {
    onFiltersChange({
      searchQuery: '',
      methodFilter: 'all',
      statusFilter: 'all',
      hostFilter: 'all',
      contentTypeFilter: 'all',
      minSize: '',
      maxSize: '',
      minTime: '',
      maxTime: '',
    })
  }

  const hasActiveFilters =
    filters.searchQuery ||
    filters.methodFilter !== 'all' ||
    filters.statusFilter !== 'all' ||
    filters.hostFilter !== 'all' ||
    filters.contentTypeFilter !== 'all' ||
    filters.minSize ||
    filters.maxSize ||
    filters.minTime ||
    filters.maxTime

  const activeFilterCount = [
    filters.searchQuery,
    filters.methodFilter !== 'all' && filters.methodFilter,
    filters.statusFilter !== 'all' && filters.statusFilter,
    filters.hostFilter !== 'all' && filters.hostFilter,
    filters.contentTypeFilter !== 'all' && filters.contentTypeFilter,
    filters.minSize,
    filters.maxSize,
    filters.minTime,
    filters.maxTime,
  ].filter(Boolean).length

  return (
    <div className="border-b border-gray-200 bg-white">
      {/* 主筛选栏 */}
      <div className="flex items-center px-3 py-2 gap-2">
        {/* 搜索框 */}
        <div className="relative flex-1 max-w-sm">
          <Search className="absolute left-2 top-1/2 -translate-y-1/2 w-3.5 h-3.5 text-gray-400" />
          <Input
            value={filters.searchQuery}
            onChange={(e) => updateFilter('searchQuery', e.target.value)}
            placeholder="搜索 URL、Host、Path..."
            className="h-7 pl-7 text-xs pr-7"
          />
          {filters.searchQuery && (
            <button
              onClick={() => updateFilter('searchQuery', '')}
              className="absolute right-2 top-1/2 -translate-y-1/2 text-gray-400 hover:text-gray-600"
            >
              <X className="w-3.5 h-3.5" />
            </button>
          )}
        </div>

        {/* 方法筛选 */}
        <Select value={filters.methodFilter} onValueChange={(v) => updateFilter('methodFilter', v)}>
          <SelectTrigger className="w-20 h-7 text-xs">
            <SelectValue placeholder="方法" />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="all">全部方法</SelectItem>
            <SelectItem value="GET">GET</SelectItem>
            <SelectItem value="POST">POST</SelectItem>
            <SelectItem value="PUT">PUT</SelectItem>
            <SelectItem value="DELETE">DELETE</SelectItem>
            <SelectItem value="PATCH">PATCH</SelectItem>
            <SelectItem value="OPTIONS">OPTIONS</SelectItem>
            <SelectItem value="HEAD">HEAD</SelectItem>
          </SelectContent>
        </Select>

        {/* 状态码筛选 */}
        <Select value={filters.statusFilter} onValueChange={(v) => updateFilter('statusFilter', v)}>
          <SelectTrigger className="w-20 h-7 text-xs">
            <SelectValue placeholder="状态" />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="all">全部状态</SelectItem>
            <SelectItem value="2xx">2xx 成功</SelectItem>
            <SelectItem value="3xx">3xx 重定向</SelectItem>
            <SelectItem value="4xx">4xx 客户端错误</SelectItem>
            <SelectItem value="5xx">5xx 服务端错误</SelectItem>
            <SelectItem value="error">错误 (≥400)</SelectItem>
          </SelectContent>
        </Select>

        {/* 展开/收起 */}
        <Button
          size="sm"
          variant="ghost"
          onClick={() => setExpanded(!expanded)}
          className={cn('h-7 px-2 text-xs', expanded && 'bg-gray-100')}
        >
          <Filter className="w-3.5 h-3.5 mr-1" />
          高级
          {activeFilterCount > 0 && (
            <Badge className="ml-1 h-4 px-1 text-[10px] bg-blue-500">{activeFilterCount}</Badge>
          )}
          {expanded ? <ChevronUp className="w-3 h-3 ml-1" /> : <ChevronDown className="w-3 h-3 ml-1" />}
        </Button>

        <div className="flex-1" />

        {/* 结果统计 */}
        <div className="text-xs text-gray-500">
          <span className="font-medium text-gray-700">{resultCount}</span>
          <span className="mx-1">/</span>
          <span>{totalCount}</span>
          <span className="ml-1">条</span>
        </div>

        {/* 清除筛选 */}
        {hasActiveFilters && (
          <Button
            size="sm"
            variant="ghost"
            onClick={clearFilters}
            className="h-7 px-2 text-xs text-red-500 hover:text-red-600 hover:bg-red-50"
          >
            <X className="w-3.5 h-3.5 mr-1" />
            清除
          </Button>
        )}
      </div>

      {/* 高级筛选栏 */}
      {expanded && (
        <div className="flex items-center px-3 py-2 gap-2 border-t border-gray-100 bg-gray-50/50">
          {/* Host 筛选 */}
          <div className="flex items-center gap-1">
            <Globe className="w-3.5 h-3.5 text-gray-400" />
            <Select value={filters.hostFilter} onValueChange={(v) => updateFilter('hostFilter', v)}>
              <SelectTrigger className="w-40 h-7 text-xs">
                <SelectValue placeholder="Host" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="all">全部 Host</SelectItem>
                {hostList.map((host) => (
                  <SelectItem key={host} value={host}>
                    {host}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>

          {/* Content-Type 筛选 */}
          <div className="flex items-center gap-1">
            <FileCode className="w-3.5 h-3.5 text-gray-400" />
            <Select value={filters.contentTypeFilter} onValueChange={(v) => updateFilter('contentTypeFilter', v)}>
              <SelectTrigger className="w-36 h-7 text-xs">
                <SelectValue placeholder="类型" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="all">全部类型</SelectItem>
                <SelectItem value="json">JSON</SelectItem>
                <SelectItem value="html">HTML</SelectItem>
                <SelectItem value="javascript">JavaScript</SelectItem>
                <SelectItem value="css">CSS</SelectItem>
                <SelectItem value="image">图片</SelectItem>
                <SelectItem value="form">Form Data</SelectItem>
                <SelectItem value="xml">XML</SelectItem>
                <SelectItem value="text">Text</SelectItem>
              </SelectContent>
            </Select>
          </div>

          {/* 响应大小筛选 */}
          <div className="flex items-center gap-1">
            <HardDrive className="w-3.5 h-3.5 text-gray-400" />
            <span className="text-[10px] text-gray-500">大小:</span>
            <Input
              value={filters.minSize}
              onChange={(e) => updateFilter('minSize', e.target.value)}
              placeholder="最小"
              className="w-14 h-7 text-xs text-center"
              type="number"
            />
            <span className="text-[10px] text-gray-400">-</span>
            <Input
              value={filters.maxSize}
              onChange={(e) => updateFilter('maxSize', e.target.value)}
              placeholder="最大"
              className="w-14 h-7 text-xs text-center"
              type="number"
            />
            <span className="text-[10px] text-gray-400">KB</span>
          </div>

          {/* 响应时间筛选 */}
          <div className="flex items-center gap-1">
            <Clock className="w-3.5 h-3.5 text-gray-400" />
            <span className="text-[10px] text-gray-500">时间:</span>
            <Input
              value={filters.minTime}
              onChange={(e) => updateFilter('minTime', e.target.value)}
              placeholder="最小"
              className="w-14 h-7 text-xs text-center"
              type="number"
            />
            <span className="text-[10px] text-gray-400">-</span>
            <Input
              value={filters.maxTime}
              onChange={(e) => updateFilter('maxTime', e.target.value)}
              placeholder="最大"
              className="w-14 h-7 text-xs text-center"
              type="number"
            />
            <span className="text-[10px] text-gray-400">ms</span>
          </div>
        </div>
      )}
    </div>
  )
}

// 辅助函数：应用筛选
export function applyTrafficFilters(items: any[], filters: TrafficFilterState): any[] {
  return items.filter((item) => {
    // 搜索查询
    if (filters.searchQuery) {
      const q = filters.searchQuery.toLowerCase()
      const matchUrl = (item.host + item.path).toLowerCase().includes(q)
      const matchMethod = item.method?.toLowerCase().includes(q)
      if (!matchUrl && !matchMethod) return false
    }

    // 方法筛选
    if (filters.methodFilter !== 'all' && item.method !== filters.methodFilter) {
      return false
    }

    // 状态码筛选
    if (filters.statusFilter !== 'all') {
      const code = item.statusCode
      switch (filters.statusFilter) {
        case '2xx':
          if (code < 200 || code >= 300) return false
          break
        case '3xx':
          if (code < 300 || code >= 400) return false
          break
        case '4xx':
          if (code < 400 || code >= 500) return false
          break
        case '5xx':
          if (code < 500) return false
          break
        case 'error':
          if (code < 400) return false
          break
      }
    }

    // Host 筛选
    if (filters.hostFilter !== 'all' && item.host !== filters.hostFilter) {
      return false
    }

    // Content-Type 筛选
    if (filters.contentTypeFilter !== 'all') {
      const ct = (item.contentType || '').toLowerCase()
      switch (filters.contentTypeFilter) {
        case 'json':
          if (!ct.includes('json')) return false
          break
        case 'html':
          if (!ct.includes('html')) return false
          break
        case 'javascript':
          if (!ct.includes('javascript') && !ct.includes('js')) return false
          break
        case 'css':
          if (!ct.includes('css')) return false
          break
        case 'image':
          if (!ct.includes('image')) return false
          break
        case 'form':
          if (!ct.includes('form')) return false
          break
        case 'xml':
          if (!ct.includes('xml')) return false
          break
        case 'text':
          if (!ct.includes('text') || ct.includes('html')) return false
          break
      }
    }

    // 大小筛选 (KB)
    const sizeKB = (item.responseSize || 0) / 1024
    if (filters.minSize && sizeKB < parseFloat(filters.minSize)) return false
    if (filters.maxSize && sizeKB > parseFloat(filters.maxSize)) return false

    // 时间筛选 (ms)
    const time = item.duration || 0
    if (filters.minTime && time < parseFloat(filters.minTime)) return false
    if (filters.maxTime && time > parseFloat(filters.maxTime)) return false

    return true
  })
}

// 辅助函数：从流量列表提取 Host 列表
export function extractHosts(items: any[]): string[] {
  const hosts = new Set<string>()
  items.forEach((item) => {
    if (item.host) hosts.add(item.host)
  })
  return Array.from(hosts).sort()
}
