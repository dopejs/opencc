import { useState, useEffect } from "react";

const GITHUB_API = "https://api.github.com/repos/dopejs/opencc/releases/latest";
const CACHE_KEY = "opencc-latest-version";
const CACHE_TTL = 1000 * 60 * 30; // 30 minutes

interface CacheEntry {
  version: string;
  timestamp: number;
}

export function useVersion() {
  const [version, setVersion] = useState<string>("");

  useEffect(() => {
    const cached = localStorage.getItem(CACHE_KEY);
    if (cached) {
      try {
        const entry: CacheEntry = JSON.parse(cached);
        if (Date.now() - entry.timestamp < CACHE_TTL) {
          setVersion(entry.version);
          return;
        }
      } catch {
        // ignore parse errors
      }
    }

    fetch(GITHUB_API)
      .then((res) => res.json())
      .then((data) => {
        const tag = data.tag_name as string;
        if (tag) {
          const v = tag.startsWith("v") ? tag : `v${tag}`;
          setVersion(v);
          localStorage.setItem(
            CACHE_KEY,
            JSON.stringify({ version: v, timestamp: Date.now() })
          );
        }
      })
      .catch(() => {
        // silently fail
      });
  }, []);

  return version;
}
