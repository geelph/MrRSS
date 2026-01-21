#!/usr/bin/env python3
"""
i18n Usage Analyzer for MrRSS

This script analyzes i18n key usage across the codebase and generates a markdown table
with line numbers and clickable links to source files.
"""

import os
import re
from pathlib import Path
from typing import Dict, List, Tuple, Set
from urllib.parse import quote

# File paths (relative to project root)
EN_FILE = Path("frontend/src/i18n/locales/en.ts")
ZH_FILE = Path("frontend/src/i18n/locales/zh.ts")
TYPES_FILE = Path("frontend/src/i18n/types.ts")
FRONTEND_DIR = Path("frontend/src")


def extract_keys_from_locale(file_path: Path) -> Tuple[Dict[str, int], Set[str]]:
    """
    Extract i18n keys and their line numbers from locale files.
    Returns tuple of (keys_dict, category_keys).
    - keys_dict: maps key name to line number
    - category_keys: set of keys that are categories (objects with children, no direct value)
    """
    keys = {}
    categories = set()

    if not file_path.exists():
        print(f"Warning: {file_path} not found")
        return keys, categories

    with open(file_path, "r", encoding="utf-8") as f:
        lines = f.readlines()

    # Track current nesting path
    path_stack = []

    for line_num, line in enumerate(lines, start=1):
        # Ignore commented lines
        stripped = line.strip()
        if stripped.startswith("//"):
            continue

        # Calculate current indentation level
        indent_match = re.match(r'^(\s*)', line)
        if not indent_match:
            continue
        indent_level = len(indent_match.group(1))

        # Update path stack based on indentation
        while path_stack and path_stack[-1][1] >= indent_level:
            path_stack.pop()

        # Match object key pattern:  key: { or  key: 'value'
        obj_match = re.match(r'^\s*([a-zA-Z_][a-zA-Z0-9_]*)\s*:\s*\{', line)
        value_match = re.match(r'^\s*([a-zA-Z_][a-zA-Z0-9_]*)\s*:\s*[\'"]', line)

        if obj_match:
            # This is a nested object (category)
            key = obj_match.group(1)
            path_stack.append((key, indent_level))

            # Build full key path
            full_key = ".".join([p[0] for p in path_stack])
            if full_key not in keys:  # Only keep first occurrence
                keys[full_key] = line_num
            # Mark as category (will be excluded from report)
            categories.add(full_key)

        elif value_match:
            # This is a leaf value (actual translation)
            key = value_match.group(1)
            full_key = ".".join([p[0] for p in path_stack] + [key]) if path_stack else key

            if full_key not in keys:  # Only keep first occurrence
                keys[full_key] = line_num

    return keys, categories


def extract_keys_from_types(file_path: Path) -> Dict[str, int]:
    """Extract i18n keys and their line numbers from types.ts"""
    keys = {}

    if not file_path.exists():
        print(f"Warning: {file_path} not found")
        return keys

    with open(file_path, "r", encoding="utf-8") as f:
        lines = f.readlines()

    # Track current nesting path
    path_stack = []

    for line_num, line in enumerate(lines, start=1):
        stripped = line.strip()

        # Calculate current indentation level
        indent_match = re.match(r'^(\s*)', line)
        if not indent_match:
            continue
        indent_level = len(indent_match.group(1))

        # Update path stack based on indentation
        while path_stack and path_stack[-1][1] >= indent_level:
            path_stack.pop()

        # Match nested type pattern:  key: { or  key: string;
        obj_match = re.match(r'^([a-zA-Z_][a-zA-Z0-9_]*)\s*:\s*\{', stripped)
        value_match = re.match(r'^([a-zA-Z_][a-zA-Z0-9_]*)\s*:\s*string;', stripped)

        if obj_match:
            # This is a nested object
            key = obj_match.group(1)
            path_stack.append((key, indent_level))

            # Build full key path
            full_key = ".".join([p[0] for p in path_stack])
            if full_key not in keys:  # Only keep first occurrence
                keys[full_key] = line_num

        elif value_match:
            # This is a leaf value
            key = value_match.group(1)
            full_key = ".".join([p[0] for p in path_stack] + [key]) if path_stack else key

            if full_key not in keys:  # Only keep first occurrence
                keys[full_key] = line_num

    return keys


def find_i18n_usage(key: str, search_dir: Path) -> List[Tuple[str, int]]:
    """
    Find all usages of an i18n key in the codebase.
    Returns list of (file_path, line_number) tuples.
    """
    usages = []
    pattern = re.compile(rf"t\(['\"]{re.escape(key)}['\"]\)|t\(\s*['\"]{re.escape(key)}['\"]\s*\)")

    # Walk through all TypeScript, Vue, and JavaScript files
    for ext in ["*.vue", "*.ts", "*.js"]:
        for file_path in search_dir.rglob(ext):
            # Skip the i18n locale files themselves
            if "i18n/locales" in str(file_path):
                continue

            try:
                with open(file_path, "r", encoding="utf-8") as f:
                    lines = f.readlines()

                for line_num, line in enumerate(lines, start=1):
                    if pattern.search(line):
                        # Convert to relative path from search_dir
                        try:
                            rel_path = file_path.relative_to(search_dir)
                            usages.append((str(rel_path), line_num))
                        except ValueError:
                            # On Windows with different drives or path issues
                            usages.append((str(file_path), line_num))
            except Exception as e:
                print(f"Warning: Could not read {file_path}: {e}")

    return usages


def create_file_link(file_path: str, line_num: int, is_windows: bool = False) -> str:
    """
    Create a clickable markdown link for a file location.
    The report is in tools/ folder, so we need relative path from tools/ to frontend/src/
    """
    # Convert backslashes to forward slashes for URLs on Windows
    if is_windows:
        file_path = file_path.replace("\\", "/")

    # The report is in tools/ folder, files are in frontend/src/
    # So we need to go up one level (../) then into frontend/src/
    relative_path = f"../frontend/src/{file_path}"
    display_text = f"{file_path}:{line_num}"

    # Create a clickable markdown link
    # Using relative path from tools/ to the file
    return f"[{display_text}]({relative_path}#L{line_num})"


def create_usage_links(usages: List[Tuple[str, int]], is_windows: bool = False) -> str:
    """Create comma-separated clickable links for usages."""
    if not usages:
        return ""

    links = []
    for file_path, line_num in sorted(usages)[:10]:  # Limit to 10 usages to avoid huge output
        # Shorten path by removing common prefix
        short_path = file_path.replace("frontend/src/", "")
        links.append(create_file_link(short_path, line_num, is_windows))

    result = ", ".join(links)
    if len(usages) > 10:
        result += f" ... ({len(usages) - 10} more)"

    return result


def get_nesting_depth(key: str) -> int:
    """Get the nesting depth of a key (number of dots + 1)."""
    return key.count(".") + 1


def is_top_level_key(key: str) -> bool:
    """Check if a key is a top-level key (not in any category)."""
    # Top-level keys have no dots
    return "." not in key


def main():
    # Determine if we're on Windows
    is_windows = os.name == "nt"

    # Change to project root directory (tools/ parent)
    script_dir = Path(__file__).parent
    project_root = script_dir.parent
    os.chdir(project_root)

    print(f"Analyzing i18n usage in {project_root}")
    print("=" * 80)

    # Extract keys from locale files with category information
    en_keys, en_categories = extract_keys_from_locale(EN_FILE)
    zh_keys, zh_categories = extract_keys_from_locale(ZH_FILE)
    types_keys = extract_keys_from_types(TYPES_FILE)

    # Combine all categories from both locale files
    all_categories = en_categories | zh_categories

    # Get all unique keys, excluding all categories
    all_keys = sorted(set(en_keys.keys()) | set(zh_keys.keys()) | set(types_keys.keys()))
    filtered_keys = [key for key in all_keys if key not in all_categories]

    # Group keys by nesting depth
    keys_by_depth: Dict[int, List[str]] = {}
    for key in filtered_keys:
        depth = get_nesting_depth(key)
        if depth not in keys_by_depth:
            keys_by_depth[depth] = []
        keys_by_depth[depth].append(key)

    print(f"Found {len(all_keys)} unique i18n keys")
    print(f"  - en.ts: {len(en_keys)} keys")
    print(f"  - zh.ts: {len(zh_keys)} keys")
    print(f"  - types.ts: {len(types_keys)} keys")
    print(f"  - Category keys (excluded from report): {len(all_categories)}")
    print(f"  - Leaf keys with values (included in report): {len(filtered_keys)}")
    print(f"\nNesting depth distribution:")
    for depth in sorted(keys_by_depth.keys()):
        print(f"  - Level {depth}: {len(keys_by_depth[depth])} keys")
    print("=" * 80)
    print()

    # Generate markdown table
    output_lines = []
    output_lines.append("# i18n Usage Analysis Report")
    output_lines.append("")
    output_lines.append("| Key | Depth | en.ts | zh.ts | types.ts | Usage Count | Locations |")
    output_lines.append("|-----|-------|-------|-------|----------|-------------|-----------|")

    for key in filtered_keys:
        en_line = en_keys.get(key, "")
        zh_line = zh_keys.get(key, "")
        types_line = types_keys.get(key, "")
        depth = get_nesting_depth(key)

        # Bold format for top-level keys (no dots)
        display_key = f"**{key}**" if is_top_level_key(key) else key

        # Find usages in the codebase
        usages = find_i18n_usage(key, FRONTEND_DIR)
        usage_count = len(usages)

        # Create clickable links
        usage_links = create_usage_links(usages, is_windows)

        output_lines.append(
            f"| {display_key} | {depth} | {en_line} | {zh_line} | {types_line} | {usage_count} | {usage_links} |"
        )

    # Write to output file
    output_file = script_dir / "i18n_usage_report.md"
    with open(output_file, "w", encoding="utf-8") as f:
        f.write("\n".join(output_lines))

    print(f"Report generated: {output_file}")
    print(f"Total keys analyzed: {len(filtered_keys)}")


if __name__ == "__main__":
    main()
