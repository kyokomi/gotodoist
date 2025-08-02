package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const (
	expectedBaseURL  = "https://api.todoist.com/api/v1"
	expectedLanguage = "en"
)

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	if config == nil {
		t.Fatal("DefaultConfig() returned nil")
	}

	if config.BaseURL != expectedBaseURL {
		t.Errorf("expected BaseURL to be %s, got %s", expectedBaseURL, config.BaseURL)
	}

	if config.APIToken != "" {
		t.Errorf("expected APIToken to be empty, got %s", config.APIToken)
	}

	if config.Language != expectedLanguage {
		t.Errorf("expected Language to be %q, got %s", expectedLanguage, config.Language)
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
			// テスト用の一時ディレクトリを作成
			tempDir, err := os.MkdirTemp("", "gotodoist-test")
			if err != nil {
				t.Fatalf("failed to create temp dir: %v", err)
			}
			defer func() {
				_ = os.RemoveAll(tempDir)
			}()

			// 環境変数をクリーンアップ
			oldToken := os.Getenv("TODOIST_API_TOKEN")
			oldConfigHome := os.Getenv("XDG_CONFIG_HOME")
			defer func() {
				if oldToken != "" {
					_ = os.Setenv("TODOIST_API_TOKEN", oldToken)
				} else {
					_ = os.Unsetenv("TODOIST_API_TOKEN")
				}
				if oldConfigHome != "" {
					_ = os.Setenv("XDG_CONFIG_HOME", oldConfigHome)
				} else {
					_ = os.Unsetenv("XDG_CONFIG_HOME")
				}
			}()

			// テスト用の設定ディレクトリを設定
			_ = os.Setenv("XDG_CONFIG_HOME", tempDir)

			// テスト用の環境変数を設定
			if tt.envToken != "" {
				_ = os.Setenv("TODOIST_API_TOKEN", tt.envToken)
			} else {
				_ = os.Unsetenv("TODOIST_API_TOKEN")
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

			if config.BaseURL != expectedBaseURL {
				t.Errorf("expected BaseURL to be %s, got %s", expectedBaseURL, config.BaseURL)
			}

			if config.Language != expectedLanguage {
				t.Errorf("expected Language to be %q, got %s", expectedLanguage, config.Language)
			}

			// 設定ファイルが自動生成されているかチェック
			configDir := filepath.Join(tempDir, "gotodoist")
			configPath := filepath.Join(configDir, "config.yaml")
			if _, err := os.Stat(configPath); os.IsNotExist(err) {
				t.Error("config file was not generated")
			}
		})
	}
}

func TestGetConfigDir(t *testing.T) {
	tests := []struct {
		name           string
		xdgConfigHome  string
		expectedSuffix string
	}{
		{
			name:           "with XDG_CONFIG_HOME",
			xdgConfigHome:  "/custom/config",
			expectedSuffix: "gotodoist",
		},
		{
			name:           "without XDG_CONFIG_HOME",
			xdgConfigHome:  "",
			expectedSuffix: filepath.Join(".config", "gotodoist"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 環境変数のクリーンアップ
			oldConfigHome := os.Getenv("XDG_CONFIG_HOME")
			defer func() {
				if oldConfigHome != "" {
					_ = os.Setenv("XDG_CONFIG_HOME", oldConfigHome)
				} else {
					_ = os.Unsetenv("XDG_CONFIG_HOME")
				}
			}()

			// テスト用の環境変数を設定
			if tt.xdgConfigHome != "" {
				_ = os.Setenv("XDG_CONFIG_HOME", tt.xdgConfigHome)
			} else {
				_ = os.Unsetenv("XDG_CONFIG_HOME")
			}

			configDir, err := GetConfigDir()
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if configDir == "" {
				t.Error("GetConfigDir() returned empty string")
				return
			}

			// パスの期待値を計算
			var expectedDir string
			if tt.xdgConfigHome != "" {
				expectedDir = filepath.Join(tt.xdgConfigHome, "gotodoist")
			} else {
				homeDir, err := os.UserHomeDir()
				if err != nil {
					t.Fatalf("failed to get home directory: %v", err)
				}
				expectedDir = filepath.Join(homeDir, ".config", "gotodoist")
			}

			if configDir != expectedDir {
				t.Errorf("expected config dir %s, got %s", expectedDir, configDir)
			}
		})
	}
}

func TestGenerateConfigFile(t *testing.T) {
	// テスト用の一時ディレクトリを作成
	tempDir, err := os.MkdirTemp("", "gotodoist-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() {
		_ = os.RemoveAll(tempDir)
	}()

	configPath := filepath.Join(tempDir, "config.yaml")
	defaultConfig := DefaultConfig()

	// 設定ファイルを生成
	err = generateConfigFile(configPath, defaultConfig)
	if err != nil {
		t.Errorf("generateConfigFile() failed: %v", err)
		return
	}

	// ファイルが作成されたかチェック
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Error("config file was not created")
		return
	}

	// ファイルの内容をチェック
	content, err := os.ReadFile(filepath.Clean(configPath))
	if err != nil {
		t.Errorf("failed to read config file: %v", err)
		return
	}

	contentStr := string(content)

	// ディレクトリが作成されていることを確認
	if _, err := os.Stat(filepath.Dir(configPath)); os.IsNotExist(err) {
		t.Error("config directory was not created")
	}

	// 基本的な設定項目が含まれているかチェック
	expectedContents := []string{
		"base_url: \"https://api.todoist.com/api/v1\"",
		"language: \"en\"",
		"# api_token:",
	}

	for _, expected := range expectedContents {
		if !strings.Contains(contentStr, expected) {
			t.Errorf("config file should contain %q, but got:\n%s", expected, contentStr)
		}
	}
}

func TestLoadConfigFromFile(t *testing.T) {
	// テスト用の一時ディレクトリを作成
	tempDir, err := os.MkdirTemp("", "gotodoist-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() {
		_ = os.RemoveAll(tempDir)
	}()

	// 環境変数をクリーンアップ
	oldToken := os.Getenv("TODOIST_API_TOKEN")
	oldConfigHome := os.Getenv("XDG_CONFIG_HOME")
	defer func() {
		if oldToken != "" {
			_ = os.Setenv("TODOIST_API_TOKEN", oldToken)
		} else {
			_ = os.Unsetenv("TODOIST_API_TOKEN")
		}
		if oldConfigHome != "" {
			_ = os.Setenv("XDG_CONFIG_HOME", oldConfigHome)
		} else {
			_ = os.Unsetenv("XDG_CONFIG_HOME")
		}
	}()

	// 環境変数でのトークンを削除
	_ = os.Unsetenv("TODOIST_API_TOKEN")
	_ = os.Setenv("XDG_CONFIG_HOME", tempDir)

	// テスト用の設定ファイルを作成
	configDir := filepath.Join(tempDir, "gotodoist")
	configPath := filepath.Join(configDir, "config.yaml")
	err = os.MkdirAll(configDir, 0750)
	if err != nil {
		t.Fatalf("failed to create config dir: %v", err)
	}

	configContent := `# gotodoist CLI設定ファイル
api_token: "test-file-token"
base_url: "https://api.todoist.com/api/v1"
language: "ja"
`
	err = os.WriteFile(configPath, []byte(configContent), 0600)
	if err != nil {
		t.Fatalf("failed to write config file: %v", err)
	}

	// 設定を読み込み
	config, err := LoadConfig()
	if err != nil {
		t.Errorf("LoadConfig() failed: %v", err)
		return
	}

	// 設定ファイルからの値が正しく読み込まれているかチェック
	if config.APIToken != "test-file-token" {
		t.Errorf("expected APIToken 'test-file-token', got %s", config.APIToken)
	}

	if config.Language != "ja" {
		t.Errorf("expected Language 'ja', got %s", config.Language)
	}

	if config.BaseURL != "https://api.todoist.com/api/v1" {
		t.Errorf("expected BaseURL 'https://api.todoist.com/api/v1', got %s", config.BaseURL)
	}
}
