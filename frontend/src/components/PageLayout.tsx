import { ReactNode } from 'react'
import { RightRail } from './RightRail'
import { Sidebar } from './Sidebar'

export function PageLayout({ children }: { children: ReactNode }) {
  return (
    <div className="app-shell">
      <Sidebar />
      <main className="page-content">{children}</main>
      <RightRail />
    </div>
  )
}
