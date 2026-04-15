#!/usr/bin/env python3
"""Backfill thumbnail_url into existing content JSON files.

For each content file that lacks a thumbnail_url, derives the URL from the
source platform and source_url:

  - Bilibili:    https://i0.hdslb.com/bfs/archive/{bvid_hash}.jpg  — but since
                 we can't derive the hash offline, we use the Bilibili API:
                 https://api.bilibili.com/x/web-interface/view?bvid={BVID}
                 which returns .data.pic (the cover image URL).
  - YouTube:     https://img.youtube.com/vi/{VIDEO_ID}/hqdefault.jpg
  - Douyin:      Skipped (no stable public thumbnail URL pattern).
  - Xiaohongshu: Skipped (no stable public thumbnail URL pattern).

Usage:
    python data/backfill_thumbnails.py           # dry-run (print what would change)
    python data/backfill_thumbnails.py --write   # actually modify files
"""

import json
import re
import sys
import time
import urllib.error
import urllib.request
from pathlib import Path

SCRIPT_DIR = Path(__file__).resolve().parent
TECHNIQUES_DIR = SCRIPT_DIR / "content" / "techniques"

USER_AGENT = (
    "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) "
    "AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"
)


def youtube_thumbnail(source_url: str) -> str:
    """Derive YouTube thumbnail URL from the video URL."""
    m = re.search(r"(?:v=|youtu\.be/)([A-Za-z0-9_-]{11})", source_url)
    if m:
        return f"https://img.youtube.com/vi/{m.group(1)}/hqdefault.jpg"
    return ""


def bilibili_thumbnail(source_url: str) -> str:
    """Fetch Bilibili cover image via the public API."""
    m = re.search(r"(BV[A-Za-z0-9]+)", source_url)
    if not m:
        return ""
    bvid = m.group(1)
    api_url = f"https://api.bilibili.com/x/web-interface/view?bvid={bvid}"
    try:
        req = urllib.request.Request(api_url, headers={"User-Agent": USER_AGENT})
        with urllib.request.urlopen(req, timeout=10) as resp:
            data = json.loads(resp.read().decode("utf-8"))
            pic = data.get("data", {}).get("pic", "")
            if pic:
                # Ensure https
                if pic.startswith("//"):
                    pic = "https:" + pic
                elif pic.startswith("http://"):
                    pic = "https://" + pic[7:]
                return pic
    except (urllib.error.URLError, OSError, json.JSONDecodeError) as e:
        print(f"  WARNING: Bilibili API failed for {bvid}: {e}")
    return ""


def resolve_thumbnail(source_url: str, platform: str) -> str:
    """Return a thumbnail URL for the given content, or empty string."""
    if platform == "youtube":
        return youtube_thumbnail(source_url)
    if platform == "bilibili":
        return bilibili_thumbnail(source_url)
    # douyin, xiaohongshu, wechat — no reliable public thumbnail URL
    return ""


def main() -> int:
    write_mode = "--write" in sys.argv

    if not TECHNIQUES_DIR.is_dir():
        print(f"ERROR: techniques dir not found: {TECHNIQUES_DIR}")
        return 1

    updated = 0
    skipped = 0
    already_has = 0
    no_thumbnail = 0

    for fpath in sorted(TECHNIQUES_DIR.rglob("*.json")):
        if fpath.name == "_technique.json":
            continue

        with open(fpath) as f:
            data = json.load(f)

        if data.get("thumbnail_url"):
            already_has += 1
            continue

        platform = data.get("source_platform", "")
        source_url = data.get("source_url", "")
        if not source_url or not platform:
            skipped += 1
            continue

        thumb = resolve_thumbnail(source_url, platform)
        if not thumb:
            no_thumbnail += 1
            rel = fpath.relative_to(TECHNIQUES_DIR)
            print(f"  SKIP {rel} ({platform}) — no thumbnail available")
            continue

        rel = fpath.relative_to(TECHNIQUES_DIR)
        print(f"  {'WRITE' if write_mode else 'WOULD'} {rel} → {thumb}")

        if write_mode:
            data["thumbnail_url"] = thumb
            with open(fpath, "w", encoding="utf-8") as f:
                json.dump(data, f, ensure_ascii=False, indent=2)
                f.write("\n")

            # Rate-limit Bilibili API calls
            if platform == "bilibili":
                time.sleep(0.3)

        updated += 1

    print(f"\nSummary: {updated} {'updated' if write_mode else 'would update'}, "
          f"{already_has} already had thumbnail, {no_thumbnail} no thumbnail available, "
          f"{skipped} skipped")
    return 0


if __name__ == "__main__":
    sys.exit(main())
