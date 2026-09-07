import { useState } from 'react'

export function TimelineHeader() {
  const [activeTab, setActiveTab] = useState('おすすめ')
  return (
    <header className="timeline-header">
      <h1>ホーム</h1>
      <div className="tabs" role="tablist">
        {['おすすめ', 'フォロー中'].map((tab) => (
          <button
            key={tab}
            className={activeTab === tab ? 'tab active' : 'tab'}
            onClick={() => setActiveTab(tab)}
            role="tab"
            aria-selected={activeTab === tab}
          >
            {tab}
          </button>
        ))}
      </div>
    </header>
  )
}
