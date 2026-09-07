import { useEffect, useState } from 'react'
import { api } from '../../lib/api'
import { errorMessage } from '../../lib/errors'
import type { ApiKeyMeta } from '../../types/api'
import { Button } from '../../components/common/Button'
import { ErrorBanner } from '../../components/common/ErrorBanner'
import { ConfirmDialog } from '../../components/common/ConfirmDialog'
import { CreateApiKeyModal } from './CreateApiKeyModal'

export function ApiKeysList() {
  const [keys, setKeys] = useState<ApiKeyMeta[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [createOpen, setCreateOpen] = useState(false)
  const [pendingRevoke, setPendingRevoke] = useState<ApiKeyMeta | null>(null)

  async function load() {
    setLoading(true)
    setError(null)
    try {
      const result = await api.get<ApiKeyMeta[]>('/api/apikeys')
      setKeys(result)
    } catch (err) {
      setError(errorMessage(err))
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    load()
  }, [])

  function handleCreated() {
    setCreateOpen(false)
    load()
  }

  async function handleConfirmRevoke() {
    if (!pendingRevoke) return
    try {
      await api.del(`/api/apikeys/${pendingRevoke.id}`)
      setPendingRevoke(null)
      load()
    } catch (err) {
      setError(errorMessage(err))
      setPendingRevoke(null)
    }
  }

  return (
    <div>
      <div className="mb-5 flex items-center justify-between">
        <h1 className="text-xl font-semibold text-text-bright">API Keys</h1>
        <Button onClick={() => setCreateOpen(true)}>Create key</Button>
      </div>

      <ErrorBanner message={error} />

      <div className="mt-3 overflow-x-auto rounded-lg border border-border bg-surface">
        {loading ? (
          <div className="p-6 text-sm text-text-muted">Loading…</div>
        ) : keys.length === 0 ? (
          <div className="p-6 text-sm text-text-muted">No API keys yet.</div>
        ) : (
          <table className="w-full text-left text-sm">
            <thead>
              <tr className="border-b border-border text-xs uppercase tracking-wide text-text-muted">
                <th className="px-4 py-2 font-medium">Name</th>
                <th className="px-4 py-2 font-medium">Scopes</th>
                <th className="px-4 py-2 font-medium">Created</th>
                <th className="px-4 py-2 font-medium">Last used</th>
                <th className="px-4 py-2 font-medium">Status</th>
                <th className="px-4 py-2 font-medium">Actions</th>
              </tr>
            </thead>
            <tbody>
              {keys.map((k) => (
                <tr key={k.id} className="border-b border-border last:border-0">
                  <td className="px-4 py-3 font-medium text-text-bright">{k.name}</td>
                  <td className="px-4 py-3 text-text">{k.scopes.join(', ')}</td>
                  <td className="px-4 py-3 text-xs text-text-muted">
                    {new Date(k.created_at).toLocaleString()}
                  </td>
                  <td className="px-4 py-3 text-xs text-text-muted">
                    {k.last_used_at ? new Date(k.last_used_at).toLocaleString() : 'Never'}
                    {k.last_used_ip ? ` · ${k.last_used_ip}` : ''}
                  </td>
                  <td className="px-4 py-3">
                    {k.revoked_at ? (
                      <span className="rounded-full bg-danger-bg px-2 py-0.5 text-xs text-danger">
                        Revoked
                      </span>
                    ) : (
                      <span className="rounded-full bg-success/15 px-2 py-0.5 text-xs text-success">
                        Active
                      </span>
                    )}
                  </td>
                  <td className="px-4 py-3">
                    {!k.revoked_at && (
                      <Button variant="danger" onClick={() => setPendingRevoke(k)}>
                        Revoke
                      </Button>
                    )}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        )}
      </div>

      {createOpen && (
        <CreateApiKeyModal onClose={() => setCreateOpen(false)} onCreated={handleCreated} />
      )}

      {pendingRevoke && (
        <ConfirmDialog
          title="Revoke API key"
          message={`Revoke "${pendingRevoke.name}"? Anything using this key will stop working.`}
          confirmLabel="Revoke"
          onConfirm={handleConfirmRevoke}
          onCancel={() => setPendingRevoke(null)}
        />
      )}
    </div>
  )
}
