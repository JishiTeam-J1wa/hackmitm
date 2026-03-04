import Editor, { OnMount } from '@monaco-editor/react'
import { cn } from '@/lib/utils'

interface HttpMessageEditorProps {
  content: string
  headers?: Record<string, string>
  contentType?: string
  readOnly?: boolean
  onChange?: (content: string) => void
  className?: string
}

export function HttpMessageEditor({
  content,
  headers,
  readOnly = true,
  onChange,
  className,
}: HttpMessageEditorProps) {
  // Detect content type from headers or content
  const detectLanguage = (): string => {
    // Check contentType prop first
    if (headers?.['content-type']) {
      const ct = headers['content-type'].toLowerCase()
      if (ct.includes('json')) return 'json'
      if (ct.includes('html')) return 'html'
      if (ct.includes('xml')) return 'xml'
      if (ct.includes('javascript')) return 'javascript'
    }

    // Try to detect from content
    const trimmed = content.trim()
    if (trimmed.startsWith('{') || trimmed.startsWith('[')) return 'json'
    if (trimmed.startsWith('<!DOCTYPE') || trimmed.startsWith('<html')) return 'html'
    if (trimmed.startsWith('<')) return 'xml'

    return 'plaintext'
  }

  const handleEditorMount: OnMount = () => {
    // Editor mounted
  }

  return (
    <div className={cn('h-full', className)}>
      <Editor
        height="100%"
        language={detectLanguage()}
        value={content}
        theme="vs-dark"
        onMount={handleEditorMount}
        onChange={(value) => !readOnly && onChange?.(value || '')}
        options={{
          readOnly,
          minimap: { enabled: false },
          fontSize: 12,
          fontFamily: 'JetBrains Mono, Fira Code, Menlo, Monaco, Consolas, monospace',
          lineNumbers: 'on',
          wordWrap: 'on',
          scrollBeyondLastLine: false,
          automaticLayout: true,
          tabSize: 2,
          renderWhitespace: 'selection',
          contextmenu: true,
          copyWithSyntaxHighlighting: true,
          scrollbar: {
            vertical: 'auto',
            horizontal: 'auto',
            verticalScrollbarSize: 10,
            horizontalScrollbarSize: 10,
          },
        }}
      />
    </div>
  )
}
