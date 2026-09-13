import { ReactNode } from 'react'
import { useAuth } from '../state/AuthContext'

export function AuthGate({ children }: { children: ReactNode }) {
  const { user, isLoading, error, login } = useAuth()

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

  return children
}
