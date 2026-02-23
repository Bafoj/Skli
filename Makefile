# Variables
BINARY_NAME=skli
MAIN_PATH=cmd/skli/main.go

VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "none")
DATE ?= $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS=-ldflags "-s -w -X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.date=$(DATE)"

.PHONY: all build run clean test tidy help releases goreleaser tag

all: build

## build: Compila el binario
build:
	@echo "Compilando versión $(VERSION)..."
	@go build $(LDFLAGS) -o bin/$(BINARY_NAME) $(MAIN_PATH)

## releases: Compila los binarios para todas las plataformas
releases: clean
	@echo "Compilando para múltiples plataformas..."
	@mkdir -p releases
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o releases/$(BINARY_NAME)-darwin-amd64 $(MAIN_PATH)
	GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o releases/$(BINARY_NAME)-darwin-arm64 $(MAIN_PATH)
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o releases/$(BINARY_NAME)-linux-amd64 $(MAIN_PATH)
	GOOS=linux GOARCH=arm64 go build $(LDFLAGS) -o releases/$(BINARY_NAME)-linux-arm64 $(MAIN_PATH)
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o releases/$(BINARY_NAME)-windows-amd64.exe $(MAIN_PATH)
	@echo "Binarios generados en releases/"

## release: Crea una release real y la sube a GitHub (Requiere GITHUB_TOKEN)
release:
	@goreleaser release --clean

## snapshot: Genera una release local sin subir nada
snapshot:
	@goreleaser release --snapshot --clean

## tag: Crea un tag de git, sincroniza versiones en scripts y lo sube (Uso: make tag version=0.1.0 o make tag type=patch|minor|major)
tag:
	@if [ -n "$(type)" ]; then \
		LATEST=$$(git tag --sort=-v:refname 2>/dev/null | head -n 1); \
		if [ -z "$$LATEST" ]; then LATEST="v0.0.0"; fi; \
		LATEST=$${LATEST#v}; \
		MAJOR=$$(echo $$LATEST | cut -d. -f1); \
		MINOR=$$(echo $$LATEST | cut -d. -f2); \
		PATCH=$$(echo $$LATEST | cut -d. -f3); \
		if [ "$(type)" = "major" ]; then MAJOR=$$((MAJOR + 1)); MINOR=0; PATCH=0; \
		elif [ "$(type)" = "minor" ]; then MINOR=$$((MINOR + 1)); PATCH=0; \
		elif [ "$(type)" = "patch" ]; then PATCH=$$((PATCH + 1)); \
		else echo "Error: tipo no válido. Usa type=major|minor|patch"; exit 1; fi; \
		VERSION="$$MAJOR.$$MINOR.$$PATCH"; \
	elif [ -n "$(version)" ]; then \
		VERSION="$(version)"; \
	else \
		echo "Error: Debes indicar la versión o el tipo. Ejemplo: make tag version=0.1.0 o make tag type=patch"; \
		exit 1; \
	fi; \
	if git rev-parse v$$VERSION >/dev/null 2>&1 || git ls-remote --tags origin v$$VERSION 2>/dev/null | grep -q v$$VERSION; then \
		echo "Error: El tag v$$VERSION ya existe local o remotamente. Por favor, usa una versión superior."; \
		exit 1; \
	fi; \
	echo "Sincronizando versiones en scripts a la versión $$VERSION..."; \
	sed -i '' "s/VERSION=\".*\"/VERSION=\"$$VERSION\"/" scripts/install.sh; \
	sed -i '' "s/\\$$version = \".*\"/\\$$version = \"$$VERSION\"/" scripts/install.ps1; \
	git add scripts/install.sh scripts/install.ps1; \
	git commit -m "chore: update version to v$$VERSION in installation scripts" || echo "No hay cambios en los scripts para commitear"; \
	git push origin main; \
	git tag -a v$$VERSION -m "Release v$$VERSION"; \
	git push origin v$$VERSION; \
	echo "Tag v$$VERSION creado y subido tras sincronizar scripts."

## run: Ejecuta la aplicación directamente
run:
	@go run $(MAIN_PATH) $(args)

## clean: Elimina los binarios y el directorio de distribución
clean:
	@echo "Limpiando..."
	@rm -rf bin/ releases/ dist/

## test: Ejecuta los tests del proyecto
test:
	@go test ./...

## tidy: Limpia y descarga las dependencias
tidy:
	@go mod tidy

## help: Muestra esta ayuda
help:
	@echo "Uso: make [comando]"
	@echo ""
	@echo "Comandos disponibles:"
	@sed -n 's/^##//p' $(MAKEFILE_LIST) | column -t -s ':' |  sed -e 's/^/ /'
