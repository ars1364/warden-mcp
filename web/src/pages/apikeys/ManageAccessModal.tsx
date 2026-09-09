import { useEffect, useState } from 'react'
import { api } from '../../lib/api'
import { errorMessage } from '../../lib/errors'
import type { ApiKeyMeta, HostMeta, ResourceGrant, SecretMeta } from '../../types/api'
import { Modal } from '../../components/common/Modal'
import { Button } from '../../components/common/Button'
import { ErrorBanner } from '../../components/common/ErrorBanner'
import { ResourceAccessPicker, initGrantState, grantStateToPayload, type GrantState } from '../../components/common/ResourceAccessPicker'

interface ManageAccessModalProps {
  apiKey: ApiKeyMeta
  onClose: () => void
}

export function ManageAccessModal({ apiKey, onClose }: ManageAccessModalProps) {
  const [secretNames, setSecretNames] = useState<string[]>([])
  const [hostNames, setHostNames] = useState<string[]>([])
  const [secretState, setSecretState] = useState<GrantState>({})
  const [hostState, setHostState] = useState<GrantState>({})
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
        const sNames = secrets.map((s) => s.name)
        const hNames = hosts.map((h) => h.name)
        setSecretNames(sNames)
        setHostNames(hNames)
        setSecretState(initGrantState(sNames, grants.filter((g) => g.resource_type === 'secret')))
        setHostState(initGrantState(hNames, grants.filter((g) => g.resource_type === 'host')))
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
        ...grantStateToPayload(secretNames, secretState).map((g) => ({ resource_type: 'secret' as const, ...g })),
        ...grantStateToPayload(hostNames, hostState).map((g) => ({ resource_type: 'host' as const, ...g })),
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
    <Modal title={`Access — ${apiKey.name}`} onClose={onClose} widthClass="max-w-xl">
      {loading ? (
        <div className="text-sm text-text-muted">Loading…</div>
      ) : (
        <div className="space-y-4">
          <ResourceAccessPicker
            secretNames={secretNames}
            hostNames={hostNames}
            secretState={secretState}
            hostState={hostState}
            onChangeSecretState={setSecretState}
            onChangeHostState={setHostState}
          />

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
