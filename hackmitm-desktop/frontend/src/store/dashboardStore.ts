import { create } from 'zustand'
import { DashboardMetrics, TrafficStats } from '@/types'

interface DashboardState {
  metrics: DashboardMetrics
  trafficStats: TrafficStats
  chartData: {
    time: string
    requests: number
    responseTime: number
  }[]

  // Actions
  setMetrics: (metrics: Partial<DashboardMetrics>) => void
  setTrafficStats: (stats: Partial<TrafficStats>) => void
  addChartData: (data: { time: string; requests: number; responseTime: number }) => void
}

export const useDashboardStore = create<DashboardState>((set) => ({
  metrics: {
    qps: 0,
    avgResponseTime: 0,
    activeConnections: 0,
    totalRequests: 0,
    totalBytesIn: 0,
    totalBytesOut: 0,
    errorRate: 0,
    uptime: 0
  },
  trafficStats: {
    requestsPerSecond: [],
    responseTimes: [],
    statusCodes: {},
    topHosts: [],
    methods: {}
  },
  chartData: [],

  setMetrics: (metrics) => set((state) => ({
    metrics: { ...state.metrics, ...metrics }
  })),

  setTrafficStats: (stats) => set((state) => ({
    trafficStats: { ...state.trafficStats, ...stats }
  })),

  addChartData: (data) => set((state) => ({
    chartData: [...state.chartData, data].slice(-60) // Keep last 60 data points
  }))
}))
