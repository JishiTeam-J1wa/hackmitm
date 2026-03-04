import { useState } from 'react'
import { Settings, Search, Puzzle } from 'lucide-react'
import { cn } from '@/lib/utils'
import { ScanManager } from './ScanManager'
import { ScanResults } from './ScanResults'
import { PluginManager } from './PluginManager'

type SubTab = 'management' | 'results' | 'plugins'

const subTabs: { id: SubTab; label: string }[] = [
  { id: 'management', label: 'Management' },
  { id: 'results', label: 'Results' },
  { id: 'plugins', label: 'Plugins' },
]

export function PassiveScanTab() {
  const [activeSubTab, setActiveSubTab] = useState<SubTab>('results')

  const renderSubTab = () => {
    switch (activeSubTab) {
      case 'management':
        return <ScanManager />
      case 'results':
        return <ScanResults />
      case 'plugins':
        return <PluginManager />
      default:
        return <ScanResults />
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
            {tab.id === 'management' && <Settings className="w-3.5 h-3.5 mr-1" />}
            {tab.id === 'results' && <Search className="w-3.5 h-3.5 mr-1" />}
            {tab.id === 'plugins' && <Puzzle className="w-3.5 h-3.5 mr-1" />}
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
