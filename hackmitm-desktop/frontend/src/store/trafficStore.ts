import { create } from 'zustand'
import { TrafficItem } from '@/types'

// 拦截请求类型
export interface InterceptedRequest {
  id: string
  timestamp: string
  method: string
  url: string
  host: string
  path: string
  headers: Record<string, string>
  requestHeaders: Record<string, string>
  body: string
  requestBody: string
  contentType: string
  clientIP: string
  protocol: string
  requestSize: number
  statusCode?: number
}

interface TrafficState {
  // 流量列表
  items: TrafficItem[]
  selectedItem: TrafficItem | null

  // 拦截功能
  interceptMode: boolean
  interceptQueue: InterceptedRequest[]
  selectedInterceptItem: InterceptedRequest | null

  // 筛选
  filter: {
    method: string
    host: string
    path: string
    statusCode: string
    search: string
  }

  // 流量操作
  addItem: (item: TrafficItem) => void
  addItems: (items: TrafficItem[]) => void
  selectItem: (item: TrafficItem | null) => void
  clearItems: () => void
  deleteItem: (id: string) => void

  // 拦截操作
  setInterceptMode: (enabled: boolean) => void
  addToInterceptQueue: (request: InterceptedRequest) => void
  selectInterceptItem: (item: InterceptedRequest | null) => void
  setSelectedInterceptItem: (item: InterceptedRequest | null) => void
  removeFromInterceptQueue: (id: string) => void
  clearInterceptQueue: () => void
  updateInterceptedRequest: (id: string, updates: Partial<InterceptedRequest>) => void
  setInterceptEnabled: (enabled: boolean) => void
  forwardIntercepted: () => void
  dropIntercepted: () => void
  interceptedItem: InterceptedRequest | null

  // 筛选操作
  setFilter: (filter: Partial<TrafficState['filter']>) => void
  resetFilter: () => void
}

const initialFilter = {
  method: '',
  host: '',
  path: '',
  statusCode: '',
  search: ''
}

export const useTrafficStore = create<TrafficState>((set) => ({
  items: [],
  selectedItem: null,
  interceptMode: false,
  interceptQueue: [],
  selectedInterceptItem: null,
  filter: initialFilter,

  // 流量操作
  addItem: (item) => {
    set((state) => ({
      items: [item, ...state.items].slice(0, 10000)
    }))
  },

  addItems: (items) => {
    set((state) => ({
      items: [...items, ...state.items].slice(0, 10000)
    }))
  },

  selectItem: (item) => set({ selectedItem: item }),

  clearItems: () => set({ items: [], selectedItem: null }),

  deleteItem: (id) => set((state) => ({
    items: state.items.filter(item => item.id !== id),
    selectedItem: state.selectedItem?.id === id ? null : state.selectedItem
  })),

  // 拦截操作
  setInterceptMode: (enabled) => set({ interceptMode: enabled }),

  addToInterceptQueue: (request) => set((state) => ({
    interceptQueue: [...state.interceptQueue, request]
  })),

  selectInterceptItem: (item) => set({ selectedInterceptItem: item }),

  removeFromInterceptQueue: (id) => set((state) => ({
    interceptQueue: state.interceptQueue.filter(r => r.id !== id),
    selectedInterceptItem: state.selectedInterceptItem?.id === id ? null : state.selectedInterceptItem
  })),

  clearInterceptQueue: () => set({ interceptQueue: [], selectedInterceptItem: null }),

  updateInterceptedRequest: (id, updates) => set((state) => ({
    interceptQueue: state.interceptQueue.map(r => r.id === id ? { ...r, ...updates } : r),
    selectedInterceptItem: state.selectedInterceptItem?.id === id
      ? { ...state.selectedInterceptItem, ...updates }
      : state.selectedInterceptItem
  })),

  setSelectedInterceptItem: (item) => set({ selectedInterceptItem: item }),

  setInterceptEnabled: (enabled) => set({ interceptMode: enabled }),

  forwardIntercepted: () => set({ selectedInterceptItem: null }),

  dropIntercepted: () => set((state) => ({
    selectedInterceptItem: null,
    interceptQueue: state.interceptQueue.filter(r => r.id !== state.selectedInterceptItem?.id)
  })),

  interceptedItem: null,

  // 筛选操作
  setFilter: (filter) => set((state) => ({
    filter: { ...state.filter, ...filter }
  })),

  resetFilter: () => set({ filter: initialFilter }),
}))

// 向后兼容的拦截 store 导出
// 现在直接使用 useTrafficStore 的拦截相关方法
export const useInterceptStore = {
  get queue() {
    return useTrafficStore.getState().interceptQueue
  },
  get selectedRequest() {
    return useTrafficStore.getState().selectedInterceptItem
  },
  get interceptEnabled() {
    return useTrafficStore.getState().interceptMode
  },
  addRequest: (request: InterceptedRequest) => {
    useTrafficStore.getState().addToInterceptQueue(request)
  },
  removeRequest: (id: string) => {
    useTrafficStore.getState().removeFromInterceptQueue(id)
  },
  selectRequest: (request: InterceptedRequest | null) => {
    useTrafficStore.getState().selectInterceptItem(request)
  },
  updateRequest: (id: string, updates: Partial<InterceptedRequest>) => {
    useTrafficStore.getState().updateInterceptedRequest(id, updates)
  },
  clearQueue: () => {
    useTrafficStore.getState().clearInterceptQueue()
  },
  setInterceptEnabled: (enabled: boolean) => {
    useTrafficStore.getState().setInterceptMode(enabled)
  },
}
