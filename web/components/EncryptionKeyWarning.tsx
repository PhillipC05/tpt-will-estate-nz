'use client'

export function EncryptionKeyWarning() {
  return (
    <div
      role="alert"
      style={{
        background: '#FAEAE3',
        border: '2px solid #C44000',
        borderRadius: 4,
        padding: '1rem 1.25rem',
        marginBottom: '1.5rem',
      }}
    >
      <strong style={{ color: '#C44000', display: 'block', marginBottom: '0.25rem' }}>
        Keep your encryption key safe
      </strong>
      <p style={{ margin: 0, fontSize: '0.9rem', color: '#444' }}>
        Your will is encrypted with a key that only you hold. If you lose this
        key, your will <strong>cannot be recovered</strong> — not by this service,
        not by your executor, and not by a court order. Store it securely (e.g.
        with a trusted solicitor or in a physical safe).
      </p>
    </div>
  )
}
