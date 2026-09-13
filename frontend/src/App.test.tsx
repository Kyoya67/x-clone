import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import App from './App'

function requestURL(input: RequestInfo | URL) {
  return input instanceof Request ? input.url : input.toString()
}

beforeEach(() => {
  vi.stubGlobal(
    'fetch',
    vi.fn((input: RequestInfo | URL) => {
      const url = requestURL(input)
      if (url === '/auth/me') {
        return Promise.resolve(
          new Response(
            JSON.stringify({
              id: '00000000-0000-0000-0000-000000000001',
              handle: 'taro_tanaka',
              displayName: '田中 太郎',
              bio: '',
              createdAt: '2026-09-09T00:00:00Z',
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
                  id: 'post-1',
                  content: 'テスト投稿です。',
                  createdAt: '2026-09-09T00:00:00Z',
                  author: {
                    id: '00000000-0000-0000-0000-000000000003',
                    handle: 'ken_yamamoto',
                    displayName: '山本 健',
                  },
                },
              ],
            }),
            { status: 200, headers: { 'Content-Type': 'application/json' } },
          ),
        )
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
    expect(screen.getByText('@ken_yamamoto')).toBeInTheDocument()
    expect(screen.getByText('山本 健')).toBeInTheDocument()
  })
})
