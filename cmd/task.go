package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// taskCmd はタスク関連のコマンド
var taskCmd = &cobra.Command{
	Use:   "task",
	Short: "Manage Todoist tasks",
	Long:  `Manage your Todoist tasks including listing, adding, updating, and deleting tasks.`,
}

// taskListCmd はタスク一覧表示コマンド
var taskListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all tasks",
	Long:  `Display a list of all your Todoist tasks.`,
	Run: func(cmd *cobra.Command, args []string) {
		// TODO: APIクライアント実装後に実際の処理を追加
		fmt.Println("Listing tasks...")
		if verbose {
			fmt.Println("Verbose mode enabled")
		}
		if debug {
			fmt.Println("Debug mode enabled")
		}
	},
}

// taskAddCmd はタスク追加コマンド
var taskAddCmd = &cobra.Command{
	Use:   "add [task content]",
	Short: "Add a new task",
	Long:  `Add a new task to your Todoist.`,
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		// TODO: APIクライアント実装後に実際の処理を追加
		taskContent := args[0]
		fmt.Printf("Adding task: %s\n", taskContent)
	},
}

// taskUpdateCmd はタスク更新コマンド
var taskUpdateCmd = &cobra.Command{
	Use:   "update [task ID] [new content]",
	Short: "Update an existing task",
	Long:  `Update the content of an existing task.`,
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		// TODO: APIクライアント実装後に実際の処理を追加
		taskID := args[0]
		newContent := args[1]
		fmt.Printf("Updating task %s: %s\n", taskID, newContent)
	},
}

// taskDeleteCmd はタスク削除コマンド
var taskDeleteCmd = &cobra.Command{
	Use:   "delete [task ID]",
	Short: "Delete a task",
	Long:  `Delete a task from your Todoist.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		// TODO: APIクライアント実装後に実際の処理を追加
		taskID := args[0]
		fmt.Printf("Deleting task: %s\n", taskID)
	},
}

// taskCompleteCmd はタスク完了コマンド
var taskCompleteCmd = &cobra.Command{
	Use:   "complete [task ID]",
	Short: "Mark a task as completed",
	Long:  `Mark a task as completed in your Todoist.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		// TODO: APIクライアント実装後に実際の処理を追加
		taskID := args[0]
		fmt.Printf("Completing task: %s\n", taskID)
	},
}

func init() {
	// サブコマンドを追加
	taskCmd.AddCommand(taskListCmd)
	taskCmd.AddCommand(taskAddCmd)
	taskCmd.AddCommand(taskUpdateCmd)
	taskCmd.AddCommand(taskDeleteCmd)
	taskCmd.AddCommand(taskCompleteCmd)

	// タスクコマンドをルートコマンドに追加
	rootCmd.AddCommand(taskCmd)

	// task list用のフラグ
	taskListCmd.Flags().StringP("project", "p", "", "filter by project")
	taskListCmd.Flags().StringP("filter", "f", "", "filter expression")
	taskListCmd.Flags().BoolP("all", "a", false, "show all tasks including completed")
}
