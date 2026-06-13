'use client'

import { useState } from 'react'

interface Clause {
  id: string
  text: string
}

interface Beneficiary {
  name: string
  email: string
  relation: string
}

interface Executor {
  name: string
  email: string
}

export function WillClauseEditor() {
  const [clauses, setClauses] = useState<Clause[]>([])
  const [beneficiaries, setBeneficiaries] = useState<Beneficiary[]>([])
  const [executors, setExecutors] = useState<Executor[]>([])
  const [step, setStep] = useState<'clauses' | 'beneficiaries' | 'executors' | 'review'>('clauses')
  const [saving, setSaving] = useState(false)

  function addClause() {
    setClauses(prev => [...prev, { id: crypto.randomUUID(), text: '' }])
  }

  function updateClause(id: string, text: string) {
    setClauses(prev => prev.map(c => c.id === id ? { ...c, text } : c))
  }

  function removeClause(id: string) {
    setClauses(prev => prev.filter(c => c.id !== id))
  }

  async function handleSubmit() {
    setSaving(true)
    try {
      await fetch('/api/wills', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ clauses, beneficiaries, executors }),
      })
    } finally {
      setSaving(false)
    }
  }

  return (
    <div>
      {/* Step indicator */}
      <nav style={{ display: 'flex', gap: 8, marginBottom: '1.5rem' }}>
        {(['clauses', 'beneficiaries', 'executors', 'review'] as const).map(s => (
          <button
            key={s}
            onClick={() => setStep(s)}
            style={{
              padding: '6px 14px',
              borderRadius: 4,
              border: '1px solid #DEDEDE',
              background: step === s ? '#00538C' : '#F8F8F8',
              color: step === s ? '#fff' : '#333',
              cursor: 'pointer',
              textTransform: 'capitalize',
            }}
          >
            {s}
          </button>
        ))}
      </nav>

      {step === 'clauses' && (
        <section>
          <h2 style={{ fontSize: '1.1rem', fontWeight: 600 }}>Will Clauses</h2>
          {clauses.map((c, i) => (
            <div key={c.id} style={{ marginBottom: '0.75rem' }}>
              <label style={{ display: 'block', fontWeight: 500, marginBottom: 4 }}>
                Clause {i + 1}
              </label>
              <textarea
                value={c.text}
                onChange={e => updateClause(c.id, e.target.value)}
                rows={3}
                style={{ width: '100%', padding: 8, border: '1px solid #DEDEDE', borderRadius: 4 }}
                placeholder="Describe this clause..."
              />
              <button
                onClick={() => removeClause(c.id)}
                style={{ color: '#C44000', background: 'none', border: 'none', cursor: 'pointer', padding: 0 }}
              >
                Remove
              </button>
            </div>
          ))}
          <button
            onClick={addClause}
            style={{ padding: '8px 16px', background: '#E8F0F7', border: '1px solid #00538C', borderRadius: 4, cursor: 'pointer' }}
          >
            + Add Clause
          </button>
          <div style={{ marginTop: '1rem' }}>
            <button
              onClick={() => setStep('beneficiaries')}
              disabled={clauses.length === 0}
              style={{ padding: '10px 20px', background: '#00538C', color: '#fff', border: 'none', borderRadius: 4, cursor: 'pointer' }}
            >
              Next: Beneficiaries
            </button>
          </div>
        </section>
      )}

      {step === 'beneficiaries' && (
        <section>
          <h2 style={{ fontSize: '1.1rem', fontWeight: 600 }}>Beneficiaries</h2>
          {beneficiaries.map((b, i) => (
            <div key={i} style={{ display: 'grid', gridTemplateColumns: '1fr 1fr 1fr auto', gap: 8, marginBottom: 8 }}>
              <input placeholder="Full name" value={b.name}
                onChange={e => setBeneficiaries(prev => prev.map((x, j) => j === i ? { ...x, name: e.target.value } : x))}
                style={{ padding: 8, border: '1px solid #DEDEDE', borderRadius: 4 }} />
              <input placeholder="Email" value={b.email}
                onChange={e => setBeneficiaries(prev => prev.map((x, j) => j === i ? { ...x, email: e.target.value } : x))}
                style={{ padding: 8, border: '1px solid #DEDEDE', borderRadius: 4 }} />
              <input placeholder="Relation" value={b.relation}
                onChange={e => setBeneficiaries(prev => prev.map((x, j) => j === i ? { ...x, relation: e.target.value } : x))}
                style={{ padding: 8, border: '1px solid #DEDEDE', borderRadius: 4 }} />
              <button onClick={() => setBeneficiaries(prev => prev.filter((_, j) => j !== i))}
                style={{ color: '#C44000', background: 'none', border: 'none', cursor: 'pointer' }}>✕</button>
            </div>
          ))}
          <button onClick={() => setBeneficiaries(prev => [...prev, { name: '', email: '', relation: '' }])}
            style={{ padding: '8px 16px', background: '#E8F0F7', border: '1px solid #00538C', borderRadius: 4, cursor: 'pointer' }}>
            + Add Beneficiary
          </button>
          <div style={{ marginTop: '1rem' }}>
            <button onClick={() => setStep('executors')}
              style={{ padding: '10px 20px', background: '#00538C', color: '#fff', border: 'none', borderRadius: 4, cursor: 'pointer' }}>
              Next: Executors
            </button>
          </div>
        </section>
      )}

      {step === 'executors' && (
        <section>
          <h2 style={{ fontSize: '1.1rem', fontWeight: 600 }}>Executors</h2>
          <p style={{ color: '#565656' }}>Executors must have a RealMe verified identity to access the will after your death.</p>
          {executors.map((e, i) => (
            <div key={i} style={{ display: 'grid', gridTemplateColumns: '1fr 1fr auto', gap: 8, marginBottom: 8 }}>
              <input placeholder="Full name" value={e.name}
                onChange={ev => setExecutors(prev => prev.map((x, j) => j === i ? { ...x, name: ev.target.value } : x))}
                style={{ padding: 8, border: '1px solid #DEDEDE', borderRadius: 4 }} />
              <input placeholder="Email" value={e.email}
                onChange={ev => setExecutors(prev => prev.map((x, j) => j === i ? { ...x, email: ev.target.value } : x))}
                style={{ padding: 8, border: '1px solid #DEDEDE', borderRadius: 4 }} />
              <button onClick={() => setExecutors(prev => prev.filter((_, j) => j !== i))}
                style={{ color: '#C44000', background: 'none', border: 'none', cursor: 'pointer' }}>✕</button>
            </div>
          ))}
          <button onClick={() => setExecutors(prev => [...prev, { name: '', email: '' }])}
            style={{ padding: '8px 16px', background: '#E8F0F7', border: '1px solid #00538C', borderRadius: 4, cursor: 'pointer' }}>
            + Add Executor
          </button>
          <div style={{ marginTop: '1rem' }}>
            <button onClick={() => setStep('review')}
              disabled={executors.length === 0}
              style={{ padding: '10px 20px', background: '#00538C', color: '#fff', border: 'none', borderRadius: 4, cursor: 'pointer' }}>
              Review & Submit
            </button>
          </div>
        </section>
      )}

      {step === 'review' && (
        <section>
          <h2 style={{ fontSize: '1.1rem', fontWeight: 600 }}>Review Your Will</h2>
          <p><strong>Clauses:</strong> {clauses.length}</p>
          <p><strong>Beneficiaries:</strong> {beneficiaries.length}</p>
          <p><strong>Executors:</strong> {executors.length}</p>
          <p style={{ color: '#565656', fontSize: '0.875rem' }}>
            By submitting, you confirm this is your current testamentary intention.
            Your will is encrypted before storage and cannot be read by this service.
          </p>
          <button
            onClick={handleSubmit}
            disabled={saving}
            style={{ padding: '10px 24px', background: '#007B40', color: '#fff', border: 'none', borderRadius: 4, cursor: 'pointer' }}
          >
            {saving ? 'Saving...' : 'Submit Will'}
          </button>
        </section>
      )}
    </div>
  )
}
