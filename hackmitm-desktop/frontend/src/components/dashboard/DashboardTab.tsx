import { useEffect } from 'react'
import { Activity, Zap, Clock, Network, ArrowUpRight, ArrowDownRight, AlertCircle, TrendingUp, Timer } from 'lucide-react'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { useDashboardStore, useProxyStore } from '@/store'
import { GetMetrics } from '../../../wailsjs/go/main/App'
import { EventsOn } from '../../../wailsjs/runtime/runtime'
import { formatBytes, formatDuration } from '@/lib/utils'
import type { DashboardMetrics } from '@/types'
import {
  LineChart,
  Line,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  ResponsiveContainer,
} from 'recharts'

export function DashboardTab() {
  const { metrics, chartData, setMetrics, addChartData } = useDashboardStore()
  const { connected, status } = useProxyStore()

  // Listen for metrics updates from backend
  useEffect(() => {
    const unsubscribe = EventsOn('dashboard:metrics', (newMetrics: DashboardMetrics) => {
      setMetrics(newMetrics)
      addChartData({
        time: new Date().toLocaleTimeString(),
        requests: newMetrics.qps || 0,
        responseTime: newMetrics.avgResponseTime || 0
      })
    })
    return () => {
      unsubscribe()
    }
  }, [setMetrics, addChartData])

  // Fetch initial data
  useEffect(() => {
    if (connected) {
      GetMetrics().then(setMetrics).catch(console.error)
    }
  }, [connected, setMetrics])

  const statCards = [
    {
      title: 'Requests/sec',
      value: metrics.qps?.toFixed(1) || '0.0',
      icon: Zap,
      iconColor: 'text-orange-500',
      iconBg: 'bg-orange-50',
    },
    {
      title: 'Avg Response',
      value: formatDuration(metrics.avgResponseTime || 0),
      icon: Clock,
      iconColor: 'text-blue-600',
      iconBg: 'bg-blue-50',
    },
    {
      title: 'Active Conns',
      value: String(status.activeConnections || metrics.activeConnections || 0),
      icon: Network,
      iconColor: 'text-green-600',
      iconBg: 'bg-green-50',
    },
    {
      title: 'Total Requests',
      value: (metrics.totalRequests || 0).toLocaleString(),
      icon: Activity,
      iconColor: 'text-purple-600',
      iconBg: 'bg-purple-50',
    },
  ]

  const trafficCards = [
    {
      title: 'Bytes In',
      value: formatBytes(metrics.totalBytesIn || 0),
      icon: ArrowDownRight,
      iconColor: 'text-green-600',
    },
    {
      title: 'Bytes Out',
      value: formatBytes(metrics.totalBytesOut || 0),
      icon: ArrowUpRight,
      iconColor: 'text-blue-600',
    },
    {
      title: 'Error Rate',
      value: `${((metrics.errorRate || 0) * 100).toFixed(1)}%`,
      icon: AlertCircle,
      iconColor: (metrics.errorRate || 0) > 0.05 ? 'text-red-500' : 'text-green-600',
    },
  ]

  return (
    <div className="flex flex-col h-full overflow-auto bg-gray-50">
      {/* Header section */}
      <div className="flex items-center justify-between px-6 py-4 bg-white border-b border-gray-200">
        <div className="flex items-center gap-4">
          <h2 className="text-xl font-bold text-gray-800">Dashboard</h2>
          <Badge
            variant="outline"
            className={connected ? 'border-green-500 text-green-600' : 'border-red-500 text-red-600'}
          >
            {connected ? 'Connected' : 'Disconnected'}
          </Badge>
        </div>
        <div className="text-sm text-gray-500">
          Uptime: {formatDuration((metrics.uptime || 0) * 1000)}
        </div>
      </div>

      <div className="flex-1 p-6 space-y-6">
        {/* Top stat cards - 一行显示 */}
        <div className="grid grid-cols-7 gap-4">
          {statCards.map((stat) => {
            const Icon = stat.icon
            return (
              <Card key={stat.title} className="bg-white shadow-sm border-gray-200">
                <CardContent className="p-4">
                  <div className="flex items-center gap-3">
                    <div className={`p-2.5 rounded-lg ${stat.iconBg}`}>
                      <Icon className={`w-5 h-5 ${stat.iconColor}`} />
                    </div>
                    <div>
                      <p className="text-xs text-gray-500">{stat.title}</p>
                      <p className="text-xl font-bold text-gray-800">{stat.value}</p>
                    </div>
                  </div>
                </CardContent>
              </Card>
            )
          })}

          {/* Traffic cards - 合并到同一行 */}
          {trafficCards.map((stat) => {
            const Icon = stat.icon
            return (
              <Card key={stat.title} className="bg-white shadow-sm border-gray-200">
                <CardContent className="p-4">
                  <div className="flex items-center gap-3">
                    <Icon className={`w-5 h-5 ${stat.iconColor}`} />
                    <div>
                      <p className="text-xs text-gray-500">{stat.title}</p>
                      <p className="text-xl font-bold text-gray-800">{stat.value}</p>
                    </div>
                  </div>
                </CardContent>
              </Card>
            )
          })}
        </div>

        {/* Charts - 垂直堆叠，每个图表占满宽度 */}
        <div className="space-y-4">
          {/* Requests/sec Chart */}
          <Card className="bg-white shadow-sm border-gray-200">
            <CardHeader className="py-3 px-4 border-b border-gray-100">
              <div className="flex items-center gap-2">
                <TrendingUp className="w-4 h-4 text-blue-600" />
                <CardTitle className="text-sm font-medium text-gray-700">Requests/sec</CardTitle>
              </div>
            </CardHeader>
            <CardContent className="p-4">
              <div className="h-48">
                <ResponsiveContainer width="100%" height="100%">
                  <LineChart data={chartData} margin={{ top: 5, right: 20, left: 0, bottom: 5 }}>
                    <CartesianGrid strokeDasharray="3 3" stroke="#E5E7EB" />
                    <XAxis
                      dataKey="time"
                      tick={{ fill: '#6B7280', fontSize: 10 }}
                      tickLine={false}
                    />
                    <YAxis
                      tick={{ fill: '#6B7280', fontSize: 10 }}
                      tickLine={false}
                      axisLine={false}
                    />
                    <Tooltip
                      contentStyle={{
                        backgroundColor: 'white',
                        border: '1px solid #E5E7EB',
                        borderRadius: '6px',
                        fontSize: '12px',
                      }}
                    />
                    <Line
                      type="monotone"
                      dataKey="requests"
                      stroke="#0054A6"
                      strokeWidth={2}
                      dot={false}
                    />
                  </LineChart>
                </ResponsiveContainer>
              </div>
            </CardContent>
          </Card>

          {/* Response Time Chart */}
          <Card className="bg-white shadow-sm border-gray-200">
            <CardHeader className="py-3 px-4 border-b border-gray-100">
              <div className="flex items-center gap-2">
                <Timer className="w-4 h-4 text-orange-500" />
                <CardTitle className="text-sm font-medium text-gray-700">Response Time (ms)</CardTitle>
              </div>
            </CardHeader>
            <CardContent className="p-4">
              <div className="h-48">
                <ResponsiveContainer width="100%" height="100%">
                  <LineChart data={chartData} margin={{ top: 5, right: 20, left: 0, bottom: 5 }}>
                    <CartesianGrid strokeDasharray="3 3" stroke="#E5E7EB" />
                    <XAxis
                      dataKey="time"
                      tick={{ fill: '#6B7280', fontSize: 10 }}
                      tickLine={false}
                    />
                    <YAxis
                      tick={{ fill: '#6B7280', fontSize: 10 }}
                      tickLine={false}
                      axisLine={false}
                    />
                    <Tooltip
                      contentStyle={{
                        backgroundColor: 'white',
                        border: '1px solid #E5E7EB',
                        borderRadius: '6px',
                        fontSize: '12px',
                      }}
                    />
                    <Line
                      type="monotone"
                      dataKey="responseTime"
                      stroke="#FF6600"
                      strokeWidth={2}
                      dot={false}
                    />
                  </LineChart>
                </ResponsiveContainer>
              </div>
            </CardContent>
          </Card>
        </div>
      </div>
    </div>
  )
}
