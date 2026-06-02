#!/usr/bin/env python3

from __future__ import annotations

import argparse
import json
import re
import sys
import zipfile
from dataclasses import asdict, dataclass
from pathlib import Path


APP_ID = "343050"
CUSTOMIZE_LUA_PATH = "scripts/map/customize.lua"


@dataclass
class ItemField:
    key: str
    default: str | None
    worlds: list[str]
    description_ref: str | None
    options: list[str] | None
    master_option: bool
    master_controlled: bool
    master_sync: bool


@dataclass
class FieldGroup:
    key: str
    title_ref: str | None
    description_ref: str | None
    items: list[ItemField]


@dataclass
class ExtractionReport:
    managed_root: str
    build_id: str | None
    scripts_zip: str
    language_files: list[str]
    worldgen_groups: list[FieldGroup]
    worldsettings_groups: list[FieldGroup]


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(
        description="Extract DST configuration field inventories from a managed DST install."
    )
    parser.add_argument(
        "--managed-root",
        default=str(Path.home() / ".local" / "share" / "dst-server-ctl"),
        help="Managed root that contains dst/, clusters/, and state/.",
    )
    parser.add_argument(
        "--format",
        choices=("markdown", "json"),
        default="markdown",
        help="Output format.",
    )
    parser.add_argument(
        "--output",
        help="Optional output path. Defaults to stdout.",
    )
    return parser.parse_args()


def main() -> int:
    args = parse_args()
    managed_root = Path(args.managed_root).expanduser().resolve()
    dst_root = managed_root / "dst"
    scripts_zip = dst_root / "data" / "databundles" / "scripts.zip"
    appmanifest = dst_root / "steamapps" / f"appmanifest_{APP_ID}.acf"

    if not scripts_zip.is_file():
        print(f"scripts.zip not found: {scripts_zip}", file=sys.stderr)
        return 1

    customize_source = read_zip_text(scripts_zip, CUSTOMIZE_LUA_PATH)
    description_options = parse_description_tables(customize_source)
    worldgen_groups = parse_group_table(
        customize_source, "WORLDGEN_GROUP", description_options
    )
    worldsettings_groups = parse_group_table(
        customize_source, "WORLDSETTINGS_GROUP", description_options
    )

    report = ExtractionReport(
        managed_root=str(managed_root),
        build_id=read_build_id(appmanifest),
        scripts_zip=str(scripts_zip),
        language_files=list_language_files(scripts_zip),
        worldgen_groups=worldgen_groups,
        worldsettings_groups=worldsettings_groups,
    )

    if args.format == "json":
        rendered = json.dumps(asdict(report), indent=2, ensure_ascii=True)
    else:
        rendered = render_markdown(report)

    if args.output:
        output_path = Path(args.output).expanduser()
        output_path.write_text(rendered + "\n", encoding="utf-8")
    else:
        sys.stdout.write(rendered + "\n")

    return 0


def read_zip_text(zip_path: Path, member: str) -> str:
    with zipfile.ZipFile(zip_path) as archive:
        with archive.open(member) as handle:
            return handle.read().decode("utf-8")


def read_build_id(appmanifest: Path) -> str | None:
    if not appmanifest.is_file():
        return None

    content = appmanifest.read_text(encoding="utf-8")
    match = re.search(r'"buildid"\s+"([^"]+)"', content)
    return match.group(1) if match else None


def list_language_files(zip_path: Path) -> list[str]:
    with zipfile.ZipFile(zip_path) as archive:
        return sorted(
            name.removeprefix("scripts/languages/").removesuffix(".po")
            for name in archive.namelist()
            if name.startswith("scripts/languages/") and name.endswith(".po")
        )


def parse_description_tables(source: str) -> dict[str, list[str]]:
    result: dict[str, list[str]] = {}
    for match in re.finditer(r"local\s+([A-Za-z0-9_]+)\s*=\s*\{", source):
        name = match.group(1)
        body, _ = extract_braced_block(source, match.end() - 1)
        data_values = re.findall(r'data\s*=\s*"([^"]+)"', body)
        if data_values:
            result[name] = data_values
    return result


def parse_group_table(
    source: str, table_name: str, description_options: dict[str, list[str]]
) -> list[FieldGroup]:
    token = f"local {table_name} ="
    start = source.find(token)
    if start < 0:
        raise ValueError(f"table not found: {table_name}")

    brace_start = source.find("{", start)
    table_body, _ = extract_braced_block(source, brace_start)
    groups: list[FieldGroup] = []

    for key, group_body in iter_named_tables(table_body):
        items_body = extract_named_table(group_body, "items")
        description_ref = parse_string_or_identifier(group_body, "desc")
        title_ref = parse_string_or_identifier(group_body, "text")
        items = [
            parse_item(item_key, item_body, description_options)
            for item_key, item_body in iter_named_tables(items_body)
        ]
        groups.append(
            FieldGroup(
                key=key,
                title_ref=title_ref,
                description_ref=description_ref,
                items=items,
            )
        )

    return groups


def parse_item(
    key: str, body: str, description_options: dict[str, list[str]]
) -> ItemField:
    description_ref = parse_string_or_identifier(body, "desc")
    return ItemField(
        key=key,
        default=parse_quoted_value(body, "value"),
        worlds=parse_string_list(body, "world"),
        description_ref=description_ref,
        options=description_options.get(description_ref) if description_ref else None,
        master_option=parse_boolean_flag(body, "masteroption"),
        master_controlled=parse_boolean_flag(body, "master_controlled"),
        master_sync=parse_boolean_flag(body, "master_sync"),
    )


def parse_quoted_value(body: str, field_name: str) -> str | None:
    match = re.search(rf"{re.escape(field_name)}\s*=\s*\"([^\"]+)\"", body)
    return match.group(1) if match else None


def parse_string_or_identifier(body: str, field_name: str) -> str | None:
    match = re.search(
        rf"{re.escape(field_name)}\s*=\s*(\"([^\"]+)\"|([A-Za-z0-9_\.]+))", body
    )
    if not match:
        return None
    return match.group(2) or match.group(3)


def parse_string_list(body: str, field_name: str) -> list[str]:
    match = re.search(rf"{re.escape(field_name)}\s*=\s*\{{", body)
    if not match:
        return []
    table_body, _ = extract_braced_block(body, match.end() - 1)
    return re.findall(r'"([^"]+)"', table_body)


def parse_boolean_flag(body: str, field_name: str) -> bool:
    return re.search(rf"{re.escape(field_name)}\s*=\s*true", body) is not None


def extract_named_table(body: str, field_name: str) -> str:
    match = re.search(rf"{re.escape(field_name)}\s*=\s*\{{", body)
    if not match:
        raise ValueError(f"named table not found: {field_name}")
    table_body, _ = extract_braced_block(body, match.end() - 1)
    return table_body


def iter_named_tables(body: str) -> list[tuple[str, str]]:
    results: list[tuple[str, str]] = []
    idx = 0
    while True:
        match = re.search(r'\["([^"]+)"\]\s*=\s*\{', body[idx:])
        if not match:
            return results
        key = match.group(1)
        brace_index = idx + match.end() - 1
        block, next_index = extract_braced_block(body, brace_index)
        results.append((key, block))
        idx = next_index


def extract_braced_block(source: str, brace_index: int) -> tuple[str, int]:
    if source[brace_index] != "{":
        raise ValueError("brace_index must point at '{'")

    depth = 0
    in_string = False
    escaped = False
    for idx in range(brace_index, len(source)):
        char = source[idx]
        if in_string:
            if escaped:
                escaped = False
            elif char == "\\":
                escaped = True
            elif char == '"':
                in_string = False
            continue

        if char == '"':
            in_string = True
            continue
        if char == "{":
            depth += 1
            continue
        if char == "}":
            depth -= 1
            if depth == 0:
                return source[brace_index + 1 : idx], idx + 1

    raise ValueError("unterminated brace block")


def render_markdown(report: ExtractionReport) -> str:
    lines: list[str] = []
    lines.append("# DST Config Field Extraction")
    lines.append("")
    lines.append(f"- Managed root: `{report.managed_root}`")
    lines.append(f"- scripts.zip: `{report.scripts_zip}`")
    lines.append(f"- Build ID: `{report.build_id or 'unknown'}`")
    lines.append("")
    lines.append("## Languages")
    lines.append("")
    for language in report.language_files:
        lines.append(f"- `{language}`")
    lines.append("")
    lines.extend(render_group_section("WORLDGEN_GROUP", report.worldgen_groups))
    lines.append("")
    lines.extend(render_group_section("WORLDSETTINGS_GROUP", report.worldsettings_groups))
    return "\n".join(lines)


def render_group_section(title: str, groups: list[FieldGroup]) -> list[str]:
    lines = [f"## {title}", ""]
    for group in groups:
        lines.append(f"### {group.key}")
        lines.append("")
        lines.append(f"- `text`: `{group.title_ref or '-'} `".rstrip())
        lines.append(f"- `desc`: `{group.description_ref or '-'} `".rstrip())
        lines.append(f"- `items`: `{len(group.items)}`")
        lines.append("")
        lines.append("| key | default | worlds | desc | options | flags |")
        lines.append("| --- | --- | --- | --- | --- | --- |")
        for item in group.items:
            worlds = ", ".join(item.worlds) if item.worlds else "-"
            options = ", ".join(item.options) if item.options else "-"
            flags = []
            if item.master_option:
                flags.append("masteroption")
            if item.master_controlled:
                flags.append("master_controlled")
            if item.master_sync:
                flags.append("master_sync")
            lines.append(
                "| {key} | {default} | {worlds} | {desc} | {options} | {flags} |".format(
                    key=item.key,
                    default=item.default or "-",
                    worlds=worlds,
                    desc=item.description_ref or "-",
                    options=options,
                    flags=", ".join(flags) if flags else "-",
                )
            )
        lines.append("")
    return lines


if __name__ == "__main__":
    raise SystemExit(main())
