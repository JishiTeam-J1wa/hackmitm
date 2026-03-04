import { useEffect, useState } from 'react'
import { createPortal } from 'react-dom'
import { motion, AnimatePresence } from 'framer-motion'
import { CheckCircle, AlertCircle, Info, X, AlertTriangle } from 'lucide-react'
import { cn } from '@/lib/utils'

export type ToastType = 'success' | 'error' | 'info' | 'warning'

interface Toast {
  id: string
  type: ToastType
  message: string
  duration?: number
}

interface ToastProps {
  toast: Toast
  onClose: (id: string) => void
}

const icons = {
  success: CheckCircle,
  error: AlertCircle,
  info: Info,
  warning: AlertTriangle,
}

const colors = {
  success: 'bg-green-50 border-green-200 text-green-800',
  error: 'bg-red-50 border-red-200 text-red-800',
  info: 'bg-blue-50 border-blue-200 text-blue-800',
  warning: 'bg-yellow-50 border-yellow-200 text-yellow-800',
}

const iconColors = {
  success: 'text-green-500',
  error: 'text-red-500',
  info: 'text-blue-500',
  warning: 'text-yellow-500',
}

function ToastItem({ toast, onClose }: ToastProps) {
  const Icon = icons[toast.type]

  useEffect(() => {
    const timer = setTimeout(() => {
      onClose(toast.id)
    }, toast.duration || 3000)

    return () => clearTimeout(timer)
  }, [toast.id, toast.duration, onClose])

  return (
    <motion.div
      initial={{ opacity: 0, y: -10, scale: 0.95 }}
      animate={{ opacity: 1, y: 0, scale: 1 }}
      exit={{ opacity: 0, y: -10, scale: 0.95 }}
      transition={{ duration: 0.2 }}
      className={cn(
        'flex items-center gap-2 px-3 py-2 rounded-lg border shadow-lg text-sm',
        colors[toast.type]
      )}
    >
      <Icon className={cn('w-4 h-4 flex-shrink-0', iconColors[toast.type])} />
      <span className="flex-1">{toast.message}</span>
      <button
        onClick={() => onClose(toast.id)}
        className="p-0.5 hover:opacity-70 transition-opacity"
      >
        <X className="w-3.5 h-3.5" />
      </button>
    </motion.div>
  )
}

// Toast Container
interface ToastContainerProps {
  toasts: Toast[]
  onClose: (id: string) => void
}

export function ToastContainer({ toasts, onClose }: ToastContainerProps) {
  if (toasts.length === 0) return null

  return createPortal(
    <div className="fixed top-4 right-4 z-[10000] flex flex-col gap-2 max-w-sm">
      <AnimatePresence mode="popLayout">
        {toasts.map((toast) => (
          <ToastItem key={toast.id} toast={toast} onClose={onClose} />
        ))}
      </AnimatePresence>
    </div>,
    document.body
  )
}

// Toast Store
let toastId = 0
let listeners: Array<(toasts: Toast[]) => void> = []
let currentToasts: Toast[] = []

function notifyListeners() {
  listeners.forEach((listener) => listener([...currentToasts]))
}

export function useToast() {
  const [toasts, setToasts] = useState<Toast[]>(currentToasts)

  useEffect(() => {
    listeners.push(setToasts)
    return () => {
      listeners = listeners.filter((l) => l !== setToasts)
    }
  }, [])

  const addToast = (type: ToastType, message: string, duration?: number) => {
    const id = `toast-${++toastId}`
    currentToasts = [...currentToasts, { id, type, message, duration }]
    notifyListeners()
  }

  const removeToast = (id: string) => {
    currentToasts = currentToasts.filter((t) => t.id !== id)
    notifyListeners()
  }

  return {
    toasts,
    addToast,
    removeToast,
    success: (msg: string) => addToast('success', msg),
    error: (msg: string) => addToast('error', msg),
    info: (msg: string) => addToast('info', msg),
    warning: (msg: string) => addToast('warning', msg),
  }
}

// 全局 toast 函数
export const toast = {
  success: (message: string) => {
    const id = `toast-${++toastId}`
    currentToasts = [...currentToasts, { id, type: 'success', message }]
    notifyListeners()
  },
  error: (message: string) => {
    const id = `toast-${++toastId}`
    currentToasts = [...currentToasts, { id, type: 'error', message }]
    notifyListeners()
  },
  info: (message: string) => {
    const id = `toast-${++toastId}`
    currentToasts = [...currentToasts, { id, type: 'info', message }]
    notifyListeners()
  },
  warning: (message: string) => {
    const id = `toast-${++toastId}`
    currentToasts = [...currentToasts, { id, type: 'warning', message }]
    notifyListeners()
  },
}

export default ToastContainer
