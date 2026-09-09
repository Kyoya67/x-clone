import { render, screen, waitFor } from '@testing-library/react'
import { within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { MemoryRouter } from 'react-router-dom'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { RightRail } from './RightRail'
import { FollowingProvider } from '../state/FollowingContext'

describe('RightRail', () => {
  beforeEach(() => {
    vi.stubGlobal(
      'fetch',
      vi.fn((input: RequestInfo | URL) => {
        const url = input instanceof Request ? input.url : input.toString()
        if (url === '/api/me/following') {
          return Promise.resolve(
            new Response(JSON.stringify({ userIds: ['00000000-0000-0000-0000-000000000004'] }), {
              status: 200,
              headers: { 'Content-Type': 'application/json' },
            }),
          )
        }
        return Promise.resolve(new Response(null, { status: 204 }))
      }),
    )
  })

  function productTeamFollowButton() {
    const followCard = screen.getByRole('heading', { name: 'おすすめユーザー' }).closest('section')
    if (!followCard) throw new Error('follow card was not found')
    const profileLink = within(followCard).getByRole('link', { name: /プロダクト開発部/ })
    return within(profileLink.parentElement as HTMLElement).getByRole('button')
  }

  afterEach(() => {
    vi.unstubAllGlobals()
  })

  it('toggles follow and unfollow', async () => {
    const user = userEvent.setup()
    render(
      <MemoryRouter>
        <FollowingProvider>
          <RightRail />
        </FollowingProvider>
      </MemoryRouter>,
    )
    const followButton = productTeamFollowButton()
    await user.click(followButton)
    await waitFor(() => expect(followButton).toHaveTextContent('フォロー中'))
    expect(fetch).toHaveBeenCalledWith('/api/users/00000000-0000-0000-0000-000000000005/follow', {
      method: 'PUT',
    })
    await user.click(followButton)
    await waitFor(() => expect(followButton).toHaveTextContent('フォロー'))
    expect(fetch).toHaveBeenLastCalledWith(
      '/api/users/00000000-0000-0000-0000-000000000005/follow',
      {
        method: 'DELETE',
      },
    )
  })

  it('keeps the follow state when the API request fails', async () => {
    const user = userEvent.setup()
    vi.mocked(fetch).mockResolvedValue(new Response(null, { status: 500 }))
    render(
      <MemoryRouter>
        <FollowingProvider>
          <RightRail />
        </FollowingProvider>
      </MemoryRouter>,
    )

    const followButton = productTeamFollowButton()
    await user.click(followButton)

    expect(await screen.findByRole('alert')).toHaveTextContent(
      'フォロー操作に失敗しました。時間をおいて再度お試しください。',
    )
    expect(followButton).toHaveTextContent('フォロー')
  })
})
