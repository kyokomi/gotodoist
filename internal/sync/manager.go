// Package sync はTodoist APIとの同期機能を提供する
package sync

import (
	"context"
	"fmt"
	"time"

	"github.com/kyokomi/gotodoist/internal/api"
	"github.com/kyokomi/gotodoist/internal/storage"
)

// Manager は同期処理を管理する
type Manager struct {
	apiClient *api.Client
	storage   *storage.SQLiteDB
	verbose   bool
}

// NewManager は新しいSyncManagerを作成する
func NewManager(apiClient *api.Client, storage *storage.SQLiteDB, verbose bool) *Manager {
	return &Manager{
		apiClient: apiClient,
		storage:   storage,
		verbose:   verbose,
	}
}

// InitialSync は初期同期を実行する（全データを取得）
func (m *Manager) InitialSync(ctx context.Context) error {
	if m.verbose {
		fmt.Println("🔄 Starting initial sync...")
	}

	// sync_token="*"で全データを取得
	resp, err := m.apiClient.Sync(ctx, &api.SyncRequest{
		SyncToken:     "*",
		ResourceTypes: []string{api.ResourceItems, api.ResourceProjects, api.ResourceSections},
	})
	if err != nil {
		return fmt.Errorf("failed to fetch initial data: %w", err)
	}

	// トランザクション内で全データを保存
	tx, err := m.storage.BeginTx()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		if err != nil {
			if rollbackErr := tx.Rollback(); rollbackErr != nil {
				fmt.Printf("Warning: failed to rollback transaction: %v\n", rollbackErr)
			}
		}
	}()

	// プロジェクトを保存
	if m.verbose {
		fmt.Printf("📁 Saving %d projects...\n", len(resp.Projects))
	}
	for _, project := range resp.Projects {
		if err := m.storage.InsertProject(project); err != nil {
			return fmt.Errorf("failed to insert project %s: %w", project.ID, err)
		}
	}

	// セクションを保存
	if m.verbose {
		fmt.Printf("📂 Saving %d sections...\n", len(resp.Sections))
	}
	for _, section := range resp.Sections {
		if err := m.storage.InsertSection(section); err != nil {
			return fmt.Errorf("failed to insert section %s: %w", section.ID, err)
		}
	}

	// タスクを保存
	if m.verbose {
		fmt.Printf("📝 Saving %d tasks...\n", len(resp.Items))
	}
	for _, task := range resp.Items {
		if err := m.storage.InsertTask(task); err != nil {
			return fmt.Errorf("failed to insert task %s: %w", task.ID, err)
		}
	}

	// sync_tokenと同期状態を更新
	if err := m.storage.SetSyncToken(resp.SyncToken); err != nil {
		return fmt.Errorf("failed to set sync token: %w", err)
	}

	if err := m.storage.SetLastSyncTime(time.Now()); err != nil {
		return fmt.Errorf("failed to set last sync time: %w", err)
	}

	if err := m.storage.SetInitialSyncDone(true); err != nil {
		return fmt.Errorf("failed to set initial sync done: %w", err)
	}

	// トランザクションコミット
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	if m.verbose {
		fmt.Println("✅ Initial sync completed successfully!")
	}

	return nil
}

// IncrementalSync は増分同期を実行する（差分のみ取得）
func (m *Manager) IncrementalSync(ctx context.Context) error {
	if err := m.checkInitialSyncStatus(ctx); err != nil {
		return err
	}

	if m.verbose {
		fmt.Println("🔄 Starting incremental sync...")
	}

	resp, err := m.fetchIncrementalData(ctx)
	if err != nil {
		return err
	}

	if m.hasNoChanges(resp) {
		if m.verbose {
			fmt.Println("📭 No changes since last sync")
		}
		return nil
	}

	if err := m.applyIncrementalChanges(resp); err != nil {
		return err
	}

	if m.verbose {
		fmt.Println("✅ Incremental sync completed successfully!")
	}

	return nil
}

// checkInitialSyncStatus は初期同期が完了しているかチェックし、未完了の場合は初期同期を実行する
func (m *Manager) checkInitialSyncStatus(ctx context.Context) error {
	initialDone, err := m.storage.IsInitialSyncDone()
	if err != nil {
		return fmt.Errorf("failed to check initial sync status: %w", err)
	}

	if !initialDone {
		if m.verbose {
			fmt.Println("Initial sync not done, running initial sync first...")
		}
		return m.InitialSync(ctx)
	}

	return nil
}

// fetchIncrementalData は前回のsync_tokenを使って差分データを取得する
func (m *Manager) fetchIncrementalData(ctx context.Context) (*api.SyncResponse, error) {
	lastToken, err := m.storage.GetSyncToken()
	if err != nil {
		return nil, fmt.Errorf("failed to get last sync token: %w", err)
	}

	resp, err := m.apiClient.Sync(ctx, &api.SyncRequest{
		SyncToken:     lastToken,
		ResourceTypes: []string{api.ResourceItems, api.ResourceProjects, api.ResourceSections},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to fetch incremental data: %w", err)
	}

	return resp, nil
}

// hasNoChanges は同期レスポンスに変更がないかチェックする
func (m *Manager) hasNoChanges(resp *api.SyncResponse) bool {
	return len(resp.Projects) == 0 && len(resp.Sections) == 0 && len(resp.Items) == 0
}

// applyIncrementalChanges はトランザクション内で差分変更を適用する
func (m *Manager) applyIncrementalChanges(resp *api.SyncResponse) error {
	tx, err := m.storage.BeginTx()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		if err != nil {
			if rollbackErr := tx.Rollback(); rollbackErr != nil {
				fmt.Printf("Warning: failed to rollback transaction: %v\n", rollbackErr)
			}
		}
	}()

	if err := m.applyProjectChanges(resp.Projects); err != nil {
		return err
	}

	if err := m.applySectionChanges(resp.Sections); err != nil {
		return err
	}

	if err := m.applyTaskChanges(resp.Items); err != nil {
		return err
	}

	if err := m.updateSyncMetadata(resp.SyncToken); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// applyProjectChanges はプロジェクトの変更を適用する
func (m *Manager) applyProjectChanges(projects []api.Project) error {
	if len(projects) == 0 {
		return nil
	}

	if m.verbose {
		fmt.Printf("📁 Processing %d project changes...\n", len(projects))
	}

	for _, project := range projects {
		if project.IsDeleted {
			if err := m.storage.DeleteProject(project.ID); err != nil {
				return fmt.Errorf("failed to delete project %s: %w", project.ID, err)
			}
		} else {
			if err := m.storage.InsertProject(project); err != nil {
				return fmt.Errorf("failed to upsert project %s: %w", project.ID, err)
			}
		}
	}

	return nil
}

// applySectionChanges はセクションの変更を適用する
func (m *Manager) applySectionChanges(sections []api.Section) error {
	if len(sections) == 0 {
		return nil
	}

	if m.verbose {
		fmt.Printf("📂 Processing %d section changes...\n", len(sections))
	}

	for _, section := range sections {
		if section.IsDeleted {
			if err := m.storage.DeleteSection(section.ID); err != nil {
				return fmt.Errorf("failed to delete section %s: %w", section.ID, err)
			}
		} else {
			if err := m.storage.InsertSection(section); err != nil {
				return fmt.Errorf("failed to upsert section %s: %w", section.ID, err)
			}
		}
	}

	return nil
}

// applyTaskChanges はタスクの変更を適用する
func (m *Manager) applyTaskChanges(tasks []api.Item) error {
	if len(tasks) == 0 {
		return nil
	}

	if m.verbose {
		fmt.Printf("📝 Processing %d task changes...\n", len(tasks))
	}

	for _, task := range tasks {
		if task.IsDeleted {
			if err := m.storage.DeleteTask(task.ID); err != nil {
				return fmt.Errorf("failed to delete task %s: %w", task.ID, err)
			}
		} else {
			if err := m.storage.InsertTask(task); err != nil {
				return fmt.Errorf("failed to upsert task %s: %w", task.ID, err)
			}
		}
	}

	return nil
}

// updateSyncMetadata はsync_tokenと同期時刻を更新する
func (m *Manager) updateSyncMetadata(syncToken string) error {
	if err := m.storage.SetSyncToken(syncToken); err != nil {
		return fmt.Errorf("failed to set sync token: %w", err)
	}

	if err := m.storage.SetLastSyncTime(time.Now()); err != nil {
		return fmt.Errorf("failed to set last sync time: %w", err)
	}

	return nil
}

// GetSyncStatus は同期状態の情報を返す
func (m *Manager) GetSyncStatus() (*Status, error) {
	initialDone, err := m.storage.IsInitialSyncDone()
	if err != nil {
		return nil, fmt.Errorf("failed to check initial sync status: %w", err)
	}

	lastSync, err := m.storage.GetLastSyncTime()
	if err != nil {
		lastSync = time.Time{} // ゼロ値
	}

	syncToken, err := m.storage.GetSyncToken()
	if err != nil {
		syncToken = "*"
	}

	return &Status{
		InitialSyncDone: initialDone,
		LastSyncTime:    lastSync,
		SyncToken:       syncToken,
	}, nil
}

// ForceInitialSync は強制的に初期同期を実行する
func (m *Manager) ForceInitialSync(ctx context.Context) error {
	if m.verbose {
		fmt.Println("🔄 Starting forced initial sync...")
	}
	return m.InitialSync(ctx)
}
