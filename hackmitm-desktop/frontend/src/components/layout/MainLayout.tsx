import { useState } from 'react'
import { Header } from './Header'
import { StatusBar } from './StatusBar'
import { ProxyTab } from '@/components/proxy/ProxyTab'
import { TargetTab } from '@/components/target/TargetTab'
import { RepeaterTab } from '@/components/repeater/RepeaterTab'
import { IntruderTab } from '@/components/intruder'
import { FingerprintTab } from '@/components/fingerprint/FingerprintTab'
import { DashboardTab } from '@/components/dashboard/DashboardTab'
import { WebSocketTab } from '@/components/proxy/WebSocketTab'
import { VulnTab } from '@/components/vuln'
import { ScanTab } from '@/components/scan'
import { cn } from '@/lib/utils'
import type { TabId } from '@/types'

// 顶部标签页配置 - 类似 Burp Suite 风格
const mainTabs: { id: TabId; label: string }[] = [
  { id: 'dashboard', label: 'Dashboard' },
  { id: 'proxy', label: 'Proxy' },
  { id: 'target', label: 'Target' },
  { id: 'repeater', label: 'Repeater' },
  { id: 'intruder', label: 'Intruder' },
  { id: 'fingerprint', label: 'Fingerprint' },
  { id: 'vuln', label: 'Vulns' },
  { id: 'scan', label: 'Scan' },
]

export function MainLayout() {
  const [activeTab, setActiveTab] = useState<TabId>('proxy')

  const renderTab = () => {
    switch (activeTab) {
      case 'dashboard':
        return <DashboardTab />
      case 'proxy':
        return <ProxyTab />
      case 'target':
        return <TargetTab />
      case 'repeater':
        return <RepeaterTab />
      case 'intruder':
        return <IntruderTab />
      case 'fingerprint':
        return <FingerprintTab />
      case 'websocket':
        return <WebSocketTab />
      case 'vuln':
        return <VulnTab />
      case 'scan':
        return <ScanTab />
      default:
        return <ProxyTab />
    }
  }

  return (
    <div className="app-container">
      {/* 顶部标签栏 - 类似 Burp Suite */}
      <div className="top-tabs-bar">
        {/* Logo */}
        <div className="top-tabs-logo">
          <img
            src="/j1shi-logo.png"
            alt="击势安全团队"
            className="w-6 h-6"
          />
          <span className="text-xs font-medium text-gray-600 ml-1.5">HackMITM</span>
        </div>

        {/* 主标签页 */}
        <div className="top-tabs">
          {mainTabs.map((tab) => (
            <button
              key={tab.id}
              onClick={() => setActiveTab(tab.id)}
              className={cn(
                'top-tab',
                activeTab === tab.id && 'top-tab-active'
              )}
            >
              {tab.label}
            </button>
          ))}
        </div>

        {/* 右侧工具区 */}
        <div className="top-tabs-right">
          <Header />
        </div>
      </div>

      {/* 主内容区 */}
      <div className="main-content-area">
        {renderTab()}
      </div>

      {/* 状态栏 */}
      <StatusBar />
    </div>
  )
}
