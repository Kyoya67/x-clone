type TimelineHeaderProps = { activeTab: string; onTabChange: (tab: string) => void }

export function TimelineHeader({ activeTab, onTabChange }: TimelineHeaderProps) {
  return (
    <header className="timeline-header">
      <h1>ホーム</h1>
      <div className="tabs" role="tablist">
        {['おすすめ', 'フォロー中'].map((tab) => (
          <button
            key={tab}
            className={activeTab === tab ? 'tab active' : 'tab'}
            onClick={() => onTabChange(tab)}
            role="tab"
            aria-selected={activeTab === tab}
            tabIndex={-1}
          >
            {tab}
          </button>
        ))}
      </div>
    </header>
  )
}
