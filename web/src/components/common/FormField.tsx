import type { InputHTMLAttributes, ReactNode, TextareaHTMLAttributes } from 'react'

interface FormFieldProps {
  label: string
  children: ReactNode
  error?: string
}

export function FormField({ label, children, error }: FormFieldProps) {
  return (
    <label className="block text-sm">
      <span className="mb-1 block font-medium text-text">{label}</span>
      {children}
      {error && <span className="mt-1 block text-xs text-danger">{error}</span>}
    </label>
  )
}

export function TextInput(props: InputHTMLAttributes<HTMLInputElement>) {
  const { className = '', ...rest } = props
  return (
    <input
      {...rest}
      className={`w-full rounded-md border border-border bg-bg px-3 py-1.5 text-sm text-text outline-none focus:border-accent ${className}`}
    />
  )
}

// A single-line <input> silently drops newlines on paste in most browsers — use this
// instead of TextInput for anything that might be multi-line (SSH keys, certs, JSON blobs).
export function Textarea(props: TextareaHTMLAttributes<HTMLTextAreaElement>) {
  const { className = '', rows = 4, ...rest } = props
  return (
    <textarea
      {...rest}
      rows={rows}
      className={`w-full resize-y rounded-md border border-border bg-bg px-3 py-1.5 text-sm text-text outline-none focus:border-accent ${className}`}
    />
  )
}
