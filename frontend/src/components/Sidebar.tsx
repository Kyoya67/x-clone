import { NavLink } from 'react-router-dom'
import { Icon } from './Icon'
import { currentUser } from '../config/currentUser'
import { useOptionalAuth } from '../state/AuthContext'

export function Sidebar() {
  const auth = useOptionalAuth()
  const user = auth?.user
  const displayName = user?.displayName ?? currentUser.displayName
  const handle = user?.handle ?? currentUser.handle.replace(/^@/, '')
  const avatar = displayName.slice(0, 1) || 'U'

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
      <div className="account-menu">
        <div className="account-popover" role="menu" aria-label="アカウントメニュー">
          <button type="button" role="menuitem">
            既存のアカウントを追加
          </button>
          {auth && (
            <button type="button" role="menuitem" onClick={() => void auth.logout()}>
              @{handle}からログアウト
            </button>
          )}
        </div>
        <button className="account-card" type="button" aria-haspopup="menu">
          <span className="avatar avatar-blue">{avatar}</span>
          <span className="account-copy">
            <strong>{displayName}</strong>
            <small>@{handle}</small>
          </span>
          <span className="more">•••</span>
        </button>
      </div>
    </aside>
  )
}
