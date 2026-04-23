# Architecture

## Overview

`dst-server-ctl` is a single-binary controller for one managed Don't Starve Together dedicated server installation. It installs and updates DST with SteamCMD, generates DST configuration files from structured controller state, runs Master and Caves processes, streams logs, and exposes a local web UI.

The controller does not manage or mutate existing manual DST installations. All managed data is stored under a dedicated managed root, defaulting to `$XDG_DATA_HOME/dst-server-ctl` or `~/.local/share/dst-server-ctl`.

## Backend Layers

- `domain`: core types such as install layout, world, shard, mod, task, and server status.
- `service`: use-case orchestration such as install, update, start, stop, world selection, and config rendering.
- `adapter`: integrations with filesystem, process runner, SteamCMD, SQLite, log tailing, and DST config writers.
- `http`: API routing, request validation, authentication, and response shaping.

Dependencies point inward: HTTP calls services, services use domain types and adapter interfaces, adapters implement external effects.

## Runtime Data

Controller state is stored in SQLite. DST-native files are generated into the managed root from structured state.

The managed root layout will use stable subdirectories:

- `steamcmd/`: SteamCMD installation.
- `dst/`: DST dedicated server installation.
- `clusters/`: generated cluster directories, worlds, saves, tokens, and shard configs.
- `logs/`: controller logs and task logs.
- `state/`: SQLite database and local controller metadata.

## DST File Policy

The controller owns generated files for managed servers:

- `cluster.ini`
- `Master/server.ini`
- `Caves/server.ini`
- `Master/modoverrides.lua`
- `Caves/modoverrides.lua`
- `cluster_token.txt`
- allow/block/admin list files

Writers for DST files live in adapter code. Handlers and UI code must not hand-roll these formats.

## Process Policy

The controller directly launches and supervises Master and Caves processes for the single managed server instance. It is responsible for stop/start/restart, status detection, update safety, and log streaming.

The first release does not generate systemd units. A later installer may optionally wrap the controller itself in systemd, but DST child process management remains inside the controller.

## Security

The default listener is `127.0.0.1`. The controller generates an admin token on first run and requires it for mutating API requests. Secrets are masked in logs and ordinary API responses.

