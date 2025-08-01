package cmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/kyokomi/gotodoist/internal/config"
)

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
	cfg, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	if !cfg.LocalStorage.Enabled {
		return fmt.Errorf("local storage is disabled. Enable it in config to use sync command")
	}

	repository, err := cfg.NewRepository(verbose)
	if err != nil {
		return fmt.Errorf("failed to create repository: %w", err)
	}
	defer func() {
		if err := repository.Close(); err != nil {
			fmt.Printf("Warning: failed to close repository: %v\n", err)
		}
	}()

	ctx := context.Background()

	// 増分同期を実行
	if err := repository.Sync(ctx); err != nil {
		return fmt.Errorf("failed to sync: %w", err)
	}

	fmt.Println("✅ Synchronization completed successfully!")

	// 同期状態を表示
	status, err := repository.GetSyncStatus()
	if err != nil {
		return fmt.Errorf("failed to get sync status: %w", err)
	}
	fmt.Printf("📊 %s\n", status.String())

	return nil
}

// runSyncInit は初期同期の実際の処理
func runSyncInit(_ *cobra.Command, _ []string) error {
	cfg, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	if !cfg.LocalStorage.Enabled {
		return fmt.Errorf("local storage is disabled. Enable it in config to use sync command")
	}

	repository, err := cfg.NewRepository(verbose)
	if err != nil {
		return fmt.Errorf("failed to create repository: %w", err)
	}
	defer func() {
		if err := repository.Close(); err != nil {
			fmt.Printf("Warning: failed to close repository: %v\n", err)
		}
	}()

	ctx := context.Background()

	// 強制的に初期同期を実行
	fmt.Println("🔄 Starting initial synchronization...")
	if err := repository.ForceInitialSync(ctx); err != nil {
		return fmt.Errorf("failed to run initial sync: %w", err)
	}

	fmt.Println("✅ Initial synchronization completed successfully!")

	// 同期状態を表示
	status, err := repository.GetSyncStatus()
	if err != nil {
		return fmt.Errorf("failed to get sync status: %w", err)
	}
	fmt.Printf("📊 %s\n", status.String())

	return nil
}

// runSyncStatus は同期状態表示の実際の処理
func runSyncStatus(_ *cobra.Command, _ []string) error {
	cfg, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	if !cfg.LocalStorage.Enabled {
		fmt.Println("📭 Local storage is disabled")
		fmt.Printf("   Enable it in %s to use local-first features\n", "~/.config/gotodoist/config.yaml")
		return nil
	}

	repository, err := cfg.NewRepository(verbose)
	if err != nil {
		return fmt.Errorf("failed to create repository: %w", err)
	}
	defer func() {
		if err := repository.Close(); err != nil {
			fmt.Printf("Warning: failed to close repository: %v\n", err)
		}
	}()

	// 同期状態を取得（初期化せずに直接取得）
	status, err := repository.GetSyncStatus()
	if err != nil {
		return fmt.Errorf("failed to get sync status: %w", err)
	}

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

	return nil
}

func init() {
	// サブコマンドを追加
	syncCmd.AddCommand(syncInitCmd)
	syncCmd.AddCommand(syncStatusCmd)

	// syncコマンドをルートコマンドに追加
	rootCmd.AddCommand(syncCmd)
}
