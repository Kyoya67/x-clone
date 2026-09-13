import { createContext, ReactNode, useContext, useEffect, useState } from 'react'
import { fetchFollowingUserIDs, followUser, unfollowUser } from '../api/follows'

type FollowingContextValue = {
  followingUserIDs: string[]
  toggleFollowing: (userID: string) => Promise<void>
  isFollowing: (userID: string) => boolean
  isUpdating: (userID: string) => boolean
  error: string
}
const FollowingContext = createContext<FollowingContextValue | null>(null)

export function FollowingProvider({ children }: { children: ReactNode }) {
  const [followingUserIDs, setFollowingUserIDs] = useState<string[]>([])
  const [updatingUserIDs, setUpdatingUserIDs] = useState<string[]>([])
  const [error, setError] = useState('')

  useEffect(() => {
    const loadFollowing = async () => {
      try {
        const userIDs = await fetchFollowingUserIDs()
        setFollowingUserIDs(userIDs)
      } catch {
        setError('フォロー状態の取得に失敗しました。時間をおいて再度お試しください。')
      }
    }

    void loadFollowing()
  }, [])

  const toggleFollowing = async (userID: string) => {
    if (updatingUserIDs.includes(userID)) return

    const following = followingUserIDs.includes(userID)
    setError('')
    setUpdatingUserIDs((current) => [...current, userID])
    try {
      if (following) {
        await unfollowUser(userID)
        setFollowingUserIDs((current) => current.filter((item) => item !== userID))
      } else {
        await followUser(userID)
        setFollowingUserIDs((current) => [...current, userID])
      }
    } catch {
      setError('フォロー操作に失敗しました。時間をおいて再度お試しください。')
    } finally {
      setUpdatingUserIDs((current) => current.filter((item) => item !== userID))
    }
  }

  const isFollowing = (userID: string) => followingUserIDs.includes(userID)
  const isUpdating = (userID: string) => updatingUserIDs.includes(userID)

  return (
    <FollowingContext.Provider
      value={{ followingUserIDs, toggleFollowing, isFollowing, isUpdating, error }}
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
