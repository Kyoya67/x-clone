export type CurrentUserResponse = {
  id: string
  handle: string
  displayName: string
  bio: string
  createdAt: string
}

export async function fetchCurrentUser(): Promise<CurrentUserResponse | null> {
  const response = await fetch('/auth/me')
  if (response.status === 401) return null
  if (!response.ok) throw new Error('ログイン状態の確認に失敗しました')
  return response.json() as Promise<CurrentUserResponse>
}

export async function logout() {
  const response = await fetch('/auth/logout', { method: 'POST' })
  if (!response.ok) throw new Error('ログアウトに失敗しました')
}
