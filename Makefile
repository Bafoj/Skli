# Variables
BINARY_NAME=skli
MAIN_PATH=cmd/skli/main.go

.PHONY: all build run clean test tidy help

all: build

## build: Compila el binario
build:
	@echo "Compilando..."
	@go build -o bin/$(BINARY_NAME) $(MAIN_PATH)

## run: Ejecuta la aplicación directamente
run:
	@go run $(MAIN_PATH)

## clean: Elimina el binario compilado
clean:
	@echo "Limpiando..."
	@rm -rf bin/

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
