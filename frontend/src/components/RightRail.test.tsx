import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { MemoryRouter } from 'react-router-dom'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { RightRail } from './RightRail'
import { FollowingProvider } from '../state/FollowingContext'

describe('RightRail', () => {
  beforeEach(() => {
    vi.stubGlobal('fetch', vi.fn().mockResolvedValue(new Response(null, { status: 204 })))
  })

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
    const followButton = screen.getByRole('button', { name: 'フォロー' })
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

    const followButton = screen.getByRole('button', { name: 'フォロー' })
    await user.click(followButton)

    expect(await screen.findByRole('alert')).toHaveTextContent(
      'フォロー操作に失敗しました。時間をおいて再度お試しください。',
    )
    expect(followButton).toHaveTextContent('フォロー')
  })
})
