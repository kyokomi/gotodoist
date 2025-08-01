// Package local はローカルファーストクライアントの実装を提供する
package local

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/kyokomi/gotodoist/internal/api"
	"github.com/kyokomi/gotodoist/internal/storage"
	"github.com/kyokomi/gotodoist/internal/sync"
)

// Client はローカルファーストのAPIクライアント
type Client struct {
	apiClient      *api.Client
	storage        *storage.SQLiteDB
	syncManager    *sync.Manager
	backgroundSync *sync.BackgroundSyncer
	config         *Config
	verbose        bool
}

// Config はローカルストレージの設定
type Config struct {
	Enabled            bool          `yaml:"enabled" mapstructure:"enabled"`
	DatabasePath       string        `yaml:"database_path" mapstructure:"database_path"`
	AutoSyncInterval   time.Duration `yaml:"auto_sync_interval" mapstructure:"auto_sync_interval"`
	InitialSyncOnStart bool          `yaml:"initial_sync_on_startup" mapstructure:"initial_sync_on_startup"`
	BackgroundSync     bool          `yaml:"background_sync" mapstructure:"background_sync"`
}

// DefaultConfig はデフォルトのローカルストレージ設定を返す
func DefaultConfig() *Config {
	return &Config{
		Enabled:            true,
		DatabasePath:       getDefaultDatabasePath(),
		AutoSyncInterval:   5 * time.Minute,
		InitialSyncOnStart: true,
		BackgroundSync:     true,
	}
}

// NewClient は新しいローカルファーストクライアントを作成する
func NewClient(apiClient *api.Client, config *Config, verbose bool) (*Client, error) {
	if !config.Enabled {
		// ローカルストレージが無効の場合は、APIクライアントをそのまま返すためのラッパー
		return &Client{
			apiClient: apiClient,
			config:    config,
			verbose:   verbose,
		}, nil
	}

	// SQLiteストレージを初期化
	storage, err := storage.NewSQLiteDB(config.DatabasePath)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize local storage: %w", err)
	}

	// 同期マネージャーを初期化
	syncManager := sync.NewManager(apiClient, storage, verbose)

	// バックグラウンド同期を初期化
	var backgroundSync *sync.BackgroundSyncer
	if config.BackgroundSync {
		backgroundSync = sync.NewBackgroundSyncer(syncManager, config.AutoSyncInterval)
	}

	client := &Client{
		apiClient:      apiClient,
		storage:        storage,
		syncManager:    syncManager,
		backgroundSync: backgroundSync,
		config:         config,
		verbose:        verbose,
	}

	return client, nil
}

// Initialize はクライアントを初期化する（必要に応じて初期同期を実行）
func (c *Client) Initialize(ctx context.Context) error {
	if !c.config.Enabled {
		return nil // ローカルストレージが無効の場合は何もしない
	}

	// 初期同期が必要かチェック
	if c.config.InitialSyncOnStart {
		initialDone, err := c.storage.IsInitialSyncDone()
		if err != nil {
			return fmt.Errorf("failed to check initial sync status: %w", err)
		}

		if !initialDone {
			if c.verbose {
				fmt.Println("🔄 Running initial sync...")
			}
			if err := c.syncManager.InitialSync(ctx); err != nil {
				return fmt.Errorf("failed to run initial sync: %w", err)
			}
		}
	}

	// NOTE: task list高速化のためバックグラウンド同期開始を一時的に無効化
	// バックグラウンド同期を開始
	// if c.backgroundSync != nil {
	// 	c.backgroundSync.Start(ctx)
	// 	if c.verbose {
	// 		fmt.Printf("🔄 Background sync started (interval: %v)\n", c.config.AutoSyncInterval)
	// 	}
	// }

	return nil
}

// Close はクライアントを終了する
func (c *Client) Close() error {
	if c.backgroundSync != nil {
		c.backgroundSync.Stop()
	}

	if c.storage != nil {
		return c.storage.Close()
	}

	return nil
}

// GetTasks は全てのタスクを取得する（ローカル優先）
func (c *Client) GetTasks(ctx context.Context) ([]api.Item, error) {
	if !c.config.Enabled {
		// ローカルストレージが無効の場合はAPIから直接取得
		return c.apiClient.GetTasks(ctx)
	}

	// NOTE: task list高速化のためバックグラウンド同期を一時的に無効化
	// バックグラウンド同期をトリガー（非同期）
	// if c.backgroundSync != nil {
	// 	go func() {
	// 		if err := c.syncManager.AutoSync(ctx, c.config.AutoSyncInterval); err != nil {
	// 			if c.verbose {
	// 				log.Printf("Background sync failed: %v", err)
	// 			}
	// 		}()
	// 	}()
	// }

	// ローカルから高速取得
	return c.storage.GetTasks()
}

// GetTasksByProject はプロジェクト指定でタスクを取得する（ローカル優先）
func (c *Client) GetTasksByProject(ctx context.Context, projectID string) ([]api.Item, error) {
	if !c.config.Enabled {
		return c.apiClient.GetTasksByProject(ctx, projectID)
	}

	// バックグラウンド同期をトリガー（非同期）
	if c.backgroundSync != nil {
		go func() {
			if err := c.syncManager.AutoSync(ctx, c.config.AutoSyncInterval); err != nil {
				if c.verbose {
					log.Printf("Background sync failed: %v", err)
				}
			}
		}()
	}

	// ローカルから高速取得
	return c.storage.GetTasksByProject(projectID)
}

// GetAllProjects は全てのプロジェクトを取得する（ローカル優先）
func (c *Client) GetAllProjects(ctx context.Context) ([]api.Project, error) {
	if !c.config.Enabled {
		return c.apiClient.GetAllProjects(ctx)
	}

	// バックグラウンド同期をトリガー（非同期）
	if c.backgroundSync != nil {
		go func() {
			if err := c.syncManager.AutoSync(ctx, c.config.AutoSyncInterval); err != nil {
				if c.verbose {
					log.Printf("Background sync failed: %v", err)
				}
			}
		}()
	}

	// ローカルから高速取得
	return c.storage.GetAllProjects()
}

// GetAllSections は全てのセクションを取得する（ローカル優先）
func (c *Client) GetAllSections(ctx context.Context) ([]api.Section, error) {
	if !c.config.Enabled {
		return c.apiClient.GetAllSections(ctx)
	}

	// バックグラウンド同期をトリガー（非同期）
	if c.backgroundSync != nil {
		go func() {
			if err := c.syncManager.AutoSync(ctx, c.config.AutoSyncInterval); err != nil {
				if c.verbose {
					log.Printf("Background sync failed: %v", err)
				}
			}
		}()
	}

	// ローカルから高速取得
	return c.storage.GetAllSections()
}

// CreateTask はタスクを作成する（API実行 + ローカル反映）
func (c *Client) CreateTask(ctx context.Context, req *api.CreateTaskRequest) (*api.SyncResponse, error) {
	// API実行
	resp, err := c.apiClient.CreateTask(ctx, req)
	if err != nil {
		return nil, err
	}

	// ローカルストレージが有効な場合は即座に反映
	if c.config.Enabled {
		// temp_id_mappingから実際のタスクを取得
		for tempID, realID := range resp.TempIDMapping {
			if tempID != realID {
				// 新しく作成されたタスクをローカルに保存
				// TODO: 実際のタスクオブジェクトが必要
				// 現時点では sync_token を更新するのみ
				if err := c.storage.SetSyncToken(resp.SyncToken); err != nil {
					log.Printf("Failed to update sync token after task creation: %v", err)
				}
				break
			}
		}

		// バックグラウンド同期をトリガーして最新状態に同期
		if c.backgroundSync != nil {
			c.backgroundSync.TriggerSync(ctx)
		}
	}

	return resp, nil
}

// UpdateTask はタスクを更新する（API実行 + ローカル反映）
func (c *Client) UpdateTask(ctx context.Context, taskID string, req *api.UpdateTaskRequest) (*api.SyncResponse, error) {
	// API実行
	resp, err := c.apiClient.UpdateTask(ctx, taskID, req)
	if err != nil {
		return nil, err
	}

	// ローカルストレージが有効な場合は sync_token を更新
	if c.config.Enabled {
		if err := c.storage.SetSyncToken(resp.SyncToken); err != nil {
			log.Printf("Failed to update sync token after task update: %v", err)
		}

		// バックグラウンド同期をトリガー
		if c.backgroundSync != nil {
			c.backgroundSync.TriggerSync(ctx)
		}
	}

	return resp, nil
}

// DeleteTask はタスクを削除する（API実行 + ローカル反映）
func (c *Client) DeleteTask(ctx context.Context, taskID string) (*api.SyncResponse, error) {
	// API実行
	resp, err := c.apiClient.DeleteTask(ctx, taskID)
	if err != nil {
		return nil, err
	}

	// ローカルストレージが有効な場合は即座に反映
	if c.config.Enabled {
		if err := c.storage.DeleteTask(taskID); err != nil {
			log.Printf("Failed to delete task from local storage: %v", err)
		}

		if err := c.storage.SetSyncToken(resp.SyncToken); err != nil {
			log.Printf("Failed to update sync token after task deletion: %v", err)
		}
	}

	return resp, nil
}

// CloseTask はタスクを完了にする（API実行 + ローカル反映）
func (c *Client) CloseTask(ctx context.Context, taskID string) (*api.SyncResponse, error) {
	// API実行
	resp, err := c.apiClient.CloseTask(ctx, taskID)
	if err != nil {
		return nil, err
	}

	// ローカルストレージが有効な場合は sync_token を更新
	if c.config.Enabled {
		if err := c.storage.SetSyncToken(resp.SyncToken); err != nil {
			log.Printf("Failed to update sync token after task completion: %v", err)
		}

		// バックグラウンド同期をトリガー
		if c.backgroundSync != nil {
			c.backgroundSync.TriggerSync(ctx)
		}
	}

	return resp, nil
}

// Sync は手動で同期を実行する
func (c *Client) Sync(ctx context.Context) error {
	if !c.config.Enabled {
		return fmt.Errorf("local storage is disabled")
	}

	return c.syncManager.IncrementalSync(ctx)
}

// GetSyncStatus は同期状態を取得する
func (c *Client) GetSyncStatus() (*sync.SyncStatus, error) {
	if !c.config.Enabled {
		return nil, fmt.Errorf("local storage is disabled")
	}

	return c.syncManager.GetSyncStatus()
}

// IsLocalStorageEnabled はローカルストレージが有効かどうかを返す
func (c *Client) IsLocalStorageEnabled() bool {
	return c.config.Enabled
}

// ForceInitialSync は強制的に初期同期を実行する
func (c *Client) ForceInitialSync(ctx context.Context) error {
	if !c.config.Enabled {
		return fmt.Errorf("local storage is disabled")
	}

	return c.syncManager.ForceInitialSync(ctx)
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
