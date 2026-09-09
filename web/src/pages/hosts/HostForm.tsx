import { useState, type FormEvent, type SelectHTMLAttributes } from 'react'
import { api } from '../../lib/api'
import { errorMessage } from '../../lib/errors'
import type { HostAddress, HostInput, HostMeta, HostStatus, HostType } from '../../types/api'
import { Modal } from '../../components/common/Modal'
import { Button } from '../../components/common/Button'
import { ErrorBanner } from '../../components/common/ErrorBanner'
import { FormField, TextInput } from '../../components/common/FormField'
import { TagInput } from '../../components/common/TagInput'
import { FieldListEditor } from '../../components/common/FieldListEditor'

interface HostFormProps {
  editing: HostMeta | null
  allHosts: HostMeta[]
  onClose: () => void
  onSaved: () => void
}

const HOST_TYPES: HostType[] = ['physical', 'vm', 'vps', 'container', 'switch', 'router', 'other']
const STATUSES: HostStatus[] = ['active', 'maintenance', 'decommissioned']

function Select(props: SelectHTMLAttributes<HTMLSelectElement>) {
  const { className = '', ...rest } = props
  return (
    <select
      {...rest}
      className={`w-full rounded-md border border-border bg-bg px-3 py-1.5 text-sm text-text outline-none focus:border-accent ${className}`}
    />
  )
}

export function HostForm({ editing, allHosts, onClose, onSaved }: HostFormProps) {
  const [name, setName] = useState(editing?.name ?? '')
  const [hostType, setHostType] = useState<HostType>(editing?.host_type ?? 'other')
  const [status, setStatus] = useState<HostStatus>(editing?.status ?? 'active')
  const [description, setDescription] = useState(editing?.description ?? '')
  const [tags, setTags] = useState<string[]>(editing?.tags ?? [])
  const [parentHostName, setParentHostName] = useState(editing?.parent_host_name ?? '')
  const [locationKind, setLocationKind] = useState(editing?.location_kind ?? '')
  const [cloudProvider, setCloudProvider] = useState(editing?.cloud_provider ?? '')
  const [cloudAccount, setCloudAccount] = useState(editing?.cloud_account ?? '')
  const [physicalLocation, setPhysicalLocation] = useState(editing?.physical_location ?? '')
  const [addresses, setAddresses] = useState<HostAddress[]>(editing?.addresses ?? [])
  const [sshUsername, setSshUsername] = useState(editing?.ssh_username ?? '')
  const [sshPort, setSshPort] = useState(editing?.ssh_port ?? 22)
  const [sshSecretName, setSshSecretName] = useState(editing?.ssh_secret_name ?? '')
  const [sshJumpHostName, setSshJumpHostName] = useState(editing?.ssh_jump_host_name ?? '')
  const [error, setError] = useState<string | null>(null)
  const [submitting, setSubmitting] = useState(false)

  const otherHosts = allHosts.filter((h) => h.id !== editing?.id)

  async function handleSubmit(e: FormEvent) {
    e.preventDefault()
    setError(null)
    setSubmitting(true)
    const body: HostInput = {
      name, host_type: hostType, status, description, tags,
      parent_host_name: parentHostName, location_kind: locationKind,
      cloud_provider: cloudProvider, cloud_account: cloudAccount, physical_location: physicalLocation,
      ssh_port: sshPort, ssh_username: sshUsername, ssh_secret_name: sshSecretName,
      ssh_jump_host_name: sshJumpHostName, addresses,
    }
    try {
      if (editing) {
        await api.put<HostMeta>(`/api/hosts/${editing.id}`, body)
      } else {
        await api.post<HostMeta>('/api/hosts', body)
      }
      onSaved()
    } catch (err) {
      setError(errorMessage(err))
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <Modal title={editing ? `Edit ${editing.name}` : 'Add host'} onClose={onClose}>
      <form onSubmit={handleSubmit} className="space-y-4">
        <div className="grid grid-cols-3 gap-3">
          <FormField label="Name">
            <TextInput value={name} onChange={(e) => setName(e.target.value)} disabled={!!editing} required />
          </FormField>
          <FormField label="Type">
            <Select value={hostType} onChange={(e) => setHostType(e.target.value as HostType)}>
              {HOST_TYPES.map((t) => (
                <option key={t} value={t}>{t}</option>
              ))}
            </Select>
          </FormField>
          <FormField label="Status">
            <Select value={status} onChange={(e) => setStatus(e.target.value as HostStatus)}>
              {STATUSES.map((s) => (
                <option key={s} value={s}>{s}</option>
              ))}
            </Select>
          </FormField>
        </div>

        <FormField label="Description">
          <TextInput value={description} onChange={(e) => setDescription(e.target.value)} />
        </FormField>
        <FormField label="Tags">
          <TagInput tags={tags} onChange={setTags} />
        </FormField>

        <div className="grid grid-cols-2 gap-3">
          <FormField label="Lives inside (parent host)">
            <Select value={parentHostName} onChange={(e) => setParentHostName(e.target.value)}>
              <option value="">— none —</option>
              {otherHosts.map((h) => (
                <option key={h.id} value={h.name}>{h.name}</option>
              ))}
            </Select>
          </FormField>
          <FormField label="Location kind">
            <Select value={locationKind} onChange={(e) => setLocationKind(e.target.value)}>
              <option value="">— not set —</option>
              <option value="on_prem">on_prem</option>
              <option value="cloud">cloud</option>
              <option value="colo">colo</option>
            </Select>
          </FormField>
        </div>

        <div className="grid grid-cols-2 gap-3">
          <FormField label="Cloud provider">
            <TextInput value={cloudProvider} onChange={(e) => setCloudProvider(e.target.value)} placeholder="Hetzner, AWS, OpenStack…" />
          </FormField>
          <FormField label="Cloud account">
            <TextInput value={cloudAccount} onChange={(e) => setCloudAccount(e.target.value)} />
          </FormField>
        </div>
        <FormField label="Physical location">
          <TextInput
            value={physicalLocation}
            onChange={(e) => setPhysicalLocation(e.target.value)}
            placeholder="office2 server room, rack 3…"
          />
        </FormField>

        <FormField label="Addresses">
          <FieldListEditor
            fields={addresses.map((a) => ({ key: a.label, value: a.address }))}
            onChange={(fields) => setAddresses(fields.map((f) => ({ label: f.key, address: f.value })))}
            keyPlaceholder="label (public, private, mgmt…)"
            valuePlaceholder="IP or hostname"
          />
        </FormField>

        <div className="grid grid-cols-3 gap-3">
          <FormField label="SSH username">
            <TextInput value={sshUsername} onChange={(e) => setSshUsername(e.target.value)} />
          </FormField>
          <FormField label="SSH port">
            <TextInput
              type="number"
              value={sshPort}
              onChange={(e) => setSshPort(Number(e.target.value) || 22)}
            />
          </FormField>
          <FormField label="Jump host">
            <Select value={sshJumpHostName} onChange={(e) => setSshJumpHostName(e.target.value)}>
              <option value="">— direct —</option>
              {otherHosts.map((h) => (
                <option key={h.id} value={h.name}>{h.name}</option>
              ))}
            </Select>
          </FormField>
        </div>
        <FormField label="SSH credential (secret name)">
          <TextInput
            value={sshSecretName}
            onChange={(e) => setSshSecretName(e.target.value)}
            placeholder="name of a secret already in the vault, e.g. ahmad_ssh_private_key"
            className="font-mono"
          />
          <span className="mt-1 block text-xs text-text-muted">
            Points at an existing secret — the key/password itself isn't stored here.
          </span>
        </FormField>

        <ErrorBanner message={error} />

        <div className="flex justify-end gap-2 pt-2">
          <Button type="button" variant="secondary" onClick={onClose}>
            Cancel
          </Button>
          <Button type="submit" disabled={submitting}>
            {submitting ? 'Saving…' : 'Save'}
          </Button>
        </div>
      </form>
    </Modal>
  )
}
