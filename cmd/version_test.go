package cmd

import (
	"testing"
)

func TestVersionCommandDefinition(t *testing.T) {
	if versionCmd.Use != "version" {
		t.Errorf("expected Use to be 'version', got %s", versionCmd.Use)
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

func TestVersionVariables(t *testing.T) {
	// バージョン変数が適切なデフォルト値を持っていることを確認
	if version != "dev" {
		t.Errorf("expected default version to be 'dev', got %s", version)
	}

	if commit != "none" {
		t.Errorf("expected default commit to be 'none', got %s", commit)
	}

	if date != "unknown" {
		t.Errorf("expected default date to be 'unknown', got %s", date)
	}
}

func TestVersionCommand_Integration(t *testing.T) {
	// versionコマンドがrootコマンドに追加されているかテスト
	found := false
	for _, cmd := range rootCmd.Commands() {
		if cmd.Use == "version" {
			found = true
			break
		}
	}
	
	if !found {
		t.Error("version command should be added to root command")
	}
}