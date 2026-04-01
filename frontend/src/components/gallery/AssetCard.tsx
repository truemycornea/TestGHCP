'use client';

import Image from 'next/image';
import { useState } from 'react';
import type { Asset } from '@/lib/types/asset';

interface AssetCardProps {
  asset: Asset;
  isSelected: boolean;
  onSelect: (id: string) => void;
}

/**
 * A single media card in the gallery grid.
 *
 * - Renders a lazy-loaded WebP thumbnail (falls back to a placeholder).
 * - Shows a play-icon overlay for videos.
 * - Displays a tick overlay when selected.
 * - Renders an IntersectionObserver-friendly <Image> for lazy loading.
 */
export function AssetCard({ asset, isSelected, onSelect }: AssetCardProps) {
  const [loaded, setLoaded] = useState(false);

  const thumbUrl = asset.thumb_path
    ? `/api/v1/assets/${asset.id}/thumb`
    : '/placeholder.svg';

  return (
    <div
      role="checkbox"
      aria-checked={isSelected}
      tabIndex={0}
      onClick={() => onSelect(asset.id)}
      onKeyDown={(e) => e.key === 'Enter' && onSelect(asset.id)}
      className={`
        relative aspect-square cursor-pointer overflow-hidden rounded-md bg-muted
        ring-offset-background transition-all
        focus:outline-none focus:ring-2 focus:ring-primary focus:ring-offset-2
        ${isSelected ? 'ring-2 ring-primary' : 'hover:ring-1 hover:ring-border'}
      `}
    >
      {/* Skeleton while loading */}
      {!loaded && (
        <div className="absolute inset-0 animate-pulse bg-muted" />
      )}

      <Image
        src={thumbUrl}
        alt={asset.filename}
        fill
        sizes="(max-width: 640px) 120px, 200px"
        className={`object-cover transition-opacity duration-300 ${loaded ? 'opacity-100' : 'opacity-0'}`}
        onLoad={() => setLoaded(true)}
        loading="lazy"
      />

      {/* Video badge */}
      {asset.media_type === 'video' && (
        <div className="absolute bottom-1 right-1 rounded bg-black/60 px-1 py-0.5">
          <span className="text-[10px] text-white">▶</span>
        </div>
      )}

      {/* Selection tick */}
      {isSelected && (
        <div className="absolute inset-0 flex items-start justify-start p-1">
          <div className="flex h-5 w-5 items-center justify-center rounded-full bg-primary text-white text-xs">
            ✓
          </div>
        </div>
      )}

      {/* Favourite indicator */}
      {asset.is_favourite && (
        <div className="absolute bottom-1 left-1">
          <span className="text-xs">❤️</span>
        </div>
      )}
    </div>
  );
}
