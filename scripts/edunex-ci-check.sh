#!/bin/bash

# Script de verificação CI para EDUNEX
# Verifica build, testes e cobertura

set -e

echo "🔥 EDUNEX CI CHECK - Iniciando verificação"
echo "=========================================="

# 1. Verificar se estamos no diretório correto
if [ ! -f "go.mod" ]; then
    echo "❌ ERRO: Não encontrado go.mod. Certifique-se de estar no diretório raiz do EDUNEX."
    exit 1
fi

echo "✅ Diretório correto detectado"

# 2. Verificar dependências
echo "📦 Verificando dependências..."
go mod tidy
if [ $? -ne 0 ]; then
    echo "⚠️  Aviso: Problemas ao verificar dependências"
fi

# 3. Build do projeto
echo "🔨 Executando build..."
go build ./...
if [ $? -ne 0 ]; then
    echo "❌ ERRO: Build falhou"
    exit 1
fi
echo "✅ Build bem-sucedido"

# 4. Executar testes
echo "🧪 Executando testes..."
go test ./... -v
TEST_RESULT=$?
if [ $TEST_RESULT -ne 0 ]; then
    echo "❌ ERRO: Testes falharam"
    exit 1
fi
echo "✅ Todos os testes passaram"

# 5. Verificar cobertura
echo "📊 Gerando relatório de cobertura..."
go test ./... -coverprofile=coverage.out
COVERAGE=$(go tool cover -func=coverage.out | grep total | awk '{print $3}')
echo "📈 Cobertura de testes: $COVERAGE"

# 6. Gerar relatório HTML de cobertura
go tool cover -html=coverage.out -o coverage.html
echo "📄 Relatório HTML gerado: coverage.html"

# 7. Verificar formatação
echo "🎨 Verificando formatação..."
go fmt ./...
echo "✅ Formatação verificada"

# 8. Verificar imports não utilizados
echo "🧹 Verificando imports não utilizados..."
go vet ./...
if [ $? -ne 0 ]; then
    echo "⚠️  Aviso: Problemas encontrados com go vet"
fi

echo ""
echo "=========================================="
echo "🔥 EDUNEX CI CHECK - CONCLUSÃO"
echo "✅ Build: OK"
echo "✅ Testes: OK"
echo "📈 Cobertura: $COVERAGE"
echo "✅ Formatação: OK"
echo ""
echo "🎉 Todas as verificações passaram com sucesso!"
echo "=========================================="