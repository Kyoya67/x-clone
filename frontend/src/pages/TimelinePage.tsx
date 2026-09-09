import { FormEvent, useEffect, useState } from 'react'
import { createPost } from '../api/posts'
import { fetchTimeline, TimelineFeed } from '../api/timeline'
import { Composer } from '../components/Composer'
import { Feed } from '../components/Feed'
import { PageLayout } from '../components/PageLayout'
import { TimelineHeader } from '../components/TimelineHeader'
import { Post } from '../types/post'

export function TimelinePage() {
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
  const toggleLike = (id: string) =>
    setPosts((current) =>
      current.map((post) =>
        post.id === id
          ? { ...post, liked: !post.liked, likes: post.likes + (post.liked ? -1 : 1) }
          : post,
      ),
    )
  return (
    <PageLayout>
      <div className="timeline">
        <TimelineHeader activeTab={activeTab} onTabChange={setActiveTab} />
        <Composer draft={draft} onDraftChange={setDraft} onPublish={publish} />
        {publishError && <p role="alert">{publishError}</p>}
        {timelineError && <p role="alert">{timelineError}</p>}
        {isLoading ? (
          <p role="status">タイムラインを読み込んでいます。</p>
        ) : posts.length === 0 ? (
          <p>表示する投稿はありません。</p>
        ) : (
          <Feed posts={posts} onToggleLike={toggleLike} />
        )}
      </div>
    </PageLayout>
  )
}
