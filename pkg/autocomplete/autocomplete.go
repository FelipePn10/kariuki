package autocomplete

import (
	"container/list"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/FelipePn10/kariuki/cmd/terminal"
	"github.com/chzyer/readline"
	"github.com/sahilm/fuzzy"
)

type LRUCache struct {
	capacity int
	cache    map[string]*list.Element
	list     *list.List
}

func NewLRUCache(capacity int) *LRUCache {
	return &LRUCache{
		capacity: capacity,
		cache:    make(map[string]*list.Element),
		list:     list.New(),
	}
}

func (c *LRUCache) Put(command string) {
	if elem, ok := c.cache[command]; ok {
		c.list.MoveToFront(elem)
		return
	}
	if c.list.Len() >= c.capacity {
		back := c.list.Back()
		if back != nil {
			delete(c.cache, back.Value.(string))
			c.list.Remove(back)
		}
	}
	elem := c.list.PushFront(command)
	c.cache[command] = elem
}

func (c *LRUCache) GetSuggestions(input string, limit int) []string {
	var suggestions []string
	count := 0
	for elem := c.list.Front(); elem != nil && count < limit; elem = elem.Next() {
		cmd := elem.Value.(string)
		if strings.Contains(strings.ToLower(cmd), strings.ToLower(input)) {
			suggestions = append(suggestions, cmd)
			count++
		}
	}
	return suggestions
}

type AutoComplete struct {
	completer      *readline.PrefixCompleter
	history        []string
	historyPath    string
	startIndex     int
	size           int
	config         *terminal.TerminalConfig
	maxHistorySize int
	allCommands    []string
	lruCache       *LRUCache
}

func NewAutocomplete(config *terminal.TerminalConfig) *AutoComplete {
	lruCacheSize := 100
	if config.LRUCacheSize > 0 {
		lruCacheSize = config.LRUCacheSize
	}
	a := &AutoComplete{
		config:         config,
		historyPath:    config.HistoryFile,
		history:        make([]string, 0, config.HistorySize),
		maxHistorySize: config.HistorySize,
		startIndex:     0,
		size:           0,
		lruCache:       NewLRUCache(lruCacheSize),
	}
	a.loadHistoryFromDisk()
	a.allCommands = a.collectAllCommands()
	a.completer = a.buildCompleter()
	return a
}

func (a *AutoComplete) collectAllCommands() []string {
	var commands []string
	hardcoded := []string{
		"mode", "login", "say", "hello", "bye", "setprompt",
		"clear", "exit", "setpassword", "help", "go", "sleep",
	}
	commands = append(commands, hardcoded...)
	if a.config != nil {
		commands = append(commands, a.config.AllowedCommands...)
	}
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
			start := time.Now()
			defer func() {
				log.Printf("Autocomplete took %v", time.Since(start))
			}()

			// Pre-filtering with LRU
			lruSuggestions := a.lruCache.GetSuggestions(line, 5)
			remainingLimit := 15
			candidates := append([]string{}, a.allCommands...)
			candidates = append(candidates, a.history...)
			for _, cmd := range lruSuggestions {
				for i, c := range candidates {
					if c == cmd {
						candidates = append(candidates[:i], candidates[i+1:]...)
						break
					}
				}
			}

			// fuzzy
			matches := fuzzy.Find(line, candidates)
			var suggestions []string
			for _, match := range matches {
				if len(suggestions) >= remainingLimit {
					break
				}
				suggestions = append(suggestions, match.Str)
			}
			suggestions = append(lruSuggestions, suggestions...)
			if len(suggestions) > 10 {
				suggestions = suggestions[:10]
			}

			scoredSuggestions := make([]struct {
				cmd   string
				score int
			}, len(suggestions))
			for i, cmd := range suggestions {
				scoredSuggestions[i] = struct {
					cmd   string
					score int
				}{cmd: cmd, score: 0}
				if elem, ok := a.lruCache.cache[cmd]; ok {
					position := 0
					for e := a.lruCache.list.Front(); e != nil; e = e.Next() {
						if e == elem {
							break
						}
						position++
					}
					C := a.lruCache.list.Len()
					if C > 0 {
						recencyScore := float64(C-position) / float64(C)
						scoredSuggestions[i].score += int(20.0 * recencyScore)
					}
				}
				for _, match := range matches {
					if match.Str == cmd {
						scoredSuggestions[i].score += match.Score
						break
					}
				}
			}

			sort.Slice(scoredSuggestions, func(i, j int) bool {
				return scoredSuggestions[i].score > scoredSuggestions[j].score
			})

			finalSuggestions := make([]string, 0, len(scoredSuggestions))
			for _, s := range scoredSuggestions {
				finalSuggestions = append(finalSuggestions, s.cmd)
			}
			return finalSuggestions
		}),
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
	a.lruCache.Put(command)
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
	if input == "" {
		return []string{}
	}
	if a.history == nil {
		a.loadHistoryFromDisk()
	}
	matches := fuzzy.Find(input, a.history)
	// The slice size will be the same as matches, but it is still empty.
	suggestions := make([]string, 0, len(matches))
	for _, match := range matches {
		if len(suggestions) >= 10 {
			break
		}
		suggestions = append(suggestions, match.Str)
	}
	return suggestions
}
