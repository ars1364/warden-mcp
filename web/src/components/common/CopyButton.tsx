import { useState } from 'react'

export function CopyButton({ value }: { value: string }) {
  const [copied, setCopied] = useState(false)

  async function handleCopy() {
    try {
      await navigator.clipboard.writeText(value)
      setCopied(true)
      setTimeout(() => setCopied(false), 1500)
    } catch {
      setCopied(false)
    }
  }

  return (
    <button
      onClick={handleCopy}
      type="button"
      className="rounded-md border border-border px-2.5 py-1 text-xs font-medium text-text hover:bg-surface-hover"
    >
      {copied ? 'Copied!' : 'Copy'}
    </button>
  )
}
