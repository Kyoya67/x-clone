import { FormEvent, useEffect, useState } from 'react'
import { likePost, unlikePost } from '../api/likes'
import { createPost } from '../api/posts'
import { fetchTimeline, TimelineFeed } from '../api/timeline'
import { Composer } from '../components/Composer'
import { Feed } from '../components/Feed'
import { PageLayout } from '../components/PageLayout'
import { TimelineHeader } from '../components/TimelineHeader'
import { currentUser } from '../config/currentUser'
import { useOptionalAuth } from '../state/AuthContext'
import { Post } from '../types/post'
import { avatarColorClass } from '../utils/avatarColor'

export function TimelinePage() {
  const auth = useOptionalAuth()
  const user = auth?.user
  const displayName = user?.displayName ?? currentUser.displayName
  const avatar = displayName.slice(0, 1) || 'U'
  const avatarClass = avatarColorClass(user?.id ?? user?.handle ?? currentUser.id)
  const [posts, setPosts] = useState<Post[]>([])
  const [draft, setDraft] = useState('')
  const [activeTab, setActiveTab] = useState('おすすめ')
  const [publishError, setPublishError] = useState('')
  const [timelineError, setTimelineError] = useState('')
  const [isLoading, setIsLoading] = useState(true)
  const [isPublishing, setIsPublishing] = useState(false)
  const [refreshKey, setRefreshKey] = useState(0)
  const feed: TimelineFeed = activeTab === 'フォロー中' ? 'following' : 'for-you'

  useEffect(() => {
    let cancelled = false

    const loadTimeline = async () => {
      setIsLoading(true)
      setTimelineError('')
      try {
        const timelinePosts = await fetchTimeline(feed)
        if (!cancelled) setPosts(timelinePosts)
      } catch {
        if (!cancelled) {
          setPosts([])
          setTimelineError('タイムラインの取得に失敗しました。時間をおいて再度お試しください。')
        }
      } finally {
        if (!cancelled) setIsLoading(false)
      }
    }

    void loadTimeline()
    return () => {
      cancelled = true
    }
  }, [feed, refreshKey])

  const publish = async (event: FormEvent) => {
    event.preventDefault()
    const body = draft.trim()
    if (!body || isPublishing) return

    setPublishError('')
    setIsPublishing(true)
    try {
      await createPost(body)
      setDraft('')
      setRefreshKey((current) => current + 1)
    } catch {
      setPublishError('投稿に失敗しました。時間をおいて再度お試しください。')
    } finally {
      setIsPublishing(false)
    }
  }
  const toggleLike = async (id: string) => {
    const target = posts.find((post) => post.id === id)
    if (!target) return

    setPosts((current) =>
      current.map((post) =>
        post.id === id
          ? { ...post, liked: !post.liked, likes: post.likes + (post.liked ? -1 : 1) }
          : post,
      ),
    )
    try {
      if (target.liked) {
        await unlikePost(id)
      } else {
        await likePost(id)
      }
    } catch {
      setPosts((current) =>
        current.map((post) =>
          post.id === id ? { ...post, liked: target.liked, likes: target.likes } : post,
        ),
      )
      setTimelineError('いいね操作に失敗しました。時間をおいて再度お試しください。')
    }
  }
  return (
    <PageLayout>
      <div className="timeline">
        <TimelineHeader activeTab={activeTab} onTabChange={setActiveTab} />
        <Composer
          avatar={avatar}
          avatarClass={avatarClass}
          draft={draft}
          onDraftChange={setDraft}
          onPublish={publish}
        />
        {publishError && <p role="alert">{publishError}</p>}
        {timelineError && <p role="alert">{timelineError}</p>}
        {isLoading ? (
          <p role="status">タイムラインを読み込んでいます。</p>
        ) : posts.length === 0 ? (
          <p>表示する投稿はありません。</p>
        ) : (
          <Feed posts={posts} onToggleLike={toggleLike} currentUserHandle={user?.handle} />
        )}
      </div>
    </PageLayout>
  )
}
