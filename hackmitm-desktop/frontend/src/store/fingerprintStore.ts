import { create } from 'zustand'
import { FingerprintResult } from '@/types'

interface FingerprintState {
  results: FingerprintResult[]
  selectedResult: FingerprintResult | null
  stats: {
    totalScans: number
    uniqueTechs: number
    topTechnologies: { name: string; count: number }[]
  }

  // Actions
  addResult: (result: FingerprintResult) => void
  selectResult: (result: FingerprintResult | null) => void
  setStats: (stats: Partial<FingerprintState['stats']>) => void
  clearResults: () => void
}

export const useFingerprintStore = create<FingerprintState>((set) => ({
  results: [],
  selectedResult: null,
  stats: {
    totalScans: 0,
    uniqueTechs: 0,
    topTechnologies: []
  },

  addResult: (result) => set((state) => {
    const newResults = [result, ...state.results].slice(0, 1000)

    // Calculate stats
    const techCounts: Record<string, number> = {}
    newResults.forEach(r => {
      r.fingerprints.forEach(f => {
        techCounts[f] = (techCounts[f] || 0) + 1
      })
    })

    const topTechnologies = Object.entries(techCounts)
      .map(([name, count]) => ({ name, count }))
      .sort((a, b) => b.count - a.count)
      .slice(0, 10)

    return {
      results: newResults,
      stats: {
        totalScans: newResults.length,
        uniqueTechs: Object.keys(techCounts).length,
        topTechnologies
      }
    }
  }),

  selectResult: (result) => set({ selectedResult: result }),

  setStats: (stats) => set((state) => ({
    stats: { ...state.stats, ...stats }
  })),

  clearResults: () => set({
    results: [],
    selectedResult: null,
    stats: {
      totalScans: 0,
      uniqueTechs: 0,
      topTechnologies: []
    }
  })
}))
