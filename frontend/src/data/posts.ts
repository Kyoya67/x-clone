export type Post = {
  id: string
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
  {
    id: 'mock-1',
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
    id: 'mock-2',
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
    id: 'mock-3',
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

export const followingPosts: Post[] = [
  {
    id: 'mock-4',
    name: '佐藤 翔',
    handle: '@sho_sato',
    body: '今日は新しい機能のプロトタイプを作っています。小さく試して改善するのが楽しい。',
    time: '30分',
    avatar: '翔',
    avatarClass: 'avatar-green',
    likes: 12,
    replies: 1,
    reposts: 2,
  },
  {
    id: 'mock-5',
    name: '鈴木 花子',
    handle: '@hanako_s',
    body: 'チームで仕様を確認しました。使う人の目線で考えることを忘れないようにしたい。',
    time: '1時間',
    avatar: '花',
    avatarClass: 'avatar-pink',
    likes: 31,
    replies: 4,
    reposts: 6,
  },
]
