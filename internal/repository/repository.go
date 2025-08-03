// Package repository はRepositoryの実装を提供する
package repository

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/kyokomi/gotodoist/internal/api"
	"github.com/kyokomi/gotodoist/internal/storage"
	"github.com/kyokomi/gotodoist/internal/sync"
)

// Repository はローカルファーストのTodoistリポジトリ
type Repository struct {
	apiClient   api.Interface
	storage     *storage.SQLiteDB
	syncManager *sync.Manager
	config      *Config
	verbose     bool
}

// NewRepository は新しいローカルファーストリポジトリを作成する
func NewRepository(apiClient api.Interface, config *Config, verbose bool) (*Repository, error) {
	if !config.Enabled {
		// ローカルストレージが無効の場合は、APIを直接呼び出すRepository
		return &Repository{
			apiClient: apiClient,
			config:    config,
			verbose:   verbose,
		}, nil
	}

	// SQLiteストレージを初期化
	st, err := storage.NewSQLiteDB(config.DatabasePath)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize local storage: %w", err)
	}

	// 同期マネージャーを初期化
	syncManager := sync.NewManager(apiClient, st, verbose)

	client := &Repository{
		apiClient:   apiClient,
		storage:     st,
		syncManager: syncManager,
		config:      config,
		verbose:     verbose,
	}

	return client, nil
}

// Initialize はRepositoryを初期化する（必要に応じて初期同期を実行）
func (c *Repository) Initialize(ctx context.Context) error {
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

	return nil
}

// Close はRepositoryを終了する
func (c *Repository) Close() error {
	if c.storage != nil {
		return c.storage.Close()
	}

	return nil
}

// GetTasks は全てのタスクを取得する（ローカル優先）
func (c *Repository) GetTasks(ctx context.Context) ([]api.Item, error) {
	if !c.config.Enabled {
		// ローカルストレージが無効の場合はAPIから直接取得
		return c.apiClient.GetTasks(ctx)
	}

	// ローカルから高速取得
	return c.storage.GetTasks()
}

// GetTasksByProject はプロジェクト指定でタスクを取得する（ローカル優先）
func (c *Repository) GetTasksByProject(ctx context.Context, projectID string) ([]api.Item, error) {
	if !c.config.Enabled {
		return c.apiClient.GetTasksByProject(ctx, projectID)
	}

	// ローカルから高速取得
	return c.storage.GetTasksByProject(projectID)
}

// GetAllProjects は全てのプロジェクトを取得する（ローカル優先）
func (c *Repository) GetAllProjects(ctx context.Context) ([]api.Project, error) {
	if !c.config.Enabled {
		return c.apiClient.GetAllProjects(ctx)
	}

	// ローカルから高速取得
	return c.storage.GetAllProjects()
}

// GetAllSections は全てのセクションを取得する（ローカル優先）
func (c *Repository) GetAllSections(ctx context.Context) ([]api.Section, error) {
	if !c.config.Enabled {
		return c.apiClient.GetAllSections(ctx)
	}

	// ローカルから高速取得
	return c.storage.GetAllSections()
}

// CreateTask はタスクを作成する（API実行 + ローカル反映）
func (c *Repository) CreateTask(ctx context.Context, req *api.CreateTaskRequest) (*api.SyncResponse, error) {
	// API実行
	resp, err := c.apiClient.CreateTask(ctx, req)
	if err != nil {
		return nil, err
	}

	// ローカルストレージが有効な場合は増分同期で最新データを取得
	if c.config.Enabled {
		// 作成後に増分同期を実行して新しいタスクを取得
		if err := c.syncManager.IncrementalSync(ctx); err != nil {
			log.Printf("Failed to sync after task creation: %v", err)
		}
	}

	return resp, nil
}

// UpdateTask はタスクを更新する（API実行 + ローカル反映）
func (c *Repository) UpdateTask(ctx context.Context, taskID string, req *api.UpdateTaskRequest) (*api.SyncResponse, error) {
	// API実行
	resp, err := c.apiClient.UpdateTask(ctx, taskID, req)
	if err != nil {
		return nil, err
	}

	// ローカルストレージが有効な場合は増分同期で最新データを取得
	if c.config.Enabled {
		// 更新後に増分同期を実行して変更をローカルに反映
		if err := c.syncManager.IncrementalSync(ctx); err != nil {
			log.Printf("Failed to sync after task update: %v", err)
		}
	}

	return resp, nil
}

// DeleteTask はタスクを削除する（API実行 + ローカル反映）
func (c *Repository) DeleteTask(ctx context.Context, taskID string) (*api.SyncResponse, error) {
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

		// 削除後に増分同期を実行してAPI側の変更をローカルに反映
		if err := c.syncManager.IncrementalSync(ctx); err != nil {
			log.Printf("Failed to sync after task deletion: %v", err)
		}
	}

	return resp, nil
}

// CloseTask はタスクを完了にする（API実行 + ローカル反映）
func (c *Repository) CloseTask(ctx context.Context, taskID string) (*api.SyncResponse, error) {
	// API実行
	resp, err := c.apiClient.CloseTask(ctx, taskID)
	if err != nil {
		return nil, err
	}

	// ローカルストレージが有効な場合
	if c.config.Enabled {
		// sync_token を更新
		if err := c.storage.SetSyncToken(resp.SyncToken); err != nil {
			log.Printf("Failed to update sync token after task completion: %v", err)
		}

		// ローカルストレージでタスクを完了状態に更新
		if err := c.storage.UpdateTaskCompleted(taskID, true); err != nil {
			log.Printf("Failed to update task completion status in local storage: %v", err)
		}
	}

	return resp, nil
}

// ReopenTask はタスクを未完了に戻す（API実行 + ローカル反映）
func (c *Repository) ReopenTask(ctx context.Context, taskID string) (*api.SyncResponse, error) {
	// API実行
	resp, err := c.apiClient.ReopenTask(ctx, taskID)
	if err != nil {
		return nil, err
	}

	// ローカルストレージが有効な場合
	if c.config.Enabled {
		// sync_token を更新
		if err := c.storage.SetSyncToken(resp.SyncToken); err != nil {
			log.Printf("Failed to update sync token after task reopening: %v", err)
		}

		// ローカルストレージでタスクを未完了状態に更新
		if err := c.storage.UpdateTaskCompleted(taskID, false); err != nil {
			log.Printf("Failed to update task completion status in local storage: %v", err)
		}
	}

	return resp, nil
}

// Sync は手動で同期を実行する
func (c *Repository) Sync(ctx context.Context) error {
	if !c.config.Enabled {
		return fmt.Errorf("local storage is disabled")
	}

	return c.syncManager.IncrementalSync(ctx)
}

// GetSyncStatus は同期状態を取得する
func (c *Repository) GetSyncStatus() (*sync.Status, error) {
	if !c.config.Enabled {
		return nil, fmt.Errorf("local storage is disabled")
	}

	return c.syncManager.GetSyncStatus()
}

// IsLocalStorageEnabled はローカルストレージが有効かどうかを返す
func (c *Repository) IsLocalStorageEnabled() bool {
	return c.config.Enabled
}

// ForceInitialSync は強制的に初期同期を実行する
func (c *Repository) ForceInitialSync(ctx context.Context) error {
	if !c.config.Enabled {
		return fmt.Errorf("local storage is disabled")
	}

	return c.syncManager.ForceInitialSync(ctx)
}

// CreateProject はプロジェクトを作成する（API実行 + ローカル反映）
func (c *Repository) CreateProject(ctx context.Context, req *api.CreateProjectRequest) (*api.SyncResponse, error) {
	// API実行
	resp, err := c.apiClient.CreateProject(ctx, req)
	if err != nil {
		return nil, err
	}

	// ローカルストレージが有効な場合は増分同期で最新データを取得
	if c.config.Enabled {
		// 作成後に増分同期を実行して新しいプロジェクトを取得
		if err := c.syncManager.IncrementalSync(ctx); err != nil {
			log.Printf("Failed to sync after project creation: %v", err)
		}
	}

	return resp, nil
}

// UpdateProject はプロジェクトを更新する（API実行 + ローカル反映）
func (c *Repository) UpdateProject(ctx context.Context, projectID string, req *api.UpdateProjectRequest) (*api.SyncResponse, error) {
	// API実行
	resp, err := c.apiClient.UpdateProject(ctx, projectID, req)
	if err != nil {
		return nil, err
	}

	// ローカルストレージが有効な場合は増分同期で最新データを取得
	if c.config.Enabled {
		// 更新後に増分同期を実行して変更をローカルに反映
		if err := c.syncManager.IncrementalSync(ctx); err != nil {
			log.Printf("Failed to sync after project update: %v", err)
		}
	}

	return resp, nil
}

// DeleteProject はプロジェクトを削除する（API実行 + ローカル反映）
func (c *Repository) DeleteProject(ctx context.Context, projectID string) (*api.SyncResponse, error) {
	// API実行
	resp, err := c.apiClient.DeleteProject(ctx, projectID)
	if err != nil {
		return nil, err
	}

	// ローカルストレージが有効な場合は即座に反映
	if c.config.Enabled {
		// プロジェクトに属するタスクを先に削除（カスケード削除）
		if err := c.storage.DeleteTasksByProject(projectID); err != nil {
			log.Printf("Failed to delete tasks for project %s: %v", projectID, err)
		}

		// プロジェクト自体を削除
		if err := c.storage.DeleteProject(projectID); err != nil {
			log.Printf("Failed to delete project from local storage: %v", err)
		}

		// 削除後に増分同期を実行してAPI側の変更をローカルに反映
		if err := c.syncManager.IncrementalSync(ctx); err != nil {
			log.Printf("Failed to sync after project deletion: %v", err)
		}
	}

	return resp, nil
}

// ArchiveProject はプロジェクトをアーカイブする（API実行 + ローカル反映）
func (c *Repository) ArchiveProject(ctx context.Context, projectID string) (*api.SyncResponse, error) {
	// API実行
	resp, err := c.apiClient.ArchiveProject(ctx, projectID)
	if err != nil {
		return nil, err
	}

	// ローカルストレージが有効な場合は増分同期で最新データを取得
	if c.config.Enabled {
		// アーカイブ後に増分同期を実行して変更をローカルに反映
		if err := c.syncManager.IncrementalSync(ctx); err != nil {
			log.Printf("Failed to sync after project archive: %v", err)
		}
	}

	return resp, nil
}

// UnarchiveProject はプロジェクトのアーカイブを解除する（API実行 + ローカル反映）
func (c *Repository) UnarchiveProject(ctx context.Context, projectID string) (*api.SyncResponse, error) {
	// API実行
	resp, err := c.apiClient.UnarchiveProject(ctx, projectID)
	if err != nil {
		return nil, err
	}

	// ローカルストレージが有効な場合は増分同期で最新データを取得
	if c.config.Enabled {
		// アンアーカイブ後に増分同期を実行して変更をローカルに反映
		if err := c.syncManager.IncrementalSync(ctx); err != nil {
			log.Printf("Failed to sync after project unarchive: %v", err)
		}
	}

	return resp, nil
}

// FindProjectIDByName はプロジェクト名またはIDからプロジェクトIDを検索する
// 検索順序: 1. ID完全一致 2. 名前完全一致 3. 名前部分一致
func (c *Repository) FindProjectIDByName(ctx context.Context, nameOrID string) (string, error) {
	projects, err := c.GetAllProjects(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get projects: %w", err)
	}

	nameOrIDLower := strings.ToLower(nameOrID)

	// 1. ID完全一致（最優先・最高速）
	for _, project := range projects {
		if project.ID == nameOrID {
			return project.ID, nil
		}
	}

	// 2. 名前完全一致（大文字小文字を無視）
	for _, project := range projects {
		if strings.EqualFold(project.Name, nameOrID) {
			return project.ID, nil
		}
	}

	// 3. 名前部分一致（最後の手段）
	for _, project := range projects {
		if strings.Contains(strings.ToLower(project.Name), nameOrIDLower) {
			return project.ID, nil
		}
	}

	return "", fmt.Errorf("project not found: %s", nameOrID)
}

// ResetLocalStorage はローカルストレージを完全にリセットする
func (c *Repository) ResetLocalStorage(_ context.Context) error {
	if !c.config.Enabled {
		return fmt.Errorf("local storage is disabled")
	}

	return c.storage.ResetAllData()
}
