import { render, screen } from '@testing-library/react'
import { MemoryRouter } from 'react-router-dom'
import { describe, expect, it } from 'vitest'
import { TimelinePage } from './TimelinePage'
import { FollowingProvider } from '../state/FollowingContext'

function renderAtWidth(width: number) {
  window.innerWidth = width
  window.dispatchEvent(new Event('resize'))
  return render(
    <MemoryRouter>
      <FollowingProvider>
        <TimelinePage />
      </FollowingProvider>
    </MemoryRouter>,
  )
}

describe('レスポンシブレイアウト', () => {
  it('デスクトップ幅で3つのレイアウト領域を表示する', () => {
    const { container } = renderAtWidth(1280)
    expect(container.querySelector('.sidebar')).toBeInTheDocument()
    expect(container.querySelector('.timeline')).toBeInTheDocument()
    expect(container.querySelector('.right-rail')).toBeInTheDocument()
    expect(screen.getByPlaceholderText('いまどうしてる？')).toBeInTheDocument()
  })

  it('モバイル幅でもタイムラインと操作領域を表示する', () => {
    const { container } = renderAtWidth(375)
    expect(container.querySelector('.sidebar')).toBeInTheDocument()
    expect(container.querySelector('.timeline')).toBeInTheDocument()
    expect(container.querySelector('.right-rail')).toBeInTheDocument()
    expect(screen.getByRole('tab', { name: 'フォロー中' })).toBeInTheDocument()
  })
})
