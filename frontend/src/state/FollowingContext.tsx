import { createContext, ReactNode, useContext, useEffect, useState } from 'react'
import { fetchFollowingUserIDs, followUser, unfollowUser } from '../api/follows'
import { findUserByHandle, findUserByID } from '../data/users'

type FollowingContextValue = {
  followingHandles: string[]
  toggleFollowing: (handle: string) => Promise<void>
  isFollowing: (handle: string) => boolean
  isUpdating: (handle: string) => boolean
  error: string
}
const FollowingContext = createContext<FollowingContextValue | null>(null)

export function FollowingProvider({ children }: { children: ReactNode }) {
  const [followingHandles, setFollowingHandles] = useState<string[]>([])
  const [updatingHandles, setUpdatingHandles] = useState<string[]>([])
  const [error, setError] = useState('')

  useEffect(() => {
    const loadFollowing = async () => {
      try {
        const userIDs = await fetchFollowingUserIDs()
        const handles = userIDs
          .map((userID) => findUserByID(userID)?.handle)
          .filter((handle): handle is string => handle !== undefined)
        setFollowingHandles(handles)
      } catch {
        setError('フォロー状態の取得に失敗しました。時間をおいて再度お試しください。')
      }
    }

    void loadFollowing()
  }, [])

  const toggleFollowing = async (handle: string) => {
    const user = findUserByHandle(handle)
    if (!user || updatingHandles.includes(handle)) return

    const following = followingHandles.includes(handle)
    setError('')
    setUpdatingHandles((current) => [...current, handle])
    try {
      if (following) {
        await unfollowUser(user.id)
        setFollowingHandles((current) => current.filter((item) => item !== handle))
      } else {
        await followUser(user.id)
        setFollowingHandles((current) => [...current, handle])
      }
    } catch {
      setError('フォロー操作に失敗しました。時間をおいて再度お試しください。')
    } finally {
      setUpdatingHandles((current) => current.filter((item) => item !== handle))
    }
  }

  const isFollowing = (handle: string) => followingHandles.includes(handle)
  const isUpdating = (handle: string) => updatingHandles.includes(handle)

  return (
    <FollowingContext.Provider
      value={{ followingHandles, toggleFollowing, isFollowing, isUpdating, error }}
    >
      {children}
    </FollowingContext.Provider>
  )
}

export function useFollowing() {
  const context = useContext(FollowingContext)
  if (!context) throw new Error('useFollowing must be used within FollowingProvider')
  return context
}
