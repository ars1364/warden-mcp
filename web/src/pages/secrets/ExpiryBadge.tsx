const SOON_DAYS = 30

export function ExpiryBadge({ expiresAt }: { expiresAt?: string }) {
  if (!expiresAt) {
    return <span className="text-xs text-text-muted">—</span>
  }

  const daysLeft = Math.ceil((new Date(expiresAt + 'T00:00:00').getTime() - Date.now()) / 86_400_000)
  const className =
    daysLeft < 0
      ? 'text-danger font-medium'
      : daysLeft <= SOON_DAYS
        ? 'text-warn font-medium'
        : 'text-text-muted'

  return (
    <span className={`text-xs ${className}`} title={expiresAt}>
      {expiresAt}
      {daysLeft < 0 ? ' (expired)' : daysLeft <= SOON_DAYS ? ` (${daysLeft}d)` : ''}
    </span>
  )
}
