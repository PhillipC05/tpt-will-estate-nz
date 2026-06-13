'use client'

import { WillClauseEditor } from '../../../components/WillClauseEditor'

export default function CreateWillPage() {
  return (
    <main style={{ maxWidth: 760, margin: '0 auto', padding: '2rem 1rem' }}>
      <header style={{ marginBottom: '2rem' }}>
        <h1 style={{ fontSize: '1.75rem', fontWeight: 700, margin: 0 }}>
          Create Your Digital Will
        </h1>
        <p style={{ color: '#565656', marginTop: '0.5rem' }}>
          Your will is encrypted client-side before being stored. Only you and
          your nominated executors can access it. This service does not provide
          legal advice — consult a solicitor for complex estates.
        </p>
      </header>
      <WillClauseEditor />
    </main>
  )
}
