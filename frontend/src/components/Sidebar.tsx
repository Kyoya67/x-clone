import { useEffect, useRef, useState } from 'react'
import { NavLink } from 'react-router-dom'
import { fetchNotifications } from '../api/notifications'
import { Icon } from './Icon'
import { currentUser } from '../config/currentUser'
import { useOptionalAuth } from '../state/AuthContext'

type Theme = 'light' | 'dark'

export function Sidebar() {
  const [isAccountMenuOpen, setIsAccountMenuOpen] = useState(false)
  const [notificationCount, setNotificationCount] = useState(0)
  const [theme, setTheme] = useState<Theme>(() => {
    const savedTheme = localStorage.getItem('theme')
    return savedTheme === 'dark' || savedTheme === 'light' ? savedTheme : 'light'
  })
  const accountMenuRef = useRef<HTMLDivElement | null>(null)
  const auth = useOptionalAuth()
  const user = auth?.user
  const displayName = user?.displayName ?? currentUser.displayName
  const handle = user?.handle ?? currentUser.handle.replace(/^@/, '')
  const avatar = displayName.slice(0, 1) || 'U'

  useEffect(() => {
    let cancelled = false
    const loadNotifications = async () => {
      try {
        const notifications = await fetchNotifications()
        if (!cancelled) setNotificationCount(notifications.length)
      } catch {
        if (!cancelled) setNotificationCount(0)
      }
    }
    void loadNotifications()
    return () => {
      cancelled = true
    }
  }, [])

  useEffect(() => {
    const closeOnOutsideClick = (event: MouseEvent) => {
      if (!accountMenuRef.current?.contains(event.target as Node)) {
        setIsAccountMenuOpen(false)
      }
    }
    document.addEventListener('click', closeOnOutsideClick)
    return () => document.removeEventListener('click', closeOnOutsideClick)
  }, [])

  useEffect(() => {
    document.documentElement.dataset.theme = theme
    localStorage.setItem('theme', theme)
  }, [theme])

  const nextTheme = theme === 'dark' ? 'light' : 'dark'

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
          {notificationCount > 0 && <span className="badge">{notificationCount}</span>}
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
      <button
        className="theme-toggle"
        type="button"
        aria-label={nextTheme === 'dark' ? 'ダークモード' : 'ライトモード'}
        onClick={() => setTheme(nextTheme)}
      >
        <Icon name={nextTheme === 'dark' ? 'moon' : 'sun'} />
      </button>
      <div className="account-menu" ref={accountMenuRef}>
        {isAccountMenuOpen && (
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
        )}
        <button
          className="account-card"
          type="button"
          aria-haspopup="menu"
          aria-expanded={isAccountMenuOpen}
          onClick={() => setIsAccountMenuOpen((current) => !current)}
        >
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
