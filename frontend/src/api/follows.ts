async function requestFollow(userID: string, method: 'PUT' | 'DELETE') {
  const response = await fetch(`/api/users/${userID}/follow`, { method })
  if (!response.ok) throw new Error('フォロー操作に失敗しました')
}

export async function fetchFollowingUserIDs() {
  const response = await fetch('/api/me/following')
  if (!response.ok) throw new Error('フォロー状態の取得に失敗しました')
  const body: { userIds: string[] } = await response.json()
  return body.userIds
}

export function followUser(userID: string) {
  return requestFollow(userID, 'PUT')
}

export function unfollowUser(userID: string) {
  return requestFollow(userID, 'DELETE')
}
