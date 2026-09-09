import type { HostStatus } from '../../types/api'

const CLASSES: Record<HostStatus, string> = {
  active: 'bg-success-bg text-success',
  maintenance: 'bg-warn-bg text-warn',
  decommissioned: 'bg-surface-hover text-text-muted',
}

export function StatusBadge({ status }: { status: HostStatus }) {
  return (
    <span className={`rounded-full px-1.5 py-0.5 text-[10px] font-medium uppercase tracking-wide ${CLASSES[status]}`}>
      {status}
    </span>
  )
}
