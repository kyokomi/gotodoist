package cmd

import (
	"testing"
)

const cmdVersion = "version"

func TestVersionCommandDefinition(t *testing.T) {
	if versionCmd.Use != cmdVersion {
		t.Errorf("expected Use to be '%s', got %s", cmdVersion, versionCmd.Use)
	}

	if versionCmd.Short != "Print version information" {
		t.Errorf("expected Short to be 'Print version information', got %s", versionCmd.Short)
	}

	if versionCmd.Run == nil {
		t.Error("expected Run function to be defined")
	}

	// ArgsやFlagsがないことを確認
	if versionCmd.Args != nil {
		t.Error("version command should not require arguments")
	}
}

func TestVersionCommand_Integration(t *testing.T) {
	// versionコマンドがrootコマンドに追加されているかテスト
	found := false
	for _, cmd := range rootCmd.Commands() {
		if cmd.Use == cmdVersion {
			found = true
			break
		}
	}

	if !found {
		t.Error("version command should be added to root command")
	}
}
