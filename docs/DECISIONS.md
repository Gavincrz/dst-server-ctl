# Decisions

## 2026-04-23: No Docker Core

The project will not use Docker for the core deployment path. The controller installs and manages a native DST dedicated server under its own managed root.

## 2026-04-23: Go + Svelte

The backend is Go. The frontend is Svelte. Release artifacts should become a single Go binary with embedded static assets after the UI build pipeline is connected.

## 2026-04-23: SQLite State

Controller state uses SQLite. DST-native configuration remains as generated files under the managed root.

## 2026-04-23: Structured Configuration Source

The UI and API manage structured controller state. DST config files are generated from that state instead of edited in place.

## 2026-04-23: Controller-Managed Processes

The controller directly starts, stops, and monitors DST Master and Caves processes. systemd may manage the controller process later, but not the shard lifecycle in the first architecture.

