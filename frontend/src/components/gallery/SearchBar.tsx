'use client';

import { useRef } from 'react';

interface SearchBarProps {
  value: string;
  onChange: (value: string) => void;
}

/**
 * Semantic search bar — queries the /search endpoint via natural language.
 * Results are powered by CLIP embeddings stored in pgvector.
 */
export function SearchBar({ value, onChange }: SearchBarProps) {
  const inputRef = useRef<HTMLInputElement>(null);

  return (
    <div className="sticky top-0 z-10 border-b border-border bg-background/80 backdrop-blur-sm px-4 py-2">
      <div className="relative mx-auto max-w-2xl">
        <span className="absolute left-3 top-1/2 -translate-y-1/2 text-muted-foreground" aria-hidden>
          🔍
        </span>
        <input
          ref={inputRef}
          type="search"
          placeholder='Search your library… e.g. "sunset over the mountains in 2022"'
          value={value}
          onChange={(e) => onChange(e.target.value)}
          className="w-full rounded-full border border-border bg-muted py-2 pl-9 pr-4 text-sm focus:outline-none focus:ring-2 focus:ring-primary"
        />
        {value && (
          <button
            onClick={() => {
              onChange('');
              inputRef.current?.focus();
            }}
            className="absolute right-3 top-1/2 -translate-y-1/2 text-xs text-muted-foreground hover:text-foreground"
            aria-label="Clear search"
          >
            ✕
          </button>
        )}
      </div>
    </div>
  );
}
