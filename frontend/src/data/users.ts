export type User = {
  id: string
  handle: string
  displayName: string
  bio: string
  avatar: string
  avatarClass?: string
}

export const users: User[] = [
  {
    id: '00000000-0000-0000-0000-000000000001',
    handle: '@taro_tanaka',
    displayName: '田中 太郎',
    bio: 'Webアプリケーションをつくっています。',
    avatar: '太',
  },
  {
    id: '00000000-0000-0000-0000-000000000002',
    handle: '@hanako_s',
    displayName: '鈴木 花子',
    bio: 'デザインとプロダクトづくりが好きです。',
    avatar: '花',
    avatarClass: 'avatar-pink',
  },
  {
    id: '00000000-0000-0000-0000-000000000003',
    handle: '@ken_yamamoto',
    displayName: '山本 健',
    bio: 'バックエンドと開発体験を改善しています。',
    avatar: '健',
    avatarClass: 'avatar-green',
  },
  {
    id: '00000000-0000-0000-0000-000000000004',
    handle: '@sho_sato',
    displayName: '佐藤 翔',
    bio: '小さく試して、学びながら開発しています。',
    avatar: '翔',
    avatarClass: 'avatar-green',
  },
  {
    id: '00000000-0000-0000-0000-000000000005',
    handle: '@product_team',
    displayName: 'プロダクト開発部',
    bio: 'チームでよりよいプロダクトをつくります。',
    avatar: '開',
    avatarClass: 'avatar-pink',
  },
]

export function findUserByHandle(handle: string) {
  return users.find((user) => user.handle === handle)
}
