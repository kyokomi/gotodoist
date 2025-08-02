package factory

import (
	"testing"

	"github.com/kyokomi/gotodoist/internal/config"
	"github.com/kyokomi/gotodoist/internal/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewAPIClient(t *testing.T) {
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
			cfg := &config.Config{
				APIToken: tt.token,
				BaseURL:  tt.baseURL,
			}

			client, err := NewAPIClient(cfg)

			if tt.wantError {
				assert.Error(t, err, "エラーが期待されますが、nilが返されました")
				return
			}

			require.NoError(t, err, "予期しないエラーが発生しました")
			assert.NotNil(t, client, "クライアントが期待されますが、nilが返されました")
		})
	}
}

func TestNewAPIClient_WithCustomBaseURL(t *testing.T) {
	cfg := &config.Config{
		APIToken: "test-token",
		BaseURL:  "https://custom.api.com",
	}

	client, err := NewAPIClient(cfg)
	require.NoError(t, err, "予期しないエラーが発生しました")
	assert.NotNil(t, client, "クライアントが期待されますが、nilが返されました")

	// BaseURL が正しく設定されているかは、APIクライアント内部の実装詳細なので
	// ここでは単純にエラーが発生しないことを確認
}

func TestNewAPIClient_WithEmptyBaseURL(t *testing.T) {
	cfg := &config.Config{
		APIToken: "test-token",
		BaseURL:  "",
	}

	client, err := NewAPIClient(cfg)
	require.NoError(t, err, "予期しないエラーが発生しました")
	assert.NotNil(t, client, "クライアントが期待されますが、nilが返されました")
}

func TestNewRepository(t *testing.T) {
	cfg := &config.Config{
		APIToken:     "test-token",
		BaseURL:      "https://api.todoist.com/api/v1",
		Language:     "en",
		LocalStorage: repository.DefaultConfig(),
	}

	repo, err := NewRepository(cfg, false)
	require.NoError(t, err, "予期しないエラーが発生しました")
	assert.NotNil(t, repo, "リポジトリが期待されますが、nilが返されました")

	// Repositoryが正常に作成されたことを確認
	defer func() {
		if closeErr := repo.Close(); closeErr != nil {
			t.Errorf("failed to close repository: %v", closeErr)
		}
	}()
}
