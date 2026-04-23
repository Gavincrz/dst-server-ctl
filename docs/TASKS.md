# Tasks

## Current State

- [x] Initialize git repository.
- [x] Add harness documentation.
- [x] Add backend skeleton.
- [x] Add frontend skeleton.
- [x] Add initial tests for path and command primitives.
- [x] Configure local Go toolchain and verify backend tests.

The project is ready for the first product implementation iteration. Current code is a harness and skeleton only; it does not install or manage DST yet.

## Next Task

- [ ] Add SQLite migration layer and state repository interfaces.

## Backlog

- [ ] Define install status API and managed root initialization.
- [ ] Implement SteamCMD installer planning and task model.
- [ ] Add first Svelte status page wired to backend `/api/v1/status`.

## Do Not Do Yet

- Do not add Docker support.
- Do not import, migrate, or mutate the manual DST install under `/home/dontstarve/dst-server`.
- Do not build the full mod-management UI before the managed install, state, and process lifecycle are in place.
- Do not expose the web UI on a public interface by default.

## Completion Checklist

- [ ] Relevant backend/frontend checks have been run.
- [ ] Go files touched in the iteration have been formatted.
- [ ] Architecture or decision docs have been updated if boundaries or technical choices changed.
- [ ] This file reflects completed work and the next task.
- [ ] Changes have been committed with a focused message.

## Open Questions

- Exact installer script UX.
- Whether admin token should be file-only, env-overridable, or both.
- How much of `leveldataoverride.lua` should be visual in the first public release.
