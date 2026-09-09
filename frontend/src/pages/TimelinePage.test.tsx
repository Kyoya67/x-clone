import { render, screen, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { MemoryRouter } from 'react-router-dom'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { TimelinePage } from './TimelinePage'
import { FollowingProvider } from '../state/FollowingContext'

const initialTimelineResponse = {
  posts: [
    {
      id: 'post-1',
      content: 'タイムラインから取得した投稿です。',
      createdAt: '2026-09-09T00:00:00Z',
      author: {
        id: '00000000-0000-0000-0000-000000000001',
        handle: 'taro_tanaka',
        displayName: '田中 太郎',
      },
    },
    {
      id: 'post-2',
      content: 'フォロー中ユーザーの投稿です。',
      createdAt: '2026-09-08T00:00:00Z',
      author: {
        id: '00000000-0000-0000-0000-000000000002',
        handle: 'hanako_s',
        displayName: '鈴木 花子',
      },
    },
  ],
}

const followingTimelineResponse = {
  posts: [initialTimelineResponse.posts[1]],
}

const postResponse = {
  id: 'post-from-api',
  authorId: '00000000-0000-0000-0000-000000000001',
  content: 'テスト投稿です',
  createdAt: '2026-09-09T00:00:00Z',
}

const timelineAfterPostingResponse = {
  posts: [
    {
      id: 'post-from-api',
      content: 'テスト投稿です',
      createdAt: '2026-09-09T00:00:00Z',
      author: {
        id: '00000000-0000-0000-0000-000000000001',
        handle: 'taro_tanaka',
        displayName: '田中 太郎',
      },
    },
    ...initialTimelineResponse.posts,
  ],
}

function requestURL(input: RequestInfo | URL) {
  return input instanceof Request ? input.url : input.toString()
}

function jsonResponse(body: unknown, status = 200) {
  return new Response(JSON.stringify(body), {
    status,
    headers: { 'Content-Type': 'application/json' },
  })
}

function createFetchMock() {
  let forYouRequestCount = 0

  return vi.fn((input: RequestInfo | URL) => {
    const url = requestURL(input)
    if (url === '/api/me/following') {
      return Promise.resolve(
        jsonResponse({
          userIds: [
            '00000000-0000-0000-0000-000000000002',
            '00000000-0000-0000-0000-000000000004',
          ],
        }),
      )
    }
    if (url === '/api/timeline?feed=following') {
      return Promise.resolve(jsonResponse(followingTimelineResponse))
    }
    if (url === '/api/timeline?feed=for-you') {
      forYouRequestCount += 1
      const response =
        forYouRequestCount === 1 ? initialTimelineResponse : timelineAfterPostingResponse
      return Promise.resolve(jsonResponse(response))
    }
    if (url === '/api/posts') {
      return Promise.resolve(jsonResponse(postResponse, 201))
    }
    return Promise.resolve(new Response(null, { status: 404 }))
  })
}

beforeEach(() => {
  vi.stubGlobal('fetch', createFetchMock())
})

afterEach(() => {
  vi.unstubAllGlobals()
})

function renderTimeline() {
  return render(
    <MemoryRouter>
      <FollowingProvider>
        <TimelinePage />
      </FollowingProvider>
    </MemoryRouter>,
  )
}

describe('TimelinePage', () => {
  it('loads and displays the for-you timeline from the API', async () => {
    renderTimeline()

    expect(await screen.findByText('タイムラインから取得した投稿です。')).toBeInTheDocument()
    expect(fetch).toHaveBeenCalledWith('/api/timeline?feed=for-you')
  })

  it('loads the following timeline when the tab is selected', async () => {
    const user = userEvent.setup()
    renderTimeline()

    await screen.findByText('タイムラインから取得した投稿です。')
    await user.click(screen.getByRole('tab', { name: 'フォロー中' }))

    expect(await screen.findByText('フォロー中ユーザーの投稿です。')).toBeInTheDocument()
    expect(screen.queryByText('タイムラインから取得した投稿です。')).not.toBeInTheDocument()
    expect(fetch).toHaveBeenCalledWith('/api/timeline?feed=following')
  })

  it('reloads the timeline after posting', async () => {
    const user = userEvent.setup()
    renderTimeline()

    await screen.findByText('タイムラインから取得した投稿です。')
    const composer = screen.getByPlaceholderText('いまどうしてる？')
    await user.type(composer, 'テスト投稿です')
    await user.click(
      within(composer.closest('form') as HTMLFormElement).getByRole('button', {
        name: 'ポストする',
      }),
    )

    expect(await screen.findByText('テスト投稿です')).toBeInTheDocument()
    expect(fetch).toHaveBeenCalledWith('/api/posts', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ content: 'テスト投稿です' }),
    })
  })

  it('displays an error when the timeline API fails', async () => {
    vi.mocked(fetch).mockImplementation((input: RequestInfo | URL) => {
      const url = requestURL(input)
      if (url.startsWith('/api/timeline')) {
        return Promise.resolve(new Response(null, { status: 500 }))
      }
      return Promise.resolve(jsonResponse({ userIds: [] }))
    })
    renderTimeline()

    expect(await screen.findByRole('alert')).toHaveTextContent(
      'タイムラインの取得に失敗しました。時間をおいて再度お試しください。',
    )
  })

  it('displays an error when the post API fails', async () => {
    const user = userEvent.setup()
    vi.mocked(fetch).mockImplementation((input: RequestInfo | URL) => {
      const url = requestURL(input)
      if (url === '/api/posts') return Promise.resolve(new Response(null, { status: 500 }))
      if (url.startsWith('/api/timeline')) return Promise.resolve(jsonResponse(initialTimelineResponse))
      return Promise.resolve(jsonResponse({ userIds: [] }))
    })
    renderTimeline()

    await screen.findByText('タイムラインから取得した投稿です。')
    const composer = screen.getByPlaceholderText('いまどうしてる？')
    await user.type(composer, '失敗する投稿')
    await user.click(
      within(composer.closest('form') as HTMLFormElement).getByRole('button', {
        name: 'ポストする',
      }),
    )

    expect(await screen.findByRole('alert')).toHaveTextContent(
      '投稿に失敗しました。時間をおいて再度お試しください。',
    )
  })

  it('updates the like count and state when liked', async () => {
    const user = userEvent.setup()
    renderTimeline()

    const post = (await screen.findByText('タイムラインから取得した投稿です。')).closest(
      '.post',
    ) as HTMLElement
    const likeButton = within(post).getByRole('button', { name: 'いいね' })
    expect(likeButton).toHaveTextContent('0')
    await user.click(likeButton)
    expect(likeButton).toHaveClass('liked')
    expect(likeButton).toHaveTextContent('1')
  })
})
