import { BrowserRouter, Route, Routes } from 'react-router-dom'
import { ExplorePage } from './pages/ExplorePage'
import { MessagesPage } from './pages/MessagesPage'
import { MorePage } from './pages/MorePage'
import { NotificationsPage } from './pages/NotificationsPage'
import { NotFoundPage } from './pages/NotFoundPage'
import { ProfilePage } from './pages/ProfilePage'
import { TimelinePage } from './pages/TimelinePage'

function App() {
  return (
    <BrowserRouter>
      <Routes>
        <Route path="/" element={<TimelinePage />} />
        <Route path="/explore" element={<ExplorePage />} />
        <Route path="/notifications" element={<NotificationsPage />} />
        <Route path="/messages" element={<MessagesPage />} />
        <Route path="/profile" element={<ProfilePage />} />
        <Route path="/more" element={<MorePage />} />
        <Route path="*" element={<NotFoundPage />} />
      </Routes>
    </BrowserRouter>
  )
}

export default App
