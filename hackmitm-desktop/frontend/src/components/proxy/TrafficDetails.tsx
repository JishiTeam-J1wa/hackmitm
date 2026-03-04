import { useState } from 'react'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { ScrollArea } from '@/components/ui/scroll-area'
import { Badge } from '@/components/ui/badge'
import { HttpMessageEditor } from './HttpMessageEditor'
import { InterceptActions } from './InterceptActions'
import { cn, getStatusCodeClass, formatBytes, formatDuration, headersToString } from '@/lib/utils'
import type { TrafficItem } from '@/types'

interface TrafficDetailsProps {
  item: TrafficItem
  isIntercepted?: boolean
}

export function TrafficDetails({ item, isIntercepted = false }: TrafficDetailsProps) {
  const [activeTab, setActiveTab] = useState('request')

  return (
    <div className="flex flex-col h-full">
      {/* Summary bar */}
      <div className="flex items-center gap-4 px-4 py-2 border-b border-border bg-muted/20">
        <Badge
          variant="outline"
          className={cn(
            'font-mono',
            item.method === 'GET' && 'text-blue-500',
            item.method === 'POST' && 'text-green-500',
            item.method === 'PUT' && 'text-yellow-500',
            item.method === 'DELETE' && 'text-red-500'
          )}
        >
          {item.method}
        </Badge>

        <span className="font-mono text-sm truncate flex-1">{item.url}</span>

        <Badge
          variant="outline"
          className={cn('font-mono', getStatusCodeClass(item.statusCode))}
        >
          {item.statusCode}
        </Badge>

        <span className="text-xs text-muted-foreground">
          {formatBytes(item.responseSize)} | {formatDuration(item.duration)}
        </span>

        {isIntercepted && (
          <Badge variant="destructive" className="animate-pulse">
            Intercepted
          </Badge>
        )}
      </div>

      {/* Intercept actions */}
      {isIntercepted && <InterceptActions item={item} />}

      {/* Tabs */}
      <Tabs value={activeTab} onValueChange={setActiveTab} className="flex-1 flex flex-col">
        <TabsList className="mx-4 mt-2">
          <TabsTrigger value="request">Request</TabsTrigger>
          <TabsTrigger value="response">Response</TabsTrigger>
          <TabsTrigger value="headers">Headers</TabsTrigger>
        </TabsList>

        <TabsContent value="request" className="flex-1 mt-2">
          <ScrollArea className="h-full">
            <HttpMessageEditor
              content={item.requestBody}
              headers={item.requestHeaders}
              readOnly={!isIntercepted}
              onChange={() => {
                // Handle edit for intercepted requests
              }}
            />
          </ScrollArea>
        </TabsContent>

        <TabsContent value="response" className="flex-1 mt-2">
          <ScrollArea className="h-full">
            <HttpMessageEditor
              content={item.responseBody}
              headers={item.responseHeaders}
              readOnly
            />
          </ScrollArea>
        </TabsContent>

        <TabsContent value="headers" className="flex-1 mt-2">
          <ScrollArea className="h-full p-4">
            <div className="space-y-4">
              <div>
                <h3 className="text-sm font-medium mb-2">Request Headers</h3>
                <pre className="text-xs font-mono bg-muted/30 p-3 rounded-md overflow-x-auto">
                  {headersToString(item.requestHeaders)}
                </pre>
              </div>

              <div>
                <h3 className="text-sm font-medium mb-2">Response Headers</h3>
                <pre className="text-xs font-mono bg-muted/30 p-3 rounded-md overflow-x-auto">
                  {headersToString(item.responseHeaders)}
                </pre>
              </div>
            </div>
          </ScrollArea>
        </TabsContent>
      </Tabs>
    </div>
  )
}
