import { Icon } from './Icon'

export function RightRail() {
  return (
    <aside className="right-rail">
      <label className="search-box">
        <Icon name="search" />
        <input placeholder="検索" />
      </label>
      <section className="rail-card">
        <h2>いまどうしてる？</h2>
        <div className="trend">
          <small>日本のトレンド</small>
          <strong>#ソフトウェア開発</strong>
          <small>12,438件のポスト</small>
        </div>
        <div className="trend">
          <small>テクノロジー · トレンド</small>
          <strong>TypeScript</strong>
          <small>8,210件のポスト</small>
        </div>
        <button className="show-more" type="button">
          さらに表示
        </button>
      </section>
      <section className="rail-card follow-card">
        <h2>おすすめユーザー</h2>
        {['佐藤 翔', 'プロダクト開発部'].map((name, index) => (
          <div className="follow-row" key={name}>
            <span className={`avatar ${index ? 'avatar-pink' : 'avatar-green'}`}>
              {index ? '開' : '翔'}
            </span>
            <span className="account-copy">
              <strong>{name}</strong>
              <small>@{index ? 'product_team' : 'sho_sato'}</small>
            </span>
            <button type="button" className="follow-button">
              フォロー
            </button>
          </div>
        ))}
      </section>
      <footer className="footer-links">利用規約　プライバシーポリシー　© 2026 ENG-1103</footer>
    </aside>
  )
}
