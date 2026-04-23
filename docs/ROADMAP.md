# Roadmap

## Phase 0: Harness and Skeleton

- Establish agent instructions and architecture documents.
- Create Go backend skeleton with package boundaries.
- Create Svelte frontend skeleton.
- Add tests for safe path and command-runner primitives.

## Phase 1: Managed Installation MVP

- Initialize managed root.
- Download/install SteamCMD.
- Install DST app `343050`.
- Persist install state in SQLite.
- Show install status in the web UI.

## Phase 2: Server Lifecycle MVP

- Create one managed cluster with Master and Caves.
- Generate core `cluster.ini` and `server.ini` files.
- Start, stop, restart, and inspect shard status.
- Stream Master and Caves logs in the UI.

## Phase 3: Updates

- Run manual SteamCMD updates.
- Check local and remote version state.
- Add daily scheduled update checks.
- Require explicit confirmation before stopping a running world for update.

## Phase 4: World and Server Config UI

- Configure server name, description, password, language, max players, PvP, pause behavior, and game mode.
- Manage token input without echoing the token.
- Manage admin, block, and allow lists.
- Add world templates and generated `leveldataoverride.lua`.

## Phase 5: Mod Management

- Add Workshop IDs and generate `dedicated_server_mods_setup.lua`.
- Update/download mods.
- Read local `modinfo.lua` metadata after download.
- Generate visual forms for supported `configuration_options`.
- Generate `modoverrides.lua` for Master and Caves.

## Phase 6: Packaging

- Add release builds for Linux amd64/arm64.
- Add optional installer script.
- Add clean uninstall workflow with prompts to preserve or delete worlds, saves, mods, and DST install files.

