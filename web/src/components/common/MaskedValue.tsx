import { useState } from 'react'
import { CopyButton } from './CopyButton'

interface MaskedValueProps {
  value: string
  revealed: boolean
  onToggle: () => void
}

export function MaskedValue({ value, revealed, onToggle }: MaskedValueProps) {
  const [dots] = useState('•'.repeat(24))

  return (
    <div
      className={`flex items-center gap-2 rounded-md border px-2 py-1 ${
        revealed ? 'border-warn bg-warn-bg' : 'border-border bg-bg'
      }`}
    >
      <code
        className={`max-h-40 flex-1 overflow-auto whitespace-pre-wrap break-all font-mono text-xs ${
          revealed ? 'text-warn' : 'truncate text-text-muted'
        }`}
      >
        {revealed ? value : dots}
      </code>
      <button
        onClick={onToggle}
        type="button"
        className="shrink-0 text-xs font-medium text-accent hover:underline"
      >
        {revealed ? 'Hide' : 'Reveal'}
      </button>
      {revealed && <CopyButton value={value} />}
    </div>
  )
}
