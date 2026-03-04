import { useEffect, useState, useRef } from 'react'
import { motion, AnimatePresence } from 'framer-motion'

interface InitScreenProps {
  onComplete: () => void
  duration?: number
}

export function InitScreen({ onComplete, duration = 2500 }: InitScreenProps) {
  const [progress, setProgress] = useState(0)
  const [loadingText, setLoadingText] = useState('正在初始化...')
  const onCompleteRef = useRef(onComplete)

  // Keep the ref updated
  useEffect(() => {
    onCompleteRef.current = onComplete
  }, [onComplete])

  useEffect(() => {
    const loadingTexts = [
      '正在初始化...',
      '加载配置文件...',
      '检查数据库连接...',
      '准备界面组件...',
      '启动完成...',
    ]

    const progressInterval = setInterval(() => {
      setProgress((prev) => {
        const next = prev + 2
        if (next >= 100) {
          clearInterval(progressInterval)
          return 100
        }

        // Update loading text based on progress
        const textIndex = Math.min(Math.floor(next / 25), loadingTexts.length - 1)
        setLoadingText(loadingTexts[textIndex])

        return next
      })
    }, duration / 50)

    const timer = setTimeout(() => {
      onCompleteRef.current()
    }, duration)

    return () => {
      clearTimeout(timer)
      clearInterval(progressInterval)
    }
  }, [duration])

  return (
    <AnimatePresence>
      <motion.div
        className="fixed inset-0 z-50 flex flex-col items-center justify-center bg-white"
        initial={{ opacity: 0 }}
        animate={{ opacity: 1 }}
        exit={{ opacity: 0, transition: { duration: 0.5 } }}
      >
        {/* Logo Container */}
        <motion.div
          className="relative mb-8"
          initial={{ scale: 0.5, opacity: 0 }}
          animate={{ scale: 1, opacity: 1 }}
          transition={{ duration: 0.6, ease: 'easeOut' }}
        >
          {/* Glow Effect */}
          <motion.div
            className="absolute inset-0 blur-3xl opacity-30"
            style={{
              background: 'radial-gradient(circle, #f97316 0%, transparent 70%)',
            }}
            animate={{
              scale: [1, 1.2, 1],
              opacity: [0.3, 0.5, 0.3],
            }}
            transition={{
              duration: 2,
              repeat: Infinity,
              ease: 'easeInOut',
            }}
          />

          {/* Logo */}
          <motion.img
            src="/logo.png"
            alt="击势安全团队"
            className="w-48 h-48 object-contain relative z-10"
            animate={{
              y: [0, -10, 0],
            }}
            transition={{
              duration: 3,
              repeat: Infinity,
              ease: 'easeInOut',
            }}
          />
        </motion.div>

        {/* Title */}
        <motion.div
          className="text-center mb-8"
          initial={{ y: 20, opacity: 0 }}
          animate={{ y: 0, opacity: 1 }}
          transition={{ delay: 0.3, duration: 0.5 }}
        >
          <h1 className="text-3xl font-bold bg-gradient-to-r from-orange-500 to-orange-600 bg-clip-text text-transparent mb-2">
            HackMITM
          </h1>
          <p className="text-gray-500 text-sm">击势安全团队 · 高性能代理工具</p>
        </motion.div>

        {/* Progress Bar */}
        <motion.div
          className="w-64 mb-4"
          initial={{ width: 0, opacity: 0 }}
          animate={{ width: 256, opacity: 1 }}
          transition={{ delay: 0.5, duration: 0.5 }}
        >
          <div className="h-1 bg-gray-100 rounded-full overflow-hidden">
            <motion.div
              className="h-full bg-gradient-to-r from-orange-400 to-orange-500 rounded-full"
              style={{ width: `${progress}%` }}
              transition={{ duration: 0.1 }}
            />
          </div>
        </motion.div>

        {/* Loading Text */}
        <motion.p
          className="text-gray-400 text-sm"
          initial={{ opacity: 0 }}
          animate={{ opacity: 1 }}
          transition={{ delay: 0.6 }}
        >
          {loadingText}
        </motion.p>

        {/* Version Info */}
        <motion.div
          className="absolute bottom-8 text-center"
          initial={{ opacity: 0 }}
          animate={{ opacity: 1 }}
          transition={{ delay: 0.8 }}
        >
          <p className="text-gray-300 text-xs">Version 1.0.0</p>
          <p className="text-gray-300 text-xs mt-1">© 2024 击势安全团队</p>
        </motion.div>
      </motion.div>
    </AnimatePresence>
  )
}
