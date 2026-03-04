import { useState, useMemo } from 'react'
import {
  ArrowRight,
  Copy,
  Check,
  X,
  GitCompare,
  AlignLeft,
  AlignCenter,
} from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { ScrollArea } from '@/components/ui/scroll-area'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { cn } from '@/lib/utils'
import type { TrafficItem } from '@/types'

interface RequestCompareProps {
  items: TrafficItem[]
  initialLeft?: TrafficItem
  initialRight?: TrafficItem
  onClose?: () => void
}

type CompareMode = 'split' | 'unified'
type CompareView = 'headers' | 'body' | 'all'

export function RequestCompare({
  items,
  initialLeft,
  initialRight,
  onClose,
}: RequestCompareProps) {
  const [leftItem, setLeftItem] = useState<TrafficItem | null>(initialLeft || null)
  const [rightItem, setRightItem] = useState<TrafficItem | null>(initialRight || null)
  const [mode, setMode] = useState<CompareMode>('split')
  const [view, setView] = useState<CompareView>('all')
  const [copied, setCopied] = useState<string | null>(null)

  const handleCopy = (text: string, id: string) => {
    navigator.clipboard.writeText(text)
    setCopied(id)
    setTimeout(() => setCopied(null), 2000)
  }

  const formatHeaders = (headers: Record<string, string>) => {
    return Object.entries(headers)
      .map(([key, value]) => `${key}: ${value}`)
      .join('\n')
  }

  const getRequestContent = (item: TrafficItem | null, view: CompareView) => {
    if (!item) return ''

    const headerStr = formatHeaders(item.requestHeaders)
    const requestLine = `${item.method} ${item.url}`

    switch (view) {
      case 'headers':
        return `${requestLine}\n${headerStr}`
      case 'body':
        return item.requestBody || '(empty body)'
      default:
        return `${requestLine}\n\n${headerStr}\n\n${item.requestBody || '(empty body)'}`
    }
  }

  // Find differences between two strings
  const findDifferences = (left: string, right: string) => {
    const leftLines = left.split('\n')
    const rightLines = right.split('\n')
    const maxLines = Math.max(leftLines.length, rightLines.length)

    const result: Array<{
      leftLine?: string
      rightLine?: string
      type: 'equal' | 'added' | 'removed' | 'modified'
    }> = []

    for (let i = 0; i < maxLines; i++) {
      const leftLine = leftLines[i]
      const rightLine = rightLines[i]

      if (leftLine === undefined) {
        result.push({ rightLine, type: 'added' })
      } else if (rightLine === undefined) {
        result.push({ leftLine, type: 'removed' })
      } else if (leftLine === rightLine) {
        result.push({ leftLine, rightLine, type: 'equal' })
      } else {
        result.push({ leftLine, rightLine, type: 'modified' })
      }
    }

    return result
  }

  const unifiedDiff = useMemo(() => {
    if (!leftItem || !rightItem || mode !== 'unified') return null
    const leftContent = getRequestContent(leftItem, view)
    const rightContent = getRequestContent(rightItem, view)
    return findDifferences(leftContent, rightContent)
  }, [leftItem, rightItem, mode, view])

  return (
    <div className="flex flex-col h-full bg-gray-50">
      {/* Header */}
      <div className="flex items-center justify-between px-4 py-2 bg-white border-b border-gray-200">
        <div className="flex items-center gap-3">
          <GitCompare className="w-5 h-5 text-blue-500" />
          <h2 className="text-sm font-semibold text-gray-800">Request Comparison</h2>
        </div>

        <div className="flex items-center gap-2">
          {/* Mode toggle */}
          <div className="flex items-center border rounded-md">
            <Button
              variant={mode === 'split' ? 'secondary' : 'ghost'}
              size="sm"
              onClick={() => setMode('split')}
              className="h-7 text-xs"
            >
              <AlignLeft className="w-3 h-3 mr-1" />
              Split
            </Button>
            <Button
              variant={mode === 'unified' ? 'secondary' : 'ghost'}
              size="sm"
              onClick={() => setMode('unified')}
              className="h-7 text-xs"
            >
              <AlignCenter className="w-3 h-3 mr-1" />
              Unified
            </Button>
          </div>

          {/* View select */}
          <Select value={view} onValueChange={(v) => setView(v as CompareView)}>
            <SelectTrigger className="w-28 h-7 text-xs">
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="all">All</SelectItem>
              <SelectItem value="headers">Headers</SelectItem>
              <SelectItem value="body">Body</SelectItem>
            </SelectContent>
          </Select>

          {onClose && (
            <Button variant="ghost" size="icon" onClick={onClose} className="w-7 h-7">
              <X className="w-4 h-4" />
            </Button>
          )}
        </div>
      </div>

      {/* Request selectors */}
      <div className="flex items-center gap-4 px-4 py-2 bg-white border-b border-gray-200">
        <div className="flex-1">
          <Select
            value={leftItem?.id || ''}
            onValueChange={(id) => setLeftItem(items.find((i) => i.id === id) || null)}
          >
            <SelectTrigger className="h-8 text-xs">
              <SelectValue placeholder="Select first request" />
            </SelectTrigger>
            <SelectContent>
              {items.map((item) => (
                <SelectItem key={item.id} value={item.id}>
                  <div className="flex items-center gap-2">
                    <Badge variant="outline" className="text-[10px]">
                      {item.method}
                    </Badge>
                    <span className="truncate">{item.path}</span>
                  </div>
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
        </div>

        <ArrowRight className="w-4 h-4 text-gray-400" />

        <div className="flex-1">
          <Select
            value={rightItem?.id || ''}
            onValueChange={(id) => setRightItem(items.find((i) => i.id === id) || null)}
          >
            <SelectTrigger className="h-8 text-xs">
              <SelectValue placeholder="Select second request" />
            </SelectTrigger>
            <SelectContent>
              {items.map((item) => (
                <SelectItem key={item.id} value={item.id}>
                  <div className="flex items-center gap-2">
                    <Badge variant="outline" className="text-[10px]">
                      {item.method}
                    </Badge>
                    <span className="truncate">{item.path}</span>
                  </div>
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
        </div>
      </div>

      {/* Comparison content */}
      <div className="flex-1 overflow-hidden">
        {!leftItem || !rightItem ? (
          <div className="flex items-center justify-center h-full text-gray-400">
            <p className="text-sm">Select two requests to compare</p>
          </div>
        ) : mode === 'split' ? (
          <div className="flex h-full">
            {/* Left panel */}
            <div className="flex-1 flex flex-col border-r border-gray-200">
              <div className="flex items-center justify-between px-3 py-1.5 bg-gray-100 border-b border-gray-200">
                <span className="text-xs font-medium text-gray-600">Request A</span>
                <Button
                  variant="ghost"
                  size="sm"
                  onClick={() => handleCopy(getRequestContent(leftItem, view), 'left')}
                  className="h-6 text-xs"
                >
                  {copied === 'left' ? (
                    <Check className="w-3 h-3 mr-1" />
                  ) : (
                    <Copy className="w-3 h-3 mr-1" />
                  )}
                  {copied === 'left' ? 'Copied' : 'Copy'}
                </Button>
              </div>
              <ScrollArea className="flex-1">
                <pre className="p-3 text-xs font-mono whitespace-pre-wrap">
                  {getRequestContent(leftItem, view)}
                </pre>
              </ScrollArea>
            </div>

            {/* Right panel */}
            <div className="flex-1 flex flex-col">
              <div className="flex items-center justify-between px-3 py-1.5 bg-gray-100 border-b border-gray-200">
                <span className="text-xs font-medium text-gray-600">Request B</span>
                <Button
                  variant="ghost"
                  size="sm"
                  onClick={() => handleCopy(getRequestContent(rightItem, view), 'right')}
                  className="h-6 text-xs"
                >
                  {copied === 'right' ? (
                    <Check className="w-3 h-3 mr-1" />
                  ) : (
                    <Copy className="w-3 h-3 mr-1" />
                  )}
                  {copied === 'right' ? 'Copied' : 'Copy'}
                </Button>
              </div>
              <ScrollArea className="flex-1">
                <pre className="p-3 text-xs font-mono whitespace-pre-wrap">
                  {getRequestContent(rightItem, view)}
                </pre>
              </ScrollArea>
            </div>
          </div>
        ) : (
          // Unified diff view
          <ScrollArea className="h-full">
            <div className="p-3">
              {unifiedDiff?.map((line, index) => (
                <div
                  key={index}
                  className={cn(
                    'font-mono text-xs leading-5 px-2',
                    line.type === 'equal' && 'text-gray-700',
                    line.type === 'added' && 'bg-green-50 text-green-800',
                    line.type === 'removed' && 'bg-red-50 text-red-800 line-through',
                    line.type === 'modified' && 'bg-yellow-50 text-yellow-800'
                  )}
                >
                  <span className="inline-block w-6 text-gray-400 select-none">
                    {line.type === 'added' && '+'}
                    {line.type === 'removed' && '-'}
                    {line.type === 'modified' && '~'}
                    {line.type === 'equal' && ' '}
                  </span>
                  {line.leftLine || line.rightLine}
                </div>
              ))}
            </div>
          </ScrollArea>
        )}
      </div>
    </div>
  )
}

export default RequestCompare
