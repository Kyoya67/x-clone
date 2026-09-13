import { ContentPage } from './ContentPage'
import { showDemoUsers } from '../config/currentUser'

export function NotificationsPage() {
  return (
    <ContentPage title="通知" description="あなたに関する最新の通知を表示します。">
      {showDemoUsers ? (
        <>
          <div className="notification-item">
            <span className="avatar avatar-blue">太</span>
            <p>
              <strong>田中 太郎</strong>さんがあなたのポストをいいねしました。
            </p>
          </div>
          <div className="notification-item">
            <span className="avatar avatar-green">翔</span>
            <p>
              <strong>佐藤 翔</strong>さんがあなたをフォローしました。
            </p>
          </div>
        </>
      ) : (
        <p>通知はまだありません。</p>
      )}
    </ContentPage>
  )
}
