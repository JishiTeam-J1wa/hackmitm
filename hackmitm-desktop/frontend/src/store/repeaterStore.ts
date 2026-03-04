import { create } from 'zustand'
import { RepeaterRequest, RepeaterResponse } from '@/types'

interface RepeaterTab {
  id: string
  name: string
  request: RepeaterRequest
  response: RepeaterResponse | null
  loading: boolean
  hasNewContent: boolean  // Indicates tab has new content from Proxy (for highlighting)
}

interface RepeaterState {
  tabs: RepeaterTab[]
  activeTabId: string | null

  // Actions
  addTab: (tab?: Partial<RepeaterTab>) => string
  addRequest: (request: RepeaterRequest) => string
  removeTab: (id: string) => void
  setActiveTab: (id: string) => void
  updateRequest: (id: string, request: Partial<RepeaterRequest>) => void
  setResponse: (id: string, response: RepeaterResponse) => void
  setLoading: (id: string, loading: boolean) => void
  duplicateTab: (id: string) => string
}

const createDefaultTab = (index: number): RepeaterTab => ({
  id: `tab-${Date.now()}-${Math.random().toString(36).substr(2, 9)}`,
  name: `Request ${index}`,
  request: {
    id: '',
    name: '',
    method: 'GET',
    url: '',
    headers: {},
    body: '',
    createdAt: new Date().toISOString(),
    updatedAt: new Date().toISOString()
  },
  response: null,
  loading: false,
  hasNewContent: false
})

export const useRepeaterStore = create<RepeaterState>((set, get) => ({
  tabs: [createDefaultTab(1)],
  activeTabId: null,

  addTab: (tab) => {
    const { tabs } = get()
    const newTab = { ...createDefaultTab(tabs.length + 1), ...tab }
    set({ tabs: [...tabs, newTab] })
    return newTab.id
  },

  addRequest: (request) => {
    const { tabs } = get()
    const newTab: RepeaterTab = {
      id: `tab-${Date.now()}-${Math.random().toString(36).substr(2, 9)}`,
      name: request.name || `Request ${tabs.length + 1}`,
      request,
      response: null,
      loading: false,
      hasNewContent: true  // Mark as new when sent from Proxy
    }
    set({ tabs: [...tabs, newTab], activeTabId: newTab.id })
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

  setActiveTab: (id) => set((state) => ({
    activeTabId: id,
    // Clear hasNewContent flag when tab becomes active
    tabs: state.tabs.map(tab =>
      tab.id === id ? { ...tab, hasNewContent: false } : tab
    )
  })),

  updateRequest: (id, request) => set((state) => ({
    tabs: state.tabs.map(tab =>
      tab.id === id
        ? { ...tab, request: { ...tab.request, ...request } }
        : tab
    )
  })),

  setResponse: (id, response) => set((state) => ({
    tabs: state.tabs.map(tab =>
      tab.id === id ? { ...tab, response } : tab
    )
  })),

  setLoading: (id, loading) => set((state) => ({
    tabs: state.tabs.map(tab =>
      tab.id === id ? { ...tab, loading } : tab
    )
  })),

  duplicateTab: (id) => {
    const { tabs } = get()
    const tabToDuplicate = tabs.find(t => t.id === id)
    if (tabToDuplicate) {
      const newTab: RepeaterTab = {
        ...tabToDuplicate,
        id: `tab-${Date.now()}-${Math.random().toString(36).substr(2, 9)}`,
        name: `${tabToDuplicate.name} (copy)`,
        response: null,
        hasNewContent: false  // Duplicated tab is not "new"
      }
      set({ tabs: [...tabs, newTab] })
      return newTab.id
    }
    return ''
  }
}))
