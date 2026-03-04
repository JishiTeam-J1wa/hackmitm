export interface TrafficItem {
  id: string
  timestamp: string
  method: string
  url: string
  host: string
  path: string
  statusCode: number
  contentType: string
  requestSize: number
  responseSize: number
  duration: number
  requestHeaders: Record<string, string>
  responseHeaders: Record<string, string>
  requestBody: string
  responseBody: string
  clientIP: string
  protocol: string
  intercepted: boolean
}

export interface FingerprintResult {
  url: string
  fingerprints: string[]
  confidence: number
  processTime: number
  title: string
  statusCode: number
  timestamp: string
}

export interface ProxyStatus {
  running: boolean
  port: number
  interceptMode: boolean
  activeConnections: number
  totalRequests: number
  uptime: number
}

export interface DashboardMetrics {
  qps: number
  avgResponseTime: number
  activeConnections: number
  totalRequests: number
  totalBytesIn: number
  totalBytesOut: number
  errorRate: number
  uptime: number
}

export interface TrafficStats {
  requestsPerSecond: number[]
  responseTimes: number[]
  statusCodes: Record<string, number>
  topHosts: { host: string; count: number }[]
  methods: Record<string, number>
}

export interface Target {
  id: string
  host: string
  port: number
  protocol: string
  title?: string
  technologies?: string[]
  inScope: boolean
  lastAccessed: string
  requestCount: number
}

export interface TargetTree {
  [host: string]: {
    paths: string[]
    technologies?: string[]
    title?: string
    requestCount: number
  }
}

export interface RepeaterRequest {
  id: string
  name: string
  method: string
  url: string
  headers: Record<string, string>
  body: string
  createdAt: string
  updatedAt: string
}

export interface RepeaterResponse {
  statusCode: number
  statusText: string
  headers: Record<string, string>
  body: string
  responseTime: number
  contentLength: number
}

export interface ScopeConfig {
  includePatterns: string[]
  excludePatterns: string[]
  enabled: boolean
}

export type TabId = 'proxy' | 'target' | 'repeater' | 'intruder' | 'fingerprint' | 'dashboard' | 'websocket' | 'vuln' | 'scan'

// WebSocket 相关类型
export interface WebSocketMessage {
  id: string
  timestamp: string
  direction: 'incoming' | 'outgoing'
  type: 'text' | 'binary' | 'ping' | 'pong' | 'close'
  url: string
  host: string
  size: number
  content: string
  contentType: string
  connectionId: string
}

export interface WebSocketConnection {
  id: string
  url: string
  host: string
  protocol: string
  openedAt: string
  closedAt?: string
  messageCount: number
  status: 'open' | 'closed'
}

// 漏洞相关类型
export interface Vulnerability {
  id: string
  title: string
  severity: 'critical' | 'high' | 'medium' | 'low'
  type: string
  url: string
  method: string
  request: string
  response: string
  description: string
  remediation: string
  references: string[]
  status: 'open' | 'fixed' | 'ignored'
  createdAt: string
  updatedAt: string
  source: 'passive' | 'active'
  cwe?: string
  cvss?: number
}

// 扫描结果类型
export interface ScanResult {
  id: string
  pluginName: string
  pluginId: string
  severity: 'critical' | 'high' | 'medium' | 'low' | 'info'
  title: string
  description: string
  url: string
  method: string
  evidence: string
  request: string
  response: string
  timestamp: string
  falsePositive: boolean
  tags?: string[]
}

// 扫描插件类型
export interface ScanPlugin {
  id: string
  name: string
  description: string
  author: string
  version: string
  enabled: boolean
  category: string
  severity: string
  config: Record<string, any>
}

// 被动扫描配置
export interface PassiveScanConfig {
  enabled: boolean
  includePatterns: string[]
  excludePatterns: string[]
  maxRequestsPerSecond: number
  timeout: number
}

// 扫描统计
export interface ScanStats {
  totalScanned: number
  totalFindings: number
  criticalCount: number
  highCount: number
  mediumCount: number
  lowCount: number
  infoCount: number
}

export interface AppSettings {
  proxyHost: string
  proxyPort: number
  apiEndpoint: string
  theme: 'light' | 'dark' | 'system'
  interceptMode: boolean
}
