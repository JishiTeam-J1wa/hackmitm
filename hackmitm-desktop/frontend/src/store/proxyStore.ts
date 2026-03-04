import { create } from 'zustand'
import type { ProxyStatus } from '@/types'

interface ProxyState {
  status: ProxyStatus
  serverHost: string
  serverPort: number
  apiEndpoint: string
  connected: boolean

  // Actions
  setStatus: (status: Partial<ProxyStatus>) => void
  setConnected: (connected: boolean) => void
  setServerConfig: (host: string, port: number) => void
  setApiEndpoint: (endpoint: string) => void
}

const initialState: Omit<ProxyState, 'setStatus' | 'setConnected' | 'setServerConfig' | 'setApiEndpoint'> = {
  status: {
    running: false,
    port: 4443,  // 代理端口
    interceptMode: false,
    activeConnections: 0,
    totalRequests: 0,
    uptime: 0
  },
  serverHost: 'localhost',
  serverPort: 4443,  // 代理端口
  apiEndpoint: 'http://localhost:9090',  // 监控API端口
  connected: false,
}

export const useProxyStore = create<ProxyState>((set) => ({
  ...initialState,

  setStatus: (status) => set((state) => ({
    status: { ...state.status, ...status }
  })),

  setConnected: (connected) => set({ connected }),

  setServerConfig: (host, port) => set({ serverHost: host, serverPort: port }),

  setApiEndpoint: (apiEndpoint) => set({ apiEndpoint }),
}))
