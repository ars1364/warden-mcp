import { useState, type FormEvent } from 'react'
import { api } from '../../lib/api'
import { errorMessage } from '../../lib/errors'
import type { SecretField, SecretMeta, SecretType, SecretWithValue } from '../../types/api'
import { Modal } from '../../components/common/Modal'
import { Button } from '../../components/common/Button'
import { ErrorBanner } from '../../components/common/ErrorBanner'
import { FormField, TextInput, Textarea } from '../../components/common/FormField'
import { TagInput } from '../../components/common/TagInput'
import { FieldListEditor } from '../../components/common/FieldListEditor'

interface SecretFormProps {
  editing: SecretMeta | null
  onClose: () => void
  onSaved: () => void
}

const TYPE_OPTIONS: { value: SecretType; label: string; hint: string }[] = [
  { value: 'opaque', label: 'Opaque', hint: 'A single value — API keys, tokens.' },
  { value: 'structured', label: 'Structured', hint: 'Multiple named fields — username/password, client id/secret.' },
  { value: 'totp', label: 'TOTP', hint: 'A 2FA seed — Claude can fetch a live 6-digit code without ever seeing the seed.' },
  { value: 'reference', label: 'Reference', hint: "Not a secret — a login URL and instructions for creds that live elsewhere in infra." },
]

export function SecretForm({ editing, onClose, onSaved }: SecretFormProps) {
  const [name, setName] = useState(editing?.name ?? '')
  const [description, setDescription] = useState(editing?.description ?? '')
  const [tags, setTags] = useState<string[]>(editing?.tags ?? [])
  const [type, setType] = useState<SecretType>(editing?.type ?? 'opaque')
  const [value, setValue] = useState('')
  const [fields, setFields] = useState<SecretField[]>(
    editing?.type === 'totp' ? [{ key: 'seed', value: '' }] : [],
  )
  const [error, setError] = useState<string | null>(null)
  const [submitting, setSubmitting] = useState(false)

  function handleTypeChange(next: SecretType) {
    setType(next)
    if (next === 'totp' && fields.length === 0) {
      setFields([{ key: 'seed', value: '' }])
    }
  }

  function validate(): string | null {
    if (type !== 'opaque') {
      if (fields.length === 0) return 'Add at least one field.'
      if (fields.some((f) => !f.key.trim())) return 'Every field needs a key.'
      if (type === 'totp' && !fields.some((f) => f.key === 'seed' && f.value.trim())) {
        return 'A totp credential requires a non-empty "seed" field.'
      }
    }
    return null
  }

  async function handleSubmit(e: FormEvent) {
    e.preventDefault()
    setError(null)
    const validationError = validate()
    if (validationError) {
      setError(validationError)
      return
    }
    setSubmitting(true)
    try {
      if (editing) {
        await api.put<SecretWithValue>(`/api/secrets/${editing.id}`, {
          description,
          tags,
          type,
          value,
          fields,
        })
      } else {
        await api.post<SecretWithValue>('/api/secrets', { name, description, tags, type, value, fields })
      }
      onSaved()
    } catch (err) {
      setError(errorMessage(err))
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <Modal title={editing ? `Edit ${editing.name}` : 'Add secret'} onClose={onClose}>
      <form onSubmit={handleSubmit} className="space-y-4">
        <FormField label="Name">
          <TextInput
            value={name}
            onChange={(e) => setName(e.target.value)}
            disabled={!!editing}
            required
          />
        </FormField>
        <FormField label="Description">
          <TextInput value={description} onChange={(e) => setDescription(e.target.value)} />
        </FormField>
        <FormField label="Tags">
          <TagInput tags={tags} onChange={setTags} />
        </FormField>

        <FormField label="Type">
          <select
            value={type}
            onChange={(e) => handleTypeChange(e.target.value as SecretType)}
            disabled={!!editing}
            className="w-full rounded-md border border-border bg-bg px-3 py-1.5 text-sm text-text outline-none focus:border-accent"
          >
            {TYPE_OPTIONS.map((opt) => (
              <option key={opt.value} value={opt.value}>
                {opt.label}
              </option>
            ))}
          </select>
          <span className="mt-1 block text-xs text-text-muted">
            {TYPE_OPTIONS.find((o) => o.value === type)?.hint}
          </span>
        </FormField>

        {type === 'opaque' ? (
          <FormField label="Value">
            <Textarea
              value={value}
              onChange={(e) => setValue(e.target.value)}
              placeholder={editing ? 'Re-enter value to save changes' : 'Paste any value — single line or multi-line (SSH keys, certs, JSON)'}
              className="font-mono"
              required
            />
          </FormField>
        ) : (
          <FormField label="Fields">
            {editing && (
              <span className="mb-1 block text-xs text-text-muted">
                Re-enter every field to save changes — this replaces all stored fields.
              </span>
            )}
            <FieldListEditor
              fields={fields}
              onChange={setFields}
              keyPlaceholder={type === 'totp' ? 'seed' : 'key'}
            />
          </FormField>
        )}

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
