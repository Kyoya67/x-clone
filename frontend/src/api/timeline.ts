import { Post } from '../types/post'
import { avatarColorClass } from '../utils/avatarColor'

type TimelinePostResponse = {
  id: string
  content: string
  createdAt: string
  author: {
    id: string
    handle: string
    displayName: string
  }
}

type TimelineResponse = {
  posts: TimelinePostResponse[]
}

export type TimelineFeed = 'for-you' | 'following'

export async function fetchTimeline(feed: TimelineFeed): Promise<Post[]> {
  const response = await fetch(`/api/timeline?feed=${feed}`)
  if (!response.ok) throw new Error('タイムラインの取得に失敗しました')

  const body = (await response.json()) as TimelineResponse
  return body.posts.map(toPost)
}

function toPost(post: TimelinePostResponse): Post {
  return {
    id: post.id,
    authorId: post.author.id,
    name: post.author.displayName,
    handle: `@${post.author.handle}`,
    body: post.content,
    time: formatPostTime(post.createdAt),
    avatar: post.author.displayName.slice(0, 1),
    avatarClass: avatarColorClass(post.author.id || post.author.handle),
    likes: 0,
    replies: 0,
    reposts: 0,
  }
}

function formatPostTime(createdAt: string) {
  const date = new Date(createdAt)
  if (Number.isNaN(date.getTime())) return ''

  const elapsedMinutes = Math.floor((Date.now() - date.getTime()) / 60_000)
  if (elapsedMinutes < 60) return `${Math.max(elapsedMinutes, 0)}分`
  if (elapsedMinutes < 24 * 60) return `${Math.floor(elapsedMinutes / 60)}時間`
  return `${date.getMonth() + 1}/${date.getDate()}`
}
