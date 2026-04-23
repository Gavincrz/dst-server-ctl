# Agent Instructions

Read this file before changing code.

## Required Context

Before implementation work, read:

- `docs/ARCHITECTURE.md`
- `docs/TASKS.md`
- `docs/DECISIONS.md`

If the change affects deployment, DST files, process management, or configuration generation, also update the relevant docs.

## Architecture Rules

- Preserve the backend layering: `domain` defines concepts, `service` orchestrates use cases, `adapter` talks to the OS/DST/SQLite, and `http` exposes APIs.
- Do not put DST file-format generation in HTTP handlers or UI code.
- Do not let frontend code encode DST config file formats directly; use API schemas.
- Execute external commands only through the shared command runner.
- Pass command arguments as arrays. Never shell-concatenate user input.
- Treat tokens, server passwords, and admin credentials as secrets. Do not log or echo them in normal API responses.
- Add or update tests for behavior changes.
- If a design decision changes, update `docs/DECISIONS.md`.
- If a module boundary changes, update `docs/ARCHITECTURE.md`.

## Product Constraints

- No Docker support in the core architecture.
- Do not import or mutate a user's existing manual DST installation unless a future explicit migration feature is designed.
- The managed DST install, worlds, saves, and mod cache must live under the controller's managed root.
- The default web listener must bind to `127.0.0.1`.

