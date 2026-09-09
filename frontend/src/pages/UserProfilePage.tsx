import { useParams } from 'react-router-dom'
import { ContentPage } from './ContentPage'
import { useFollowing } from '../state/FollowingContext'
import { findUserByHandle } from '../data/users'

export function UserProfilePage() {
  const { handle = 'unknown' } = useParams()
  const user = findUserByHandle(`@${handle}`)
  const displayHandle = user?.handle ?? `@${handle}`
  const displayName = user?.displayName ?? 'ユーザー'
  const profileHandle = user?.handle ?? `@${handle}`
  const { toggleFollowing, isFollowing, isUpdating, error } = useFollowing()

  return (
    <ContentPage title="プロフィール" description="ユーザーのプロフィールとポストを表示します。">
      <div className="profile-card">
        <span className={`avatar avatar-blue profile-avatar ${user?.avatarClass ?? ''}`}>
          {user?.avatar ?? '?'}
        </span>
        <h2>{displayName}</h2>
        <p>{displayHandle}</p>
        <p>{user?.bio ?? '自己紹介はまだありません。'}</p>
        <button
          className={`follow-button ${isFollowing(profileHandle) ? 'following' : ''}`}
          disabled={!user || isUpdating(profileHandle)}
          type="button"
          onClick={() => void toggleFollowing(profileHandle)}
        >
          {isFollowing(profileHandle) ? 'フォロー中' : 'フォロー'}
        </button>
        {error && <p role="alert">{error}</p>}
      </div>
    </ContentPage>
  )
}
