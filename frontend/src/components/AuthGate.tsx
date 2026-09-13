import { FormEvent, ReactNode, useEffect, useState } from 'react'
import { useAuth } from '../state/AuthContext'

export function AuthGate({ children }: { children: ReactNode }) {
  const { user, isLoading, error, login, updateProfile } = useAuth()
  const [form, setForm] = useState({ handle: '', displayName: '', bio: '' })
  const [setupError, setSetupError] = useState('')
  const [isSaving, setIsSaving] = useState(false)

  useEffect(() => {
    if (user?.needsProfileSetup) {
      setForm({
        handle: '',
        displayName: user.displayName,
        bio: user.bio,
      })
    }
  }, [user])

  const saveProfile = async (event: FormEvent) => {
    event.preventDefault()
    setSetupError('')
    if (!/^[A-Za-z0-9_]{1,13}$/.test(form.handle)) {
      setSetupError('ハンドルは半角英数字と_のみ、13文字以下で入力してください。')
      return
    }
    if (!form.displayName.trim()) {
      setSetupError('表示名を入力してください。')
      return
    }
    setIsSaving(true)
    try {
      await updateProfile({
        handle: form.handle.trim(),
        displayName: form.displayName.trim(),
        bio: form.bio.trim(),
      })
    } catch {
      setSetupError('プロフィール更新に失敗しました。時間をおいて再度お試しください。')
    } finally {
      setIsSaving(false)
    }
  }

  if (isLoading) return <main className="login-page">ログイン状態を確認しています。</main>

  if (!user) {
    return (
      <main className="login-page">
        <section className="login-card">
          <h1>X Clone</h1>
          <p>Googleアカウントでログインしてください。</p>
          {error && <p role="alert">{error}</p>}
          <button type="button" onClick={login}>
            Googleでログイン
          </button>
        </section>
      </main>
    )
  }

  if (user.needsProfileSetup) {
    return (
      <main className="login-page">
        <form className="login-card profile-form" onSubmit={saveProfile}>
          <h1>プロフィール設定</h1>
          <p>ハンドルは半角英数字と_のみ、13文字以下で設定してください。</p>
          <label>
            ハンドル
            <input
              value={form.handle}
              onChange={(event) =>
                setForm((current) => ({ ...current, handle: event.target.value }))
              }
              maxLength={13}
              pattern="[A-Za-z0-9_]{1,13}"
              required
            />
          </label>
          <label>
            表示名
            <input
              value={form.displayName}
              onChange={(event) =>
                setForm((current) => ({ ...current, displayName: event.target.value }))
              }
              maxLength={100}
              required
            />
          </label>
          <label>
            自己紹介
            <textarea
              value={form.bio}
              onChange={(event) => setForm((current) => ({ ...current, bio: event.target.value }))}
              maxLength={160}
              rows={3}
            />
          </label>
          {setupError && <p role="alert">{setupError}</p>}
          <button disabled={isSaving} type="submit">
            保存して始める
          </button>
        </form>
      </main>
    )
  }

  return children
}
