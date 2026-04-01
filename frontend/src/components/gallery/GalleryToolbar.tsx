'use client';

interface GalleryToolbarProps {
  total: number;
  selected: number;
  filter: 'all' | 'photo' | 'video';
  onFilterChange: (f: 'all' | 'photo' | 'video') => void;
  onClearSelection: () => void;
}

const FILTERS: { label: string; value: 'all' | 'photo' | 'video' }[] = [
  { label: 'All', value: 'all' },
  { label: 'Photos', value: 'photo' },
  { label: 'Videos', value: 'video' },
];

/**
 * Top toolbar with media-type filter tabs and selection state indicator.
 */
export function GalleryToolbar({
  total,
  selected,
  filter,
  onFilterChange,
  onClearSelection,
}: GalleryToolbarProps) {
  return (
    <header className="flex items-center justify-between border-b border-border px-4 py-3">
      <div className="flex items-center gap-3">
        <h2 className="text-base font-semibold">
          {selected > 0 ? `${selected} selected` : `${total.toLocaleString()} items`}
        </h2>
        {selected > 0 && (
          <button
            onClick={onClearSelection}
            className="text-xs text-muted-foreground hover:text-foreground"
          >
            Clear
          </button>
        )}
      </div>

      <nav className="flex gap-1">
        {FILTERS.map(({ label, value }) => (
          <button
            key={value}
            onClick={() => onFilterChange(value)}
            className={`
              rounded-md px-3 py-1.5 text-xs font-medium transition-colors
              ${filter === value
                ? 'bg-primary text-primary-foreground'
                : 'text-muted-foreground hover:bg-muted hover:text-foreground'}
            `}
          >
            {label}
          </button>
        ))}
      </nav>
    </header>
  );
}
