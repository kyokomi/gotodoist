// Package factory provides factory functions for creating various instances.
package factory

import (
	"fmt"

	"github.com/kyokomi/gotodoist/internal/api"
	"github.com/kyokomi/gotodoist/internal/config"
	"github.com/kyokomi/gotodoist/internal/repository"
)

// NewAPIClient は設定からAPIクライアントを作成する
func NewAPIClient(cfg *config.Config) (*api.Client, error) {
	client, err := api.NewClient(cfg.APIToken)
	if err != nil {
		return nil, fmt.Errorf("failed to create API client: %w", err)
	}

	if cfg.BaseURL != "" {
		if err := client.SetBaseURL(cfg.BaseURL); err != nil {
			return nil, fmt.Errorf("failed to set base URL: %w", err)
		}
	}

	return client, nil
}

// NewRepository は設定からRepositoryを作成する
func NewRepository(cfg *config.Config, verbose bool) (*repository.Repository, error) {
	// 基本APIクライアントを作成
	apiClient, err := NewAPIClient(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create API client: %w", err)
	}

	// Repositoryを作成
	localRepository, err := repository.NewRepository(apiClient, cfg.LocalStorage, verbose)
	if err != nil {
		return nil, fmt.Errorf("failed to create Repository: %w", err)
	}

	return localRepository, nil
}
