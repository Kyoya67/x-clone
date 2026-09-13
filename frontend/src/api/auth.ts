export type CurrentUserResponse = {
  id: string
  handle: string
  displayName: string
  bio: string
  createdAt: string
  needsProfileSetup: boolean
}

export type UpdateProfileRequest = {
  handle: string
  displayName: string
  bio: string
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

export async function updateCurrentUser(
  request: UpdateProfileRequest,
): Promise<CurrentUserResponse> {
  const response = await fetch('/auth/me', {
    method: 'PATCH',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(request),
  })
  if (!response.ok) throw new Error('プロフィール更新に失敗しました')
  return response.json() as Promise<CurrentUserResponse>
}
