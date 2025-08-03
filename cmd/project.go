package cmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/kyokomi/gotodoist/internal/api"
	"github.com/kyokomi/gotodoist/internal/cli"
	"github.com/kyokomi/gotodoist/internal/config"
	"github.com/kyokomi/gotodoist/internal/factory"
	"github.com/kyokomi/gotodoist/internal/repository"
)

const (
	iconFolder = "📁"
	iconInbox  = "📥"
	iconShared = "👥"
)

func init() {
	// サブコマンドを追加
	projectCmd.AddCommand(projectListCmd)
	projectCmd.AddCommand(projectAddCmd)
	projectCmd.AddCommand(projectUpdateCmd)
	projectCmd.AddCommand(projectDeleteCmd)
	projectCmd.AddCommand(projectArchiveCmd)
	projectCmd.AddCommand(projectUnarchiveCmd)

	// プロジェクトコマンドをルートコマンドに追加
	rootCmd.AddCommand(projectCmd)

	// project list用のフラグ
	projectListCmd.Flags().BoolP("tree", "t", false, "show projects in tree structure")
	projectListCmd.Flags().BoolP("archived", "a", false, "show archived projects")
	projectListCmd.Flags().BoolP("favorites", "f", false, "show favorite projects only")

	// project add用のフラグ
	projectAddCmd.Flags().StringP("color", "c", "", "project color (e.g., red, blue, green)")
	projectAddCmd.Flags().StringP("parent", "p", "", "parent project ID or name")
	projectAddCmd.Flags().BoolP("favorite", "f", false, "mark as favorite project")

	// project update用のフラグ
	projectUpdateCmd.Flags().StringP("name", "n", "", "new project name")
	projectUpdateCmd.Flags().StringP("color", "c", "", "project color")
	projectUpdateCmd.Flags().BoolP("favorite", "f", false, "toggle favorite status")

	// project delete用のフラグ
	projectDeleteCmd.Flags().BoolP("force", "f", false, "skip confirmation prompt")
}

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
	RunE:  runProjectList,
}

// projectAddCmd はプロジェクト追加コマンド
var projectAddCmd = &cobra.Command{
	Use:   "add [project name]",
	Short: "Add a new project",
	Long:  `Add a new project to your Todoist.`,
	Args:  cobra.MinimumNArgs(1),
	RunE:  runProjectAdd,
}

// projectUpdateCmd はプロジェクト更新コマンド
var projectUpdateCmd = &cobra.Command{
	Use:   "update [project ID or name]",
	Short: "Update an existing project",
	Long:  `Update an existing project. Use --name, --color, or --favorite flags to specify what to update.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runProjectUpdate,
}

// projectDeleteCmd はプロジェクト削除コマンド
var projectDeleteCmd = &cobra.Command{
	Use:   "delete [project ID or name]",
	Short: "Delete a project",
	Long:  `Delete a project from your Todoist.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runProjectDelete,
}

// projectArchiveCmd はプロジェクトアーカイブコマンド
var projectArchiveCmd = &cobra.Command{
	Use:   "archive [project ID or name]",
	Short: "Archive a project",
	Long:  `Archive a project in your Todoist.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runProjectArchive,
}

// projectUnarchiveCmd はプロジェクトアーカイブ解除コマンド
var projectUnarchiveCmd = &cobra.Command{
	Use:   "unarchive [project ID or name]",
	Short: "Unarchive a project",
	Long:  `Unarchive a project in your Todoist.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runProjectUnarchive,
}

// projectListParams はプロジェクトリストのパラメータ
type projectListParams struct {
	showTree      bool
	showArchived  bool
	showFavorites bool
}

// projectListData はプロジェクトリスト実行で取得したデータ
type projectListData struct {
	projects []api.Project
}

// getProjectListParams はコマンドフラグからパラメータを取得する
func getProjectListParams(cmd *cobra.Command) *projectListParams {
	showTree, _ := cmd.Flags().GetBool("tree")
	showArchived, _ := cmd.Flags().GetBool("archived")
	showFavorites, _ := cmd.Flags().GetBool("favorites")

	return &projectListParams{
		showTree:      showTree,
		showArchived:  showArchived,
		showFavorites: showFavorites,
	}
}

// runProjectList はプロジェクト一覧表示の実際の処理
func runProjectList(cmd *cobra.Command, _ []string) error {
	ctx := createBaseContext()

	// セットアップ
	executor, err := setupProjectExecution(ctx)
	if err != nil {
		return err
	}
	defer executor.cleanup()

	// パラメータ取得と実行
	params := getProjectListParams(cmd)
	return executor.executeProjectList(ctx, params)
}

// executeProjectList はプロジェクト一覧表示の実行処理（テスト可能）
func (e *projectExecutor) executeProjectList(ctx context.Context, params *projectListParams) error {
	// 1. データ取得
	data, err := e.fetchProjectListData(ctx, params)
	if err != nil {
		return err
	}

	// 2. フィルタリング
	filteredProjects := applyProjectFilters(data.projects, params)

	// 3. 出力
	e.displayProjectResults(filteredProjects, params)

	return nil
}

// projectAddParams はプロジェクト追加のパラメータ
type projectAddParams struct {
	name       string
	color      string
	parentName string
	isFavorite bool
}

// getProjectAddParams はプロジェクト追加のパラメータを取得する
func getProjectAddParams(cmd *cobra.Command, args []string) *projectAddParams {
	color, _ := cmd.Flags().GetString("color")
	parentName, _ := cmd.Flags().GetString("parent")
	isFavorite, _ := cmd.Flags().GetBool("favorite")

	return &projectAddParams{
		name:       strings.Join(args, " "),
		color:      color,
		parentName: parentName,
		isFavorite: isFavorite,
	}
}

// runProjectAdd はプロジェクト追加の実際の処理
func runProjectAdd(cmd *cobra.Command, args []string) error {
	ctx := createBaseContext()

	// セットアップ
	executor, err := setupProjectExecution(ctx)
	if err != nil {
		return err
	}
	defer executor.cleanup()

	// パラメータ取得と実行
	params := getProjectAddParams(cmd, args)
	return executor.executeProjectAddWithOutput(ctx, params)
}

// executeProjectAddWithOutput はプロジェクト追加と結果表示を実行する（テスト可能）
func (e *projectExecutor) executeProjectAddWithOutput(ctx context.Context, params *projectAddParams) error {
	// 1. プロジェクト追加実行
	resp, err := e.executeProjectAdd(ctx, params)
	if err != nil {
		return fmt.Errorf("failed to create project: %w", err)
	}

	// 2. 結果表示
	e.displayProjectAddResult(params, resp)

	return nil
}

// projectUpdateParams はプロジェクト更新のパラメータ
type projectUpdateParams struct {
	projectIDOrName string
	newName         string
	color           string
	isFavorite      bool
	favoriteChanged bool
}

// getProjectUpdateParams はプロジェクト更新のパラメータを取得する
func getProjectUpdateParams(cmd *cobra.Command, args []string) *projectUpdateParams {
	newName, _ := cmd.Flags().GetString("name")
	color, _ := cmd.Flags().GetString("color")
	isFavorite, _ := cmd.Flags().GetBool("favorite")

	return &projectUpdateParams{
		projectIDOrName: args[0],
		newName:         newName,
		color:           color,
		isFavorite:      isFavorite,
		favoriteChanged: cmd.Flags().Changed("favorite"),
	}
}

// runProjectUpdate はプロジェクト更新の実際の処理
func runProjectUpdate(cmd *cobra.Command, args []string) error {
	ctx := createBaseContext()

	// セットアップ
	executor, err := setupProjectExecution(ctx)
	if err != nil {
		return err
	}
	defer executor.cleanup()

	// パラメータ取得と実行
	params := getProjectUpdateParams(cmd, args)
	return executor.executeProjectUpdateWithOutput(ctx, params)
}

// executeProjectUpdateWithOutput はプロジェクト更新と結果表示を実行する（テスト可能）
func (e *projectExecutor) executeProjectUpdateWithOutput(ctx context.Context, params *projectUpdateParams) error {
	// 1. 更新内容の確認
	if params.newName == "" && params.color == "" && !params.favoriteChanged {
		return fmt.Errorf("at least one update field must be specified (--name, --color, --favorite)")
	}

	// 2. プロジェクト更新実行
	resp, err := e.executeProjectUpdate(ctx, params)
	if err != nil {
		return fmt.Errorf("failed to update project: %w", err)
	}

	// 3. 結果表示
	e.displayProjectUpdateResult(params, resp)

	return nil
}

// projectDeleteParams はプロジェクト削除のパラメータ
type projectDeleteParams struct {
	projectIDOrName string
	force           bool
}

// getProjectDeleteParams はプロジェクト削除のパラメータを取得する
func getProjectDeleteParams(cmd *cobra.Command, args []string) *projectDeleteParams {
	force, _ := cmd.Flags().GetBool("force")
	return &projectDeleteParams{
		projectIDOrName: args[0],
		force:           force,
	}
}

// runProjectDelete はプロジェクト削除の実際の処理
func runProjectDelete(cmd *cobra.Command, args []string) error {
	ctx := createBaseContext()

	// セットアップ
	executor, err := setupProjectExecution(ctx)
	if err != nil {
		return err
	}
	defer executor.cleanup()

	// パラメータ取得と実行
	params := getProjectDeleteParams(cmd, args)
	return executor.executeProjectDeleteWithOutput(ctx, params)
}

// executeProjectDeleteWithOutput はプロジェクト削除と結果表示を実行する（テスト可能）
func (e *projectExecutor) executeProjectDeleteWithOutput(ctx context.Context, params *projectDeleteParams) error {
	// 1. 削除対象の確認
	project, shouldDelete, err := e.confirmProjectDeletion(ctx, params)
	if err != nil {
		return err
	}
	if !shouldDelete {
		return nil // ユーザーがキャンセル
	}

	// 2. プロジェクト削除実行
	resp, err := e.deleteProject(ctx, project.ID)
	if err != nil {
		return fmt.Errorf("failed to delete project: %w", err)
	}

	// 3. 結果表示
	e.displayProjectDeleteResult(project, resp)

	return nil
}

// projectArchiveParams はプロジェクトアーカイブのパラメータ
type projectArchiveParams struct {
	projectIDOrName string
}

// getProjectArchiveParams はプロジェクトアーカイブのパラメータを取得する
func getProjectArchiveParams(args []string) *projectArchiveParams {
	return &projectArchiveParams{
		projectIDOrName: args[0],
	}
}

// runProjectArchive はプロジェクトアーカイブの実際の処理
func runProjectArchive(_ *cobra.Command, args []string) error {
	ctx := createBaseContext()

	// セットアップ
	executor, err := setupProjectExecution(ctx)
	if err != nil {
		return err
	}
	defer executor.cleanup()

	// パラメータ取得と実行
	params := getProjectArchiveParams(args)
	return executor.executeProjectArchiveWithOutput(ctx, params)
}

// executeProjectArchiveWithOutput はプロジェクトアーカイブと結果表示を実行する（テスト可能）
func (e *projectExecutor) executeProjectArchiveWithOutput(ctx context.Context, params *projectArchiveParams) error {
	// 1. プロジェクトアーカイブ実行
	resp, err := e.executeProjectArchive(ctx, params)
	if err != nil {
		return fmt.Errorf("failed to archive project: %w", err)
	}

	// 2. 結果表示
	e.displaySuccessMessageForProject("📦 Project archived successfully!", resp.SyncToken)

	return nil
}

// runProjectUnarchive はプロジェクトアーカイブ解除の実際の処理
func runProjectUnarchive(_ *cobra.Command, args []string) error {
	ctx := createBaseContext()

	// セットアップ
	executor, err := setupProjectExecution(ctx)
	if err != nil {
		return err
	}
	defer executor.cleanup()

	// パラメータ取得と実行
	params := getProjectArchiveParams(args)
	return executor.executeProjectUnarchiveWithOutput(ctx, params)
}

// executeProjectUnarchiveWithOutput はプロジェクトアーカイブ解除と結果表示を実行する（テスト可能）
func (e *projectExecutor) executeProjectUnarchiveWithOutput(ctx context.Context, params *projectArchiveParams) error {
	// 1. プロジェクトアーカイブ解除実行
	resp, err := e.executeProjectUnarchive(ctx, params)
	if err != nil {
		return fmt.Errorf("failed to unarchive project: %w", err)
	}

	// 2. 結果表示
	e.displaySuccessMessageForProject("📁 Project unarchived successfully!", resp.SyncToken)

	return nil
}

// applyProjectFilters はプロジェクトにフィルタを適用する
func applyProjectFilters(projects []api.Project, params *projectListParams) []api.Project {
	// アーカイブ状態でフィルタリング
	if !params.showArchived {
		projects = filterActiveProjects(projects)
	} else {
		projects = filterArchivedProjects(projects)
	}

	return projects
}

// filterActiveProjects はアクティブなプロジェクトのみを返す
func filterActiveProjects(projects []api.Project) []api.Project {
	filtered := make([]api.Project, 0, len(projects))
	for _, project := range projects {
		if !project.IsArchived {
			filtered = append(filtered, project)
		}
	}
	return filtered
}

// filterArchivedProjects はアーカイブ済みプロジェクトのみを返す
func filterArchivedProjects(projects []api.Project) []api.Project {
	filtered := make([]api.Project, 0, len(projects))
	for _, project := range projects {
		if project.IsArchived {
			filtered = append(filtered, project)
		}
	}
	return filtered
}

// displayProjectResults はプロジェクト結果を表示する
func (e *projectExecutor) displayProjectResults(projects []api.Project, params *projectListParams) {
	// タイトルを取得
	title, emptyMessage := getProjectListTitle(params.showArchived, params.showFavorites)

	if len(projects) == 0 {
		e.output.Infof(emptyMessage)
		return
	}

	// プロジェクトを表示
	e.output.Projectf("%s (%d):", title, len(projects))
	e.output.Plainf("")

	if params.showTree {
		e.displayProjectsTree(projects)
	} else {
		e.displayProjectsList(projects)
	}
}

// getProjectListTitle はプロジェクトリストのタイトルを取得する
func getProjectListTitle(showArchived, showFavorites bool) (title, emptyMessage string) {
	switch {
	case showArchived:
		return "📦 Archived Projects", "📦 No archived projects found"
	case showFavorites:
		return "⭐ Favorite Projects", "⭐ No favorite projects found"
	default:
		return "📁 Projects", "📁 No projects found"
	}
}

// displayProjectsList はプロジェクトをリスト形式で表示する
func (e *projectExecutor) displayProjectsList(projects []api.Project) {
	for i, project := range projects {
		icon := iconFolder
		if project.InboxProject {
			icon = iconInbox
		} else if project.Shared {
			icon = iconShared
		}

		favoriteIcon := ""
		if project.IsFavorite {
			favoriteIcon = " ⭐"
		}
		archivedIcon := ""
		if project.IsArchived {
			archivedIcon = " 📦"
		}
		e.output.Plainf("%d. %s %s%s%s", i+1, icon, project.Name, favoriteIcon, archivedIcon)

		if IsVerbose() {
			e.output.Plainf("   ID: %s", project.ID)
			e.output.Plainf("   Color: %s", project.Color)
			if project.ParentID != "" {
				e.output.Plainf("   Parent ID: %s", project.ParentID)
			}
			if project.Shared {
				e.output.Plainf("   Shared: Yes")
			}
			e.output.Plainf("   Child Order: %d", project.ChildOrder)
		}

		e.output.Plainf("")
	}
}

// displayProjectsTree はプロジェクトをツリー形式で表示する
func (e *projectExecutor) displayProjectsTree(projects []api.Project) {
	// 親プロジェクトマップを作成
	parentMap := make(map[string][]api.Project)
	rootProjects := []api.Project{}

	for _, project := range projects {
		if project.ParentID == "" {
			rootProjects = append(rootProjects, project)
		} else {
			parentMap[project.ParentID] = append(parentMap[project.ParentID], project)
		}
	}

	// ルートプロジェクトから表示
	for i := range rootProjects {
		e.displayProjectTreeNode(&rootProjects[i], parentMap, 0)
	}
}

// displayProjectTreeNode は単一のプロジェクトノードをツリー形式で表示する
func (e *projectExecutor) displayProjectTreeNode(project *api.Project, parentMap map[string][]api.Project, depth int) {
	indent := strings.Repeat("  ", depth)
	icon := iconFolder
	if project.InboxProject {
		icon = iconInbox
	} else if project.Shared {
		icon = iconShared
	}

	favoriteIcon := ""
	if project.IsFavorite {
		favoriteIcon = " ⭐"
	}
	archivedIcon := ""
	if project.IsArchived {
		archivedIcon = " 📦"
	}
	e.output.Plainf("%s├─ %s %s%s%s", indent, icon, project.Name, favoriteIcon, archivedIcon)

	if IsVerbose() {
		e.output.Plainf("%s   ID: %s, Color: %s", indent, project.ID, project.Color)
	}

	// 子プロジェクトを表示
	if children, exists := parentMap[project.ID]; exists {
		for i := range children {
			e.displayProjectTreeNode(&children[i], parentMap, depth+1)
		}
	}
}

// displayProjectAddResult はプロジェクト追加結果を表示する
func (e *projectExecutor) displayProjectAddResult(params *projectAddParams, resp *api.SyncResponse) {
	e.output.Successf("📁 Project created successfully!")
	e.output.Plainf("   Name: %s", params.name)
	if params.color != "" {
		e.output.Plainf("   Color: %s", params.color)
	}
	if params.isFavorite {
		e.output.Plainf("   Favorite: Yes ⭐")
	}
	if IsVerbose() && resp.SyncToken != "" {
		e.output.Plainf("   Sync token: %s", resp.SyncToken)
	}
}

// displayProjectUpdateResult はプロジェクト更新結果を表示する
func (e *projectExecutor) displayProjectUpdateResult(params *projectUpdateParams, resp *api.SyncResponse) {
	e.output.Successf("✏️  Project updated successfully!")
	if params.newName != "" {
		e.output.Plainf("   New name: %s", params.newName)
	}
	if params.color != "" {
		e.output.Plainf("   Color: %s", params.color)
	}
	if params.favoriteChanged {
		if params.isFavorite {
			e.output.Plainf("   Favorite: Yes ⭐")
		} else {
			e.output.Plainf("   Favorite: No")
		}
	}
	if IsVerbose() && resp.SyncToken != "" {
		e.output.Plainf("   Sync token: %s", resp.SyncToken)
	}
}

// displayProjectDeleteResult はプロジェクト削除結果を表示する
func (e *projectExecutor) displayProjectDeleteResult(project *api.Project, resp *api.SyncResponse) {
	e.output.Successf("🗑️  Project deleted successfully!")
	e.output.Infof("    Deleted: %s", project.Name)
	if IsVerbose() && resp.SyncToken != "" {
		e.output.Plainf("    Sync token: %s", resp.SyncToken)
	}
}

// projectExecutor はプロジェクト実行に必要な情報をまとめた構造体
type projectExecutor struct {
	cfg        *config.Config
	repository *repository.Repository
	output     *cli.Output
}

// setupProjectExecution はプロジェクト実行環境をセットアップする
func setupProjectExecution(ctx context.Context) (*projectExecutor, error) {
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

	return &projectExecutor{
		cfg:        cfg,
		repository: repo,
		output:     output,
	}, nil
}

// cleanup はRepositoryのリソースクリーンアップを行う
func (e *projectExecutor) cleanup() {
	if err := e.repository.Close(); err != nil {
		e.output.Warningf("failed to close repository: %v", err)
	}
}

// fetchProjectListData はプロジェクトリストのデータを取得する
func (e *projectExecutor) fetchProjectListData(ctx context.Context, params *projectListParams) (*projectListData, error) {
	var projects []api.Project
	var err error

	if params.showFavorites {
		// お気に入りプロジェクトのみ取得
		allProjects, err := e.repository.GetAllProjects(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to get projects: %w", err)
		}
		// お気に入りのみフィルタリング
		for _, p := range allProjects {
			if p.IsFavorite {
				projects = append(projects, p)
			}
		}
	} else {
		// 全プロジェクトを取得
		projects, err = e.repository.GetAllProjects(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to get projects: %w", err)
		}
	}

	return &projectListData{
		projects: projects,
	}, nil
}

// findProjectIDByName はプロジェクト名からIDを検索する（Repository層に移植済み）
func (e *projectExecutor) findProjectIDByName(ctx context.Context, nameOrID string) (string, error) {
	return e.repository.FindProjectIDByName(ctx, nameOrID)
}

// findProjectByID は指定されたIDのプロジェクトを取得する
func (e *projectExecutor) findProjectByID(ctx context.Context, projectID string) (*api.Project, error) {
	projects, err := e.repository.GetAllProjects(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get projects: %w", err)
	}

	for i := range projects {
		if projects[i].ID == projectID {
			return &projects[i], nil
		}
	}

	return nil, fmt.Errorf("project not found")
}

// executeProjectAdd はプロジェクト追加を実行する
func (e *projectExecutor) executeProjectAdd(ctx context.Context, params *projectAddParams) (*api.SyncResponse, error) {
	// リクエストを構築
	req := &api.CreateProjectRequest{
		Name:       params.name,
		Color:      params.color,
		IsFavorite: params.isFavorite,
	}

	if params.parentName != "" {
		// 親プロジェクトIDを解決
		parentID, err := e.findProjectIDByName(ctx, params.parentName)
		if err != nil {
			return nil, fmt.Errorf("failed to find parent project: %w", err)
		}
		req.ParentID = parentID
	}

	// プロジェクトを作成
	return e.repository.CreateProject(ctx, req)
}

// executeProjectUpdate はプロジェクト更新を実行する
func (e *projectExecutor) executeProjectUpdate(ctx context.Context, params *projectUpdateParams) (*api.SyncResponse, error) {
	// プロジェクトIDを解決
	projectID, err := e.findProjectIDByName(ctx, params.projectIDOrName)
	if err != nil {
		return nil, fmt.Errorf("failed to find project: %w", err)
	}

	// リクエストを構築
	req := &api.UpdateProjectRequest{
		Name:       params.newName,
		Color:      params.color,
		IsFavorite: params.isFavorite,
	}

	// プロジェクトを更新
	return e.repository.UpdateProject(ctx, projectID, req)
}

// confirmProjectDeletion はプロジェクト削除の確認を行う
func (e *projectExecutor) confirmProjectDeletion(ctx context.Context, params *projectDeleteParams) (*api.Project, bool, error) {
	// プロジェクトIDを解決
	projectID, err := e.findProjectIDByName(ctx, params.projectIDOrName)
	if err != nil {
		return nil, false, fmt.Errorf("failed to find project: %w", err)
	}

	// プロジェクトの詳細を取得
	targetProject, err := e.findProjectByID(ctx, projectID)
	if err != nil {
		return nil, false, fmt.Errorf("project not found: %s - %w", params.projectIDOrName, err)
	}

	// 確認処理（forceフラグが無い場合）
	if !params.force {
		if !e.promptProjectDeletionConfirmation(targetProject) {
			return nil, false, nil // キャンセルされた
		}
	}

	return targetProject, true, nil
}

// deleteProject はプロジェクトを削除する
func (e *projectExecutor) deleteProject(ctx context.Context, projectID string) (*api.SyncResponse, error) {
	return e.repository.DeleteProject(ctx, projectID)
}

// executeProjectArchive はプロジェクトアーカイブを実行する
func (e *projectExecutor) executeProjectArchive(ctx context.Context, params *projectArchiveParams) (*api.SyncResponse, error) {
	// プロジェクトIDを解決
	projectID, err := e.findProjectIDByName(ctx, params.projectIDOrName)
	if err != nil {
		return nil, fmt.Errorf("failed to find project: %w", err)
	}

	// プロジェクトをアーカイブ
	return e.repository.ArchiveProject(ctx, projectID)
}

// executeProjectUnarchive はプロジェクトアーカイブ解除を実行する
func (e *projectExecutor) executeProjectUnarchive(ctx context.Context, params *projectArchiveParams) (*api.SyncResponse, error) {
	// プロジェクトIDを解決
	projectID, err := e.findProjectIDByName(ctx, params.projectIDOrName)
	if err != nil {
		return nil, fmt.Errorf("failed to find project: %w", err)
	}

	// プロジェクトのアーカイブを解除
	return e.repository.UnarchiveProject(ctx, projectID)
}

// promptProjectDeletionConfirmation はプロジェクト削除の確認プロンプトを表示する
func (e *projectExecutor) promptProjectDeletionConfirmation(project *api.Project) bool {
	e.output.Warningf("Are you sure you want to delete this project? (y/N)")
	e.output.Plainf("    ID: %s", project.ID)
	e.output.Plainf("    Name: %s", project.Name)
	e.output.Plainf("    Color: %s", project.Color)
	if project.IsFavorite {
		e.output.Plainf("    Favorite: Yes ⭐")
	}
	if project.Shared {
		e.output.Plainf("    Shared: Yes 👥")
	}
	e.output.PlainNoNewlinef("Enter your choice: ")

	var confirmation string
	_, err := fmt.Scanln(&confirmation)
	if err != nil {
		e.output.Errorf("Project deletion canceled")
		return false
	}

	if confirmation != "y" && confirmation != "Y" {
		e.output.Errorf("Project deletion canceled")
		return false
	}

	return true
}

// displaySuccessMessageForProject はプロジェクト用の成功メッセージを表示する
func (e *projectExecutor) displaySuccessMessageForProject(message string, syncToken string) {
	e.output.Successf("%s", message)
	if IsVerbose() && syncToken != "" {
		e.output.Plainf("Sync token: %s", syncToken)
	}
}
