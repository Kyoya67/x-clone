async function requestLike(postID: string, method: 'PUT' | 'DELETE') {
  const response = await fetch(`/api/posts/${postID}/like`, { method })
  if (!response.ok) throw new Error('いいね操作に失敗しました')
}

export function likePost(postID: string) {
  return requestLike(postID, 'PUT')
}

export function unlikePost(postID: string) {
  return requestLike(postID, 'DELETE')
}
