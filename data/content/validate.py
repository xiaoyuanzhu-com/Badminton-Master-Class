#!/usr/bin/env python3
"""Validate content-as-code JSON files against schemas and cross-references."""

import json
import os
import re
import sys
from pathlib import Path

# jsonschema is optional — fall back to manual checks if not installed
try:
    import jsonschema
    HAS_JSONSCHEMA = True
except ImportError:
    HAS_JSONSCHEMA = False

CONTENT_DIR = Path(__file__).parent
SCHEMAS_DIR = CONTENT_DIR / "schemas"
TECHNIQUES_DIR = CONTENT_DIR / "techniques"
PEOPLE_DIR = CONTENT_DIR / "people"


def load_json(path: Path) -> dict | None:
    """Load and parse a JSON file, returning None on error."""
    try:
        with open(path) as f:
            return json.load(f)
    except (json.JSONDecodeError, OSError) as e:
        return None


def load_schema(name: str) -> dict | None:
    """Load a schema file by name."""
    return load_json(SCHEMAS_DIR / name)


def validate_against_schema(data: dict, schema: dict, filepath: str) -> list[str]:
    """Validate data against a JSON Schema. Returns list of error messages."""
    if not HAS_JSONSCHEMA:
        return validate_manually(data, schema, filepath)
    errors = []
    v = jsonschema.Draft7Validator(schema)
    for err in v.iter_errors(data):
        errors.append(f"  {filepath}: {err.message}")
    return errors


def validate_manually(data: dict, schema: dict, filepath: str) -> list[str]:
    """Basic manual validation when jsonschema is not installed."""
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
        if "enum" in prop and val not in prop["enum"]:
            errors.append(f"  {filepath}: field '{key}' value '{val}' not in {prop['enum']}")
        if "pattern" in prop and isinstance(val, str):
            if not re.match(prop["pattern"], val):
                errors.append(f"  {filepath}: field '{key}' value '{val}' doesn't match pattern {prop['pattern']}")
        if "minLength" in prop and isinstance(val, str) and len(val) < prop["minLength"]:
            errors.append(f"  {filepath}: field '{key}' is too short (min {prop['minLength']})")

    return errors


def collect_people_slugs() -> set[str]:
    """Collect all people slugs from the people directory."""
    slugs = set()
    if not PEOPLE_DIR.is_dir():
        return slugs
    for f in PEOPLE_DIR.iterdir():
        if f.suffix == ".json":
            slugs.add(f.stem)
    return slugs


def main() -> int:
    errors: list[str] = []
    warnings: list[str] = []

    # Load schemas
    technique_schema = load_schema("technique.schema.json")
    person_schema = load_schema("person.schema.json")
    content_schema = load_schema("content.schema.json")

    if not technique_schema:
        errors.append("Cannot load technique.schema.json")
    if not person_schema:
        errors.append("Cannot load person.schema.json")
    if not content_schema:
        errors.append("Cannot load content.schema.json")

    if errors:
        for e in errors:
            print(f"FATAL: {e}", file=sys.stderr)
        return 1

    # Collect people slugs
    people_slugs = collect_people_slugs()
    people_count = 0

    # Validate people files
    if PEOPLE_DIR.is_dir():
        for f in sorted(PEOPLE_DIR.iterdir()):
            if f.suffix != ".json":
                continue
            people_count += 1
            data = load_json(f)
            if data is None:
                errors.append(f"  {f.relative_to(CONTENT_DIR)}: invalid JSON")
                continue
            errors.extend(validate_against_schema(data, person_schema, str(f.relative_to(CONTENT_DIR))))

    # Walk technique directories, validate techniques and content
    technique_count = 0
    content_count = 0
    source_urls: dict[str, str] = {}  # url -> filepath (for duplicate check)

    if not TECHNIQUES_DIR.is_dir():
        errors.append("techniques/ directory not found")
    else:
        for dirpath, dirnames, filenames in os.walk(TECHNIQUES_DIR):
            dirpath = Path(dirpath)
            rel_dir = dirpath.relative_to(CONTENT_DIR)

            # Check for _technique.json in every directory under techniques/
            # (skip the root techniques/ directory itself — it's just a container)
            if dirpath == TECHNIQUES_DIR:
                continue
            technique_file = dirpath / "_technique.json"
            if not technique_file.exists():
                errors.append(f"  {rel_dir}: missing _technique.json")
            else:
                technique_count += 1
                data = load_json(technique_file)
                if data is None:
                    errors.append(f"  {rel_dir}/_technique.json: invalid JSON")
                else:
                    errors.extend(validate_against_schema(
                        data, technique_schema, str(rel_dir / "_technique.json")))

            # Validate content JSON files (anything that's not _technique.json)
            for fname in sorted(filenames):
                if fname == "_technique.json" or not fname.endswith(".json"):
                    continue
                content_count += 1
                fpath = dirpath / fname
                rel_path = fpath.relative_to(CONTENT_DIR)
                data = load_json(fpath)
                if data is None:
                    errors.append(f"  {rel_path}: invalid JSON")
                    continue

                errors.extend(validate_against_schema(data, content_schema, str(rel_path)))

                # Check person reference
                person = data.get("person")
                if person and person not in people_slugs:
                    errors.append(f"  {rel_path}: person '{person}' not found in people/")

                # Check duplicate source_url
                url = data.get("source_url")
                if url:
                    if url in source_urls:
                        errors.append(
                            f"  {rel_path}: duplicate source_url '{url}' "
                            f"(also in {source_urls[url]})")
                    else:
                        source_urls[url] = str(rel_path)

    # Summary
    print(f"Validated: {technique_count} techniques, {people_count} people, {content_count} content items")

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
