import type { ReactNode } from 'react'

interface ModalProps {
  title: string
  onClose: () => void
  children: ReactNode
  widthClass?: string
}

export function Modal({ title, onClose, children, widthClass = 'max-w-md' }: ModalProps) {
  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 p-4">
      <div className={`flex max-h-[90vh] w-full ${widthClass} flex-col rounded-lg border border-border bg-surface shadow-xl`}>
        <div className="flex shrink-0 items-center justify-between border-b border-border px-4 py-3 sm:px-5">
          <h2 className="text-sm font-semibold text-text-bright">{title}</h2>
          <button
            onClick={onClose}
            className="text-text-muted hover:text-text-bright"
            aria-label="Close"
          >
            ✕
          </button>
        </div>
        <div className="overflow-y-auto p-4 sm:p-5">{children}</div>
      </div>
    </div>
  )
}
