import { useEffect, useState } from 'react'
import { api } from '../../lib/api'
import { errorMessage } from '../../lib/errors'
import type { ApiKeyMeta, HostMeta, ResourceGrant, SecretMeta } from '../../types/api'
import { Modal } from '../../components/common/Modal'
import { Button } from '../../components/common/Button'
import { ErrorBanner } from '../../components/common/ErrorBanner'
import { GrantListEditor, type GrantRow } from '../../components/common/GrantListEditor'

interface ManageAccessModalProps {
  apiKey: ApiKeyMeta
  onClose: () => void
}

export function ManageAccessModal({ apiKey, onClose }: ManageAccessModalProps) {
  const [secretNames, setSecretNames] = useState<string[]>([])
  const [hostNames, setHostNames] = useState<string[]>([])
  const [secretGrants, setSecretGrants] = useState<GrantRow[]>([])
  const [hostGrants, setHostGrants] = useState<GrantRow[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [submitting, setSubmitting] = useState(false)

  useEffect(() => {
    async function load() {
      setLoading(true)
      setError(null)
      try {
        const [secrets, hosts, grants] = await Promise.all([
          api.get<SecretMeta[]>('/api/secrets'),
          api.get<HostMeta[]>('/api/hosts'),
          api.get<ResourceGrant[]>(`/api/apikeys/${apiKey.id}/access`),
        ])
        setSecretNames(secrets.map((s) => s.name))
        setHostNames(hosts.map((h) => h.name))
        setSecretGrants(
          grants.filter((g) => g.resource_type === 'secret').map((g) => ({ name: g.resource_name, canWrite: g.can_write })),
        )
        setHostGrants(
          grants.filter((g) => g.resource_type === 'host').map((g) => ({ name: g.resource_name, canWrite: g.can_write })),
        )
      } catch (err) {
        setError(errorMessage(err))
      } finally {
        setLoading(false)
      }
    }
    load()
  }, [apiKey.id])

  async function handleSave() {
    setError(null)
    setSubmitting(true)
    try {
      const grants: ResourceGrant[] = [
        ...secretGrants.filter((g) => g.name).map((g) => ({ resource_type: 'secret' as const, resource_name: g.name, can_write: g.canWrite })),
        ...hostGrants.filter((g) => g.name).map((g) => ({ resource_type: 'host' as const, resource_name: g.name, can_write: g.canWrite })),
      ]
      await api.put(`/api/apikeys/${apiKey.id}/access`, { grants })
      onClose()
    } catch (err) {
      setError(errorMessage(err))
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <Modal title={`Access — ${apiKey.name}`} onClose={onClose} widthClass="max-w-lg">
      {loading ? (
        <div className="text-sm text-text-muted">Loading…</div>
      ) : (
        <div className="space-y-5">
          <p className="text-xs text-text-muted">
            Leaving a section empty means unrestricted (this key can reach every secret/host, same as today).
            Adding even one grant turns that section into an allowlist — only listed names are visible to this
            key at all.
          </p>

          <div>
            <h3 className="mb-2 text-sm font-semibold text-text-bright">Secrets</h3>
            <GrantListEditor label="Secret" options={secretNames} grants={secretGrants} onChange={setSecretGrants} />
          </div>

          <div>
            <h3 className="mb-2 text-sm font-semibold text-text-bright">Hosts</h3>
            <GrantListEditor label="Host" options={hostNames} grants={hostGrants} onChange={setHostGrants} />
          </div>

          <ErrorBanner message={error} />

          <div className="flex justify-end gap-2 pt-2">
            <Button type="button" variant="secondary" onClick={onClose}>
              Cancel
            </Button>
            <Button type="button" onClick={handleSave} disabled={submitting}>
              {submitting ? 'Saving…' : 'Save'}
            </Button>
          </div>
        </div>
      )}
    </Modal>
  )
}
