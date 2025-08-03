package cmd

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/kyokomi/gotodoist/internal/api"
	"github.com/kyokomi/gotodoist/internal/cli"
	"github.com/kyokomi/gotodoist/internal/config"
	"github.com/kyokomi/gotodoist/internal/factory"
	"github.com/kyokomi/gotodoist/internal/repository"
)

func init() {
	// サブコマンドを追加
	taskCmd.AddCommand(taskListCmd)
	taskCmd.AddCommand(taskAddCmd)
	taskCmd.AddCommand(taskUpdateCmd)
	taskCmd.AddCommand(taskDeleteCmd)
	taskCmd.AddCommand(taskCompleteCmd)
	taskCmd.AddCommand(taskUncompleteCmd)

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

// taskUncompleteCmd はタスク未完了コマンド
var taskUncompleteCmd = &cobra.Command{
	Use:   "uncomplete [task ID]",
	Short: "Mark a task as uncompleted",
	Long:  `Mark a completed task as uncompleted in your Todoist.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runTaskUncomplete,
}

// taskListParams はタスクリスト実行のパラメータ
type taskListParams struct {
	projectFilter    string
	filterExpression string
	showAll          bool
}

// taskListData はタスクリスト実行で取得したデータ
type taskListData struct {
	tasks       []api.Item
	projectsMap map[string]string
	sectionsMap map[string]string
}

// getTaskListParams はコマンドフラグからパラメータを取得する
func getTaskListParams(cmd *cobra.Command) *taskListParams {
	projectFilter, _ := cmd.Flags().GetString("project")
	filterExpression, _ := cmd.Flags().GetString("filter")
	showAll, _ := cmd.Flags().GetBool("all")

	return &taskListParams{
		projectFilter:    projectFilter,
		filterExpression: filterExpression,
		showAll:          showAll,
	}
}

// runTaskList はタスク一覧表示の実際の処理
func runTaskList(cmd *cobra.Command, _ []string) error {
	ctx := createBaseContext()

	// セットアップ
	executor, err := setupTaskExecution(ctx)
	if err != nil {
		return err
	}
	defer executor.cleanup()

	// パラメータ取得と実行
	params := getTaskListParams(cmd)
	return executor.executeTaskListWithOutput(ctx, params)
}

// executeTaskListWithOutput はタスク一覧表示と結果表示を実行する（テスト可能）
func (e *taskExecutor) executeTaskListWithOutput(ctx context.Context, params *taskListParams) error {
	// 1. データ取得
	data, err := e.fetchAllTaskListData(ctx, params)
	if err != nil {
		return err
	}

	// 2. フィルタリング
	filteredTasks := applyTaskFilters(data.tasks, params)

	// 3. 出力
	e.displayTaskResults(data.projectsMap, data.sectionsMap, filteredTasks)

	return nil
}

// taskAddParams はタスク追加のパラメータ
type taskAddParams struct {
	content     string
	projectID   string
	priority    string
	dueDate     string
	description string
	labels      string
}

// getTaskAddParams はタスク追加のパラメータを取得する
func getTaskAddParams(cmd *cobra.Command, args []string) *taskAddParams {
	projectID, _ := cmd.Flags().GetString("project")
	priority, _ := cmd.Flags().GetString("priority")
	dueDate, _ := cmd.Flags().GetString("due")
	description, _ := cmd.Flags().GetString("description")
	labels, _ := cmd.Flags().GetString("labels")

	return &taskAddParams{
		content:     strings.Join(args, " "),
		projectID:   projectID,
		priority:    priority,
		dueDate:     dueDate,
		description: description,
		labels:      labels,
	}
}

// runTaskAdd はタスク追加の実際の処理
func runTaskAdd(cmd *cobra.Command, args []string) error {
	ctx := createBaseContext()

	// セットアップ
	executor, err := setupTaskExecution(ctx)
	if err != nil {
		return err
	}
	defer executor.cleanup()

	// パラメータ取得と実行
	params := getTaskAddParams(cmd, args)
	return executor.executeTaskAddWithOutput(ctx, params)
}

// executeTaskAddWithOutput はタスク追加と結果表示を実行する（テスト可能）
func (e *taskExecutor) executeTaskAddWithOutput(ctx context.Context, params *taskAddParams) error {
	// 1. タスク追加実行
	resp, err := e.executeTaskAdd(ctx, params)
	if err != nil {
		return fmt.Errorf("failed to create task: %w", err)
	}

	// 2. 結果表示
	e.displaySuccessMessage("Task created successfully!", resp.SyncToken)

	return nil
}

// taskCompleteParams はタスク完了のパラメータ
type taskCompleteParams struct {
	taskID string
}

// getTaskCompleteParams はタスク完了のパラメータを取得する
func getTaskCompleteParams(args []string) *taskCompleteParams {
	return &taskCompleteParams{
		taskID: args[0],
	}
}

// runTaskComplete はタスク完了の実際の処理
func runTaskComplete(_ *cobra.Command, args []string) error {
	ctx := createBaseContext()

	// セットアップ
	executor, err := setupTaskExecution(ctx)
	if err != nil {
		return err
	}
	defer executor.cleanup()

	// パラメータ取得と実行
	params := getTaskCompleteParams(args)
	return executor.executeTaskCompleteWithOutput(ctx, params)
}

// executeTaskCompleteWithOutput はタスク完了と結果表示を実行する（テスト可能）
func (e *taskExecutor) executeTaskCompleteWithOutput(ctx context.Context, params *taskCompleteParams) error {
	// 1. タスク完了実行
	resp, err := e.executeTaskComplete(ctx, params)
	if err != nil {
		return fmt.Errorf("failed to complete task: %w", err)
	}

	// 2. 結果表示
	e.displaySuccessMessage("Task completed successfully!", resp.SyncToken)

	return nil
}

// runTaskUncomplete はタスク未完了の実際の処理
func runTaskUncomplete(_ *cobra.Command, args []string) error {
	ctx := createBaseContext()

	// セットアップ
	executor, err := setupTaskExecution(ctx)
	if err != nil {
		return err
	}
	defer executor.cleanup()

	// パラメータ取得と実行
	params := getTaskCompleteParams(args) // 同じパラメータ構造を使用
	return executor.executeTaskUncompleteWithOutput(ctx, params)
}

// executeTaskUncompleteWithOutput はタスク未完了と結果表示を実行する（テスト可能）
func (e *taskExecutor) executeTaskUncompleteWithOutput(ctx context.Context, params *taskCompleteParams) error {
	// 1. タスク未完了実行
	resp, err := e.executeTaskUncomplete(ctx, params)
	if err != nil {
		return fmt.Errorf("failed to uncomplete task: %w", err)
	}

	// 2. 結果表示
	e.displaySuccessMessage("Task marked as uncompleted successfully!", resp.SyncToken)

	return nil
}

// taskDeleteParams はタスク削除のパラメータ
type taskDeleteParams struct {
	taskID string
	force  bool
}

// getTaskDeleteParams はタスク削除のパラメータを取得する
func getTaskDeleteParams(cmd *cobra.Command, args []string) *taskDeleteParams {
	force, _ := cmd.Flags().GetBool("force")
	return &taskDeleteParams{
		taskID: args[0],
		force:  force,
	}
}

// runTaskDelete はタスク削除の実際の処理
func runTaskDelete(cmd *cobra.Command, args []string) error {
	ctx := createBaseContext()

	// セットアップ
	executor, err := setupTaskExecution(ctx)
	if err != nil {
		return err
	}
	defer executor.cleanup()

	// パラメータ取得と実行
	params := getTaskDeleteParams(cmd, args)
	return executor.executeTaskDeleteWithOutput(ctx, params)
}

// executeTaskDeleteWithOutput はタスク削除と結果表示を実行する（テスト可能）
func (e *taskExecutor) executeTaskDeleteWithOutput(ctx context.Context, params *taskDeleteParams) error {
	// 1. 削除対象の確認
	task, shouldDelete, err := e.confirmTaskDeletion(ctx, params)
	if err != nil {
		return err
	}
	if !shouldDelete {
		return nil // ユーザーがキャンセルまたはタスクが見つからない
	}

	// 2. タスク削除実行
	resp, err := e.deleteTask(ctx, task.ID)
	if err != nil {
		return fmt.Errorf("failed to delete task: %w", err)
	}

	// 3. 結果表示
	e.displayTaskDeleteResult(task, resp)

	return nil
}

// taskUpdateParams はタスク更新のパラメータ
type taskUpdateParams struct {
	taskID      string
	content     string
	priority    string
	dueDate     string
	description string
	labels      string
}

// getTaskUpdateParams はタスク更新のパラメータを取得する
func getTaskUpdateParams(cmd *cobra.Command, args []string) *taskUpdateParams {
	content, _ := cmd.Flags().GetString("content")
	priority, _ := cmd.Flags().GetString("priority")
	dueDate, _ := cmd.Flags().GetString("due")
	description, _ := cmd.Flags().GetString("description")
	labels, _ := cmd.Flags().GetString("labels")

	return &taskUpdateParams{
		taskID:      args[0],
		content:     content,
		priority:    priority,
		dueDate:     dueDate,
		description: description,
		labels:      labels,
	}
}

// runTaskUpdate はタスク更新の実際の処理
func runTaskUpdate(cmd *cobra.Command, args []string) error {
	ctx := createBaseContext()

	// セットアップ
	executor, err := setupTaskExecution(ctx)
	if err != nil {
		return err
	}
	defer executor.cleanup()

	// パラメータ取得と実行
	params := getTaskUpdateParams(cmd, args)
	return executor.executeTaskUpdateWithOutput(ctx, params)
}

// executeTaskUpdateWithOutput はタスク更新と結果表示を実行する（テスト可能）
func (e *taskExecutor) executeTaskUpdateWithOutput(ctx context.Context, params *taskUpdateParams) error {
	// 1. タスク更新実行
	resp, err := e.executeTaskUpdate(ctx, params)
	if err != nil {
		return fmt.Errorf("failed to update task: %w", err)
	}

	// 2. 結果表示
	e.displaySuccessMessage("Task updated successfully!", resp.SyncToken)

	return nil
}

// applyTaskFilters はタスクにフィルタを適用する
func applyTaskFilters(tasks []api.Item, params *taskListParams) []api.Item {
	// アクティブタスクフィルタ
	tasks = filterActiveTasks(tasks, params.showAll)

	// フィルタ式による絞り込み
	if params.filterExpression != "" {
		tasks = filterTasks(tasks, params.filterExpression)
	}

	return tasks
}

// filterTasks は指定されたフィルタ式でタスクを絞り込む
func filterTasks(tasks []api.Item, filter string) []api.Item {
	filtered := make([]api.Item, 0, len(tasks))
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

// displayTaskResults はタスク結果を表示する
func (e *taskExecutor) displayTaskResults(projectsMap, sectionsMap map[string]string, tasks []api.Item) {
	if len(tasks) == 0 {
		e.output.Infof("📭 No tasks found")
		return
	}

	// タスクを表示
	e.output.Listf("Found %d task(s):", len(tasks))
	e.output.Plainf("")
	for i := range tasks {
		e.displayTask(&tasks[i], projectsMap, sectionsMap)
	}
}

// displayTask はタスクを表示する
func (e *taskExecutor) displayTask(task *api.Item, projects map[string]string, sections map[string]string) {
	priorityIcon := getPriorityIcon(task.Priority)

	// セクション名を取得
	sectionName := ""
	if task.SectionID != "" {
		if name, exists := sections[task.SectionID]; exists {
			sectionName = fmt.Sprintf(" [%s]", name)
		}
	}

	e.output.Plainf("%s %s%s", priorityIcon, task.Content, sectionName)

	if IsVerbose() {
		e.output.Plainf("   ID: %s", task.ID)
		projectName, exists := projects[task.ProjectID]
		if exists {
			e.output.Plainf("   Project: %s (%s)", projectName, task.ProjectID)
		} else {
			e.output.Plainf("   Project: %s", task.ProjectID)
		}
		if task.Due != nil {
			e.output.Plainf("   Due: %s", task.Due.String)
		}
		if len(task.Labels) > 0 {
			e.output.Plainf("   Labels: %s", strings.Join(task.Labels, ", "))
		}
		if !task.DateAdded.IsZero() {
			e.output.Plainf("   Created: %s", task.DateAdded.Format("2006-01-02 15:04"))
		} else {
			e.output.Plainf("   Created: Unknown")
		}
	}

	if task.Description != "" && IsVerbose() {
		e.output.Plainf("   Description: %s", task.Description)
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

// filterActiveTasks は完了済みタスクを除外する
func filterActiveTasks(tasks []api.Item, showAll bool) []api.Item {
	if showAll {
		return tasks
	}

	activeTasks := make([]api.Item, 0, len(tasks))
	for i := range tasks {
		if tasks[i].DateCompleted == nil {
			activeTasks = append(activeTasks, tasks[i])
		}
	}
	return activeTasks
}

// displaySuccessMessage は共通の成功メッセージを表示する
func (e *taskExecutor) displaySuccessMessage(message string, syncToken string) {
	e.output.Successf("%s", message)
	if IsVerbose() && syncToken != "" {
		e.output.Plainf("Sync token: %s", syncToken)
	}
}

// displayTaskDeleteResult はタスク削除結果を表示する
func (e *taskExecutor) displayTaskDeleteResult(task *api.Item, resp *api.SyncResponse) {
	if task == nil || resp == nil {
		return // キャンセルまたはタスクが見つからない場合
	}

	e.output.Successf("🗑️  Task deleted successfully!")
	e.output.Infof("    Deleted: %s", task.Content)
	if IsVerbose() {
		e.output.Plainf("    Sync token: %s", resp.SyncToken)
	}
}

// taskExecutor はタスク実行に必要な情報をまとめた構造体
type taskExecutor struct {
	cfg        *config.Config
	repository *repository.Repository
	output     *cli.Output
}

// setupTaskExecution はタスク実行環境をセットアップする
func setupTaskExecution(ctx context.Context) (*taskExecutor, error) {
	cfg, err := config.LoadConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	output := cli.New(IsVerbose())

	repo, err := factory.NewRepository(cfg, IsVerbose())
	if err != nil {
		return nil, fmt.Errorf("failed to create Repository: %w", err)
	}

	// Repositoryの初期化
	if err := repo.Initialize(ctx); err != nil {
		if closeErr := repo.Close(); closeErr != nil {
			output.Warningf("failed to close repository after initialization error: %v", closeErr)
		}
		return nil, fmt.Errorf("failed to initialize repository: %w", err)
	}

	return &taskExecutor{
		cfg:        cfg,
		repository: repo,
		output:     output,
	}, nil
}

// cleanup はRepositoryのリソースクリーンアップを行う
func (e *taskExecutor) cleanup() {
	if err := e.repository.Close(); err != nil {
		e.output.Warningf("failed to close repository: %v", err)
	}
}

// buildProjectsMap はプロジェクトマップを構築する
func (e *taskExecutor) buildProjectsMap(ctx context.Context, verbose bool) map[string]string {
	if !verbose {
		return nil
	}

	projects, err := e.repository.GetAllProjects(ctx)
	if err != nil {
		// プロジェクト情報の取得に失敗してもタスク表示は続行
		e.output.Warningf("Failed to load project names: %v", err)
		return make(map[string]string)
	}

	projectsMap := make(map[string]string)
	for _, project := range projects {
		projectsMap[project.ID] = project.Name
	}
	return projectsMap
}

// buildSectionsMap はセクションマップを構築する
func (e *taskExecutor) buildSectionsMap(ctx context.Context) map[string]string {
	sections, err := e.repository.GetAllSections(ctx)
	if err != nil {
		// セクション情報の取得に失敗してもタスク表示は続行
		e.output.Warningf("Failed to load section names: %v", err)
		return make(map[string]string)
	}

	sectionsMap := make(map[string]string)
	for _, section := range sections {
		sectionsMap[section.ID] = section.Name
	}
	return sectionsMap
}

// findProjectIDByName はプロジェクト検索を実行する（Repository層に移植済み）
func (e *taskExecutor) findProjectIDByName(ctx context.Context, nameOrID string) (string, error) {
	return e.repository.FindProjectIDByName(ctx, nameOrID)
}

// fetchAllTaskListData は必要なデータを全て取得する
func (e *taskExecutor) fetchAllTaskListData(ctx context.Context, params *taskListParams) (*taskListData, error) {
	repo := e.repository

	// プロジェクト情報を取得（ローカル優先）
	projectsMap := e.buildProjectsMap(ctx, IsVerbose())

	// セクション情報を取得（ローカル優先）
	sectionsMap := e.buildSectionsMap(ctx)

	// タスクデータを取得
	var tasks []api.Item
	var err error

	if params.projectFilter != "" {
		// プロジェクト指定がある場合
		projectID, err := e.findProjectIDByName(ctx, params.projectFilter)
		if err != nil {
			return nil, fmt.Errorf("failed to find project: %w", err)
		}
		tasks, err = repo.GetTasksByProject(ctx, projectID)
		if err != nil {
			return nil, fmt.Errorf("failed to get tasks: %w", err)
		}
	} else {
		// 全タスクを取得（ローカル優先）
		tasks, err = repo.GetTasks(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to get tasks: %w", err)
		}
	}

	return &taskListData{
		tasks:       tasks,
		projectsMap: projectsMap,
		sectionsMap: sectionsMap,
	}, nil
}

// executeTaskAdd はタスク追加を実行する
func (e *taskExecutor) executeTaskAdd(ctx context.Context, params *taskAddParams) (*api.SyncResponse, error) {
	repo := e.repository

	// リクエストを構築
	req := &api.CreateTaskRequest{
		Content:     params.content,
		Description: params.description,
	}

	if params.projectID != "" {
		// プロジェクト名からIDを解決
		resolvedProjectID, err := e.findProjectIDByName(ctx, params.projectID)
		if err != nil {
			return nil, fmt.Errorf("failed to find project: %w", err)
		}
		req.ProjectID = resolvedProjectID
	}

	if params.priority != "" {
		priority, err := strconv.Atoi(params.priority)
		if err != nil {
			return nil, fmt.Errorf("invalid priority: %s", params.priority)
		}
		if priority < 1 || priority > 4 {
			return nil, fmt.Errorf("priority must be between 1 and 4")
		}
		req.Priority = priority
	}

	if params.dueDate != "" {
		req.DueString = params.dueDate
	}

	if params.labels != "" {
		labels := strings.Split(params.labels, ",")
		for i, label := range labels {
			labels[i] = strings.TrimSpace(label)
		}
		req.Labels = labels
	}

	// タスクを作成
	return repo.CreateTask(ctx, req)
}

// executeTaskComplete はタスク完了を実行する
func (e *taskExecutor) executeTaskComplete(ctx context.Context, params *taskCompleteParams) (*api.SyncResponse, error) {
	repo := e.repository
	return repo.CloseTask(ctx, params.taskID)
}

// executeTaskUncomplete はタスク未完了を実行する
func (e *taskExecutor) executeTaskUncomplete(ctx context.Context, params *taskCompleteParams) (*api.SyncResponse, error) {
	repo := e.repository
	return repo.ReopenTask(ctx, params.taskID)
}

// findTaskByID はタスクIDからタスクを検索する
func (e *taskExecutor) findTaskByID(ctx context.Context, taskID string) (*api.Item, error) {
	repo := e.repository
	tasks, err := repo.GetTasks(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get tasks: %w", err)
	}

	for i := range tasks {
		if tasks[i].ID == taskID {
			return &tasks[i], nil
		}
	}
	return nil, nil // タスクが見つからない場合
}

// confirmTaskDeletion は削除対象タスクの確認を行う
func (e *taskExecutor) confirmTaskDeletion(ctx context.Context, params *taskDeleteParams) (*api.Item, bool, error) {
	// タスクの存在確認
	targetTask, err := e.findTaskByID(ctx, params.taskID)
	if err != nil {
		return nil, false, err
	}

	if targetTask == nil {
		e.output.Errorf("Task with ID '%s' not found.", params.taskID)
		e.output.Plainf("")
		e.output.Infof("💡 To find the correct task ID, use one of these commands:")
		e.output.Infof("   gotodoist task list -v                    # Show all tasks with IDs")
		e.output.Infof("   gotodoist task list -v -f \"keyword\"       # Search tasks containing 'keyword'")
		e.output.Infof("   gotodoist task list -v -p \"project name\"  # Show tasks in specific project")
		return nil, false, nil // エラーではなく、単にタスクが見つからない
	}

	// 確認処理（forceフラグが無い場合）
	if !params.force {
		if !e.promptTaskDeletionConfirmation(targetTask) {
			return nil, false, nil // キャンセルされた
		}
	}

	return targetTask, true, nil
}

// deleteTask はタスクを削除する
func (e *taskExecutor) deleteTask(ctx context.Context, taskID string) (*api.SyncResponse, error) {
	return e.repository.DeleteTask(ctx, taskID)
}

// executeTaskUpdate はタスク更新を実行する
func (e *taskExecutor) executeTaskUpdate(ctx context.Context, params *taskUpdateParams) (*api.SyncResponse, error) {
	// リクエストを構築
	req, err := e.buildUpdateTaskRequest(params)
	if err != nil {
		return nil, err
	}

	// タスクを更新
	repo := e.repository
	return repo.UpdateTask(ctx, params.taskID, req)
}

// buildUpdateTaskRequest はタスク更新リクエストを構築する
func (e *taskExecutor) buildUpdateTaskRequest(params *taskUpdateParams) (*api.UpdateTaskRequest, error) {
	// 何も更新内容がない場合はエラー
	if params.content == "" && params.priority == "" && params.dueDate == "" &&
		params.description == "" && params.labels == "" {
		return nil, fmt.Errorf("at least one update field must be specified (--content, --priority, --due, --description, --labels)")
	}

	// リクエストを構築
	req := &api.UpdateTaskRequest{}

	if params.content != "" {
		req.Content = params.content
	}

	if params.description != "" {
		req.Description = params.description
	}

	if params.priority != "" {
		priority, err := strconv.Atoi(params.priority)
		if err != nil {
			return nil, fmt.Errorf("invalid priority: %s", params.priority)
		}
		if priority < 1 || priority > 4 {
			return nil, fmt.Errorf("priority must be between 1 and 4")
		}
		req.Priority = priority
	}

	if params.dueDate != "" {
		req.DueString = params.dueDate
	}

	if params.labels != "" {
		labels := strings.Split(params.labels, ",")
		for i, label := range labels {
			labels[i] = strings.TrimSpace(label)
		}
		req.Labels = labels
	}

	return req, nil
}

// promptTaskDeletionConfirmation はタスク削除の確認プロンプトを表示する
func (e *taskExecutor) promptTaskDeletionConfirmation(task *api.Item) bool {
	e.output.Warningf("Are you sure you want to delete this task? (y/N)")
	e.output.Plainf("    ID: %s", task.ID)
	e.output.Plainf("    Content: %s", task.Content)
	if task.Description != "" {
		e.output.Plainf("    Description: %s", task.Description)
	}
	if len(task.Labels) > 0 {
		e.output.Plainf("    Labels: %s", strings.Join(task.Labels, ", "))
	}
	e.output.PlainNoNewlinef("Enter your choice: ")

	var confirmation string
	_, err := fmt.Scanln(&confirmation)
	if err != nil {
		e.output.Errorf("Task deletion canceled")
		return false
	}
	if confirmation != "y" && confirmation != "Y" {
		e.output.Errorf("Task deletion canceled")
		return false
	}
	return true
}
