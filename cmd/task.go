package cmd

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/kyokomi/gotodoist/internal/api"
	"github.com/kyokomi/gotodoist/internal/config"
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
	RunE: func(cmd *cobra.Command, args []string) error {
		return runTaskList(cmd, args)
	},
}

// taskAddCmd はタスク追加コマンド
var taskAddCmd = &cobra.Command{
	Use:   "add [task content]",
	Short: "Add a new task",
	Long:  `Add a new task to your Todoist.`,
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return runTaskAdd(cmd, args)
	},
}

// taskUpdateCmd はタスク更新コマンド
var taskUpdateCmd = &cobra.Command{
	Use:   "update [task ID]",
	Short: "Update an existing task",
	Long:  `Update the content of an existing task.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return runTaskUpdate(cmd, args)
	},
}

// taskDeleteCmd はタスク削除コマンド
var taskDeleteCmd = &cobra.Command{
	Use:   "delete [task ID]",
	Short: "Delete a task",
	Long:  `Delete a task from your Todoist.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return runTaskDelete(cmd, args)
	},
}

// taskCompleteCmd はタスク完了コマンド
var taskCompleteCmd = &cobra.Command{
	Use:   "complete [task ID]",
	Short: "Mark a task as completed",
	Long:  `Mark a task as completed in your Todoist.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return runTaskComplete(cmd, args)
	},
}

// runTaskList はタスク一覧表示の実際の処理
func runTaskList(cmd *cobra.Command, args []string) error {
	cfg, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	client, err := cfg.NewAPIClient()
	if err != nil {
		return fmt.Errorf("failed to create API client: %w", err)
	}

	ctx := context.Background()

	// フラグから設定を取得
	projectFilter, _ := cmd.Flags().GetString("project")
	showAll, _ := cmd.Flags().GetBool("all")

	var tasks []api.Item
	if projectFilter != "" {
		// プロジェクト指定がある場合
		tasks, err = client.GetTasksByProject(ctx, projectFilter)
	} else {
		// 全タスクを取得
		tasks, err = client.GetTasks(ctx)
	}

	if err != nil {
		return fmt.Errorf("failed to get tasks: %w", err)
	}

	// フィルタリング
	if !showAll {
		// 完了済みタスクを除外（実際には削除済みタスクは既に除外されている）
		var activeTasks []api.Item
		for _, task := range tasks {
			if task.DateCompleted == nil {
				activeTasks = append(activeTasks, task)
			}
		}
		tasks = activeTasks
	}

	if len(tasks) == 0 {
		fmt.Println("📭 No tasks found")
		return nil
	}

	// タスクを表示
	fmt.Printf("📝 Found %d task(s):\n\n", len(tasks))
	for i, task := range tasks {
		displayTask(task, i+1)
	}

	return nil
}

// displayTask はタスクを表示する
func displayTask(task api.Item, index int) {
	priorityIcon := getPriorityIcon(task.Priority)
	
	fmt.Printf("%d. %s %s\n", index, priorityIcon, task.Content)
	
	if verbose {
		fmt.Printf("   ID: %s\n", task.ID)
		fmt.Printf("   Project: %s\n", task.ProjectID)
		if task.Due != nil {
			fmt.Printf("   Due: %s\n", task.Due.String)
		}
		if len(task.Labels) > 0 {
			fmt.Printf("   Labels: %s\n", strings.Join(task.Labels, ", "))
		}
		fmt.Printf("   Created: %s\n", task.DateAdded.Format("2006-01-02 15:04"))
	}
	
	if task.Description != "" && verbose {
		fmt.Printf("   Description: %s\n", task.Description)
	}
	
	fmt.Println()
}

// getPriorityIcon は優先度に応じたアイコンを返す
func getPriorityIcon(priority int) string {
	switch priority {
	case 4:
		return "🔴" // Urgent
	case 3:
		return "🟡" // Very High
	case 2:
		return "🟢" // High
	default:
		return "⚪" // Normal
	}
}

// runTaskAdd はタスク追加の実際の処理
func runTaskAdd(cmd *cobra.Command, args []string) error {
	cfg, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	client, err := cfg.NewAPIClient()
	if err != nil {
		return fmt.Errorf("failed to create API client: %w", err)
	}

	ctx := context.Background()

	// フラグから設定を取得
	projectID, _ := cmd.Flags().GetString("project")
	priorityStr, _ := cmd.Flags().GetString("priority")
	dueDate, _ := cmd.Flags().GetString("due")
	description, _ := cmd.Flags().GetString("description")
	labelsStr, _ := cmd.Flags().GetString("labels")

	// タスク内容を結合
	content := strings.Join(args, " ")

	// リクエストを構築
	req := &api.CreateTaskRequest{
		Content:     content,
		Description: description,
	}

	if projectID != "" {
		req.ProjectID = projectID
	}

	if priorityStr != "" {
		priority, err := strconv.Atoi(priorityStr)
		if err != nil {
			return fmt.Errorf("invalid priority: %s", priorityStr)
		}
		if priority < 1 || priority > 4 {
			return fmt.Errorf("priority must be between 1 and 4")
		}
		req.Priority = priority
	}

	if dueDate != "" {
		req.DueString = dueDate
	}

	if labelsStr != "" {
		labels := strings.Split(labelsStr, ",")
		for i, label := range labels {
			labels[i] = strings.TrimSpace(label)
		}
		req.Labels = labels
	}

	// タスクを作成
	resp, err := client.CreateTask(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to create task: %w", err)
	}

	fmt.Printf("✅ Task created successfully!\n")
	if verbose {
		fmt.Printf("Sync token: %s\n", resp.SyncToken)
	}

	return nil
}

// runTaskComplete はタスク完了の実際の処理
func runTaskComplete(cmd *cobra.Command, args []string) error {
	cfg, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	client, err := cfg.NewAPIClient()
	if err != nil {
		return fmt.Errorf("failed to create API client: %w", err)
	}

	ctx := context.Background()
	taskID := args[0]

	// タスクを完了にする
	resp, err := client.CloseTask(ctx, taskID)
	if err != nil {
		return fmt.Errorf("failed to complete task: %w", err)
	}

	fmt.Printf("✅ Task completed successfully!\n")
	if verbose {
		fmt.Printf("Sync token: %s\n", resp.SyncToken)
	}

	return nil
}

// runTaskDelete はタスク削除の実際の処理
func runTaskDelete(cmd *cobra.Command, args []string) error {
	cfg, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	client, err := cfg.NewAPIClient()
	if err != nil {
		return fmt.Errorf("failed to create API client: %w", err)
	}

	ctx := context.Background()
	taskID := args[0]

	// 確認フラグをチェック
	force, _ := cmd.Flags().GetBool("force")
	if !force {
		fmt.Printf("⚠️  Are you sure you want to delete task %s? (y/N): ", taskID)
		var confirmation string
		fmt.Scanln(&confirmation)
		if confirmation != "y" && confirmation != "Y" {
			fmt.Println("❌ Task deletion cancelled")
			return nil
		}
	}

	// タスクを削除する
	resp, err := client.DeleteTask(ctx, taskID)
	if err != nil {
		return fmt.Errorf("failed to delete task: %w", err)
	}

	fmt.Printf("🗑️  Task deleted successfully!\n")
	if verbose {
		fmt.Printf("Sync token: %s\n", resp.SyncToken)
	}

	return nil
}

// runTaskUpdate はタスク更新の実際の処理
func runTaskUpdate(cmd *cobra.Command, args []string) error {
	cfg, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	client, err := cfg.NewAPIClient()
	if err != nil {
		return fmt.Errorf("failed to create API client: %w", err)
	}

	ctx := context.Background()
	taskID := args[0]

	// フラグから設定を取得
	content, _ := cmd.Flags().GetString("content")
	priorityStr, _ := cmd.Flags().GetString("priority")
	dueDate, _ := cmd.Flags().GetString("due")
	description, _ := cmd.Flags().GetString("description")
	labelsStr, _ := cmd.Flags().GetString("labels")

	// 何も更新内容がない場合はエラー
	if content == "" && priorityStr == "" && dueDate == "" && description == "" && labelsStr == "" {
		return fmt.Errorf("at least one update field must be specified (--content, --priority, --due, --description, --labels)")
	}

	// リクエストを構築
	req := &api.UpdateTaskRequest{}

	if content != "" {
		req.Content = content
	}

	if description != "" {
		req.Description = description
	}

	if priorityStr != "" {
		priority, err := strconv.Atoi(priorityStr)
		if err != nil {
			return fmt.Errorf("invalid priority: %s", priorityStr)
		}
		if priority < 1 || priority > 4 {
			return fmt.Errorf("priority must be between 1 and 4")
		}
		req.Priority = priority
	}

	if dueDate != "" {
		req.DueString = dueDate
	}

	if labelsStr != "" {
		labels := strings.Split(labelsStr, ",")
		for i, label := range labels {
			labels[i] = strings.TrimSpace(label)
		}
		req.Labels = labels
	}

	// タスクを更新
	resp, err := client.UpdateTask(ctx, taskID, req)
	if err != nil {
		return fmt.Errorf("failed to update task: %w", err)
	}

	fmt.Printf("✏️  Task updated successfully!\n")
	if verbose {
		fmt.Printf("Sync token: %s\n", resp.SyncToken)
	}

	return nil
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

	// task add用のフラグ
	taskAddCmd.Flags().StringP("project", "p", "", "project ID to add task to")
	taskAddCmd.Flags().StringP("priority", "P", "", "task priority (1-4)")
	taskAddCmd.Flags().StringP("due", "d", "", "due date (e.g., 'today', 'tomorrow', '2024-12-25')")
	taskAddCmd.Flags().StringP("description", "D", "", "task description")
	taskAddCmd.Flags().StringP("labels", "l", "", "comma-separated labels")

	// task update用のフラグ
	taskUpdateCmd.Flags().StringP("content", "c", "", "new task content")
	taskUpdateCmd.Flags().StringP("priority", "P", "", "task priority (1-4)")
	taskUpdateCmd.Flags().StringP("due", "d", "", "due date (e.g., 'today', 'tomorrow', '2024-12-25')")
	taskUpdateCmd.Flags().StringP("description", "D", "", "task description")
	taskUpdateCmd.Flags().StringP("labels", "l", "", "comma-separated labels")

	// task delete用のフラグ
	taskDeleteCmd.Flags().BoolP("force", "f", false, "skip confirmation prompt")
}
