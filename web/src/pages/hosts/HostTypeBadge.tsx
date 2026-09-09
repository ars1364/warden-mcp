import type { HostType } from '../../types/api'

export function HostTypeBadge({ type }: { type: HostType }) {
  return (
    <span className="rounded-full bg-surface-hover px-1.5 py-0.5 text-[10px] font-medium uppercase tracking-wide text-text-muted">
      {type}
    </span>
  )
}
