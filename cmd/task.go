package cmd

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/kyokomi/gotodoist/internal/api"
	"github.com/kyokomi/gotodoist/internal/benchmark"
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
	RunE:  runTaskList,
}

// taskAddCmd はタスク追加コマンド
var taskAddCmd = &cobra.Command{
	Use:   "add [task content]",
	Short: "Add a new task",
	Long:  `Add a new task to your Todoist.`,
	Args:  cobra.MinimumNArgs(1),
	RunE:  runTaskAdd,
}

// taskUpdateCmd はタスク更新コマンド
var taskUpdateCmd = &cobra.Command{
	Use:   "update [task ID]",
	Short: "Update an existing task",
	Long:  `Update the content of an existing task.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runTaskUpdate,
}

// taskDeleteCmd はタスク削除コマンド
var taskDeleteCmd = &cobra.Command{
	Use:   "delete [task ID]",
	Short: "Delete a task",
	Long:  `Delete a task from your Todoist.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runTaskDelete,
}

// taskCompleteCmd はタスク完了コマンド
var taskCompleteCmd = &cobra.Command{
	Use:   "complete [task ID]",
	Short: "Mark a task as completed",
	Long:  `Mark a task as completed in your Todoist.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runTaskComplete,
}

// runTaskList はタスク一覧表示の実際の処理
func runTaskList(cmd *cobra.Command, _ []string) error {
	// ベンチマークタイマーを開始
	timer := benchmark.NewTimer(showBenchmark)

	cfg, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}
	timer.Step("Config loaded")

	// ローカルファーストクライアントを作成
	client, err := cfg.NewLocalFirstClient(verbose)
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}
	defer client.Close()
	timer.Step("Client created")

	ctx := context.Background()

	// クライアントを初期化（必要に応じて初期同期）
	if err := client.Initialize(ctx); err != nil {
		return fmt.Errorf("failed to initialize client: %w", err)
	}
	timer.Step("Client initialized (inc. sync check)")

	// フラグから設定を取得
	projectFilter, _ := cmd.Flags().GetString("project")
	filterExpression, _ := cmd.Flags().GetString("filter")
	showAll, _ := cmd.Flags().GetBool("all")
	compare, _ := cmd.Flags().GetBool("compare")

	if compare {
		return runTaskListComparison(projectFilter, filterExpression, showAll)
	}

	// プロジェクト情報を取得（ローカル優先）
	projectsMap := buildProjectsMapLocal(ctx, client, verbose)
	timer.Step("Projects loaded")

	// セクション情報を取得（ローカル優先）
	sectionsMap := buildSectionsMapLocal(ctx, client)
	timer.Step("Sections loaded")

	var tasks []api.Item
	if projectFilter != "" {
		// プロジェクト指定がある場合
		// まずプロジェクト名で検索を試み、見つからなければIDとして扱う
		projectID, err := findProjectIDByNameLocal(ctx, client, projectFilter)
		if err != nil {
			return fmt.Errorf("failed to find project: %w", err)
		}
		tasks, err = client.GetTasksByProject(ctx, projectID)
		if err != nil {
			return fmt.Errorf("failed to get tasks: %w", err)
		}
	} else {
		// 全タスクを取得（ローカル優先）
		tasks, err = client.GetTasks(ctx)
		if err != nil {
			return fmt.Errorf("failed to get tasks: %w", err)
		}
	}
	timer.Step("Tasks loaded")

	// フィルタリング
	tasks = filterActiveTasks(tasks, showAll)

	// フィルタ式による絞り込み
	if filterExpression != "" {
		filteredTasks := filterTasks(tasks, filterExpression)
		tasks = filteredTasks
	}
	timer.Step("Tasks filtered")

	if len(tasks) == 0 {
		fmt.Println("📭 No tasks found")
		timer.PrintResults()
		return nil
	}

	// タスクを表示
	fmt.Printf("📝 Found %d task(s):\n\n", len(tasks))
	for i := range tasks {
		displayTask(&tasks[i], projectsMap, sectionsMap)
	}
	timer.Step("Tasks displayed")

	// ベンチマーク結果を表示
	timer.PrintResults()

	return nil
}

// runTaskListComparison はローカルファーストとAPIの性能比較
func runTaskListComparison(projectFilter, filterExpression string, showAll bool) error {
	fmt.Println("🔍 Performance Comparison: Local-First vs API Direct")
	fmt.Println(strings.Repeat("=", 60))

	// 1. ローカルファーストクライアントでの実行
	fmt.Println("\n📦 Local-First Client:")
	localTimer := benchmark.NewTimer(true)

	cfg, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}
	localTimer.Step("Config loaded")

	localClient, err := cfg.NewLocalFirstClient(verbose)
	if err != nil {
		return fmt.Errorf("failed to create local client: %w", err)
	}
	defer localClient.Close()
	localTimer.Step("Local client created")

	ctx := context.Background()
	if err := localClient.Initialize(ctx); err != nil {
		return fmt.Errorf("failed to initialize local client: %w", err)
	}
	localTimer.Step("Local client initialized")

	var localTasks []api.Item
	if projectFilter != "" {
		projectID, err := findProjectIDByNameLocal(ctx, localClient, projectFilter)
		if err != nil {
			return fmt.Errorf("failed to find project: %w", err)
		}
		localTasks, err = localClient.GetTasksByProject(ctx, projectID)
		if err != nil {
			return fmt.Errorf("failed to get tasks: %w", err)
		}
	} else {
		localTasks, err = localClient.GetTasks(ctx)
		if err != nil {
			return fmt.Errorf("failed to get tasks: %w", err)
		}
	}
	localTimer.Step("Tasks loaded from local")

	localTasks = filterActiveTasks(localTasks, showAll)
	if filterExpression != "" {
		localTasks = filterTasks(localTasks, filterExpression)
	}
	localTimer.Step("Tasks filtered")

	localDuration := localTimer.GetTotalDuration()

	// 2. APIクライアントでの実行
	fmt.Println("\n🌐 API Client:")
	apiTimer := benchmark.NewTimer(true)

	apiClient, err := cfg.NewAPIClient()
	if err != nil {
		return fmt.Errorf("failed to create API client: %w", err)
	}
	apiTimer.Step("API client created")

	var apiTasks []api.Item
	if projectFilter != "" {
		projectID, err := findProjectIDByName(ctx, apiClient, projectFilter)
		if err != nil {
			return fmt.Errorf("failed to find project: %w", err)
		}
		apiTasks, err = apiClient.GetTasksByProject(ctx, projectID)
		if err != nil {
			return fmt.Errorf("failed to get tasks: %w", err)
		}
	} else {
		apiTasks, err = apiClient.GetTasks(ctx)
		if err != nil {
			return fmt.Errorf("failed to get tasks: %w", err)
		}
	}
	apiTimer.Step("Tasks loaded from API")

	apiTasks = filterActiveTasks(apiTasks, showAll)
	if filterExpression != "" {
		apiTasks = filterTasks(apiTasks, filterExpression)
	}
	apiTimer.Step("Tasks filtered")

	apiDuration := apiTimer.GetTotalDuration()

	// 結果の表示
	localTimer.PrintResults()
	apiTimer.PrintResults()

	// 比較結果
	fmt.Printf("📊 Comparison Results:\n")
	fmt.Printf("%s\n", strings.Repeat("─", 50))
	fmt.Printf("Local-First:  %s (%d tasks)\n", benchmark.FormatDuration(localDuration), len(localTasks))
	fmt.Printf("API Direct:   %s (%d tasks)\n", benchmark.FormatDuration(apiDuration), len(apiTasks))

	if localDuration < apiDuration {
		speedup := float64(apiDuration) / float64(localDuration)
		fmt.Printf("Speed-up:     %.1fx faster with Local-First! 🚀\n", speedup)
	} else {
		slowdown := float64(localDuration) / float64(apiDuration)
		fmt.Printf("Speed-down:   %.1fx slower with Local-First 😅\n", slowdown)
	}
	fmt.Printf("%s\n", strings.Repeat("─", 50))

	return nil
}

// displayTask はタスクを表示する
func displayTask(task *api.Item, projects map[string]string, sections map[string]string) {
	priorityIcon := getPriorityIcon(task.Priority)

	// セクション名を取得
	sectionName := ""
	if task.SectionID != "" {
		if name, exists := sections[task.SectionID]; exists {
			sectionName = fmt.Sprintf(" [%s]", name)
		}
	}

	fmt.Printf("%s %s%s\n", priorityIcon, task.Content, sectionName)

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
}

// getPriorityIcon は優先度に応じたアイコンを返す
func getPriorityIcon(priority int) string {
	switch priority {
	case int(api.PriorityUrgent):
		return "🔴" // Urgent
	case int(api.PriorityVeryHigh):
		return "🟡" // Very High
	case int(api.PriorityHigh):
		return "🟢" // High
	default:
		return "⚪" // Normal
	}
}

// filterTasks は指定されたフィルタ式でタスクを絞り込む
func filterTasks(tasks []api.Item, filter string) []api.Item {
	var filtered []api.Item
	filter = strings.ToLower(filter)

	for i := range tasks {
		if matchesFilter(&tasks[i], filter) {
			filtered = append(filtered, tasks[i])
		}
	}

	return filtered
}

// matchesFilter はタスクがフィルタ条件にマッチするかチェック
func matchesFilter(task *api.Item, filter string) bool {
	// 基本的なキーワード検索
	content := strings.ToLower(task.Content)
	description := strings.ToLower(task.Description)

	// 特別なフィルタ
	switch {
	case strings.HasPrefix(filter, "p1"):
		return task.Priority == int(api.PriorityNormal)
	case strings.HasPrefix(filter, "p2"):
		return task.Priority == int(api.PriorityHigh)
	case strings.HasPrefix(filter, "p3"):
		return task.Priority == int(api.PriorityVeryHigh)
	case strings.HasPrefix(filter, "p4"):
		return task.Priority == int(api.PriorityUrgent)
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
		if strings.EqualFold(project.Name, nameOrID) {
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
func runTaskComplete(_ *cobra.Command, args []string) error {
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

	// フラグからリクエストを構築
	req, err := buildUpdateTaskRequestFromFlags(cmd)
	if err != nil {
		return err
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

// buildUpdateTaskRequestFromFlags はフラグからUpdateTaskRequestを構築する
func buildUpdateTaskRequestFromFlags(cmd *cobra.Command) (*api.UpdateTaskRequest, error) {
	// フラグから設定を取得
	content, _ := cmd.Flags().GetString("content")
	priorityStr, _ := cmd.Flags().GetString("priority")
	dueDate, _ := cmd.Flags().GetString("due")
	description, _ := cmd.Flags().GetString("description")
	labelsStr, _ := cmd.Flags().GetString("labels")

	// 何も更新内容がない場合はエラー
	if content == "" && priorityStr == "" && dueDate == "" && description == "" && labelsStr == "" {
		return nil, fmt.Errorf("at least one update field must be specified (--content, --priority, --due, --description, --labels)")
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
			return nil, fmt.Errorf("invalid priority: %s", priorityStr)
		}
		if priority < 1 || priority > 4 {
			return nil, fmt.Errorf("priority must be between 1 and 4")
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

	return req, nil
}

// buildProjectsMapLocal はローカルファーストクライアント用のプロジェクトマップを構築する
func buildProjectsMapLocal(ctx context.Context, client interface {
	GetAllProjects(ctx context.Context) ([]api.Project, error)
}, verbose bool) map[string]string {
	if !verbose {
		return nil
	}

	projects, err := client.GetAllProjects(ctx)
	if err != nil {
		// プロジェクト情報の取得に失敗してもタスク表示は続行
		fmt.Printf("Warning: Failed to load project names: %v\n", err)
		return make(map[string]string)
	}

	projectsMap := make(map[string]string)
	for _, project := range projects {
		projectsMap[project.ID] = project.Name
	}
	return projectsMap
}

// buildSectionsMapLocal はローカルファーストクライアント用のセクションマップを構築する
func buildSectionsMapLocal(ctx context.Context, client interface {
	GetAllSections(ctx context.Context) ([]api.Section, error)
}) map[string]string {
	sections, err := client.GetAllSections(ctx)
	if err != nil {
		// セクション情報の取得に失敗してもタスク表示は続行
		fmt.Printf("Warning: Failed to load section names: %v\n", err)
		return make(map[string]string)
	}

	sectionsMap := make(map[string]string)
	for _, section := range sections {
		sectionsMap[section.ID] = section.Name
	}
	return sectionsMap
}

// findProjectIDByNameLocal はローカルファーストクライアント用のプロジェクト検索
func findProjectIDByNameLocal(ctx context.Context, client interface {
	GetAllProjects(ctx context.Context) ([]api.Project, error)
}, nameOrID string) (string, error) {
	// 全プロジェクトを取得
	projects, err := client.GetAllProjects(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get projects: %w", err)
	}

	nameOrID = strings.ToLower(nameOrID)

	// 完全一致で検索
	for _, project := range projects {
		if strings.EqualFold(project.Name, nameOrID) {
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

// filterActiveTasks は完了済みタスクを除外する
func filterActiveTasks(tasks []api.Item, showAll bool) []api.Item {
	if showAll {
		return tasks
	}

	var activeTasks []api.Item
	for i := range tasks {
		if tasks[i].DateCompleted == nil {
			activeTasks = append(activeTasks, tasks[i])
		}
	}
	return activeTasks
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
	taskListCmd.Flags().Bool("compare", false, "compare local-first vs API performance")

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
