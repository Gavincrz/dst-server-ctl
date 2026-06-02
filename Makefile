SHELL := /bin/bash

.PHONY: dev check extract-config-fields

dev:
	./scripts/dev.sh

check:
	go test ./...
	cd web && npm run check
	cd web && npm run build

extract-config-fields:
	python3 ./scripts/extract_dst_config_fields.py
