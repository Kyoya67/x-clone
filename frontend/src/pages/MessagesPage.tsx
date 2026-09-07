import { ContentPage } from './ContentPage'

export function MessagesPage() {
  return (
    <ContentPage title="チャット" description="メッセージを送受信できます。">
      <div className="message-row">
        <span className="avatar avatar-pink">開</span>
        <span className="account-copy">
          <strong>プロダクト開発部</strong>
          <small>新しいプロジェクトについて</small>
        </span>
        <span className="message-time">昨日</span>
      </div>
    </ContentPage>
  )
}
