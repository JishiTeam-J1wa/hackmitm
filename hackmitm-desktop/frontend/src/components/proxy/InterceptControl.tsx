import { Trash2 } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Switch } from '@/components/ui/switch'
import { Badge } from '@/components/ui/badge'
import { useTrafficStore } from '@/store'
import { SetInterceptMode, ClearTraffic } from '../../../wailsjs/go/main/App'

export function InterceptControl() {
  const { interceptMode, setInterceptMode, interceptedItem } = useTrafficStore()

  const handleInterceptToggle = async () => {
    try {
      await SetInterceptMode(!interceptMode)
      setInterceptMode(!interceptMode)
    } catch (error) {
      console.error('Failed to toggle intercept mode:', error)
    }
  }

  const handleClear = async () => {
    try {
      await ClearTraffic()
      useTrafficStore.getState().clearItems()
    } catch (error) {
      console.error('Failed to clear traffic:', error)
    }
  }

  return (
    <div className="flex items-center gap-3">
      {/* Intercept toggle */}
      <div className="flex items-center gap-2">
        <Switch
          checked={interceptMode}
          onCheckedChange={handleInterceptToggle}
        />
        <span className="text-sm">Intercept</span>
        {interceptMode && (
          <Badge variant="destructive" className="animate-pulse text-xs">
            ON
          </Badge>
        )}
      </div>

      {/* Clear button */}
      <Button
        variant="ghost"
        size="sm"
        onClick={handleClear}
        title="Clear traffic"
      >
        <Trash2 className="w-4 h-4" />
      </Button>

      {/* Intercepted item indicator */}
      {interceptedItem && (
        <Badge variant="destructive" className="animate-pulse">
          Request Intercepted
        </Badge>
      )}
    </div>
  )
}
