import { ContentPage } from './ContentPage'
import { currentUser } from '../config/currentUser'
import { useOptionalAuth } from '../state/AuthContext'

export function ProfilePage() {
  const auth = useOptionalAuth()
  const user = auth?.user
  const displayName = user?.displayName ?? currentUser.displayName
  const handle = user?.handle ?? currentUser.handle.replace(/^@/, '')
  const bio = user?.bio ?? currentUser.bio
  const avatar = displayName.slice(0, 1) || 'U'

  return (
    <ContentPage title="プロフィール">
      <div className="profile-card">
        <span className="avatar avatar-blue profile-avatar">{avatar}</span>
        <h2>{displayName}</h2>
        <p>@{handle}</p>
        <p>{bio}</p>
      </div>
    </ContentPage>
  )
}
