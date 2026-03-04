import { create } from 'zustand'

// Payload position marker
export const PAYLOAD_MARKER_START = '§'
export const PAYLOAD_MARKER_END = '§'

// Payload position in request
export interface PayloadPosition {
  id: string
  startIndex: number
  endIndex: number
  originalValue: string
}

// Attack type
export type AttackType = 'sniper' | 'battering_ram' | 'pitchfork' | 'cluster_bomb'

// Payload configuration
export interface PayloadConfig {
  type: 'list' | 'numbers' | 'bruteforcer' | 'file'
  items: string[]
  // For numbers
  min?: number
  max?: number
  step?: number
  // For file
  filePath?: string
}

// Intruder tab
export interface IntruderTab {
  id: string
  name: string
  // Original request
  request: {
    method: string
    url: string
    headers: Record<string, string>
    body: string
  }
  // Positions marked for payload injection
  positions: PayloadPosition[]
  // Payload configurations for each position
  payloadConfigs: Record<string, PayloadConfig>
  // Attack type
  attackType: AttackType
  // Attack results
  results: AttackResult[]
  // Status
  status: 'idle' | 'running' | 'paused' | 'completed'
  // Progress
  progress: {
    current: number
    total: number
  }
}

// Attack result
export interface AttackResult {
  id: string
  positionValues: Record<string, string>
  request: string
  response: string
  statusCode: number
  responseTime: number
  responseLength: number
  timestamp: string
  error?: string
}

interface IntruderState {
  tabs: IntruderTab[]
  activeTabId: string | null

  // Tab management
  addTab: (tab?: Partial<IntruderTab>) => string
  removeTab: (id: string) => void
  setActiveTab: (id: string) => void
  duplicateTab: (id: string) => string

  // Request management
  setRequest: (id: string, request: IntruderTab['request']) => void

  // Position management
  addPosition: (tabId: string, position: Omit<PayloadPosition, 'id'>) => string
  removePosition: (tabId: string, positionId: string) => void
  clearPositions: (tabId: string) => void

  // Payload configuration
  setPayloadConfig: (tabId: string, positionId: string, config: PayloadConfig) => void

  // Attack configuration
  setAttackType: (tabId: string, attackType: AttackType) => void

  // Results management
  addResult: (tabId: string, result: AttackResult) => void
  clearResults: (tabId: string) => void

  // Attack control
  setStatus: (tabId: string, status: IntruderTab['status']) => void
  setProgress: (tabId: string, current: number, total: number) => void
}

const createDefaultTab = (index: number): IntruderTab => ({
  id: `intruder-${Date.now()}-${Math.random().toString(36).substr(2, 9)}`,
  name: `Attack ${index}`,
  request: {
    method: 'GET',
    url: '',
    headers: {},
    body: ''
  },
  positions: [],
  payloadConfigs: {},
  attackType: 'sniper',
  results: [],
  status: 'idle',
  progress: {
    current: 0,
    total: 0
  }
})

export const useIntruderStore = create<IntruderState>((set, get) => ({
  tabs: [createDefaultTab(1)],
  activeTabId: null,

  addTab: (tab) => {
    const { tabs } = get()
    const newTab = { ...createDefaultTab(tabs.length + 1), ...tab }
    set({ tabs: [...tabs, newTab] })
    return newTab.id
  },

  removeTab: (id) => {
    const { tabs, activeTabId } = get()
    const newTabs = tabs.filter(t => t.id !== id)
    if (newTabs.length === 0) {
      newTabs.push(createDefaultTab(1))
    }
    set({
      tabs: newTabs,
      activeTabId: activeTabId === id ? newTabs[0]?.id : activeTabId
    })
  },

  setActiveTab: (id) => set({ activeTabId: id }),

  duplicateTab: (id) => {
    const { tabs } = get()
    const tabToDuplicate = tabs.find(t => t.id === id)
    if (tabToDuplicate) {
      const newTab: IntruderTab = {
        ...tabToDuplicate,
        id: `intruder-${Date.now()}-${Math.random().toString(36).substr(2, 9)}`,
        name: `${tabToDuplicate.name} (copy)`,
        results: [],
        status: 'idle',
        progress: { current: 0, total: 0 }
      }
      set({ tabs: [...tabs, newTab] })
      return newTab.id
    }
    return ''
  },

  setRequest: (id, request) => set((state) => ({
    tabs: state.tabs.map(tab =>
      tab.id === id ? { ...tab, request } : tab
    )
  })),

  addPosition: (tabId, position) => {
    const id = `pos-${Date.now()}-${Math.random().toString(36).substr(2, 9)}`
    set((state) => ({
      tabs: state.tabs.map(tab =>
        tab.id === tabId
          ? { ...tab, positions: [...tab.positions, { ...position, id }] }
          : tab
      )
    }))
    return id
  },

  removePosition: (tabId, positionId) => set((state) => ({
    tabs: state.tabs.map(tab =>
      tab.id === tabId
        ? { ...tab, positions: tab.positions.filter(p => p.id !== positionId) }
        : tab
    )
  })),

  clearPositions: (tabId) => set((state) => ({
    tabs: state.tabs.map(tab =>
      tab.id === tabId ? { ...tab, positions: [] } : tab
    )
  })),

  setPayloadConfig: (tabId, positionId, config) => set((state) => ({
    tabs: state.tabs.map(tab =>
      tab.id === tabId
        ? { ...tab, payloadConfigs: { ...tab.payloadConfigs, [positionId]: config } }
        : tab
    )
  })),

  setAttackType: (tabId, attackType) => set((state) => ({
    tabs: state.tabs.map(tab =>
      tab.id === tabId ? { ...tab, attackType } : tab
    )
  })),

  addResult: (tabId, result) => set((state) => ({
    tabs: state.tabs.map(tab =>
      tab.id === tabId
        ? { ...tab, results: [...tab.results, result] }
        : tab
    )
  })),

  clearResults: (tabId) => set((state) => ({
    tabs: state.tabs.map(tab =>
      tab.id === tabId ? { ...tab, results: [], progress: { current: 0, total: 0 } } : tab
    )
  })),

  setStatus: (tabId, status) => set((state) => ({
    tabs: state.tabs.map(tab =>
      tab.id === tabId ? { ...tab, status } : tab
    )
  })),

  setProgress: (tabId, current, total) => set((state) => ({
    tabs: state.tabs.map(tab =>
      tab.id === tabId ? { ...tab, progress: { current, total } } : tab
    )
  }))
}))
