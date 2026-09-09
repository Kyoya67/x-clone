export type CreatePostResponse = {
  id: string
  authorId: string
  content: string
  createdAt: string
}

export async function createPost(content: string): Promise<CreatePostResponse> {
  const response = await fetch('/api/posts', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ content }),
  })

  if (!response.ok) throw new Error('投稿に失敗しました')
  return response.json() as Promise<CreatePostResponse>
}
