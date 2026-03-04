import { useState } from 'react'
import { Shield, Zap } from 'lucide-react'
import { cn } from '@/lib/utils'
import { PassiveScanTab } from './PassiveScanTab'
import { ActiveScanTab } from './ActiveScanTab'

type ScanSubTab = 'passive' | 'active'

const subTabs: { id: ScanSubTab; label: string }[] = [
  { id: 'passive', label: 'Passive Scan' },
  { id: 'active', label: 'Active Scan' },
]

export function ScanTab() {
  const [activeSubTab, setActiveSubTab] = useState<ScanSubTab>('passive')

  const renderSubTab = () => {
    switch (activeSubTab) {
      case 'passive':
        return <PassiveScanTab />
      case 'active':
        return <ActiveScanTab />
      default:
        return <PassiveScanTab />
    }
  }

  return (
    <div className="flex h-full flex-col">
      {/* Sub-tabs bar - Burp style */}
      <div className="sub-tabs-bar">
        {subTabs.map((tab) => (
          <button
            key={tab.id}
            onClick={() => setActiveSubTab(tab.id)}
            className={cn('sub-tab', activeSubTab === tab.id && 'sub-tab-active')}
          >
            {tab.id === 'passive' && <Shield className="w-3.5 h-3.5 mr-1" />}
            {tab.id === 'active' && <Zap className="w-3.5 h-3.5 mr-1" />}
            {tab.label}
          </button>
        ))}
      </div>

      {/* Sub-tab content */}
      <div className="flex-1 overflow-hidden">
        {renderSubTab()}
      </div>
    </div>
  )
}
