package cmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/kyokomi/gotodoist/internal/config"
	"github.com/kyokomi/gotodoist/internal/repository"
	"github.com/kyokomi/gotodoist/internal/sync"
)

func init() {
	// サブコマンドを追加
	syncCmd.AddCommand(syncInitCmd)
	syncCmd.AddCommand(syncStatusCmd)

	// syncコマンドをルートコマンドに追加
	rootCmd.AddCommand(syncCmd)
}

// syncCmd は同期関連のコマンド
var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Synchronize local data with Todoist API",
	Long: `Synchronize your local database with Todoist API to get the latest tasks, projects, and sections.
	
This command is useful for:
- Initial setup of local storage
- Getting latest data from Todoist
- Refreshing local cache`,
	RunE: runSync,
}

// syncInitCmd は初期同期コマンド
var syncInitCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize local storage with full sync",
	Long: `Perform initial synchronization to populate local storage with all your Todoist data.
	
This command will:
- Download all tasks, projects, and sections
- Set up local SQLite database
- Mark initial sync as completed`,
	RunE: runSyncInit,
}

// syncStatusCmd は同期状態確認コマンド
var syncStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show synchronization status",
	Long:  `Display current synchronization status including last sync time and data counts.`,
	RunE:  runSyncStatus,
}

// runSync は増分同期の実際の処理
func runSync(_ *cobra.Command, _ []string) error {
	ctx := context.Background()

	// 1. セットアップ
	executor, err := setupSyncExecution(ctx)
	if err != nil {
		return err
	}
	defer executor.cleanup()

	// 2. ローカルストレージの確認
	if !executor.isLocalStorageEnabled() {
		return fmt.Errorf("local storage is disabled. Enable it in config to use sync command")
	}

	// 3. 増分同期を実行
	status, err := executor.executeIncrementalSync(ctx)
	if err != nil {
		return fmt.Errorf("failed to sync: %w", err)
	}

	// 4. 結果表示
	displaySyncResult(status)

	return nil
}

// runSyncInit は初期同期の実際の処理
func runSyncInit(_ *cobra.Command, _ []string) error {
	ctx := context.Background()

	// 1. セットアップ
	executor, err := setupSyncExecution(ctx)
	if err != nil {
		return err
	}
	defer executor.cleanup()

	// 2. ローカルストレージの確認
	if !executor.isLocalStorageEnabled() {
		return fmt.Errorf("local storage is disabled. Enable it in config to use sync command")
	}

	// 3. 初期同期を実行
	fmt.Println("🔄 Starting initial synchronization...")
	status, err := executor.executeInitialSync(ctx)
	if err != nil {
		return fmt.Errorf("failed to run initial sync: %w", err)
	}

	// 4. 結果表示
	displayInitialSyncResult(status)

	return nil
}

// runSyncStatus は同期状態表示の実際の処理
func runSyncStatus(_ *cobra.Command, _ []string) error {
	ctx := context.Background()

	// 1. セットアップ
	executor, err := setupSyncExecution(ctx)
	if err != nil {
		return err
	}
	defer executor.cleanup()

	// 2. ローカルストレージの確認
	if !executor.isLocalStorageEnabled() {
		displayLocalStorageDisabled()
		return nil
	}

	// 3. 同期状態を取得
	status, err := executor.getSyncStatus()
	if err != nil {
		return fmt.Errorf("failed to get sync status: %w", err)
	}

	// 4. 結果表示
	displaySyncStatus(status)

	return nil
}

// displaySyncResult は同期結果を表示する
func displaySyncResult(status *sync.Status) {
	fmt.Println("✅ Synchronization completed successfully!")
	if status != nil {
		fmt.Printf("📊 %s\n", status.String())
	}
}

// displayInitialSyncResult は初期同期結果を表示する
func displayInitialSyncResult(status *sync.Status) {
	fmt.Println("✅ Initial synchronization completed successfully!")
	if status != nil {
		fmt.Printf("📊 %s\n", status.String())
	}
}

// displaySyncStatus は同期状態を表示する
func displaySyncStatus(status *sync.Status) {
	fmt.Printf("📊 Synchronization Status:\n")
	fmt.Printf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
	fmt.Printf("%s\n", status.String())
	fmt.Printf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")

	if status.InitialSyncDone {
		fmt.Printf("💡 Use 'gotodoist sync' for incremental sync\n")
		fmt.Printf("💡 Use 'gotodoist sync init' for full resync\n")
	} else {
		fmt.Printf("⚠️  Initial sync has not been completed\n")
		fmt.Printf("💡 Use 'gotodoist sync init' to initialize local storage\n")
	}
}

// displayLocalStorageDisabled はローカルストレージ無効時のメッセージを表示する
func displayLocalStorageDisabled() {
	fmt.Println("📭 Local storage is disabled")
	fmt.Printf("   Enable it in %s to use local-first features\n", "~/.config/gotodoist/config.yaml")
}

// syncExecutor は同期実行に必要な情報をまとめた構造体
type syncExecutor struct {
	cfg        *config.Config
	repository *repository.Repository
}

// setupSyncExecution は同期実行環境をセットアップする
func setupSyncExecution(ctx context.Context) (*syncExecutor, error) {
	cfg, err := config.LoadConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	repo, err := cfg.NewRepository(verbose)
	if err != nil {
		return nil, fmt.Errorf("failed to create repository: %w", err)
	}

	// ローカルストレージが有効な場合のみ初期化
	if cfg.LocalStorage.Enabled {
		if err := repo.Initialize(ctx); err != nil {
			if closeErr := repo.Close(); closeErr != nil {
				fmt.Printf("Warning: failed to close repository after initialization error: %v\n", closeErr)
			}
			return nil, fmt.Errorf("failed to initialize repository: %w", err)
		}
	}

	return &syncExecutor{
		cfg:        cfg,
		repository: repo,
	}, nil
}

// cleanup はRepositoryのリソースクリーンアップを行う
func (e *syncExecutor) cleanup() {
	if err := e.repository.Close(); err != nil {
		fmt.Printf("Warning: failed to close repository: %v\n", err)
	}
}

// isLocalStorageEnabled はローカルストレージが有効かどうかを返す
func (e *syncExecutor) isLocalStorageEnabled() bool {
	return e.cfg.LocalStorage.Enabled
}

// executeIncrementalSync は増分同期を実行する
func (e *syncExecutor) executeIncrementalSync(ctx context.Context) (*sync.Status, error) {
	if err := e.repository.Sync(ctx); err != nil {
		return nil, err
	}

	return e.repository.GetSyncStatus()
}

// executeInitialSync は初期同期を実行する
func (e *syncExecutor) executeInitialSync(ctx context.Context) (*sync.Status, error) {
	if err := e.repository.ForceInitialSync(ctx); err != nil {
		return nil, err
	}

	return e.repository.GetSyncStatus()
}

// getSyncStatus は同期状態を取得する
func (e *syncExecutor) getSyncStatus() (*sync.Status, error) {
	return e.repository.GetSyncStatus()
}
