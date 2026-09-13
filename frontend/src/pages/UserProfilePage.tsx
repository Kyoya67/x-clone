import { useEffect, useState } from 'react'
import { useParams } from 'react-router-dom'
import { fetchTimeline } from '../api/timeline'
import { Feed } from '../components/Feed'
import { ContentPage } from './ContentPage'
import { useFollowing } from '../state/FollowingContext'
import { findUserByHandle } from '../data/users'
import { Post } from '../types/post'

export function UserProfilePage() {
  const { handle = 'unknown' } = useParams()
  const user = findUserByHandle(`@${handle}`)
  const displayHandle = user?.handle ?? `@${handle}`
  const displayName = user?.displayName ?? 'ユーザー'
  const profileHandle = user?.handle ?? `@${handle}`
  const { toggleFollowing, isFollowing, isUpdating, error } = useFollowing()
  const [posts, setPosts] = useState<Post[]>([])
  const [isLoading, setIsLoading] = useState(true)
  const [timelineError, setTimelineError] = useState('')

  useEffect(() => {
    let cancelled = false

    const loadPosts = async () => {
      setIsLoading(true)
      setTimelineError('')
      try {
        const timelinePosts = await fetchTimeline('for-you')
        if (!cancelled) {
          setPosts(timelinePosts.filter((post) => post.handle === profileHandle))
        }
      } catch {
        if (!cancelled) {
          setPosts([])
          setTimelineError('ポストの取得に失敗しました。時間をおいて再度お試しください。')
        }
      } finally {
        if (!cancelled) setIsLoading(false)
      }
    }

    void loadPosts()
    return () => {
      cancelled = true
    }
  }, [profileHandle])

  const toggleLike = (id: string) =>
    setPosts((current) =>
      current.map((post) =>
        post.id === id
          ? { ...post, liked: !post.liked, likes: post.likes + (post.liked ? -1 : 1) }
          : post,
      ),
    )

  return (
    <ContentPage title="プロフィール">
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
      {timelineError && <p role="alert">{timelineError}</p>}
      {isLoading ? (
        <p role="status">ポストを読み込んでいます。</p>
      ) : posts.length === 0 ? (
        <p>表示する投稿はありません。</p>
      ) : (
        <Feed posts={posts} onToggleLike={toggleLike} />
      )}
    </ContentPage>
  )
}
