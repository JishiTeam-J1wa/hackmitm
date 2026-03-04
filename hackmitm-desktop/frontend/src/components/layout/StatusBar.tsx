import { Circle } from 'lucide-react'
import { cn } from '@/lib/utils'
import { useProxyStore, useDashboardStore } from '@/store'
import { formatBytes, formatDuration } from '@/lib/utils'
import { useEffect } from 'react'
import { EventsOn } from '../../../wailsjs/runtime/runtime'
import { GetProxyStatus } from '../../../wailsjs/go/main/App'

export function StatusBar() {
  const { status, connected } = useProxyStore()
  const { metrics } = useDashboardStore()

  // 监听代理状态更新
  useEffect(() => {
    if (!connected) return

    // 获取初始状态
    GetProxyStatus().then(_status => {
      // 状态更新逻辑
    }).catch(console.error)

    // 监听状态更新事件
    const unsubscribe = EventsOn('proxy:status', (_newStatus: unknown) => {
      // 状态更新逻辑
    })

    return () => {
      unsubscribe()
    }
  }, [connected])

  return (
    <footer className="status-bar">
      {/* Connection status */}
      <div className="status-bar-item flex items-center gap-1">
        <Circle
          className={cn(
            'w-2 h-2',
            connected ? 'fill-green-500 text-green-500' : 'fill-red-500 text-red-500'
          )}
        />
        <span>{connected ? '已连接' : '未连接'}</span>
      </div>

      {/* Port - 从服务器状态获取 */}
      <div className="status-bar-item">
        端口: {status.port || 4443}
      </div>

      {/* Connections */}
      <div className="status-bar-item">
        连接: {status.activeConnections || 0}
      </div>

      {/* Requests */}
      <div className="status-bar-item">
        请求: {status.totalRequests || 0}
      </div>

      {/* Spacer */}
      <div className="status-bar-spacer" />

      {/* QPS */}
      <div className="status-bar-item">
        QPS: {metrics.qps?.toFixed(1) || '0.0'}
      </div>

      {/* Average response time */}
      <div className="status-bar-item">
        延迟: {formatDuration(metrics.avgResponseTime || 0)}
      </div>

      {/* Traffic */}
      <div className="status-bar-item">
        流量: {formatBytes(metrics.totalBytesIn || 0)} / {formatBytes(metrics.totalBytesOut || 0)}
      </div>

      {/* Version */}
      <div className="status-bar-item">
        HackMITM v1.0.0
      </div>
    </footer>
  )
}
