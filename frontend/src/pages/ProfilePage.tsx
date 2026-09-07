import { ContentPage } from './ContentPage'

export function ProfilePage() {
  return (
    <ContentPage title="プロフィール" description="あなたのプロフィールとポストを表示します。">
      <div className="profile-card">
        <span className="avatar avatar-blue profile-avatar">太</span>
        <h2>田中 太郎</h2>
        <p>@taro_tanaka</p>
        <p>プロダクト開発とユーザー体験について考えています。</p>
      </div>
    </ContentPage>
  )
}
