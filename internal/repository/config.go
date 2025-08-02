package repository

import (
	"os"
	"path/filepath"
)

// Config はローカルストレージの設定
type Config struct {
	Enabled            bool   `yaml:"enabled" mapstructure:"enabled"`
	DatabasePath       string `yaml:"database_path" mapstructure:"database_path"`
	InitialSyncOnStart bool   `yaml:"initial_sync_on_startup" mapstructure:"initial_sync_on_startup"`
}

// DefaultConfig はデフォルトのローカルストレージ設定を返す
func DefaultConfig() *Config {
	return &Config{
		Enabled:            true,
		DatabasePath:       getDefaultDatabasePath(),
		InitialSyncOnStart: true,
	}
}

// getDefaultDatabasePath はデフォルトのデータベースパスを返す
func getDefaultDatabasePath() string {
	// XDG Base Directory Specification に従う
	dataHome := os.Getenv("XDG_DATA_HOME")
	if dataHome == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return "./gotodoist.db" // フォールバック
		}
		dataHome = filepath.Join(homeDir, ".local", "share")
	}

	return filepath.Join(dataHome, "gotodoist", "data.db")
}
