import { FormEvent, useState } from 'react'
import { createPost } from '../api/posts'
import { Composer } from '../components/Composer'
import { Feed } from '../components/Feed'
import { PageLayout } from '../components/PageLayout'
import { TimelineHeader } from '../components/TimelineHeader'
import { initialPosts, Post } from '../data/posts'
import { useFollowing } from '../state/FollowingContext'

export function TimelinePage() {
  const [posts, setPosts] = useState<Post[]>(initialPosts)
  const [draft, setDraft] = useState('')
  const [activeTab, setActiveTab] = useState('おすすめ')
  const [error, setError] = useState('')
  const [isPublishing, setIsPublishing] = useState(false)
  const { followingHandles } = useFollowing()
  const publish = async (event: FormEvent) => {
    event.preventDefault()
    const body = draft.trim()
    if (!body || isPublishing) return

    setError('')
    setIsPublishing(true)
    try {
      const post = await createPost(body)
      setPosts((current) => [
        {
          id: post.id,
          name: '田中 太郎',
          handle: '@taro_tanaka',
          body: post.content,
          time: '今',
          avatar: '太',
          likes: 0,
          replies: 0,
          reposts: 0,
        },
        ...current,
      ])
      setDraft('')
    } catch {
      setError('投稿に失敗しました。時間をおいて再度お試しください。')
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
  const visiblePosts =
    activeTab === 'フォロー中'
      ? posts.filter((post) => followingHandles.includes(post.handle))
      : posts

  return (
    <PageLayout>
      <div className="timeline">
        <TimelineHeader activeTab={activeTab} onTabChange={setActiveTab} />
        <Composer draft={draft} onDraftChange={setDraft} onPublish={publish} />
        {error && <p role="alert">{error}</p>}
        <Feed posts={visiblePosts} onToggleLike={toggleLike} />
      </div>
    </PageLayout>
  )
}
