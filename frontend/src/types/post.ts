export type Post = {
  id: string
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
