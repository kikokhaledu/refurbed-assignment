GO ?= go
NPM ?= npm
COMPOSE ?= docker compose
BACKEND_DIR ?= backend
FRONTEND_VUE_DIR ?= assignment_vue/frontend-vue
ENV_FILE ?= $(if $(wildcard .env),.env,.env.example)
GOLANGCI_LINT_VERSION ?= v1.64.8
GOLANGCI_LINT ?= $(GO) run github.com/golangci/golangci-lint/cmd/golangci-lint@$(GOLANGCI_LINT_VERSION)
COMPOSE_WITH_ENV = $(COMPOSE) --env-file $(ENV_FILE)

ifeq ($(OS),Windows_NT)
PWSH := powershell -NoProfile -Command
ECHO_BLANK := @$(PWSH) "Write-Output ''"
CHECK_ENV_FILE := @$(PWSH) "if (-not (Test-Path '$(ENV_FILE)')) { Write-Error 'Missing $(ENV_FILE)'; exit 1 }"
CHECK_METADATA_FILE := @$(PWSH) "if (-not (Test-Path '$(BACKEND_DIR)/data/metadata.json')) { Write-Error 'Missing $(BACKEND_DIR)/data/metadata.json'; exit 1 }"
CHECK_DETAILS_FILE := @$(PWSH) "if (-not (Test-Path '$(BACKEND_DIR)/data/details.json')) { Write-Error 'Missing $(BACKEND_DIR)/data/details.json'; exit 1 }"
RUN_DOCKER_CHECK := @$(PWSH) 'docker --version > $$null 2>&1; if ($$LASTEXITCODE -ne 0) { Write-Error "[docker-check] Docker CLI not found. Install Docker Desktop and retry."; exit 1 }; docker info > $$null 2>&1; if ($$LASTEXITCODE -ne 0) { Write-Error "[docker-check] Docker engine is not reachable. Start Docker Desktop and wait until it is running, then retry make up."; exit 1 }; Write-Output "[docker-check] Docker engine reachable [OK]"'
RUN_FMT_CHECK := @$(PWSH) 'Push-Location "$(BACKEND_DIR)"; $$unformatted = gofmt -l .; Pop-Location; if ($$unformatted) { $$unformatted; Write-Error "[fmt-check] Unformatted files found"; exit 1 }; Write-Output "[fmt-check] Passed [OK]"'
RUN_COMPOSE_CHECK := @$(PWSH) '$(COMPOSE_WITH_ENV) config *> $$null; if ($$LASTEXITCODE -ne 0) { exit 1 }; Write-Output "[compose-check] Passed [OK]"'
RUN_TEST_RACE := @$(PWSH) 'if (-not (Get-Command gcc -ErrorAction SilentlyContinue)) { [Console]::Error.WriteLine("[test-race] gcc not found. Install gcc (for example via MSYS2/MinGW) and retry."); exit 1 }; Push-Location "$(BACKEND_DIR)"; $$env:CGO_ENABLED = "1"; $(GO) test ./... -race; $$exitCode = $$LASTEXITCODE; Pop-Location; if ($$exitCode -ne 0) { exit $$exitCode }; Write-Output "[test-race] Passed [OK]"'
RUN_FRONTEND_DEPS := @$(PWSH) 'Push-Location "$(FRONTEND_VUE_DIR)"; $(NPM) ci; if ($$LASTEXITCODE -ne 0) { Write-Output "[frontend-deps] npm ci failed, retrying with npx npm@10 ci..."; npx --yes npm@10 ci; if ($$LASTEXITCODE -ne 0) { Pop-Location; exit $$LASTEXITCODE } }; Pop-Location; Write-Output "[frontend-deps] Installed [OK]"'
else
ECHO_BLANK := @printf '\n'
CHECK_ENV_FILE := @test -f $(ENV_FILE)
CHECK_METADATA_FILE := @test -f $(BACKEND_DIR)/data/metadata.json
CHECK_DETAILS_FILE := @test -f $(BACKEND_DIR)/data/details.json
RUN_DOCKER_CHECK := @command -v docker > /dev/null 2>&1 || { echo "[docker-check] Docker CLI not found. Install Docker and retry."; exit 1; }; docker info > /dev/null 2>&1 || { echo "[docker-check] Docker engine is not reachable. Start Docker daemon/Desktop and retry."; exit 1; }; echo [docker-check] Docker engine reachable [OK]
RUN_FMT_CHECK := @cd $(BACKEND_DIR) && files="$$(gofmt -l .)"; if [ -n "$$files" ]; then echo "$$files"; echo "[fmt-check] Unformatted files found"; exit 1; fi && echo [fmt-check] Passed [OK]
RUN_COMPOSE_CHECK := @$(COMPOSE_WITH_ENV) config > /dev/null && echo [compose-check] Passed [OK]
RUN_TEST_RACE := @command -v gcc > /dev/null 2>&1 || { echo "[test-race] gcc not found. Install gcc and retry."; exit 1; }; cd $(BACKEND_DIR) && CGO_ENABLED=1 $(GO) test ./... -race && echo [test-race] Passed [OK]
RUN_FRONTEND_DEPS := @cd $(FRONTEND_VUE_DIR) && ($(NPM) ci || (echo "[frontend-deps] npm ci failed, retrying with npx npm@10 ci..."; command -v npx > /dev/null 2>&1 || { echo "[frontend-deps] npx not found. Install Node.js 18+ and npm 9+."; exit 1; }; npx --yes npm@10 ci)) && echo [frontend-deps] Installed [OK]
endif

.PHONY: help env-file-check docker-check up down reset seed logs test test-backend frontend-deps test-frontend frontend-lint frontend-check test-e2e e2e-install test-race fmt fmt-check vet lint compose-check pre-push clean

help:
	@echo Available targets:
	$(ECHO_BLANK)
	@echo [Environment]
	@echo   make env-file-check - Verify selected env file exists: .env or .env.example
	@echo   make docker-check    - Verify Docker CLI + daemon availability
	@echo   make seed           - Validate backend seed data files
	$(ECHO_BLANK)
	@echo [Docker Stack]
	@echo   make up             - Build and start backend + Vue frontend with Docker Compose
	@echo   make down           - Stop Compose services
	@echo   make reset          - Recreate full stack from scratch
	@echo   make logs           - Tail backend + frontend logs
	@echo   make clean          - Stop Compose services and remove volumes
	@echo   make compose-check  - Validate docker-compose.yml
	$(ECHO_BLANK)
	@echo [Backend Quality]
	@echo   make fmt            - Format backend Go code
	@echo   make fmt-check      - Check backend Go formatting without modifying files
	@echo   make vet            - Run go vet on backend
	@echo   make lint           - Run golangci-lint on backend
	$(ECHO_BLANK)
	@echo [Testing]
	@echo   make test           - Run backend + frontend unit tests
	@echo   make test-backend   - Run backend tests
	@echo   make frontend-deps  - Install frontend dependencies via npm ci
	@echo   make test-frontend  - Run frontend unit tests via Vitest
	@echo   make frontend-lint  - Run frontend ESLint checks
	@echo   make frontend-check - Run frontend production build check
	@echo   make test-e2e       - Run frontend Playwright E2E tests
	@echo   make e2e-install    - Install Playwright Chromium browser
	@echo   make test-race      - Run backend race tests and requires CGO plus gcc
	$(ECHO_BLANK)
	@echo [CI / Gate]
	@echo   make pre-push       - Run fmt, vet, lint, unit tests, E2E tests, and compose validation

env-file-check:
	$(CHECK_ENV_FILE)
	@echo [env] Using env file: $(ENV_FILE) [OK]

docker-check:
	$(RUN_DOCKER_CHECK)

up: env-file-check docker-check compose-check
	@echo [up] Starting backend + frontend with Docker Compose...
	@$(COMPOSE_WITH_ENV) up --build -d backend frontend && echo [up] Services are up [OK]

down: env-file-check docker-check
	@echo [down] Stopping Docker Compose services...
	@$(COMPOSE_WITH_ENV) down --remove-orphans && echo [down] Services stopped [OK]

reset: down seed up

seed:
	@echo [seed] Checking seed files...
	$(CHECK_METADATA_FILE)
	$(CHECK_DETAILS_FILE)
	@echo [seed] Seed data is present in $(BACKEND_DIR)/data [OK]

logs: env-file-check docker-check
	$(COMPOSE_WITH_ENV) logs -f backend frontend

test: test-backend test-frontend
	@echo [test] Backend + frontend unit tests passed [OK]

test-backend:
	@echo [test-backend] Running Go tests...
	@cd $(BACKEND_DIR) && $(GO) test ./... && echo [test-backend] Passed [OK]

frontend-deps:
	@echo [frontend-deps] Installing frontend dependencies...
	$(RUN_FRONTEND_DEPS)

test-frontend: frontend-deps
	@echo [test-frontend] Running frontend unit tests...
	@cd $(FRONTEND_VUE_DIR) && $(NPM) run test:run && echo [test-frontend] Passed [OK]

frontend-lint: frontend-deps
	@echo [frontend-lint] Running frontend ESLint...
	@cd $(FRONTEND_VUE_DIR) && $(NPM) run lint && echo [frontend-lint] Passed [OK]

frontend-check: frontend-deps
	@echo [frontend-check] Running frontend production build check...
	@cd $(FRONTEND_VUE_DIR) && $(NPM) run build && echo [frontend-check] Passed [OK]

test-e2e: e2e-install
	@echo [test-e2e] Running Playwright E2E tests...
	@cd $(FRONTEND_VUE_DIR) && $(NPM) run e2e && echo [test-e2e] Passed [OK]

e2e-install: frontend-deps
	@echo [e2e-install] Installing Playwright Chromium...
	@cd $(FRONTEND_VUE_DIR) && $(NPM) run e2e:install && echo [e2e-install] Done [OK]

test-race:
	@echo [test-race] Running Go race tests - requires CGO and gcc...
	$(RUN_TEST_RACE)

fmt:
	@echo [fmt] Formatting backend Go code...
	@cd $(BACKEND_DIR) && $(GO) fmt ./... && echo [fmt] Passed [OK]

fmt-check:
	@echo [fmt-check] Checking backend Go formatting...
	$(RUN_FMT_CHECK)

vet:
	@echo [vet] Running go vet...
	@cd $(BACKEND_DIR) && $(GO) vet ./... && echo [vet] Passed [OK]

lint:
	@echo [lint] Running golangci-lint...
	@cd $(BACKEND_DIR) && $(GOLANGCI_LINT) run ./... && echo [lint] Passed [OK]

compose-check: env-file-check docker-check
	@echo [compose-check] Validating docker-compose.yml...
	$(RUN_COMPOSE_CHECK)

pre-push: fmt-check vet lint frontend-lint frontend-check test test-e2e compose-check
	@echo [pre-push] All checks passed [OK]

clean: env-file-check docker-check
	@echo [clean] Stopping services and removing volumes...
	@$(COMPOSE_WITH_ENV) down --remove-orphans --volumes && echo [clean] Done [OK]
