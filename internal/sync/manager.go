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
			tx.Rollback()
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
	// 初期同期が完了しているかチェック
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

	if m.verbose {
		fmt.Println("🔄 Starting incremental sync...")
	}

	// 前回のsync_tokenを取得
	lastToken, err := m.storage.GetSyncToken()
	if err != nil {
		return fmt.Errorf("failed to get last sync token: %w", err)
	}

	// 差分データを取得
	resp, err := m.apiClient.Sync(ctx, &api.SyncRequest{
		SyncToken:     lastToken,
		ResourceTypes: []string{api.ResourceItems, api.ResourceProjects, api.ResourceSections},
	})
	if err != nil {
		return fmt.Errorf("failed to fetch incremental data: %w", err)
	}

	// 変更がない場合は何もしない
	if len(resp.Projects) == 0 && len(resp.Sections) == 0 && len(resp.Items) == 0 {
		if m.verbose {
			fmt.Println("📭 No changes since last sync")
		}
		return nil
	}

	// トランザクション内で差分を適用
	tx, err := m.storage.BeginTx()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	// プロジェクトの差分を適用
	if len(resp.Projects) > 0 {
		if m.verbose {
			fmt.Printf("📁 Processing %d project changes...\n", len(resp.Projects))
		}
		for _, project := range resp.Projects {
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
	}

	// セクションの差分を適用
	if len(resp.Sections) > 0 {
		if m.verbose {
			fmt.Printf("📂 Processing %d section changes...\n", len(resp.Sections))
		}
		for _, section := range resp.Sections {
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
	}

	// タスクの差分を適用
	if len(resp.Items) > 0 {
		if m.verbose {
			fmt.Printf("📝 Processing %d task changes...\n", len(resp.Items))
		}
		for _, task := range resp.Items {
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
	}

	// sync_tokenと同期時刻を更新
	if err := m.storage.SetSyncToken(resp.SyncToken); err != nil {
		return fmt.Errorf("failed to set sync token: %w", err)
	}

	if err := m.storage.SetLastSyncTime(time.Now()); err != nil {
		return fmt.Errorf("failed to set last sync time: %w", err)
	}

	// トランザクションコミット
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	if m.verbose {
		fmt.Println("✅ Incremental sync completed successfully!")
	}

	return nil
}

// ShouldSync は同期が必要かどうかを判定する
func (m *Manager) ShouldSync(interval time.Duration) (bool, error) {
	lastSync, err := m.storage.GetLastSyncTime()
	if err != nil {
		// 同期時刻が取得できない場合は同期が必要
		return true, nil
	}

	return time.Since(lastSync) > interval, nil
}

// AutoSync は自動同期を実行する（必要に応じて）
func (m *Manager) AutoSync(ctx context.Context, interval time.Duration) error {
	shouldSync, err := m.ShouldSync(interval)
	if err != nil {
		return fmt.Errorf("failed to check if sync needed: %w", err)
	}

	if !shouldSync {
		if m.verbose {
			lastSync, _ := m.storage.GetLastSyncTime()
			fmt.Printf("⏰ Sync not needed (last: %s)\n", lastSync.Format("15:04:05"))
		}
		return nil
	}

	return m.IncrementalSync(ctx)
}

// GetSyncStatus は同期状態の情報を返す
func (m *Manager) GetSyncStatus() (*SyncStatus, error) {
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

	return &SyncStatus{
		InitialSyncDone: initialDone,
		LastSyncTime:    lastSync,
		SyncToken:       syncToken,
	}, nil
}

// SyncStatus は同期状態を表す
type SyncStatus struct {
	InitialSyncDone bool      `json:"initial_sync_done"`
	LastSyncTime    time.Time `json:"last_sync_time"`
	SyncToken       string    `json:"sync_token"`
}

// String は同期状態を文字列として表現する
func (s *SyncStatus) String() string {
	status := "❌ Not initialized"
	if s.InitialSyncDone {
		if s.LastSyncTime.IsZero() {
			status = "✅ Initialized (never synced)"
		} else {
			status = fmt.Sprintf("✅ Last sync: %s", s.LastSyncTime.Format("2006-01-02 15:04:05"))
		}
	}
	return fmt.Sprintf("Sync Status: %s (token: %s)", status, s.SyncToken[:8]+"...")
}
