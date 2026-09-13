import { useEffect, useState } from 'react'
import { fetchNotifications, Notification } from '../api/notifications'
import { ContentPage } from './ContentPage'
import { avatarColorClass } from '../utils/avatarColor'

export function NotificationsPage() {
  const [notifications, setNotifications] = useState<Notification[]>([])
  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState('')

  useEffect(() => {
    let cancelled = false

    const loadNotifications = async () => {
      setIsLoading(true)
      setError('')
      try {
        const nextNotifications = await fetchNotifications()
        if (!cancelled) setNotifications(nextNotifications)
      } catch {
        if (!cancelled) {
          setNotifications([])
          setError('通知の取得に失敗しました。時間をおいて再度お試しください。')
        }
      } finally {
        if (!cancelled) setIsLoading(false)
      }
    }

    void loadNotifications()
    return () => {
      cancelled = true
    }
  }, [])

  return (
    <ContentPage title="通知" description="あなたに関する最新の通知を表示します。">
      {error && <p role="alert">{error}</p>}
      {isLoading ? (
        <p role="status">通知を読み込んでいます。</p>
      ) : notifications.length === 0 ? (
        <p>通知はまだありません。</p>
      ) : (
        notifications.map((notification) => (
          <div className="notification-item" key={notification.id}>
            <span className={`avatar ${avatarColorClass(notification.actor.id)}`}>
              {notification.actor.displayName.slice(0, 1)}
            </span>
            <p>
              <strong>{notification.actor.displayName}</strong>
              {notification.type === 'follow'
                ? 'さんがあなたをフォローしました。'
                : 'さんがあなたのポストをいいねしました。'}
            </p>
          </div>
        ))
      )}
    </ContentPage>
  )
}
