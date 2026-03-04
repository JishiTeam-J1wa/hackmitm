import { create } from 'zustand'

// Active Scan Types
export type ScanStatus = 'idle' | 'running' | 'paused' | 'completed' | 'error' | 'cancelled'
export type Severity = 'critical' | 'high' | 'medium' | 'low' | 'info'

export interface ScanTarget {
  id: string
  url: string
  method: string
  headers: Record<string, string>
  body: string
  enabled: boolean
}

export interface ScanPlugin {
  id: string
  name: string
  description: string
  severity: Severity
  enabled: boolean
}

export interface ScanFinding {
  id: string
  pluginId: string
  pluginName: string
  severity: Severity
  title: string
  description: string
  url: string
  method: string
  payload: string
  evidence: string
  request: string
  response: string
  confidence: number
  timestamp: string
}

export interface ScanProgress {
  totalTargets: number
  scannedTargets: number
  totalRequests: number
  completedReqs: number
  findingsCount: number
  errorCount: number
  status: ScanStatus
  currentTarget: string
  currentPlugin: string
  startTime: string
  elapsedTime: number
  estimatedTime: number
  requestsPerSec: number
}

export interface ScanConfig {
  id: string
  name: string
  concurrency: number
  rateLimit: number
  timeout: number
  followRedirects: boolean
  enabledPlugins: string[]
}

interface ActiveScanState {
  // Scan configuration
  scans: Record<string, ScanConfig>
  activeScanId: string | null

  // Targets
  targets: ScanTarget[]

  // Plugins
  plugins: ScanPlugin[]

  // Progress
  progress: Record<string, ScanProgress>

  // Findings
  findings: Record<string, ScanFinding[]>

  // UI state
  isLoading: boolean
  selectedFinding: ScanFinding | null

  // Actions - Scan management
  createScan: (config: ScanConfig) => Promise<string>
  startScan: (scanId: string) => Promise<void>
  pauseScan: (scanId: string) => Promise<void>
  resumeScan: (scanId: string) => Promise<void>
  stopScan: (scanId: string) => Promise<void>
  removeScan: (scanId: string) => Promise<void>
  setActiveScan: (scanId: string | null) => void

  // Actions - Target management
  addTarget: (target: Omit<ScanTarget, 'id' | 'enabled'>) => Promise<void>
  removeTarget: (targetId: string) => Promise<void>
  updateTarget: (targetId: string, updates: Partial<ScanTarget>) => void
  clearTargets: () => void

  // Actions - Plugin management
  loadDefaultPlugins: () => Promise<void>
  enablePlugin: (pluginId: string) => Promise<void>
  disablePlugin: (pluginId: string) => Promise<void>
  togglePlugin: (pluginId: string) => void

  // Actions - Data loading
  loadProgress: (scanId: string) => Promise<void>
  loadFindings: (scanId: string) => Promise<void>

  // Actions - UI
  selectFinding: (finding: ScanFinding | null) => void
  setLoading: (loading: boolean) => void
}

// Default plugins
const defaultPlugins: ScanPlugin[] = [
  {
    id: 'SQL-INJECTION',
    name: 'SQL Injection',
    description: 'Detects SQL injection vulnerabilities by injecting various SQL payloads',
    severity: 'high',
    enabled: true,
  },
  {
    id: 'XSS',
    name: 'Cross-Site Scripting',
    description: 'Detects reflected and stored XSS vulnerabilities',
    severity: 'high',
    enabled: true,
  },
  {
    id: 'PATH-TRAVERSAL',
    name: 'Path Traversal',
    description: 'Detects path traversal vulnerabilities that allow reading files outside web root',
    severity: 'high',
    enabled: true,
  },
  {
    id: 'COMMAND-INJECTION',
    name: 'Command Injection',
    description: 'Detects OS command injection vulnerabilities',
    severity: 'critical',
    enabled: true,
  },
]

// Default progress
const defaultProgress: ScanProgress = {
  totalTargets: 0,
  scannedTargets: 0,
  totalRequests: 0,
  completedReqs: 0,
  findingsCount: 0,
  errorCount: 0,
  status: 'idle',
  currentTarget: '',
  currentPlugin: '',
  startTime: '',
  elapsedTime: 0,
  estimatedTime: 0,
  requestsPerSec: 0,
}

export const useActiveScanStore = create<ActiveScanState>((set, get) => ({
  scans: {},
  activeScanId: null,
  targets: [],
  plugins: defaultPlugins,
  progress: {},
  findings: {},
  isLoading: false,
  selectedFinding: null,

  // Scan management
  createScan: async (config) => {
    const { CreateActiveScan } = await import('../../wailsjs/go/main/App')
    const scanId = config.id || `scan-${Date.now()}`
    await CreateActiveScan(
      scanId,
      config.name,
      config.concurrency,
      config.rateLimit,
      config.timeout,
      config.followRedirects,
      config.enabledPlugins
    )
    set((state) => ({
      scans: { ...state.scans, [scanId]: { ...config, id: scanId } },
      progress: { ...state.progress, [scanId]: { ...defaultProgress } },
      findings: { ...state.findings, [scanId]: [] },
    }))
    return scanId
  },

  startScan: async (scanId) => {
    const { StartActiveScan, AddActiveScanTarget } = await import('../../wailsjs/go/main/App')
    const { targets } = get()

    // Add all targets to the scan
    for (const target of targets) {
      await AddActiveScanTarget(
        scanId,
        target.id,
        target.url,
        target.method,
        target.headers,
        target.body
      )
    }

    await StartActiveScan(scanId)
    set((state) => ({
      progress: {
        ...state.progress,
        [scanId]: { ...state.progress[scanId], status: 'running' },
      },
    }))
  },

  pauseScan: async (scanId) => {
    const { PauseActiveScan } = await import('../../wailsjs/go/main/App')
    await PauseActiveScan(scanId)
    set((state) => ({
      progress: {
        ...state.progress,
        [scanId]: { ...state.progress[scanId], status: 'paused' },
      },
    }))
  },

  resumeScan: async (scanId) => {
    const { ResumeActiveScan } = await import('../../wailsjs/go/main/App')
    await ResumeActiveScan(scanId)
    set((state) => ({
      progress: {
        ...state.progress,
        [scanId]: { ...state.progress[scanId], status: 'running' },
      },
    }))
  },

  stopScan: async (scanId) => {
    const { StopActiveScan } = await import('../../wailsjs/go/main/App')
    await StopActiveScan(scanId)
    set((state) => ({
      progress: {
        ...state.progress,
        [scanId]: { ...state.progress[scanId], status: 'cancelled' },
      },
    }))
  },

  removeScan: async (scanId) => {
    const { RemoveActiveScan } = await import('../../wailsjs/go/main/App')
    await RemoveActiveScan(scanId)
    set((state) => {
      const { [scanId]: _, ...remainingScans } = state.scans
      const { [scanId]: __, ...remainingProgress } = state.progress
      const { [scanId]: ___, ...remainingFindings } = state.findings
      return {
        scans: remainingScans,
        progress: remainingProgress,
        findings: remainingFindings,
        activeScanId: state.activeScanId === scanId ? null : state.activeScanId,
      }
    })
  },

  setActiveScan: (scanId) => set({ activeScanId: scanId }),

  // Target management
  addTarget: async (target) => {
    const targetId = `target-${Date.now()}`
    const newTarget: ScanTarget = {
      ...target,
      id: targetId,
      enabled: true,
    }
    set((state) => ({ targets: [...state.targets, newTarget] }))
  },

  removeTarget: async (targetId) => {
    const scanId = get().activeScanId
    if (scanId) {
      const { RemoveActiveScanTarget } = await import('../../wailsjs/go/main/App')
      await RemoveActiveScanTarget(scanId, targetId)
    }
    set((state) => ({
      targets: state.targets.filter((t) => t.id !== targetId),
    }))
  },

  updateTarget: (targetId, updates) => {
    set((state) => ({
      targets: state.targets.map((t) =>
        t.id === targetId ? { ...t, ...updates } : t
      ),
    }))
  },

  clearTargets: () => set({ targets: [] }),

  // Plugin management
  loadDefaultPlugins: async () => {
    try {
      const { GetDefaultActiveScanPlugins } = await import('../../wailsjs/go/main/App')
      const plugins = await GetDefaultActiveScanPlugins()
      if (plugins && plugins.length > 0) {
        set({ plugins: plugins as ScanPlugin[] })
      }
    } catch (error) {
      console.error('Failed to load plugins:', error)
    }
  },

  enablePlugin: async (pluginId) => {
    const scanId = get().activeScanId
    if (scanId) {
      const { EnableActiveScanPlugin } = await import('../../wailsjs/go/main/App')
      await EnableActiveScanPlugin(scanId, pluginId)
    }
    set((state) => ({
      plugins: state.plugins.map((p) =>
        p.id === pluginId ? { ...p, enabled: true } : p
      ),
    }))
  },

  disablePlugin: async (pluginId) => {
    const scanId = get().activeScanId
    if (scanId) {
      const { DisableActiveScanPlugin } = await import('../../wailsjs/go/main/App')
      await DisableActiveScanPlugin(scanId, pluginId)
    }
    set((state) => ({
      plugins: state.plugins.map((p) =>
        p.id === pluginId ? { ...p, enabled: false } : p
      ),
    }))
  },

  togglePlugin: (pluginId) => {
    const plugin = get().plugins.find((p) => p.id === pluginId)
    if (plugin) {
      if (plugin.enabled) {
        get().disablePlugin(pluginId)
      } else {
        get().enablePlugin(pluginId)
      }
    }
  },

  // Data loading
  loadProgress: async (scanId) => {
    try {
      const { GetActiveScanProgress } = await import('../../wailsjs/go/main/App')
      const progress = await GetActiveScanProgress(scanId)
      set((state) => ({
        progress: { ...state.progress, [scanId]: progress as ScanProgress },
      }))
    } catch (error) {
      console.error('Failed to load progress:', error)
    }
  },

  loadFindings: async (scanId) => {
    try {
      const { GetActiveScanFindings } = await import('../../wailsjs/go/main/App')
      const findings = await GetActiveScanFindings(scanId)
      set((state) => ({
        findings: { ...state.findings, [scanId]: (findings || []) as ScanFinding[] },
      }))
    } catch (error) {
      console.error('Failed to load findings:', error)
    }
  },

  // UI
  selectFinding: (finding) => set({ selectedFinding: finding }),
  setLoading: (loading) => set({ isLoading: loading }),
}))
