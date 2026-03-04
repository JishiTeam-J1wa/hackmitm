import { useState, useEffect } from 'react'
import { Wifi, WifiOff, Loader2, Circle, Monitor, Cloud } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { useProxyStore, useTrafficStore } from '@/store'
import { useAppStore } from '@/store/appStore'
import { Connect, Disconnect, SetAPIEndpoint, GetProxyStatus, GetConnectionMode } from '../../../wailsjs/go/main/App'

export function Header() {
  const { connected, apiEndpoint, setApiEndpoint, setConnected, setStatus } = useProxyStore()
  const { interceptMode, items } = useTrafficStore()
  const { appConfig } = useAppStore()
  const [connecting, setConnecting] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [connectionMode, setConnectionMode] = useState<string>('')

  // Get connection mode on mount
  useEffect(() => {
    const fetchMode = async () => {
      try {
        const mode = await GetConnectionMode()
        setConnectionMode(mode)
      } catch (e) {
        // Use stored config if API not available yet
        setConnectionMode(appConfig.connectionMode || '')
      }
    }
    fetchMode()
  }, [appConfig.connectionMode])

  // 监听连接状态变化
  useEffect(() => {
    if (connected) {
      // 定期获取状态
      const interval = setInterval(async () => {
        try {
          const status = await GetProxyStatus()
          setStatus({
            running: status.running,
            port: status.port,
            interceptMode: status.interceptMode,
            activeConnections: status.activeConnections,
            totalRequests: status.totalRequests,
            uptime: status.uptime
          })
        } catch (e) {
          console.error('Failed to get status:', e)
        }
      }, 3000)
      return () => clearInterval(interval)
    }
  }, [connected, setStatus])

  const handleConnect = async () => {
    if (connected) {
      try {
        setConnecting(true)
        setError(null)
        await Disconnect()
        setConnected(false)
      } catch (error) {
        setError(error instanceof Error ? error.message : String(error))
      } finally {
        setConnecting(false)
      }
    } else {
      try {
        setConnecting(true)
        setError(null)

        // 先设置API端点
        await SetAPIEndpoint(apiEndpoint)

        // 然后连接
        await Connect()

        setConnected(true)

        // 获取代理状态
        try {
          const status = await GetProxyStatus()
          setStatus({
            running: status.running,
            port: status.port,
            interceptMode: status.interceptMode,
            activeConnections: status.activeConnections,
            totalRequests: status.totalRequests,
            uptime: status.uptime
          })
        } catch {
          // Status fetch failed, but connection succeeded
        }
      } catch (error) {
        setError(error instanceof Error ? error.message : String(error))
        setConnected(false)
      } finally {
        setConnecting(false)
      }
    }
  }

  // Check if this is local mode - connection is auto-managed
  const isLocalMode = connectionMode === 'local'

  return (
    <div className="api-connection-panel">
      {/* Connection Mode Indicator */}
      {connectionMode && (
        <div className="flex items-center gap-1.5 px-2 py-0.5 rounded bg-gray-100 text-gray-600">
          {connectionMode === 'local' ? (
            <>
              <Monitor className="w-3 h-3" />
              <span className="text-[10px]">本地</span>
            </>
          ) : (
            <>
              <Cloud className="w-3 h-3" />
              <span className="text-[10px]">远程</span>
            </>
          )}
        </div>
      )}

      {/* API 端点输入 - only show in remote mode or when not connected in local mode */}
      {(!isLocalMode || !connected) && (
        <div className="flex items-center gap-2">
          <span className="text-gray-500 text-[10px] hidden sm:inline">API:</span>
          <Input
            value={apiEndpoint}
            onChange={(e) => setApiEndpoint(e.target.value)}
            placeholder="http://localhost:9090"
            className="w-36 h-6 text-xs bg-white border-gray-200 text-gray-700 placeholder:text-gray-400"
            disabled={connected || connecting}
          />

          {/* 连接/断开按钮 */}
          <Button
            variant={connected ? 'destructive' : 'default'}
            size="sm"
            onClick={handleConnect}
            className="h-6 text-xs px-3"
            disabled={connecting}
          >
            {connecting ? (
              <>
                <Loader2 className="w-3 h-3 mr-1 animate-spin" />
                {connected ? '断开中...' : '连接中...'}
              </>
            ) : connected ? (
              <>
                <WifiOff className="w-3 h-3 mr-1" />
                断开
              </>
            ) : (
              <>
                <Wifi className="w-3 h-3 mr-1" />
                连接
              </>
            )}
          </Button>
        </div>
      )}

      {/* 状态指示器 */}
      <div className={`api-status-indicator ${connected ? 'api-status-connected' : 'api-status-disconnected'}`}>
        <Circle
          className={`w-2 h-2 ${connected ? 'fill-green-500 text-green-500' : 'fill-red-500 text-red-500'}`}
        />
        <span>{connected ? '已连接' : '未连接'}</span>
      </div>

      {/* 拦截模式指示 */}
      {interceptMode && (
        <div className="api-status-indicator" style={{ background: 'rgba(249, 115, 22, 0.1)', color: '#ea580c' }}>
          <Circle className="w-2 h-2 fill-orange-500 text-orange-500 animate-pulse" />
          <span>拦截中</span>
        </div>
      )}

      {/* 请求数量 */}
      <div className="text-gray-500 text-[10px]">
        {items.length} 请求
      </div>

      {/* 错误提示 */}
      {error && (
        <div className="text-red-500 text-[10px] max-w-[200px] truncate" title={error}>
          {error}
        </div>
      )}
    </div>
  )
}
