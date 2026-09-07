import { useState, type FormEvent } from 'react'
import { api } from '../../lib/api'
import { errorMessage } from '../../lib/errors'
import type { SecretMeta, SecretWithValue } from '../../types/api'
import { Modal } from '../../components/common/Modal'
import { Button } from '../../components/common/Button'
import { ErrorBanner } from '../../components/common/ErrorBanner'
import { FormField, TextInput } from '../../components/common/FormField'
import { TagInput } from '../../components/common/TagInput'

interface SecretFormProps {
  editing: SecretMeta | null
  onClose: () => void
  onSaved: () => void
}

export function SecretForm({ editing, onClose, onSaved }: SecretFormProps) {
  const [name, setName] = useState(editing?.name ?? '')
  const [description, setDescription] = useState(editing?.description ?? '')
  const [tags, setTags] = useState<string[]>(editing?.tags ?? [])
  const [value, setValue] = useState('')
  const [error, setError] = useState<string | null>(null)
  const [submitting, setSubmitting] = useState(false)

  async function handleSubmit(e: FormEvent) {
    e.preventDefault()
    setError(null)
    setSubmitting(true)
    try {
      if (editing) {
        await api.put<SecretWithValue>(`/api/secrets/${editing.id}`, {
          description,
          tags,
          value,
        })
      } else {
        await api.post<SecretWithValue>('/api/secrets', { name, description, tags, value })
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
        <FormField label="Value">
          <TextInput
            type="password"
            value={value}
            onChange={(e) => setValue(e.target.value)}
            placeholder={editing ? 'Re-enter value to save changes' : undefined}
            className="font-mono"
            required
          />
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
