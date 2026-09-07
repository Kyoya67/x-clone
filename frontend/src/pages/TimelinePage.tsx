import { FormEvent, useState } from 'react'
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
  const { followingHandles } = useFollowing()
  const publish = (event: FormEvent) => {
    event.preventDefault()
    const body = draft.trim()
    if (!body) return
    setPosts((current) => [
      {
        id: Date.now(),
        name: '田中 太郎',
        handle: '@taro_tanaka',
        body,
        time: '今',
        avatar: '太',
        likes: 0,
        replies: 0,
        reposts: 0,
      },
      ...current,
    ])
    setDraft('')
  }
  const toggleLike = (id: number) =>
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
        <Feed posts={visiblePosts} onToggleLike={toggleLike} />
      </div>
    </PageLayout>
  )
}
