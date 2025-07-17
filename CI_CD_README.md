# 🔄 CI/CD para Framework deco

Este documento descreve os fluxos de CI/CD adaptados especificamente para o framework **deco**, que é um package Go, não uma aplicação.

## 📋 Visão Geral

O framework deco utiliza fluxos de CI/CD otimizados para packages Go, focando em:

- ✅ **Testes multiplataforma** (Linux, Windows, macOS)
- ✅ **Linting e validação de código**
- ✅ **Verificação de segurança**
- ✅ **Build e distribuição de binários**
- ✅ **Publicação no Go Proxy**
- ✅ **Geração automática de documentação**
- ✅ **Release management**

## 🚀 Workflows Disponíveis

### 1. CI Package (`.github/workflows/ci-package.yaml`)

**Trigger:** Push para `main` ou Pull Requests

**Jobs:**
- **test**: Testes em múltiplas plataformas e versões do Go
- **lint**: Linting com golangci-lint, goimports, go vet
- **security**: Verificação de vulnerabilidades com govulncheck
- **build**: Build do binário em múltiplas plataformas
- **validate**: Validação do go.mod e dependências
- **notify**: Notificações via Telegram

### 2. CD Package (`.github/workflows/cd-package.yaml`)

**Trigger:** Após CI Package bem-sucedido na branch `main`

**Jobs:**
- **get_commit_messages**: Coleta informações dos commits
- **build_and_release**: Build, versionamento e criação de release
- **publish_to_go_proxy**: Publicação no Go Proxy
- **generate_documentation**: Atualização automática de documentação
- **notify**: Notificações de sucesso/erro

### 3. Documentation (`.github/workflows/docs.yaml`)

**Trigger:** Mudanças em arquivos de código ou documentação

**Jobs:**
- **generate_docs**: Geração automática de documentação
- **validate_docs**: Validação da documentação gerada
- **update_main_readme**: Atualização do README principal
- **notify**: Notificações de atualização de docs

### 4. Release Drafter (`.github/workflows/release-drafter.yml`)

**Trigger:** Push para `main` ou Pull Requests

**Jobs:**
- **update_release_draft**: Geração automática de notas de release

## 🔧 Configurações

### GolangCI-Lint (`.golangci.yml`)

```yaml
# Linters habilitados
- gofmt, goimports, govet
- staticcheck, gosimple, ineffassign
- unused, misspell, gosec
- errcheck, gocritic

# Configurações específicas
- Timeout: 5 minutos
- Go version: 1.23
- Exclusões para arquivos de teste
```

### Codecov (`.codecov.yml`)

```yaml
# Configurações de cobertura
- Target: 80%
- Threshold: 5%
- Ignora: main.go, exemplos, testes
```

### Dependabot (`.github/dependabot.yml`)

```yaml
# Atualizações automáticas
- Go modules: Semanal
- GitHub Actions: Semanal
- Ignora atualizações major de dependências críticas
```

## 📦 Processo de Release

### 1. Versionamento Automático

```bash
# Incremento automático de versão
./.config/scripts/increment_version.sh
```

### 2. Build Multiplataforma

```bash
# Build para Linux, Windows, macOS
go build -ldflags="-s -w -X main.version=$VERSION" -o deco ./cmd/decorate-gen
```

### 3. Distribuição

- **GitHub Releases**: Binários para download
- **Go Proxy**: Package disponível via `go install`
- **Documentação**: Atualizada automaticamente

### 4. Instalação

```bash
# Instalação da versão mais recente
go install github.com/RodolfoBonis/deco/cmd/decorate-gen@latest

# Instalação de versão específica
go install github.com/RodolfoBonis/deco/cmd/decorate-gen@v1.0.0
```

## 🛠️ Comandos Locais

### Makefile

```bash
# Ver todos os comandos disponíveis
make help

# Pipeline completo
make all

# Apenas build
make build

# Testes com cobertura
make test-coverage

# Linting
make lint

# Verificação de segurança
make security

# Modo desenvolvimento
make dev
```

### Comandos Manuais

```bash
# Build local
go build -o deco ./cmd/decorate-gen

# Testes
go test -v -race ./...

# Linting
golangci-lint run

# Verificação de segurança
govulncheck ./...

# Geração de documentação
go doc -all ./pkg/decorators > docs/api.md
```

## 🔍 Monitoramento

### Notificações Telegram

- ✅ **Sucesso**: Detalhes do release, versão, links
- ❌ **Erro**: Informações de debug, logs, troubleshooting
- 📚 **Documentação**: Status de atualização de docs

### Métricas

- **Cobertura de testes**: Target 80%
- **Tempo de build**: Monitorado por job
- **Vulnerabilidades**: Bloqueia release se encontradas

## 🚨 Troubleshooting

### Problemas Comuns

1. **Build falha em Windows/macOS**
   - Verificar compatibilidade de código
   - Testar localmente em diferentes OS

2. **Linting falha**
   - Executar `make lint-fix`
   - Verificar configuração do golangci-lint

3. **Vulnerabilidades detectadas**
   - Atualizar dependências
   - Verificar se são falsos positivos

4. **Documentação não gera**
   - Verificar se o binário compila
   - Checar permissões de escrita

### Logs e Debug

```bash
# Ver logs detalhados do CI
# GitHub Actions > Workflows > [Workflow] > [Job] > [Step]

# Testar localmente
make all

# Verificar configurações
cat .golangci.yml
cat .codecov.yml
```

## 🔗 Links Úteis

- [GitHub Actions](https://github.com/RodolfoBonis/deco/actions)
- [Releases](https://github.com/RodolfoBonis/deco/releases)
- [Go Package](https://pkg.go.dev/github.com/RodolfoBonis/deco)
- [Documentação](https://github.com/RodolfoBonis/deco/tree/main/docs)

## 📝 Notas Importantes

1. **Não é uma aplicação**: Este framework não é deployado na AWS
2. **Package Go**: Foco em distribuição via Go Proxy
3. **Binário CLI**: O produto principal é um comando CLI
4. **Multiplataforma**: Build para Linux, Windows, macOS
5. **Documentação**: Gerada automaticamente a cada mudança

---

**Última atualização:** $(date)
**Versão do framework:** $(cat version.txt 2>/dev/null || echo "dev") 