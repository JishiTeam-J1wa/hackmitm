import { create } from 'zustand'

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

interface InterceptState {
  // 拦截队列
  queue: InterceptedRequest[]
  // 当前选中的拦截请求
  selectedRequest: InterceptedRequest | null
  // 是否处于拦截模式
  interceptEnabled: boolean

  // Actions
  addRequest: (request: InterceptedRequest) => void
  removeRequest: (id: string) => void
  selectRequest: (request: InterceptedRequest | null) => void
  updateRequest: (id: string, updates: Partial<InterceptedRequest>) => void
  clearQueue: () => void
  setInterceptEnabled: (enabled: boolean) => void
}

export const useInterceptStore = create<InterceptState>((set) => ({
  queue: [],
  selectedRequest: null,
  interceptEnabled: false,

  addRequest: (request) => set((state) => ({
    queue: [...state.queue, request]
  })),

  removeRequest: (id) => set((state) => ({
    queue: state.queue.filter(r => r.id !== id),
    selectedRequest: state.selectedRequest?.id === id ? null : state.selectedRequest
  })),

  selectRequest: (request) => set({ selectedRequest: request }),

  updateRequest: (id, updates) => set((state) => ({
    queue: state.queue.map(r => r.id === id ? { ...r, ...updates } : r),
    selectedRequest: state.selectedRequest?.id === id
      ? { ...state.selectedRequest, ...updates }
      : state.selectedRequest
  })),

  clearQueue: () => set({ queue: [], selectedRequest: null }),

  setInterceptEnabled: (enabled) => set({ interceptEnabled: enabled }),
}))
