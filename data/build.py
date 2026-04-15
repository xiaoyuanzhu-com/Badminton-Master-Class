#!/usr/bin/env python3
"""
Compiler: walks data/content/ and produces bmc.db.

Usage:
    python3 data/build.py              # compile only
    python3 data/build.py --upload     # compile + upload to OSS

Thumbnail strategy (BUILD-2 → THUMB pipeline):
    Content JSON files store a thumbnail_url pointing to the platform-hosted
    thumbnail (e.g. Bilibili cover image, YouTube hqdefault.jpg). These URLs
    are stable, public, and don't need our own CDN. build.py reads this field
    and writes it into the thumbnail_url column in the compiled DB.
"""

import json
import os
import shutil
import sqlite3
import subprocess
import sys
from pathlib import Path

ROOT = Path(__file__).resolve().parent.parent  # project root
CONTENT_DIR = ROOT / "data" / "content"
TECHNIQUES_DIR = CONTENT_DIR / "techniques"
PEOPLE_DIR = CONTENT_DIR / "people"
PATHS_DIR = CONTENT_DIR / "paths"
DB_PATH = ROOT / "data" / "bmc.db"

IOS_BUNDLE = ROOT / "ios" / "BadmintonMasterClass" / "Resources" / "bmc.db"
ANDROID_BUNDLE = ROOT / "android" / "app" / "src" / "main" / "assets" / "bmc.db"

SCHEMA_VERSION = 3

SCHEMA_SQL = """\
CREATE TABLE categories (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    icon TEXT NOT NULL DEFAULT '',
    sort_order INTEGER NOT NULL DEFAULT 0,
    parent_id INTEGER REFERENCES categories(id)
);

CREATE TABLE people (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    slug TEXT NOT NULL UNIQUE,
    name TEXT NOT NULL,
    bio TEXT NOT NULL DEFAULT '',
    platforms_json TEXT NOT NULL DEFAULT '{}'
);

CREATE TABLE contents (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    title TEXT NOT NULL,
    summary TEXT NOT NULL DEFAULT '',
    thumbnail_url TEXT NOT NULL DEFAULT '',
    source_url TEXT NOT NULL,
    source_platform TEXT NOT NULL CHECK(source_platform IN ('bilibili', 'xiaohongshu', 'douyin', 'wechat', 'youtube', 'other')),
    author_name TEXT NOT NULL DEFAULT '',
    person_id INTEGER REFERENCES people(id),
    difficulty TEXT NOT NULL DEFAULT '',
    duration TEXT NOT NULL DEFAULT '',
    editor_notes TEXT NOT NULL DEFAULT '',
    category_id INTEGER NOT NULL REFERENCES categories(id),
    sort_order INTEGER NOT NULL DEFAULT 0,
    created_at TEXT NOT NULL DEFAULT (datetime('now')),
    updated_at TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE INDEX idx_contents_category ON contents(category_id);
CREATE INDEX idx_contents_person ON contents(person_id);

CREATE TABLE learning_paths (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    title TEXT NOT NULL,
    summary TEXT NOT NULL DEFAULT '',
    difficulty TEXT NOT NULL DEFAULT '',
    sort_order INTEGER NOT NULL DEFAULT 0
);

CREATE TABLE path_steps (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    path_id INTEGER NOT NULL REFERENCES learning_paths(id),
    step_order INTEGER NOT NULL DEFAULT 0,
    day INTEGER,
    title TEXT NOT NULL,
    note TEXT NOT NULL DEFAULT ''
);

CREATE TABLE path_step_contents (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    step_id INTEGER NOT NULL REFERENCES path_steps(id),
    content_id INTEGER NOT NULL REFERENCES contents(id),
    sort_order INTEGER NOT NULL DEFAULT 0
);

CREATE INDEX idx_path_steps_path ON path_steps(path_id);
CREATE INDEX idx_path_step_contents_step ON path_step_contents(step_id);

CREATE TABLE schema_version (
    version INTEGER NOT NULL
);
"""


def load_json(path: Path) -> dict:
    with open(path) as f:
        return json.load(f)


def build_people(cur: sqlite3.Cursor) -> dict[str, tuple[int, str]]:
    """Insert people rows. Returns {slug: (id, name)}."""
    people_map: dict[str, tuple[int, str]] = {}
    if not PEOPLE_DIR.is_dir():
        return people_map

    for f in sorted(PEOPLE_DIR.iterdir()):
        if f.suffix != ".json":
            continue
        slug = f.stem
        data = load_json(f)
        name = data["name"]
        bio = data.get("bio", "")
        platforms = json.dumps(data.get("platforms", {}), ensure_ascii=False)

        cur.execute(
            "INSERT INTO people (slug, name, bio, platforms_json) VALUES (?, ?, ?, ?)",
            (slug, name, bio, platforms),
        )
        people_map[slug] = (cur.lastrowid, name)

    return people_map


def build_categories(cur: sqlite3.Cursor) -> dict[str, int]:
    """Walk technique folders and insert categories. Returns {rel_path: id}."""
    cat_map: dict[str, int] = {}

    # Walk breadth-first by sorting directory entries at each level
    def walk_dir(dirpath: Path, parent_id: int | None, sort_start: int):
        entries = sorted(
            [e for e in dirpath.iterdir() if e.is_dir()],
            key=lambda e: e.name,
        )
        for sort_order, entry in enumerate(entries, start=sort_start):
            technique_file = entry / "_technique.json"
            if not technique_file.exists():
                print(f"WARNING: {entry} has no _technique.json, skipping", file=sys.stderr)
                continue

            data = load_json(technique_file)
            rel_path = str(entry.relative_to(TECHNIQUES_DIR))

            explicit_order = data.get("sort_order", sort_order)

            cur.execute(
                "INSERT INTO categories (name, icon, sort_order, parent_id) VALUES (?, ?, ?, ?)",
                (data["name"], data.get("icon", ""), explicit_order, parent_id),
            )
            cat_id = cur.lastrowid
            cat_map[rel_path] = cat_id

            # Recurse into subdirectories
            walk_dir(entry, cat_id, 0)

    walk_dir(TECHNIQUES_DIR, None, 0)
    return cat_map


def build_contents(
    cur: sqlite3.Cursor,
    cat_map: dict[str, int],
    people_map: dict[str, tuple[int, str]],
):
    """Walk technique folders and insert content items."""
    content_count = 0

    for dirpath, dirnames, filenames in os.walk(TECHNIQUES_DIR):
        dirpath = Path(dirpath)
        if dirpath == TECHNIQUES_DIR:
            continue

        rel_dir = str(dirpath.relative_to(TECHNIQUES_DIR))
        cat_id = cat_map.get(rel_dir)
        if cat_id is None:
            continue

        content_files = sorted(
            f for f in filenames if f.endswith(".json") and f != "_technique.json"
        )

        for sort_order, fname in enumerate(content_files):
            fpath = dirpath / fname
            data = load_json(fpath)
            slug = Path(fname).stem

            # Use thumbnail_url from content JSON (platform-hosted URL)
            thumbnail_url = data.get("thumbnail_url", "")

            # Resolve person
            person_slug = data.get("person", "")
            person_id = None
            author_name = ""
            if person_slug and person_slug in people_map:
                person_id, author_name = people_map[person_slug]

            cur.execute(
                """INSERT INTO contents
                   (title, summary, thumbnail_url, source_url, source_platform,
                    author_name, person_id, difficulty, duration, editor_notes,
                    category_id, sort_order)
                   VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)""",
                (
                    data["title"],
                    data.get("summary", ""),
                    thumbnail_url,
                    data["source_url"],
                    data["source_platform"],
                    author_name,
                    person_id,
                    data.get("difficulty", ""),
                    data.get("duration", ""),
                    data.get("editor_notes", ""),
                    cat_id,
                    sort_order,
                ),
            )
            content_count += 1

    return content_count


def build_content_slug_map(cur: sqlite3.Cursor) -> dict[str, int]:
    """Build a mapping of content file slug -> contents.id from the DB."""
    slug_map: dict[str, int] = {}
    for dirpath, _dirnames, filenames in os.walk(TECHNIQUES_DIR):
        dirpath = Path(dirpath)
        for fname in sorted(filenames):
            if fname.endswith(".json") and fname != "_technique.json":
                slug = Path(fname).stem
                # Look up the content by title (slug is the filename stem)
                # We need to match by source_url or title; safer to query by
                # iterating what we just inserted. Instead, collect during insert.
                pass
    # Query all contents and map by scanning technique files for slug->title
    # Simpler approach: scan technique dirs and match slug to DB row by source_url
    # Actually, the simplest: re-scan files and look up by source_url.
    slug_map = {}
    for dirpath, _dirnames, filenames in os.walk(TECHNIQUES_DIR):
        dirpath_p = Path(dirpath)
        for fname in sorted(filenames):
            if fname.endswith(".json") and fname != "_technique.json":
                slug = Path(fname).stem
                fpath = dirpath_p / fname
                data = load_json(fpath)
                source_url = data.get("source_url", "")
                row = cur.execute(
                    "SELECT id FROM contents WHERE source_url = ?", (source_url,)
                ).fetchone()
                if row:
                    slug_map[slug] = row[0]
    return slug_map


def build_paths(cur: sqlite3.Cursor, content_slug_map: dict[str, int]) -> int:
    """Read path JSON files and insert into learning_paths, path_steps, path_step_contents."""
    path_count = 0
    if not PATHS_DIR.is_dir():
        return path_count

    for sort_order, f in enumerate(sorted(PATHS_DIR.iterdir())):
        if f.suffix != ".json":
            continue
        data = load_json(f)
        path_count += 1

        cur.execute(
            "INSERT INTO learning_paths (title, summary, difficulty, sort_order) VALUES (?, ?, ?, ?)",
            (data["title"], data.get("summary", ""), data.get("difficulty", ""), sort_order),
        )
        path_id = cur.lastrowid

        for step_order, step in enumerate(data.get("steps", [])):
            cur.execute(
                "INSERT INTO path_steps (path_id, step_order, day, title, note) VALUES (?, ?, ?, ?, ?)",
                (path_id, step_order, step.get("day"), step["title"], step.get("note", "")),
            )
            step_id = cur.lastrowid

            for content_sort, slug in enumerate(step.get("content_slugs", [])):
                content_id = content_slug_map.get(slug)
                if content_id is not None:
                    cur.execute(
                        "INSERT INTO path_step_contents (step_id, content_id, sort_order) VALUES (?, ?, ?)",
                        (step_id, content_id, content_sort),
                    )

    return path_count


def compile_db():
    """Main compilation: create bmc.db from content files."""
    # Remove existing DB
    if DB_PATH.exists():
        DB_PATH.unlink()

    conn = sqlite3.connect(str(DB_PATH))
    cur = conn.cursor()

    # Enable foreign keys
    cur.execute("PRAGMA foreign_keys = ON")

    # Create schema
    cur.executescript(SCHEMA_SQL)

    # Record schema version
    cur.execute("INSERT INTO schema_version (version) VALUES (?)", (SCHEMA_VERSION,))

    # Build data
    people_map = build_people(cur)
    cat_map = build_categories(cur)
    content_count = build_contents(cur, cat_map, people_map)
    content_slug_map = build_content_slug_map(cur)
    path_count = build_paths(cur, content_slug_map)

    conn.commit()

    # Print summary
    print(f"Compiled bmc.db:")
    print(f"  Categories: {len(cat_map)}")
    print(f"  People:     {len(people_map)}")
    print(f"  Contents:   {content_count}")
    print(f"  Paths:      {path_count}")
    print(f"  Schema:     v{SCHEMA_VERSION}")
    print(f"  Output:     {DB_PATH}")

    conn.close()
    return len(cat_map), len(people_map), content_count, path_count


def copy_to_bundles():
    """Copy bmc.db to iOS and Android app bundles."""
    for dest in [IOS_BUNDLE, ANDROID_BUNDLE]:
        dest.parent.mkdir(parents=True, exist_ok=True)
        shutil.copy2(str(DB_PATH), str(dest))
        print(f"  Copied to {dest.relative_to(ROOT)}")


def upload_to_oss():
    """Upload bmc.db to OSS using env vars (same as admin panel)."""
    endpoint = os.environ.get("BMC_OSS_ENDPOINT", "")
    bucket = os.environ.get("BMC_OSS_BUCKET", "")
    key_id = os.environ.get("BMC_OSS_ACCESS_KEY_ID", "")
    key_secret = os.environ.get("BMC_OSS_ACCESS_KEY_SECRET", "")

    if not all([endpoint, bucket, key_id, key_secret]):
        print("OSS env vars not set, skipping upload.")
        print("  Required: BMC_OSS_ENDPOINT, BMC_OSS_BUCKET, BMC_OSS_ACCESS_KEY_ID, BMC_OSS_ACCESS_KEY_SECRET")
        return False

    object_key = os.environ.get("BMC_OSS_OBJECT_KEY", "bmc.db")

    # Use ossutil or the admin Go binary for upload — for now, use a simple
    # Python approach via subprocess calling the admin export endpoint,
    # or we can shell out to a Go helper. For simplicity, we use the
    # aliyun-oss Python SDK if available, otherwise skip.
    try:
        import oss2
        auth = oss2.Auth(key_id, key_secret)
        bucket_obj = oss2.Bucket(auth, endpoint, bucket)
        bucket_obj.put_object_from_file(object_key, str(DB_PATH))
        print(f"  Uploaded to OSS: {bucket}/{object_key}")
        return True
    except ImportError:
        # Fall back to ossutil CLI
        try:
            subprocess.run(
                [
                    "ossutil", "cp", str(DB_PATH),
                    f"oss://{bucket}/{object_key}",
                    "-e", endpoint,
                    "-i", key_id,
                    "-k", key_secret,
                    "-f",
                ],
                check=True,
            )
            print(f"  Uploaded to OSS: {bucket}/{object_key}")
            return True
        except FileNotFoundError:
            print("  ERROR: Neither oss2 Python package nor ossutil CLI found.", file=sys.stderr)
            print("  Install: pip install oss2  OR  brew install ossutil", file=sys.stderr)
            return False


def main():
    upload = "--upload" in sys.argv

    # Step 1: Validate
    print("=== Validating content files ===")
    validate_script = CONTENT_DIR / "validate.py"
    result = subprocess.run([sys.executable, str(validate_script)], cwd=str(ROOT))
    if result.returncode != 0:
        print("Validation failed. Fix errors before compiling.", file=sys.stderr)
        return 1

    # Step 2: Compile
    print("\n=== Compiling bmc.db ===")
    cat_count, people_count, content_count, path_count = compile_db()

    # Step 3: Copy to bundles
    print("\n=== Copying to app bundles ===")
    copy_to_bundles()

    # Step 4: Upload (optional)
    if upload:
        print("\n=== Uploading to OSS ===")
        upload_to_oss()

    print("\nDone.")
    return 0


if __name__ == "__main__":
    sys.exit(main())
