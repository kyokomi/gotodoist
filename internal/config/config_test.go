package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	expectedBaseURL  = "https://api.todoist.com/api/v1"
	expectedLanguage = "en"
)

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	require.NotNil(t, config, "DefaultConfig()がnilを返しました")
	assert.Equal(t, expectedBaseURL, config.BaseURL, "BaseURLが期待値と異なります")
	assert.Empty(t, config.APIToken, "APITokenが空ではありません")
	assert.Equal(t, expectedLanguage, config.Language, "Languageが期待値と異なります")
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
			require.NoError(t, err, "一時ディレクトリの作成に失敗しました")
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
				assert.Error(t, err, "エラーが期待されますが、nilが返されました")
				return
			}

			require.NoError(t, err, "予期しないエラーが発生しました")
			require.NotNil(t, config, "LoadConfig()がnil設定を返しました")

			assert.Equal(t, tt.wantToken, config.APIToken, "APITokenが期待値と異なります")
			assert.Equal(t, expectedBaseURL, config.BaseURL, "BaseURLが期待値と異なります")
			assert.Equal(t, expectedLanguage, config.Language, "Languageが期待値と異なります")

			// 設定ファイルが自動生成されているかチェック
			configDir := filepath.Join(tempDir, "gotodoist")
			configPath := filepath.Join(configDir, "config.yaml")
			_, err = os.Stat(configPath)
			assert.False(t, os.IsNotExist(err), "設定ファイルが生成されていません")
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
			require.NoError(t, err, "予期しないエラーが発生しました")
			assert.NotEmpty(t, configDir, "GetConfigDir()が空文字を返しました")

			// パスの期待値を計算
			var expectedDir string
			if tt.xdgConfigHome != "" {
				expectedDir = filepath.Join(tt.xdgConfigHome, "gotodoist")
			} else {
				homeDir, err := os.UserHomeDir()
				require.NoError(t, err, "ホームディレクトリの取得に失敗しました")
				expectedDir = filepath.Join(homeDir, ".config", "gotodoist")
			}

			assert.Equal(t, expectedDir, configDir, "configディレクトリが期待値と異なります")
		})
	}
}

func TestGenerateConfigFile(t *testing.T) {
	// テスト用の一時ディレクトリを作成
	tempDir, err := os.MkdirTemp("", "gotodoist-test")
	require.NoError(t, err, "一時ディレクトリの作成に失敗しました")
	defer func() {
		_ = os.RemoveAll(tempDir)
	}()

	configPath := filepath.Join(tempDir, "config.yaml")
	defaultConfig := DefaultConfig()

	// 設定ファイルを生成
	err = generateConfigFile(configPath, defaultConfig)
	require.NoError(t, err, "generateConfigFile()が失敗しました")

	// ファイルが作成されたかチェック
	_, err = os.Stat(configPath)
	assert.False(t, os.IsNotExist(err), "設定ファイルが作成されていません")

	// ファイルの内容をチェック
	content, err := os.ReadFile(filepath.Clean(configPath))
	require.NoError(t, err, "設定ファイルの読み込みに失敗しました")

	contentStr := string(content)

	// ディレクトリが作成されていることを確認
	_, err = os.Stat(filepath.Dir(configPath))
	assert.False(t, os.IsNotExist(err), "設定ディレクトリが作成されていません")

	// 基本的な設定項目が含まれているかチェック
	expectedContents := []string{
		"base_url: \"https://api.todoist.com/api/v1\"",
		"language: \"en\"",
		"# api_token:",
	}

	for _, expected := range expectedContents {
		assert.Contains(t, contentStr, expected, "設定ファイルに %q が含まれていません", expected)
	}
}

func TestLoadConfigFromFile(t *testing.T) {
	// テスト用の一時ディレクトリを作成
	tempDir, err := os.MkdirTemp("", "gotodoist-test")
	require.NoError(t, err, "一時ディレクトリの作成に失敗しました")
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
	require.NoError(t, err, "設定ディレクトリの作成に失敗しました")

	configContent := `# gotodoist CLI設定ファイル
api_token: "test-file-token"
base_url: "https://api.todoist.com/api/v1"
language: "ja"
`
	err = os.WriteFile(configPath, []byte(configContent), 0600)
	require.NoError(t, err, "設定ファイルの書き込みに失敗しました")

	// 設定を読み込み
	config, err := LoadConfig()
	require.NoError(t, err, "LoadConfig()が失敗しました")

	// 設定ファイルからの値が正しく読み込まれているかチェック
	assert.Equal(t, "test-file-token", config.APIToken, "APITokenが期待値と異なります")
	assert.Equal(t, "ja", config.Language, "Languageが期待値と異なります")
	assert.Equal(t, "https://api.todoist.com/api/v1", config.BaseURL, "BaseURLが期待値と異なります")
}
