import { useRef, useCallback } from 'react'
import Editor, { OnMount, OnChange } from '@monaco-editor/react'
import type { editor } from 'monaco-editor'
import { initializeHTTPLanguageSupport } from '@/lib/monaco-http-language'

/**
 * HttpEditor Props
 */
export interface HttpEditorProps {
  /** Editor content value */
  value: string
  /** Callback when content changes */
  onChange?: (value: string) => void
  /** Language mode: 'http-request' or 'http-response' */
  language?: 'http-request' | 'http-response' | 'json' | 'xml' | 'html' | 'plaintext'
  /** Whether editor is read-only */
  readOnly?: boolean
  /** Placeholder text when empty */
  placeholder?: string
  /** Additional CSS class names */
  className?: string
  /** Editor height */
  height?: string | number
  /** Word wrap enabled */
  wordWrap?: boolean
  /** Show minimap */
  minimap?: boolean
  /** Line numbers */
  lineNumbers?: 'on' | 'off' | 'relative' | 'interval'
  /** Font size */
  fontSize?: number
  /** Auto-detect content type for syntax highlighting */
  autoDetectContentType?: string
}

// Track if HTTP languages have been registered
let httpLanguagesRegistered = false

/**
 * HttpEditor - Monaco Editor wrapper for HTTP content editing
 *
 * Provides syntax highlighting for HTTP request/response messages,
 * JSON, XML, and other common formats.
 */
export function HttpEditor({
  value,
  onChange,
  language = 'plaintext',
  readOnly = false,
  placeholder,
  className = '',
  height = '100%',
  wordWrap = true,
  minimap = false,
  lineNumbers = 'on',
  fontSize = 12,
  autoDetectContentType,
}: HttpEditorProps) {
  const editorRef = useRef<editor.IStandaloneCodeEditor | null>(null)

  /**
   * Handle editor mount
   */
  const handleEditorMount: OnMount = useCallback((editor, monaco) => {
    editorRef.current = editor

    // Initialize HTTP language support once
    if (!httpLanguagesRegistered) {
      initializeHTTPLanguageSupport(monaco)
      httpLanguagesRegistered = true
    }

    // Focus editor
    editor.focus()
  }, [])

  /**
   * Handle content changes
   */
  const handleChange: OnChange = useCallback((value) => {
    onChange?.(value || '')
  }, [onChange])

  /**
   * Detect language from content type
   */
  const detectLanguage = useCallback((): string => {
    if (autoDetectContentType) {
      const ct = autoDetectContentType.toLowerCase()
      if (ct.includes('json')) return 'json'
      if (ct.includes('xml') || ct.includes('html')) return 'html'
      if (ct.includes('javascript')) return 'javascript'
    }

    // Try to auto-detect from content
    const trimmedValue = value.trim()
    if (trimmedValue.startsWith('{') || trimmedValue.startsWith('[')) return 'json'
    if (trimmedValue.startsWith('<')) return 'html'

    return language
  }, [autoDetectContentType, language, value])

  /**
   * Get effective language for Monaco
   */
  const effectiveLanguage = language.startsWith('http') ? language : detectLanguage()

  return (
    <div className={`http-editor ${className}`} style={{ height }}>
      <Editor
        height={height}
        language={effectiveLanguage}
        value={value}
        onChange={handleChange}
        onMount={handleEditorMount}
        theme="vs-dark"
        options={{
          readOnly,
          wordWrap: wordWrap ? 'on' : 'off',
          minimap: { enabled: minimap },
          lineNumbers,
          fontSize,
          fontFamily: "'Fira Code', 'Monaco', 'Menlo', monospace",
          fontLigatures: true,
          scrollBeyondLastLine: false,
          automaticLayout: true,
          tabSize: 2,
          insertSpaces: true,
          renderWhitespace: 'selection',
          contextmenu: true,
          copyWithSyntaxHighlighting: true,
          links: false,
          colorDecorators: true,
          accessibilitySupport: 'auto',
          cursorBlinking: 'smooth',
          cursorSmoothCaretAnimation: 'on',
          smoothScrolling: true,
          padding: { top: 8, bottom: 8 },
        }}
      />
      {!value && placeholder && (
        <div
          className="absolute top-2 left-3 text-muted-foreground pointer-events-none text-xs"
          style={{ fontFamily: 'inherit' }}
        >
          {placeholder}
        </div>
      )}
    </div>
  )
}

export default HttpEditor
