# Future work — track as GitHub issues before v1.0

## High priority

- **Executor identity verification** — store nominated executor RealMe FLTs at
  will-creation time and cross-check at access time (currently any verified
  user can call the executor endpoint if they know the will ID).

- **Email delivery** — wire an email provider (SES, Postmark, etc.) into
  `BeneficiaryHandler.NotifyAll`. The endpoint currently logs intent and
  returns 202 without sending email.

- **Testator liveness check** — annual prompt for the testator to re-confirm
  the will is current; escalate to email reminders if no login is detected.

## Medium priority

- **Will supersession** — when a testator creates a new will, automatically
  mark the previous one as superseded to prevent competing instruments.

- **Codicil support** — allow amendments to a locked will without full
  revocation (append-only codicil model with re-witness flow).

- **Time-lock vault** — configurable cooling-off period after BDM death
  notification before the vault transitions to `unlocked_at_death`
  (e.g., 30 days for estate administration).

- **Audit trail** — hash-chained audit log for each state transition, making
  the provenance of the will tamper-evident.

## Low priority / ideas

- **Digital asset clauses** — structured clause type for crypto wallets,
  social media accounts, and email with access instructions.

- **LINZ property cross-reference** — verify assets named in the will against
  LINZ title data at will-creation time.

- **Probate submission API** — integrate with the NZ High Court probate filing
  endpoint once a public API is available.

- **Multi-instance session store** — replace the default JWT cookie session
  with the Redis-backed store in `packages/nz-common/auth` for horizontal
  scaling.
