import { useState } from 'react'
import { Search, RefreshCw, ExternalLink } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Badge } from '@/components/ui/badge'
import { Card, CardContent } from '@/components/ui/card'
import { ScrollArea } from '@/components/ui/scroll-area'
import { useFingerprintStore } from '@/store'
import { IdentifyFingerprint } from '../../../wailsjs/go/main/App'
import { cn } from '@/lib/utils'
import type { FingerprintResult } from '@/types'

export function FingerprintTab() {
  const { results, selectedResult, selectResult, stats, addResult } = useFingerprintStore()
  const [url, setUrl] = useState('')
  const [loading, setLoading] = useState(false)

  const handleIdentify = async () => {
    if (!url) return

    setLoading(true)
    try {
      const result = await IdentifyFingerprint(url)
      if (result) {
        const fingerprintResult: FingerprintResult = {
          url: result.url,
          fingerprints: result.fingerprints || [],
          confidence: result.confidence || 0,
          processTime: result.processTime || 0,
          title: result.title || '',
          statusCode: result.statusCode || 0,
          timestamp: new Date().toISOString()
        }
        addResult(fingerprintResult)
      }
    } catch (error) {
      console.error('Identification failed:', error)
    } finally {
      setLoading(false)
    }
  }

  const getTechIcon = (tech: string): string => {
    const icons: Record<string, string> = {
      'nginx': '🌐',
      'apache': '🪶',
      'php': '🐘',
      'python': '🐍',
      'node': '💚',
      'react': '⚛️',
      'vue': '💚',
      'angular': '🅰️',
      'mysql': '🐬',
      'redis': '🔴',
      'mongodb': '🍃',
      'wordpress': '📝',
      'cloudflare': '☁️',
    }
    return icons[tech.toLowerCase()] || '🔧'
  }

  return (
    <div className="flex h-full">
      {/* Results list */}
      <div className="w-80 border-r border-border flex flex-col">
        {/* Search */}
        <div className="p-4 border-b border-border">
          <div className="relative">
            <Search className="absolute left-2 top-1/2 -translate-y-1/2 w-4 h-4 text-muted-foreground" />
            <Input
              placeholder="Enter URL to identify..."
              value={url}
              onChange={(e) => setUrl(e.target.value)}
              onKeyDown={(e) => e.key === 'Enter' && handleIdentify()}
              className="pl-8 pr-16"
            />
            <Button
              size="sm"
              className="absolute right-1 top-1/2 -translate-y-1/2 h-7"
              onClick={handleIdentify}
              disabled={loading || !url}
            >
              {loading ? (
                <RefreshCw className="w-4 h-4 animate-spin" />
              ) : (
                'Scan'
              )}
            </Button>
          </div>
        </div>

        {/* Stats */}
        <div className="flex items-center gap-4 px-4 py-2 border-b border-border text-sm">
          <span className="text-muted-foreground">
            Scans: <strong>{stats.totalScans}</strong>
          </span>
          <span className="text-muted-foreground">
            Techs: <strong>{stats.uniqueTechs}</strong>
          </span>
        </div>

        {/* Results list */}
        <ScrollArea className="flex-1">
          <div className="p-2">
            {results.map((result) => (
              <div
                key={result.url + result.timestamp}
                className={cn(
                  'p-3 rounded-lg cursor-pointer mb-2 border',
                  selectedResult?.url === result.url && selectedResult?.timestamp === result.timestamp
                    ? 'bg-primary/10 border-primary'
                    : 'bg-muted/30 hover:bg-muted/50 border-transparent'
                )}
                onClick={() => selectResult(result)}
              >
                <div className="flex items-center gap-2 mb-1">
                  <span className="font-mono text-sm truncate flex-1">
                    {result.url ? new URL(result.url).hostname : 'Unknown'}
                  </span>
                  <Badge variant="outline" className="text-xs">
                    {result.fingerprints.length}
                  </Badge>
                </div>
                <div className="flex items-center gap-1 text-xs text-muted-foreground">
                  <span>{result.statusCode}</span>
                  <span>•</span>
                  <span>{(result.confidence * 100).toFixed(0)}% conf</span>
                </div>
              </div>
            ))}

            {results.length === 0 && (
              <div className="text-center text-muted-foreground py-8">
                No fingerprints scanned yet
              </div>
            )}
          </div>
        </ScrollArea>
      </div>

      {/* Details panel */}
      <div className="flex-1 flex flex-col">
        {selectedResult ? (
          <>
            {/* Header */}
            <div className="px-6 py-4 border-b border-border">
              <div className="flex items-center gap-2 mb-2">
                <h2 className="text-lg font-semibold flex-1 truncate">
                  {selectedResult.title || (selectedResult.url ? new URL(selectedResult.url).hostname : 'Unknown')}
                </h2>
                <Button variant="ghost" size="sm">
                  <ExternalLink className="w-4 h-4" />
                </Button>
              </div>
              <div className="text-sm text-muted-foreground font-mono truncate">
                {selectedResult.url}
              </div>
              <div className="flex items-center gap-4 mt-2">
                <Badge variant={selectedResult.statusCode < 400 ? 'success' : 'destructive'}>
                  {selectedResult.statusCode}
                </Badge>
                <span className="text-sm text-muted-foreground">
                  Confidence: {(selectedResult.confidence * 100).toFixed(0)}%
                </span>
                <span className="text-sm text-muted-foreground">
                  Time: {selectedResult.processTime}ms
                </span>
              </div>
            </div>

            {/* Technologies */}
            <ScrollArea className="flex-1 p-6">
              <h3 className="text-sm font-medium mb-4">Detected Technologies</h3>
              <div className="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-4 gap-3">
                {selectedResult.fingerprints.map((tech, index) => (
                  <Card key={index} className="bg-muted/30">
                    <CardContent className="p-4 flex items-center gap-3">
                      <span className="text-2xl">{getTechIcon(tech)}</span>
                      <div>
                        <div className="font-medium text-sm">{tech}</div>
                        <div className="text-xs text-muted-foreground">Technology</div>
                      </div>
                    </CardContent>
                  </Card>
                ))}
              </div>

              {selectedResult.fingerprints.length === 0 && (
                <div className="text-center text-muted-foreground py-12">
                  No technologies detected
                </div>
              )}
            </ScrollArea>
          </>
        ) : (
          <div className="flex items-center justify-center h-full text-muted-foreground">
            Select a scan result to view details
          </div>
        )}
      </div>
    </div>
  )
}
