package cmd

import (
	"testing"
)

func TestRootCommandDefinition(t *testing.T) {
	if rootCmd.Use != "gotodoist" {
		t.Errorf("expected Use to be 'gotodoist', got %s", rootCmd.Use)
	}

	if rootCmd.Short != "A CLI tool for Todoist" {
		t.Errorf("expected Short to be 'A CLI tool for Todoist', got %s", rootCmd.Short)
	}

	if rootCmd.Long == "" {
		t.Error("expected Long description to be set")
	}

	// Long description should contain key information
	expectedTexts := []string{"gotodoist", "command-line", "Todoist", "tasks"}
	for _, text := range expectedTexts {
		if !containsText(rootCmd.Long, text) {
			t.Errorf("expected Long description to contain %q", text)
		}
	}
}

func TestRootCommandFlags(t *testing.T) {
	// Check that persistent flags are defined
	verboseFlag := rootCmd.PersistentFlags().Lookup("verbose")
	if verboseFlag == nil {
		t.Error("expected verbose flag to be defined")
	} else {
		if verboseFlag.Shorthand != "v" {
			t.Errorf("expected verbose flag shorthand to be 'v', got %s", verboseFlag.Shorthand)
		}
		if verboseFlag.Usage != "enable verbose output" {
			t.Errorf("expected verbose flag usage to be 'enable verbose output', got %s", verboseFlag.Usage)
		}
	}

	debugFlag := rootCmd.PersistentFlags().Lookup("debug")
	if debugFlag == nil {
		t.Error("expected debug flag to be defined")
	} else if debugFlag.Usage != "enable debug mode" {
		t.Errorf("expected debug flag usage to be 'enable debug mode', got %s", debugFlag.Usage)
	}

	langFlag := rootCmd.PersistentFlags().Lookup("lang")
	if langFlag == nil {
		t.Error("expected lang flag to be defined")
	} else if langFlag.Usage != "language preference (en/ja)" {
		t.Errorf("expected lang flag usage to be 'language preference (en/ja)', got %s", langFlag.Usage)
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
		if !found {
			t.Errorf("expected subcommand %q to be registered", expectedCmd)
		}
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
