# Variables
BINARY_NAME=skli
MAIN_PATH=cmd/skli/main.go

.PHONY: all build run clean test tidy help releases goreleaser tag

all: build

## build: Compila el binario
build:
	@echo "Compilando..."
	@go build -o bin/$(BINARY_NAME) $(MAIN_PATH)

## releases: Compila los binarios para todas las plataformas
releases: clean
	@echo "Compilando para múltiples plataformas..."
	@mkdir -p releases
	GOOS=darwin GOARCH=amd64 go build -o releases/$(BINARY_NAME)-darwin-amd64 $(MAIN_PATH)
	GOOS=darwin GOARCH=arm64 go build -o releases/$(BINARY_NAME)-darwin-arm64 $(MAIN_PATH)
	GOOS=linux GOARCH=amd64 go build -o releases/$(BINARY_NAME)-linux-amd64 $(MAIN_PATH)
	GOOS=linux GOARCH=arm64 go build -o releases/$(BINARY_NAME)-linux-arm64 $(MAIN_PATH)
	GOOS=windows GOARCH=amd64 go build -o releases/$(BINARY_NAME)-windows-amd64.exe $(MAIN_PATH)
	@echo "Binarios generados en releases/"

## release: Crea una release real y la sube a GitHub (Requiere GITHUB_TOKEN)
release:
	@goreleaser release --clean

## snapshot: Genera una release local sin subir nada
snapshot:
	@goreleaser release --snapshot --clean

## tag: Crea un tag de git, sincroniza versiones en scripts y lo sube (Uso: make tag version=0.1.0)
tag:
	@if [ -z "$(version)" ]; then \
		echo "Error: Debes indicar la versión (sin 'v'). Ejemplo: make tag version=0.1.0"; \
		exit 1; \
	fi
	@if git rev-parse v$(version) >/dev/null 2>&1 || git ls-remote --tags origin v$(version) 2>/dev/null | grep -q v$(version); then \
		echo "Error: El tag v$(version) ya existe local o remotamente. Por favor, usa una versión superior."; \
		exit 1; \
	fi
	@echo "Sincronizando versiones en scripts..."
	@sed -i '' 's/VERSION=".*"/VERSION="$(version)"/' scripts/install.sh
	@sed -i '' 's/$$version = ".*"/$$version = "$(version)"/' scripts/install.ps1
	@git add scripts/install.sh scripts/install.ps1
	@git commit -m "chore: update version to v$(version) in installation scripts" || echo "No hay cambios en los scripts para commitear"
	@git push origin main
	@git tag -a v$(version) -m "Release v$(version)"
	@git push origin v$(version)
	@echo "Tag v$(version) creado y subido tras sincronizar scripts."

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
