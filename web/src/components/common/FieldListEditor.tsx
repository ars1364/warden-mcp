import type { SecretField } from '../../types/api'
import { TextInput, Textarea } from './FormField'
import { Button } from './Button'

interface FieldListEditorProps {
  fields: SecretField[]
  onChange: (fields: SecretField[]) => void
  keyPlaceholder?: string
  valuePlaceholder?: string
}

export function FieldListEditor({
  fields,
  onChange,
  keyPlaceholder = 'key',
  valuePlaceholder = 'value',
}: FieldListEditorProps) {
  function update(index: number, patch: Partial<SecretField>) {
    onChange(fields.map((f, i) => (i === index ? { ...f, ...patch } : f)))
  }

  function remove(index: number) {
    onChange(fields.filter((_, i) => i !== index))
  }

  function add() {
    onChange([...fields, { key: '', value: '' }])
  }

  return (
    <div className="space-y-2">
      {fields.map((field, i) => (
        <div key={i} className="flex items-start gap-2">
          <TextInput
            value={field.key}
            onChange={(e) => update(i, { key: e.target.value })}
            placeholder={keyPlaceholder}
            className="w-1/3 font-mono"
          />
          <Textarea
            value={field.value}
            onChange={(e) => update(i, { value: e.target.value })}
            placeholder={valuePlaceholder}
            rows={1}
            className="flex-1 font-mono"
          />
          <Button type="button" variant="ghost" onClick={() => remove(i)} aria-label="Remove field">
            ×
          </Button>
        </div>
      ))}
      <Button type="button" variant="secondary" onClick={add}>
        Add field
      </Button>
    </div>
  )
}
