import { ContentPage } from './ContentPage'
import { currentUser } from '../config/currentUser'

export function ProfilePage() {
  return (
    <ContentPage title="プロフィール" description="あなたのプロフィールとポストを表示します。">
      <div className="profile-card">
        <span className="avatar avatar-blue profile-avatar">{currentUser.avatar}</span>
        <h2>{currentUser.displayName}</h2>
        <p>{currentUser.handle}</p>
        <p>{currentUser.bio}</p>
      </div>
    </ContentPage>
  )
}
