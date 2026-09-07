import { FormEvent, useState } from 'react'
import { BrowserRouter, NavLink, Route, Routes } from 'react-router-dom'
import { Composer } from './components/Composer'
import { Feed } from './components/Feed'
import { RightRail } from './components/RightRail'
import { Sidebar } from './components/Sidebar'
import { TimelineHeader } from './components/TimelineHeader'
import { initialPosts, Post } from './data/posts'

function Timeline() {
  const [posts, setPosts] = useState<Post[]>(initialPosts)
  const [draft, setDraft] = useState('')

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

  return (
    <div className="app-shell">
      <Sidebar />
      <main className="timeline" id="timeline">
        <TimelineHeader />
        <Composer draft={draft} onDraftChange={setDraft} onPublish={publish} />
        <Feed posts={posts} onToggleLike={toggleLike} />
      </main>
      <RightRail />
    </div>
  )
}

function Placeholder({ title }: { title: string }) {
  return (
    <div className="placeholder-page">
      <div className="brand-mark">𝕏</div>
      <h1>{title}</h1>
      <p>この画面は今後実装予定です。</p>
      <NavLink className="back-home" to="/">
        ホームに戻る
      </NavLink>
    </div>
  )
}

function App() {
  return (
    <BrowserRouter>
      <Routes>
        <Route path="/" element={<Timeline />} />
        <Route path="/explore" element={<Placeholder title="話題を検索" />} />
        <Route path="/notifications" element={<Placeholder title="通知" />} />
        <Route path="/messages" element={<Placeholder title="チャット" />} />
        <Route path="/profile" element={<Placeholder title="プロフィール" />} />
        <Route path="/more" element={<Placeholder title="もっと見る" />} />
        <Route path="*" element={<Placeholder title="ページが見つかりません" />} />
      </Routes>
    </BrowserRouter>
  )
}

export default App
