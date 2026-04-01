'use client';

import { useState, useCallback } from 'react';
import useSWRInfinite from 'swr/infinite';
import { apiClient } from '@/lib/api/client';
import type { Asset, AssetListResponse } from '@/lib/types/asset';

interface UseAssetsOptions {
  query?: string;
  mediaType?: 'photo' | 'video';
  pageSize?: number;
}

async function fetchAssets(url: string): Promise<AssetListResponse> {
  return apiClient.get<AssetListResponse>(url);
}

/**
 * SWR-based hook for paginated, filtered asset fetching with infinite scroll.
 */
export function useAssets({
  query = '',
  mediaType,
  pageSize = 100,
}: UseAssetsOptions = {}) {
  const getKey = useCallback(
    (pageIndex: number, previousPageData: AssetListResponse | null) => {
      // Reached end of list.
      if (previousPageData && previousPageData.assets.length < pageSize) return null;

      const params = new URLSearchParams({
        limit: String(pageSize),
        offset: String(pageIndex * pageSize),
      });
      if (query) params.set('q', query);
      if (mediaType) params.set('media_type', mediaType);

      const base = query ? '/search' : '/assets';
      return `${base}?${params.toString()}`;
    },
    [query, mediaType, pageSize],
  );

  const { data, size, setSize, isLoading, isValidating } = useSWRInfinite<AssetListResponse>(
    getKey,
    fetchAssets,
    { revalidateFirstPage: false },
  );

  const assets: Asset[] = data ? data.flatMap((page) => page.assets ?? []) : [];
  const total = data?.[0]?.total ?? 0;

  const loadMore = useCallback(() => {
    if (!isValidating) setSize((s) => s + 1);
  }, [isValidating, setSize]);

  return { assets, total, loadMore, isLoading, isValidating };
}
