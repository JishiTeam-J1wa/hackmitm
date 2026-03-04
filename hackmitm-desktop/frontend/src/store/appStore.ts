import { create } from 'zustand'
import { persist } from 'zustand/middleware'

export type ConnectionMode = 'remote' | 'local' | null

export interface LocalConfig {
  dataDir: string
  apiPort: number
  proxyPort: number
}

export interface RemoteConfig {
  host: string
  port: number
  apiKey?: string
}

export interface AppConfig {
  connectionMode: ConnectionMode
  localConfig: LocalConfig | null
  remoteConfig: RemoteConfig | null
  initialized: boolean
  // UI settings
  theme?: string
  interceptEnabled?: boolean
  defaultMethodFilter?: string
  defaultStatusFilter?: string
}

interface AppState {
  appConfig: AppConfig
  backendSynced: boolean

  // Actions
  setConnectionMode: (mode: ConnectionMode) => void
  setLocalConfig: (config: LocalConfig) => void
  setRemoteConfig: (config: RemoteConfig) => void
  completeInitialization: () => void
  resetInitialization: () => void
  setTheme: (theme: string) => void
  setInterceptEnabled: (enabled: boolean) => void
  setDefaultFilters: (method: string, status: string) => void

  // Backend sync
  syncFromBackend: () => Promise<void>
  syncToBackend: () => Promise<void>
  loadSavedConfig: () => void
}

const initialConfig: AppConfig = {
  connectionMode: null,
  localConfig: null,
  remoteConfig: null,
  initialized: false,
  theme: 'light',
  interceptEnabled: false,
  defaultMethodFilter: 'all',
  defaultStatusFilter: 'all',
}

export const useAppStore = create<AppState>()(
  persist(
    (set, get) => ({
      appConfig: initialConfig,
      backendSynced: false,

      setConnectionMode: (mode) => {
        set((state) => ({
          appConfig: { ...state.appConfig, connectionMode: mode }
        }))
        // Auto-sync to backend
        get().syncToBackend()
      },

      setLocalConfig: (config) => {
        set((state) => ({
          appConfig: { ...state.appConfig, localConfig: config }
        }))
        get().syncToBackend()
      },

      setRemoteConfig: (config) => {
        set((state) => ({
          appConfig: { ...state.appConfig, remoteConfig: config }
        }))
        get().syncToBackend()
      },

      completeInitialization: () => {
        set((state) => ({
          appConfig: { ...state.appConfig, initialized: true },
        }))
        get().syncToBackend()
      },

      resetInitialization: () => {
        set({
          appConfig: initialConfig,
          backendSynced: false,
        })
      },

      setTheme: (theme) => {
        set((state) => ({
          appConfig: { ...state.appConfig, theme }
        }))
        get().syncToBackend()
      },

      setInterceptEnabled: (enabled) => {
        set((state) => ({
          appConfig: { ...state.appConfig, interceptEnabled: enabled }
        }))
        get().syncToBackend()
      },

      setDefaultFilters: (method, status) => {
        set((state) => ({
          appConfig: {
            ...state.appConfig,
            defaultMethodFilter: method,
            defaultStatusFilter: status
          }
        }))
      },

      syncFromBackend: async () => {
        try {
          const { GetAppConfig } = await import('../../wailsjs/go/main/App')
          const backendConfig = await GetAppConfig()

          if (backendConfig) {
            const newConfig: AppConfig = {
              connectionMode: backendConfig.lastConnectionMode as ConnectionMode || null,
              localConfig: backendConfig.localDataDir ? {
                dataDir: backendConfig.localDataDir,
                apiPort: backendConfig.localApiPort || 9090,
                proxyPort: backendConfig.localProxyPort || 4443,
              } : null,
              remoteConfig: backendConfig.remoteHost ? {
                host: backendConfig.remoteHost,
                port: backendConfig.remotePort || 9090,
                apiKey: backendConfig.remoteApiKey,
              } : null,
              initialized: !!backendConfig.lastConnectionMode,
              theme: backendConfig.theme || 'light',
              interceptEnabled: backendConfig.interceptEnabled || false,
              defaultMethodFilter: backendConfig.defaultMethodFilter || 'all',
              defaultStatusFilter: backendConfig.defaultStatusFilter || 'all',
            }

            set({ appConfig: newConfig, backendSynced: true })
          }
        } catch (e) {
          console.warn('Failed to sync from backend:', e)
        }
      },

      syncToBackend: async () => {
        try {
          const config = get().appConfig
          const { UpdateAppConfig } = await import('../../wailsjs/go/main/App')

          const updates: Record<string, any> = {
            lastConnectionMode: config.connectionMode || '',
            theme: config.theme || 'light',
            interceptEnabled: config.interceptEnabled || false,
            defaultMethodFilter: config.defaultMethodFilter || 'all',
            defaultStatusFilter: config.defaultStatusFilter || 'all',
          }

          if (config.localConfig) {
            updates.localDataDir = config.localConfig.dataDir
            updates.localApiPort = config.localConfig.apiPort
            updates.localProxyPort = config.localConfig.proxyPort
          }

          if (config.remoteConfig) {
            updates.remoteHost = config.remoteConfig.host
            updates.remotePort = config.remoteConfig.port
            if (config.remoteConfig.apiKey) {
              updates.remoteApiKey = config.remoteConfig.apiKey
            }
          }

          await UpdateAppConfig(updates)
          set({ backendSynced: true })
        } catch (e) {
          console.warn('Failed to sync to backend:', e)
        }
      },

      loadSavedConfig: () => {
        const state = get()
        if (state.appConfig.initialized) {
          // Config already loaded from persist, now sync with backend
          state.syncFromBackend()
        }
      },
    }),
    {
      name: 'hackmitm-app-config',
      partialize: (state) => ({
        appConfig: state.appConfig,
      }),
    }
  )
)
