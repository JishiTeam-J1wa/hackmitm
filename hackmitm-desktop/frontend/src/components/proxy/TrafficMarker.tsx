import { useState, useCallback, useMemo } from 'react'
import {
  Tag,
  Star,
  Flag,
  Bookmark,
  AlertTriangle,
  CheckCircle,
  X,
  Plus,
  Edit2,
} from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogFooter,
} from '@/components/ui/dialog'
import { cn } from '@/lib/utils'

export interface TrafficMarker {
  id: string
  name: string
  color: string
  icon?: string
  createdAt: string
}

export interface TrafficNote {
  id: string
  trafficId: string
  markerId: string
  note: string
  createdAt: string
  updatedAt: string
}

interface TrafficMarkerProps {
  trafficId: string
  currentMarkerId?: string
  currentNote?: string
  markers: TrafficMarker[]
  onMark: (trafficId: string, markerId: string, note?: string) => void
  onUnmark: (trafficId: string) => void
  onCreateMarker: (marker: Omit<TrafficMarker, 'id' | 'createdAt'>) => void
  onDeleteMarker: (markerId: string) => void
}

const defaultColors = [
  '#EF4444', // red
  '#F97316', // orange
  '#EAB308', // yellow
  '#22C55E', // green
  '#3B82F6', // blue
  '#8B5CF6', // purple
  '#EC4899', // pink
  '#6B7280', // gray
]

const presetMarkers: TrafficMarker[] = [
  { id: 'important', name: 'Important', color: '#EAB308', icon: 'star', createdAt: new Date().toISOString() },
  { id: 'suspicious', name: 'Suspicious', color: '#EF4444', icon: 'alert', createdAt: new Date().toISOString() },
  { id: 'reviewed', name: 'Reviewed', color: '#22C55E', icon: 'check', createdAt: new Date().toISOString() },
  { id: 'todo', name: 'To Do', color: '#3B82F6', icon: 'flag', createdAt: new Date().toISOString() },
]

export function TrafficMarker({
  trafficId,
  currentMarkerId,
  currentNote,
  markers,
  onMark,
  onUnmark,
  onCreateMarker,
  onDeleteMarker: _onDeleteMarker,
}: TrafficMarkerProps) {
  const [showCreateDialog, setShowCreateDialog] = useState(false)
  const [showNoteDialog, setShowNoteDialog] = useState(false)
  const [newMarkerName, setNewMarkerName] = useState('')
  const [newMarkerColor, setNewMarkerColor] = useState(defaultColors[0])
  const [note, setNote] = useState(currentNote || '')

  const allMarkers = useMemo(() => [...presetMarkers, ...markers], [markers])

  const currentMarker = useMemo(
    () => allMarkers.find((m) => m.id === currentMarkerId),
    [allMarkers, currentMarkerId]
  )

  const handleMark = useCallback(
    (markerId: string) => {
      if (currentMarkerId === markerId) {
        onUnmark(trafficId)
      } else {
        onMark(trafficId, markerId, note)
      }
    },
    [trafficId, currentMarkerId, note, onMark, onUnmark]
  )

  const handleCreateMarker = useCallback(() => {
    if (newMarkerName.trim()) {
      onCreateMarker({
        name: newMarkerName.trim(),
        color: newMarkerColor,
      })
      setNewMarkerName('')
      setShowCreateDialog(false)
    }
  }, [newMarkerName, newMarkerColor, onCreateMarker])

  const handleAddNote = useCallback(() => {
    if (currentMarkerId) {
      onMark(trafficId, currentMarkerId, note)
    }
    setShowNoteDialog(false)
  }, [currentMarkerId, trafficId, note, onMark])

  const getMarkerIcon = (marker: TrafficMarker) => {
    switch (marker.icon) {
      case 'star':
        return <Star className="w-3 h-3" />
      case 'alert':
        return <AlertTriangle className="w-3 h-3" />
      case 'check':
        return <CheckCircle className="w-3 h-3" />
      case 'flag':
        return <Flag className="w-3 h-3" />
      default:
        return <Bookmark className="w-3 h-3" />
    }
  }

  return (
    <>
      <DropdownMenu>
        <DropdownMenuTrigger asChild>
          <Button
            variant="ghost"
            size="sm"
            className={cn(
              'h-7 gap-1 text-xs',
              currentMarker && 'bg-opacity-10'
            )}
            style={currentMarker ? { backgroundColor: `${currentMarker.color}20` } : undefined}
          >
            {currentMarker ? (
              <>
                <div
                  className="w-3 h-3 rounded-full flex items-center justify-center"
                  style={{ backgroundColor: currentMarker.color }}
                >
                  {getMarkerIcon(currentMarker)}
                </div>
                <span style={{ color: currentMarker.color }}>{currentMarker.name}</span>
              </>
            ) : (
              <>
                <Tag className="w-3 h-3" />
                Mark
              </>
            )}
          </Button>
        </DropdownMenuTrigger>

        <DropdownMenuContent align="end" className="w-48">
          {allMarkers.map((marker) => (
            <DropdownMenuItem
              key={marker.id}
              onClick={() => handleMark(marker.id)}
              className="flex items-center justify-between"
            >
              <div className="flex items-center gap-2">
                <div
                  className="w-3 h-3 rounded-full flex items-center justify-center text-white"
                  style={{ backgroundColor: marker.color }}
                >
                  {getMarkerIcon(marker)}
                </div>
                <span>{marker.name}</span>
              </div>
              {currentMarkerId === marker.id && (
                <CheckCircle className="w-4 h-4 text-green-500" />
              )}
            </DropdownMenuItem>
          ))}

          <DropdownMenuSeparator />

          {currentMarker && (
            <>
              <DropdownMenuItem onClick={() => setShowNoteDialog(true)}>
                <Edit2 className="w-4 h-4 mr-2" />
                Edit Note
              </DropdownMenuItem>
              <DropdownMenuItem onClick={() => onUnmark(trafficId)} className="text-red-500">
                <X className="w-4 h-4 mr-2" />
                Remove Mark
              </DropdownMenuItem>
              <DropdownMenuSeparator />
            </>
          )}

          <DropdownMenuItem onClick={() => setShowCreateDialog(true)}>
            <Plus className="w-4 h-4 mr-2" />
            Create Marker
          </DropdownMenuItem>
        </DropdownMenuContent>
      </DropdownMenu>

      {/* Create Marker Dialog */}
      <Dialog open={showCreateDialog} onOpenChange={setShowCreateDialog}>
        <DialogContent className="sm:max-w-[400px]">
          <DialogHeader>
            <DialogTitle className="text-sm">Create New Marker</DialogTitle>
          </DialogHeader>

          <div className="space-y-4 py-4">
            <div className="space-y-2">
              <label className="text-xs font-medium text-gray-700">Name</label>
              <Input
                value={newMarkerName}
                onChange={(e) => setNewMarkerName(e.target.value)}
                placeholder="Marker name"
                className="h-8 text-xs"
              />
            </div>

            <div className="space-y-2">
              <label className="text-xs font-medium text-gray-700">Color</label>
              <div className="flex gap-2 flex-wrap">
                {defaultColors.map((color) => (
                  <button
                    key={color}
                    onClick={() => setNewMarkerColor(color)}
                    className={cn(
                      'w-6 h-6 rounded-full transition-all',
                      newMarkerColor === color && 'ring-2 ring-offset-2 ring-gray-400'
                    )}
                    style={{ backgroundColor: color }}
                  />
                ))}
              </div>
            </div>
          </div>

          <DialogFooter>
            <Button variant="outline" size="sm" onClick={() => setShowCreateDialog(false)}>
              Cancel
            </Button>
            <Button size="sm" onClick={handleCreateMarker} disabled={!newMarkerName.trim()}>
              Create
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Note Dialog */}
      <Dialog open={showNoteDialog} onOpenChange={setShowNoteDialog}>
        <DialogContent className="sm:max-w-[400px]">
          <DialogHeader>
            <DialogTitle className="text-sm">Add Note</DialogTitle>
          </DialogHeader>

          <div className="py-4">
            <Input
              value={note}
              onChange={(e) => setNote(e.target.value)}
              placeholder="Add a note for this traffic item"
              className="text-xs"
            />
          </div>

          <DialogFooter>
            <Button variant="outline" size="sm" onClick={() => setShowNoteDialog(false)}>
              Cancel
            </Button>
            <Button size="sm" onClick={handleAddNote}>
              Save Note
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </>
  )
}

// Marker Badge component for display in lists
export function MarkerBadge({ marker, note }: { marker: TrafficMarker; note?: string }) {
  return (
    <div
      className="inline-flex items-center gap-1 px-1.5 py-0.5 rounded text-[10px]"
      style={{ backgroundColor: `${marker.color}20`, color: marker.color }}
      title={note}
    >
      <Bookmark className="w-2.5 h-2.5" />
      {marker.name}
    </div>
  )
}

export default TrafficMarker
