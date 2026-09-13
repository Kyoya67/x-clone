import { FormEvent, useEffect, useState } from 'react'
import { fetchTimeline } from '../api/timeline'
import { Feed } from '../components/Feed'
import { ContentPage } from './ContentPage'
import { currentUser } from '../config/currentUser'
import { useOptionalAuth } from '../state/AuthContext'
import { Post } from '../types/post'

export function ProfilePage() {
  const auth = useOptionalAuth()
  const user = auth?.user
  const displayName = user?.displayName ?? currentUser.displayName
  const handle = user?.handle ?? currentUser.handle.replace(/^@/, '')
  const bio = user?.bio ?? currentUser.bio
  const avatar = displayName.slice(0, 1) || 'U'
  const [form, setForm] = useState({ handle, displayName, bio })
  const [message, setMessage] = useState('')
  const [isEditing, setIsEditing] = useState(user?.needsProfileSetup ?? false)
  const [isSaving, setIsSaving] = useState(false)
  const [posts, setPosts] = useState<Post[]>([])
  const [isLoadingPosts, setIsLoadingPosts] = useState(true)
  const [postsError, setPostsError] = useState('')

  useEffect(() => {
    setForm({ handle, displayName, bio })
    setIsEditing(user?.needsProfileSetup ?? false)
  }, [handle, displayName, bio, user?.needsProfileSetup])

  useEffect(() => {
    let cancelled = false

    const loadPosts = async () => {
      setIsLoadingPosts(true)
      setPostsError('')
      try {
        const timelinePosts = await fetchTimeline('for-you')
        if (!cancelled) {
          setPosts(timelinePosts.filter((post) => post.handle === `@${handle}`))
        }
      } catch {
        if (!cancelled) {
          setPosts([])
          setPostsError('ポストの取得に失敗しました。時間をおいて再度お試しください。')
        }
      } finally {
        if (!cancelled) setIsLoadingPosts(false)
      }
    }

    void loadPosts()
    return () => {
      cancelled = true
    }
  }, [handle])

  const saveProfile = async (event: FormEvent) => {
    event.preventDefault()
    setMessage('')
    if (!/^[A-Za-z0-9_]{1,13}$/.test(form.handle)) {
      setMessage('ハンドルは半角英数字と_のみ、13文字以下で入力してください。')
      return
    }
    if (!form.displayName.trim()) {
      setMessage('表示名を入力してください。')
      return
    }
    setIsSaving(true)
    try {
      await auth?.updateProfile({
        handle: form.handle.trim(),
        displayName: form.displayName.trim(),
        bio: form.bio.trim(),
      })
      setIsEditing(false)
      setMessage('プロフィールを更新しました。')
    } catch {
      setMessage('プロフィール更新に失敗しました。時間をおいて再度お試しください。')
    } finally {
      setIsSaving(false)
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
    <ContentPage title="プロフィール">
      <div className="profile-card">
        <span className="avatar avatar-blue profile-avatar">{avatar}</span>
        {isEditing ? (
          <form className="profile-form" onSubmit={saveProfile}>
            <label>
              ハンドル
              <input
                value={form.handle}
                onChange={(event) =>
                  setForm((current) => ({ ...current, handle: event.target.value }))
                }
                maxLength={13}
                pattern="[A-Za-z0-9_]{1,13}"
                required
              />
            </label>
            <label>
              表示名
              <input
                value={form.displayName}
                onChange={(event) =>
                  setForm((current) => ({ ...current, displayName: event.target.value }))
                }
                maxLength={100}
                required
              />
            </label>
            <label>
              自己紹介
              <textarea
                value={form.bio}
                onChange={(event) =>
                  setForm((current) => ({ ...current, bio: event.target.value }))
                }
                maxLength={160}
                rows={3}
              />
            </label>
            {message && (
              <p role={message.includes('失敗') || message.includes('入力') ? 'alert' : 'status'}>
                {message}
              </p>
            )}
            <button className="follow-button" disabled={isSaving} type="submit">
              保存
            </button>
          </form>
        ) : (
          <>
            <button className="follow-button" type="button" onClick={() => setIsEditing(true)}>
              編集
            </button>
            <h2>{displayName}</h2>
            <p>@{handle}</p>
            <p>{bio}</p>
            {message && <p role="status">{message}</p>}
          </>
        )}
      </div>
      {postsError && <p role="alert">{postsError}</p>}
      {isLoadingPosts ? (
        <p role="status">ポストを読み込んでいます。</p>
      ) : posts.length === 0 ? (
        <p>表示する投稿はありません。</p>
      ) : (
        <Feed posts={posts} onToggleLike={toggleLike} currentUserHandle={handle} />
      )}
    </ContentPage>
  )
}
