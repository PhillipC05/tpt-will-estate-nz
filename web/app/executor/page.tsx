'use client'

export default function ExecutorPage() {
  return (
    <main style={{ maxWidth: 760, margin: '0 auto', padding: '2rem 1rem' }}>
      <h1 style={{ fontSize: '1.75rem', fontWeight: 700, marginBottom: '0.5rem' }}>
        Executor Access
      </h1>
      <p style={{ color: '#565656', marginBottom: '1.5rem' }}>
        This page is accessible only after a registered death notification has
        been received from Births, Deaths and Marriages (BDM). You must be
        logged in with the RealMe identity that was nominated as an executor.
      </p>

      <section style={{
        background: '#FDF4E3',
        border: '1px solid #E07500',
        borderRadius: 4,
        padding: '1rem 1.25rem',
        marginBottom: '2rem',
      }}>
        <strong>Awaiting death notification</strong>
        <p style={{ margin: '0.25rem 0 0' }}>
          The will vault cannot be unlocked until BDM confirms the testator&apos;s
          death. If you believe this is an error, contact support.
        </p>
      </section>

      <p style={{ fontSize: '0.875rem', color: '#787878' }}>
        This service does not constitute legal advice. Estate administration
        must comply with the Administration Act 1969 and the Wills Act 2007.
        Consult a solicitor or Public Trust for complex estates.
      </p>
    </main>
  )
}
