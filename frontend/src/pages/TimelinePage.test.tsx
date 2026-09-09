import { render, screen, waitFor, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { MemoryRouter } from 'react-router-dom'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { TimelinePage } from './TimelinePage'
import { FollowingProvider } from '../state/FollowingContext'

const postResponse = {
  id: 'post-from-api',
  authorId: '00000000-0000-0000-0000-000000000001',
  content: 'テスト投稿です',
  createdAt: '2026-09-08T00:00:00Z',
}

beforeEach(() => {
  vi.stubGlobal(
    'fetch',
    vi.fn((input: RequestInfo | URL) => {
      const url = input instanceof Request ? input.url : input.toString()
      if (url === '/api/me/following') {
        return Promise.resolve(
          new Response(
            JSON.stringify({
              userIds: [
                '00000000-0000-0000-0000-000000000002',
                '00000000-0000-0000-0000-000000000004',
              ],
            }),
            { status: 200, headers: { 'Content-Type': 'application/json' } },
          ),
        )
      }
      return Promise.resolve(
        new Response(JSON.stringify(postResponse), {
          status: 201,
          headers: { 'Content-Type': 'application/json' },
        }),
      )
    }),
  )
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

function productTeamFollowButton() {
  const followCard = screen.getByRole('heading', { name: 'おすすめユーザー' }).closest('section')
  if (!followCard) throw new Error('follow card was not found')
  const profileLink = within(followCard).getByRole('link', { name: /プロダクト開発部/ })
  return within(profileLink.parentElement as HTMLElement).getByRole('button')
}

describe('TimelinePage', () => {
  it('adds a post through the composer', async () => {
    const user = userEvent.setup()
    renderTimeline()
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

  it('displays an error when the post API fails', async () => {
    const user = userEvent.setup()
    vi.mocked(fetch).mockImplementation((input: RequestInfo | URL) => {
      const url = input instanceof Request ? input.url : input.toString()
      if (url === '/api/me/following') {
        return Promise.resolve(
          new Response(JSON.stringify({ userIds: [] }), {
            status: 200,
            headers: { 'Content-Type': 'application/json' },
          }),
        )
      }
      return Promise.resolve(new Response(null, { status: 500 }))
    })
    renderTimeline()
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
    expect(document.querySelector('.feed')).not.toHaveTextContent('失敗する投稿')
  })

  it('shows followed users posts in the following tab', async () => {
    const user = userEvent.setup()
    renderTimeline()
    await waitFor(() => expect(fetch).toHaveBeenCalledWith('/api/me/following'))
    await user.click(screen.getByRole('tab', { name: 'フォロー中' }))
    expect(
      screen.getByText('新しいサービスの最初の一歩。ユーザーが迷わず使える体験を大切にしたい。'),
    ).toBeInTheDocument()
    expect(
      screen.queryByText(
        '今日はチームでタイムライン機能の設計をしました。小さく作って検証するのが気持ちいい。',
      ),
    ).not.toBeInTheDocument()
  })

  it('shows a newly followed users posts in the following tab', async () => {
    const user = userEvent.setup()
    renderTimeline()

    await waitFor(() => expect(fetch).toHaveBeenCalledWith('/api/me/following'))
    await user.click(productTeamFollowButton())
    await user.click(screen.getByRole('tab', { name: 'フォロー中' }))

    expect(
      await screen.findByText(
        '今週は開発環境の改善に取り組んでいます。小さな改善を積み重ねていきます。',
      ),
    ).toBeInTheDocument()
  })

  it('updates the like count and state when liked', async () => {
    const user = userEvent.setup()
    renderTimeline()
    const post = screen
      .getByText(
        '今日はチームでタイムライン機能の設計をしました。小さく作って検証するのが気持ちいい。',
      )
      .closest('.post') as HTMLElement
    const likeButton = within(post).getByRole('button', { name: 'いいね' })
    expect(likeButton).toHaveTextContent('18')
    await user.click(likeButton)
    expect(likeButton).toHaveClass('liked')
    expect(likeButton).toHaveTextContent('19')
  })
})
