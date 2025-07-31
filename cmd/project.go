package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// projectCmd はプロジェクト関連のコマンド
var projectCmd = &cobra.Command{
	Use:   "project",
	Short: "Manage Todoist projects",
	Long:  `Manage your Todoist projects including listing, adding, updating, and deleting projects.`,
}

// projectListCmd はプロジェクト一覧表示コマンド
var projectListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all projects",
	Long:  `Display a list of all your Todoist projects.`,
	Run: func(cmd *cobra.Command, args []string) {
		// TODO: APIクライアント実装後に実際の処理を追加
		fmt.Println("Listing projects...")
	},
}

// projectAddCmd はプロジェクト追加コマンド
var projectAddCmd = &cobra.Command{
	Use:   "add [project name]",
	Short: "Add a new project",
	Long:  `Add a new project to your Todoist.`,
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		// TODO: APIクライアント実装後に実際の処理を追加
		projectName := args[0]
		fmt.Printf("Adding project: %s\n", projectName)
	},
}

// projectUpdateCmd はプロジェクト更新コマンド
var projectUpdateCmd = &cobra.Command{
	Use:   "update [project ID] [new name]",
	Short: "Update an existing project",
	Long:  `Update the name of an existing project.`,
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		// TODO: APIクライアント実装後に実際の処理を追加
		projectID := args[0]
		newName := args[1]
		fmt.Printf("Updating project %s: %s\n", projectID, newName)
	},
}

// projectDeleteCmd はプロジェクト削除コマンド
var projectDeleteCmd = &cobra.Command{
	Use:   "delete [project ID]",
	Short: "Delete a project",
	Long:  `Delete a project from your Todoist.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		// TODO: APIクライアント実装後に実際の処理を追加
		projectID := args[0]
		fmt.Printf("Deleting project: %s\n", projectID)
	},
}

// projectArchiveCmd はプロジェクトアーカイブコマンド
var projectArchiveCmd = &cobra.Command{
	Use:   "archive [project ID]",
	Short: "Archive a project",
	Long:  `Archive a project in your Todoist.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		// TODO: APIクライアント実装後に実際の処理を追加
		projectID := args[0]
		fmt.Printf("Archiving project: %s\n", projectID)
	},
}

func init() {
	// サブコマンドを追加
	projectCmd.AddCommand(projectListCmd)
	projectCmd.AddCommand(projectAddCmd)
	projectCmd.AddCommand(projectUpdateCmd)
	projectCmd.AddCommand(projectDeleteCmd)
	projectCmd.AddCommand(projectArchiveCmd)

	// プロジェクトコマンドをルートコマンドに追加
	rootCmd.AddCommand(projectCmd)

	// project list用のフラグ
	projectListCmd.Flags().BoolP("tree", "t", false, "show projects in tree structure")
	projectListCmd.Flags().BoolP("archived", "a", false, "show archived projects")
}
