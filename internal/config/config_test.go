package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	if config == nil {
		t.Fatal("DefaultConfig() returned nil")
	}

	if config.BaseURL != "https://api.todoist.com/api/v1" {
		t.Errorf("expected BaseURL to be https://api.todoist.com/api/v1, got %s", config.BaseURL)
	}

	if config.APIToken != "" {
		t.Errorf("expected APIToken to be empty, got %s", config.APIToken)
	}
}

func TestLoadConfig(t *testing.T) {
	tests := []struct {
		name      string
		envToken  string
		wantError bool
		wantToken string
	}{
		{
			name:      "valid token from env",
			envToken:  "test-api-token",
			wantError: false,
			wantToken: "test-api-token",
		},
		{
			name:      "empty token",
			envToken:  "",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 環境変数をクリーンアップ
			oldToken := os.Getenv("TODOIST_API_TOKEN")
			defer func() {
				if oldToken != "" {
					os.Setenv("TODOIST_API_TOKEN", oldToken)
				} else {
					os.Unsetenv("TODOIST_API_TOKEN")
				}
			}()

			// テスト用の環境変数を設定
			if tt.envToken != "" {
				os.Setenv("TODOIST_API_TOKEN", tt.envToken)
			} else {
				os.Unsetenv("TODOIST_API_TOKEN")
			}

			config, err := LoadConfig()

			if tt.wantError {
				if err == nil {
					t.Error("expected error but got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if config == nil {
				t.Fatal("LoadConfig() returned nil config")
			}

			if config.APIToken != tt.wantToken {
				t.Errorf("expected APIToken %s, got %s", tt.wantToken, config.APIToken)
			}

			if config.BaseURL != "https://api.todoist.com/api/v1" {
				t.Errorf("expected BaseURL to be https://api.todoist.com/api/v1, got %s", config.BaseURL)
			}
		})
	}
}

func TestConfig_NewAPIClient(t *testing.T) {
	tests := []struct {
		name      string
		token     string
		baseURL   string
		wantError bool
	}{
		{
			name:      "valid config",
			token:     "test-token",
			baseURL:   "https://api.todoist.com/api/v1",
			wantError: false,
		},
		{
			name:      "empty token",
			token:     "",
			baseURL:   "https://api.todoist.com/api/v1",
			wantError: true,
		},
		{
			name:      "invalid base URL",
			token:     "test-token",
			baseURL:   "ht!tp://invalid-url with spaces",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &Config{
				APIToken: tt.token,
				BaseURL:  tt.baseURL,
			}

			client, err := config.NewAPIClient()

			if tt.wantError {
				if err == nil {
					t.Error("expected error but got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if client == nil {
				t.Error("expected client but got nil")
			}
		})
	}
}

func TestGetConfigDir(t *testing.T) {
	configDir, err := GetConfigDir()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
		return
	}

	if configDir == "" {
		t.Error("GetConfigDir() returned empty string")
		return
	}

	// 設定ディレクトリのパスが期待通りの形式かチェック
	homeDir, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("failed to get home directory: %v", err)
	}

	expectedDir := filepath.Join(homeDir, ".gotodoist")
	if configDir != expectedDir {
		t.Errorf("expected config dir %s, got %s", expectedDir, configDir)
	}
}

func TestConfig_NewAPIClient_WithCustomBaseURL(t *testing.T) {
	config := &Config{
		APIToken: "test-token",
		BaseURL:  "https://custom.api.com",
	}

	client, err := config.NewAPIClient()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
		return
	}

	if client == nil {
		t.Error("expected client but got nil")
		return
	}

	// BaseURL が正しく設定されているかは、APIクライアント内部の実装詳細なので
	// ここでは単純にエラーが発生しないことを確認
}

func TestConfig_NewAPIClient_WithEmptyBaseURL(t *testing.T) {
	config := &Config{
		APIToken: "test-token",
		BaseURL:  "",
	}

	client, err := config.NewAPIClient()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
		return
	}

	if client == nil {
		t.Error("expected client but got nil")
	}
}
