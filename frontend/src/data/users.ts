import { currentUser, showDemoUsers } from '../config/currentUser'

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
    ...currentUser,
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

export const visibleUsers = showDemoUsers ? users : [users[0]]

export function findUserByHandle(handle: string) {
  return visibleUsers.find((user) => user.handle === handle)
}

export function findUserByID(id: string) {
  return visibleUsers.find((user) => user.id === id)
}
