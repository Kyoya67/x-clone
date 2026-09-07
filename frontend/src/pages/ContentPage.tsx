import { ReactNode } from 'react'
import { PageLayout } from '../components/PageLayout'

type ContentPageProps = { title: string; description: string; children?: ReactNode }

export function ContentPage({ title, description, children }: ContentPageProps) {
  return (
    <PageLayout>
      <header className="content-header">
        <h1>{title}</h1>
      </header>
      <section className="content-body">
        <p>{description}</p>
        {children}
      </section>
    </PageLayout>
  )
}
