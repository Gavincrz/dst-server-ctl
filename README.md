# dst-server-ctl

`dst-server-ctl` is a planned web controller for installing, configuring, updating, and running a single Don't Starve Together dedicated server instance without Docker.

The project is currently in the harness and architecture phase. See:

- [AGENTS.md](AGENTS.md) for coding-agent rules.
- [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md) for module boundaries.
- [docs/ROADMAP.md](docs/ROADMAP.md) for staged product goals.
- [docs/TASKS.md](docs/TASKS.md) for current progress.

## Development

```sh
make dev
```

This starts:

- the Go backend on `http://127.0.0.1:8737`
- the Vite frontend on `http://127.0.0.1:5173`

Use `Ctrl+C` to stop both processes together.

The frontend is not embedded into the Go binary yet. During development, open `http://127.0.0.1:5173` in the browser; Vite proxies `/api` requests to the backend on `127.0.0.1:8737`.

Other useful checks:

```sh
make check
```

If you need to run services separately:

```sh
go test ./...
go run ./cmd/dst-server-ctl
```

The web frontend lives in `web/` and is intended to be embedded into the Go binary after the first UI milestone.
