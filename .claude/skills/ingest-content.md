# Ingest Content

Ingest a video/article URL into the badminton content library.

## Triggers
- "ingest this video: URL"
- "add this content: URL"
- "ingest URL"
- "/ingest URL"

## Workflow

1. **Parse the user's message** to extract the URL and optional category/person hints.

2. **Run the ingest script** from the project root:
   ```bash
   python data/ingest.py "<url>" [category_path] [person_slug]
   ```
   - If no category is specified, run without it first to list categories, then ask the user to pick one and re-run with the chosen category.
   - If the user mentions a person by name, look up the slug in `data/content/people/` and pass it as the third argument.

3. **Show the user what was created:**
   - Print the new/modified files with `git diff` or `git status`.
   - Highlight: content JSON path, person file (if new), thumbnail (if downloaded).

4. **Ask the user to review the diff** before committing.

5. **On approval**, commit the new files:
   ```bash
   git add data/content/
   git commit -m "content: ingest <title> from <platform>"
   ```

## Supported Platforms
- Bilibili (bilibili.com, b23.tv)
- YouTube (youtube.com, youtu.be)
- Xiaohongshu (xiaohongshu.com, xhslink.com)
- Douyin (douyin.com)
- WeChat Articles (mp.weixin.qq.com)

## Category Paths
Categories are technique folders under `data/content/techniques/`. Examples:
- `techniques/basics/grip`
- `techniques/basics/clear`
- `techniques/attack/smash`
- `techniques/defense/smash-return`
- `techniques/doubles/rotation`

Run `python data/ingest.py` with no arguments to see full list.

## Notes
- The script uses only Python standard library (no pip install needed).
- If `pypinyin` is installed, Chinese titles produce cleaner slugs; otherwise a hash fallback is used.
- Thumbnails are optional; if download fails the content file is still valid.
- The validator (`data/content/validate.py`) runs automatically after ingestion.
