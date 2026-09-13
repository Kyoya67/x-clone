import { FormEvent } from 'react'

export function Composer({
  draft,
  avatar,
  avatarClass = 'avatar-blue',
  onDraftChange,
  onPublish,
}: {
  draft: string
  avatar: string
  avatarClass?: string
  onDraftChange: (value: string) => void
  onPublish: (event: FormEvent) => void | Promise<void>
}) {
  return (
    <form className="composer" onSubmit={onPublish}>
      <span className={`avatar ${avatarClass}`}>{avatar}</span>
      <div className="composer-main">
        <textarea
          id="composer"
          value={draft}
          onChange={(event) => onDraftChange(event.target.value)}
          placeholder="いまどうしてる？"
          maxLength={280}
          rows={3}
        />
        <div className="composer-footer">
          <span className="char-count">{draft.length}/280</span>
          <button className="publish-button" disabled={!draft.trim()} type="submit">
            ポストする
          </button>
        </div>
      </div>
    </form>
  )
}
