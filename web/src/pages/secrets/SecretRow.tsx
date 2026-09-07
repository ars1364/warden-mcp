import { useState } from 'react'
import { api } from '../../lib/api'
import { errorMessage } from '../../lib/errors'
import type { SecretMeta, SecretType, SecretWithValue } from '../../types/api'
import { MaskedValue } from '../../components/common/MaskedValue'
import { CopyButton } from '../../components/common/CopyButton'
import { TagChips } from '../../components/common/TagChips'
import { Button } from '../../components/common/Button'
import { TypeBadge } from './TypeBadge'
import { TotpCodeButton } from './TotpCodeButton'

interface SecretRowProps {
  secret: SecretMeta
  onEdit: (secret: SecretMeta) => void
  onDelete: (secret: SecretMeta) => void
}

export function SecretRow({ secret, onEdit, onDelete }: SecretRowProps) {
  const [revealed, setRevealed] = useState(false)
  const [detail, setDetail] = useState<SecretWithValue | null>(null)
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)

  async function handleToggleReveal() {
    if (revealed) {
      setRevealed(false)
      return
    }
    if (detail !== null) {
      setRevealed(true)
      return
    }
    setLoading(true)
    setError(null)
    try {
      const full = await api.get<SecretWithValue>(`/api/secrets/${secret.id}`)
      setDetail(full)
      setRevealed(true)
    } catch (err) {
      setError(errorMessage(err))
    } finally {
      setLoading(false)
    }
  }

  return (
    <tr className="border-b border-border last:border-0">
      <td className="px-4 py-3 align-top">
        <div className="flex items-center gap-2">
          <span className="font-medium text-text-bright">{secret.name}</span>
          <TypeBadge type={secret.type} />
        </div>
        <div className="text-xs text-text-muted">{secret.description}</div>
      </td>
      <td className="px-4 py-3 align-top">
        <TagChips tags={secret.tags} />
      </td>
      <td className="px-4 py-3 align-top">
        <div className="w-64">
          {secret.type === 'totp' && <TotpCodeButton secretId={secret.id} />}
          {loading ? (
            <span className="text-xs text-text-muted">Loading…</span>
          ) : error ? (
            <span className="text-xs text-danger">{error}</span>
          ) : (
            <SecretValueCell type={secret.type} detail={detail} revealed={revealed} onToggle={handleToggleReveal} />
          )}
        </div>
      </td>
      <td className="px-4 py-3 align-top text-xs text-text-muted">
        {new Date(secret.updated_at).toLocaleString()}
      </td>
      <td className="px-4 py-3 align-top">
        <div className="flex gap-2">
          <Button variant="secondary" onClick={() => onEdit(secret)}>
            Edit
          </Button>
          <Button variant="danger" onClick={() => onDelete(secret)}>
            Delete
          </Button>
        </div>
      </td>
    </tr>
  )
}

function SecretValueCell({
  type,
  detail,
  revealed,
  onToggle,
}: {
  type: SecretType
  detail: SecretWithValue | null
  revealed: boolean
  onToggle: () => void
}) {
  if (type === 'opaque') {
    return <MaskedValue value={detail?.value ?? ''} revealed={revealed} onToggle={onToggle} />
  }

  if (!revealed) {
    return (
      <button onClick={onToggle} type="button" className="text-xs font-medium text-accent hover:underline">
        Reveal {detail?.fields?.length ?? ''} field(s)
      </button>
    )
  }

  return (
    <div className="space-y-1 rounded-md border border-warn bg-warn-bg p-2">
      {(detail?.fields ?? []).map((f) => (
        <div key={f.key} className="flex items-start gap-2">
          <code className="w-1/3 shrink-0 truncate text-xs text-text-muted">{f.key}</code>
          <code className="max-h-40 flex-1 overflow-auto whitespace-pre-wrap break-all font-mono text-xs text-warn">
            {f.value}
          </code>
          <CopyButton value={f.value} />
        </div>
      ))}
      <button onClick={onToggle} type="button" className="text-xs font-medium text-accent hover:underline">
        Hide
      </button>
    </div>
  )
}
