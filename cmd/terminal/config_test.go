package terminal_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/FelipePn10/kariuki/cmd/terminal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadConfig(t *testing.T) {
	t.Run("Default values", func(t *testing.T) {
		cfg, err := terminal.LoadConfig("", "testapp")
		require.NoError(t, err)

		assert.Equal(t, "> ", cfg.Prompt)
		assert.Equal(t, "black", cfg.BgColor)
		assert.Equal(t, 1000, cfg.HistorySize)
		assert.Equal(t, []string{"rm -rf /", "dd if=/dev/random"}, cfg.BlockedCommands)
	})

	t.Run("File loading", func(t *testing.T) {
		dir := t.TempDir()
		cfgPath := filepath.Join(dir, "config.yaml")

		data := `
prompt: "TEST $ "
bg_color: navy
history_size: 500
allowed_commands: ["ls", "pwd"]
`
		require.NoError(t, os.WriteFile(cfgPath, []byte(data), 0644))

		cfg, err := terminal.LoadConfig(cfgPath, "testapp")
		require.NoError(t, err)

		assert.Equal(t, "TEST $ ", cfg.Prompt)
		assert.Equal(t, "navy", cfg.BgColor)
		assert.Equal(t, 500, cfg.HistorySize)
		assert.Equal(t, []string{"ls", "pwd"}, cfg.AllowedCommands)
	})

	t.Run("Environment override", func(t *testing.T) {
		t.Setenv("PTY_BACKGROUND_COLOR", "green")
		t.Setenv("PTY_TEXT_COLOR", "yellow")

		cfg, err := terminal.LoadConfig("", "testapp")
		require.NoError(t, err)

		assert.Equal(t, "green", cfg.BgColor)
		assert.Equal(t, "yellow", cfg.TextColor)
	})

	t.Run("Reload config", func(t *testing.T) {
		dir := t.TempDir()
		cfgPath := filepath.Join(dir, "reload.yaml")

		// Config inicial
		require.NoError(t, os.WriteFile(cfgPath, []byte("prompt: \"V1 $\""), 0644))

		cfg, err := terminal.LoadConfig(cfgPath, "testapp")
		require.NoError(t, err)
		assert.Equal(t, "V1 $", cfg.Prompt)

		// Atualiza arquivo
		require.NoError(t, os.WriteFile(cfgPath, []byte("prompt: \"V2 $\""), 0644))

		// Força reload
		err = terminal.ReloadConfig(cfgPath, "testapp")
		require.NoError(t, err)

		cfg, _ = terminal.GetConfig()
		assert.Equal(t, "V2 $", cfg.Prompt)
	})
}
