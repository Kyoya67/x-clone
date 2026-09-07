import { FormEvent } from 'react'

export function Composer({
  draft,
  onDraftChange,
  onPublish,
}: {
  draft: string
  onDraftChange: (value: string) => void
  onPublish: (event: FormEvent) => void
}) {
  return (
    <form className="composer" onSubmit={onPublish}>
      <span className="avatar avatar-blue">太</span>
      <div className="composer-main">
        <textarea
          id="composer"
          value={draft}
          onChange={(event) => onDraftChange(event.target.value)}
          placeholder="いまどうしてる？"
          maxLength={140}
          rows={3}
        />
        <div className="composer-footer">
          <span className="char-count">{draft.length}/140</span>
          <button className="publish-button" disabled={!draft.trim()} type="submit">
            ポストする
          </button>
        </div>
      </div>
    </form>
  )
}
