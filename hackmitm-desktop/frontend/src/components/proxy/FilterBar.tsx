import { useState, useCallback } from 'react'
import {
  Search,
  Filter,
  X,
  Save,
  Bookmark,
  Regex,
  ChevronDown,
  Eraser,
} from 'lucide-react'
import { Input } from '@/components/ui/input'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import { useTrafficStore } from '@/store'
import { cn } from '@/lib/utils'

const methods = ['ALL', 'GET', 'POST', 'PUT', 'DELETE', 'PATCH', 'HEAD', 'OPTIONS']

const statusGroups = [
  { value: 'ALL', label: 'All Status' },
  { value: '2xx', label: '2xx (Success)' },
  { value: '3xx', label: '3xx (Redirect)' },
  { value: '4xx', label: '4xx (Client Error)' },
  { value: '5xx', label: '5xx (Server Error)' },
]

interface SavedFilter {
  id: string
  name: string
  filter: {
    method: string
    host: string
    path: string
    statusCode: string
    search: string
    useRegex: boolean
  }
}

export function FilterBar() {
  const { filter, setFilter, resetFilter } = useTrafficStore()
  const [useRegex, setUseRegex] = useState(false)
  const [savedFilters, setSavedFilters] = useState<SavedFilter[]>([])
  const [showAdvanced, setShowAdvanced] = useState(false)

  const hasActiveFilters =
    filter.method || filter.host || filter.path || filter.statusCode || filter.search

  const handleSaveFilter = useCallback(() => {
    const name = prompt('Enter filter name:')
    if (name) {
      const newFilter: SavedFilter = {
        id: Date.now().toString(),
        name,
        filter: { ...filter, useRegex },
      }
      setSavedFilters((prev) => [...prev, newFilter])
    }
  }, [filter, useRegex])

  const handleLoadFilter = useCallback((saved: SavedFilter) => {
    setFilter(saved.filter)
    setUseRegex(saved.filter.useRegex)
  }, [setFilter])

  const handleDeleteFilter = useCallback((id: string) => {
    setSavedFilters((prev) => prev.filter((f) => f.id !== id))
  }, [])

  const activeFilterCount = [
    filter.method,
    filter.host,
    filter.path,
    filter.statusCode,
    filter.search,
  ].filter(Boolean).length

  return (
    <div className="flex flex-col gap-2">
      {/* Main filter row */}
      <div className="flex items-center gap-2 flex-wrap">
        {/* Method filter */}
        <Select
          value={filter.method || 'ALL'}
          onValueChange={(value) => setFilter({ method: value === 'ALL' ? '' : value })}
        >
          <SelectTrigger className="w-24 h-8 text-xs">
            <SelectValue placeholder="Method" />
          </SelectTrigger>
          <SelectContent>
            {methods.map((method) => (
              <SelectItem key={method} value={method} className="text-xs">
                {method}
              </SelectItem>
            ))}
          </SelectContent>
        </Select>

        {/* Status code filter */}
        <Select
          value={filter.statusCode || 'ALL'}
          onValueChange={(value) => setFilter({ statusCode: value === 'ALL' ? '' : value })}
        >
          <SelectTrigger className="w-28 h-8 text-xs">
            <SelectValue placeholder="Status" />
          </SelectTrigger>
          <SelectContent>
            {statusGroups.map((group) => (
              <SelectItem key={group.value} value={group.value} className="text-xs">
                {group.label}
              </SelectItem>
            ))}
          </SelectContent>
        </Select>

        {/* Host filter */}
        <Input
          placeholder="Host"
          value={filter.host}
          onChange={(e) => setFilter({ host: e.target.value })}
          className="w-32 h-8 text-xs"
        />

        {/* Search with regex toggle */}
        <div className="relative flex-1 min-w-[200px] max-w-[400px]">
          <Search className="absolute left-2 top-1/2 -translate-y-1/2 w-3.5 h-3.5 text-muted-foreground" />
          <Input
            placeholder={useRegex ? 'Search (regex)...' : 'Search...'}
            value={filter.search}
            onChange={(e) => setFilter({ search: e.target.value })}
            className={cn('h-8 pl-7 pr-8 text-xs', useRegex && 'font-mono')}
          />
          <Button
            variant="ghost"
            size="icon"
            className="absolute right-1 top-1/2 -translate-y-1/2 w-6 h-6"
            onClick={() => setUseRegex(!useRegex)}
            title={useRegex ? 'Disable regex' : 'Enable regex'}
          >
            <Regex className={cn('w-3 h-3', useRegex ? 'text-blue-500' : 'text-gray-400')} />
          </Button>
        </div>

        {/* Clear filters */}
        {hasActiveFilters && (
          <Button variant="ghost" size="sm" onClick={resetFilter} className="h-8 text-xs">
            <Eraser className="w-3 h-3 mr-1" />
            Clear
          </Button>
        )}

        {/* Advanced toggle */}
        <Button
          variant="ghost"
          size="sm"
          onClick={() => setShowAdvanced(!showAdvanced)}
          className={cn('h-8 text-xs', showAdvanced && 'bg-blue-50 text-blue-600')}
        >
          <Filter className="w-3 h-3 mr-1" />
          Advanced
          <ChevronDown className={cn('w-3 h-3 ml-1', showAdvanced && 'rotate-180')} />
        </Button>

        {/* Saved filters dropdown */}
        <DropdownMenu>
          <DropdownMenuTrigger asChild>
            <Button variant="outline" size="sm" className="h-8 text-xs">
              <Bookmark className="w-3 h-3 mr-1" />
              Saved
              {savedFilters.length > 0 && (
                <Badge variant="secondary" className="ml-1 h-4 px-1 text-[10px]">
                  {savedFilters.length}
                </Badge>
              )}
            </Button>
          </DropdownMenuTrigger>
          <DropdownMenuContent align="end" className="w-56">
            {savedFilters.length === 0 ? (
              <div className="px-2 py-4 text-center text-xs text-gray-500">
                No saved filters
              </div>
            ) : (
              savedFilters.map((saved) => (
                <DropdownMenuItem
                  key={saved.id}
                  className="flex items-center justify-between"
                  onSelect={() => handleLoadFilter(saved)}
                >
                  <span className="truncate">{saved.name}</span>
                  <Button
                    variant="ghost"
                    size="icon"
                    className="w-6 h-6"
                    onClick={(e) => {
                      e.stopPropagation()
                      handleDeleteFilter(saved.id)
                    }}
                  >
                    <X className="w-3 h-3" />
                  </Button>
                </DropdownMenuItem>
              ))
            )}
            <DropdownMenuSeparator />
            <DropdownMenuItem onClick={handleSaveFilter}>
              <Save className="w-3 h-3 mr-2" />
              Save current filter
            </DropdownMenuItem>
          </DropdownMenuContent>
        </DropdownMenu>

        {/* Active filter count badge */}
        {activeFilterCount > 0 && (
          <Badge variant="secondary" className="text-xs">
            {activeFilterCount} filter{activeFilterCount > 1 ? 's' : ''} active
          </Badge>
        )}
      </div>

      {/* Advanced filters row */}
      {showAdvanced && (
        <div className="flex items-center gap-2 pl-2 border-l-2 border-blue-200">
          <span className="text-xs text-gray-500">Path:</span>
          <Input
            placeholder="Path pattern"
            value={filter.path}
            onChange={(e) => setFilter({ path: e.target.value })}
            className="w-40 h-7 text-xs"
          />
        </div>
      )}
    </div>
  )
}
