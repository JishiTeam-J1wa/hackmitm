import { create } from 'zustand'
import type { Vulnerability } from '@/types'
import {
  GetVulnerabilities,
  AddVulnerability,
  UpdateVulnerabilityStatus,
  DeleteVulnerability,
} from '../../wailsjs/go/main/App'
import type { api } from '../../wailsjs/go/models'

interface VulnState {
  vulnerabilities: Vulnerability[]
  selectedVuln: Vulnerability | null
  isLoading: boolean
  filters: {
    severity: string
    type: string
    status: string
    search: string
  }
  viewMode: 'cards' | 'list'

  // Actions - Data loading
  loadVulnerabilities: () => Promise<void>

  // Actions
  addVulnerability: (vuln: Vulnerability) => Promise<void>
  addVulnerabilities: (vulns: Vulnerability[]) => void
  updateVulnerability: (id: string, updates: Partial<Vulnerability>) => Promise<void>
  deleteVulnerability: (id: string) => Promise<void>
  selectVuln: (vuln: Vulnerability | null) => void
  setFilters: (filters: Partial<VulnState['filters']>) => void
  setViewMode: (mode: 'cards' | 'list') => void
  clearAll: () => Promise<void>

  // Statistics
  getStats: () => {
    critical: number
    high: number
    medium: number
    low: number
    total: number
    open: number
    fixed: number
    ignored: number
  }
}

// Convert from backend Vulnerability to frontend Vulnerability
function convertVuln(v: api.Vulnerability): Vulnerability {
  return {
    id: String(v.id),
    title: v.title,
    severity: v.severity as any,
    type: v.type,
    url: v.url,
    method: v.method || '',
    request: v.request || '',
    response: v.response || '',
    description: v.description || '',
    remediation: v.remediation || '',
    references: v.references || [],
    status: v.status as any,
    createdAt: v.createdAt,
    updatedAt: v.updatedAt,
    source: (v.source || '') as any,
    cwe: v.cwe || '',
    cvss: v.cvss || 0,
  }
}

// Convert from frontend Vulnerability to backend format
function toBackendVuln(v: Vulnerability): Omit<api.Vulnerability, 'id' | 'createdAt' | 'updatedAt'> {
  return {
    title: v.title,
    severity: v.severity,
    type: v.type,
    url: v.url,
    method: v.method,
    request: v.request,
    response: v.response,
    description: v.description,
    remediation: v.remediation,
    references: v.references,
    status: v.status,
    source: v.source,
    cwe: v.cwe || '',
    cvss: v.cvss || 0,
  }
}

export const useVulnStore = create<VulnState>((set, get) => ({
  vulnerabilities: [],
  selectedVuln: null,
  isLoading: false,
  filters: {
    severity: 'all',
    type: 'all',
    status: 'all',
    search: '',
  },
  viewMode: 'cards',

  loadVulnerabilities: async () => {
    set({ isLoading: true })
    try {
      const { filters } = get()
      const results = await GetVulnerabilities(
        filters.severity === 'all' ? '' : filters.severity,
        filters.status === 'all' ? '' : filters.status,
        filters.type === 'all' ? '' : filters.type,
        1000
      )
      const vulns = results.map(convertVuln)
      set({ vulnerabilities: vulns })
    } catch (error) {
      console.error('Failed to load vulnerabilities:', error)
    } finally {
      set({ isLoading: false })
    }
  },

  addVulnerability: async (vuln) => {
    try {
      const backendVuln = toBackendVuln(vuln)
      await AddVulnerability(backendVuln as any)
      set((state) => ({
        vulnerabilities: [vuln, ...state.vulnerabilities]
      }))
    } catch (error) {
      console.error('Failed to add vulnerability:', error)
    }
  },

  addVulnerabilities: (vulns) => set((state) => ({
    vulnerabilities: [...vulns, ...state.vulnerabilities]
  })),

  updateVulnerability: async (id, updates) => {
    try {
      if (updates.status) {
        await UpdateVulnerabilityStatus(parseInt(id), updates.status)
      }
      set((state) => ({
        vulnerabilities: state.vulnerabilities.map((v) =>
          v.id === id ? { ...v, ...updates, updatedAt: new Date().toISOString() } : v
        ),
        selectedVuln: state.selectedVuln?.id === id
          ? { ...state.selectedVuln, ...updates, updatedAt: new Date().toISOString() }
          : state.selectedVuln,
      }))
    } catch (error) {
      console.error('Failed to update vulnerability:', error)
    }
  },

  deleteVulnerability: async (id) => {
    try {
      await DeleteVulnerability(parseInt(id))
      set((state) => ({
        vulnerabilities: state.vulnerabilities.filter((v) => v.id !== id),
        selectedVuln: state.selectedVuln?.id === id ? null : state.selectedVuln,
      }))
    } catch (error) {
      console.error('Failed to delete vulnerability:', error)
    }
  },

  selectVuln: (vuln) => set({ selectedVuln: vuln }),

  setFilters: (filters) => set((state) => ({
    filters: { ...state.filters, ...filters }
  })),

  setViewMode: (mode) => set({ viewMode: mode }),

  clearAll: async () => {
    const { vulnerabilities } = get()
    for (const vuln of vulnerabilities) {
      try {
        await DeleteVulnerability(parseInt(vuln.id))
      } catch (e) {
        // Ignore errors
      }
    }
    set({
      vulnerabilities: [],
      selectedVuln: null
    })
  },

  getStats: () => {
    const vulns = get().vulnerabilities
    return {
      critical: vulns.filter((v) => v.severity === 'critical').length,
      high: vulns.filter((v) => v.severity === 'high').length,
      medium: vulns.filter((v) => v.severity === 'medium').length,
      low: vulns.filter((v) => v.severity === 'low').length,
      total: vulns.length,
      open: vulns.filter((v) => v.status === 'open').length,
      fixed: vulns.filter((v) => v.status === 'fixed').length,
      ignored: vulns.filter((v) => v.status === 'ignored').length,
    }
  },
}))
