// Package config provides configuration management for gotodoist CLI.
package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"

	"github.com/kyokomi/gotodoist/internal/api"
)

const (
	// DefaultLanguage はデフォルト言語設定
	DefaultLanguage = "en"
	// ConfigDirPerm は設定ディレクトリのパーミッション
	ConfigDirPerm = 0750
	// ConfigFilePerm は設定ファイルのパーミッション
	ConfigFilePerm = 0600
)

// Config はアプリケーション設定を管理する
type Config struct {
	APIToken string `yaml:"api_token" mapstructure:"api_token"`
	BaseURL  string `yaml:"base_url,omitempty" mapstructure:"base_url"`
	Language string `yaml:"language,omitempty" mapstructure:"language"`
}

// DefaultConfig はデフォルト設定を返す
func DefaultConfig() *Config {
	return &Config{
		BaseURL:  "https://api.todoist.com/api/v1",
		Language: DefaultLanguage,
	}
}

// LoadConfig は設定を読み込む（環境変数優先）
func LoadConfig() (*Config, error) {
	// Viperの初期化
	v := viper.New()

	// デフォルト値の設定
	defaultConfig := DefaultConfig()
	v.SetDefault("api_token", defaultConfig.APIToken)
	v.SetDefault("base_url", defaultConfig.BaseURL)
	v.SetDefault("language", defaultConfig.Language)

	// 環境変数の設定（優先度最高）
	v.SetEnvPrefix("TODOIST")
	v.AutomaticEnv()

	// 設定ファイルの設定
	configDir, err := GetConfigDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get config directory: %w", err)
	}

	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath(configDir)

	// 設定ファイルが存在しない場合は自動生成
	configPath := filepath.Join(configDir, "config.yaml")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		if err := generateConfigFile(configPath, defaultConfig); err != nil {
			return nil, fmt.Errorf("failed to generate config file: %w", err)
		}
	}

	// 設定ファイルの読み込み
	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// 設定を構造体にマッピング
	var config Config
	if err := v.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// APIトークンが設定されているかチェック
	if config.APIToken == "" {
		return nil, fmt.Errorf("API token is required. Set TODOIST_API_TOKEN environment variable or add api_token to %s", configPath)
	}

	return &config, nil
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

// GetConfigDir は設定ディレクトリのパスを返す（XDG Base Directory仕様に準拠）
func GetConfigDir() (string, error) {
	// XDG_CONFIG_HOME環境変数があればそれを使用
	if configHome := os.Getenv("XDG_CONFIG_HOME"); configHome != "" {
		return filepath.Join(configHome, "gotodoist"), nil
	}

	// フォールバック: ~/.config/gotodoist
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}
	return filepath.Join(homeDir, ".config", "gotodoist"), nil
}

// generateConfigFile は設定ファイルを自動生成する
func generateConfigFile(configPath string, defaultConfig *Config) error {
	// 設定ディレクトリが存在しない場合は作成
	configDir := filepath.Dir(configPath)
	if err := os.MkdirAll(configDir, ConfigDirPerm); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// 設定ファイルの内容を作成
	configContent := `# gotodoist CLI設定ファイル
# API トークンはTodoist API設定ページから取得してください: https://todoist.com/prefs/integrations
# セキュリティ上の理由から、APIトークンは環境変数TODOIST_API_TOKENで設定することを推奨します

# Todoist API トークン（必須）
# api_token: "your-api-token-here"

# Todoist API ベースURL（通常は変更不要）
base_url: "` + defaultConfig.BaseURL + `"

# 言語設定（en/ja）
language: "` + defaultConfig.Language + `"
`

	// ファイルに書き込み
	if err := os.WriteFile(configPath, []byte(configContent), ConfigFilePerm); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}
