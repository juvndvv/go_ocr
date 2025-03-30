.PHONY: build watch run dev install stop clean

APP_NAME := build/ocr-api
PID_FILE := build/pid.txt

install:
	@echo "Instalando dependencias..."
	@go mod tidy
	@echo "Dependencias instaladas."
	@cp .env.example .env

build:
	@echo "Compilando aplicación..."
	@echo "Copiando archivo .env..."
	@cp .env build/.env
	@go build -o $(APP_NAME) main.go
	@echo "Compilación completada."

run: build
	@echo "Iniciando aplicación..."
	@./$(APP_NAME) & echo $$! > $(PID_FILE)
	@echo "Aplicación ejecutándose en segundo plano (PID: $$(cat $(PID_FILE)))"

stop:
	@if [ -f $(PID_FILE) ]; then \
		echo "Deteniendo aplicación (PID: $$(cat $(PID_FILE)))..."; \
		kill $$(cat $(PID_FILE)) || true; \
		rm -f $(PID_FILE); \
		echo "Aplicación detenida."; \
	else \
		echo "No hay aplicación en ejecución."; \
	fi

watch:
	@echo "Observando cambios en archivos .go (Ctrl+C para salir)..."
	@fswatch -or --event=Updated --include='.*\.go$$' . | while read; do \
		make stop; \
		make run; \
	done

clean:
	@make stop
	@rm -f $(APP_NAME)

dev: watch