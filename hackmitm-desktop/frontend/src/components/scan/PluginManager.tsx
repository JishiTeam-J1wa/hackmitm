import { useState, useMemo } from 'react'
import {
  Search,
  ToggleLeft,
  ToggleRight,
  Info,
  ChevronDown,
  ChevronRight,
  Shield,
} from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Badge } from '@/components/ui/badge'
import { Switch } from '@/components/ui/switch'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { cn } from '@/lib/utils'
import { useScanStore } from '@/store'
import type { ScanPlugin } from '@/types'

const categoryColors: Record<string, string> = {
  'Injection': 'bg-red-100 text-red-700',
  'Information Disclosure': 'bg-yellow-100 text-yellow-700',
  'Security Misconfiguration': 'bg-purple-100 text-purple-700',
  'Authentication': 'bg-blue-100 text-blue-700',
  'default': 'bg-gray-100 text-gray-700',
}

const severityColors: Record<string, string> = {
  'high': 'text-red-600',
  'medium': 'text-orange-600',
  'low': 'text-blue-600',
}

export function PluginManager() {
  const { plugins, togglePlugin, updatePluginConfig } = useScanStore()
  const [searchQuery, setSearchQuery] = useState('')
  const [categoryFilter, setCategoryFilter] = useState('all')
  const [expandedPlugin, setExpandedPlugin] = useState<string | null>(null)
  const [expandedCategories, setExpandedCategories] = useState<Set<string>>(new Set(['Injection', 'Information Disclosure']))

  // Get unique categories
  const categories = useMemo(() => {
    const cats = new Set(plugins.map((p) => p.category))
    return Array.from(cats).sort()
  }, [plugins])

  // Filter plugins
  const filteredPlugins = useMemo(() => {
    let result = [...plugins]

    if (categoryFilter !== 'all') {
      result = result.filter((p) => p.category === categoryFilter)
    }

    if (searchQuery) {
      const query = searchQuery.toLowerCase()
      result = result.filter(
        (p) =>
          p.name.toLowerCase().includes(query) ||
          p.description.toLowerCase().includes(query)
      )
    }

    return result
  }, [plugins, categoryFilter, searchQuery])

  // Group plugins by category
  const groupedPlugins = useMemo(() => {
    const groups: Record<string, ScanPlugin[]> = {}
    filteredPlugins.forEach((plugin) => {
      if (!groups[plugin.category]) {
        groups[plugin.category] = []
      }
      groups[plugin.category].push(plugin)
    })
    return groups
  }, [filteredPlugins])

  const toggleCategory = (category: string) => {
    const newExpanded = new Set(expandedCategories)
    if (newExpanded.has(category)) {
      newExpanded.delete(category)
    } else {
      newExpanded.add(category)
    }
    setExpandedCategories(newExpanded)
  }

  const PluginCard = ({ plugin }: { plugin: ScanPlugin }) => {
    const isExpanded = expandedPlugin === plugin.id

    return (
      <div className="border border-gray-200 rounded-lg overflow-hidden">
        {/* Plugin header */}
        <div
          className={cn(
            'px-3 py-2 flex items-center gap-3 cursor-pointer',
            plugin.enabled ? 'bg-white' : 'bg-gray-50'
          )}
          onClick={() => setExpandedPlugin(isExpanded ? null : plugin.id)}
        >
          <Switch
            checked={plugin.enabled}
            onCheckedChange={() => togglePlugin(plugin.id)}
            onClick={(e) => e.stopPropagation()}
          />

          <div className="flex-1 min-w-0">
            <div className="flex items-center gap-2">
              <span className={cn(
                'text-sm font-medium',
                plugin.enabled ? 'text-gray-900' : 'text-gray-500'
              )}>
                {plugin.name}
              </span>
              <Badge
                variant="outline"
                className={cn('text-[10px]', categoryColors[plugin.category] || categoryColors.default)}
              >
                {plugin.category}
              </Badge>
              <span className={cn('text-xs', severityColors[plugin.severity] || 'text-gray-500')}>
                {plugin.severity}
              </span>
            </div>
            <p className="text-xs text-gray-500 truncate">{plugin.description}</p>
          </div>

          <div className="text-xs text-gray-400">
            v{plugin.version}
          </div>

          {isExpanded ? (
            <ChevronDown className="w-4 h-4 text-gray-400" />
          ) : (
            <ChevronRight className="w-4 h-4 text-gray-400" />
          )}
        </div>

        {/* Expanded details */}
        {isExpanded && (
          <div className="border-t border-gray-200 bg-gray-50 p-3">
            <div className="grid grid-cols-2 gap-4 mb-3">
              <div>
                <span className="text-[10px] text-gray-500 block">Author</span>
                <span className="text-xs text-gray-700">{plugin.author}</span>
              </div>
              <div>
                <span className="text-[10px] text-gray-500 block">Severity</span>
                <span className={cn('text-xs font-medium', severityColors[plugin.severity])}>
                  {plugin.severity.toUpperCase()}
                </span>
              </div>
            </div>

            {/* Plugin config */}
            {Object.keys(plugin.config).length > 0 && (
              <div>
                <span className="text-[10px] text-gray-500 block mb-2">Configuration</span>
                <div className="bg-white rounded border p-2 space-y-2">
                  {Object.entries(plugin.config).map(([key, value]) => (
                    <div key={key} className="flex items-center justify-between">
                      <span className="text-xs text-gray-600">{key}</span>
                      {typeof value === 'boolean' ? (
                        <Switch
                          checked={value}
                          onCheckedChange={(checked) =>
                            updatePluginConfig(plugin.id, { [key]: checked })
                          }
                        />
                      ) : typeof value === 'string' ? (
                        <Input
                          value={value}
                          onChange={(e) =>
                            updatePluginConfig(plugin.id, { [key]: e.target.value })
                          }
                          className="h-6 w-40 text-xs"
                        />
                      ) : Array.isArray(value) ? (
                        <span className="text-xs text-gray-500">{value.length} items</span>
                      ) : (
                        <span className="text-xs text-gray-500">{String(value)}</span>
                      )}
                    </div>
                  ))}
                </div>
              </div>
            )}
          </div>
        )}
      </div>
    )
  }

  // Calculate stats
  const enabledCount = plugins.filter((p) => p.enabled).length
  const totalCount = plugins.length

  return (
    <div className="flex h-full flex-col">
      {/* Header */}
      <div className="h-9 border-b border-gray-200 bg-white px-3 flex items-center gap-2 flex-shrink-0">
        <Shield className="w-4 h-4 text-blue-500" />
        <span className="text-sm font-medium text-gray-700">Plugins</span>

        <div className="w-px h-4 bg-gray-200 mx-2" />

        <Badge variant="secondary" className="text-xs">
          {enabledCount}/{totalCount} enabled
        </Badge>

        <div className="flex-1" />

        {/* Enable/Disable all */}
        <Button
          variant="ghost"
          size="sm"
          onClick={() => {
            plugins.forEach((p) => {
              if (!p.enabled) togglePlugin(p.id)
            })
          }}
          className="h-6 text-xs"
        >
          <ToggleRight className="w-3 h-3 mr-1" />
          Enable All
        </Button>
        <Button
          variant="ghost"
          size="sm"
          onClick={() => {
            plugins.forEach((p) => {
              if (p.enabled) togglePlugin(p.id)
            })
          }}
          className="h-6 text-xs"
        >
          <ToggleLeft className="w-3 h-3 mr-1" />
          Disable All
        </Button>
      </div>

      {/* Toolbar */}
      <div className="h-9 border-b border-gray-200 bg-white px-3 flex items-center gap-2 flex-shrink-0">
        {/* Search */}
        <div className="relative flex-1 max-w-xs">
          <Search className="absolute left-2 top-1/2 -translate-y-1/2 w-3.5 h-3.5 text-gray-400" />
          <Input
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            placeholder="Search plugins..."
            className="h-7 pl-7 text-xs"
          />
        </div>

        {/* Category filter */}
        <Select value={categoryFilter} onValueChange={setCategoryFilter}>
          <SelectTrigger className="w-36 h-7 text-xs">
            <SelectValue />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="all">All Categories</SelectItem>
            {categories.map((cat) => (
              <SelectItem key={cat} value={cat}>
                {cat}
              </SelectItem>
            ))}
          </SelectContent>
        </Select>
      </div>

      {/* Plugin list */}
      <div className="flex-1 overflow-auto p-3">
        {categoryFilter === 'all' ? (
          // Grouped view
          <div className="space-y-4">
            {Object.entries(groupedPlugins).map(([category, categoryPlugins]) => (
              <div key={category}>
                {/* Category header */}
                <div
                  onClick={() => toggleCategory(category)}
                  className="flex items-center gap-2 mb-2 cursor-pointer select-none"
                >
                  {expandedCategories.has(category) ? (
                    <ChevronDown className="w-4 h-4 text-gray-400" />
                  ) : (
                    <ChevronRight className="w-4 h-4 text-gray-400" />
                  )}
                  <Badge
                    variant="outline"
                    className={cn('text-xs', categoryColors[category] || categoryColors.default)}
                  >
                    {category}
                  </Badge>
                  <span className="text-xs text-gray-500">
                    {categoryPlugins.filter((p) => p.enabled).length}/{categoryPlugins.length} enabled
                  </span>
                </div>

                {/* Plugins in category */}
                {expandedCategories.has(category) && (
                  <div className="space-y-2 ml-6">
                    {categoryPlugins.map((plugin) => (
                      <PluginCard key={plugin.id} plugin={plugin} />
                    ))}
                  </div>
                )}
              </div>
            ))}
          </div>
        ) : (
          // Flat view
          <div className="space-y-2">
            {filteredPlugins.map((plugin) => (
              <PluginCard key={plugin.id} plugin={plugin} />
            ))}
          </div>
        )}

        {filteredPlugins.length === 0 && (
          <div className="flex flex-col items-center justify-center h-32 text-gray-400">
            <Info className="w-8 h-8 mb-2" />
            <p className="text-sm">No plugins found</p>
          </div>
        )}
      </div>
    </div>
  )
}
