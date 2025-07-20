package terminal

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/go-viper/mapstructure/v2"
	"github.com/spf13/viper"
)

type TerminalConfig struct {
	// Section: Appearance TerminalConfig
	Prompt         string `mapstructure:"prompt"`
	BgColor        string `mapstructure:"bg_color"`
	TextColor      string `mapstructure:"text_color"`
	CursorStyle    string `mapstructure:"cursor_style"`
	CursorBlink    bool   `mapstructure:"cursor_blink"`
	WelcomeMessage string `mapstructure:"welcome_message"`
	Font           string `mapstructure:"font"`

	// Section: Behavior and History
	HistorySize     int           `mapstructure:"history_size"`
	HistoryFile     string        `mapstructure:"history_file"`
	TypeAhead       bool          `mapstructure:"type_ahead"`
	AutoSuggest     bool          `mapstructure:"auto_suggest"`
	InactivityClose time.Duration `mapstructure:"inactivity_close"`

	// Section: Security and Access
	MaxSessionTime  time.Duration `mapstructure:"max_session_time"`
	AllowedCommands []string      `mapstructure:"allowed_commands"`
	BlockedCommands []string      `mapstructure:"blocked_commands"`
	EnableLogging   bool          `mapstructure:"enable_logging"` // Log executed commands

	// Section: Config PTY
	Rows         int    `mapstructure:"rows"`
	Cols         int    `mapstructure:"cols"`
	ScrollBuffer int    `mapstructure:"scroll_buffer"`
	Encoding     string `mapstructure:"encoding"` // UTF-8 etc..
	BellSound    string `mapstructure:"bell_sound"`
	EnableMouse  bool   `mapstructure:"enable_mouse"` // Mouse event support
}

var (
	configInstance *TerminalConfig // Instance of TerminalConfig
	configMutex    sync.RWMutex
	configErr      error
	configOnce     sync.Once // Ensures a single boot
)

// Loads terminal settings from a configuration file.
// Loads terminal settings from a configuration file.
func LoadConfig(configPath, kariuki string) (*TerminalConfig, error) {
	configOnce.Do(func() {
		configInstance = &TerminalConfig{}
		v := viper.New()

		setDefaultConfig(v)
		v.SetConfigType("yaml")

		// Explicit configuration file (highest priority)
		if configPath != "" {
			v.SetConfigFile(configPath)
		} else {
			v.SetConfigName("pty-config")
			// Path Search - Current directory
			v.AddConfigPath(".")
			// e.g. ~/.config/<kariuki>
			if userConfigDir, err := os.UserConfigDir(); err == nil {
				appConfigDir := filepath.Join(userConfigDir, kariuki)
				v.AddConfigPath(appConfigDir)
			}
			v.AddConfigPath("/etc/" + kariuki) // Global config
		}

		// Config System
		v.SetEnvPrefix("PTY") // PTY_ prefix for all variables
		v.AutomaticEnv()      // Bind all environment variables automatically

		v.BindEnv("bg_color", "PTY_BACKGROUND_COLOR")
		v.BindEnv("text_color", "PTY_TEXT_COLOR")
		v.BindEnv("inactivity_close", "PTY_SESSION_TIMEOUT")

		// Load and error handling
		if err := v.ReadInConfig(); err != nil {
			// Check if the error is just "file not found"
			if _, ok := err.(viper.ConfigFileNotFoundError); ok {
				fmt.Println("Config: Using default settings (file not found)")
			} else {
				configErr = fmt.Errorf("error reading configuration file: %w", err)
				return
			}
		}

		// Create a custom decoder
		decoderConfig := &mapstructure.DecoderConfig{
			Result: configInstance,
			DecodeHook: mapstructure.ComposeDecodeHookFunc(
				mapstructure.StringToTimeDurationHookFunc(),
				mapstructure.StringToSliceHookFunc(","),
			),
			// We'll handle unused keys manually
			ErrorUnused: false,
		}

		decoder, err := mapstructure.NewDecoder(decoderConfig)
		if err != nil {
			configErr = fmt.Errorf("failed to create config decoder: %w", err)
			return
		}

		// Get all settings and handle legacy keys
		settings := v.AllSettings()
		handleLegacyKeys(settings)

		// Decode using custom decoder
		if err := decoder.Decode(settings); err != nil {
			configErr = fmt.Errorf("failed to decode config: %w", err)
			return
		}

		// Validate configuration keys
		if err := validateConfigKeys(v, settings); err != nil {
			configErr = err
			return
		}

		// Process settings
		configMutex.Lock()
		configInstance.postProcessConfig()
		configMutex.Unlock()

		// Start monitoring file changes
		if configPath != "" {
			go watchConfigFile(configPath, kariuki)
		}
	})
	return configInstance, configErr
}

// setDefaultConfig sets the default values for all configurations
func setDefaultConfig(v *viper.Viper) {
	v.SetDefault("prompt", "> ")
	v.SetDefault("bg_color", "black")
	v.SetDefault("text_color", "white")
	v.SetDefault("cursor_style", "block")
	v.SetDefault("cursor_blink", true)
	v.SetDefault("welcome_message", "Welcome to the Kariuki!")
	v.SetDefault("font", "Monospace")

	v.SetDefault("history_size", 1000)
	v.SetDefault("history_file", ".pty_history")
	v.SetDefault("type_ahead", true)
	v.SetDefault("auto_suggest", true)
	v.SetDefault("inactivity_close", time.Hour)

	v.SetDefault("max_session_time", 8*time.Hour)
	v.SetDefault("allowed_commands", []string{})
	v.SetDefault("blocked_commands", []string{"rm -rf /", "dd if=/dev/random"})

	v.SetDefault("enable_logging", false)

	v.SetDefault("rows", 24)
	v.SetDefault("cols", 80)
	v.SetDefault("scroll_buffer", 1000)
	v.SetDefault("encoding", "UTF-8")
	v.SetDefault("bell_sound", "system")
	v.SetDefault("enable_mouse", true)
}

// watchConfigFile monitors configuration file changes
func watchConfigFile(configPath, kariuki string) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		fmt.Printf("Error creating configuration watcher: %v\n", err)
		return
	}
	defer watcher.Close()

	dir := filepath.Dir(configPath)
	if err := watcher.Add(dir); err != nil {
		fmt.Printf("Error adding directory to watcher: %v\n", err)
		return
	}

	fmt.Printf("Watching directory: %s\n", configPath)

	// Process watcher events.
	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return
			}
			if event.Name == configPath && event.Op&fsnotify.Write == fsnotify.Write {
				fmt.Println("Configuration file changed, reloading...")
				if err := ReloadConfig(configPath, kariuki); err != nil {
					fmt.Printf("Error reloading configuration: %v\n", err)
				} else {
					fmt.Println("Configuration reloaded successfully")
				}
			}
		// If an error occurs in the watcher
		case err, ok := <-watcher.Errors:
			if !ok {
				// Close channel and goruntine stop
				return
			}
			fmt.Printf("Error watching configuration file: %v\n", err)
		}
	}
}

//----------------------------------------------------------------------------//
//Deserialise to struct :
//
// 			configErr = v.UnmarshalExact(&configInstance, viper.DecoderConfigOption(
// 				func(c *mapstructure.DecoderConfig) {
// 					c.DecodeHook = mapstructure.ComposeDecodeHookFunc(
// 						mapstructure.StringToTimeDurationHookFunc(),
// 						func(f reflect.Kind, t reflect.Kind, data interface{}) (interface{}, error) {
// 							if f != reflect.String || t != reflect.Slice {
// 								return data, nil
// 							}
// 							str := data.(string)
// 							if str == "" {
// 								return []string{}, nil
// 							}
// 							return strings.Split(str, ","), nil
// 						},
// 					)
// 				},
// 			))

// 			if configErr == nil {
// 				configInstance.postProcessConfig()
// 			}
// 		}
// 	})
// 	return configInstance, configErr
// }
//
//----------------------------------------------------------------------------//

// Function and Hooks
// String -> Time.Duration
func StringToTimeDurationHookFunc() mapstructure.DecodeHookFunc {
	return mapstructure.StringToTimeDurationHookFunc()
}

// String delimiters -> Slices
func StringToSliceHookFunc(delimiter string) mapstructure.DecodeHookFunc {
	return func(f reflect.Kind, t reflect.Kind, data interface{}) (interface{}, error) {
		// Check if the source is a string and the destination is a slice.
		if f != reflect.String || t != reflect.Slice {
			return data, nil
		}
		// Convert the data to a string.
		str, ok := data.(string)
		if !ok || str == "" {
			// If not a string or empty, return an empty slice.
			return []string{}, nil
		}
		// Split the string by the delimiter and return the slice.
		return strings.Split(str, delimiter), nil
	}
}

func (c *TerminalConfig) postProcessConfig() {
	if c.HistorySize < 100 {
		c.HistorySize = 100
	}

	if c.Rows < 10 {
		c.Rows = 10
	}
	if c.Cols < 40 {
		c.Cols = 40
	}

	// Ex.: ".kariuki_history" -> "/home/user/.kariuki_history".
	if c.HistoryFile != "" && !filepath.IsAbs(c.HistoryFile) {
		if home, err := os.UserHomeDir(); err == nil {
			c.HistoryFile = filepath.Join(home, c.HistoryFile)
		}
	}

	c.BgColor = strings.ToLower(c.BgColor)
	c.TextColor = strings.ToLower(c.TextColor)

	if len(c.BlockedCommands) == 0 {
		c.BlockedCommands = []string{
			"rm -rf /",
			"mkfs",
			"dd if=/dev/random",
		}
	}

	// "block", "underline" or "bar" are allowed.
	validCursors := []string{"block", "underline", "bar"}
	found := false
	for _, vc := range validCursors {
		if c.CursorStyle == vc {
			found = true
			break
		}
	}
	if !found {
		c.CursorStyle = "block"
	}
}

func GetConfig() (*TerminalConfig, error) {
	configMutex.RLock()
	defer configMutex.RUnlock()

	if configInstance == nil {
		return nil, fmt.Errorf("configuration not initialized. Call LoadConfig first")
	}
	return configInstance, configErr
}

func ReloadConfig(configPath, kariuki string) error {
	configMutex.Lock()
	defer configMutex.Unlock()

	configOnce = sync.Once{}
	configInstance = nil

	_, err := LoadConfig(configPath, kariuki)
	return err
}

// Handle legacy keys by mapping them to new keys
func handleLegacyKeys(settings map[string]interface{}) {
	// Handle enconding -> encoding
	if val, ok := settings["enconding"]; ok {
		if _, exists := settings["encoding"]; !exists {
			settings["encoding"] = val
		}
		delete(settings, "enconding")
	}

	// Handle welcome_msg -> welcome_message
	if val, ok := settings["welcome_msg"]; ok {
		if _, exists := settings["welcome_message"]; !exists {
			settings["welcome_message"] = val
		}
		delete(settings, "welcome_msg")
	}
}

// Validate configuration keys against struct tags
func validateConfigKeys(v *viper.Viper, settings map[string]interface{}) error {
	validKeys := getValidKeysForStruct(reflect.TypeOf(TerminalConfig{}))

	var invalidKeys []string
	for key := range settings {
		if _, valid := validKeys[key]; !valid && key != "" {
			invalidKeys = append(invalidKeys, key)
		}
	}

	if len(invalidKeys) > 0 {
		return fmt.Errorf("decoding failed due to the following error(s):\n\n'%s' has invalid keys: %s",
			v.ConfigFileUsed(), strings.Join(invalidKeys, ", "))
	}
	return nil
}

// Get valid mapstructure keys from struct tags
func getValidKeysForStruct(t reflect.Type) map[string]bool {
	validKeys := make(map[string]bool)

	// Handle struct and pointer types
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	if t.Kind() != reflect.Struct {
		return validKeys
	}

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		tag := field.Tag.Get("mapstructure")

		// Skip fields without mapstructure tag
		if tag == "" || tag == "-" {
			continue
		}

		// Handle comma-separated options (e.g. "field,omitempty")
		if commaIdx := strings.Index(tag, ","); commaIdx != -1 {
			tag = tag[:commaIdx]
		}

		validKeys[tag] = true
	}

	return validKeys
}
