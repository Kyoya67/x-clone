import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { describe, expect, it } from 'vitest'
import App from './App'

describe('App routes', () => {
  it.each([
    ['/', 'ホーム'],
    ['/explore', '話題を検索'],
    ['/notifications', '通知'],
    ['/messages', 'チャット'],
    ['/profile', 'プロフィール'],
    ['/more', 'もっと見る'],
  ])('renders %s at %s', (path, title) => {
    window.history.pushState({}, '', path)
    render(<App />)
    expect(screen.getByRole('heading', { name: title })).toBeInTheDocument()
  })

  it('navigates to a profile when clicking an author name', async () => {
    const user = userEvent.setup()
    window.history.pushState({}, '', '/')
    render(<App />)
    await user.click(screen.getByRole('link', { name: '山本 健' }))
    expect(screen.getByRole('heading', { name: 'プロフィール' })).toBeInTheDocument()
    expect(screen.getByText('@ken_yamamoto')).toBeInTheDocument()
    expect(screen.getByText('山本 健')).toBeInTheDocument()
  })
})
