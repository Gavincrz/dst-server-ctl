# dst-server-ctl

`dst-server-ctl` is a planned web controller for installing, configuring, updating, and running a single Don't Starve Together dedicated server instance without Docker.

The project is currently in the harness and architecture phase. See:

- [AGENTS.md](AGENTS.md) for coding-agent rules.
- [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md) for module boundaries.
- [docs/ROADMAP.md](docs/ROADMAP.md) for staged product goals.
- [docs/TASKS.md](docs/TASKS.md) for current progress.

## Development

```sh
go test ./...
go run ./cmd/dst-server-ctl
```

The web frontend lives in `web/` and is intended to be embedded into the Go binary after the first UI milestone.

