/**
 * Thin type-safe API client for communicating with the Aura Go backend.
 * Uses the browser's native Fetch API so it works in both App Router server
 * components (via Next.js fetch cache) and client components.
 */

const BASE_URL =
  process.env.NEXT_PUBLIC_API_URL ?? 'http://localhost:8080/api/v1';

function getToken(): string | null {
  if (typeof window === 'undefined') return null;
  return localStorage.getItem('aura_token');
}

async function request<T>(
  method: string,
  path: string,
  body?: unknown,
  options?: RequestInit,
): Promise<T> {
  const token = getToken();
  const headers: HeadersInit = {
    'Content-Type': 'application/json',
    ...(token ? { Authorization: `Bearer ${token}` } : {}),
    ...options?.headers,
  };

  const res = await fetch(`${BASE_URL}${path}`, {
    method,
    headers,
    body: body !== undefined ? JSON.stringify(body) : undefined,
    ...options,
  });

  if (!res.ok) {
    const err = await res.json().catch(() => ({ message: res.statusText }));
    throw new Error(err.message ?? 'API error');
  }

  // 204 No Content
  if (res.status === 204) return undefined as unknown as T;

  return res.json() as Promise<T>;
}

/** Upload a file using multipart/form-data (no JSON body). */
async function upload<T>(path: string, formData: FormData): Promise<T> {
  const token = getToken();
  const res = await fetch(`${BASE_URL}${path}`, {
    method: 'POST',
    headers: token ? { Authorization: `Bearer ${token}` } : {},
    body: formData,
  });

  if (!res.ok) {
    const err = await res.json().catch(() => ({ message: res.statusText }));
    throw new Error(err.message ?? 'Upload error');
  }

  return res.json() as Promise<T>;
}

export const apiClient = {
  get: <T>(path: string, options?: RequestInit) =>
    request<T>('GET', path, undefined, options),
  post: <T>(path: string, body: unknown, options?: RequestInit) =>
    request<T>('POST', path, body, options),
  patch: <T>(path: string, body: unknown, options?: RequestInit) =>
    request<T>('PATCH', path, body, options),
  delete: <T>(path: string, options?: RequestInit) =>
    request<T>('DELETE', path, undefined, options),
  upload,
};
