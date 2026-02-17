#!/bin/bash

# Script de análise de cobertura por pacote
# Identifica pacotes com baixa cobertura para priorização

set -e

echo "📊 ANÁLISE DE COBERTURA POR PACOTE"
echo "=================================="

# Gerar cobertura detalhada
go test ./... -coverprofile=coverage.out

# Analisar por pacote
echo ""
echo "📈 Cobertura por pacote:"
echo "------------------------"
go tool cover -func=coverage.out | grep -E "(github.com|total)" | while read line; do
    if echo "$line" | grep -q "total"; then
        echo "📊 $line"
    else
        pkg=$(echo "$line" | awk '{print $1}')
        cov=$(echo "$line" | awk '{print $3}')
        if [ "$cov" = "0.0%" ]; then
            echo "🔴 $pkg - $cov (SEM TESTES)"
        else
            # Remove % e converte para inteiro (bash não lida bem com floats)
            cov_int=$(echo "${cov%\%}" | cut -d. -f1)
            if [ "$cov_int" -lt 50 ]; then
                echo "🟡 $pkg - $cov (BAIXA)"
            else
                echo "🟢 $pkg - $cov (OK)"
            fi
        fi
    fi
done

# Contar pacotes sem testes
NO_TESTS=$(go tool cover -func=coverage.out | grep "0.0%" | wc -l)
TOTAL_PKGS=$(go list ./... | wc -l)
WITH_TESTS=$((TOTAL_PKGS - NO_TESTS))

echo ""
echo "📊 ESTATÍSTICAS:"
echo "----------------"
echo "📦 Total de pacotes: $TOTAL_PKGS"
echo "✅ Pacotes com testes: $WITH_TESTS"
echo "❌ Pacotes sem testes: $NO_TESTS"
echo "📈 Percentual com testes: $((WITH_TESTS * 100 / TOTAL_PKGS))%"

# Sugerir prioridades
echo ""
echo "🎯 SUGESTÕES DE PRIORIDADE:"
echo "---------------------------"
if [ $NO_TESTS -gt 0 ]; then
    echo "1. Começar pelos pacotes CRÍTICOS sem testes:"
    go tool cover -func=coverage.out | grep "0.0%" | head -5 | awk '{print "   - " $1}'
fi

echo ""
echo "2. Pacotes com maior impacto no sistema:"
echo "   - internal/api/handlers (já tem testes)"
echo "   - internal/core/services"
echo "   - internal/infra/postgres"
echo "   - internal/infra/storage"

echo ""
echo "🔥 Dica: Use 'go test -cover ./path/to/package' para focar em um pacote específico"