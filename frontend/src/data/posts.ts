export type Post = {
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

export const initialPosts: Post[] = [
  { id: 1, name: '田中 太郎', handle: '@taro_tanaka', body: '今日はチームでタイムライン機能の設計をしました。小さく作って検証するのが気持ちいい。', time: '2時間', avatar: '太', likes: 18, replies: 3, reposts: 4 },
  { id: 2, name: '鈴木 花子', handle: '@hanako_s', body: '新しいサービスの最初の一歩。ユーザーが迷わず使える体験を大切にしたい。', time: '5時間', avatar: '花', avatarClass: 'avatar-pink', likes: 42, replies: 8, reposts: 9 },
  { id: 3, name: '山本 健', handle: '@ken_yamamoto', body: 'コードレビューで設計の意図を共有する時間が好きです。', time: '昨日', avatar: '健', avatarClass: 'avatar-green', likes: 27, replies: 2, reposts: 5 },
]
