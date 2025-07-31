package cmd

import (
	"testing"
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
			if result != tt.expected {
				t.Errorf("maskToken(%q) = %q, want %q", tt.token, result, tt.expected)
			}
		})
	}
}

func TestConfigCommandDefinition(t *testing.T) {
	// configコマンドが正しく定義されているかテスト
	if configCmd.Use != configCmdName {
		t.Errorf("expected Use to be %q, got %s", configCmdName, configCmd.Use)
	}

	if configCmd.Short != "Manage gotodoist configuration" {
		t.Errorf("expected Short to be 'Manage gotodoist configuration', got %s", configCmd.Short)
	}

	if configCmd.Run == nil {
		t.Error("expected Run function to be defined")
	}

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
		if !found {
			t.Errorf("expected subcommand %q to be registered", expectedCmd)
		}
	}
}

func TestConfigSubcommandDefinitions(t *testing.T) {
	// showコマンドのテスト
	if configShowCmd.Use != "show" {
		t.Errorf("expected configShowCmd.Use to be 'show', got %s", configShowCmd.Use)
	}
	if configShowCmd.Run == nil {
		t.Error("expected configShowCmd.Run to be defined")
	}

	// pathコマンドのテスト
	if configPathCmd.Use != "path" {
		t.Errorf("expected configPathCmd.Use to be 'path', got %s", configPathCmd.Use)
	}
	if configPathCmd.Run == nil {
		t.Error("expected configPathCmd.Run to be defined")
	}

	// initコマンドのテスト
	if configInitCmd.Use != "init" {
		t.Errorf("expected configInitCmd.Use to be 'init', got %s", configInitCmd.Use)
	}
	if configInitCmd.Run == nil {
		t.Error("expected configInitCmd.Run to be defined")
	}
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

	if !found {
		t.Error("config command should be added to root command")
	}
}
