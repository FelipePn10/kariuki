#!/bin/bash


set -e  # Parar em caso de erro

echo "🚀 Executando Benchmarks do AutoComplete"
echo "========================================"

# Detectar estrutura do projeto
PROJECT_ROOT=""
AUTOCOMPLETE_DIR=""

# Verificar se estamos no diretório raiz do projeto
if [ -f "go.mod" ] && [ -d "pkg/autocomplete" ]; then
    PROJECT_ROOT="$(pwd)"
    AUTOCOMPLETE_DIR="pkg/autocomplete"
elif [ -f "../go.mod" ] && [ -d "../pkg/autocomplete" ]; then
    PROJECT_ROOT="$(cd .. && pwd)"
    AUTOCOMPLETE_DIR="pkg/autocomplete"
elif [ -f "autocomplete/autocomplete.go" ]; then
    PROJECT_ROOT="$(pwd)"
    AUTOCOMPLETE_DIR="autocomplete"
elif [ -f "go.mod" ] && [ -f "autocomplete.go" ]; then
    # Estamos diretamente no diretório do autocomplete
    PROJECT_ROOT="$(cd .. && pwd)"
    AUTOCOMPLETE_DIR="$(basename $(pwd))"
else
    echo "❌ Error: Could not find autocomplete.go. Please ensure you're in the project root or autocomplete directory."
    echo "   Looking for one of these structures:"
    echo "   - pkg/autocomplete/autocomplete.go"
    echo "   - autocomplete/autocomplete.go"
    echo "   Current directory: $(pwd)"
    echo "   Files found: $(ls -la)"
    exit 1
fi

echo "📁 Detected project structure:"
echo "   Project root: $PROJECT_ROOT"
echo "   Autocomplete: $AUTOCOMPLETE_DIR"

# Criar estrutura de diretórios
RESULTS_DIR="$PROJECT_ROOT/benchmark_results"
FULL_AUTOCOMPLETE_PATH="$PROJECT_ROOT/$AUTOCOMPLETE_DIR"

echo "   Results dir: $RESULTS_DIR"
echo "   Full autocomplete path: $FULL_AUTOCOMPLETE_PATH"
echo ""

# Verificar se o arquivo autocomplete.go existe
if [ ! -f "$FULL_AUTOCOMPLETE_PATH/autocomplete.go" ]; then
    echo "❌ Error: autocomplete.go not found at $FULL_AUTOCOMPLETE_PATH/autocomplete.go"
    exit 1
fi

# Criar diretórios necessários
mkdir -p "$RESULTS_DIR"
mkdir -p "$FULL_AUTOCOMPLETE_PATH/testdata"

echo "📁 Results will be saved to: $RESULTS_DIR"
echo ""

# Função para executar benchmark com tratamento de erro
run_benchmark() {
    local name="$1"
    local pattern="$2"
    local output_file="$3"
    local count="${4:-3}"

    echo "🔍 Running $name..."

    if go test -bench="$pattern" -benchmem -count="$count" > "$output_file" 2>&1; then
        echo "✅ $name completed successfully"
    else
        echo "⚠️  $name had issues, but continuing..."
        echo "Check $output_file for details"
    fi
}

# Navegar para o diretório do autocomplete
cd "$FULL_AUTOCOMPLETE_PATH"

echo "📍 Current directory: $(pwd)"
echo "📋 Go module: $(go list -m 2>/dev/null || echo 'No module found')"
echo ""

# Verificar se há testes
if [ ! -f "autocomplete_test.go" ]; then
    echo "⚠️  Warning: autocomplete_test.go not found. Benchmarks may not work."
fi

# Executar benchmarks principais
echo "📊 Running main benchmarks..."
if go test -bench=. -benchmem -count=3 -timeout=30m > "$RESULTS_DIR/full_results.txt" 2>&1; then
    echo "✅ Main benchmarks completed"
else
    echo "❌ Main benchmarks failed. Check $RESULTS_DIR/full_results.txt for details"
    cat "$RESULTS_DIR/full_results.txt"
    exit 1
fi

# Executar benchmarks específicos
run_benchmark "creation benchmarks" "BenchmarkNewAutocomplete|BenchmarkCollectAllCommands|BenchmarkBuildCompleter" "$RESULTS_DIR/creation_benchmark.txt" 5

run_benchmark "fuzzy suggestion benchmarks" "BenchmarkGetFuzzySuggestions" "$RESULTS_DIR/fuzzy_benchmark.txt" 5

run_benchmark "history benchmarks" "BenchmarkAddToHistory|BenchmarkSuggestHistory" "$RESULTS_DIR/history_benchmark.txt" 5

run_benchmark "I/O benchmarks" "BenchmarkSaveHistoryToDisk|BenchmarkLoadHistoryFromDisk" "$RESULTS_DIR/io_benchmark.txt" 3

run_benchmark "scalability benchmarks" "BenchmarkScalability" "$RESULTS_DIR/scalability_benchmark.txt" 3

run_benchmark "memory benchmarks" "BenchmarkMemory" "$RESULTS_DIR/memory_benchmark.txt" 3

run_benchmark "latency benchmarks" "BenchmarkLatency|BenchmarkWorstCase" "$RESULTS_DIR/latency_benchmark.txt" 3

run_benchmark "special cases benchmarks" "BenchmarkUnicode|BenchmarkOverhead" "$RESULTS_DIR/special_cases_benchmark.txt" 3

run_benchmark "concurrency benchmarks" "BenchmarkConcurrency" "$RESULTS_DIR/concurrency_benchmark.txt" 3

# CPU Profiling
echo "🔬 Generating CPU profile..."
if go test -bench=BenchmarkCompleteFlow -cpuprofile="$RESULTS_DIR/cpu.prof" -benchmem -count=1 > "$RESULTS_DIR/cpu_profile_output.txt" 2>&1; then
    echo "✅ CPU profile generated successfully"
else
    echo "⚠️  CPU profiling had issues"
    cat "$RESULTS_DIR/cpu_profile_output.txt"
fi

# Memory Profiling
echo "🧮 Generating memory profile..."
if go test -bench=BenchmarkMemory_FuzzySuggestions -memprofile="$RESULTS_DIR/mem.prof" -benchmem -count=1 > "$RESULTS_DIR/mem_profile_output.txt" 2>&1; then
    echo "✅ Memory profile generated successfully"
else
    echo "⚠️  Memory profiling had issues"
    cat "$RESULTS_DIR/mem_profile_output.txt"
fi

# Memory leak test
echo "🔍 Running memory leak test..."
if go test -run=TestMemoryLeak -v > "$RESULTS_DIR/memory_leak_test.txt" 2>&1; then
    echo "✅ Memory leak test completed"
else
    echo "⚠️  Memory leak test had issues"
fi

# Voltar para o diretório raiz
cd "$PROJECT_ROOT"

echo ""
echo "✅ Benchmarks completed!"
echo ""
echo "📈 Results saved in benchmark_results/:"

# Verificar quais arquivos foram criados com sucesso
check_file() {
    local file="$1"
    local description="$2"

    if [ -f "$RESULTS_DIR/$file" ] && [ -s "$RESULTS_DIR/$file" ]; then
        echo "  ✅ $file ($description)"
    else
        echo "  ❌ $file ($description) - not created or empty"
    fi
}

check_file "full_results.txt" "all benchmarks"
check_file "creation_benchmark.txt" "AutoComplete creation"
check_file "fuzzy_benchmark.txt" "fuzzy suggestions"
check_file "history_benchmark.txt" "history operations"
check_file "io_benchmark.txt" "disk I/O operations"
check_file "scalability_benchmark.txt" "scalability tests"
check_file "memory_benchmark.txt" "memory usage"
check_file "latency_benchmark.txt" "response time"
check_file "special_cases_benchmark.txt" "unicode, edge cases"
check_file "concurrency_benchmark.txt" "parallel execution"
check_file "cpu.prof" "CPU profile"
check_file "mem.prof" "memory profile"
check_file "memory_leak_test.txt" "memory leak test"

echo ""
echo "🔧 To analyze profiles:"
if [ -f "$RESULTS_DIR/cpu.prof" ]; then
    echo "  go tool pprof $RESULTS_DIR/cpu.prof"
fi
if [ -f "$RESULTS_DIR/mem.prof" ]; then
    echo "  go tool pprof $RESULTS_DIR/mem.prof"
fi
echo ""
echo "📊 To compare results over time:"
echo "  benchstat old_results.txt new_results.txt"
echo ""
echo "🎯 Next steps:"
echo "  1. Run: ./analyze_results.sh (if available)"
echo "  2. Review the full_results.txt for overall performance"
echo "  3. Check any failed benchmarks for issues"
