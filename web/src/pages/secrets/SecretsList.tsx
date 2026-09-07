import { useEffect, useMemo, useState } from 'react'
import { api } from '../../lib/api'
import { errorMessage } from '../../lib/errors'
import type { SecretMeta } from '../../types/api'
import { Button } from '../../components/common/Button'
import { ErrorBanner } from '../../components/common/ErrorBanner'
import { TextInput } from '../../components/common/FormField'
import { ConfirmDialog } from '../../components/common/ConfirmDialog'
import { SecretForm } from './SecretForm'
import { SecretRow } from './SecretRow'

export function SecretsList() {
  const [secrets, setSecrets] = useState<SecretMeta[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [search, setSearch] = useState('')
  const [formOpen, setFormOpen] = useState(false)
  const [editing, setEditing] = useState<SecretMeta | null>(null)
  const [pendingDelete, setPendingDelete] = useState<SecretMeta | null>(null)

  async function load() {
    setLoading(true)
    setError(null)
    try {
      const result = await api.get<SecretMeta[]>('/api/secrets')
      setSecrets(result)
    } catch (err) {
      setError(errorMessage(err))
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    load()
  }, [])

  const filtered = useMemo(() => {
    const q = search.trim().toLowerCase()
    if (!q) return secrets
    return secrets.filter(
      (s) => s.name.toLowerCase().includes(q) || s.tags.some((t) => t.toLowerCase().includes(q)),
    )
  }, [secrets, search])

  function handleAdd() {
    setEditing(null)
    setFormOpen(true)
  }

  function handleEdit(secret: SecretMeta) {
    setEditing(secret)
    setFormOpen(true)
  }

  function handleFormSaved() {
    setFormOpen(false)
    setEditing(null)
    load()
  }

  async function handleConfirmDelete() {
    if (!pendingDelete) return
    try {
      await api.del(`/api/secrets/${pendingDelete.id}`)
      setPendingDelete(null)
      load()
    } catch (err) {
      setError(errorMessage(err))
      setPendingDelete(null)
    }
  }

  return (
    <div>
      <div className="mb-5 flex items-center justify-between">
        <h1 className="text-xl font-semibold text-text-bright">Secrets</h1>
        <Button onClick={handleAdd}>Add secret</Button>
      </div>

      <div className="mb-4">
        <TextInput
          placeholder="Search by name or tag…"
          value={search}
          onChange={(e) => setSearch(e.target.value)}
          className="max-w-xs"
        />
      </div>

      <ErrorBanner message={error} />

      <div className="mt-3 overflow-x-auto rounded-lg border border-border bg-surface">
        {loading ? (
          <div className="p-6 text-sm text-text-muted">Loading…</div>
        ) : filtered.length === 0 ? (
          <div className="p-6 text-sm text-text-muted">No secrets found.</div>
        ) : (
          <table className="w-full text-left text-sm">
            <thead>
              <tr className="border-b border-border text-xs uppercase tracking-wide text-text-muted">
                <th className="px-4 py-2 font-medium">Name</th>
                <th className="px-4 py-2 font-medium">Tags</th>
                <th className="px-4 py-2 font-medium">Value</th>
                <th className="px-4 py-2 font-medium">Updated</th>
                <th className="px-4 py-2 font-medium">Actions</th>
              </tr>
            </thead>
            <tbody>
              {filtered.map((secret) => (
                <SecretRow
                  key={secret.id}
                  secret={secret}
                  onEdit={handleEdit}
                  onDelete={setPendingDelete}
                />
              ))}
            </tbody>
          </table>
        )}
      </div>

      {formOpen && (
        <SecretForm
          editing={editing}
          onClose={() => setFormOpen(false)}
          onSaved={handleFormSaved}
        />
      )}

      {pendingDelete && (
        <ConfirmDialog
          title="Delete secret"
          message={`Delete "${pendingDelete.name}"? This cannot be undone.`}
          confirmLabel="Delete"
          onConfirm={handleConfirmDelete}
          onCancel={() => setPendingDelete(null)}
        />
      )}
    </div>
  )
}
