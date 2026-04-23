# Research

## DST Dedicated Server

Key operational facts:

- DST dedicated server is installed with SteamCMD app `343050`.
- A normal caves-enabled world runs two shard processes: `Master` and `Caves`.
- Shared server settings live in `cluster.ini`.
- Shard-specific network settings live in each shard's `server.ini`.
- The cluster token is stored in `cluster_token.txt`.
- Server mods are downloaded through `dedicated_server_mods_setup.lua` entries such as `ServerModSetup("workshop_id")`.
- Enabled mods and mod configuration live in `modoverrides.lua`.

Local reference observed on this development VPS:

- Master process: `./dontstarve_dedicated_server_nullrenderer -persistent_storage_root /home/dontstarve/dst-server/dontstarve-config -conf_dir server_dir -cluster lhy_server -console -shard Master`
- Caves process: `./dontstarve_dedicated_server_nullrenderer -persistent_storage_root /home/dontstarve/dst-server/dontstarve-config -conf_dir server_dir -cluster lhy_server -console -shard Caves`
- This manual deployment is reference-only. The project must not import, mutate, or assume ownership of `/home/dontstarve/dst-server`.

Useful references:

- Klei Linux dedicated server guide: https://forums.kleientertainment.com/forums/topic/64441-dedicated-server-quick-setup-guide-linux/
- DST dedicated server guide: https://dontstarve.wiki.gg/wiki/Guides/Don%E2%80%99t_Starve_Together_Dedicated_Servers
- SteamCMD documentation: https://developer.valvesoftware.com/wiki/SteamCMD

## Comparable Projects

Existing projects are strongest around Docker deployment and CLI automation. This project intentionally differentiates through non-Docker single-binary deployment and a visual web UI for worlds and mods.

Examples:

- https://github.com/Jamesits/docker-dst-server
- https://hub.docker.com/r/superjump22/dontstarvetogether

## Agent Harness

The harness uses a short `AGENTS.md`, current architecture docs, decisions, and task tracking. The goal is to give coding agents persistent context without overloading every turn with stale or low-signal instructions.

References:

- OpenAI harness engineering: https://openai.com/index/harness-engineering/
- OpenAI Codex guide: https://openai.com/business/guides-and-resources/how-openai-uses-codex/
