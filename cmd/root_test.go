package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRootCommandDefinition(t *testing.T) {
	assert.Equal(t, "gotodoist", rootCmd.Use, "rootコマンドのUseが期待値と異なります")
	assert.Equal(t, "A CLI tool for Todoist", rootCmd.Short, "rootコマンドのShortが期待値と異なります")
	assert.NotEmpty(t, rootCmd.Long, "Long記述が設定されていません")

	// Long description should contain key information
	expectedTexts := []string{"gotodoist", "command-line", "Todoist", "tasks"}
	for _, text := range expectedTexts {
		assert.True(t, containsText(rootCmd.Long, text), "Long記述に %q が含まれていません", text)
	}
}

func TestRootCommandFlags(t *testing.T) {
	// Check that persistent flags are defined
	verboseFlag := rootCmd.PersistentFlags().Lookup("verbose")
	assert.NotNil(t, verboseFlag, "verboseフラグが定義されていません")
	if verboseFlag != nil {
		assert.Equal(t, "v", verboseFlag.Shorthand, "verboseフラグのショートハンドが期待値と異なります")
		assert.Equal(t, "enable verbose output", verboseFlag.Usage, "verboseフラグのUsageが期待値と異なります")
	}

	debugFlag := rootCmd.PersistentFlags().Lookup("debug")
	assert.NotNil(t, debugFlag, "debugフラグが定義されていません")
	if debugFlag != nil {
		assert.Equal(t, "enable debug mode", debugFlag.Usage, "debugフラグのUsageが期待値と異なります")
	}

	langFlag := rootCmd.PersistentFlags().Lookup("lang")
	assert.NotNil(t, langFlag, "langフラグが定義されていません")
	if langFlag != nil {
		assert.Equal(t, "language preference (en/ja)", langFlag.Usage, "langフラグのUsageが期待値と異なります")
	}
}

func TestRootCommandSubcommands(t *testing.T) {
	// Test that expected subcommands are registered
	commands := rootCmd.Commands()

	expectedCommands := []string{"version", "task", "project"}

	for _, expectedCmd := range expectedCommands {
		found := false
		for _, cmd := range commands {
			if cmd.Use == expectedCmd {
				found = true
				break
			}
		}
		assert.True(t, found, "サブコマンド %q が登録されていません", expectedCmd)
	}
}

// Helper function to check if text contains a substring (case-insensitive)
func containsText(text, substr string) bool {
	// Simple case-insensitive contains check
	textLower := toLower(text)
	substrLower := toLower(substr)
	return contains(textLower, substrLower)
}

// Simple toLowerCase implementation for testing
func toLower(s string) string {
	result := make([]rune, len(s))
	for i, r := range s {
		if r >= 'A' && r <= 'Z' {
			result[i] = r + 32
		} else {
			result[i] = r
		}
	}
	return string(result)
}

// Simple contains implementation for testing
func contains(s, substr string) bool {
	if substr == "" {
		return true
	}
	if len(substr) > len(s) {
		return false
	}
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
