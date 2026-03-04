import { create } from 'zustand'
import type { ScanResult, ScanPlugin, PassiveScanConfig, ScanStats } from '@/types'
import { GetTrafficPatterns } from '../../wailsjs/go/main/App'
import {
  GetScanResults,
  MarkScanResultFalsePositive,
  DeleteScanResult,
  ClearScanResults,
} from '../../wailsjs/go/main/App'
import type { api } from '../../wailsjs/go/models'

interface TrafficPattern {
  type: string
  count: number
  percentage?: number
  confidence?: number
}

interface ScanState {
  // Scan status
  isScanning: boolean
  isLoading: boolean
  config: PassiveScanConfig
  stats: ScanStats
  trafficPatterns: TrafficPattern[]

  // Results - from database
  results: ScanResult[]
  selectedResult: ScanResult | null

  // Plugins - mapped from backend security plugin
  plugins: ScanPlugin[]

  // Filters
  filters: {
    severity: string
    pluginId: string
    falsePositive: string
    search: string
  }

  // Actions - Data loading
  loadTrafficPatterns: () => Promise<void>
  loadScanResults: () => Promise<void>
  loadStats: () => Promise<void>

  // Actions - Config
  setEnabled: (enabled: boolean) => void
  updateConfig: (config: Partial<PassiveScanConfig>) => void

  // Actions - Results
  selectResult: (result: ScanResult | null) => void
  markFalsePositive: (id: string, isFalsePositive: boolean) => Promise<void>
  deleteResult: (id: string) => Promise<void>
  clearResults: () => Promise<void>

  // Actions - Plugins
  togglePlugin: (id: string) => void
  updatePluginConfig: (id: string, config: Record<string, any>) => void

  // Actions - Filters
  setFilters: (filters: Partial<ScanState['filters']>) => void

  // Helpers
  getFilteredResults: () => ScanResult[]
}

// Default plugins - matching backend SecurityPlugin capabilities
const defaultPlugins: ScanPlugin[] = [
  {
    id: 'sql_injection_check',
    name: 'SQL Injection Detection',
    description: 'Detects potential SQL injection vulnerabilities in parameters',
    author: 'HackMITM',
    version: '1.0.0',
    enabled: true,
    category: 'Injection',
    severity: 'high',
    config: { checkBody: true, checkHeaders: false },
  },
  {
    id: 'xss_check',
    name: 'XSS Detection',
    description: 'Detects reflected and stored cross-site scripting vulnerabilities',
    author: 'HackMITM',
    version: '1.0.0',
    enabled: true,
    category: 'Injection',
    severity: 'high',
    config: { checkDOM: true, checkReflected: true },
  },
  {
    id: 'path_traversal_check',
    name: 'Path Traversal Detection',
    description: 'Detects directory traversal vulnerabilities',
    author: 'HackMITM',
    version: '1.0.0',
    enabled: true,
    category: 'Injection',
    severity: 'high',
    config: { checkEncoded: true },
  },
  {
    id: 'command_injection_check',
    name: 'Command Injection Detection',
    description: 'Detects command injection vulnerabilities',
    author: 'HackMITM',
    version: '1.0.0',
    enabled: true,
    category: 'Injection',
    severity: 'high',
    config: {},
  },
  {
    id: 'sensitive_file_check',
    name: 'Sensitive File Detection',
    description: 'Detects access to sensitive files like .env, .git, backup files',
    author: 'HackMITM',
    version: '1.0.0',
    enabled: true,
    category: 'Information Disclosure',
    severity: 'medium',
    config: { checkGit: true, checkEnv: true, checkBackups: true },
  },
  {
    id: 'rate_limit',
    name: 'Rate Limiting',
    description: 'Blocks requests exceeding rate limits',
    author: 'HackMITM',
    version: '1.0.0',
    enabled: true,
    category: 'Security Control',
    severity: 'medium',
    config: { maxRequests: 100, timeWindow: 60 },
  },
  {
    id: 'ip_blacklist',
    name: 'IP Blacklist',
    description: 'Block requests from specific IPs',
    author: 'HackMITM',
    version: '1.0.0',
    enabled: true,
    category: 'Security Control',
    severity: 'medium',
    config: { ips: [] },
  },
  {
    id: 'path_blacklist',
    name: 'Path Blacklist',
    description: 'Block requests to specific paths',
    author: 'HackMITM',
    version: '1.0.0',
    enabled: true,
    category: 'Security Control',
    severity: 'medium',
    config: { paths: [] },
  },
]

const defaultConfig: PassiveScanConfig = {
  enabled: true,
  includePatterns: ['*'],
  excludePatterns: ['*.js', '*.css', '*.png', '*.jpg'],
  maxRequestsPerSecond: 100,
  timeout: 30,
}

const defaultStats: ScanStats = {
  totalScanned: 0,
  totalFindings: 0,
  criticalCount: 0,
  highCount: 0,
  mediumCount: 0,
  lowCount: 0,
  infoCount: 0,
}

const defaultFilters = {
  severity: 'all',
  pluginId: 'all',
  falsePositive: 'all',
  search: '',
}

// Convert from backend ScanResult to frontend ScanResult
function convertScanResult(r: api.ScanResult): ScanResult {
  return {
    id: String(r.id),
    pluginName: r.pluginName,
    pluginId: r.pluginId,
    severity: r.severity as any,
    title: r.title,
    description: r.description || '',
    url: r.url,
    method: r.method || '',
    evidence: r.evidence || '',
    request: r.request || '',
    response: r.response || '',
    timestamp: r.timestamp,
    falsePositive: r.falsePositive,
    tags: r.tags || [],
  }
}

export const useScanStore = create<ScanState>((set, get) => ({
  isScanning: true,
  isLoading: false,
  config: defaultConfig,
  stats: defaultStats,
  trafficPatterns: [],
  results: [],
  selectedResult: null,
  plugins: defaultPlugins,
  filters: defaultFilters,

  // Data loading actions
  loadTrafficPatterns: async () => {
    set({ isLoading: true })
    try {
      const patterns = await GetTrafficPatterns()
      if (patterns) {
        const patternArray: TrafficPattern[] = []

        if (patterns.pattern_stats) {
          for (const [type, count] of Object.entries(patterns.pattern_stats as Record<string, number>)) {
            patternArray.push({ type, count })
          }
        } else {
          for (const [type, data] of Object.entries(patterns)) {
            if (typeof data === 'object' && data !== null) {
              const d = data as any
              patternArray.push({
                type,
                count: d.count || 0,
                percentage: d.percentage,
                confidence: d.confidence,
              })
            } else if (typeof data === 'number') {
              patternArray.push({ type, count: data })
            }
          }
        }

        set({ trafficPatterns: patternArray })

        const attackPattern = patternArray.find(p => p.type === 'attack')
        const totalScanned = patternArray.reduce((sum, p) => sum + p.count, 0)

        set((state) => ({
          stats: {
            ...state.stats,
            totalScanned,
            highCount: attackPattern?.count || 0,
            totalFindings: (attackPattern?.count || 0) + state.stats.mediumCount + state.stats.lowCount,
          }
        }))
      }
    } catch (error) {
      console.error('Failed to load traffic patterns:', error)
    } finally {
      set({ isLoading: false })
    }
  },

  loadScanResults: async () => {
    set({ isLoading: true })
    try {
      const { filters } = get()
      const results = await GetScanResults(
        filters.severity === 'all' ? '' : filters.severity,
        filters.pluginId === 'all' ? '' : filters.pluginId,
        filters.falsePositive === 'all' ? '' : filters.falsePositive,
        500
      )

      const convertedResults = results.map(convertScanResult)

      const highCount = convertedResults.filter(r => r.severity === 'high').length
      const mediumCount = convertedResults.filter(r => r.severity === 'medium').length
      const lowCount = convertedResults.filter(r => r.severity === 'low').length

      set({
        results: convertedResults,
        stats: {
          ...get().stats,
          totalFindings: convertedResults.length,
          highCount,
          mediumCount,
          lowCount,
        }
      })
    } catch (error) {
      console.error('Failed to load scan results:', error)
    } finally {
      set({ isLoading: false })
    }
  },

  loadStats: async () => {
    const { loadTrafficPatterns, loadScanResults } = get()
    await Promise.all([loadTrafficPatterns(), loadScanResults()])
  },

  // Config actions
  setEnabled: (enabled) => set((state) => ({
    isScanning: enabled,
    config: { ...state.config, enabled }
  })),

  updateConfig: (config) => set((state) => ({
    config: { ...state.config, ...config }
  })),

  // Result actions
  selectResult: (result) => set({ selectedResult: result }),

  markFalsePositive: async (id, isFalsePositive) => {
    try {
      await MarkScanResultFalsePositive(parseInt(id), isFalsePositive)
      set((state) => ({
        results: state.results.map((r) =>
          r.id === id ? { ...r, falsePositive: isFalsePositive } : r
        ),
        selectedResult: state.selectedResult?.id === id
          ? { ...state.selectedResult, falsePositive: isFalsePositive }
          : state.selectedResult,
      }))
    } catch (error) {
      console.error('Failed to mark false positive:', error)
    }
  },

  deleteResult: async (id) => {
    try {
      await DeleteScanResult(parseInt(id))
      set((state) => ({
        results: state.results.filter((r) => r.id !== id),
        selectedResult: state.selectedResult?.id === id ? null : state.selectedResult,
      }))
    } catch (error) {
      console.error('Failed to delete scan result:', error)
    }
  },

  clearResults: async () => {
    try {
      await ClearScanResults()
      set({
        results: [],
        selectedResult: null,
        stats: { ...get().stats, totalFindings: 0, criticalCount: 0, highCount: 0, mediumCount: 0, lowCount: 0, infoCount: 0 }
      })
    } catch (error) {
      console.error('Failed to clear scan results:', error)
    }
  },

  // Plugin actions
  togglePlugin: (id) => set((state) => ({
    plugins: state.plugins.map((p) =>
      p.id === id ? { ...p, enabled: !p.enabled } : p
    )
  })),

  updatePluginConfig: (id, config) => set((state) => ({
    plugins: state.plugins.map((p) =>
      p.id === id ? { ...p, config: { ...p.config, ...config } } : p
    )
  })),

  // Filter actions
  setFilters: (filters) => set((state) => ({
    filters: { ...state.filters, ...filters }
  })),

  // Helper to get filtered results
  getFilteredResults: () => {
    const { results, filters } = get()
    let filtered = [...results]

    if (filters.severity !== 'all') {
      filtered = filtered.filter((r) => r.severity === filters.severity)
    }

    if (filters.pluginId !== 'all') {
      filtered = filtered.filter((r) => r.pluginId === filters.pluginId)
    }

    if (filters.falsePositive === 'true') {
      filtered = filtered.filter((r) => r.falsePositive)
    } else if (filters.falsePositive === 'false') {
      filtered = filtered.filter((r) => !r.falsePositive)
    }

    if (filters.search) {
      const search = filters.search.toLowerCase()
      filtered = filtered.filter((r) =>
        r.title.toLowerCase().includes(search) ||
        r.url.toLowerCase().includes(search) ||
        r.pluginName.toLowerCase().includes(search)
      )
    }

    return filtered
  },
}))
