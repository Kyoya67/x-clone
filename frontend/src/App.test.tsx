import { render, screen, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import App from './App'

function requestURL(input: RequestInfo | URL) {
  return input instanceof Request ? input.url : input.toString()
}

beforeEach(() => {
  vi.stubGlobal(
    'fetch',
    vi.fn((input: RequestInfo | URL, init?: RequestInit) => {
      const url = requestURL(input)
      if (url === '/auth/me') {
        if ((input instanceof Request && input.method === 'PATCH') || init?.method === 'PATCH') {
          return Promise.resolve(
            new Response(
              JSON.stringify({
                id: '00000000-0000-0000-0000-000000000001',
                handle: 'kyoya_dev',
                displayName: 'dev kyoya',
                bio: '',
                createdAt: '2026-09-09T00:00:00Z',
                needsProfileSetup: false,
              }),
              { status: 200, headers: { 'Content-Type': 'application/json' } },
            ),
          )
        }
        return Promise.resolve(
          new Response(
            JSON.stringify({
              id: '00000000-0000-0000-0000-000000000001',
              handle: 'taro_tanaka',
              displayName: '田中 太郎',
              bio: '',
              createdAt: '2026-09-09T00:00:00Z',
              needsProfileSetup: false,
            }),
            { status: 200, headers: { 'Content-Type': 'application/json' } },
          ),
        )
      }
      if (url === '/api/timeline?feed=for-you') {
        return Promise.resolve(
          new Response(
            JSON.stringify({
              posts: [
                {
                  id: 'post-0',
                  content: '自分の投稿です。',
                  createdAt: '2026-09-09T00:00:00Z',
                  author: {
                    id: '00000000-0000-0000-0000-000000000001',
                    handle: 'taro_tanaka',
                    displayName: '田中 太郎',
                  },
                },
                {
                  id: 'post-1',
                  content: 'テスト投稿です。',
                  createdAt: '2026-09-09T00:00:00Z',
                  author: {
                    id: '00000000-0000-0000-0000-000000000003',
                    handle: 'ken_yamamoto',
                    displayName: '山本 健',
                  },
                },
                {
                  id: 'post-2',
                  content: '佐藤さんの投稿です。',
                  createdAt: '2026-09-09T00:00:00Z',
                  author: {
                    id: '00000000-0000-0000-0000-000000000004',
                    handle: 'sho_sato',
                    displayName: '佐藤 翔',
                  },
                },
              ],
            }),
            { status: 200, headers: { 'Content-Type': 'application/json' } },
          ),
        )
      }
      if (url === '/api/notifications') {
        return Promise.resolve(
          new Response(
            JSON.stringify({
              notifications: [
                {
                  id: 'notification-1',
                  type: 'follow',
                  actor: {
                    id: '00000000-0000-0000-0000-000000000004',
                    handle: 'sho_sato',
                    displayName: '佐藤 翔',
                  },
                  createdAt: '2026-09-09T00:00:00Z',
                },
              ],
            }),
            { status: 200, headers: { 'Content-Type': 'application/json' } },
          ),
        )
      }
      if (url.endsWith('/like')) {
        return Promise.resolve(new Response(null, { status: 204 }))
      }
      return Promise.resolve(
        new Response(JSON.stringify({ userIds: [] }), {
          status: 200,
          headers: { 'Content-Type': 'application/json' },
        }),
      )
    }),
  )
})

afterEach(() => {
  localStorage.clear()
  document.documentElement.removeAttribute('data-theme')
  vi.unstubAllGlobals()
})

describe('App routes', () => {
  it.each([
    ['/', 'ホーム'],
    ['/explore', '話題を検索'],
    ['/notifications', '通知'],
    ['/messages', 'チャット'],
    ['/profile', 'プロフィール'],
    ['/more', 'もっと見る'],
  ])('renders %s at %s', async (path, title) => {
    window.history.pushState({}, '', path)
    render(<App />)
    expect(await screen.findByRole('heading', { name: title })).toBeInTheDocument()
  })

  it('navigates to a profile when clicking an author name', async () => {
    const user = userEvent.setup()
    window.history.pushState({}, '', '/')
    render(<App />)
    await user.click(await screen.findByRole('link', { name: '山本 健' }))
    expect(screen.getByRole('heading', { name: 'プロフィール' })).toBeInTheDocument()
    expect(screen.getAllByText('@ken_yamamoto')).not.toHaveLength(0)
    expect(screen.getAllByText('山本 健')).not.toHaveLength(0)
  })

  it('shows only posts from the selected user profile', async () => {
    window.history.pushState({}, '', '/users/sho_sato')
    render(<App />)

    expect(await screen.findAllByText('佐藤 翔')).not.toHaveLength(0)
    expect(await screen.findByText('佐藤さんの投稿です。')).toBeInTheDocument()
    expect(screen.queryByText('テスト投稿です。')).not.toBeInTheDocument()
    expect(
      screen.queryByText('ユーザーのプロフィールとポストを表示します。'),
    ).not.toBeInTheDocument()
  })

  it('navigates to the own profile page when clicking the current user post', async () => {
    const user = userEvent.setup()
    window.history.pushState({}, '', '/')
    render(<App />)

    await user.click(await screen.findByRole('link', { name: '田中 太郎' }))

    expect(window.location.pathname).toBe('/profile')
    expect(await screen.findByText('自分の投稿です。')).toBeInTheDocument()
  })

  it('follows the selected profile user instead of a fixed user', async () => {
    const user = userEvent.setup()
    window.history.pushState({}, '', '/users/sho_sato')
    render(<App />)

    await screen.findByText('佐藤さんの投稿です。')
    const profileCard = screen
      .getByText('小さく試して、学びながら開発しています。')
      .closest('.profile-card')
    if (!profileCard) throw new Error('profile card was not found')
    await user.click(within(profileCard as HTMLElement).getByRole('button', { name: 'フォロー' }))

    expect(fetch).toHaveBeenCalledWith('/api/users/00000000-0000-0000-0000-000000000004/follow', {
      method: 'PUT',
    })
  })

  it('shows notifications from the API', async () => {
    window.history.pushState({}, '', '/notifications')
    render(<App />)

    expect(await screen.findByText(/さんがあなたをフォローしました。/)).toBeInTheDocument()
  })

  it('shows the account menu and logs out from the current account', async () => {
    const user = userEvent.setup()
    window.history.pushState({}, '', '/')
    render(<App />)

    expect(
      screen.queryByRole('menuitem', { name: '既存のアカウントを追加' }),
    ).not.toBeInTheDocument()
    await user.click(await screen.findByRole('button', { name: /田中 太郎/ }))
    expect(
      screen.queryByRole('menuitem', { name: '既存のアカウントを追加' }),
    ).not.toBeInTheDocument()
    await user.click(screen.getByRole('menuitem', { name: '@taro_tanakaからログアウト' }))

    expect(fetch).toHaveBeenCalledWith('/auth/logout', { method: 'POST' })
  })

  it('switches between light and dark theme', async () => {
    const user = userEvent.setup()
    window.history.pushState({}, '', '/')
    render(<App />)

    await user.click(await screen.findByRole('button', { name: 'ダークモード' }))
    expect(document.documentElement.dataset.theme).toBe('dark')
    expect(localStorage.getItem('theme')).toBe('dark')

    await user.click(screen.getByRole('button', { name: 'ライトモード' }))
    expect(document.documentElement.dataset.theme).toBe('light')
    expect(localStorage.getItem('theme')).toBe('light')
  })
})
