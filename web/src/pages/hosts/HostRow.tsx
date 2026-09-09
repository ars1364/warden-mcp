import type { HostMeta } from '../../types/api'
import { TagChips } from '../../components/common/TagChips'
import { CopyButton } from '../../components/common/CopyButton'
import { Button } from '../../components/common/Button'
import { HostTypeBadge } from './HostTypeBadge'
import { StatusBadge } from './StatusBadge'

interface HostRowProps {
  host: HostMeta
  onEdit: (host: HostMeta) => void
  onDelete: (host: HostMeta) => void
}

function locationLabel(host: HostMeta): string {
  if (host.parent_host_name) return `on ${host.parent_host_name}`
  if (host.cloud_provider) return [host.cloud_provider, host.cloud_account].filter(Boolean).join(' · ')
  if (host.physical_location) return host.physical_location
  return host.location_kind || '—'
}

export function HostRow({ host, onEdit, onDelete }: HostRowProps) {
  const addresses = host.addresses ?? []
  const primaryAddress = addresses[0]?.address ?? host.name
  const sshCommand = host.ssh_username
    ? `ssh -p ${host.ssh_port} ${host.ssh_username}@${primaryAddress}`
    : ''

  return (
    <tr className="border-b border-border last:border-0">
      <td className="px-4 py-3 align-top">
        <div className="flex items-center gap-2">
          <span className="font-medium text-text-bright">{host.name}</span>
          <HostTypeBadge type={host.host_type} />
          <StatusBadge status={host.status} />
        </div>
        <div className="text-xs text-text-muted">{host.description}</div>
      </td>
      <td className="px-4 py-3 align-top text-xs text-text-muted">{locationLabel(host)}</td>
      <td className="px-4 py-3 align-top">
        {addresses.length === 0 ? (
          <span className="text-xs text-text-muted">—</span>
        ) : (
          <div className="space-y-0.5">
            {addresses.map((a) => (
              <div key={a.label + a.address} className="flex items-center gap-1.5 font-mono text-xs">
                <span className="text-text-muted">{a.label}</span>
                <span className="text-text">{a.address}</span>
              </div>
            ))}
          </div>
        )}
      </td>
      <td className="px-4 py-3 align-top">
        {sshCommand ? (
          <div className="flex items-center gap-1.5">
            <code className="text-xs text-text-muted">{sshCommand}</code>
            <CopyButton value={sshCommand} />
          </div>
        ) : (
          <span className="text-xs text-text-muted">—</span>
        )}
        {host.ssh_secret_name && (
          <div className="mt-0.5 text-xs text-text-muted">
            key: <span className="font-mono text-accent">{host.ssh_secret_name}</span>
          </div>
        )}
      </td>
      <td className="px-4 py-3 align-top">
        <TagChips tags={host.tags} />
      </td>
      <td className="px-4 py-3 align-top">
        <div className="flex gap-2">
          <Button variant="secondary" onClick={() => onEdit(host)}>
            Edit
          </Button>
          <Button variant="danger" onClick={() => onDelete(host)}>
            Delete
          </Button>
        </div>
      </td>
    </tr>
  )
}
