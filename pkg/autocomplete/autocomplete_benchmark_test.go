package autocomplete

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/FelipePn10/kariuki/cmd/terminal"
)

// Dados de teste para os benchmarks
var (
	testCommands = []string{
		"mode", "login", "say", "hello", "bye", "setprompt",
		"clear", "exit", "setpassword", "help", "go", "sleep",
		"build", "test", "run", "install", "deploy", "start",
		"stop", "restart", "status", "config", "update", "delete",
		"create", "list", "show", "edit", "copy", "move", "remove",
	}

	testHistory = []string{
		"go build main.go",
		"go test ./...",
		"say hello world",
		"mode vi",
		"setprompt 'custom> '",
		"help autocomplete",
		"clear screen",
		"login admin",
		"go install -v",
		"sleep 5",
		"setpassword newpass",
		"bye goodbye",
		"go build -o app",
		"say with following items",
		"mode emacs",
	}
)

// Função auxiliar para criar um AutoComplete de teste
func createTestAutoComplete() *AutoComplete {
	config := &terminal.TerminalConfig{
		HistoryFile:     "", // Não usar arquivo para testes
		HistorySize:     100,
		AllowedCommands: []string{"custom1", "custom2", "custom3"},
	}

	ac := NewAutocomplete(config)

	// Adicionar histórico de teste
	for _, cmd := range testHistory {
		ac.AddToHistory(cmd)
	}

	return ac
}

// Benchmark: Criação do AutoComplete
func BenchmarkNewAutocomplete(b *testing.B) {
	config := &terminal.TerminalConfig{
		HistoryFile:     "",
		HistorySize:     100,
		AllowedCommands: testCommands,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = NewAutocomplete(config)
	}
}

// Benchmark: Coleta de todos os comandos
func BenchmarkCollectAllCommands(b *testing.B) {
	config := &terminal.TerminalConfig{
		HistoryFile:     "",
		HistorySize:     100,
		AllowedCommands: testCommands,
	}
	ac := &AutoComplete{config: config}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ac.collectAllCommands()
	}
}

// Benchmark: Sugestões fuzzy com diferentes tamanhos de entrada
func BenchmarkGetFuzzySuggestions_SmallInput(b *testing.B) {
	input := "go"
	candidates := testCommands

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = getFuzzySuggestions(input, candidates, 5)
	}
}

func BenchmarkGetFuzzySuggestions_MediumInput(b *testing.B) {
	input := "build"
	candidates := testHistory

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = getFuzzySuggestions(input, candidates, 5)
	}
}

func BenchmarkGetFuzzySuggestions_LongInput(b *testing.B) {
	input := "go build main.go"
	candidates := testHistory

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = getFuzzySuggestions(input, candidates, 10)
	}
}

// Benchmark: Sugestões fuzzy com diferentes tamanhos de candidatos
func BenchmarkGetFuzzySuggestions_SmallCandidates(b *testing.B) {
	input := "test"
	candidates := testCommands[:5]

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = getFuzzySuggestions(input, candidates, 5)
	}
}

func BenchmarkGetFuzzySuggestions_LargeCandidates(b *testing.B) {
	input := "test"
	// Criar uma lista maior de candidatos
	largeCandidates := make([]string, 0, 1000)
	for i := 0; i < 100; i++ {
		largeCandidates = append(largeCandidates, testCommands...)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = getFuzzySuggestions(input, largeCandidates, 10)
	}
}

// Benchmark: Adição ao histórico
func BenchmarkAddToHistory(b *testing.B) {
	ac := createTestAutoComplete()
	commands := []string{
		"new command 1",
		"new command 2",
		"new command 3",
		"new command 4",
		"new command 5",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ac.AddToHistory(commands[i%len(commands)])
	}
}

// Benchmark: Adição ao histórico com buffer cheio
func BenchmarkAddToHistory_FullBuffer(b *testing.B) {
	config := &terminal.TerminalConfig{
		HistoryFile: "",
		HistorySize: 10, // Buffer pequeno para forçar rotação
	}
	ac := NewAutocomplete(config)

	// Preencher o buffer
	for i := 0; i < 10; i++ {
		ac.AddToHistory("initial command " + string(rune(i)))
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ac.AddToHistory("rotating command " + string(rune(i)))
	}
}

// Benchmark: Sugestões do histórico
func BenchmarkSuggestHistory(b *testing.B) {
	ac := createTestAutoComplete()
	input := "go"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ac.SuggestHistory(input)
	}
}

// Benchmark: Listagem de arquivos
func BenchmarkListFiles(b *testing.B) {
	ac := createTestAutoComplete()
	// Usar diretório atual
	listFunc := ac.listFiles("./")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = listFunc("")
	}
}

// Benchmark: Salvar histórico em disco (usando arquivo temporário)
func BenchmarkSaveHistoryToDisk(b *testing.B) {
	tempDir := b.TempDir()
	historyFile := filepath.Join(tempDir, "test_history")

	config := &terminal.TerminalConfig{
		HistoryFile: historyFile,
		HistorySize: 100,
	}
	ac := NewAutocomplete(config)

	// Adicionar alguns comandos ao histórico
	for _, cmd := range testHistory {
		ac.AddToHistory(cmd)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ac.SaveHistoryToDisk()
	}
}

// Benchmark: Carregar histórico do disco
func BenchmarkLoadHistoryFromDisk(b *testing.B) {
	tempDir := b.TempDir()
	historyFile := filepath.Join(tempDir, "test_history")

	// Criar arquivo de histórico de teste
	content := strings.Join(testHistory, "\n")
	err := os.WriteFile(historyFile, []byte(content), 0644)
	if err != nil {
		b.Fatal("Falha ao criar arquivo de teste:", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		config := &terminal.TerminalConfig{
			HistoryFile: historyFile,
			HistorySize: 100,
		}
		ac := &AutoComplete{
			config:         config,
			historyPath:    historyFile,
			history:        make([]string, 0, 100),
			maxHistorySize: 100,
		}
		ac.loadHistoryFromDisk()
	}
}

// Benchmark: Construção do completer
func BenchmarkBuildCompleter(b *testing.B) {
	ac := createTestAutoComplete()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ac.buildCompleter()
	}
}

// Benchmark: Fluxo completo de autocomplete
func BenchmarkCompleteFlow(b *testing.B) {
	ac := createTestAutoComplete()
	inputs := []string{"g", "go", "say", "hel", "mod"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		input := inputs[i%len(inputs)]
		// Simular sugestões de comandos e histórico
		_ = getFuzzySuggestions(input, ac.allCommands, 5)
		_ = getFuzzySuggestions(input, ac.history, 5)
	}
}

// Benchmark de memória: Criação do AutoComplete
func BenchmarkMemory_NewAutocomplete(b *testing.B) {
	config := &terminal.TerminalConfig{
		HistoryFile:     "",
		HistorySize:     1000,
		AllowedCommands: testCommands,
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ac := NewAutocomplete(config)
		_ = ac
	}
}

// Benchmark de memória: Sugestões fuzzy
func BenchmarkMemory_FuzzySuggestions(b *testing.B) {
	input := "test"
	candidates := testHistory

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		suggestions := getFuzzySuggestions(input, candidates, 10)
		_ = suggestions
	}
}
