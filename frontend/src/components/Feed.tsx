import { Post } from '../data/posts'

export function Feed({
  posts,
  onToggleLike,
}: {
  posts: Post[]
  onToggleLike: (id: number) => void
}) {
  return (
    <div className="feed">
      {posts.map((post) => (
        <article className="post" key={post.id}>
          <span className={`avatar avatar-blue ${post.avatarClass ?? ''}`}>{post.avatar}</span>
          <div className="post-content">
            <div className="post-meta">
              <strong>{post.name}</strong>
              <span>{post.handle}</span>
              <span>·</span>
              <span>{post.time}</span>
              <button className="post-more" aria-label="その他">
                •••
              </button>
            </div>
            <p>{post.body}</p>
            <div className="post-actions">
              <button aria-label="返信">
                <span className="action-icon">◌</span>
                <span>{post.replies}</span>
              </button>
              <button aria-label="リポスト">
                <span className="action-icon">↻</span>
                <span>{post.reposts}</span>
              </button>
              <button
                className={post.liked ? 'liked' : ''}
                aria-label="いいね"
                onClick={() => onToggleLike(post.id)}
              >
                <span className="action-icon">♡</span>
                <span>{post.likes}</span>
              </button>
              <button aria-label="共有">
                <span className="action-icon">⇧</span>
              </button>
            </div>
          </div>
        </article>
      ))}
    </div>
  )
}
