import { FormEvent, useState } from 'react'
import { BrowserRouter, NavLink, Route, Routes } from 'react-router-dom'

type Post = {
  id: number
  name: string
  handle: string
  body: string
  time: string
  avatar: string
  avatarClass?: string
  likes: number
  replies: number
  reposts: number
  liked?: boolean
}

const initialPosts: Post[] = [
  {
    id: 1,
    name: '田中 太郎',
    handle: '@taro_tanaka',
    body: '今日はチームでタイムライン機能の設計をしました。小さく作って検証するのが気持ちいい。',
    time: '2時間',
    avatar: '太',
    likes: 18,
    replies: 3,
    reposts: 4,
  },
  {
    id: 2,
    name: '鈴木 花子',
    handle: '@hanako_s',
    body: '新しいサービスの最初の一歩。ユーザーが迷わず使える体験を大切にしたい。',
    time: '5時間',
    avatar: '花',
    avatarClass: 'avatar-pink',
    likes: 42,
    replies: 8,
    reposts: 9,
  },
  {
    id: 3,
    name: '山本 健',
    handle: '@ken_yamamoto',
    body: 'コードレビューで設計の意図を共有する時間が好きです。',
    time: '昨日',
    avatar: '健',
    avatarClass: 'avatar-green',
    likes: 27,
    replies: 2,
    reposts: 5,
  },
]

type IconName = 'home' | 'search' | 'bell' | 'chat' | 'profile' | 'more'

function Icon({ name }: { name: IconName }) {
  const paths: Record<IconName, string> = {
    home: 'M3 10.5 12 3l9 7.5v9a1.5 1.5 0 0 1-1.5 1.5h-5v-6h-5v6h-5A1.5 1.5 0 0 1 3 19.5v-9Z',
    search: 'm21 21-4.35-4.35M10.75 18a7.25 7.25 0 1 1 0-14.5 7.25 7.25 0 0 1 0 14.5Z',
    bell: 'M18 8a6 6 0 0 0-12 0c0 7-3 7-3 9h18c0-2-3-2-3-9ZM10 21h4',
    chat: 'M20 11.5a8 8 0 0 1-8 8 8.6 8.6 0 0 1-3.3-.65L4 20l1.15-4.3A8 8 0 1 1 20 11.5Z',
    profile: 'M20 21a8 8 0 0 0-16 0M12 11a4 4 0 1 0 0-8 4 4 0 0 0 0 8Z',
    more: 'M5 12h.01M12 12h.01M19 12h.01',
  }

  return <svg className="icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" aria-hidden="true"><path d={paths[name]} /></svg>
}

function Timeline() {
  const [posts, setPosts] = useState(initialPosts)
  const [draft, setDraft] = useState('')
  const [activeTab, setActiveTab] = useState('おすすめ')

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

  const toggleLike = (id: number) => {
    setPosts((current) => current.map((post) => post.id === id
      ? { ...post, liked: !post.liked, likes: post.likes + (post.liked ? -1 : 1) }
      : post))
  }

  return (
    <div className="app-shell">
      <aside className="sidebar">
        <div className="brand-mark">𝕏</div>
        <nav className="main-nav" aria-label="メインメニュー">
          <NavLink className="nav-item" to="/"><Icon name="home" /><span>ホーム</span></NavLink>
          <NavLink className="nav-item" to="/explore"><Icon name="search" /><span>話題を検索</span></NavLink>
          <NavLink className="nav-item" to="/notifications"><Icon name="bell" /><span>通知</span><span className="badge">3</span></NavLink>
          <NavLink className="nav-item" to="/messages"><Icon name="chat" /><span>チャット</span></NavLink>
          <NavLink className="nav-item" to="/profile"><Icon name="profile" /><span>プロフィール</span></NavLink>
          <NavLink className="nav-item" to="/more"><Icon name="more" /><span>もっと見る</span></NavLink>
        </nav>
        <button className="post-button" type="button" onClick={() => document.getElementById('composer')?.focus()}>ポストする</button>
        <button className="account-card" type="button">
          <span className="avatar avatar-blue">太</span>
          <span className="account-copy"><strong>田中 太郎</strong><small>@taro_tanaka</small></span>
          <span className="more">•••</span>
        </button>
      </aside>

      <main className="timeline" id="timeline">
        <header className="timeline-header">
          <h1>ホーム</h1>
          <div className="tabs" role="tablist">
            {['おすすめ', 'フォロー中'].map((tab) => (
              <button key={tab} className={activeTab === tab ? 'tab active' : 'tab'} onClick={() => setActiveTab(tab)} role="tab" aria-selected={activeTab === tab}>{tab}</button>
            ))}
          </div>
        </header>

        <form className="composer" onSubmit={publish}>
          <span className="avatar avatar-blue">太</span>
          <div className="composer-main">
            <textarea id="composer" value={draft} onChange={(event) => setDraft(event.target.value)} placeholder="いまどうしてる？" maxLength={140} rows={3} />
            <div className="composer-footer">
              <div className="composer-tools" aria-label="添付メニュー"><button type="button">▧</button><button type="button">◎</button><button type="button">☺</button></div>
              <span className="char-count">{draft.length}/140</span>
              <button className="publish-button" disabled={!draft.trim()} type="submit">ポストする</button>
            </div>
          </div>
        </form>

        <div className="feed">
          {posts.map((post) => (
            <article className="post" key={post.id}>
              <span className={`avatar avatar-blue ${post.avatarClass ?? ''}`}>{post.avatar}</span>
              <div className="post-content">
                <div className="post-meta"><strong>{post.name}</strong><span>{post.handle}</span><span>·</span><span>{post.time}</span><button className="post-more" aria-label="その他">•••</button></div>
                <p>{post.body}</p>
                <div className="post-actions">
                  <button aria-label="返信"><span className="action-icon">◌</span><span>{post.replies}</span></button>
                  <button aria-label="リポスト"><span className="action-icon">↻</span><span>{post.reposts}</span></button>
                  <button className={post.liked ? 'liked' : ''} aria-label="いいね" onClick={() => toggleLike(post.id)}><span className="action-icon">♡</span><span>{post.likes}</span></button>
                  <button aria-label="共有"><span className="action-icon">⇧</span></button>
                </div>
              </div>
            </article>
          ))}
        </div>
      </main>

      <aside className="right-rail">
        <label className="search-box"><Icon name="search" /><input placeholder="検索" /></label>
        <section className="rail-card">
          <h2>いまどうしてる？</h2>
          <div className="trend"><small>日本のトレンド</small><strong>#ソフトウェア開発</strong><small>12,438件のポスト</small></div>
          <div className="trend"><small>テクノロジー · トレンド</small><strong>TypeScript</strong><small>8,210件のポスト</small></div>
          <button className="show-more" type="button">さらに表示</button>
        </section>
        <section className="rail-card follow-card">
          <h2>おすすめユーザー</h2>
          {['佐藤 翔', 'プロダクト開発部'].map((name, index) => <div className="follow-row" key={name}><span className={`avatar ${index ? 'avatar-pink' : 'avatar-green'}`}>{index ? '開' : '翔'}</span><span className="account-copy"><strong>{name}</strong><small>@{index ? 'product_team' : 'sho_sato'}</small></span><button type="button" className="follow-button">フォロー</button></div>)}
        </section>
        <footer className="footer-links">利用規約　プライバシーポリシー　© 2026 ENG-1103</footer>
      </aside>
    </div>
  )
}

function Placeholder({ title }: { title: string }) {
  return (
    <div className="placeholder-page">
      <div className="brand-mark">𝕏</div>
      <h1>{title}</h1>
      <p>この画面は今後実装予定です。</p>
      <NavLink className="back-home" to="/">ホームに戻る</NavLink>
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
