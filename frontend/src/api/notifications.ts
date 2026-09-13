export type Notification = {
  id: string
  type: 'follow' | 'like'
  actor: {
    id: string
    handle: string
    displayName: string
  }
  postId?: string
  createdAt: string
}

type NotificationsResponse = {
  notifications: Notification[]
}

export async function fetchNotifications(): Promise<Notification[]> {
  const response = await fetch('/api/notifications')
  if (!response.ok) throw new Error('通知の取得に失敗しました')

  const body = (await response.json()) as NotificationsResponse
  return body.notifications
}
