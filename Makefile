SHELL := /bin/bash

.PHONY: dev check

dev:
	./scripts/dev.sh

check:
	go test ./...
	cd web && npm run check
	cd web && npm run build
