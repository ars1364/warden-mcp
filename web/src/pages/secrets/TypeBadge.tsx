import type { SecretType } from '../../types/api'

const LABELS: Record<SecretType, string> = {
  opaque: 'opaque',
  structured: 'structured',
  totp: 'totp',
  reference: 'reference',
}

const CLASSES: Record<SecretType, string> = {
  opaque: 'bg-surface-hover text-text-muted',
  structured: 'bg-accent/15 text-accent',
  totp: 'bg-warn-bg text-warn',
  reference: 'bg-surface-hover text-text-muted',
}

export function TypeBadge({ type }: { type: SecretType }) {
  return (
    <span className={`rounded-full px-1.5 py-0.5 text-[10px] font-medium uppercase tracking-wide ${CLASSES[type]}`}>
      {LABELS[type]}
    </span>
  )
}
