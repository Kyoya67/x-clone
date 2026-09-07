import { render, screen, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { MemoryRouter } from 'react-router-dom'
import { describe, expect, it } from 'vitest'
import { TimelinePage } from './TimelinePage'
import { FollowingProvider } from '../state/FollowingContext'

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
  it('投稿フォームから投稿を追加できる', async () => {
    const user = userEvent.setup()
    renderTimeline()
    const composer = screen.getByPlaceholderText('いまどうしてる？')
    await user.type(composer, 'テスト投稿です')
    await user.click(
      within(composer.closest('form') as HTMLFormElement).getByRole('button', {
        name: 'ポストする',
      }),
    )
    expect(screen.getByText('テスト投稿です')).toBeInTheDocument()
  })

  it('フォロー中タブではフォロー中ユーザーの投稿を表示する', async () => {
    const user = userEvent.setup()
    renderTimeline()
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

  it('いいねを押すと件数と状態が変わる', async () => {
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
