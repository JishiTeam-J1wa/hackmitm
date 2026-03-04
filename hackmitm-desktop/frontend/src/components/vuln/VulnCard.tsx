import { Badge } from '@/components/ui/badge'
import { cn } from '@/lib/utils'
import type { Vulnerability } from '@/types'
import {
  AlertTriangle,
  AlertCircle,
  Info,
  CheckCircle,
  XCircle,
  Clock,
  Globe,
  Shield,
} from 'lucide-react'

interface VulnCardProps {
  vuln: Vulnerability
  isSelected: boolean
  onClick: () => void
  onContextMenu?: (e: React.MouseEvent) => void
}

const severityConfig = {
  critical: {
    icon: AlertTriangle,
    color: 'text-red-600',
    bgColor: 'bg-red-50 border-red-200',
    badgeColor: 'bg-red-500 text-white',
    label: 'Critical',
  },
  high: {
    icon: AlertCircle,
    color: 'text-orange-600',
    bgColor: 'bg-orange-50 border-orange-200',
    badgeColor: 'bg-orange-500 text-white',
    label: 'High',
  },
  medium: {
    icon: AlertCircle,
    color: 'text-yellow-600',
    bgColor: 'bg-yellow-50 border-yellow-200',
    badgeColor: 'bg-yellow-500 text-white',
    label: 'Medium',
  },
  low: {
    icon: Info,
    color: 'text-blue-600',
    bgColor: 'bg-blue-50 border-blue-200',
    badgeColor: 'bg-blue-500 text-white',
    label: 'Low',
  },
}

const statusConfig = {
  open: { icon: AlertCircle, color: 'text-orange-500', label: 'Open' },
  fixed: { icon: CheckCircle, color: 'text-green-500', label: 'Fixed' },
  ignored: { icon: XCircle, color: 'text-gray-400', label: 'Ignored' },
}

export function VulnCard({ vuln, isSelected, onClick, onContextMenu }: VulnCardProps) {
  const severity = severityConfig[vuln.severity]
  const status = statusConfig[vuln.status]
  const SeverityIcon = severity.icon
  const StatusIcon = status.icon

  return (
    <div
      onClick={onClick}
      onContextMenu={onContextMenu}
      className={cn(
        'p-3 border rounded-lg cursor-pointer transition-all duration-200',
        isSelected
          ? 'border-blue-500 bg-blue-50 shadow-sm'
          : 'border-gray-200 hover:border-gray-300 hover:shadow-sm',
        vuln.status === 'ignored' && 'opacity-60'
      )}
    >
      {/* Header */}
      <div className="flex items-start gap-2 mb-2">
        <SeverityIcon className={cn('w-5 h-5 mt-0.5 flex-shrink-0', severity.color)} />
        <div className="flex-1 min-w-0">
          <h4 className="text-sm font-medium text-gray-900 truncate">{vuln.title}</h4>
          <div className="flex items-center gap-2 mt-1">
            <Badge className={cn('text-[10px] px-1.5 py-0', severity.badgeColor)}>
              {severity.label}
            </Badge>
            <Badge variant="outline" className="text-[10px] px-1.5 py-0">
              {vuln.type}
            </Badge>
            {vuln.status !== 'open' && (
              <Badge
                variant="outline"
                className={cn('text-[10px] px-1.5 py-0', status.color)}
              >
                <StatusIcon className="w-3 h-3 mr-0.5" />
                {status.label}
              </Badge>
            )}
          </div>
        </div>
      </div>

      {/* URL */}
      <div className="flex items-center gap-1.5 text-xs text-gray-500 mb-2">
        <Globe className="w-3 h-3 flex-shrink-0" />
        <span className="truncate">{vuln.url}</span>
      </div>

      {/* Description */}
      <p className="text-xs text-gray-600 line-clamp-2 mb-2">{vuln.description}</p>

      {/* Footer */}
      <div className="flex items-center justify-between text-[10px] text-gray-400">
        <div className="flex items-center gap-1">
          <Shield className="w-3 h-3" />
          <span>{vuln.source === 'passive' ? 'Passive Scan' : 'Active Scan'}</span>
        </div>
        <div className="flex items-center gap-1">
          <Clock className="w-3 h-3" />
          <span>{new Date(vuln.createdAt).toLocaleDateString('zh-CN')}</span>
        </div>
      </div>

      {/* CVSS Score */}
      {vuln.cvss && (
        <div className="mt-2 pt-2 border-t border-gray-100">
          <div className="flex items-center justify-between">
            <span className="text-[10px] text-gray-500">CVSS Score</span>
            <div className="flex items-center gap-1">
              <div
                className={cn(
                  'w-8 h-2 rounded-full',
                  vuln.cvss >= 9 ? 'bg-red-500' :
                  vuln.cvss >= 7 ? 'bg-orange-500' :
                  vuln.cvss >= 4 ? 'bg-yellow-500' : 'bg-blue-500'
                )}
              />
              <span className={cn('text-xs font-medium', severity.color)}>
                {vuln.cvss.toFixed(1)}
              </span>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}
