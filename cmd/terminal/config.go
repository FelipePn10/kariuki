package terminal

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"sync"
	"time"

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
	configOnce     sync.Once       // Ensures a single boot
	configErr      error
)

// Loads terminal settings from a configuration file.
func LoadConfig(configPath, kariuki string) (*TerminalConfig, error) {
	configOnce.Do(func() {
		v := viper.New()

		v.SetDefault("prompt", "> ")
		v.SetDefault("bg_color", "black")
		v.SetDefault("text_color", "white")
		v.SetDefault("cursor_style", "block")
		v.SetDefault("cursor_blink", true)
		v.SetDefault("welcome_msg", "Welcome to the Kariuki!")
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
		v.SetDefault("enconding", "UTF-8")
		v.SetDefault("bell_sound", "system")
		v.SetDefault("enable_mouse", true)

		// Fonts - Config -
		// Explicit configuration file (highest priority)
		if configPath != "" {
			v.SetConfigFile(configPath)
		} else {
			v.SetConfigName("pty-config")
			// Path Search - Directory current
			v.AddConfigPath(".")
			// es: ~/.config/<kariuki>
			if userConfigDir, err := os.UserConfigDir(); err == nil {
				appConfigDir := filepath.Join(userConfigDir, kariuki)
				v.AddConfigPath(appConfigDir)
			}
			// Config System
			v.AddConfigPath("/etc/" + kariuki)

			v.SetEnvPrefix("PTY") // PTY_ prefix for all variables
			v.AutomaticEnv()      // Bind automatically all environment variables

			v.BindEnv("bg_color", "PTY_BACKGROUND_COLOR")
			v.BindEnv("text_color", "PTY_TEXT_COLOR")
			v.BindEnv("inactivity_close", "PTY_SESSION_TIMEOUT")

			// Load and error handling
			if err := v.ReadInConfig(); err != nil {
				switch err.(type) {
				case viper.ConfigFileNotFoundError:
					fmt.Printf("Config: Using defaults (file not found)\n")
				default:
					configErr = fmt.Errorf("Error config file: %w", err)
					return
				}
			}

			// Deserialise to struct
			configErr = v.UnmarshalExact(&configInstance,
				viper.DecodeHook(
					StringToTimeDurationHookFunc(),
					StringToSliceHookFunc(","), // String -> Slices
				),
			)

			if configErr == nil {
				configInstance.postProcessConfig()
			}
		}
	})
	return configInstance, configErr
}

// Function and Hooks
// String -> Time.Duration
func StringToTimeDurationHookFunc() viper.DecoderConfigOption {
	return func(c *mapstructure.DecoderConfig) {
		c.DecodeHook = mapstructure.ComposeDecodeHookFunc(
			c.DecodeHook,
			mapstructure.StringToTimeDurationHookFunc(),
		)
	}
}

// String delimiters -> Slices
func StringToSliceHookFunc(delimiter string) viper.DecoderConfigOption {
	return func(c *mapstructure.DecoderConfig) {
		c.DecodeHook = mapstructure.ComposeDecodeHookFunc(
			c.DecodeHook,
			func(f reflect.Kind, t reflect.Kind, data interface{}) (interface{}, error) {
				if f != reflect.String || t != reflect.Slice {
					return data, nil
				}
				str := data.(string)
				if str == "" {
					return []string{}, nil
				}
				return strings.Split(str, delimiter), nil
			},
		)
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

	// Expand relative paths for history
	if c.HistoryFile != "" && !filepath.IsAbs(c.HistoryFile) {
		if home, err := os.UserHomeDir(); err == nil {
			c.HistoryFile = filepath.Join(home, c.HistoryFile)
		}
	}

	// Normalize colors to lower-case
	c.BgColor = strings.ToLower(c.BgColor)
	c.TextColor = strings.ToLower(c.TextColor)

	// Config Security: Blocked AllowedCommands
	if len(c.BlockedCommands) == 0 {
		c.BlockedCommands = []string{"rm -rf", "mkfs", "dd if="}
	}
}

// GetConfig returns the loaded configuration instance
// (Concurrency-safe after initialization)
func GetConfig() (*TerminalConfig, error) {
	if configInstance == nil {
		panic("Config instance not initialized. Call InitConfig() first.")
	}
	return configInstance, nil
}

func ReloadConfig(configPath, kariuki string) error {
	configOnce = sync.Once{} // Reset singleton
	_, err := LoadConfig(configPath, kariuki)
	return err
}
