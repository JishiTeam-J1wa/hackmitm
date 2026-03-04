import { Play, X } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { useTrafficStore } from '@/store'
import { ForwardIntercepted, DropIntercepted } from '../../../wailsjs/go/main/App'
import type { TrafficItem } from '@/types'

interface InterceptActionsProps {
  item: TrafficItem
}

export function InterceptActions({ item }: InterceptActionsProps) {
  const { forwardIntercepted, dropIntercepted } = useTrafficStore()

  const handleForward = async () => {
    try {
      await ForwardIntercepted(item.id)
      forwardIntercepted()
    } catch (error) {
      console.error('Failed to forward request:', error)
    }
  }

  const handleDrop = async () => {
    try {
      await DropIntercepted(item.id)
      dropIntercepted()
    } catch (error) {
      console.error('Failed to drop request:', error)
    }
  }

  return (
    <div className="flex items-center justify-center gap-4 px-4 py-3 bg-destructive/10 border-b border-destructive/20">
      <span className="text-sm font-medium text-destructive">
        Request Intercepted - Choose an action:
      </span>

      <div className="flex gap-2">
        <Button
          variant="default"
          size="sm"
          onClick={handleForward}
          className="bg-green-600 hover:bg-green-700"
        >
          <Play className="w-4 h-4 mr-1" />
          Forward
        </Button>

        <Button
          variant="destructive"
          size="sm"
          onClick={handleDrop}
        >
          <X className="w-4 h-4 mr-1" />
          Drop
        </Button>
      </div>
    </div>
  )
}
