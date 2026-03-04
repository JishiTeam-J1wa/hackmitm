import { useEffect, useRef, useCallback, useState } from 'react'
import { useTrafficStore, useVulnStore, useInterceptStore } from '@/store'
import type { TrafficItem, Vulnerability } from '@/types'

interface WebSocketMessage {
  type: string
  channel: string
  data: any
}

interface UseWebSocketOptions {
  url?: string
  autoConnect?: boolean
  onConnected?: () => void
  onDisconnected?: () => void
  onError?: (error: Event) => void
}

export function useWebSocket(options: UseWebSocketOptions = {}) {
  const {
    url = 'ws://localhost:9090/ws',
    autoConnect = true,
    onConnected,
    onDisconnected,
    onError,
  } = options

  const wsRef = useRef<WebSocket | null>(null)
  const reconnectTimeoutRef = useRef<NodeJS.Timeout | null>(null)
  const [connected, setConnected] = useState(false)
  const [reconnecting, setReconnecting] = useState(false)

  const { addItem: addTrafficItem } = useTrafficStore()
  const { addVulnerability } = useVulnStore()
  const { addRequest: addInterceptedRequest } = useInterceptStore

  // Handle incoming messages
  const handleMessage = useCallback((event: MessageEvent) => {
    try {
      const message: WebSocketMessage = JSON.parse(event.data)

      switch (message.type) {
        case 'traffic':
          handleTrafficMessage(message.data)
          break
        case 'vulnerability':
          handleVulnerabilityMessage(message.data)
          break
        case 'intercepted':
          handleInterceptedMessage(message.data)
          break
        case 'ack':
          console.log('[WebSocket] Ack received:', message.data)
          break
        default:
          console.log('[WebSocket] Unknown message type:', message.type)
      }
    } catch (error) {
      console.error('[WebSocket] Failed to parse message:', error)
    }
  }, [])

  const handleTrafficMessage = useCallback((data: any) => {
    if (data && data.id) {
      const trafficItem: TrafficItem = {
        id: data.id,
        timestamp: data.timestamp || new Date().toISOString(),
        method: data.method || 'GET',
        url: data.url || '',
        host: data.host || '',
        path: data.path || '',
        statusCode: data.status_code || data.statusCode || 0,
        contentType: data.content_type || data.contentType || '',
        requestSize: data.request_size || data.requestSize || 0,
        responseSize: data.response_size || data.responseSize || 0,
        duration: data.duration || 0,
        requestHeaders: data.request_headers || data.requestHeaders || {},
        responseHeaders: data.response_headers || data.responseHeaders || {},
        requestBody: data.request_body || data.requestBody || '',
        responseBody: data.response_body || data.responseBody || '',
        clientIP: data.client_ip || data.clientIP || '',
        protocol: data.protocol || 'HTTP/1.1',
        intercepted: data.intercepted || false,
      }
      addTrafficItem(trafficItem)
    }
  }, [addTrafficItem])

  const handleVulnerabilityMessage = useCallback((data: any) => {
    if (data && data.id) {
      const vuln: Vulnerability = {
        id: data.id,
        title: data.title || data.name || 'Unknown Vulnerability',
        severity: data.severity || 'medium',
        type: data.type || 'unknown',
        url: data.url || '',
        method: data.method || 'GET',
        request: data.request || '',
        response: data.response || '',
        description: data.description || '',
        remediation: data.remediation || '',
        references: data.references || [],
        status: data.status || 'open',
        createdAt: data.created_at || data.createdAt || new Date().toISOString(),
        updatedAt: data.updated_at || data.updatedAt || new Date().toISOString(),
        source: data.source || 'passive',
        cwe: data.cwe,
        cvss: data.cvss,
      }
      addVulnerability(vuln)
    }
  }, [addVulnerability])

  const handleInterceptedMessage = useCallback((data: any) => {
    if (data && data.id) {
      addInterceptedRequest({
        id: data.id,
        timestamp: data.timestamp || new Date().toISOString(),
        method: data.method || 'GET',
        url: data.url || '',
        host: data.host || '',
        path: data.path || '',
        headers: data.headers || {},
        requestHeaders: data.request_headers || data.requestHeaders || {},
        body: data.body || '',
        requestBody: data.request_body || data.requestBody || '',
        contentType: data.content_type || data.contentType || '',
        clientIP: data.client_ip || data.clientIP || '',
        protocol: data.protocol || 'HTTP/1.1',
        requestSize: data.request_size || data.requestSize || 0,
      })
    }
  }, [addInterceptedRequest])

  // Connect to WebSocket
  const connect = useCallback(() => {
    if (wsRef.current?.readyState === WebSocket.OPEN) {
      return
    }

    try {
      const ws = new WebSocket(url)

      ws.onopen = () => {
        console.log('[WebSocket] Connected')
        setConnected(true)
        setReconnecting(false)
        onConnected?.()

        // Subscribe to channels
        ws.send(JSON.stringify({
          action: 'subscribe',
          channels: ['traffic', 'vulnerabilities', 'intercept']
        }))
      }

      ws.onclose = () => {
        console.log('[WebSocket] Disconnected')
        setConnected(false)
        onDisconnected?.()

        // Auto reconnect
        if (autoConnect) {
          setReconnecting(true)
          reconnectTimeoutRef.current = setTimeout(() => {
            connect()
          }, 3000)
        }
      }

      ws.onerror = (error) => {
        console.error('[WebSocket] Error:', error)
        onError?.(error)
      }

      ws.onmessage = handleMessage

      wsRef.current = ws
    } catch (error) {
      console.error('[WebSocket] Failed to connect:', error)
    }
  }, [url, autoConnect, handleMessage, onConnected, onDisconnected, onError])

  // Disconnect from WebSocket
  const disconnect = useCallback(() => {
    if (reconnectTimeoutRef.current) {
      clearTimeout(reconnectTimeoutRef.current)
    }
    if (wsRef.current) {
      wsRef.current.close()
      wsRef.current = null
    }
    setConnected(false)
    setReconnecting(false)
  }, [])

  // Send message
  const send = useCallback((data: any) => {
    if (wsRef.current?.readyState === WebSocket.OPEN) {
      wsRef.current.send(JSON.stringify(data))
      return true
    }
    return false
  }, [])

  // Subscribe to channel
  const subscribe = useCallback((channels: string[]) => {
    return send({ action: 'subscribe', channels })
  }, [send])

  // Unsubscribe from channel
  const unsubscribe = useCallback((channels: string[]) => {
    return send({ action: 'unsubscribe', channels })
  }, [send])

  // Intercept control
  const enableIntercept = useCallback((mode: 'all' | 'filter' = 'all') => {
    return send({ action: 'intercept_enable', data: { mode } })
  }, [send])

  const disableIntercept = useCallback(() => {
    return send({ action: 'intercept_disable' })
  }, [send])

  const forwardRequest = useCallback((requestId: string, modified?: any) => {
    return send({ action: 'forward', data: { request_id: requestId, ...modified } })
  }, [send])

  const dropRequest = useCallback((requestId: string) => {
    return send({ action: 'drop', data: { request_id: requestId } })
  }, [send])

  // Auto connect on mount
  useEffect(() => {
    if (autoConnect) {
      connect()
    }

    return () => {
      disconnect()
    }
  }, [autoConnect, connect, disconnect])

  return {
    connected,
    reconnecting,
    connect,
    disconnect,
    send,
    subscribe,
    unsubscribe,
    enableIntercept,
    disableIntercept,
    forwardRequest,
    dropRequest,
  }
}

export default useWebSocket
