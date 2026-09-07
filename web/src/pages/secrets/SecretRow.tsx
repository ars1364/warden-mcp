import { useState } from 'react'
import { api } from '../../lib/api'
import { errorMessage } from '../../lib/errors'
import type { SecretMeta, SecretWithValue } from '../../types/api'
import { MaskedValue } from '../../components/common/MaskedValue'
import { TagChips } from '../../components/common/TagChips'
import { Button } from '../../components/common/Button'

interface SecretRowProps {
  secret: SecretMeta
  onEdit: (secret: SecretMeta) => void
  onDelete: (secret: SecretMeta) => void
}

export function SecretRow({ secret, onEdit, onDelete }: SecretRowProps) {
  const [revealed, setRevealed] = useState(false)
  const [value, setValue] = useState<string | null>(null)
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)

  async function handleToggleReveal() {
    if (revealed) {
      setRevealed(false)
      return
    }
    if (value !== null) {
      setRevealed(true)
      return
    }
    setLoading(true)
    setError(null)
    try {
      const full = await api.get<SecretWithValue>(`/api/secrets/${secret.id}`)
      setValue(full.value)
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
        <div className="font-medium text-text-bright">{secret.name}</div>
        <div className="text-xs text-text-muted">{secret.description}</div>
      </td>
      <td className="px-4 py-3 align-top">
        <TagChips tags={secret.tags} />
      </td>
      <td className="px-4 py-3 align-top">
        <div className="w-56">
          {loading ? (
            <span className="text-xs text-text-muted">Loading…</span>
          ) : error ? (
            <span className="text-xs text-danger">{error}</span>
          ) : (
            <MaskedValue value={value ?? ''} revealed={revealed} onToggle={handleToggleReveal} />
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
