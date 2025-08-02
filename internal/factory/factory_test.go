package factory

import (
	"testing"

	"github.com/kyokomi/gotodoist/internal/config"
	"github.com/kyokomi/gotodoist/internal/repository"
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

func TestNewAPIClient_WithCustomBaseURL(t *testing.T) {
	cfg := &config.Config{
		APIToken: "test-token",
		BaseURL:  "https://custom.api.com",
	}

	client, err := NewAPIClient(cfg)
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

func TestNewAPIClient_WithEmptyBaseURL(t *testing.T) {
	cfg := &config.Config{
		APIToken: "test-token",
		BaseURL:  "",
	}

	client, err := NewAPIClient(cfg)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
		return
	}

	if client == nil {
		t.Error("expected client but got nil")
	}
}

func TestNewRepository(t *testing.T) {
	cfg := &config.Config{
		APIToken:     "test-token",
		BaseURL:      "https://api.todoist.com/api/v1",
		Language:     "en",
		LocalStorage: repository.DefaultConfig(),
	}

	repo, err := NewRepository(cfg, false)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
		return
	}

	if repo == nil {
		t.Error("expected repository but got nil")
		return
	}

	// Repositoryが正常に作成されたことを確認
	defer func() {
		if closeErr := repo.Close(); closeErr != nil {
			t.Errorf("failed to close repository: %v", closeErr)
		}
	}()
}
