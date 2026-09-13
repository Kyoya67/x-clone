import { useEffect, useState } from 'react'
import { useParams } from 'react-router-dom'
import { fetchTimeline } from '../api/timeline'
import { Feed } from '../components/Feed'
import { ContentPage } from './ContentPage'
import { useFollowing } from '../state/FollowingContext'
import { findUserByHandle, User } from '../data/users'
import { Post } from '../types/post'
import { avatarColorClass } from '../utils/avatarColor'

export function UserProfilePage() {
  const { handle = 'unknown' } = useParams()
  const staticUser = findUserByHandle(`@${handle}`)
  const [profileUser, setProfileUser] = useState<User | null>(staticUser ?? null)
  const user = profileUser ?? staticUser
  const displayHandle = user?.handle ?? `@${handle}`
  const displayName = user?.displayName ?? 'ユーザー'
  const { toggleFollowing, isFollowing, isUpdating, error } = useFollowing()
  const [posts, setPosts] = useState<Post[]>([])
  const [isLoading, setIsLoading] = useState(true)
  const [timelineError, setTimelineError] = useState('')

  useEffect(() => {
    let cancelled = false

    const loadPosts = async () => {
      setIsLoading(true)
      setTimelineError('')
      setProfileUser(staticUser ?? null)
      try {
        const timelinePosts = await fetchTimeline('for-you')
        if (!cancelled) {
          const userPosts = timelinePosts.filter((post) => post.handle === `@${handle}`)
          setPosts(userPosts)
          const firstPost = userPosts[0]
          if (!staticUser && firstPost) {
            setProfileUser({
              id: firstPost.authorId,
              handle: firstPost.handle,
              displayName: firstPost.name,
              bio: '自己紹介はまだありません。',
              avatar: firstPost.avatar,
            })
          }
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
  }, [handle, staticUser])

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
        <span
          className={`avatar ${user?.avatarClass ?? avatarColorClass(user?.id ?? displayHandle)} profile-avatar`}
        >
          {user?.avatar ?? '?'}
        </span>
        <h2>{displayName}</h2>
        <p>{displayHandle}</p>
        <p>{user?.bio ?? '自己紹介はまだありません。'}</p>
        <button
          className={`follow-button ${user && isFollowing(user.id) ? 'following' : ''}`}
          disabled={!user || isUpdating(user.id)}
          type="button"
          onClick={() => user && void toggleFollowing(user.id)}
        >
          {user && isFollowing(user.id) ? 'フォロー中' : 'フォロー'}
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
