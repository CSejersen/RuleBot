import { useEffect, useRef, useState } from "react";

type CacheEntry<T> = {
  data?: T;
  error?: any;
  updatedAt?: number;
  listeners: Set<() => void>;
};

const globalCache = new Map<string, CacheEntry<any>>();

export type SwrLiteOptions<T> = {
  ttlMs?: number; // how long data is considered fresh
  revalidateOnFocus?: boolean;
  revalidateIntervalMs?: number;
  initialData?: T;
};

export function useSwrLite<T>(
  key: string | null,
  fetcher: () => Promise<T>,
  opts: SwrLiteOptions<T> = {}
) {
  const { ttlMs = 60_000, revalidateOnFocus = true, revalidateIntervalMs, initialData } = opts;
  const [data, setData] = useState<T | undefined>(() => {
    if (!key) return initialData;
    const entry = globalCache.get(key);
    return (entry?.data as T | undefined) ?? initialData;
  });
  const [error, setError] = useState<any>(undefined);
  const [loading, setLoading] = useState<boolean>(() => !data);
  const keyRef = useRef(key);

  // ensure cache entry exists
  useEffect(() => {
    if (!key) return;
    if (!globalCache.has(key)) {
      globalCache.set(key, { listeners: new Set() });
    }
  }, [key]);

  // subscribe to cache updates
  useEffect(() => {
    if (!key) return;
    const entry = globalCache.get(key)!;
    const notify = () => {
      const e = globalCache.get(key!);
      setData(e?.data);
      setError(e?.error);
      setLoading(!e?.data && !e?.error);
    };
    entry.listeners.add(notify);
    return () => {
      entry.listeners.delete(notify);
    };
  }, [key]);

  async function revalidate() {
    if (!key) return;
    const entry = globalCache.get(key)!;
    try {
      const result = await fetcher();
      entry.data = result;
      entry.error = undefined;
      entry.updatedAt = Date.now();
      entry.listeners.forEach((l) => l());
    } catch (err) {
      entry.error = err;
      entry.listeners.forEach((l) => l());
    }
  }

  // initial load / SWR logic
  useEffect(() => {
    keyRef.current = key;
    if (!key) return;
    const entry = globalCache.get(key)!;
    const isStale = !entry.updatedAt || Date.now() - (entry.updatedAt ?? 0) > ttlMs;
    // if we have cached data, show immediately, then revalidate if stale
    if (entry.data !== undefined) {
      setData(entry.data);
      setLoading(false);
      if (isStale) {
        revalidate();
      }
      return;
    }
    // no cached data -> fetch
    setLoading(true);
    revalidate();
  // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [key, ttlMs]);

  // focus revalidation
  useEffect(() => {
    if (!revalidateOnFocus || !key) return;
    const onFocus = () => revalidate();
    window.addEventListener("focus", onFocus);
    return () => window.removeEventListener("focus", onFocus);
  // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [key, revalidateOnFocus]);

  // interval revalidation
  useEffect(() => {
    if (!revalidateIntervalMs || !key) return;
    const id = setInterval(revalidate, revalidateIntervalMs);
    return () => clearInterval(id);
  // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [key, revalidateIntervalMs]);

  const mutate = (updater: (prev?: T) => T) => {
    if (!key) return;
    const entry = globalCache.get(key)!;
    entry.data = updater(entry.data);
    entry.updatedAt = Date.now();
    entry.listeners.forEach((l) => l());
  };

  return { data, error, loading, revalidate, mutate };
}
