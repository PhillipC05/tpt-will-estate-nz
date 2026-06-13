'use client'

export default function BeneficiaryPage() {
  return (
    <main style={{ maxWidth: 760, margin: '0 auto', padding: '2rem 1rem' }}>
      <h1 style={{ fontSize: '1.75rem', fontWeight: 700, marginBottom: '0.5rem' }}>
        Beneficiary Notification
      </h1>
      <p style={{ color: '#565656', marginBottom: '1.5rem' }}>
        You have been notified as a beneficiary of a digital will. To view the
        specific bequest made to you, sign in with your RealMe identity.
      </p>

      <section style={{
        background: '#E5F2EB',
        border: '1px solid #007B40',
        borderRadius: 4,
        padding: '1rem 1.25rem',
        marginBottom: '2rem',
      }}>
        <strong>Will status: Unlocked</strong>
        <p style={{ margin: '0.25rem 0 0' }}>
          The executor has unlocked this will following confirmation of the
          testator&apos;s passing. Your nominated executor will be in contact to
          discuss the next steps.
        </p>
      </section>

      <p style={{ fontSize: '0.875rem', color: '#787878' }}>
        This service does not constitute legal advice. If you have questions
        about the administration of this estate, consult a solicitor or
        contact the Public Trust.
      </p>
    </main>
  )
}
