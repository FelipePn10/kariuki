package autocomplete

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

	fuzzy "github.com/paul-mannino/go-fuzzywuzzy"

	"github.com/FelipePn10/kariuki/cmd/terminal"
	"github.com/chzyer/readline"
)

type AutoComplete struct {
	completer      *readline.PrefixCompleter
	history        []string
	historyPath    string
	startIndex     int
	size           int
	config         *terminal.TerminalConfig
	maxHistorySize int
	allCommands    []string
}

func NewAutocomplete(config *terminal.TerminalConfig) *AutoComplete {
	a := &AutoComplete{
		config:         config,
		historyPath:    config.HistoryFile,
		history:        make([]string, 0, config.HistorySize),
		maxHistorySize: config.HistorySize,
		startIndex:     0,
		size:           0,
	}
	a.loadHistoryFromDisk()
	a.allCommands = a.collectAllCommands()
	a.completer = a.buildCompleter()
	return a
}

func (a *AutoComplete) collectAllCommands() []string {
	// Coletar todos os nomes de comandos
	var commands []string
	// Comandos predefinidos
	hardcoded := []string{
		"mode", "login", "say", "hello", "bye", "setprompt",
		"clear", "exit", "setpassword", "help", "go", "sleep",
	}
	commands = append(commands, hardcoded...)
	// Adicionar comandos permitidos da configuração
	if a.config != nil {
		commands = append(commands, a.config.AllowedCommands...)
	}
	// Remover duplicatas
	uniqueCommands := make(map[string]struct{})
	for _, cmd := range commands {
		uniqueCommands[cmd] = struct{}{}
	}
	var allCommands []string
	for cmd := range uniqueCommands {
		allCommands = append(allCommands, cmd)
	}
	return allCommands
}

func (a *AutoComplete) buildCompleter() *readline.PrefixCompleter {
	return readline.NewPrefixCompleter(
		readline.PcItemDynamic(func(line string) []string {
			// Obter sugestões fuzzy de comandos e histórico
			commandSuggestions := getFuzzySuggestions(line, a.allCommands, 5)
			historySuggestions := getFuzzySuggestions(line, a.history, 5)
			// Combinar sugestões
			suggestions := append(commandSuggestions, historySuggestions...)
			// Limitar a 10 sugestões
			if len(suggestions) > 10 {
				suggestions = suggestions[:10]
			}
			return suggestions
		}),
		// Manter sub-comandos específicos, se necessário
		readline.PcItem("mode",
			readline.PcItem("vi"),
			readline.PcItem("emacs"),
		),
		readline.PcItem("say",
			readline.PcItemDynamic(a.listFiles("./"),
				readline.PcItem("with",
					readline.PcItem("following"),
					readline.PcItem("items"),
				),
			),
		),
		readline.PcItem("go",
			readline.PcItem("build",
				readline.PcItem("-o"),
				readline.PcItem("-v"),
			),
			readline.PcItem("install",
				readline.PcItem("-v"),
				readline.PcItem("-vv"),
				readline.PcItem("-vvv"),
			),
			readline.PcItem("test"),
		),
	)
}

func getFuzzySuggestions(input string, candidates []string, limit int) []string {
	if len(candidates) == 0 || input == "" {
		return []string{}
	}
	results, err := fuzzy.Extract(input, candidates, limit)
	if err != nil {
		// Handle error, e.g., log it or return an empty list
		return []string{}
	}
	// Ordenar por pontuação (maior para menor)
	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})
	// Extrair as correspondências
	suggestions := make([]string, len(results))
	for i, r := range results {
		suggestions[i] = r.Match
	}
	return suggestions
}

func containsCommand(items []readline.PrefixCompleterInterface, cmd string) bool {
	for _, item := range items {
		if pc, ok := item.(*readline.PrefixCompleter); ok {
			if string(pc.Name) == cmd {
				return true
			}
		}
	}
	return false
}

func (a *AutoComplete) listFiles(path string) func(string) []string {
	return func(line string) []string {
		resolvedPath := path
		if !filepath.IsAbs(path) {
			if cwd, err := os.Getwd(); err == nil {
				resolvedPath = filepath.Join(cwd, path)
			}
		}
		names := make([]string, 0)
		files, err := os.ReadDir(resolvedPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Erro ao ler diretório %s: %v\n", resolvedPath, err)
			return names
		}

		for _, f := range files {
			name := f.Name()
			if f.IsDir() {
				name += "/"
			}
			names = append(names, name)
		}
		return names
	}
}

func (a *AutoComplete) AddToHistory(command string) {
	command = strings.TrimSpace(command)
	if command == "" {
		return
	}

	// Evitar salvar comandos duplicados consecutivos
	if a.size > 0 {
		lastIndex := (a.startIndex + a.size - 1) % len(a.history)
		if a.history[lastIndex] == command {
			return
		}
	}

	if a.size < a.maxHistorySize {
		a.history = append(a.history, command)
		a.size++
	} else {
		a.history[a.startIndex] = command
		a.startIndex = (a.startIndex + 1) % a.maxHistorySize
	}
}

func (a *AutoComplete) SaveHistoryToDisk() error {
	if a.historyPath == "" {
		return nil
	}
	var orderedHistory []string
	if a.size < a.maxHistorySize {
		orderedHistory = a.history
	} else {
		orderedHistory = append(a.history[a.startIndex:], a.history[0:a.startIndex]...)
	}
	content := strings.Join(orderedHistory, "\n")
	err := os.WriteFile(a.historyPath, []byte(content), 0644)
	if err != nil {
		log.Printf("Falha ao salvar histórico em %s: %v", a.historyPath, err)
	}
	return err
}

func (a *AutoComplete) loadHistoryFromDisk() {
	if a.historyPath == "" {
		return
	}

	data, err := os.ReadFile(a.historyPath)
	if err != nil {
		log.Printf("Nenhum arquivo de histórico encontrado: %v", err)
		return
	}

	lines := strings.Split(string(data), "\n")
	var history []string
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed != "" {
			history = append(history, trimmed)
		}
	}
	if len(history) > a.maxHistorySize {
		history = history[len(history)-a.maxHistorySize:]
	}
	a.history = history
	a.size = len(history)
	a.startIndex = 0
}

func (a *AutoComplete) SuggestHistory(input string) []string {
	return getFuzzySuggestions(input, a.history, 10)
}
