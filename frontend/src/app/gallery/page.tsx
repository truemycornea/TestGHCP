/**
 * Gallery page — uses virtual scrolling via react-virtuoso so it can handle
 * 100k+ assets without DOM bloat.  Assets are loaded in pages via SWR.
 */
'use client';

import { useState, useCallback } from 'react';
import { Virtuoso } from 'react-virtuoso';
import { useAssets } from '@/lib/hooks/useAssets';
import { AssetCard } from '@/components/gallery/AssetCard';
import { SearchBar } from '@/components/gallery/SearchBar';
import { GalleryToolbar } from '@/components/gallery/GalleryToolbar';
import type { Asset } from '@/lib/types/asset';

const PAGE_SIZE = 100;

export default function GalleryPage() {
  const [query, setQuery] = useState('');
  const [filter, setFilter] = useState<'all' | 'photo' | 'video'>('all');
  const [selected, setSelected] = useState<Set<string>>(new Set());

  const { assets, total, loadMore, isLoading } = useAssets({
    query,
    mediaType: filter === 'all' ? undefined : filter,
    pageSize: PAGE_SIZE,
  });

  const toggleSelect = useCallback((id: string) => {
    setSelected((prev) => {
      const next = new Set(prev);
      if (next.has(id)) next.delete(id);
      else next.add(id);
      return next;
    });
  }, []);

  // Pack assets into rows of 4 for the virtual list.
  const COLS = 4;
  const rows: Asset[][] = [];
  for (let i = 0; i < assets.length; i += COLS) {
    rows.push(assets.slice(i, i + COLS));
  }

  return (
    <div className="flex h-screen flex-col bg-background">
      <GalleryToolbar
        total={total}
        selected={selected.size}
        filter={filter}
        onFilterChange={setFilter}
        onClearSelection={() => setSelected(new Set())}
      />

      <SearchBar value={query} onChange={setQuery} />

      {isLoading && assets.length === 0 ? (
        <div className="flex flex-1 items-center justify-center text-muted-foreground">
          Loading…
        </div>
      ) : assets.length === 0 ? (
        <div className="flex flex-1 items-center justify-center text-muted-foreground">
          No assets found.
        </div>
      ) : (
        <Virtuoso
          className="flex-1"
          totalCount={rows.length}
          endReached={loadMore}
          itemContent={(index) => (
            <div className="gallery-grid px-4 py-1">
              {rows[index].map((asset) => (
                <AssetCard
                  key={asset.id}
                  asset={asset}
                  isSelected={selected.has(asset.id)}
                  onSelect={toggleSelect}
                />
              ))}
            </div>
          )}
        />
      )}
    </div>
  );
}
