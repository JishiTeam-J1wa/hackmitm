import * as monaco from 'monaco-editor'

/**
 * Register custom HTTP languages for Monaco Editor
 * Provides syntax highlighting for HTTP request and response messages
 */

// Token types for HTTP messages
export const HTTP_TOKEN_TYPES = {
  // Request line tokens
  METHOD: 'keyword.http.method',
  URL: 'string.http.url',
  VERSION: 'constant.http.version',

  // Response line tokens
  STATUS_CODE: 'constant.numeric.http.status',
  STATUS_TEXT: 'string.http.status.text',

  // Header tokens
  HEADER_NAME: 'type.http.header.name',
  HEADER_DELIMITER: 'delimiter.http.header',
  HEADER_VALUE: 'string.http.header.value',

  // Body tokens
  BODY: 'text.http.body',
} as const

/**
 * Tokenizer rules for HTTP request language
 */
const httpRequestLanguage: monaco.languages.IMonarchLanguage = {
  defaultToken: '',
  tokenPostfix: '.http',

  // HTTP methods
  keywords: ['GET', 'POST', 'PUT', 'DELETE', 'PATCH', 'HEAD', 'OPTIONS', 'CONNECT', 'TRACE'],

  // Common header names (for better highlighting)
  typeKeywords: [
    'Content-Type', 'Content-Length', 'Authorization', 'Accept', 'Accept-Encoding',
    'Accept-Language', 'Cache-Control', 'Cookie', 'Host', 'Origin', 'Referer',
    'User-Agent', 'X-Requested-With', 'X-Forwarded-For', 'Connection', 'Date',
    'ETag', 'Last-Modified', 'Location', 'Server', 'Set-Cookie', 'Transfer-Encoding'
  ],

  tokenizer: {
    root: [
      // Request line: METHOD /path HTTP/version
      [/^(GET|POST|PUT|DELETE|PATCH|HEAD|OPTIONS|CONNECT|TRACE)\b/, 'keyword.http.method'],
      [/^(HTTP\/[\d.]+)\b/, 'constant.http.version'],

      // Headers section
      [/^([A-Za-z][-A-Za-z0-9]*)\s*:/, ['type.http.header.name', 'delimiter.http.header']],

      // Continue header values on subsequent lines (starting with space or tab)
      [/^[ \t]+.*/, 'string.http.header.value'],

      // Empty line marks end of headers
      [/^\s*$/, { token: 'delimiter', next: '@body' }],
    ],

    body: [
      // Body content - delegate to nested language based on content-type
      [/.*/, 'text.http.body'],
    ],
  },
}

/**
 * Tokenizer rules for HTTP response language
 */
const httpResponseLanguage: monaco.languages.IMonarchLanguage = {
  defaultToken: '',
  tokenPostfix: '.http',

  // Common header names
  typeKeywords: [
    'Content-Type', 'Content-Length', 'Content-Encoding', 'Cache-Control',
    'Connection', 'Date', 'ETag', 'Expires', 'Last-Modified', 'Location',
    'Server', 'Set-Cookie', 'Transfer-Encoding', 'Vary', 'WWW-Authenticate',
    'Access-Control-Allow-Origin', 'Access-Control-Allow-Methods',
    'Access-Control-Allow-Headers', 'X-Frame-Options', 'X-XSS-Protection'
  ],

  tokenizer: {
    root: [
      // Status line: HTTP/version STATUS_CODE STATUS_TEXT
      [/^(HTTP\/[\d.]+)\s+(\d{3})\s*(.*)?$/, [
        'constant.http.version',
        'constant.numeric.http.status',
        'string.http.status.text'
      ]],

      // Headers section
      [/^([A-Za-z][-A-Za-z0-9]*)\s*:/, ['type.http.header.name', 'delimiter.http.header']],

      // Continue header values on subsequent lines
      [/^[ \t]+.*/, 'string.http.header.value'],

      // Empty line marks end of headers
      [/^\s*$/, { token: 'delimiter', next: '@body' }],
    ],

    body: [
      // Body content
      [/.*/, 'text.http.body'],
    ],
  },
}

/**
 * Configuration for HTTP request language
 */
const httpRequestConfig: monaco.languages.LanguageConfiguration = {
  comments: {
    lineComment: '#',
  },
  brackets: [
    ['{', '}'],
    ['[', ']'],
    ['(', ')'],
  ],
  autoClosingPairs: [
    { open: '{', close: '}' },
    { open: '[', close: ']' },
    { open: '(', close: ')' },
    { open: '"', close: '"' },
    { open: "'", close: "'" },
  ],
  surroundingPairs: [
    { open: '{', close: '}' },
    { open: '[', close: ']' },
    { open: '(', close: ')' },
    { open: '"', close: '"' },
    { open: "'", close: "'" },
  ],
}

/**
 * Configuration for HTTP response language
 */
const httpResponseConfig: monaco.languages.LanguageConfiguration = {
  ...httpRequestConfig,
}

/**
 * Register HTTP languages with Monaco
 * Should be called once during app initialization
 */
export function registerHTTPLanguages(monacoInstance: typeof monaco): void {
  // Register http-request language
  monacoInstance.languages.register({ id: 'http-request' })
  monacoInstance.languages.setMonarchTokensProvider('http-request', httpRequestLanguage)
  monacoInstance.languages.setLanguageConfiguration('http-request', httpRequestConfig)

  // Register http-response language
  monacoInstance.languages.register({ id: 'http-response' })
  monacoInstance.languages.setMonarchTokensProvider('http-response', httpResponseLanguage)
  monacoInstance.languages.setLanguageConfiguration('http-response', httpResponseConfig)

  console.log('HTTP languages registered with Monaco Editor')
}

/**
 * Custom token colors for dark theme
 */
export const httpDarkThemeColors: Record<string, string> = {
  'keyword.http.method': '#569CD6',      // Blue for methods
  'string.http.url': '#CE9178',           // Orange/brown for URLs
  'constant.http.version': '#4EC9B0',     // Teal for HTTP version
  'constant.numeric.http.status': '#B5CEA8', // Light green for status codes
  'string.http.status.text': '#CE9178',   // Orange for status text
  'type.http.header.name': '#9CDCFE',     // Light blue for header names
  'delimiter.http.header': '#D4D4D4',     // Gray for colon delimiter
  'string.http.header.value': '#CE9178',  // Orange for header values
  'text.http.body': '#D4D4D4',            // Gray for body text
}

/**
 * Custom token colors for light theme
 */
export const httpLightThemeColors: Record<string, string> = {
  'keyword.http.method': '#0000FF',       // Blue for methods
  'string.http.url': '#A31515',           // Dark red for URLs
  'constant.http.version': '#267F99',     // Teal for HTTP version
  'constant.numeric.http.status': '#098658', // Green for status codes
  'string.http.status.text': '#A31515',   // Dark red for status text
  'type.http.header.name': '#001080',     // Dark blue for header names
  'delimiter.http.header': '#000000',     // Black for colon delimiter
  'string.http.header.value': '#A31515',  // Dark red for header values
  'text.http.body': '#000000',            // Black for body text
}

/**
 * Define custom theme with HTTP token colors
 */
export function defineHTTPTheme(
  monacoInstance: typeof monaco,
  themeName: string,
  baseTheme: 'vs' | 'vs-dark' | 'hc-black' = 'vs-dark'
): void {
  const colors = baseTheme === 'vs' ? httpLightThemeColors : httpDarkThemeColors

  monacoInstance.editor.defineTheme(themeName, {
    base: baseTheme,
    inherit: true,
    rules: Object.entries(colors).map(([token, foreground]) => ({
      token,
      foreground,
    })),
    colors: {},
  })
}

/**
 * Initialize all HTTP language support for Monaco
 * Call this once when the app loads
 */
export function initializeHTTPLanguageSupport(monacoInstance: typeof monaco): void {
  // Register languages
  registerHTTPLanguages(monacoInstance)

  // Define custom themes
  defineHTTPTheme(monacoInstance, 'http-dark', 'vs-dark')
  defineHTTPTheme(monacoInstance, 'http-light', 'vs')

  console.log('HTTP language support initialized')
}
