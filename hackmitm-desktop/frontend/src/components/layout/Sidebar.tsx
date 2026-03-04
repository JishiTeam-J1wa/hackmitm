import { Target, RefreshCw, Fingerprint, LayoutDashboard, Settings, Circle, ArrowLeftRight } from 'lucide-react'
import { cn } from '@/lib/utils'
import { Button } from '@/components/ui/button'
import { Separator } from '@/components/ui/separator'
import { useProxyStore } from '@/store'
import type { TabId } from '@/types'

interface SidebarProps {
  activeTab: TabId
  onTabChange: (tab: TabId) => void
}

// Dashboard 保留英文，其他使用中文
const navItems: { id: TabId; label: string; icon: typeof LayoutDashboard }[] = [
  { id: 'dashboard', label: 'Dashboard', icon: LayoutDashboard },
  { id: 'proxy', label: '代理', icon: ArrowLeftRight },
  { id: 'target', label: '目标', icon: Target },
  { id: 'repeater', label: '重放', icon: RefreshCw },
  { id: 'fingerprint', label: '指纹', icon: Fingerprint },
]

export function Sidebar({ activeTab, onTabChange }: SidebarProps) {
  const { connected } = useProxyStore()

  return (
    <div className="sidebar">
      {/* Logo - 击势安全团队 */}
      <div className="sidebar-logo">
        <img
          src="/j1shi-logo.png"
          alt="击势安全团队"
          className="w-8 h-8"
        />
      </div>

      {/* Navigation */}
      <nav className="sidebar-nav">
        {navItems.map((item) => {
          const Icon = item.icon
          const isActive = activeTab === item.id
          return (
            <div key={item.id} className="flex flex-col items-center">
              <Button
                variant="ghost"
                size="icon"
                className={cn(
                  'w-10 h-10',
                  isActive && 'bg-blue-500/10 text-blue-600 hover:bg-blue-500/20'
                )}
                onClick={() => onTabChange(item.id)}
                title={item.label}
              >
                <Icon className="w-5 h-5" />
              </Button>
              <span className={cn(
                'text-[9px] mt-0.5',
                isActive ? 'text-blue-600 font-medium' : 'text-gray-500'
              )}>
                {item.label}
              </span>
            </div>
          )
        })}
      </nav>

      {/* Footer */}
      <div className="sidebar-footer">
        <Separator className="w-8 bg-gray-200" />

        {/* Connection status indicator */}
        <Circle
          className={cn(
            'w-3 h-3',
            connected ? 'fill-green-500 text-green-500' : 'fill-red-500 text-red-500'
          )}
        />

        {/* Settings button */}
        <Button
          variant="ghost"
          size="icon"
          className="w-10 h-10 text-gray-500 hover:text-gray-700"
          title="设置"
        >
          <Settings className="w-5 h-5" />
        </Button>
      </div>
    </div>
  )
}
