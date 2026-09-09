import { Icon } from './Icon'
import { NavLink } from 'react-router-dom'
import { useFollowing } from '../state/FollowingContext'
import { findUserByHandle } from '../data/users'

export function RightRail() {
  const { toggleFollowing, isFollowing, isUpdating, error } = useFollowing()
  const recommendedHandles = ['@sho_sato', '@product_team']

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
        {recommendedHandles.map((handle) => {
          const user = findUserByHandle(handle)
          if (!user) return null
          return (
            <div className="follow-row" key={user.id}>
              <span className={`avatar ${user.avatarClass ?? ''}`}>{user.avatar}</span>
              <NavLink className="account-copy user-link" to={`/users/${user.handle.slice(1)}`}>
                <strong>{user.displayName}</strong>
                <small>{user.handle}</small>
              </NavLink>
              <button
                type="button"
                className={`follow-button ${isFollowing(user.handle) ? 'following' : ''}`}
                disabled={isUpdating(user.handle)}
                onClick={() => void toggleFollowing(user.handle)}
              >
                {isFollowing(user.handle) ? 'フォロー中' : 'フォロー'}
              </button>
            </div>
          )
        })}
        {error && <p role="alert">{error}</p>}
      </section>
      <footer className="footer-links">利用規約　プライバシーポリシー　© 2026 ENG-1103</footer>
    </aside>
  )
}
