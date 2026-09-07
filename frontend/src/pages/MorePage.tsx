import { ContentPage } from './ContentPage'

export function MorePage() {
  return (
    <ContentPage title="もっと見る" description="設定やその他の機能を利用できます。">
      <div className="more-menu">
        <button>設定とプライバシー</button>
        <button>ヘルプセンター</button>
        <button>表示設定</button>
      </div>
    </ContentPage>
  )
}
