import { useState, useCallback } from 'react'
import {
  Save,
  X,
  Play,
  AlertCircle,
  CheckCircle,
  Wand2,
} from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Badge } from '@/components/ui/badge'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { Textarea } from '@/components/ui/textarea'
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/ui/card'
import { cn } from '@/lib/utils'

interface RuleConfig {
  id: string
  name: string
  description: string
  severity: 'critical' | 'high' | 'medium' | 'low' | 'info'
  enabled: boolean
  priority: number
  tags: string[]
  patterns: PatternConfig[]
  condition: 'or' | 'and'
  remediation: string
}

interface PatternConfig {
  type: 'regex' | 'keyword' | 'header' | 'status' | 'size'
  target: 'request' | 'response' | 'url' | 'body' | 'header'
  pattern: string
  description?: string
}

interface RuleEditorProps {
  initialRule?: Partial<RuleConfig>
  onSave?: (rule: RuleConfig) => Promise<void>
  onTest?: (rule: RuleConfig, testData: string) => Promise<boolean>
  onCancel?: () => void
}

const severityOptions = [
  { value: 'critical', label: 'Critical', color: 'bg-red-500' },
  { value: 'high', label: 'High', color: 'bg-orange-500' },
  { value: 'medium', label: 'Medium', color: 'bg-yellow-500' },
  { value: 'low', label: 'Low', color: 'bg-blue-500' },
  { value: 'info', label: 'Info', color: 'bg-gray-500' },
]

const patternTypes = [
  { value: 'regex', label: 'Regular Expression' },
  { value: 'keyword', label: 'Keyword Match' },
  { value: 'header', label: 'Header Check' },
  { value: 'status', label: 'Status Code' },
  { value: 'size', label: 'Response Size' },
]

const targetOptions = [
  { value: 'request', label: 'Full Request' },
  { value: 'response', label: 'Full Response' },
  { value: 'url', label: 'URL' },
  { value: 'body', label: 'Body Only' },
  { value: 'header', label: 'Headers Only' },
]

const defaultRule: RuleConfig = {
  id: '',
  name: '',
  description: '',
  severity: 'medium',
  enabled: true,
  priority: 50,
  tags: [],
  patterns: [{ type: 'regex', target: 'response', pattern: '' }],
  condition: 'or',
  remediation: '',
}

export function RuleEditor({ initialRule, onSave, onTest, onCancel }: RuleEditorProps) {
  const [rule, setRule] = useState<RuleConfig>({ ...defaultRule, ...initialRule })
  const [newTag, setNewTag] = useState('')
  const [testData, setTestData] = useState('')
  const [testResult, setTestResult] = useState<{ success: boolean; message: string } | null>(null)
  const [isSaving, setIsSaving] = useState(false)
  const [isValidating, setIsValidating] = useState(false)
  const [validationError, setValidationError] = useState<string | null>(null)

  const updateRule = useCallback((updates: Partial<RuleConfig>) => {
    setRule((prev) => ({ ...prev, ...updates }))
    setValidationError(null)
  }, [])

  const addTag = useCallback(() => {
    if (newTag.trim() && !rule.tags.includes(newTag.trim())) {
      updateRule({ tags: [...rule.tags, newTag.trim()] })
      setNewTag('')
    }
  }, [newTag, rule.tags, updateRule])

  const removeTag = useCallback((tag: string) => {
    updateRule({ tags: rule.tags.filter((t) => t !== tag) })
  }, [rule.tags, updateRule])

  const addPattern = useCallback(() => {
    updateRule({
      patterns: [...rule.patterns, { type: 'regex', target: 'response', pattern: '' }],
    })
  }, [rule.patterns, updateRule])

  const updatePattern = useCallback((index: number, updates: Partial<PatternConfig>) => {
    updateRule({
      patterns: rule.patterns.map((p, i) => (i === index ? { ...p, ...updates } : p)),
    })
  }, [rule.patterns, updateRule])

  const removePattern = useCallback((index: number) => {
    if (rule.patterns.length > 1) {
      updateRule({
        patterns: rule.patterns.filter((_, i) => i !== index),
      })
    }
  }, [rule.patterns, updateRule])

  const validateRegex = useCallback((pattern: string): boolean => {
    try {
      new RegExp(pattern)
      return true
    } catch {
      return false
    }
  }, [])

  const handleTest = useCallback(async () => {
    if (!onTest || !testData) return

    setIsValidating(true)
    setTestResult(null)

    try {
      const success = await onTest(rule, testData)
      setTestResult({
        success,
        message: success ? 'Pattern matched!' : 'No match found',
      })
    } catch (error) {
      setTestResult({
        success: false,
        message: `Test failed: ${error}`,
      })
    } finally {
      setIsValidating(false)
    }
  }, [onTest, rule, testData])

  const handleSave = useCallback(async () => {
    // Validate
    if (!rule.id || !rule.name) {
      setValidationError('Rule ID and Name are required')
      return
    }

    for (const pattern of rule.patterns) {
      if (!pattern.pattern) {
        setValidationError('All patterns must have a pattern value')
        return
      }
      if (pattern.type === 'regex' && !validateRegex(pattern.pattern)) {
        setValidationError(`Invalid regex pattern: ${pattern.pattern}`)
        return
      }
    }

    setIsSaving(true)
    try {
      await onSave?.(rule)
    } catch (error) {
      setValidationError(`Failed to save: ${error}`)
    } finally {
      setIsSaving(false)
    }
  }, [rule, onSave, validateRegex])

  const generateId = useCallback(() => {
    const id = rule.name
      .toLowerCase()
      .replace(/[^a-z0-9]+/g, '-')
      .replace(/^-|-$/g, '')
    updateRule({ id: `custom-${id}` })
  }, [rule.name, updateRule])

  return (
    <div className="flex flex-col h-full bg-gray-50">
      {/* Header */}
      <div className="flex items-center justify-between px-4 py-3 bg-white border-b border-gray-200">
        <h2 className="text-lg font-semibold text-gray-800">
          {initialRule ? 'Edit Rule' : 'Create Custom Rule'}
        </h2>
        <div className="flex items-center gap-2">
          <Button variant="outline" size="sm" onClick={onCancel} className="h-8">
            <X className="w-4 h-4 mr-1" />
            Cancel
          </Button>
          <Button size="sm" onClick={handleSave} disabled={isSaving} className="h-8">
            <Save className="w-4 h-4 mr-1" />
            {isSaving ? 'Saving...' : 'Save Rule'}
          </Button>
        </div>
      </div>

      {/* Error alert */}
      {validationError && (
        <div className="px-4 py-2 bg-red-50 border-b border-red-200 flex items-center gap-2">
          <AlertCircle className="w-4 h-4 text-red-500" />
          <span className="text-sm text-red-700">{validationError}</span>
        </div>
      )}

      {/* Content */}
      <div className="flex-1 overflow-auto p-4">
        <div className="max-w-3xl mx-auto space-y-6">
          {/* Basic Info */}
          <Card>
            <CardHeader className="py-3">
              <CardTitle className="text-sm font-medium">Basic Information</CardTitle>
              <CardDescription className="text-xs">
                Define the rule's identity and severity
              </CardDescription>
            </CardHeader>
            <CardContent className="space-y-4">
              <div className="grid grid-cols-2 gap-4">
                <div className="space-y-2">
                  <label className="text-xs font-medium text-gray-700">Rule ID</label>
                  <div className="flex gap-2">
                    <Input
                      value={rule.id}
                      onChange={(e) => updateRule({ id: e.target.value })}
                      placeholder="e.g., custom-sqli-detection"
                      className="h-8 text-xs font-mono"
                    />
                    <Button
                      variant="outline"
                      size="sm"
                      onClick={generateId}
                      disabled={!rule.name}
                      className="h-8"
                    >
                      <Wand2 className="w-3 h-3" />
                    </Button>
                  </div>
                </div>

                <div className="space-y-2">
                  <label className="text-xs font-medium text-gray-700">Name</label>
                  <Input
                    value={rule.name}
                    onChange={(e) => updateRule({ name: e.target.value })}
                    placeholder="Rule display name"
                    className="h-8 text-xs"
                  />
                </div>
              </div>

              <div className="space-y-2">
                <label className="text-xs font-medium text-gray-700">Description</label>
                <Textarea
                  value={rule.description}
                  onChange={(e) => updateRule({ description: e.target.value })}
                  placeholder="Describe what this rule detects"
                  className="text-xs min-h-[60px]"
                />
              </div>

              <div className="grid grid-cols-3 gap-4">
                <div className="space-y-2">
                  <label className="text-xs font-medium text-gray-700">Severity</label>
                  <Select
                    value={rule.severity}
                    onValueChange={(v) => updateRule({ severity: v as any })}
                  >
                    <SelectTrigger className="h-8 text-xs">
                      <SelectValue />
                    </SelectTrigger>
                    <SelectContent>
                      {severityOptions.map((opt) => (
                        <SelectItem key={opt.value} value={opt.value}>
                          <div className="flex items-center gap-2">
                            <div className={cn('w-2 h-2 rounded-full', opt.color)} />
                            {opt.label}
                          </div>
                        </SelectItem>
                      ))}
                    </SelectContent>
                  </Select>
                </div>

                <div className="space-y-2">
                  <label className="text-xs font-medium text-gray-700">Priority</label>
                  <Input
                    type="number"
                    value={rule.priority}
                    onChange={(e) => updateRule({ priority: parseInt(e.target.value) || 50 })}
                    className="h-8 text-xs"
                    min={1}
                    max={100}
                  />
                </div>

                <div className="space-y-2">
                  <label className="text-xs font-medium text-gray-700">Condition</label>
                  <Select
                    value={rule.condition}
                    onValueChange={(v) => updateRule({ condition: v as 'or' | 'and' })}
                  >
                    <SelectTrigger className="h-8 text-xs">
                      <SelectValue />
                    </SelectTrigger>
                    <SelectContent>
                      <SelectItem value="or">Match ANY pattern</SelectItem>
                      <SelectItem value="and">Match ALL patterns</SelectItem>
                    </SelectContent>
                  </Select>
                </div>
              </div>

              <div className="space-y-2">
                <label className="text-xs font-medium text-gray-700">Tags</label>
                <div className="flex items-center gap-2 flex-wrap">
                  {rule.tags.map((tag) => (
                    <Badge key={tag} variant="secondary" className="text-xs">
                      {tag}
                      <button
                        onClick={() => removeTag(tag)}
                        className="ml-1 hover:text-red-500"
                      >
                        <X className="w-3 h-3" />
                      </button>
                    </Badge>
                  ))}
                  <div className="flex items-center gap-1">
                    <Input
                      value={newTag}
                      onChange={(e) => setNewTag(e.target.value)}
                      onKeyDown={(e) => e.key === 'Enter' && addTag()}
                      placeholder="Add tag"
                      className="h-7 w-24 text-xs"
                    />
                    <Button variant="ghost" size="sm" onClick={addTag} className="h-7">
                      +
                    </Button>
                  </div>
                </div>
              </div>
            </CardContent>
          </Card>

          {/* Patterns */}
          <Card>
            <CardHeader className="py-3">
              <div className="flex items-center justify-between">
                <div>
                  <CardTitle className="text-sm font-medium">Detection Patterns</CardTitle>
                  <CardDescription className="text-xs">
                    Define patterns to match against traffic
                  </CardDescription>
                </div>
                <Button variant="outline" size="sm" onClick={addPattern} className="h-7 text-xs">
                  + Add Pattern
                </Button>
              </div>
            </CardHeader>
            <CardContent className="space-y-4">
              {rule.patterns.map((pattern, index) => (
                <div
                  key={index}
                  className="p-3 bg-gray-50 rounded-lg border border-gray-200 space-y-3"
                >
                  <div className="flex items-center justify-between">
                    <span className="text-xs font-medium text-gray-600">Pattern #{index + 1}</span>
                    {rule.patterns.length > 1 && (
                      <Button
                        variant="ghost"
                        size="sm"
                        onClick={() => removePattern(index)}
                        className="h-6 text-xs text-red-500"
                      >
                        Remove
                      </Button>
                    )}
                  </div>

                  <div className="grid grid-cols-2 gap-3">
                    <div className="space-y-1">
                      <label className="text-[10px] text-gray-500">Type</label>
                      <Select
                        value={pattern.type}
                        onValueChange={(v) => updatePattern(index, { type: v as any })}
                      >
                        <SelectTrigger className="h-7 text-xs">
                          <SelectValue />
                        </SelectTrigger>
                        <SelectContent>
                          {patternTypes.map((pt) => (
                            <SelectItem key={pt.value} value={pt.value}>
                              {pt.label}
                            </SelectItem>
                          ))}
                        </SelectContent>
                      </Select>
                    </div>

                    <div className="space-y-1">
                      <label className="text-[10px] text-gray-500">Target</label>
                      <Select
                        value={pattern.target}
                        onValueChange={(v) => updatePattern(index, { target: v as any })}
                      >
                        <SelectTrigger className="h-7 text-xs">
                          <SelectValue />
                        </SelectTrigger>
                        <SelectContent>
                          {targetOptions.map((pt) => (
                            <SelectItem key={pt.value} value={pt.value}>
                              {pt.label}
                            </SelectItem>
                          ))}
                        </SelectContent>
                      </Select>
                    </div>
                  </div>

                  <div className="space-y-1">
                    <label className="text-[10px] text-gray-500">Pattern</label>
                    <Input
                      value={pattern.pattern}
                      onChange={(e) => updatePattern(index, { pattern: e.target.value })}
                      placeholder={pattern.type === 'regex' ? 'Enter regex pattern' : 'Enter value'}
                      className={cn(
                        'h-8 text-xs font-mono',
                        pattern.type === 'regex' && !validateRegex(pattern.pattern) && pattern.pattern && 'border-red-300'
                      )}
                    />
                    {pattern.type === 'regex' && !validateRegex(pattern.pattern) && pattern.pattern && (
                      <p className="text-[10px] text-red-500">Invalid regex pattern</p>
                    )}
                  </div>
                </div>
              ))}
            </CardContent>
          </Card>

          {/* Remediation */}
          <Card>
            <CardHeader className="py-3">
              <CardTitle className="text-sm font-medium">Remediation</CardTitle>
              <CardDescription className="text-xs">
                Provide guidance on how to fix the issue
              </CardDescription>
            </CardHeader>
            <CardContent>
              <Textarea
                value={rule.remediation}
                onChange={(e) => updateRule({ remediation: e.target.value })}
                placeholder="Enter remediation steps and recommendations"
                className="text-xs min-h-[80px]"
              />
            </CardContent>
          </Card>

          {/* Test */}
          <Card>
            <CardHeader className="py-3">
              <CardTitle className="text-sm font-medium">Test Rule</CardTitle>
              <CardDescription className="text-xs">
                Test your rule against sample data
              </CardDescription>
            </CardHeader>
            <CardContent className="space-y-3">
              <Textarea
                value={testData}
                onChange={(e) => setTestData(e.target.value)}
                placeholder="Paste sample request or response data to test"
                className="text-xs font-mono min-h-[100px]"
              />
              <div className="flex items-center gap-3">
                <Button
                  variant="outline"
                  size="sm"
                  onClick={handleTest}
                  disabled={isValidating || !testData}
                  className="h-8"
                >
                  <Play className="w-4 h-4 mr-1" />
                  Test
                </Button>
                {testResult && (
                  <div
                    className={cn(
                      'flex items-center gap-1 text-xs',
                      testResult.success ? 'text-green-600' : 'text-red-600'
                    )}
                  >
                    {testResult.success ? (
                      <CheckCircle className="w-4 h-4" />
                    ) : (
                      <AlertCircle className="w-4 h-4" />
                    )}
                    {testResult.message}
                  </div>
                )}
              </div>
            </CardContent>
          </Card>
        </div>
      </div>
    </div>
  )
}

export default RuleEditor
