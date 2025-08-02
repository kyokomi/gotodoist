package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

const configCmdName = "config"

func TestMaskToken(t *testing.T) {
	tests := []struct {
		name     string
		token    string
		expected string
	}{
		{
			name:     "empty token",
			token:    "",
			expected: "(not set)",
		},
		{
			name:     "short token",
			token:    "abc",
			expected: "***",
		},
		{
			name:     "normal token",
			token:    "abcd1234567890efgh",
			expected: "abcd...efgh",
		},
		{
			name:     "exactly 8 chars",
			token:    "12345678",
			expected: "1234...5678",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := maskToken(tt.token)
			assert.Equal(t, tt.expected, result, "maskToken結果が期待値と異なります")
		})
	}
}

func TestConfigCommandDefinition(t *testing.T) {
	// configコマンドが正しく定義されているかテスト
	assert.Equal(t, configCmdName, configCmd.Use, "configコマンドのUseが期待値と異なります")
	assert.Equal(t, "Manage gotodoist configuration", configCmd.Short, "configコマンドのShortが期待値と異なります")
	assert.True(t, configCmd.Run != nil || configCmd.RunE != nil, "configコマンドのRunまたはRunE関数が定義されていません")

	// サブコマンドがあることを確認
	commands := configCmd.Commands()
	expectedSubcommands := []string{"show", "path", "init"}

	for _, expectedCmd := range expectedSubcommands {
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

func TestConfigSubcommandDefinitions(t *testing.T) {
	// showコマンドのテスト
	assert.Equal(t, "show", configShowCmd.Use, "configShowCmdのUseが期待値と異なります")
	assert.True(t, configShowCmd.Run != nil || configShowCmd.RunE != nil, "configShowCmdのRunまたはRunE関数が定義されていません")

	// pathコマンドのテスト
	assert.Equal(t, "path", configPathCmd.Use, "configPathCmdのUseが期待値と異なります")
	assert.True(t, configPathCmd.Run != nil || configPathCmd.RunE != nil, "configPathCmdのRunまたはRunE関数が定義されていません")

	// initコマンドのテスト
	assert.Equal(t, "init", configInitCmd.Use, "configInitCmdのUseが期待値と異なります")
	assert.True(t, configInitCmd.Run != nil || configInitCmd.RunE != nil, "configInitCmdのRunまたはRunE関数が定義されていません")
}

func TestConfigCommandIntegration(t *testing.T) {
	// configコマンドがrootコマンドに追加されているかテスト
	found := false
	for _, cmd := range rootCmd.Commands() {
		if cmd.Use == configCmdName {
			found = true
			break
		}
	}

	assert.True(t, found, "configコマンドがrootコマンドに追加されていません")
}
