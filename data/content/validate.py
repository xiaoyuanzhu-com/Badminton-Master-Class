#!/usr/bin/env python3
"""Validate content-as-code JSON files against schemas and cross-references.

Layout:
  techniques/<path>/<slug>/<slug>.json   — technique node (marker = matches parent dir)
                            other.json   — sub-technique leaves under the same parent
  techniques/<path>/<slug>/<slug>.png    — poster image (sibling, matching basename)
  content/<slug>.json (+ .png)           — social-media post reference + scraped preview
  people/<slug>.json (+ .png)            — author + avatar
  paths/<slug>.json                      — curated learning path over content slugs
"""

import json
import os
import re
import sys
from pathlib import Path

try:
    import jsonschema
    HAS_JSONSCHEMA = True
except ImportError:
    HAS_JSONSCHEMA = False

CONTENT_DIR = Path(__file__).parent
SCHEMAS_DIR = CONTENT_DIR / "schemas"
TECHNIQUES_DIR = CONTENT_DIR / "techniques"
PEOPLE_DIR = CONTENT_DIR / "people"
POSTS_DIR = CONTENT_DIR / "content"
PATHS_DIR = CONTENT_DIR / "paths"


def load_json(path: Path) -> dict | None:
    try:
        with open(path) as f:
            return json.load(f)
    except (json.JSONDecodeError, OSError):
        return None


def load_schema(name: str) -> dict | None:
    return load_json(SCHEMAS_DIR / name)


def validate_against_schema(data: dict, schema: dict, filepath: str) -> list[str]:
    if not HAS_JSONSCHEMA:
        return validate_manually(data, schema, filepath)
    errors = []
    v = jsonschema.Draft7Validator(schema)
    for err in v.iter_errors(data):
        errors.append(f"  {filepath}: {err.message}")
    return errors


def validate_manually(data: dict, schema: dict, filepath: str) -> list[str]:
    errors = []
    required = schema.get("required", [])
    properties = schema.get("properties", {})

    for field in required:
        if field not in data:
            errors.append(f"  {filepath}: missing required field '{field}'")

    if schema.get("additionalProperties") is False:
        for key in data:
            if key not in properties:
                errors.append(f"  {filepath}: unexpected field '{key}'")

    for key, prop in properties.items():
        if key not in data:
            continue
        val = data[key]
        expected_type = prop.get("type")
        if expected_type == "string" and not isinstance(val, str):
            errors.append(f"  {filepath}: field '{key}' should be a string")
        elif expected_type == "object" and not isinstance(val, dict):
            errors.append(f"  {filepath}: field '{key}' should be an object")
        elif expected_type == "array" and not isinstance(val, list):
            errors.append(f"  {filepath}: field '{key}' should be an array")
        if "enum" in prop and val not in prop["enum"]:
            errors.append(f"  {filepath}: field '{key}' value '{val}' not in {prop['enum']}")
        if "pattern" in prop and isinstance(val, str):
            if not re.match(prop["pattern"], val):
                errors.append(f"  {filepath}: field '{key}' value '{val}' doesn't match pattern {prop['pattern']}")
        if "minLength" in prop and isinstance(val, str) and len(val) < prop["minLength"]:
            errors.append(f"  {filepath}: field '{key}' is too short (min {prop['minLength']})")

    return errors


def collect_people_slugs() -> set[str]:
    slugs = set()
    if not PEOPLE_DIR.is_dir():
        return slugs
    for f in PEOPLE_DIR.iterdir():
        if f.suffix == ".json":
            slugs.add(f.stem)
    return slugs


def collect_technique_slugs() -> tuple[set[str], dict[str, str]]:
    """Walk techniques/. A marker file is <dir>/<dir>.json.
    Returns (set of leaf slugs, map of leaf -> full path like 'rearcourt/smash').
    """
    slugs: set[str] = set()
    paths: dict[str, str] = {}
    if not TECHNIQUES_DIR.is_dir():
        return slugs, paths
    for dirpath, _dirnames, filenames in os.walk(TECHNIQUES_DIR):
        d = Path(dirpath)
        if d == TECHNIQUES_DIR:
            continue
        marker = f"{d.name}.json"
        if marker in filenames:
            slugs.add(d.name)
            paths[d.name] = str(d.relative_to(TECHNIQUES_DIR))
    return slugs, paths


def main() -> int:
    errors: list[str] = []
    warnings: list[str] = []

    technique_schema = load_schema("technique.schema.json")
    person_schema = load_schema("person.schema.json")
    content_schema = load_schema("content.schema.json")
    path_schema = load_schema("path.schema.json")

    for name, s in [("technique", technique_schema), ("person", person_schema),
                    ("content", content_schema), ("path", path_schema)]:
        if not s:
            errors.append(f"Cannot load {name}.schema.json")
    if errors:
        for e in errors:
            print(f"FATAL: {e}", file=sys.stderr)
        return 1

    people_slugs = collect_people_slugs()
    people_count = 0

    # People
    if PEOPLE_DIR.is_dir():
        for f in sorted(PEOPLE_DIR.iterdir()):
            if f.suffix != ".json":
                continue
            people_count += 1
            data = load_json(f)
            if data is None:
                errors.append(f"  {f.relative_to(CONTENT_DIR)}: invalid JSON")
                continue
            errors.extend(validate_against_schema(
                data, person_schema, str(f.relative_to(CONTENT_DIR))))

    # Techniques: every directory under techniques/ must contain a marker file <dirname>.json
    technique_count = 0
    if not TECHNIQUES_DIR.is_dir():
        errors.append("techniques/ directory not found")
    else:
        for dirpath, _dirnames, filenames in os.walk(TECHNIQUES_DIR):
            d = Path(dirpath)
            rel = d.relative_to(CONTENT_DIR)
            if d == TECHNIQUES_DIR:
                continue
            marker_name = f"{d.name}.json"
            marker = d / marker_name
            if not marker.exists():
                errors.append(f"  {rel}: missing marker file {marker_name}")
                continue
            technique_count += 1
            data = load_json(marker)
            if data is None:
                errors.append(f"  {rel / marker_name}: invalid JSON")
                continue
            errors.extend(validate_against_schema(
                data, technique_schema, str(rel / marker_name)))

    technique_slugs, _ = collect_technique_slugs()

    # Content posts
    content_count = 0
    content_slugs: set[str] = set()
    source_urls: dict[str, str] = {}

    if POSTS_DIR.is_dir():
        for f in sorted(POSTS_DIR.iterdir()):
            if f.suffix != ".json":
                continue
            content_count += 1
            content_slugs.add(f.stem)
            rel_path = f.relative_to(CONTENT_DIR)
            data = load_json(f)
            if data is None:
                errors.append(f"  {rel_path}: invalid JSON")
                continue

            errors.extend(validate_against_schema(data, content_schema, str(rel_path)))

            person = data.get("person")
            if person and person not in people_slugs:
                errors.append(f"  {rel_path}: person '{person}' not found in people/")

            for tech in data.get("techniques", []):
                if tech not in technique_slugs:
                    errors.append(
                        f"  {rel_path}: technique '{tech}' not found in techniques/")

            url = data.get("source_url")
            if url:
                if url in source_urls:
                    errors.append(
                        f"  {rel_path}: duplicate source_url '{url}' "
                        f"(also in {source_urls[url]})")
                else:
                    source_urls[url] = str(rel_path)

    # Paths
    path_count = 0
    if PATHS_DIR.is_dir():
        for f in sorted(PATHS_DIR.iterdir()):
            if f.suffix != ".json":
                continue
            path_count += 1
            rel_path = f.relative_to(CONTENT_DIR)
            data = load_json(f)
            if data is None:
                errors.append(f"  {rel_path}: invalid JSON")
                continue

            errors.extend(validate_against_schema(data, path_schema, str(rel_path)))

            seen_slugs: set[str] = set()
            for i, step in enumerate(data.get("steps", [])):
                for slug in step.get("content_slugs", []):
                    if slug not in content_slugs:
                        errors.append(
                            f"  {rel_path}: step {i} references unknown content slug '{slug}'")
                    if slug in seen_slugs:
                        errors.append(
                            f"  {rel_path}: step {i} has duplicate content slug '{slug}'")
                    seen_slugs.add(slug)

    print(f"Validated: {technique_count} techniques, {people_count} people, "
          f"{content_count} content posts, {path_count} paths")

    if not HAS_JSONSCHEMA:
        warnings.append("jsonschema not installed — using basic manual validation")

    if warnings:
        for w in warnings:
            print(f"WARNING: {w}")

    if errors:
        print(f"\n{len(errors)} error(s) found:")
        for e in errors:
            print(e)
        return 1

    print("All checks passed.")
    return 0


if __name__ == "__main__":
    sys.exit(main())
