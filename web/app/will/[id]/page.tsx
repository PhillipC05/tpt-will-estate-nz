'use client'

import { useParams } from 'next/navigation'
import { WitnessSignatureStep } from '../../../components/WitnessSignatureStep'
import { EncryptionKeyWarning } from '../../../components/EncryptionKeyWarning'

export default function WillDetailPage() {
  const { id } = useParams<{ id: string }>()

  return (
    <main style={{ maxWidth: 760, margin: '0 auto', padding: '2rem 1rem' }}>
      <h1 style={{ fontSize: '1.75rem', fontWeight: 700, marginBottom: '0.5rem' }}>
        Will #{id}
      </h1>

      <EncryptionKeyWarning />

      <section style={{ marginTop: '2rem' }}>
        <h2 style={{ fontSize: '1.25rem', fontWeight: 600 }}>Witness Signatures</h2>
        <p style={{ color: '#565656' }}>
          Under the Wills Act 2007, your will requires two independent witnesses.
          Each witness must sign using their own RealMe verified identity.
        </p>
        <WitnessSignatureStep willId={id} />
      </section>
    </main>
  )
}
