import { useState, useEffect, useCallback } from 'react'
import { MainLayout } from '@/components/layout'
import { TooltipProvider } from '@/components/ui/tooltip'
import { ContextMenuProvider } from '@/components/ui/ContextMenu'
import { InitScreen, ModeSelectionScreen } from '@/components/init'
import { ToastContainer, useToast } from '@/components/ui/toast'
import { useAppStore } from '@/store/appStore'
import { useProxyStore } from '@/store/proxyStore'
import './style.css'

type AppPhase = 'splash' | 'mode-select' | 'main'

function App() {
  const [phase, setPhase] = useState<AppPhase>('splash')
  const [hydrated, setHydrated] = useState(false)
  const appConfig = useAppStore((state) => state.appConfig)
  const setConnected = useProxyStore((state) => state.setConnected)
  const { toasts, removeToast } = useToast()

  // Wait for zustand persist to hydrate
  useEffect(() => {
    const unsubscribe = useAppStore.persist.onFinishHydration(() => {
      setHydrated(true)
      // Check if already initialized after hydration
      const config = useAppStore.getState().appConfig
      if (config.initialized) {
        setPhase('main')
      }
    })

    // Also check immediately in case already hydrated
    if (useAppStore.persist.hasHydrated()) {
      setHydrated(true)
      if (appConfig.initialized) {
        setPhase('main')
      }
    }

    return unsubscribe
  }, [appConfig.initialized])

  const handleSplashComplete = useCallback(() => {
    // Get fresh config state
    const currentConfig = useAppStore.getState().appConfig
    if (currentConfig.initialized) {
      setConnected(true)  // Mark as connected for returning users
      setPhase('main')
    } else {
      setPhase('mode-select')
    }
  }, [setConnected])

  const handleModeSelectComplete = useCallback(() => {
    setConnected(true)  // Mark as connected to backend
    setPhase('main')
  }, [setConnected])

  // Don't render until hydrated to prevent flash
  if (!hydrated) {
    return null
  }

  return (
    <TooltipProvider>
      <ContextMenuProvider>
        {phase === 'splash' && <InitScreen onComplete={handleSplashComplete} duration={2500} />}
        {phase === 'mode-select' && <ModeSelectionScreen onComplete={handleModeSelectComplete} />}
        {phase === 'main' && <MainLayout />}
        <ToastContainer toasts={toasts} onClose={removeToast} />
      </ContextMenuProvider>
    </TooltipProvider>
  )
}

export default App
