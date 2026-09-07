import { useParams } from 'react-router-dom'
import { ContentPage } from './ContentPage'
import { useFollowing } from '../state/FollowingContext'

export function UserProfilePage() {
  const { handle = 'unknown' } = useParams()
  const users: Record<string, { name: string; avatar: string; avatarClass?: string }> = {
    taro_tanaka: { name: '田中 太郎', avatar: '太' },
    hanako_s: { name: '鈴木 花子', avatar: '花', avatarClass: 'avatar-pink' },
    ken_yamamoto: { name: '山本 健', avatar: '健', avatarClass: 'avatar-green' },
    sho_sato: { name: '佐藤 翔', avatar: '翔', avatarClass: 'avatar-green' },
    product_team: { name: 'プロダクト開発部', avatar: '開', avatarClass: 'avatar-pink' },
  }
  const user = users[handle] ?? { name: 'ユーザー', avatar: '?', avatarClass: '' }
  const displayHandle = `@${handle}`
  const displayName = user.name
  const profileHandle = `@${handle}`
  const { toggleFollowing, isFollowing } = useFollowing()

  return (
    <ContentPage title="プロフィール" description="ユーザーのプロフィールとポストを表示します。">
      <div className="profile-card">
        <span className={`avatar avatar-blue profile-avatar ${user.avatarClass ?? ''}`}>
          {user.avatar}
        </span>
        <h2>{displayName}</h2>
        <p>{displayHandle}</p>
        <p>プロダクト開発とユーザー体験について考えています。</p>
        <button
          className={`follow-button ${isFollowing(profileHandle) ? 'following' : ''}`}
          type="button"
          onClick={() => toggleFollowing(profileHandle)}
        >
          {isFollowing(profileHandle) ? 'フォロー中' : 'フォロー'}
        </button>
      </div>
    </ContentPage>
  )
}
