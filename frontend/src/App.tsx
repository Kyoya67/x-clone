import { BrowserRouter, Route, Routes } from 'react-router-dom'
import { ExplorePage } from './pages/ExplorePage'
import { MessagesPage } from './pages/MessagesPage'
import { MorePage } from './pages/MorePage'
import { NotificationsPage } from './pages/NotificationsPage'
import { NotFoundPage } from './pages/NotFoundPage'
import { ProfilePage } from './pages/ProfilePage'
import { TimelinePage } from './pages/TimelinePage'
import { UserProfilePage } from './pages/UserProfilePage'
import { FollowingProvider } from './state/FollowingContext'

function App() {
  return (
    <BrowserRouter>
      <FollowingProvider>
        <Routes>
          <Route path="/" element={<TimelinePage />} />
          <Route path="/explore" element={<ExplorePage />} />
          <Route path="/notifications" element={<NotificationsPage />} />
          <Route path="/messages" element={<MessagesPage />} />
          <Route path="/profile" element={<ProfilePage />} />
          <Route path="/users/:handle" element={<UserProfilePage />} />
          <Route path="/more" element={<MorePage />} />
          <Route path="*" element={<NotFoundPage />} />
        </Routes>
      </FollowingProvider>
    </BrowserRouter>
  )
}

export default App
