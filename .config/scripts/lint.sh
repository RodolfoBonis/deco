#!/bin/bash
# Script para rodar ferramentas de lint e análise estática no projeto Go
# Agora coleta todos os erros e só falha no final, mostrando um resumo

FAIL=0

# 1. gofmt (formatação)
FMT_OUT=$(gofmt -l .)
if [ -n "$FMT_OUT" ]; then
  echo -e "\nArquivos com problemas de formatação (gofmt):"
  echo "$FMT_OUT"
  FAIL=1
else
  echo "gofmt: OK"
fi

# 2. go vet (erros comuns)
VET_OUT=$(go vet ./... 2>&1)
if [ -n "$VET_OUT" ]; then
  echo -e "\nProblemas encontrados pelo go vet:"
  echo "$VET_OUT"
  FAIL=1
else
  echo "go vet: OK"
fi

# 3. golangci-lint (análise estática e boas práticas)
if ! command -v golangci-lint &> /dev/null; then
  echo "golangci-lint não encontrado. Instale com: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"
  FAIL=1
else
  LINT_OUT=$(golangci-lint run 2>&1)
  if [ $? -ne 0 ]; then
    echo -e "\nProblemas encontrados pelo golangci-lint:"
    echo "$LINT_OUT"
    FAIL=1
  else
    echo "golangci-lint: OK"
  fi
fi

# 4. staticcheck (análise avançada)
if ! command -v staticcheck &> /dev/null; then
  echo "staticcheck não encontrado. Instale com: go install honnef.co/go/tools/cmd/staticcheck@latest"
  FAIL=1
else
  STATIC_OUT=$(staticcheck ./... 2>&1)
  if [ -n "$STATIC_OUT" ]; then
    echo -e "\nProblemas encontrados pelo staticcheck:"
    echo "$STATIC_OUT"
    FAIL=1
  else
    echo "staticcheck: OK"
  fi
fi

# 5. goimports (organização dos imports)
if ! command -v goimports &> /dev/null; then
  echo "goimports não encontrado. Instale com: go install golang.org/x/tools/cmd/goimports@latest"
  FAIL=1
else
  IMP_OUT=$(goimports -l .)
  if [ -n "$IMP_OUT" ]; then
    echo -e "\nArquivos com imports desorganizados (goimports):"
    echo "$IMP_OUT"
    FAIL=1
  else
    echo "goimports: OK"
  fi
fi

if [ $FAIL -eq 0 ]; then
  echo -e "\n✅ Lint finalizado com sucesso!"
else
  echo -e "\n❌ Foram encontrados problemas de lint. Veja os detalhes acima."
  exit 1
fi 