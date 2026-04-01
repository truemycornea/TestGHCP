/** TypeScript types that mirror the Go backend domain entities. */

export type MediaType = 'photo' | 'video';

export interface EXIFData {
  make?: string;
  model?: string;
  lens_model?: string;
  f_number?: number;
  exposure_time?: string;
  iso?: number;
  focal_length?: number;
  width?: number;
  height?: number;
  orientation?: number;
  date_time_original?: string;
  gps_latitude?: number;
  gps_longitude?: number;
  gps_altitude?: number;
  city?: string;
  country?: string;
}

export interface Asset {
  id: string;
  owner_id: string;
  filename: string;
  original_path: string;
  thumb_path?: string;
  proxy_path?: string;
  mime_type: string;
  media_type: MediaType;
  file_size: number;
  duration?: number;
  checksum: string;
  phash_value?: string;
  exif: EXIFData;
  is_favourite: boolean;
  is_archived: boolean;
  is_trashed: boolean;
  is_processed: boolean;
  created_at: string;
  updated_at: string;
  taken_at: string;
}

export interface AssetListResponse {
  assets: Asset[];
  total: number;
  limit: number;
  offset: number;
}

export interface User {
  id: string;
  email: string;
  display_name: string;
  avatar_url?: string;
  role: 'admin' | 'user' | 'viewer' | 'contributor';
  is_active: boolean;
  created_at: string;
  updated_at: string;
  last_login_at?: string;
}

export interface TokenPair {
  access_token: string;
  refresh_token: string;
  expires_at: string;
}

export type SharePermission = 'view_only' | 'contributor';

export interface SharedLink {
  id: string;
  album_id: string;
  token: string;
  permission: SharePermission;
  expires_at?: string;
  created_at: string;
}

export interface Album {
  id: string;
  owner_id: string;
  name: string;
  description?: string;
  cover_asset_id?: string;
  asset_count: number;
  shared_links?: SharedLink[];
  created_at: string;
  updated_at: string;
}

export interface SemanticSearchResponse {
  results: Asset[];
  count: number;
}
