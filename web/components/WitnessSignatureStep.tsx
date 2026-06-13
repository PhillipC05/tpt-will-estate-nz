'use client'

interface WitnessSignatureStepProps {
  willId: string
}

export function WitnessSignatureStep({ willId }: WitnessSignatureStepProps) {
  return (
    <div style={{ marginTop: '1rem' }}>
      <p style={{ color: '#565656' }}>
        Share the following link with each witness. They must sign in with their
        own RealMe verified identity to countersign this will.
      </p>

      <div style={{
        background: '#F8F8F8',
        border: '1px solid #DEDEDE',
        borderRadius: 4,
        padding: '0.75rem 1rem',
        marginTop: '0.75rem',
        fontFamily: 'monospace',
        fontSize: '0.9rem',
        wordBreak: 'break-all',
      }}>
        {typeof window !== 'undefined'
          ? `${window.location.origin}/wills/${willId}/witness`
          : `/wills/${willId}/witness`}
      </div>

      <div style={{ marginTop: '1.5rem' }}>
        <h3 style={{ fontSize: '1rem', fontWeight: 600 }}>Witnesses</h3>
        <p style={{ color: '#787878', fontSize: '0.875rem' }}>
          No witnesses have signed yet. Both witnesses must sign before this will
          can be finalised.
        </p>
        <ul style={{ listStyle: 'none', padding: 0 }}>
          {[1, 2].map(n => (
            <li key={n} style={{
              display: 'flex',
              alignItems: 'center',
              gap: 10,
              padding: '10px 0',
              borderBottom: '1px solid #EEEEEE',
            }}>
              <span style={{
                width: 24, height: 24,
                borderRadius: '50%',
                background: '#EEEEEE',
                display: 'flex', alignItems: 'center', justifyContent: 'center',
                fontSize: '0.75rem',
                color: '#787878',
              }}>
                {n}
              </span>
              <span style={{ color: '#9A9A9A' }}>Waiting for witness {n}…</span>
            </li>
          ))}
        </ul>
      </div>
    </div>
  )
}
