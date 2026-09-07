import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { MemoryRouter } from 'react-router-dom'
import { describe, expect, it } from 'vitest'
import { RightRail } from './RightRail'
import { FollowingProvider } from '../state/FollowingContext'

describe('RightRail', () => {
  it('フォローとフォロー解除を切り替えられる', async () => {
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
    expect(followButton).toHaveTextContent('フォロー中')
    await user.click(followButton)
    expect(followButton).toHaveTextContent('フォロー')
  })
})
