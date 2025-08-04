#!/bin/bash


echo "📊 Analisando Resultados dos Benchmarks"
echo "======================================="

# Detectar onde estão os resultados
RESULTS_DIR=""
if [ -d "benchmark_results" ]; then
    RESULTS_DIR="benchmark_results"
elif [ -d "../benchmark_results" ]; then
    RESULTS_DIR="../benchmark_results"
else
    echo "❌ Diretório de resultados não encontrado. Execute ./benchmark.sh primeiro."
    echo "   Procurando por: benchmark_results/"
    echo "   Diretório atual: $(pwd)"
    exit 1
fi

echo "📁 Analisando resultados em: $RESULTS_DIR"
echo ""

# Função para extrair métricas específicas
extract_metrics() {
    local file=$1
    local benchmark_name=$2

    echo "📈 Analisando $benchmark_name:"
    echo "----------------------------------------"

    if [ -f "$RESULTS_DIR/$file" ]; then
        # Extrair tempos de execução (ns/op)
        echo "⏱️  Tempos de execução (ns/op):"
        grep -E "Benchmark.*-[0-9]+" "$RESULTS_DIR/$file" | \
        awk '{print "  " $1 ": " $3}' | head -10

        echo ""

        # Extrair uso de memória (B/op)
        echo "💾 Uso de memória (B/op):"
        grep -E "Benchmark.*-[0-9]+" "$RESULTS_DIR/$file" | \
        awk '{if($5 != "") print "  " $1 ": " $5}' | head -10

        echo ""

        # Extrair número de alocações (allocs/op)
        echo "🔢 Alocações (allocs/op):"
        grep -E "Benchmark.*-[0-9]+" "$RESULTS_DIR/$file" | \
        awk '{if($7 != "") print "  " $1 ": " $7}' | head -10

        echo ""
    else
        echo "❌ Arquivo $file não encontrado"
    fi
}

# Analisar cada categoria
extract_metrics "creation_benchmark.txt" "Criação do AutoComplete"
extract_metrics "fuzzy_benchmark.txt" "Sugestões Fuzzy"
extract_metrics "history_benchmark.txt" "Operações de Histórico"
extract_metrics "scalability_benchmark.txt" "Escalabilidade"
extract_metrics "memory_benchmark.txt" "Uso de Memória"

# Resumo geral
echo "📋 RESUMO GERAL"
echo "==============="

if [ -f "$RESULTS_DIR/full_results.txt" ]; then
    echo "🏆 Top 5 benchmarks mais rápidos:"
    grep -E "Benchmark.*-[0-9]+" "$RESULTS_DIR/full_results.txt" | \
    sort -k3 -n | head -5 | \
    awk '{print "  " $1 ": " $3 " ns/op"}'

    echo ""

    echo "🐌 Top 5 benchmarks mais lentos:"
    grep -E "Benchmark.*-[0-9]+" "$RESULTS_DIR/full_results.txt" | \
    sort -k3 -nr | head -5 | \
    awk '{print "  " $1 ": " $3 " ns/op"}'

    echo ""

    echo "🧠 Top 5 benchmarks que mais consomem memória:"
    grep -E "Benchmark.*-[0-9]+" "$RESULTS_DIR/full_results.txt" | \
    awk '{if($5 != "") print $1 " " $5}' | \
    sort -k2 -nr | head -5 | \
    awk '{print "  " $1 ": " $2 " B/op"}'

    echo ""
fi

# Análise de vazamentos de memória
echo "🔍 ANÁLISE DE VAZAMENTOS"
echo "========================"

if [ -f "$RESULTS_DIR/memory_leak_test.txt" ]; then
    echo "📊 Resultado do teste de vazamento:"
    grep -E "(PASS|FAIL|Memory growth)" "$RESULTS_DIR/memory_leak_test.txt"
    echo ""
fi

# Comparação com baseline se existir
if [ -f "baseline_results.txt" ] && [ -f "$RESULTS_DIR/full_results.txt" ]; then
    echo "📈 COMPARAÇÃO COM BASELINE"
    echo "=========================="

    if command -v benchstat >/dev/null 2>&1; then
        echo "📊 Comparação detalhada (benchstat):"
        benchstat baseline_results.txt "$RESULTS_DIR/full_results.txt"
    else
        echo "⚠️  benchstat não instalado. Instale com:"
        echo "   go install golang.org/x/perf/cmd/benchstat@latest"
    fi
    echo ""
fi

# Recomendações baseadas nos resultados
echo "💡 RECOMENDAÇÕES"
echo "================"

# Verificar se há benchmarks lentos
if [ -f "$RESULTS_DIR/full_results.txt" ]; then
    slow_benchmarks=$(grep -E "Benchmark.*-[0-9]+" "$RESULTS_DIR/full_results.txt" | \
                     awk '$3 > 1000000 {print $1}' | wc -l)

    if [ "$slow_benchmarks" -gt 0 ]; then
        echo "⚠️  $slow_benchmarks benchmark(s) levam mais de 1ms para executar"
        echo "   Considere otimizar as operações mais lentas"
    fi

    # Verificar uso alto de memória
    high_memory=$(grep -E "Benchmark.*-[0-9]+" "$RESULTS_DIR/full_results.txt" | \
                 awk '$5 > 100000 {print $1}' | wc -l)

    if [ "$high_memory" -gt 0 ]; then
        echo "⚠️  $high_memory benchmark(s) usam mais de 100KB por operação"
        echo "   Considere otimizar o uso de memória"
    fi

    # Verificar muitas alocações
    high_allocs=$(grep -E "Benchmark.*-[0-9]+" "$RESULTS_DIR/full_results.txt" | \
                 awk '$7 > 100 {print $1}' | wc -l)

    if [ "$high_allocs" -gt 0 ]; then
        echo "⚠️  $high_allocs benchmark(s) fazem mais de 100 alocações por operação"
        echo "   Considere reutilizar objetos ou usar object pooling"
    fi
fi

echo ""
echo "🔧 PRÓXIMOS PASSOS"
echo "=================="
echo "1. Analise os profiles gerados:"
echo "   go tool pprof $RESULTS_DIR/cpu.prof"
echo "   go tool pprof $RESULTS_DIR/mem.prof"
echo ""
echo "2. Para análise visual dos profiles:"
echo "   go tool pprof -http=:8080 $RESULTS_DIR/cpu.prof"
echo ""
echo "3. Salve os resultados atuais como baseline:"
echo "   cp $RESULTS_DIR/full_results.txt baseline_results.txt"
echo ""
echo "4. Execute novamente após otimizações para comparar"
echo ""

# Gerar relatório HTML se pandoc estiver disponível
if command -v pandoc >/dev/null 2>&1; then
    echo "📄 Gerando relatório HTML..."

    cat > "$RESULTS_DIR/report.md" << EOF
# Relatório de Performance - AutoComplete

## Data: $(date)

## Resumo Executivo

$(if [ -f "$RESULTS_DIR/full_results.txt" ]; then
    total_benchmarks=$(grep -c "Benchmark.*-[0-9]+" "$RESULTS_DIR/full_results.txt")
    avg_time=$(grep -E "Benchmark.*-[0-9]+" "$RESULTS_DIR/full_results.txt" | awk '{sum+=$3; count++} END {print sum/count}')
    echo "- Total de benchmarks executados: $total_benchmarks"
    echo "- Tempo médio de execução: ${avg_time%.*} ns/op"
fi)

## Detalhes por Categoria

### Criação e Inicialização
\`\`\`
$(if [ -f "$RESULTS_DIR/creation_benchmark.txt" ]; then cat "$RESULTS_DIR/creation_benchmark.txt"; fi)
\`\`\`

### Sugestões Fuzzy
\`\`\`
$(if [ -f "$RESULTS_DIR/fuzzy_benchmark.txt" ]; then cat "$RESULTS_DIR/fuzzy_benchmark.txt"; fi)
\`\`\`

### Escalabilidade
\`\`\`
$(if [ -f "$RESULTS_DIR/scalability_benchmark.txt" ]; then cat "$RESULTS_DIR/scalability_benchmark.txt"; fi)
\`\`\`

EOF

    pandoc "$RESULTS_DIR/report.md" -o "$RESULTS_DIR/report.html"
    echo "✅ Relatório HTML gerado: $RESULTS_DIR/report.html"
else
    echo "💡 Instale pandoc para gerar relatórios HTML automaticamente"
fi

echo ""
echo "✅ Análise concluída!"
