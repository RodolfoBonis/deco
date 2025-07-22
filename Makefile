.PHONY: help build test lint clean install dev deps security bench release all check format test-infra-start test-infra-stop test-infra-restart test-infra-status

# Variáveis
BINARY_NAME=deco
MAIN_PATH=./cmd/deco
VERSION=$(shell cat version.txt 2>/dev/null || echo "dev")

help: ## Mostra esta ajuda
	@echo "Comandos disponíveis:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

build: ## Compila o binário
	@echo "🔨 Compilando $(BINARY_NAME) v$(VERSION)..."
	go build -v -ldflags="-s -w -X main.version=$(VERSION)" -o $(BINARY_NAME) $(MAIN_PATH)
	@echo "✅ Binário compilado: $(BINARY_NAME)"

test: ## Executa os testes
	@echo "🧪 Executando testes..."
	go test -v -race -coverprofile=coverage.out -covermode=atomic ./...
	@echo "✅ Testes concluídos"

test-coverage: test ## Executa testes com cobertura
	@echo "📊 Gerando relatório de cobertura..."
	go tool cover -html=coverage.out -o coverage.html
	@echo "✅ Relatório de cobertura gerado: coverage.html"

lint: ## Executa linting
	@echo "🔍 Executando linting..."
	golangci-lint run --timeout=5m
	@echo "✅ Linting concluído"

lint-fix: ## Corrige problemas de linting automaticamente
	@echo "🔧 Corrigindo problemas de linting..."
	golangci-lint run --fix --timeout=5m
	@echo "✅ Linting corrigido"

clean: ## Remove arquivos temporários
	@echo "🧹 Limpando arquivos temporários..."
	rm -f $(BINARY_NAME)
	rm -f coverage.out
	rm -f coverage.html
	@echo "✅ Limpeza concluída"

install: ## Instala o binário localmente
	@echo "📦 Instalando $(BINARY_NAME)..."
	go install $(MAIN_PATH)
	@echo "✅ $(BINARY_NAME) instalado"

dev: ## Inicia modo de desenvolvimento
	@echo "🚀 Iniciando modo de desenvolvimento..."
	@if [ -f "$(BINARY_NAME)" ]; then \
		./$(BINARY_NAME) dev; \
	else \
		echo "❌ Binário não encontrado. Execute 'make build' primeiro."; \
		exit 1; \
	fi

deps: ## Atualiza dependências
	@echo "📦 Atualizando dependências..."
	go mod tidy
	go mod download
	@echo "✅ Dependências atualizadas"

security: ## Verifica vulnerabilidades
	@echo "🔒 Verificando vulnerabilidades..."
	govulncheck ./...
	@echo "✅ Verificação de segurança concluída"

bench: ## Executa benchmarks
	@echo "⚡ Executando benchmarks..."
	go test -bench=. -benchmem ./...
	@echo "✅ Benchmarks concluídos"

release: build test lint security ## Prepara release (build + test + lint + security)
	@echo "🎉 Release preparado com sucesso!"

all: clean deps build test lint security ## Executa todo o pipeline
	@echo "🎉 Pipeline completo executado com sucesso!"

check: test lint security ## Executa verificações (test + lint + security)
	@echo "✅ Todas as verificações passaram!"

format: ## Formata o código
	@echo "🎨 Formatando código..."
	gofmt -w .
	goimports -w .
	@echo "✅ Código formatado"

# Infraestrutura de Teste
test-infra-start: ## Inicia infraestrutura de teste (Redis + OpenTelemetry)
	@echo "🚀 Iniciando infraestrutura de teste..."
	@./scripts/test-infra.sh start

test-infra-stop: ## Para infraestrutura de teste
	@echo "🛑 Parando infraestrutura de teste..."
	@./scripts/test-infra.sh stop

test-infra-restart: ## Reinicia infraestrutura de teste
	@echo "🔄 Reiniciando infraestrutura de teste..."
	@./scripts/test-infra.sh restart

test-infra-status: ## Mostra status da infraestrutura de teste
	@echo "📊 Status da infraestrutura de teste..."
	@./scripts/test-infra.sh status

test-with-infra: test-infra-start ## Executa testes com infraestrutura
	@echo "🧪 Executando testes com infraestrutura..."
	@make test
	@echo "🛑 Parando infraestrutura de teste..."
	@./scripts/test-infra.sh stop 