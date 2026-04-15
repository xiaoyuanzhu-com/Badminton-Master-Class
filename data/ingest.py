#!/usr/bin/env python3
"""Content ingestion agent: fetch metadata from a URL and write content files.

Usage:
    python data/ingest.py <url> [category_path] [person_slug]

Examples:
    python data/ingest.py "https://www.bilibili.com/video/BV1xx..."
    python data/ingest.py "https://www.bilibili.com/video/BV1xx..." "techniques/basics/clear"
    python data/ingest.py "https://www.bilibili.com/video/BV1xx..." "techniques/basics/clear" "yang-chen-da-shen"
"""

import hashlib
import html.parser
import json
import os
import re
import subprocess
import sys
import urllib.error
import urllib.request
from pathlib import Path

# ---------------------------------------------------------------------------
# Constants
# ---------------------------------------------------------------------------

SCRIPT_DIR = Path(__file__).resolve().parent
CONTENT_DIR = SCRIPT_DIR / "content"
TECHNIQUES_DIR = CONTENT_DIR / "techniques"
PEOPLE_DIR = CONTENT_DIR / "people"

PLATFORM_DOMAINS = {
    "bilibili.com": "bilibili",
    "b23.tv": "bilibili",
    "youtube.com": "youtube",
    "youtu.be": "youtube",
    "xiaohongshu.com": "xiaohongshu",
    "xhslink.com": "xiaohongshu",
    "douyin.com": "douyin",
    "mp.weixin.qq.com": "wechat",
}

USER_AGENT = (
    "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) "
    "AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"
)

# ---------------------------------------------------------------------------
# Pinyin slug generation (optional dependency)
# ---------------------------------------------------------------------------

try:
    from pypinyin import lazy_pinyin

    def _chinese_to_pinyin(text: str) -> str:
        return "".join(lazy_pinyin(text))

    HAS_PINYIN = True
except ImportError:
    HAS_PINYIN = False

    def _chinese_to_pinyin(text: str) -> str:  # type: ignore[misc]
        return text


# ---------------------------------------------------------------------------
# HTML metadata extractor
# ---------------------------------------------------------------------------


class MetaExtractor(html.parser.HTMLParser):
    """Extract <meta> og:/name tags and <title> from HTML."""

    def __init__(self) -> None:
        super().__init__()
        self.meta: dict[str, str] = {}
        self._in_title = False
        self._title_text = ""

    def handle_starttag(self, tag: str, attrs: list[tuple[str, str | None]]) -> None:
        if tag == "title":
            self._in_title = True
            self._title_text = ""
            return
        if tag != "meta":
            return
        attr_dict: dict[str, str] = {}
        for k, v in attrs:
            if v is not None:
                attr_dict[k.lower()] = v

        # og:xxx or twitter:xxx
        prop = attr_dict.get("property", attr_dict.get("name", ""))
        content = attr_dict.get("content", "")
        if prop and content:
            self.meta[prop.lower()] = content

    def handle_data(self, data: str) -> None:
        if self._in_title:
            self._title_text += data

    def handle_endtag(self, tag: str) -> None:
        if tag == "title" and self._in_title:
            self._in_title = False
            if self._title_text.strip():
                self.meta["html_title"] = self._title_text.strip()


# ---------------------------------------------------------------------------
# Platform detection
# ---------------------------------------------------------------------------


def detect_platform(url: str) -> str | None:
    """Return platform slug from the URL domain, or None."""
    # strip scheme
    host = url.split("://", 1)[-1].split("/")[0].split(":")[0].lower()
    # try exact then suffix match
    for domain, platform in PLATFORM_DOMAINS.items():
        if host == domain or host.endswith("." + domain):
            return platform
    return None


# ---------------------------------------------------------------------------
# Fetch page HTML
# ---------------------------------------------------------------------------


def fetch_html(url: str) -> str:
    """Fetch the URL and return the HTML body as a string."""
    req = urllib.request.Request(url, headers={"User-Agent": USER_AGENT})
    with urllib.request.urlopen(req, timeout=20) as resp:
        charset = resp.headers.get_content_charset() or "utf-8"
        return resp.read().decode(charset, errors="replace")


# ---------------------------------------------------------------------------
# Extract metadata
# ---------------------------------------------------------------------------


def extract_metadata(html_text: str, url: str, platform: str) -> dict:
    """Parse HTML and return a metadata dict with title, author, thumbnail, duration."""
    parser = MetaExtractor()
    parser.feed(html_text)
    m = parser.meta

    title = (
        m.get("og:title")
        or m.get("twitter:title")
        or m.get("html_title")
        or ""
    )
    author = m.get("og:video:author") or m.get("author") or ""
    thumbnail = m.get("og:image") or m.get("twitter:image") or ""
    duration_raw = m.get("og:video:duration") or ""

    # Platform-specific fallbacks ------------------------------------------

    if platform == "bilibili":
        # Bilibili often puts author in specific meta tags
        if not author:
            match = re.search(r'"name"\s*:\s*"([^"]+)"', html_text)
            if match:
                author = match.group(1)
        # Duration in __INITIAL_STATE__ JSON
        if not duration_raw:
            match = re.search(r'"duration"\s*:\s*(\d+)', html_text)
            if match:
                duration_raw = match.group(1)
        # Title cleanup: remove " - bilibili" suffix
        title = re.sub(r"\s*[-_]\s*(哔哩哔哩|bilibili).*$", "", title, flags=re.IGNORECASE)

    elif platform == "youtube":
        if not author:
            match = re.search(r'"ownerChannelName"\s*:\s*"([^"]+)"', html_text)
            if match:
                author = match.group(1)
        if not duration_raw:
            # ISO 8601 PT##M##S
            match = re.search(r'"duration"\s*:\s*"PT(?:(\d+)H)?(?:(\d+)M)?(?:(\d+)S)?"', html_text)
            if match:
                h = int(match.group(1) or 0)
                mins = int(match.group(2) or 0)
                s = int(match.group(3) or 0)
                duration_raw = str(h * 3600 + mins * 60 + s)

    elif platform == "xiaohongshu":
        if not author:
            match = re.search(r'"nickname"\s*:\s*"([^"]+)"', html_text)
            if match:
                author = match.group(1)

    elif platform == "douyin":
        if not author:
            match = re.search(r'"nickname"\s*:\s*"([^"]+)"', html_text)
            if match:
                author = match.group(1)
        if not duration_raw:
            match = re.search(r'"duration"\s*:\s*(\d+)', html_text)
            if match:
                # Douyin duration is in milliseconds
                duration_raw = str(int(match.group(1)) // 1000)

    # Format duration as M:SS or MM:SS
    duration = ""
    if duration_raw and duration_raw.isdigit():
        total_secs = int(duration_raw)
        mins = total_secs // 60
        secs = total_secs % 60
        duration = f"{mins}:{secs:02d}"

    return {
        "title": title.strip(),
        "author": author.strip(),
        "thumbnail": thumbnail.strip(),
        "duration": duration,
    }


# ---------------------------------------------------------------------------
# Slug generation
# ---------------------------------------------------------------------------


def generate_slug(title: str) -> str:
    """Convert a (possibly Chinese) title to a URL-friendly slug."""
    if not title:
        return "untitled"

    text = title

    # Convert Chinese characters to pinyin if possible
    if HAS_PINYIN and re.search(r"[\u4e00-\u9fff]", text):
        text = _chinese_to_pinyin(text)
    elif re.search(r"[\u4e00-\u9fff]", text):
        # Fallback: hash the Chinese portion for uniqueness, keep ASCII parts
        ascii_parts = re.sub(r"[^\x00-\x7f]", " ", text).strip()
        chinese_hash = hashlib.md5(text.encode()).hexdigest()[:8]
        if ascii_parts:
            text = ascii_parts + "-" + chinese_hash
        else:
            text = chinese_hash

    # Normalize to lowercase ASCII with hyphens
    text = text.lower()
    text = re.sub(r"[^a-z0-9]+", "-", text)
    text = text.strip("-")

    # Truncate if very long
    if len(text) > 60:
        text = text[:60].rstrip("-")

    return text or "untitled"


def generate_person_slug(name: str) -> str:
    """Generate a slug for a person name."""
    return generate_slug(name)


# ---------------------------------------------------------------------------
# Category listing
# ---------------------------------------------------------------------------


def list_categories() -> list[str]:
    """Return all available technique category paths relative to content/."""
    categories: list[str] = []
    if not TECHNIQUES_DIR.is_dir():
        return categories
    for dirpath, dirnames, filenames in os.walk(TECHNIQUES_DIR):
        dp = Path(dirpath)
        if (dp / "_technique.json").exists():
            rel = dp.relative_to(CONTENT_DIR)
            categories.append(str(rel))
    return sorted(categories)


# ---------------------------------------------------------------------------
# File writers
# ---------------------------------------------------------------------------


def ensure_person(slug: str, name: str, platform: str, url: str) -> bool:
    """Create a person stub if the slug doesn't already exist. Returns True if created."""
    person_file = PEOPLE_DIR / f"{slug}.json"
    if person_file.exists():
        return False

    PEOPLE_DIR.mkdir(parents=True, exist_ok=True)

    # Build a platform profile URL from the source URL domain
    platform_url = url.split("?")[0]  # crude: use the video URL for now
    person_data = {
        "name": name,
        "platforms": {platform: platform_url} if platform_url else {},
    }
    with open(person_file, "w", encoding="utf-8") as f:
        json.dump(person_data, f, ensure_ascii=False, indent=2)
        f.write("\n")
    return True


def write_content(
    category_path: str,
    slug: str,
    title: str,
    source_url: str,
    platform: str,
    person_slug: str,
    duration: str,
    summary: str = "",
    thumbnail_url: str = "",
) -> Path:
    """Write the content JSON file and return the path."""
    folder = CONTENT_DIR / category_path
    if not folder.is_dir():
        raise FileNotFoundError(f"Category folder does not exist: {folder}")

    content_data: dict[str, str] = {
        "title": title,
        "source_url": source_url,
        "source_platform": platform,
        "person": person_slug,
    }
    if summary:
        content_data["summary"] = summary
    if duration:
        content_data["duration"] = duration
    if thumbnail_url:
        content_data["thumbnail_url"] = thumbnail_url

    filepath = folder / f"{slug}.json"
    with open(filepath, "w", encoding="utf-8") as f:
        json.dump(content_data, f, ensure_ascii=False, indent=2)
        f.write("\n")
    return filepath


# ---------------------------------------------------------------------------
# Thumbnail download
# ---------------------------------------------------------------------------


def download_thumbnail(thumbnail_url: str, dest_path: Path) -> bool:
    """Download thumbnail image. Returns True on success."""
    if not thumbnail_url:
        return False
    try:
        req = urllib.request.Request(thumbnail_url, headers={"User-Agent": USER_AGENT})
        with urllib.request.urlopen(req, timeout=15) as resp:
            content_type = resp.headers.get("Content-Type", "")
            # Determine extension from content type
            if "png" in content_type:
                ext = ".png"
            elif "webp" in content_type:
                ext = ".webp"
            else:
                ext = ".jpg"
            out = dest_path.with_suffix(ext)
            with open(out, "wb") as f:
                f.write(resp.read())
        print(f"  Thumbnail saved: {out.relative_to(CONTENT_DIR)}")
        return True
    except (urllib.error.URLError, OSError) as e:
        print(f"  Thumbnail download failed ({e}), skipping.")
        return False


# ---------------------------------------------------------------------------
# Validator
# ---------------------------------------------------------------------------


def run_validator() -> int:
    """Run validate.py and return the exit code."""
    validate_script = CONTENT_DIR / "validate.py"
    if not validate_script.exists():
        print("  Validator not found, skipping.")
        return 0
    result = subprocess.run(
        [sys.executable, str(validate_script)],
        cwd=str(CONTENT_DIR.parent),
        capture_output=True,
        text=True,
    )
    print(result.stdout)
    if result.stderr:
        print(result.stderr, file=sys.stderr)
    return result.returncode


# ---------------------------------------------------------------------------
# Main
# ---------------------------------------------------------------------------


def main() -> int:
    if len(sys.argv) < 2:
        print(__doc__)
        return 1

    url = sys.argv[1]
    category_path = sys.argv[2] if len(sys.argv) > 2 else None
    person_slug_arg = sys.argv[3] if len(sys.argv) > 3 else None

    # Step 1: Detect platform
    platform = detect_platform(url)
    if not platform:
        print(f"ERROR: Cannot detect platform from URL: {url}")
        print("Supported: bilibili, youtube, xiaohongshu, douyin, wechat")
        return 1
    print(f"Platform: {platform}")

    # Step 2: Fetch and extract metadata
    print(f"Fetching: {url}")
    try:
        html_text = fetch_html(url)
    except (urllib.error.URLError, OSError) as e:
        print(f"ERROR: Failed to fetch URL: {e}")
        return 1

    meta = extract_metadata(html_text, url, platform)
    print(f"  Title:     {meta['title'] or '(not found)'}")
    print(f"  Author:    {meta['author'] or '(not found)'}")
    print(f"  Duration:  {meta['duration'] or '(not found)'}")
    print(f"  Thumbnail: {'yes' if meta['thumbnail'] else 'no'}")

    if not meta["title"]:
        print("ERROR: Could not extract title from the page.")
        return 1

    # Step 3: Category
    if not category_path:
        categories = list_categories()
        if not categories:
            print("ERROR: No technique categories found.")
            return 1
        print("\nAvailable categories:")
        for i, cat in enumerate(categories, 1):
            print(f"  {i}. {cat}")
        print()
        try:
            choice = input("Pick a category number (or type the path): ").strip()
            if choice.isdigit():
                idx = int(choice) - 1
                if 0 <= idx < len(categories):
                    category_path = categories[idx]
                else:
                    print("Invalid selection.")
                    return 1
            else:
                category_path = choice
        except (EOFError, KeyboardInterrupt):
            print("\nAborted.")
            return 1

    # Validate category exists
    cat_dir = CONTENT_DIR / category_path
    if not cat_dir.is_dir():
        print(f"ERROR: Category folder does not exist: {category_path}")
        return 1

    # Step 4: Person
    person_slug = person_slug_arg
    if not person_slug and meta["author"]:
        person_slug = generate_person_slug(meta["author"])
    if not person_slug:
        print("ERROR: Could not determine author. Provide person_slug as third argument.")
        return 1

    person_created = ensure_person(person_slug, meta["author"] or person_slug, platform, url)
    if person_created:
        print(f"  Created new person: people/{person_slug}.json")
    else:
        print(f"  Person exists: people/{person_slug}.json")

    # Step 5: Generate slug
    slug = generate_slug(meta["title"])
    print(f"  Content slug: {slug}")

    # Check for slug collision
    content_file = cat_dir / f"{slug}.json"
    if content_file.exists():
        print(f"WARNING: File already exists: {content_file.relative_to(CONTENT_DIR)}")
        print("  Appending hash to avoid collision.")
        url_hash = hashlib.md5(url.encode()).hexdigest()[:6]
        slug = f"{slug}-{url_hash}"
        print(f"  New slug: {slug}")

    # Step 6: Resolve thumbnail URL
    # For YouTube, use the deterministic thumbnail URL pattern
    thumbnail_url = meta.get("thumbnail", "")
    if platform == "youtube" and not thumbnail_url:
        yt_match = re.search(r"(?:v=|youtu\.be/)([A-Za-z0-9_-]{11})", url)
        if yt_match:
            thumbnail_url = f"https://img.youtube.com/vi/{yt_match.group(1)}/hqdefault.jpg"

    # Step 7: Write content JSON
    content_path = write_content(
        category_path=category_path,
        slug=slug,
        title=meta["title"],
        source_url=url,
        platform=platform,
        person_slug=person_slug,
        duration=meta["duration"],
        thumbnail_url=thumbnail_url,
    )
    print(f"  Created: {content_path.relative_to(CONTENT_DIR)}")

    # Step 8: Download thumbnail (local copy, optional)
    thumb_dest = content_path.with_suffix("")  # strip .json, download_thumbnail adds ext
    download_thumbnail(meta["thumbnail"], thumb_dest)

    # Step 9: Validate
    print("\nRunning validator...")
    val_rc = run_validator()

    # Step 10: Summary
    print("\n--- Summary ---")
    print(f"  Platform:  {platform}")
    print(f"  Title:     {meta['title']}")
    print(f"  Author:    {meta['author']}")
    print(f"  Person:    people/{person_slug}.json {'(new)' if person_created else '(existing)'}")
    print(f"  Content:   {content_path.relative_to(CONTENT_DIR)}")
    print(f"  Category:  {category_path}")
    if meta["duration"]:
        print(f"  Duration:  {meta['duration']}")
    print(f"  Validator: {'PASS' if val_rc == 0 else 'FAIL'}")

    return val_rc


if __name__ == "__main__":
    sys.exit(main())
