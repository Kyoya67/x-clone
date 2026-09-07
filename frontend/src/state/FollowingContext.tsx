import { createContext, ReactNode, useContext, useState } from 'react'

type FollowingContextValue = {
  followingHandles: string[]
  toggleFollowing: (handle: string) => void
  isFollowing: (handle: string) => boolean
}
const FollowingContext = createContext<FollowingContextValue | null>(null)

export function FollowingProvider({ children }: { children: ReactNode }) {
  const [followingHandles, setFollowingHandles] = useState(['@sho_sato', '@hanako_s'])
  const toggleFollowing = (handle: string) =>
    setFollowingHandles((current) =>
      current.includes(handle) ? current.filter((item) => item !== handle) : [...current, handle],
    )
  const isFollowing = (handle: string) => followingHandles.includes(handle)
  return (
    <FollowingContext.Provider value={{ followingHandles, toggleFollowing, isFollowing }}>
      {children}
    </FollowingContext.Provider>
  )
}

export function useFollowing() {
  const context = useContext(FollowingContext)
  if (!context) throw new Error('useFollowing must be used within FollowingProvider')
  return context
}
