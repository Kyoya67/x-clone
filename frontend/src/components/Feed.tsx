import { NavLink } from 'react-router-dom'
import { Post } from '../types/post'

export function Feed({
  posts,
  onToggleLike,
  currentUserHandle,
}: {
  posts: Post[]
  onToggleLike: (id: string) => void
  currentUserHandle?: string
}) {
  return (
    <div className="feed">
      {posts.map((post) => {
        const authorHandle = post.handle.replace(/^@/, '')
        const profilePath =
          authorHandle === currentUserHandle ? '/profile' : `/users/${authorHandle}`

        return (
          <article className="post" key={post.id}>
            <span className={`avatar ${post.avatarClass ?? 'avatar-blue'}`}>{post.avatar}</span>
            <div className="post-content">
              <div className="post-meta">
                <NavLink className="post-author" to={profilePath}>
                  <strong>{post.name}</strong>
                </NavLink>
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
        )
      })}
    </div>
  )
}
