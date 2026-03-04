import { useMemo } from 'react'
import {
  BarChart,
  Bar,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  ResponsiveContainer,
  PieChart,
  Pie,
  Cell,
} from 'recharts'
import {
  Globe,
  Clock,
  TrendingUp,
  Activity,
  AlertTriangle,
  Zap,
} from 'lucide-react'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { cn } from '@/lib/utils'
import { useTrafficStore } from '@/store'

interface AnalysisPanelProps {
  className?: string
}

export function AnalysisPanel({ className }: AnalysisPanelProps) {
  const { items } = useTrafficStore()

  // Top URLs analysis
  const topUrls = useMemo(() => {
    const urlCounts = new Map<string, { count: number; avgTime: number; totalTime: number }>()

    items.forEach((item) => {
      const key = `${item.method} ${item.host}${item.path}`
      const existing = urlCounts.get(key)
      if (existing) {
        existing.count++
        existing.totalTime += item.duration
        existing.avgTime = existing.totalTime / existing.count
      } else {
        urlCounts.set(key, { count: 1, avgTime: item.duration, totalTime: item.duration })
      }
    })

    return Array.from(urlCounts.entries())
      .map(([url, stats]) => ({ url, ...stats }))
      .sort((a, b) => b.count - a.count)
      .slice(0, 10)
  }, [items])

  // Response time distribution
  const responseTimeDistribution = useMemo(() => {
    const ranges = [
      { name: '< 100ms', min: 0, max: 100, count: 0 },
      { name: '100-500ms', min: 100, max: 500, count: 0 },
      { name: '500ms-1s', min: 500, max: 1000, count: 0 },
      { name: '1-2s', min: 1000, max: 2000, count: 0 },
      { name: '> 2s', min: 2000, max: Infinity, count: 0 },
    ]

    items.forEach((item) => {
      const time = item.duration
      for (const range of ranges) {
        if (time >= range.min && time < range.max) {
          range.count++
          break
        }
      }
    })

    return ranges
  }, [items])

  // Status code distribution
  const statusCodeDistribution = useMemo(() => {
    const statusCodes = new Map<string, number>()

    items.forEach((item) => {
      const category = Math.floor(item.statusCode / 100) * 100
      const key = `${category}xx`
      statusCodes.set(key, (statusCodes.get(key) || 0) + 1)
    })

    return Array.from(statusCodes.entries())
      .map(([name, value]) => ({ name, value }))
      .sort((a, b) => b.value - a.value)
  }, [items])

  // Slow endpoints
  const slowEndpoints = useMemo(() => {
    return items
      .filter((item) => item.duration > 1000)
      .sort((a, b) => b.duration - a.duration)
      .slice(0, 5)
      .map((item) => ({
        url: `${item.method} ${item.host}${item.path}`,
        duration: item.duration,
        statusCode: item.statusCode,
      }))
  }, [items])

  // Stats
  const stats = useMemo(() => {
    if (items.length === 0) {
      return {
        totalRequests: 0,
        avgResponseTime: 0,
        maxResponseTime: 0,
        minResponseTime: 0,
        errorRate: 0,
        slowRequests: 0,
      }
    }

    const times = items.map((i) => i.duration)
    const errors = items.filter((i) => i.statusCode >= 400).length
    const slow = items.filter((i) => i.duration > 1000).length

    return {
      totalRequests: items.length,
      avgResponseTime: Math.round(times.reduce((a, b) => a + b, 0) / times.length),
      maxResponseTime: Math.max(...times),
      minResponseTime: Math.min(...times),
      errorRate: ((errors / items.length) * 100).toFixed(1),
      slowRequests: slow,
    }
  }, [items])

  return (
    <div className={cn('space-y-4', className)}>
      {/* Quick Stats */}
      <div className="grid grid-cols-4 gap-3">
        <Card className="bg-white">
          <CardContent className="p-3">
            <div className="flex items-center gap-2">
              <Activity className="w-4 h-4 text-blue-500" />
              <div>
                <p className="text-[10px] text-gray-500">Total Requests</p>
                <p className="text-lg font-bold text-gray-800">{stats.totalRequests.toLocaleString()}</p>
              </div>
            </div>
          </CardContent>
        </Card>

        <Card className="bg-white">
          <CardContent className="p-3">
            <div className="flex items-center gap-2">
              <Clock className="w-4 h-4 text-orange-500" />
              <div>
                <p className="text-[10px] text-gray-500">Avg Response</p>
                <p className="text-lg font-bold text-gray-800">{stats.avgResponseTime}ms</p>
              </div>
            </div>
          </CardContent>
        </Card>

        <Card className="bg-white">
          <CardContent className="p-3">
            <div className="flex items-center gap-2">
              <AlertTriangle className="w-4 h-4 text-red-500" />
              <div>
                <p className="text-[10px] text-gray-500">Error Rate</p>
                <p className="text-lg font-bold text-gray-800">{stats.errorRate}%</p>
              </div>
            </div>
          </CardContent>
        </Card>

        <Card className="bg-white">
          <CardContent className="p-3">
            <div className="flex items-center gap-2">
              <Zap className="w-4 h-4 text-yellow-500" />
              <div>
                <p className="text-[10px] text-gray-500">Slow Requests</p>
                <p className="text-lg font-bold text-gray-800">{stats.slowRequests}</p>
              </div>
            </div>
          </CardContent>
        </Card>
      </div>

      {/* Charts Row */}
      <div className="grid grid-cols-2 gap-4">
        {/* Top URLs */}
        <Card className="bg-white">
          <CardHeader className="py-3 px-4">
            <CardTitle className="text-sm font-medium flex items-center gap-2">
              <Globe className="w-4 h-4 text-blue-500" />
              Top URLs
            </CardTitle>
          </CardHeader>
          <CardContent className="px-4 pb-4">
            {topUrls.length === 0 ? (
              <div className="h-48 flex items-center justify-center text-gray-400 text-xs">
                No data available
              </div>
            ) : (
              <div className="h-48">
                <ResponsiveContainer width="100%" height="100%">
                  <BarChart data={topUrls} layout="vertical" margin={{ left: 10, right: 10 }}>
                    <CartesianGrid strokeDasharray="3 3" stroke="#E5E7EB" />
                    <XAxis type="number" tick={{ fill: '#6B7280', fontSize: 10 }} />
                    <YAxis
                      type="category"
                      dataKey="url"
                      tick={{ fill: '#6B7280', fontSize: 9 }}
                      width={150}
                      tickFormatter={(value) => value.length > 25 ? `${value.slice(0, 25)}...` : value}
                    />
                    <Tooltip
                      contentStyle={{
                        backgroundColor: 'white',
                        border: '1px solid #E5E7EB',
                        borderRadius: '6px',
                        fontSize: '11px',
                      }}
                      formatter={(value: number, name: string) => [
                        name === 'count' ? value : `${value.toFixed(0)}ms`,
                        name === 'count' ? 'Requests' : 'Avg Time',
                      ]}
                    />
                    <Bar dataKey="count" fill="#0054A6" radius={[0, 4, 4, 0]} />
                  </BarChart>
                </ResponsiveContainer>
              </div>
            )}
          </CardContent>
        </Card>

        {/* Response Time Distribution */}
        <Card className="bg-white">
          <CardHeader className="py-3 px-4">
            <CardTitle className="text-sm font-medium flex items-center gap-2">
              <Clock className="w-4 h-4 text-orange-500" />
              Response Time Distribution
            </CardTitle>
          </CardHeader>
          <CardContent className="px-4 pb-4">
            {responseTimeDistribution.every((r) => r.count === 0) ? (
              <div className="h-48 flex items-center justify-center text-gray-400 text-xs">
                No data available
              </div>
            ) : (
              <div className="h-48">
                <ResponsiveContainer width="100%" height="100%">
                  <BarChart data={responseTimeDistribution} margin={{ top: 5, right: 10, left: 10, bottom: 5 }}>
                    <CartesianGrid strokeDasharray="3 3" stroke="#E5E7EB" />
                    <XAxis dataKey="name" tick={{ fill: '#6B7280', fontSize: 10 }} />
                    <YAxis tick={{ fill: '#6B7280', fontSize: 10 }} />
                    <Tooltip
                      contentStyle={{
                        backgroundColor: 'white',
                        border: '1px solid #E5E7EB',
                        borderRadius: '6px',
                        fontSize: '11px',
                      }}
                    />
                    <Bar dataKey="count" fill="#FF6600" radius={[4, 4, 0, 0]}>
                      {responseTimeDistribution.map((entry, index) => (
                        <Cell
                          key={`cell-${index}`}
                          fill={
                            entry.name.includes('> 2s') || entry.name.includes('1-2s')
                              ? '#EF4444'
                              : entry.name.includes('500ms')
                              ? '#F97316'
                              : '#22C55E'
                          }
                        />
                      ))}
                    </Bar>
                  </BarChart>
                </ResponsiveContainer>
              </div>
            )}
          </CardContent>
        </Card>
      </div>

      {/* Second Row */}
      <div className="grid grid-cols-3 gap-4">
        {/* Status Code Distribution */}
        <Card className="bg-white">
          <CardHeader className="py-3 px-4">
            <CardTitle className="text-sm font-medium">Status Codes</CardTitle>
          </CardHeader>
          <CardContent className="px-4 pb-4">
            {statusCodeDistribution.length === 0 ? (
              <div className="h-32 flex items-center justify-center text-gray-400 text-xs">
                No data
              </div>
            ) : (
              <div className="h-32">
                <ResponsiveContainer width="100%" height="100%">
                  <PieChart>
                    <Pie
                      data={statusCodeDistribution}
                      cx="50%"
                      cy="50%"
                      innerRadius={25}
                      outerRadius={45}
                      paddingAngle={2}
                      dataKey="value"
                    >
                      {statusCodeDistribution.map((entry, index) => (
                        <Cell
                          key={`cell-${index}`}
                          fill={
                            entry.name === '2xx'
                              ? '#22C55E'
                              : entry.name === '3xx'
                              ? '#3B82F6'
                              : entry.name === '4xx'
                              ? '#F97316'
                              : '#EF4444'
                          }
                        />
                      ))}
                    </Pie>
                    <Tooltip
                      contentStyle={{
                        backgroundColor: 'white',
                        border: '1px solid #E5E7EB',
                        borderRadius: '6px',
                        fontSize: '11px',
                      }}
                    />
                  </PieChart>
                </ResponsiveContainer>
                <div className="flex justify-center gap-3 mt-2">
                  {statusCodeDistribution.map((entry) => (
                    <div key={entry.name} className="flex items-center gap-1 text-[10px]">
                      <div
                        className="w-2 h-2 rounded-full"
                        style={{
                          backgroundColor:
                            entry.name === '2xx'
                              ? '#22C55E'
                              : entry.name === '3xx'
                              ? '#3B82F6'
                              : entry.name === '4xx'
                              ? '#F97316'
                              : '#EF4444',
                        }}
                      />
                      <span className="text-gray-600">{entry.name}</span>
                    </div>
                  ))}
                </div>
              </div>
            )}
          </CardContent>
        </Card>

        {/* Slow Endpoints */}
        <Card className="bg-white col-span-2">
          <CardHeader className="py-3 px-4">
            <CardTitle className="text-sm font-medium flex items-center gap-2">
              <TrendingUp className="w-4 h-4 text-red-500" />
              Slowest Endpoints (&gt;1s)
            </CardTitle>
          </CardHeader>
          <CardContent className="px-4 pb-4">
            {slowEndpoints.length === 0 ? (
              <div className="h-32 flex items-center justify-center text-gray-400 text-xs">
                No slow endpoints detected
              </div>
            ) : (
              <div className="space-y-2">
                {slowEndpoints.map((endpoint, index) => (
                  <div
                    key={index}
                    className="flex items-center justify-between p-2 bg-gray-50 rounded text-xs"
                  >
                    <div className="flex items-center gap-2 flex-1 min-w-0">
                      <Badge
                        variant="outline"
                        className={cn(
                          'text-[10px] flex-shrink-0',
                          endpoint.statusCode >= 400 ? 'text-red-500' : 'text-gray-600'
                        )}
                      >
                        {endpoint.url.split(' ')[0]}
                      </Badge>
                      <span className="truncate text-gray-600">{endpoint.url.split(' ').slice(1).join(' ')}</span>
                    </div>
                    <div className="flex items-center gap-2 text-red-600 font-medium ml-2">
                      <Clock className="w-3 h-3" />
                      {endpoint.duration}ms
                    </div>
                  </div>
                ))}
              </div>
            )}
          </CardContent>
        </Card>
      </div>
    </div>
  )
}

export default AnalysisPanel
