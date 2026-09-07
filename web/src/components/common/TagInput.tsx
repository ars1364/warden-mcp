import { useState, type KeyboardEvent } from 'react'

interface TagInputProps {
  tags: string[]
  onChange: (tags: string[]) => void
}

export function TagInput({ tags, onChange }: TagInputProps) {
  const [draft, setDraft] = useState('')

  function commitDraft() {
    const parts = draft
      .split(',')
      .map((p) => p.trim())
      .filter((p) => p.length > 0 && !tags.includes(p))
    if (parts.length > 0) {
      onChange([...tags, ...parts])
    }
    setDraft('')
  }

  function handleKeyDown(e: KeyboardEvent<HTMLInputElement>) {
    if (e.key === 'Enter' || e.key === ',') {
      e.preventDefault()
      commitDraft()
    } else if (e.key === 'Backspace' && draft === '' && tags.length > 0) {
      onChange(tags.slice(0, -1))
    }
  }

  function removeTag(tag: string) {
    onChange(tags.filter((t) => t !== tag))
  }

  return (
    <div className="flex flex-wrap items-center gap-1.5 rounded-md border border-border bg-bg px-2 py-1.5 focus-within:border-accent">
      {tags.map((tag) => (
        <span
          key={tag}
          className="flex items-center gap-1 rounded-full bg-surface-hover px-2 py-0.5 text-xs text-text"
        >
          {tag}
          <button
            type="button"
            onClick={() => removeTag(tag)}
            className="text-text-muted hover:text-danger"
            aria-label={`Remove tag ${tag}`}
          >
            ×
          </button>
        </span>
      ))}
      <input
        value={draft}
        onChange={(e) => setDraft(e.target.value)}
        onKeyDown={handleKeyDown}
        onBlur={commitDraft}
        placeholder={tags.length === 0 ? 'tags, comma-separated' : ''}
        className="min-w-[80px] flex-1 bg-transparent text-sm text-text outline-none placeholder:text-text-muted"
      />
    </div>
  )
}
