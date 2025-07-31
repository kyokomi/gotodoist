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
	filterExpression, _ := cmd.Flags().GetString("filter")
	showAll, _ := cmd.Flags().GetBool("all")

	// プロジェクト情報を取得（verbose表示用）
	var projectsMap map[string]string
	if verbose {
		projects, err := client.GetAllProjects(ctx)
		if err != nil {
			// プロジェクト情報の取得に失敗してもタスク表示は続行
			fmt.Printf("Warning: Failed to load project names: %v\n", err)
			projectsMap = make(map[string]string)
		} else {
			projectsMap = make(map[string]string)
			for _, project := range projects {
				projectsMap[project.ID] = project.Name
			}
		}
	}

	var tasks []api.Item
	if projectFilter != "" {
		// プロジェクト指定がある場合
		// まずプロジェクト名で検索を試み、見つからなければIDとして扱う
		projectID, err := findProjectIDByName(ctx, client, projectFilter)
		if err != nil {
			return fmt.Errorf("failed to find project: %w", err)
		}
		tasks, err = client.GetTasksByProject(ctx, projectID)
		if err != nil {
			return fmt.Errorf("failed to get tasks: %w", err)
		}
	} else {
		// 全タスクを取得
		tasks, err = client.GetTasks(ctx)
		if err != nil {
			return fmt.Errorf("failed to get tasks: %w", err)
		}
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

	// フィルタ式による絞り込み
	if filterExpression != "" {
		filteredTasks := filterTasks(tasks, filterExpression)
		tasks = filteredTasks
	}

	if len(tasks) == 0 {
		fmt.Println("📭 No tasks found")
		return nil
	}

	// タスクを表示
	fmt.Printf("📝 Found %d task(s):\n\n", len(tasks))
	for i, task := range tasks {
		displayTask(task, i+1, projectsMap)
	}

	return nil
}

// displayTask はタスクを表示する
func displayTask(task api.Item, index int, projects map[string]string) {
	priorityIcon := getPriorityIcon(task.Priority)

	fmt.Printf("%d. %s %s\n", index, priorityIcon, task.Content)

	if verbose {
		fmt.Printf("   ID: %s\n", task.ID)
		projectName, exists := projects[task.ProjectID]
		if exists {
			fmt.Printf("   Project: %s (%s)\n", projectName, task.ProjectID)
		} else {
			fmt.Printf("   Project: %s\n", task.ProjectID)
		}
		if task.Due != nil {
			fmt.Printf("   Due: %s\n", task.Due.String)
		}
		if len(task.Labels) > 0 {
			fmt.Printf("   Labels: %s\n", strings.Join(task.Labels, ", "))
		}
		if !task.DateAdded.IsZero() {
			fmt.Printf("   Created: %s\n", task.DateAdded.Format("2006-01-02 15:04"))
		} else {
			fmt.Printf("   Created: Unknown\n")
		}
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

// filterTasks は指定されたフィルタ式でタスクを絞り込む
func filterTasks(tasks []api.Item, filter string) []api.Item {
	var filtered []api.Item
	filter = strings.ToLower(filter)

	for _, task := range tasks {
		if matchesFilter(task, filter) {
			filtered = append(filtered, task)
		}
	}

	return filtered
}

// matchesFilter はタスクがフィルタ条件にマッチするかチェック
func matchesFilter(task api.Item, filter string) bool {
	// 基本的なキーワード検索
	content := strings.ToLower(task.Content)
	description := strings.ToLower(task.Description)

	// 特別なフィルタ
	switch {
	case strings.HasPrefix(filter, "p1"):
		return task.Priority == 1
	case strings.HasPrefix(filter, "p2"):
		return task.Priority == 2
	case strings.HasPrefix(filter, "p3"):
		return task.Priority == 3
	case strings.HasPrefix(filter, "p4"):
		return task.Priority == 4
	case strings.HasPrefix(filter, "today"):
		return task.Due != nil && strings.Contains(strings.ToLower(task.Due.String), "today")
	case strings.HasPrefix(filter, "tomorrow"):
		return task.Due != nil && strings.Contains(strings.ToLower(task.Due.String), "tomorrow")
	case strings.HasPrefix(filter, "overdue"):
		return task.Due != nil && strings.Contains(strings.ToLower(task.Due.String), "overdue")
	case strings.HasPrefix(filter, "@"):
		// ラベル検索
		label := strings.TrimPrefix(filter, "@")
		for _, taskLabel := range task.Labels {
			if strings.Contains(strings.ToLower(taskLabel), label) {
				return true
			}
		}
		return false
	default:
		// 一般的なキーワード検索（内容と説明をチェック）
		return strings.Contains(content, filter) || strings.Contains(description, filter)
	}
}

// findProjectIDByName はプロジェクト名からIDを検索する
func findProjectIDByName(ctx context.Context, client *api.Client, nameOrID string) (string, error) {
	// まず全プロジェクトを取得
	projects, err := client.GetAllProjects(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get projects: %w", err)
	}

	nameOrID = strings.ToLower(nameOrID)

	// 完全一致で検索
	for _, project := range projects {
		if strings.ToLower(project.Name) == nameOrID {
			return project.ID, nil
		}
	}

	// 部分一致で検索
	for _, project := range projects {
		if strings.Contains(strings.ToLower(project.Name), nameOrID) {
			return project.ID, nil
		}
	}

	// IDとして直接指定されている可能性をチェック
	for _, project := range projects {
		if project.ID == nameOrID {
			return project.ID, nil
		}
	}

	return "", fmt.Errorf("project not found: %s", nameOrID)
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
		// プロジェクト名からIDを解決
		resolvedProjectID, err := findProjectIDByName(ctx, client, projectID)
		if err != nil {
			return fmt.Errorf("failed to find project: %w", err)
		}
		req.ProjectID = resolvedProjectID
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

	// タスクの存在確認
	tasks, err := client.GetTasks(ctx)
	if err != nil {
		return fmt.Errorf("failed to get tasks: %w", err)
	}

	var targetTask *api.Item
	for i := range tasks {
		if tasks[i].ID == taskID {
			targetTask = &tasks[i]
			break
		}
	}

	if targetTask == nil {
		fmt.Printf("❌ Task with ID '%s' not found.\n\n", taskID)
		fmt.Printf("💡 To find the correct task ID, use one of these commands:\n")
		fmt.Printf("   gotodoist task list -v                    # Show all tasks with IDs\n")
		fmt.Printf("   gotodoist task list -v -f \"keyword\"       # Search tasks containing 'keyword'\n")
		fmt.Printf("   gotodoist task list -v -p \"project name\"  # Show tasks in specific project\n")
		return nil
	}

	// 確認フラグをチェック
	force, _ := cmd.Flags().GetBool("force")
	if !force {
		fmt.Printf("⚠️  Are you sure you want to delete this task? (y/N)\n")
		fmt.Printf("    ID: %s\n", targetTask.ID)
		fmt.Printf("    Content: %s\n", targetTask.Content)
		if targetTask.Description != "" {
			fmt.Printf("    Description: %s\n", targetTask.Description)
		}
		if len(targetTask.Labels) > 0 {
			fmt.Printf("    Labels: %s\n", strings.Join(targetTask.Labels, ", "))
		}
		fmt.Printf("Enter your choice: ")

		var confirmation string
		_, err := fmt.Scanln(&confirmation)
		if err != nil {
			fmt.Println("❌ Task deletion canceled")
			return nil
		}
		if confirmation != "y" && confirmation != "Y" {
			fmt.Println("❌ Task deletion canceled")
			return nil
		}
	}

	// タスクを削除する
	resp, err := client.DeleteTask(ctx, taskID)
	if err != nil {
		return fmt.Errorf("failed to delete task: %w", err)
	}

	fmt.Printf("🗑️  Task deleted successfully!\n")
	fmt.Printf("    Deleted: %s\n", targetTask.Content)
	if verbose {
		fmt.Printf("    Sync token: %s\n", resp.SyncToken)
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
	taskListCmd.Flags().StringP("project", "p", "", "filter by project name or ID")
	taskListCmd.Flags().StringP("filter", "f", "", "filter expression (p1-p4 for priority, @label for labels, keywords for content)")
	taskListCmd.Flags().BoolP("all", "a", false, "show all tasks including completed")

	// task add用のフラグ
	taskAddCmd.Flags().StringP("project", "p", "", "project name or ID to add task to")
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
