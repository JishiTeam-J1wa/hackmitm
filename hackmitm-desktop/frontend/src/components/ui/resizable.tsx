import { useState, useRef, useCallback, useEffect } from 'react'
import { cn } from '@/lib/utils'

interface ResizablePanelProps {
  children: React.ReactNode
  direction?: 'horizontal' | 'vertical'
  defaultSizes?: number[]
  minSizes?: number[]
  className?: string
}

export function ResizablePanelGroup({
  children,
  direction = 'horizontal',
  defaultSizes,
  minSizes = [100, 100],
  className,
}: ResizablePanelProps) {
  const childArray = Array.isArray(children) ? children : [children]
  const [sizes, setSizes] = useState<number[]>(defaultSizes || [50, 50])
  const [isDragging, setIsDragging] = useState(false)
  const containerRef = useRef<HTMLDivElement>(null)
  const startPos = useRef(0)
  const startSizes = useRef<number[]>([])

  const handleMouseDown = useCallback((e: React.MouseEvent) => {
    e.preventDefault()
    setIsDragging(true)
    startPos.current = direction === 'horizontal' ? e.clientX : e.clientY
    startSizes.current = [...sizes]
    document.body.style.cursor = direction === 'horizontal' ? 'col-resize' : 'row-resize'
    document.body.style.userSelect = 'none'
  }, [direction, sizes])

  const handleMouseMove = useCallback((e: MouseEvent) => {
    if (!containerRef.current || startSizes.current.length === 0) return

    const rect = containerRef.current.getBoundingClientRect()
    const currentPos = direction === 'horizontal' ? e.clientX : e.clientY
    const totalSize = direction === 'horizontal' ? rect.width : rect.height
    const delta = currentPos - startPos.current
    const deltaPercent = (delta / totalSize) * 100

    const newSizes = [...startSizes.current]
    newSizes[0] = Math.max(minSizes[0], Math.min(100 - minSizes[1], startSizes.current[0] + deltaPercent))
    newSizes[1] = 100 - newSizes[0]

    setSizes(newSizes)
  }, [direction, minSizes])

  const handleMouseUp = useCallback(() => {
    setIsDragging(false)
    startSizes.current = []
    document.body.style.cursor = ''
    document.body.style.userSelect = ''
  }, [])

  useEffect(() => {
    if (isDragging) {
      document.addEventListener('mousemove', handleMouseMove)
      document.addEventListener('mouseup', handleMouseUp)
      return () => {
        document.removeEventListener('mousemove', handleMouseMove)
        document.removeEventListener('mouseup', handleMouseUp)
      }
    }
  }, [isDragging, handleMouseMove, handleMouseUp])

  return (
    <div
      ref={containerRef}
      className={cn(
        'flex h-full w-full',
        direction === 'vertical' && 'flex-col',
        className
      )}
    >
      <div
        style={{ [direction === 'horizontal' ? 'width' : 'height']: `${sizes[0]}%` }}
        className={cn(
          'min-w-0 min-h-0 overflow-hidden',
          direction === 'horizontal' ? 'flex-shrink-0' : 'flex-shrink-0'
        )}
      >
        {childArray[0]}
      </div>

      {/* Resizer handle - 增大热区域 */}
      <div
        onMouseDown={handleMouseDown}
        className={cn(
          'flex-shrink-0 relative group',
          direction === 'horizontal'
            ? 'w-2 cursor-col-resize'
            : 'h-2 cursor-row-resize'
        )}
      >
        {/* 拖拽时的高亮线 */}
        <div
          className={cn(
            'absolute transition-all duration-150',
            isDragging ? 'bg-blue-500' : 'bg-transparent group-hover:bg-blue-400',
            direction === 'horizontal'
              ? 'left-1/2 -translate-x-1/2 top-0 bottom-0 w-0.5 group-hover:w-1'
              : 'top-1/2 -translate-y-1/2 left-0 right-0 h-0.5 group-hover:h-1'
          )}
        />

        {/* 拖拽时的拖拽区域背景 */}
        <div
          className={cn(
            'absolute inset-0 transition-colors duration-150',
            isDragging
              ? 'bg-blue-100/50'
              : 'bg-transparent hover:bg-blue-100/30'
          )}
        />

        {/* 拖拽手柄指示器 */}
        <div
          className={cn(
            'absolute opacity-0 group-hover:opacity-100 transition-opacity duration-150',
            direction === 'horizontal'
              ? 'left-1/2 -translate-x-1/2 top-1/2 -translate-y-1/2 w-0.5 h-8 bg-blue-400 rounded-full'
              : 'top-1/2 -translate-y-1/2 left-1/2 -translate-x-1/2 h-0.5 w-8 bg-blue-400 rounded-full'
          )}
        />
      </div>

      <div
        style={{ [direction === 'horizontal' ? 'width' : 'height']: `${sizes[1]}%` }}
        className="min-w-0 min-h-0 flex-1 overflow-hidden"
      >
        {childArray[1]}
      </div>
    </div>
  )
}

// 三栏可调整面板
interface ThreePanelProps {
  top: React.ReactNode
  middle: React.ReactNode
  bottom: React.ReactNode
  defaultHeights?: [number, number, number]
  minHeights?: [number, number, number]
  className?: string
}

export function ResizableThreePanels({
  top,
  middle,
  bottom,
  defaultHeights = [35, 35, 30],
  minHeights = [80, 80, 80],
  className,
}: ThreePanelProps) {
  const [heights, setHeights] = useState<[number, number, number]>(defaultHeights)
  const [dragIndex, setDragIndex] = useState<number | null>(null)
  const containerRef = useRef<HTMLDivElement>(null)
  const startPos = useRef(0)
  const startHeights = useRef<[number, number, number]>([35, 35, 30])

  const handleMouseDown = useCallback((e: React.MouseEvent, index: number) => {
    e.preventDefault()
    setDragIndex(index)
    startPos.current = e.clientY
    startHeights.current = [...heights]
    document.body.style.cursor = 'row-resize'
    document.body.style.userSelect = 'none'
  }, [heights])

  const handleMouseMove = useCallback((e: MouseEvent) => {
    if (dragIndex === null || !containerRef.current) return

    const rect = containerRef.current.getBoundingClientRect()
    const totalHeight = rect.height
    const delta = e.clientY - startPos.current
    const deltaPercent = (delta / totalHeight) * 100

    const newHeights = [...startHeights.current] as [number, number, number]

    if (dragIndex === 0) {
      const newTop = Math.max(minHeights[0], Math.min(100 - minHeights[1] - minHeights[2], startHeights.current[0] + deltaPercent))
      const remaining = 100 - newTop - startHeights.current[2]
      if (remaining >= minHeights[1]) {
        newHeights[0] = newTop
        newHeights[1] = remaining
      }
    } else {
      const newBottom = Math.max(minHeights[2], Math.min(100 - minHeights[0] - minHeights[1], startHeights.current[2] - deltaPercent))
      const remaining = 100 - newBottom - startHeights.current[0]
      if (remaining >= minHeights[1]) {
        newHeights[2] = newBottom
        newHeights[1] = remaining
      }
    }

    setHeights(newHeights)
  }, [dragIndex, minHeights])

  const handleMouseUp = useCallback(() => {
    setDragIndex(null)
    document.body.style.cursor = ''
    document.body.style.userSelect = ''
  }, [])

  useEffect(() => {
    if (dragIndex !== null) {
      document.addEventListener('mousemove', handleMouseMove)
      document.addEventListener('mouseup', handleMouseUp)
      return () => {
        document.removeEventListener('mousemove', handleMouseMove)
        document.removeEventListener('mouseup', handleMouseUp)
      }
    }
  }, [dragIndex, handleMouseMove, handleMouseUp])

  return (
    <div ref={containerRef} className={cn('flex flex-col h-full', className)}>
      <div style={{ height: `${heights[0]}%` }} className="min-h-0 overflow-hidden">
        {top}
      </div>

      {/* 拖拽分割线 1 */}
      <div
        onMouseDown={(e) => handleMouseDown(e, 0)}
        className={cn(
          'h-2 flex-shrink-0 relative group cursor-row-resize',
          dragIndex === 0 ? 'bg-blue-100' : 'bg-transparent hover:bg-blue-50'
        )}
      >
        <div className={cn(
          'absolute left-0 right-0 top-1/2 -translate-y-1/2 h-0.5 transition-colors',
          dragIndex === 0 ? 'bg-blue-500' : 'bg-gray-200 group-hover:bg-blue-400'
        )} />
      </div>

      <div style={{ height: `${heights[1]}%` }} className="min-h-0 overflow-hidden">
        {middle}
      </div>

      {/* 拖拽分割线 2 */}
      <div
        onMouseDown={(e) => handleMouseDown(e, 1)}
        className={cn(
          'h-2 flex-shrink-0 relative group cursor-row-resize',
          dragIndex === 1 ? 'bg-blue-100' : 'bg-transparent hover:bg-blue-50'
        )}
      >
        <div className={cn(
          'absolute left-0 right-0 top-1/2 -translate-y-1/2 h-0.5 transition-colors',
          dragIndex === 1 ? 'bg-blue-500' : 'bg-gray-200 group-hover:bg-blue-400'
        )} />
      </div>

      <div style={{ height: `${heights[2]}%` }} className="min-h-0 overflow-hidden">
        {bottom}
      </div>
    </div>
  )
}
