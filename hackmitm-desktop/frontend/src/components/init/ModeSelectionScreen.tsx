import { useState } from 'react'
import { motion, AnimatePresence } from 'framer-motion'
import { useAppStore, ConnectionMode } from '@/store/appStore'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Card, CardContent } from '@/components/ui/card'
import {
  Cloud,
  Monitor,
  FolderOpen,
  CheckCircle2,
  Loader2,
  AlertCircle,
  Server,
  Shield,
  Database,
  ChevronRight,
  ChevronLeft,
  Plus,
} from 'lucide-react'

interface ModeSelectionScreenProps {
  onComplete: () => void
}

type Step = 'mode' | 'database'

export function ModeSelectionScreen({ onComplete }: ModeSelectionScreenProps) {
  const {
    appConfig,
    setConnectionMode,
    setLocalConfig,
    setRemoteConfig,
    completeInitialization,
  } = useAppStore()

  const [step, setStep] = useState<Step>('mode')
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)

  // Mode selection
  const [selectedMode, setSelectedMode] = useState<ConnectionMode>(appConfig.connectionMode)

  // Database config
  const [dataDir, setDataDir] = useState(appConfig.localConfig?.dataDir || '')
  const [dbName, setDbName] = useState('hackmitm')
  const [isNewDb, setIsNewDb] = useState(true)

  // Local mode - only API port, proxy port is configured in Proxy Settings
  const [apiPort] = useState(appConfig.localConfig?.apiPort || 9090)

  // Remote mode config
  const [remoteHost] = useState(appConfig.remoteConfig?.host || '')
  const [remotePort] = useState(appConfig.remoteConfig?.port || 9090)
  const [remoteApiKey] = useState(appConfig.remoteConfig?.apiKey || '')

  const handleModeSelect = (mode: ConnectionMode) => {
    setSelectedMode(mode)
    setError(null)
  }

  const handleSelectFolder = async () => {
    try {
      const { SelectDatabaseFolder } = await import('../../../wailsjs/go/main/App')
      const folder = await SelectDatabaseFolder()
      if (folder) {
        setDataDir(folder)
      }
    } catch (err) {
      console.error('Failed to select folder:', err)
      setDataDir('./data')
    }
  }

  const handleNext = () => {
    setError(null)
    if (step === 'mode') {
      if (!selectedMode) {
        setError('请选择启动模式')
        return
      }
      setConnectionMode(selectedMode)
      setStep('database')
    } else if (step === 'database') {
      if (!dataDir) {
        setError('请选择数据目录')
        return
      }
      handleStart()
    }
  }

  const handleBack = () => {
    setError(null)
    if (step === 'database') setStep('mode')
  }

  const handleStart = async () => {
    setError(null)
    setLoading(true)

    try {
      if (selectedMode === 'local') {
        setLocalConfig({
          dataDir,
          apiPort,
          proxyPort: 8080, // Default, can be changed in Proxy Settings
        })

        const { StartLocalMode } = await import('../../../wailsjs/go/main/App')
        await StartLocalMode({
          dataDir,
          apiPort,
          proxyPort: 8080, // Default proxy port
        })
      } else {
        if (!remoteHost) {
          setError('请输入服务器地址')
          setLoading(false)
          return
        }
        setRemoteConfig({
          host: remoteHost,
          port: remotePort,
          apiKey: remoteApiKey,
        })

        const { ConnectRemoteMode, SetAPIEndpoint } = await import('../../../wailsjs/go/main/App')
        const endpoint = `http://${remoteHost}:${remotePort}`
        await SetAPIEndpoint(endpoint)
        await ConnectRemoteMode({
          host: remoteHost,
          port: remotePort,
          apiKey: remoteApiKey,
        })
      }

      completeInitialization()
      onComplete()
    } catch (err) {
      console.error('Failed to start:', err)
      setError(err instanceof Error ? err.message : String(err))
    } finally {
      setLoading(false)
    }
  }

  // Step indicator
  const renderStepIndicator = () => {
    const steps = [
      { key: 'mode', label: '选择模式' },
      { key: 'database', label: '选择项目' },
    ]
    const currentIndex = steps.findIndex(s => s.key === step)

    return (
      <div className="flex items-center justify-center mb-4">
        {steps.map((s, index) => (
          <div key={s.key} className="flex items-center">
            <div
              className={`w-7 h-7 rounded-full flex items-center justify-center text-xs font-medium transition-colors ${
                index < currentIndex
                  ? 'bg-orange-500 text-white'
                  : index === currentIndex
                  ? 'bg-orange-500 text-white'
                  : 'bg-gray-100 text-gray-400'
              }`}
            >
              {index < currentIndex ? <CheckCircle2 className="w-4 h-4" /> : index + 1}
            </div>
            {index < steps.length - 1 && (
              <div
                className={`w-12 h-0.5 mx-1.5 transition-colors ${
                  index < currentIndex ? 'bg-orange-500' : 'bg-gray-200'
                }`}
              />
            )}
          </div>
        ))}
      </div>
    )
  }

  // Step 1: Mode selection
  const renderModeStep = () => (
    <div className="space-y-4">
      <div className="grid grid-cols-2 gap-4">
        <Card
          className={`cursor-pointer transition-all hover:shadow-md ${
            selectedMode === 'local'
              ? 'ring-2 ring-orange-500 bg-orange-50'
              : 'hover:bg-gray-50'
          }`}
          onClick={() => handleModeSelect('local')}
        >
          <CardContent className="p-5">
            <div className="flex flex-col items-center text-center">
              <div
                className={`w-14 h-14 rounded-full flex items-center justify-center mb-3 transition-colors ${
                  selectedMode === 'local' ? 'bg-orange-500' : 'bg-gray-100'
                }`}
              >
                <Monitor
                  className={`w-7 h-7 transition-colors ${
                    selectedMode === 'local' ? 'text-white' : 'text-gray-500'
                  }`}
                />
              </div>
              <h3 className="text-sm font-semibold text-gray-900 mb-1">本地模式</h3>
              <p className="text-xs text-gray-500">在本机启动内嵌服务</p>
              <div className="flex items-center gap-1 text-[10px] text-gray-400 mt-2">
                <Shield className="w-3 h-3" />
                <span>管理本地数据库</span>
              </div>
            </div>
          </CardContent>
        </Card>

        <Card
          className={`cursor-pointer transition-all hover:shadow-md ${
            selectedMode === 'remote'
              ? 'ring-2 ring-orange-500 bg-orange-50'
              : 'hover:bg-gray-50'
          }`}
          onClick={() => handleModeSelect('remote')}
        >
          <CardContent className="p-5">
            <div className="flex flex-col items-center text-center">
              <div
                className={`w-14 h-14 rounded-full flex items-center justify-center mb-3 transition-colors ${
                  selectedMode === 'remote' ? 'bg-orange-500' : 'bg-gray-100'
                }`}
              >
                <Cloud
                  className={`w-7 h-7 transition-colors ${
                    selectedMode === 'remote' ? 'text-white' : 'text-gray-500'
                  }`}
                />
              </div>
              <h3 className="text-sm font-semibold text-gray-900 mb-1">远程模式</h3>
              <p className="text-xs text-gray-500">连接到远程服务器</p>
              <div className="flex items-center gap-1 text-[10px] text-gray-400 mt-2">
                <Server className="w-3 h-3" />
                <span>连接远程服务</span>
              </div>
            </div>
          </CardContent>
        </Card>
      </div>
    </div>
  )

  // Step 2: Project configuration
  const renderDatabaseStep = () => (
    <motion.div
      initial={{ opacity: 0, x: 20 }}
      animate={{ opacity: 1, x: 0 }}
      exit={{ opacity: 0, x: -20 }}
      className="space-y-4"
    >
      <div className="bg-gray-50 rounded-xl p-4 space-y-4">
        <h3 className="text-sm font-medium text-gray-900 flex items-center gap-2">
          <Database className="w-4 h-4 text-orange-500" />
          项目管理
        </h3>

        {/* 新建/使用现有 */}
        <div className="grid grid-cols-2 gap-3">
          <Card
            className={`cursor-pointer transition-all ${
              isNewDb ? 'ring-2 ring-orange-500 bg-orange-50' : 'hover:bg-gray-50'
            }`}
            onClick={() => setIsNewDb(true)}
          >
            <CardContent className="p-3 flex items-center gap-3">
              <div className={`w-8 h-8 rounded-lg flex items-center justify-center ${isNewDb ? 'bg-orange-500' : 'bg-gray-100'}`}>
                <Plus className={`w-4 h-4 ${isNewDb ? 'text-white' : 'text-gray-500'}`} />
              </div>
              <div>
                <h4 className="text-xs font-medium text-gray-900">新建项目</h4>
                <p className="text-[10px] text-gray-500">创建新项目数据库</p>
              </div>
            </CardContent>
          </Card>

          <Card
            className={`cursor-pointer transition-all ${
              !isNewDb ? 'ring-2 ring-orange-500 bg-orange-50' : 'hover:bg-gray-50'
            }`}
            onClick={() => setIsNewDb(false)}
          >
            <CardContent className="p-3 flex items-center gap-3">
              <div className={`w-8 h-8 rounded-lg flex items-center justify-center ${!isNewDb ? 'bg-orange-500' : 'bg-gray-100'}`}>
                <FolderOpen className={`w-4 h-4 ${!isNewDb ? 'text-white' : 'text-gray-500'}`} />
              </div>
              <div>
                <h4 className="text-xs font-medium text-gray-900">打开项目</h4>
                <p className="text-[10px] text-gray-500">选择已有项目</p>
              </div>
            </CardContent>
          </Card>
        </div>

        {/* 项目路径 */}
        <div>
          <Label htmlFor="dataDir" className="text-xs font-medium">项目路径</Label>
          <div className="flex gap-2 mt-1">
            <Input
              id="dataDir"
              value={dataDir}
              onChange={(e) => setDataDir(e.target.value)}
              placeholder="选择项目存储目录"
              className="flex-1 h-8 text-xs"
            />
            <Button variant="outline" onClick={handleSelectFolder} type="button" className="h-8 text-xs px-3">
              <FolderOpen className="w-3.5 h-3.5 mr-1" />
              浏览
            </Button>
          </div>
        </div>

        {/* 项目名称 */}
        <div>
          <Label htmlFor="dbName" className="text-xs font-medium">项目名称</Label>
          <Input
            id="dbName"
            value={dbName}
            onChange={(e) => setDbName(e.target.value)}
            placeholder="hackmitm"
            className="mt-1 h-8 text-xs"
          />
        </div>
      </div>

      <Card className="bg-blue-50 border-blue-200">
        <CardContent className="p-3">
          <h4 className="font-medium text-gray-900 mb-2 text-xs">💡 提示</h4>
          <div className="text-xs text-gray-600 space-y-1">
            <p>• 代理端口和证书可在启动后的 <strong>Proxy Settings</strong> 中配置</p>
            <p>• 默认代理端口: <code className="px-1 bg-blue-100 rounded">127.0.0.1:8080</code></p>
            <p>• 默认 API 端口: <code className="px-1 bg-blue-100 rounded">9090</code></p>
          </div>
        </CardContent>
      </Card>
    </motion.div>
  )

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-gradient-to-br from-gray-50 to-gray-100 p-4">
      <motion.div
        className="w-full max-w-xl bg-white rounded-2xl shadow-2xl flex flex-col max-h-[85vh]"
        initial={{ opacity: 0, y: 20 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ duration: 0.3 }}
      >
        {/* Header */}
        <div className="flex-shrink-0 p-4 pb-0">
          <div className="flex justify-center mb-2">
            <img src="/logo.png" alt="HackMITM" className="w-10 h-10 object-contain" />
          </div>
          <div className="text-center mb-2">
            <h1 className="text-base font-bold text-gray-900">HackMITM</h1>
            <p className="text-gray-500 text-xs">启动后可在设置中配置代理和证书</p>
          </div>
          {renderStepIndicator()}
        </div>

        {/* Content */}
        <div className="flex-1 overflow-y-auto px-4">
          {error && (
            <motion.div
              className="mb-3 p-2 bg-red-50 border border-red-200 rounded-lg flex items-center gap-2 text-red-700"
              initial={{ opacity: 0, y: -10 }}
              animate={{ opacity: 1, y: 0 }}
            >
              <AlertCircle className="w-4 h-4 flex-shrink-0" />
              <span className="text-xs">{error}</span>
            </motion.div>
          )}

          <AnimatePresence mode="wait">
            {step === 'mode' && renderModeStep()}
            {step === 'database' && renderDatabaseStep()}
          </AnimatePresence>
        </div>

        {/* Footer */}
        <div className="flex-shrink-0 flex items-center justify-between p-4 border-t">
          <Button
            variant="ghost"
            onClick={handleBack}
            disabled={step === 'mode' || loading}
            className="h-8 text-xs text-gray-600"
          >
            <ChevronLeft className="w-4 h-4 mr-1" />
            上一步
          </Button>

          {step === 'database' ? (
            <Button
              onClick={handleNext}
              disabled={loading}
              className="bg-orange-500 hover:bg-orange-600 text-white h-8 text-xs min-w-[100px]"
            >
              {loading ? (
                <>
                  <Loader2 className="w-3.5 h-3.5 mr-1.5 animate-spin" />
                  启动中...
                </>
              ) : (
                <>
                  <CheckCircle2 className="w-3.5 h-3.5 mr-1.5" />
                  启动
                </>
              )}
            </Button>
          ) : (
            <Button
              onClick={handleNext}
              className="bg-orange-500 hover:bg-orange-600 text-white h-8 text-xs min-w-[100px]"
            >
              下一步
              <ChevronRight className="w-4 h-4 ml-1" />
            </Button>
          )}
        </div>
      </motion.div>
    </div>
  )
}
