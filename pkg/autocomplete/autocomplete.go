package autocomplete

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/FelipePn10/kariuki/cmd/terminal"
	input_autocomplete "github.com/JoaoDanielRufino/go-input-autocomplete"
	"github.com/chzyer/readline"
)

type AutoComplete struct {
	completer      *readline.PrefixCompleter
	history        []string
	historyPath    string
	startIndex     int
	size           int
	config         *terminal.TerminalConfig // Terminal configuration
	maxHistorySize int
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
	a.completer = a.buildCompleter()
	return a
}

// Builds the autocompletion tree for readline.
func (a *AutoComplete) buildCompleter() *readline.PrefixCompleter {
	items := []readline.PrefixCompleterInterface{
		readline.PcItem("mode",
			readline.PcItem("vi"),
			readline.PcItem("emacs"),
		),
		readline.PcItem("login"),
		readline.PcItem("say",
			readline.PcItemDynamic(a.listFiles("./"),
				readline.PcItem("with",
					readline.PcItem("following"),
					readline.PcItem("items"),
				),
			),
		),
		readline.PcItem("hello"),
		readline.PcItem("bye"),
		readline.PcItem("setprompt"),
		readline.PcItem("clear"),
		readline.PcItem("exit"),
		readline.PcItem("setpassword"),
		readline.PcItem("help"),
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
		readline.PcItem("sleep"),
	}

	// dynamic history suggestions
	items = append(items, readline.PcItemDynamic(func(line string) []string {
		suggestions := make([]string, 0)
		lineLower := strings.ToLower(line)
		for _, cmd := range a.history {
			if strings.HasPrefix(strings.ToLower(cmd), lineLower) {
				suggestions = append(suggestions, cmd)
			}
		}
		return suggestions
	}))

	// Add allowed commands from terminal configuration, if available.
	if a.config != nil && len(a.config.AllowedCommands) > 0 {
		for _, cmd := range a.config.AllowedCommands {
			if !containsCommand(items, cmd) {
				items = append(items, readline.PcItem(cmd))
			}
		}
	}

	return readline.NewPrefixCompleter(items...)
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

// List files in a directory
func (a *AutoComplete) listFiles(path string) func(string) []string {
	return func(line string) []string {
		resolvedPath := path
		if !filepath.IsAbs(path) {
			if cwd, err := os.Getwd(); err == nil {
				resolvedPath = filepath.Join(cwd, path)
			}
		}
		names := make([]string, 0) // Array to store file names
		files, err := os.ReadDir(resolvedPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading directory %s: %v\n", resolvedPath, err)
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

	// avoid saving consecutive duplicate commands in history
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

func (a *AutoComplete) loadHistoryFromDisk() {
	if a.historyPath == "" {
		return
	}

	// data in bytes
	data, err := os.ReadFile(a.historyPath)
	if err != nil {
		log.Printf("No existing history file found: %v", err)
		return
	}

	// data in bytes -> data in strings
	lines := strings.Split(string(data), "\n")
	var history []string
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed != "" {
			history = append(history, trimmed)
		}
	}
	// Take the last a.maxHistorySize commands
	if len(history) > a.maxHistorySize {
		history = history[len(history)-a.maxHistorySize:]
	}
	a.history = history
	a.size = len(history)
	a.startIndex = 0
}

func (a *AutoComplete) SuggestHistory(input string) []string {
	suggestions := make([]string, 0)
	input = strings.ToLower(strings.TrimSpace(input))
	if input == "" {
		return suggestions
	}

	for _, cmd := range a.history {
		if strings.HasPrefix(strings.ToLower(cmd), input) {
			suggestions = append(suggestions, cmd)
		}
	}
	return suggestions
}

func (a *AutoComplete) Input(prompt string) (string, error) {
	rl, err := readline.NewEx(&readline.Config{
		Prompt:              prompt,
		AutoComplete:        a.completer,
		HistoryFile:         a.config.HistoryFile,
		InterruptPrompt:     "^C",
		EOFPrompt:           "exit",
		HistorySearchFold:   true,
		FuncFilterInputRune: nil,
	})
	if err != nil {
		return "", fmt.Errorf("failed to initialize readline: %w", err)
	}
	defer rl.Close()

	line, err := rl.Readline()
	if err != nil {
		return "", err
	}

	a.AddToHistory(line)
	return line, nil
}

func FallbackInput(prompt string) (string, error) {
	path, err := input_autocomplete.Read(prompt)
	if err != nil {
		return "", fmt.Errorf("failed to read path entry: %w", err)
	}
	return path, nil
}

func InputAutocomplete() {
	path, err := input_autocomplete.Read("Path: ")
	if err != nil {
		panic(err)
	}
	fmt.Println(path)
}
