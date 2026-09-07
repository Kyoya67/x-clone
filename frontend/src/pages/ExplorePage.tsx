import { ContentPage } from './ContentPage'

export function ExplorePage() {
  return (
    <ContentPage title="話題を検索" description="いま話題のポストやトレンドを検索できます。">
      <div className="rail-card">
        <h2>おすすめのトレンド</h2>
        <div className="trend">
          <small>テクノロジー · トレンド</small>
          <strong>TypeScript</strong>
          <small>8,210件のポスト</small>
        </div>
        <div className="trend">
          <small>日本のトレンド</small>
          <strong>#ソフトウェア開発</strong>
          <small>12,438件のポスト</small>
        </div>
      </div>
    </ContentPage>
  )
}
