import { NavLink } from 'react-router-dom'
import { Icon } from './Icon'

export function Sidebar() {
  return (
    <aside className="sidebar">
      <div className="brand-mark">𝕏</div>
      <nav className="main-nav" aria-label="メインメニュー">
        <NavLink className="nav-item" to="/">
          <Icon name="home" />
          <span>ホーム</span>
        </NavLink>
        <NavLink className="nav-item" to="/explore">
          <Icon name="search" />
          <span>話題を検索</span>
        </NavLink>
        <NavLink className="nav-item" to="/notifications">
          <Icon name="bell" />
          <span>通知</span>
          <span className="badge">3</span>
        </NavLink>
        <NavLink className="nav-item" to="/messages">
          <Icon name="chat" />
          <span>チャット</span>
        </NavLink>
        <NavLink className="nav-item" to="/profile">
          <Icon name="profile" />
          <span>プロフィール</span>
        </NavLink>
        <NavLink className="nav-item" to="/more">
          <Icon name="more" />
          <span>もっと見る</span>
        </NavLink>
      </nav>
      <button
        className="post-button"
        type="button"
        onClick={() => document.getElementById('composer')?.focus()}
      >
        ポストする
      </button>
      <button className="account-card" type="button">
        <span className="avatar avatar-blue">太</span>
        <span className="account-copy">
          <strong>田中 太郎</strong>
          <small>@taro_tanaka</small>
        </span>
        <span className="more">•••</span>
      </button>
    </aside>
  )
}
