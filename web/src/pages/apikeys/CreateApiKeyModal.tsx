import { useEffect, useState, type FormEvent } from 'react'
import { api } from '../../lib/api'
import { errorMessage } from '../../lib/errors'
import type { ApiKeyCreated, ApiKeyScope, HostMeta, ResourceGrant, SecretMeta } from '../../types/api'
import { Modal } from '../../components/common/Modal'
import { Button } from '../../components/common/Button'
import { ErrorBanner } from '../../components/common/ErrorBanner'
import { FormField, TextInput } from '../../components/common/FormField'
import { CopyButton } from '../../components/common/CopyButton'
import { ResourceAccessPicker, initGrantState, grantStateToPayload, type GrantState } from '../../components/common/ResourceAccessPicker'

const ALL_SCOPES: ApiKeyScope[] = ['read', 'write']

interface CreateApiKeyModalProps {
  onClose: () => void
  onCreated: () => void
}

export function CreateApiKeyModal({ onClose, onCreated }: CreateApiKeyModalProps) {
  const [name, setName] = useState('')
  const [scopes, setScopes] = useState<ApiKeyScope[]>(['read'])
  const [secretNames, setSecretNames] = useState<string[]>([])
  const [hostNames, setHostNames] = useState<string[]>([])
  const [secretState, setSecretState] = useState<GrantState>({})
  const [hostState, setHostState] = useState<GrantState>({})
  const [error, setError] = useState<string | null>(null)
  const [submitting, setSubmitting] = useState(false)
  const [created, setCreated] = useState<ApiKeyCreated | null>(null)

  useEffect(() => {
    async function load() {
      try {
        const [secrets, hosts] = await Promise.all([
          api.get<SecretMeta[]>('/api/secrets'),
          api.get<HostMeta[]>('/api/hosts'),
        ])
        const sNames = secrets.map((s) => s.name)
        const hNames = hosts.map((h) => h.name)
        setSecretNames(sNames)
        setHostNames(hNames)
        setSecretState(initGrantState(sNames, [])) // no saved grants yet — starts fully checked
        setHostState(initGrantState(hNames, []))
      } catch (err) {
        setError(errorMessage(err))
      }
    }
    load()
  }, [])

  function toggleScope(scope: ApiKeyScope) {
    setScopes((prev) =>
      prev.includes(scope) ? prev.filter((s) => s !== scope) : [...prev, scope],
    )
  }

  async function handleSubmit(e: FormEvent) {
    e.preventDefault()
    setError(null)
    setSubmitting(true)
    try {
      const result = await api.post<ApiKeyCreated>('/api/apikeys', { name, scopes })
      const grants: ResourceGrant[] = [
        ...grantStateToPayload(secretNames, secretState).map((g) => ({ resource_type: 'secret' as const, ...g })),
        ...grantStateToPayload(hostNames, hostState).map((g) => ({ resource_type: 'host' as const, ...g })),
      ]
      await api.put(`/api/apikeys/${result.id}/access`, { grants })
      setCreated(result)
    } catch (err) {
      setError(errorMessage(err))
    } finally {
      setSubmitting(false)
    }
  }

  function handleDone() {
    onCreated()
  }

  if (created) {
    return (
      <Modal title="API key created" onClose={handleDone}>
        <div className="space-y-3">
          <div className="rounded-md border border-warn bg-warn-bg px-3 py-2 text-sm text-warn">
            This is the only time the full key is shown. Copy it now.
          </div>
          <div className="flex items-center gap-2 rounded-md border border-border bg-bg px-3 py-2">
            <code className="flex-1 overflow-x-auto whitespace-nowrap font-mono text-sm text-text-bright">
              {created.key}
            </code>
            <CopyButton value={created.key} />
          </div>
          <div className="flex justify-end pt-2">
            <Button onClick={handleDone}>Done</Button>
          </div>
        </div>
      </Modal>
    )
  }

  return (
    <Modal title="Create API key" onClose={onClose} widthClass="max-w-xl">
      <form onSubmit={handleSubmit} className="space-y-4">
        <FormField label="Name">
          <TextInput value={name} onChange={(e) => setName(e.target.value)} required />
        </FormField>

        <FormField label="Scopes">
          <div className="flex gap-4">
            {ALL_SCOPES.map((scope) => (
              <label key={scope} className="flex items-center gap-2 text-sm text-text">
                <input
                  type="checkbox"
                  checked={scopes.includes(scope)}
                  onChange={() => toggleScope(scope)}
                />
                {scope}
              </label>
            ))}
          </div>
        </FormField>

        <FormField label="Access (defaults to everything)">
          <ResourceAccessPicker
            secretNames={secretNames}
            hostNames={hostNames}
            secretState={secretState}
            hostState={hostState}
            onChangeSecretState={setSecretState}
            onChangeHostState={setHostState}
          />
        </FormField>

        <ErrorBanner message={error} />

        <div className="flex justify-end gap-2 pt-2">
          <Button type="button" variant="secondary" onClick={onClose}>
            Cancel
          </Button>
          <Button type="submit" disabled={submitting || scopes.length === 0}>
            {submitting ? 'Creating…' : 'Create'}
          </Button>
        </div>
      </form>
    </Modal>
  )
}
