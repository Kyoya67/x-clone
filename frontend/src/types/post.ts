export type Post = {
  id: string
  authorId: string
  name: string
  handle: string
  body: string
  time: string
  avatar: string
  likes: number
  replies: number
  reposts: number
  liked?: boolean
}
