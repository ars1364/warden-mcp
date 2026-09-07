export function TagChips({ tags }: { tags: string[] }) {
  if (tags.length === 0) {
    return <span className="text-xs text-text-muted">—</span>
  }
  return (
    <div className="flex flex-wrap gap-1">
      {tags.map((tag) => (
        <span
          key={tag}
          className="rounded-full border border-border bg-surface-hover px-2 py-0.5 text-xs text-text"
        >
          {tag}
        </span>
      ))}
    </div>
  )
}
