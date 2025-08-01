package cmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/kyokomi/gotodoist/internal/cli"
	"github.com/kyokomi/gotodoist/internal/config"
	"github.com/kyokomi/gotodoist/internal/factory"
	"github.com/kyokomi/gotodoist/internal/repository"
	"github.com/kyokomi/gotodoist/internal/sync"
)

func init() {
	// サブコマンドを追加
	syncCmd.AddCommand(syncInitCmd)
	syncCmd.AddCommand(syncStatusCmd)
	syncCmd.AddCommand(syncResetCmd)

	// syncコマンドをルートコマンドに追加
	rootCmd.AddCommand(syncCmd)

	// sync reset用のフラグ
	syncResetCmd.Flags().BoolP("force", "f", false, "skip confirmation prompt")
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

// syncResetCmd はローカルデータリセットコマンド
var syncResetCmd = &cobra.Command{
	Use:   "reset",
	Short: "Reset local storage and clear all cached data",
	Long: `Reset local storage by clearing all cached data including tasks, projects, and sections.
	
This command is useful for:
- Fixing data corruption issues
- Switching to a different API token
- Starting fresh with a clean local database
- Resolving sync conflicts

After reset, you may want to run 'gotodoist sync init' to repopulate the local storage.`,
	RunE: runSyncReset,
}

// runSync は増分同期の実際の処理
func runSync(_ *cobra.Command, _ []string) error {
	ctx := createBaseContext()

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
	executor.displaySyncResult(status)

	return nil
}

// runSyncInit は初期同期の実際の処理
func runSyncInit(_ *cobra.Command, _ []string) error {
	ctx := createBaseContext()

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
	executor.output.Syncf("Starting initial synchronization...")
	status, err := executor.executeInitialSync(ctx)
	if err != nil {
		return fmt.Errorf("failed to run initial sync: %w", err)
	}

	// 4. 結果表示
	executor.displayInitialSyncResult(status)

	return nil
}

// runSyncStatus は同期状態表示の実際の処理
func runSyncStatus(_ *cobra.Command, _ []string) error {
	ctx := createBaseContext()

	// 1. セットアップ
	executor, err := setupSyncExecution(ctx)
	if err != nil {
		return err
	}
	defer executor.cleanup()

	// 2. ローカルストレージの確認
	if !executor.isLocalStorageEnabled() {
		executor.displayLocalStorageDisabled()
		return nil
	}

	// 3. 同期状態を取得
	status, err := executor.getSyncStatus()
	if err != nil {
		return fmt.Errorf("failed to get sync status: %w", err)
	}

	// 4. 結果表示
	executor.displaySyncStatus(status)

	return nil
}

// displaySyncResult は同期結果を表示する
func (e *syncExecutor) displaySyncResult(status *sync.Status) {
	e.output.Successf("Synchronization completed successfully!")
	if status != nil {
		e.output.Infof("📊 %s", status.String())
	}
}

// displayInitialSyncResult は初期同期結果を表示する
func (e *syncExecutor) displayInitialSyncResult(status *sync.Status) {
	e.output.Successf("Initial synchronization completed successfully!")
	if status != nil {
		e.output.Infof("📊 %s", status.String())
	}
}

// displaySyncStatus は同期状態を表示する
func (e *syncExecutor) displaySyncStatus(status *sync.Status) {
	e.output.Infof("📊 Synchronization Status:")
	e.output.Plainf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	e.output.Plainf("%s", status.String())
	e.output.Plainf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	if status.InitialSyncDone {
		e.output.Infof("💡 Use 'gotodoist sync' for incremental sync")
		e.output.Infof("💡 Use 'gotodoist sync init' for full resync")
	} else {
		e.output.Warningf("Initial sync has not been completed")
		e.output.Infof("💡 Use 'gotodoist sync init' to initialize local storage")
	}
}

// displayLocalStorageDisabled はローカルストレージ無効時のメッセージを表示する
func (e *syncExecutor) displayLocalStorageDisabled() {
	e.output.Infof("📭 Local storage is disabled")
	e.output.Infof("   Enable it in %s to use local-first features", "~/.config/gotodoist/config.yaml")
}

// syncExecutor は同期実行に必要な情報をまとめた構造体
type syncExecutor struct {
	cfg        *config.Config
	repository *repository.Repository
	output     *cli.Output
}

// setupSyncExecution は同期実行環境をセットアップする
func setupSyncExecution(ctx context.Context) (*syncExecutor, error) {
	cfg, err := config.LoadConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	output := cli.New(IsVerbose())

	repo, err := factory.NewRepository(cfg, IsVerbose())
	if err != nil {
		return nil, fmt.Errorf("failed to create repository: %w", err)
	}

	// ローカルストレージが有効な場合のみ初期化
	if cfg.LocalStorage.Enabled {
		if err := repo.Initialize(ctx); err != nil {
			if closeErr := repo.Close(); closeErr != nil {
				output.Warningf("failed to close repository after initialization error: %v", closeErr)
			}
			return nil, fmt.Errorf("failed to initialize repository: %w", err)
		}
	}

	return &syncExecutor{
		cfg:        cfg,
		repository: repo,
		output:     output,
	}, nil
}

// cleanup はRepositoryのリソースクリーンアップを行う
func (e *syncExecutor) cleanup() {
	if err := e.repository.Close(); err != nil {
		e.output.Warningf("failed to close repository: %v", err)
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

// runSyncReset はローカルデータリセットの実際の処理
func runSyncReset(cmd *cobra.Command, _ []string) error {
	ctx := createBaseContext()

	// 1. セットアップ
	executor, err := setupSyncExecution(ctx)
	if err != nil {
		return err
	}
	defer executor.cleanup()

	// 2. ローカルストレージの確認
	if !executor.isLocalStorageEnabled() {
		return fmt.Errorf("local storage is disabled. Enable it in config to use reset command")
	}

	// 3. 確認プロンプト（forceフラグが無い場合）
	force, _ := cmd.Flags().GetBool("force")
	if !force {
		if !executor.promptResetConfirmation() {
			return nil // ユーザーがキャンセル
		}
	}

	// 4. リセット実行
	if err := executor.executeReset(ctx); err != nil {
		return fmt.Errorf("failed to reset local storage: %w", err)
	}

	// 5. 結果表示
	executor.displayResetResult()

	return nil
}

// promptResetConfirmation はリセットの確認プロンプトを表示する
func (e *syncExecutor) promptResetConfirmation() bool {
	e.output.Warningf("⚠️  WARNING: This will delete ALL local cached data!")
	e.output.Plainf("")
	e.output.Plainf("This includes:")
	e.output.Plainf("  • All cached tasks")
	e.output.Plainf("  • All cached projects")
	e.output.Plainf("  • All cached sections")
	e.output.Plainf("  • Sync status and tokens")
	e.output.Plainf("")
	e.output.Plainf("Your data in Todoist cloud will NOT be affected.")
	e.output.Plainf("You can repopulate local storage by running 'gotodoist sync init'.")
	e.output.Plainf("")
	e.output.PlainNoNewlinef("Are you sure you want to reset local storage? (y/N): ")

	var confirmation string
	_, err := fmt.Scanln(&confirmation)
	if err != nil {
		e.output.Errorf("Reset canceled")
		return false
	}

	if confirmation != "y" && confirmation != "Y" {
		e.output.Errorf("Reset canceled")
		return false
	}

	return true
}

// executeReset はローカルストレージのリセットを実行する
func (e *syncExecutor) executeReset(ctx context.Context) error {
	// Repositoryにリセットメソッドがあるかチェック、なければ直接ストレージを操作
	// この実装では、新しいメソッドをRepositoryに追加する必要がある
	return e.repository.ResetLocalStorage(ctx)
}

// displayResetResult はリセット結果を表示する
func (e *syncExecutor) displayResetResult() {
	e.output.Successf("🗑️  Local storage reset completed!")
	e.output.Plainf("")
	e.output.Infof("💡 Next steps:")
	e.output.Plainf("  • Run 'gotodoist sync init' to repopulate local storage")
	e.output.Plainf("  • Or use commands directly (they will fetch from API)")
}
