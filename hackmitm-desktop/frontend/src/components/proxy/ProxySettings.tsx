import { useState, useEffect } from 'react'
import {
  Settings,
  X,
  Save,
  RotateCcw,
  Globe,
  Shield,
  ShieldCheck,
  Server,
  Link2,
  Download,
  AlertCircle,
  CheckCircle,
  Wifi,
  RefreshCw,
  Loader2,
} from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Switch } from '@/components/ui/switch'
import { useProxyStore } from '@/store'
import {
  GetBindAddressOptions,
  GetProxyConfig,
  SaveProxyConfig,
  TestConnection,
} from '../../../wailsjs/go/main/App'

interface ProxySettingsProps {
  open: boolean
  onClose: () => void
  certInfo?: { installed: boolean; path: string }
  onInstallCert?: () => void
  interceptMode?: boolean
  onToggleIntercept?: (enabled: boolean) => void
  apiEndpoint?: string
  onApiEndpointChange?: (endpoint: string) => void
}

type SettingsTab = 'proxy' | 'intercept' | 'ssl' | 'upstream'

interface BindOption {
  value: string
  label: string
}

export function ProxySettings({
  open,
  onClose,
  onInstallCert,
  interceptMode = false,
  onToggleIntercept,
  apiEndpoint = 'http://localhost:9090',
  onApiEndpointChange,
}: ProxySettingsProps) {
  const { status, setStatus } = useProxyStore()
  const [activeTab, setActiveTab] = useState<SettingsTab>('proxy')
  const [loading, setLoading] = useState(false)
  const [bindOptions, setBindOptions] = useState<BindOption[]>([
    { value: '127.0.0.1', label: '127.0.0.1 (仅本机)' },
    { value: '0.0.0.0', label: '0.0.0.0 (所有接口)' },
  ])
  const [customBindAddress, setCustomBindAddress] = useState('')
  const [showCustomInput, setShowCustomInput] = useState(false)

  // Local state for settings
  const [proxyPort, setProxyPort] = useState(status.port || 8080)
  const [bindAddress, setBindAddress] = useState('127.0.0.1')
  const [localApiEndpoint, setLocalApiEndpoint] = useState(apiEndpoint)
  const [localInterceptMode, setLocalInterceptMode] = useState(interceptMode)
  const [enableHttps, setEnableHttps] = useState(true)
  const [enableHttp2, setEnableHttp2] = useState(true)
  const [enableWebsockets, setEnableWebsockets] = useState(true)
  const [upstreamProxyEnabled, setUpstreamProxyEnabled] = useState(false)
  const [upstreamProxy, setUpstreamProxy] = useState('')
  const [upstreamUsername, setUpstreamUsername] = useState('')
  const [upstreamPassword, setUpstreamPassword] = useState('')
  const [certDir, setCertDir] = useState('./certs')
  const [errors, setErrors] = useState<Record<string, string>>({})

  // Certificate status
  const [certStatus, setCertStatus] = useState<{
    installed: boolean
    checking: boolean
    path: string
    expiresAt: string
  }>({
    installed: false,
    checking: true,
    path: '',
    expiresAt: ''
  })

  // Test connection state
  const [testingConnection, setTestingConnection] = useState(false)
  const [testResult, setTestResult] = useState<{ success: boolean; message: string; latency?: number } | null>(null)

  // Load bind address options on mount
  useEffect(() => {
    loadBindOptions()
    loadConfig()
    checkCertStatus()
  }, [])

  const loadBindOptions = async () => {
    try {
      const options = await GetBindAddressOptions()
      if (options && options.length > 0) {
        setBindOptions(options.map((opt: any) => ({
          value: opt.value,
          label: opt.label,
        })))
      }
    } catch (e) {
      console.error('Failed to load bind options:', e)
    }
  }

  const handleTestConnection = async () => {
    setTestingConnection(true)
    setTestResult(null)

    try {
      const result = await TestConnection(localApiEndpoint)
      setTestResult({
        success: result.success,
        message: result.message,
        latency: result.latency,
      })
    } catch (e) {
      setTestResult({
        success: false,
        message: e instanceof Error ? e.message : String(e),
      })
    } finally {
      setTestingConnection(false)
    }
  }

  const loadConfig = async () => {
    try {
      const config = await GetProxyConfig()
      if (config) {
        setProxyPort(config.proxyPort || 8080)
        setBindAddress(config.bindAddress || '127.0.0.1')
        setEnableHttps(config.enableHttps ?? true)
        setEnableHttp2(config.enableHttp2 ?? true)
        setEnableWebsockets(config.enableWebSocket ?? true)
        setUpstreamProxyEnabled(config.upstreamEnabled ?? false)
        setUpstreamProxy(config.upstreamAddress || '')
        setUpstreamUsername(config.upstreamUsername || '')
        setUpstreamPassword(config.upstreamPassword || '')
        setCertDir(config.certDir || './certs')
      }
    } catch (e) {
      console.error('Failed to load proxy config:', e)
    }
  }

  const handleSave = async () => {
    setErrors({})
    setLoading(true)

    const config = {
      proxyPort,
      bindAddress: showCustomInput ? customBindAddress : bindAddress,
      apiPort: parseInt(localApiEndpoint.split(':').pop() || '9090'),
      enableHttps,
      enableHttp2,
      enableWebSocket: enableWebsockets,
      upstreamEnabled: upstreamProxyEnabled,
      upstreamAddress: upstreamProxy,
      upstreamUsername,
      upstreamPassword,
      interceptMode: localInterceptMode,
      certDir,
    }

    try {
      const validationErrors = await SaveProxyConfig(config)
      if (validationErrors && Object.keys(validationErrors).length > 0) {
        setErrors(validationErrors)
        setLoading(false)
        return
      }

      // Update local state
      setStatus({ ...status, port: proxyPort })
      if (onToggleIntercept) onToggleIntercept(localInterceptMode)
      if (onApiEndpointChange) onApiEndpointChange(localApiEndpoint)

      onClose()
    } catch (e) {
      setErrors({ general: '保存配置失败' })
    }

    setLoading(false)
  }

  const handleReset = async () => {
    setProxyPort(8080)
    setBindAddress('127.0.0.1')
    setLocalApiEndpoint('http://localhost:9090')
    setLocalInterceptMode(false)
    setEnableHttps(true)
    setEnableHttp2(true)
    setEnableWebsockets(true)
    setUpstreamProxyEnabled(false)
    setUpstreamProxy('')
    setUpstreamUsername('')
    setUpstreamPassword('')
    setCertDir('./certs')
    setShowCustomInput(false)
    setCustomBindAddress('')
    setErrors({})
  }

  // Check certificate status
  const checkCertStatus = async () => {
    setCertStatus(prev => ({ ...prev, checking: true }))
    try {
      // Try to access the certificate status endpoint
      const response = await fetch(`${localApiEndpoint}/cert/status`)
      if (response.ok) {
        const data = await response.json()
        setCertStatus({
          installed: data.installed || false,
          checking: false,
          path: data.path || '',
          expiresAt: data.expires_at || ''
        })
      } else {
        // If endpoint not available, check if we can reach hackmitm.ca
        setCertStatus({
          installed: false,
          checking: false,
          path: '',
          expiresAt: ''
        })
      }
    } catch (e) {
      setCertStatus({
        installed: false,
        checking: false,
        path: '',
        expiresAt: ''
      })
    }
  }

  // Open certificate download page
  const openCertDownloadPage = () => {
    // Open the special certificate download URL (like Burp's http://burp.ca)
    window.open('http://hackmitm.ca/', '_blank')
  }

  // Open certificate file location
  const openCertLocation = () => {
    if (certStatus.path) {
      // In a real implementation, this would open the file location
      console.log('Open cert location:', certStatus.path)
    }
  }

  if (!open) return null

  const tabs: { id: SettingsTab; label: string; icon: React.ReactNode }[] = [
    { id: 'proxy', label: '代理设置', icon: <Server className="w-4 h-4" /> },
    { id: 'intercept', label: '拦截模式', icon: <AlertCircle className="w-4 h-4" /> },
    { id: 'ssl', label: 'SSL证书', icon: <Shield className="w-4 h-4" /> },
    { id: 'upstream', label: '上游代理', icon: <Link2 className="w-4 h-4" /> },
  ]

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50">
      <div className="bg-white rounded-xl shadow-2xl w-full max-w-3xl overflow-hidden">
        {/* Header */}
        <div className="flex items-center justify-between px-5 py-4 border-b bg-gray-50">
          <div className="flex items-center gap-2">
            <Settings className="w-5 h-5 text-orange-500" />
            <h2 className="text-lg font-semibold text-gray-900">代理设置</h2>
          </div>
          <Button variant="ghost" size="sm" onClick={onClose} className="h-8 w-8 p-0">
            <X className="w-4 h-4" />
          </Button>
        </div>

        <div className="flex" style={{ height: '480px' }}>
          {/* Left Navigation */}
          <div className="w-44 border-r bg-gray-50 p-3 flex-shrink-0">
            <nav className="space-y-1">
              {tabs.map((tab) => (
                <button
                  key={tab.id}
                  onClick={() => setActiveTab(tab.id)}
                  className={`w-full flex items-center gap-2.5 px-3 py-2.5 rounded-lg text-sm font-medium transition-colors ${
                    activeTab === tab.id
                      ? 'bg-orange-500 text-white shadow-sm'
                      : 'text-gray-600 hover:bg-gray-100'
                  }`}
                >
                  {tab.icon}
                  {tab.label}
                </button>
              ))}
            </nav>
          </div>

          {/* Right Content */}
          <div className="flex-1 overflow-y-auto p-5">
            {/* Error Messages */}
            {Object.keys(errors).length > 0 && (
              <div className="mb-4 p-3 bg-red-50 border border-red-200 rounded-lg">
                {Object.entries(errors).map(([key, msg]) => (
                  <p key={key} className="text-sm text-red-600">{msg}</p>
                ))}
              </div>
            )}

            {/* Proxy Settings */}
            {activeTab === 'proxy' && (
              <div className="space-y-6">
                <div>
                  <h3 className="text-base font-semibold text-gray-900 mb-4">代理监听器</h3>
                  <div className="grid grid-cols-2 gap-4">
                    <div>
                      <Label htmlFor="proxyPort" className="text-sm font-medium">代理端口</Label>
                      <Input
                        id="proxyPort"
                        type="number"
                        value={proxyPort}
                        onChange={(e) => setProxyPort(parseInt(e.target.value) || 4443)}
                        className="mt-1.5"
                      />
                      <p className="text-xs text-gray-500 mt-1.5">浏览器代理设置使用此端口</p>
                    </div>
                    <div>
                      <Label htmlFor="bindAddress" className="text-sm font-medium">绑定地址</Label>
                      <div className="relative mt-1.5">
                        <select
                          value={showCustomInput ? 'custom' : bindAddress}
                          onChange={(e) => {
                            if (e.target.value === 'custom') {
                              setShowCustomInput(true)
                            } else {
                              setShowCustomInput(false)
                              setBindAddress(e.target.value)
                            }
                          }}
                          className="w-full h-9 px-3 text-sm border border-gray-200 rounded-md bg-white focus:outline-none focus:ring-2 focus:ring-gray-950"
                        >
                          {bindOptions.map((opt) => (
                            <option key={opt.value} value={opt.value}>{opt.label}</option>
                          ))}
                          <option value="custom">自定义 IP...</option>
                        </select>
                      </div>
                      {showCustomInput && (
                        <Input
                          value={customBindAddress}
                          onChange={(e) => setCustomBindAddress(e.target.value)}
                          placeholder="输入 IP 地址"
                          className="mt-2"
                        />
                      )}
                      <p className="text-xs text-gray-500 mt-1.5">
                        {showCustomInput ? '输入网卡 IP 地址' : '选择网络接口'}
                      </p>
                    </div>
                  </div>
                </div>

                <div className="border-t pt-6">
                  <h3 className="text-base font-semibold text-gray-900 mb-4">API 连接</h3>
                  <div>
                    <Label htmlFor="apiEndpoint" className="text-sm font-medium">监控 API 地址</Label>
                    <Input
                      id="apiEndpoint"
                      value={localApiEndpoint}
                      onChange={(e) => setLocalApiEndpoint(e.target.value)}
                      placeholder="http://localhost:9090"
                      className="mt-1.5"
                    />
                    <p className="text-xs text-gray-500 mt-1.5">用于连接 HackMITM 服务的 API 端点</p>
                  </div>
                </div>

                <div className="border-t pt-6">
                  <h3 className="text-base font-semibold text-gray-900 mb-4">协议支持</h3>
                  <div className="space-y-3">
                    <div className="flex items-center justify-between p-3 bg-gray-50 rounded-lg">
                      <div>
                        <p className="text-sm font-medium text-gray-700">HTTPS 支持</p>
                        <p className="text-xs text-gray-500">拦截和解密 HTTPS 流量</p>
                      </div>
                      <Switch checked={enableHttps} onCheckedChange={setEnableHttps} />
                    </div>
                    <div className="flex items-center justify-between p-3 bg-gray-50 rounded-lg">
                      <div>
                        <p className="text-sm font-medium text-gray-700">HTTP/2 支持</p>
                        <p className="text-xs text-gray-500">支持 HTTP/2 协议</p>
                      </div>
                      <Switch checked={enableHttp2} onCheckedChange={setEnableHttp2} />
                    </div>
                    <div className="flex items-center justify-between p-3 bg-gray-50 rounded-lg">
                      <div>
                        <p className="text-sm font-medium text-gray-700">WebSocket 支持</p>
                        <p className="text-xs text-gray-500">拦截 WebSocket 连接</p>
                      </div>
                      <Switch checked={enableWebsockets} onCheckedChange={setEnableWebsockets} />
                    </div>
                  </div>
                </div>
              </div>
            )}

            {/* Intercept Settings */}
            {activeTab === 'intercept' && (
              <div className="space-y-6">
                <div>
                  <h3 className="text-base font-semibold text-gray-900 mb-4">拦截模式</h3>
                  <div className={`p-4 rounded-lg border-2 transition-colors ${
                    localInterceptMode ? 'border-orange-500 bg-orange-50' : 'border-gray-200 bg-gray-50'
                  }`}>
                    <div className="flex items-start justify-between">
                      <div className="flex-1">
                        <div className="flex items-center gap-2 mb-2">
                          <AlertCircle className={`w-5 h-5 ${localInterceptMode ? 'text-orange-500' : 'text-gray-400'}`} />
                          <span className="font-medium text-gray-900">启用拦截模式</span>
                        </div>
                        <p className="text-sm text-gray-600 mb-3">
                          启用后，每个请求都会被暂停，您可以查看、修改或丢弃请求后再转发
                        </p>
                        <Switch
                          checked={localInterceptMode}
                          onCheckedChange={setLocalInterceptMode}
                        />
                      </div>
                    </div>
                  </div>
                </div>

                {localInterceptMode && (
                  <div className="bg-orange-50 border border-orange-200 rounded-lg p-4">
                    <div className="flex items-start gap-3">
                      <AlertCircle className="w-5 h-5 text-orange-500 mt-0.5" />
                      <div>
                        <p className="font-medium text-orange-800">拦截模式已启用</p>
                        <p className="text-sm text-orange-700 mt-1">
                          所有通过代理的请求都将被暂停。您可以在拦截列表中对请求进行操作：
                        </p>
                        <ul className="text-sm text-orange-700 mt-2 space-y-1 ml-4 list-disc">
                          <li><strong>Forward</strong> - 转发请求到目标服务器</li>
                          <li><strong>Drop</strong> - 丢弃请求</li>
                          <li><strong>Modify</strong> - 修改请求后再转发</li>
                        </ul>
                      </div>
                    </div>
                  </div>
                )}

                <div className="border-t pt-6">
                  <h3 className="text-base font-semibold text-gray-900 mb-4">使用说明</h3>
                  <div className="bg-gray-50 rounded-lg p-4 text-sm text-gray-600 space-y-2">
                    <p>拦截模式允许您在请求发送到目标服务器之前查看和修改请求。</p>
                    <p>这在以下场景非常有用：</p>
                    <ul className="ml-4 list-disc space-y-1">
                      <li>测试 Web 应用的输入验证</li>
                      <li>修改请求参数进行安全测试</li>
                      <li>分析 API 请求结构</li>
                    </ul>
                  </div>
                </div>
              </div>
            )}

            {/* SSL Settings */}
            {activeTab === 'ssl' && (
              <div className="space-y-6">
                {/* Certificate Status */}
                <div>
                  <div className="flex items-center justify-between mb-4">
                    <h3 className="text-base font-semibold text-gray-900">CA 证书状态</h3>
                    <Button
                      variant="outline"
                      size="sm"
                      onClick={checkCertStatus}
                      disabled={certStatus.checking}
                    >
                      {certStatus.checking ? (
                        <Loader2 className="w-3.5 h-3.5 mr-1.5 animate-spin" />
                      ) : (
                        <RefreshCw className="w-3.5 h-3.5 mr-1.5" />
                      )}
                      检测证书
                    </Button>
                  </div>

                  <div className={`p-4 rounded-lg border ${
                    certStatus.checking
                      ? 'bg-gray-50 border-gray-200'
                      : certStatus.installed
                        ? 'bg-green-50 border-green-200'
                        : 'bg-orange-50 border-orange-200'
                  }`}>
                    <div className="flex items-center gap-4">
                      {certStatus.checking ? (
                        <div className="w-12 h-12 rounded-full bg-gray-200 flex items-center justify-center">
                          <Loader2 className="w-6 h-6 text-gray-500 animate-spin" />
                        </div>
                      ) : certStatus.installed ? (
                        <div className="w-12 h-12 rounded-full bg-green-100 flex items-center justify-center">
                          <ShieldCheck className="w-6 h-6 text-green-600" />
                        </div>
                      ) : (
                        <div className="w-12 h-12 rounded-full bg-orange-100 flex items-center justify-center">
                          <Shield className="w-6 h-6 text-orange-600" />
                        </div>
                      )}
                      <div className="flex-1">
                        <p className={`font-medium ${
                          certStatus.checking
                            ? 'text-gray-600'
                            : certStatus.installed
                              ? 'text-green-800'
                              : 'text-orange-800'
                        }`}>
                          {certStatus.checking
                            ? '正在检测证书...'
                            : certStatus.installed
                              ? '证书已安装'
                              : '证书未安装或未被系统信任'
                          }
                        </p>
                        {certStatus.path && (
                          <p className="text-sm text-gray-500 mt-0.5 font-mono text-xs">{certStatus.path}</p>
                        )}
                        {certStatus.expiresAt && (
                          <p className="text-xs text-gray-500 mt-1">过期时间: {certStatus.expiresAt}</p>
                        )}
                      </div>
                      <div className="flex items-center gap-1">
                        {!certStatus.checking && (
                          certStatus.installed
                            ? <CheckCircle className="w-5 h-5 text-green-500" />
                            : <AlertCircle className="w-5 h-5 text-orange-500" />
                        )}
                      </div>
                    </div>
                  </div>
                </div>

                {/* Certificate Download (like Burp's http://burp.ca) */}
                <div className="border-t pt-6">
                  <h3 className="text-base font-semibold text-gray-900 mb-4">下载证书</h3>
                  <div className="bg-blue-50 border border-blue-200 rounded-lg p-4">
                    <div className="flex items-start gap-3">
                      <Globe className="w-5 h-5 text-blue-500 mt-0.5" />
                      <div className="flex-1">
                        <p className="text-sm font-medium text-blue-800">
                          通过浏览器下载证书
                        </p>
                        <p className="text-xs text-blue-600 mt-1">
                          配置浏览器使用代理后，访问以下地址下载 CA 证书：
                        </p>
                        <div className="mt-2 p-2 bg-white rounded border border-blue-200">
                          <code className="text-sm text-blue-700">http://hackmitm.ca/</code>
                        </div>
                        <Button
                          variant="outline"
                          size="sm"
                          className="mt-3"
                          onClick={openCertDownloadPage}
                        >
                          <Globe className="w-3.5 h-3.5 mr-1.5" />
                          打开证书下载页面
                        </Button>
                      </div>
                    </div>
                  </div>
                </div>

                {/* Certificate Directory */}
                <div>
                  <Label htmlFor="certDir" className="text-sm font-medium">证书存储目录</Label>
                  <div className="flex gap-2 mt-1.5">
                    <Input
                      id="certDir"
                      value={certDir}
                      onChange={(e) => setCertDir(e.target.value)}
                      className="flex-1"
                    />
                    <Button variant="outline" size="sm">
                      浏览
                    </Button>
                  </div>
                </div>

                {/* Certificate Actions */}
                <div>
                  <h3 className="text-base font-semibold text-gray-900 mb-4">证书操作</h3>
                  <div className="grid grid-cols-2 gap-3">
                    <Button
                      variant="outline"
                      className="h-auto py-4 flex-col"
                      onClick={onInstallCert}
                    >
                      <Download className="w-5 h-5 mb-2" />
                      <span>安装证书</span>
                      <span className="text-xs text-gray-500 mt-1">导入到系统钥匙串</span>
                    </Button>
                    <Button variant="outline" className="h-auto py-4 flex-col" onClick={openCertLocation}>
                      <Globe className="w-5 h-5 mb-2" />
                      <span>打开证书位置</span>
                      <span className="text-xs text-gray-500 mt-1">在 Finder 中显示</span>
                    </Button>
                  </div>
                </div>

                {/* Installation Instructions */}
                <div className="border-t pt-6">
                  <h3 className="text-base font-semibold text-gray-900 mb-4">安装说明</h3>
                  <div className="bg-gray-50 rounded-lg p-4 text-sm text-gray-600 space-y-3">
                    <p className="font-medium text-gray-800">方法一：通过浏览器下载（推荐）</p>
                    <ol className="ml-4 list-decimal space-y-1">
                      <li>配置浏览器代理为 <code className="px-1 bg-gray-200 rounded">127.0.0.1:8080</code></li>
                      <li>访问 <code className="px-1 bg-gray-200 rounded">http://hackmitm.ca/</code></li>
                      <li>下载并双击证书文件</li>
                      <li>在钥匙串中将证书设为"始终信任"</li>
                    </ol>

                    <p className="font-medium text-gray-800 mt-4">方法二：直接安装</p>
                    <ol className="ml-4 list-decimal space-y-1">
                      <li>点击"安装证书"按钮</li>
                      <li>在钥匙串访问中找到 HackMITM CA 证书</li>
                      <li>双击证书，将"信任"设置为"始终信任"</li>
                      <li>重启浏览器使证书生效</li>
                    </ol>
                  </div>
                </div>
              </div>
            )}

            {/* Upstream Proxy Settings */}
            {activeTab === 'upstream' && (
              <div className="space-y-6">
                <div>
                  <h3 className="text-base font-semibold text-gray-900 mb-4">上游代理</h3>
                  <div className={`p-4 rounded-lg border-2 transition-colors ${
                    upstreamProxyEnabled ? 'border-orange-500 bg-orange-50' : 'border-gray-200 bg-gray-50'
                  }`}>
                    <div className="flex items-start justify-between mb-4">
                      <div>
                        <div className="flex items-center gap-2 mb-2">
                          <Link2 className={`w-5 h-5 ${upstreamProxyEnabled ? 'text-orange-500' : 'text-gray-400'}`} />
                          <span className="font-medium text-gray-900">启用上游代理</span>
                        </div>
                        <p className="text-sm text-gray-600">
                          通过另一个代理服务器转发所有流量
                        </p>
                      </div>
                      <Switch
                        checked={upstreamProxyEnabled}
                        onCheckedChange={setUpstreamProxyEnabled}
                      />
                    </div>
                  </div>
                </div>

                {upstreamProxyEnabled && (
                  <div className="space-y-4 border-t pt-6">
                    <div>
                      <Label htmlFor="upstreamProxy" className="text-sm font-medium">代理地址</Label>
                      <Input
                        id="upstreamProxy"
                        value={upstreamProxy}
                        onChange={(e) => setUpstreamProxy(e.target.value)}
                        placeholder="http://127.0.0.1:8080 或 socks5://127.0.0.1:1080"
                        className="mt-1.5"
                      />
                      <p className="text-xs text-gray-500 mt-1.5">
                        支持 HTTP 和 SOCKS5 代理
                      </p>
                    </div>

                    <div className="grid grid-cols-2 gap-4">
                      <div>
                        <Label htmlFor="upstreamUsername" className="text-sm font-medium">用户名（可选）</Label>
                        <Input
                          id="upstreamUsername"
                          value={upstreamUsername}
                          onChange={(e) => setUpstreamUsername(e.target.value)}
                          placeholder="用户名"
                          className="mt-1.5"
                        />
                      </div>
                      <div>
                        <Label htmlFor="upstreamPassword" className="text-sm font-medium">密码（可选）</Label>
                        <Input
                          id="upstreamPassword"
                          type="password"
                          value={upstreamPassword}
                          onChange={(e) => setUpstreamPassword(e.target.value)}
                          placeholder="密码"
                          className="mt-1.5"
                        />
                      </div>
                    </div>

                    <Button
                      variant="outline"
                      size="sm"
                      onClick={handleTestConnection}
                      disabled={testingConnection}
                      className="mt-2"
                    >
                      {testingConnection ? (
                        <Loader2 className="w-3.5 h-3.5 mr-1.5 animate-spin" />
                      ) : (
                        <RefreshCw className="w-3.5 h-3.5 mr-1.5" />
                      )}
                      {testingConnection ? '测试中...' : '测试连接'}
                    </Button>

                    {testResult && (
                      <div className={`mt-2 p-2 rounded text-xs ${
                        testResult.success
                          ? 'bg-green-50 text-green-700 border border-green-200'
                          : 'bg-red-50 text-red-700 border border-red-200'
                      }`}>
                        <div className="flex items-center gap-1.5">
                          {testResult.success ? (
                            <CheckCircle className="w-3.5 h-3.5" />
                          ) : (
                            <AlertCircle className="w-3.5 h-3.5" />
                          )}
                          <span>{testResult.message}</span>
                        </div>
                        {testResult.latency !== undefined && testResult.latency > 0 && (
                          <div className="mt-1 text-[10px] opacity-75">
                            延迟: {testResult.latency}ms
                          </div>
                        )}
                      </div>
                    )}
                  </div>
                )}

                {!upstreamProxyEnabled && (
                  <div className="border-t pt-6">
                    <div className="bg-gray-50 rounded-lg p-4 text-sm text-gray-600">
                      <Wifi className="w-5 h-5 text-gray-400 mb-2" />
                      <p>
                        启用上游代理后，所有流量将通过指定的代理服务器转发。
                        这在需要通过公司代理或 VPN 访问外部网络时非常有用。
                      </p>
                    </div>
                  </div>
                )}
              </div>
            )}
          </div>
        </div>

        {/* Footer */}
        <div className="flex items-center justify-between px-5 py-3 border-t bg-gray-50">
          <Button variant="ghost" onClick={handleReset} className="text-gray-600 text-sm">
            <RotateCcw className="w-3.5 h-3.5 mr-1.5" />
            重置默认
          </Button>
          <div className="flex gap-2">
            <Button variant="outline" onClick={onClose} className="text-sm">
              取消
            </Button>
            <Button
              onClick={handleSave}
              disabled={loading}
              className="bg-orange-500 hover:bg-orange-600 text-white text-sm"
            >
              {loading ? (
                <RefreshCw className="w-3.5 h-3.5 mr-1.5 animate-spin" />
              ) : (
                <Save className="w-3.5 h-3.5 mr-1.5" />
              )}
              保存设置
            </Button>
          </div>
        </div>
      </div>
    </div>
  )
}
