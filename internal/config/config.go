package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/kyokomi/gotodoist/internal/api"
)

// Config はアプリケーション設定を管理する
type Config struct {
	APIToken string `json:"api_token"`
	BaseURL  string `json:"base_url,omitempty"`
}

// DefaultConfig はデフォルト設定を返す
func DefaultConfig() *Config {
	return &Config{
		BaseURL: "https://api.todoist.com/api/v1",
	}
}

// LoadConfig は設定を読み込む（環境変数優先）
func LoadConfig() (*Config, error) {
	config := DefaultConfig()

	// 環境変数からAPIトークンを取得
	if token := os.Getenv("TODOIST_API_TOKEN"); token != "" {
		config.APIToken = token
	}

	// 設定ファイルからの読み込みは将来的に実装
	// TODO: ~/.gotodoist/config.json からの読み込み

	if config.APIToken == "" {
		return nil, fmt.Errorf("API token is required. Set TODOIST_API_TOKEN environment variable")
	}

	return config, nil
}

// NewAPIClient は設定からAPIクライアントを作成する
func (c *Config) NewAPIClient() (*api.Client, error) {
	client, err := api.NewClient(c.APIToken)
	if err != nil {
		return nil, fmt.Errorf("failed to create API client: %w", err)
	}

	if c.BaseURL != "" {
		if err := client.SetBaseURL(c.BaseURL); err != nil {
			return nil, fmt.Errorf("failed to set base URL: %w", err)
		}
	}

	return client, nil
}

// GetConfigDir は設定ディレクトリのパスを返す
func GetConfigDir() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}
	return filepath.Join(homeDir, ".gotodoist"), nil
}