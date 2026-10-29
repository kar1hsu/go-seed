SHELL := /bin/bash
export PATH := $(CURDIR)/.tools/bin:$(PATH)
COMPOSE := docker compose --env-file .env -p goseed -f deploy/compose/compose.yaml

.PHONY: help tools doctor infra-up infra-down dev build-go build check test gen gen-api gen-rpc gen-web verify-gen smoke compose-up compose-down
help:
	@echo 'init tools doctor infra-up infra-down migrate bootstrap dev gen check test test-auth build smoke auth-smoke compose-up compose-migrate compose-bootstrap compose-down'
tools:
	bash scripts/install-tools.sh
doctor:
	bash scripts/doctor.sh
infra-up:
	$(COMPOSE) up -d --wait mysql redis
infra-down:
	$(COMPOSE) stop mysql redis
build-go:
	mkdir -p bin
	go build -o bin/admin-api ./apps/admin/api
	go build -o bin/app-api ./apps/app/api
	go build -o bin/system-rpc ./services/system/rpc
	go build -o bin/probe ./cmd/probe
dev: build-go
	python3 scripts/dev.py
build: build-go
	pnpm --dir web/admin build
check:
	python3 scripts/check.py
	go vet ./...
	pnpm --dir web/admin typecheck
	pnpm --dir web/admin format:check
	$(MAKE) verify-gen
test:
	go test ./...
gen:
	bash scripts/gen.sh all
gen-api:
	APP="$(APP)" bash scripts/gen.sh api
gen-rpc:
	DOMAIN="$(DOMAIN)" bash scripts/gen.sh rpc
gen-web:
	bash scripts/gen.sh web
verify-gen:
	python3 scripts/verify-gen.py
smoke:
	python3 scripts/smoke.py
compose-up:
	$(COMPOSE) up -d --build --wait
compose-down:
	$(COMPOSE) down

.PHONY: init migrate bootstrap compose-migrate compose-bootstrap test-auth
init:
	python3 scripts/init-env.py
migrate:
	python3 scripts/system.py migrate
bootstrap:
	python3 scripts/system.py bootstrap
compose-migrate:
	python3 scripts/system.py --compose migrate
compose-bootstrap:
	python3 scripts/system.py --compose bootstrap
test-auth:
	python3 scripts/test-auth.py

.PHONY: auth-smoke
auth-smoke:
	python3 scripts/test-auth.py --http
