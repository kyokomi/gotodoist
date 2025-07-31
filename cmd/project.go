package cmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/kyokomi/gotodoist/internal/api"
	"github.com/kyokomi/gotodoist/internal/config"
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
	RunE: func(cmd *cobra.Command, args []string) error {
		return runProjectList(cmd, args)
	},
}

// projectAddCmd はプロジェクト追加コマンド
var projectAddCmd = &cobra.Command{
	Use:   "add [project name]",
	Short: "Add a new project",
	Long:  `Add a new project to your Todoist.`,
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return runProjectAdd(cmd, args)
	},
}

// projectUpdateCmd はプロジェクト更新コマンド
var projectUpdateCmd = &cobra.Command{
	Use:   "update [project ID or name]",
	Short: "Update an existing project",
	Long:  `Update an existing project. Use --name, --color, or --favorite flags to specify what to update.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return runProjectUpdate(cmd, args)
	},
}

// projectDeleteCmd はプロジェクト削除コマンド
var projectDeleteCmd = &cobra.Command{
	Use:   "delete [project ID or name]",
	Short: "Delete a project",
	Long:  `Delete a project from your Todoist.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return runProjectDelete(cmd, args)
	},
}

// projectArchiveCmd はプロジェクトアーカイブコマンド
var projectArchiveCmd = &cobra.Command{
	Use:   "archive [project ID or name]",
	Short: "Archive a project",
	Long:  `Archive a project in your Todoist.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return runProjectArchive(cmd, args)
	},
}

// projectUnarchiveCmd はプロジェクトアーカイブ解除コマンド
var projectUnarchiveCmd = &cobra.Command{
	Use:   "unarchive [project ID or name]",
	Short: "Unarchive a project",
	Long:  `Unarchive a project in your Todoist.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return runProjectUnarchive(cmd, args)
	},
}

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

// runProjectList はプロジェクト一覧表示の実際の処理
func runProjectList(cmd *cobra.Command, args []string) error {
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
	showArchived, _ := cmd.Flags().GetBool("archived")
	showFavorites, _ := cmd.Flags().GetBool("favorites")
	showTree, _ := cmd.Flags().GetBool("tree")

	var projects []api.Project
	if showFavorites {
		projects, err = client.GetFavoriteProjects(ctx)
		if err != nil {
			return fmt.Errorf("failed to get favorite projects: %w", err)
		}
	} else {
		projects, err = client.GetAllProjects(ctx)
		if err != nil {
			return fmt.Errorf("failed to get projects: %w", err)
		}
	}

	// アーカイブフィルタリング
	if showArchived {
		// アーカイブ済みプロジェクトのみ表示
		var archivedProjects []api.Project
		for _, project := range projects {
			if project.IsArchived {
				archivedProjects = append(archivedProjects, project)
			}
		}
		projects = archivedProjects
	} else {
		// アクティブなプロジェクトのみ表示（デフォルト）
		var activeProjects []api.Project
		for _, project := range projects {
			if !project.IsArchived {
				activeProjects = append(activeProjects, project)
			}
		}
		projects = activeProjects
	}

	if len(projects) == 0 {
		if showArchived {
			fmt.Println("📦 No archived projects found")
		} else if showFavorites {
			fmt.Println("⭐ No favorite projects found")
		} else {
			fmt.Println("📁 No projects found")
		}
		return nil
	}

	// プロジェクトを表示
	title := "📁 Projects"
	if showArchived {
		title = "📦 Archived Projects"
	} else if showFavorites {
		title = "⭐ Favorite Projects"
	}
	fmt.Printf("%s (%d):\n\n", title, len(projects))

	if showTree {
		displayProjectsTree(projects)
	} else {
		displayProjectsList(projects)
	}

	return nil
}

// displayProjectsList はプロジェクトをリスト形式で表示する
func displayProjectsList(projects []api.Project) {
	for i, project := range projects {
		icon := "📁"
		if project.InboxProject {
			icon = "📥"
		} else if project.Shared {
			icon = "👥"
		}

		fmt.Printf("%d. %s %s", i+1, icon, project.Name)

		if project.IsFavorite {
			fmt.Print(" ⭐")
		}
		if project.IsArchived {
			fmt.Print(" 📦")
		}

		fmt.Println()

		if verbose {
			fmt.Printf("   ID: %s\n", project.ID)
			fmt.Printf("   Color: %s\n", project.Color)
			if project.ParentID != "" {
				fmt.Printf("   Parent ID: %s\n", project.ParentID)
			}
			if project.Shared {
				fmt.Printf("   Shared: Yes\n")
			}
			fmt.Printf("   Child Order: %d\n", project.ChildOrder)
		}

		fmt.Println()
	}
}

// displayProjectsTree はプロジェクトをツリー形式で表示する（簡易実装）
func displayProjectsTree(projects []api.Project) {
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
	for _, project := range rootProjects {
		displayProjectTreeNode(project, parentMap, 0)
	}
}

// displayProjectTreeNode は単一のプロジェクトノードをツリー形式で表示する
func displayProjectTreeNode(project api.Project, parentMap map[string][]api.Project, depth int) {
	indent := strings.Repeat("  ", depth)
	icon := "📁"
	if project.InboxProject {
		icon = "📥"
	} else if project.Shared {
		icon = "👥"
	}

	fmt.Printf("%s├─ %s %s", indent, icon, project.Name)

	if project.IsFavorite {
		fmt.Print(" ⭐")
	}
	if project.IsArchived {
		fmt.Print(" 📦")
	}

	fmt.Println()

	if verbose {
		fmt.Printf("%s   ID: %s, Color: %s\n", indent, project.ID, project.Color)
	}

	// 子プロジェクトを表示
	if children, exists := parentMap[project.ID]; exists {
		for _, child := range children {
			displayProjectTreeNode(child, parentMap, depth+1)
		}
	}
}

// findProjectIDByNameInProject はプロジェクト名からIDを検索する（プロジェクト専用）
func findProjectIDByNameInProject(ctx context.Context, client *api.Client, nameOrID string) (string, error) {
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

	// IDとして直接指定されている可能性をチェック
	for _, project := range projects {
		if project.ID == nameOrID {
			return project.ID, nil
		}
	}

	// 部分一致で検索
	for _, project := range projects {
		if strings.Contains(strings.ToLower(project.Name), nameOrID) {
			return project.ID, nil
		}
	}

	return "", fmt.Errorf("project not found: %s", nameOrID)
}

// runProjectAdd はプロジェクト追加の実際の処理
func runProjectAdd(cmd *cobra.Command, args []string) error {
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
	color, _ := cmd.Flags().GetString("color")
	parentName, _ := cmd.Flags().GetString("parent")
	isFavorite, _ := cmd.Flags().GetBool("favorite")

	// プロジェクト名を結合
	projectName := strings.Join(args, " ")

	// リクエストを構築
	req := &api.CreateProjectRequest{
		Name:       projectName,
		Color:      color,
		IsFavorite: isFavorite,
	}

	if parentName != "" {
		// 親プロジェクトIDを解決
		parentID, err := findProjectIDByNameInProject(ctx, client, parentName)
		if err != nil {
			return fmt.Errorf("failed to find parent project: %w", err)
		}
		req.ParentID = parentID
	}

	// プロジェクトを作成
	resp, err := client.CreateProject(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to create project: %w", err)
	}

	fmt.Printf("📁 Project created successfully!\n")
	fmt.Printf("   Name: %s\n", projectName)
	if color != "" {
		fmt.Printf("   Color: %s\n", color)
	}
	if isFavorite {
		fmt.Printf("   Favorite: Yes ⭐\n")
	}
	if verbose {
		fmt.Printf("   Sync token: %s\n", resp.SyncToken)
	}

	return nil
}

// runProjectUpdate はプロジェクト更新の実際の処理
func runProjectUpdate(cmd *cobra.Command, args []string) error {
	cfg, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	client, err := cfg.NewAPIClient()
	if err != nil {
		return fmt.Errorf("failed to create API client: %w", err)
	}

	ctx := context.Background()
	projectIDOrName := args[0]

	// フラグから設定を取得
	newName, _ := cmd.Flags().GetString("name")
	color, _ := cmd.Flags().GetString("color")
	favorite, _ := cmd.Flags().GetBool("favorite")

	// 何も更新内容がない場合はエラー
	if newName == "" && color == "" && !cmd.Flags().Changed("favorite") {
		return fmt.Errorf("at least one update field must be specified (--name, --color, --favorite)")
	}

	// プロジェクトIDを解決
	projectID, err := findProjectIDByNameInProject(ctx, client, projectIDOrName)
	if err != nil {
		return fmt.Errorf("failed to find project: %w", err)
	}

	// リクエストを構築
	req := &api.UpdateProjectRequest{
		Name:       newName,
		Color:      color,
		IsFavorite: favorite,
	}

	// プロジェクトを更新
	resp, err := client.UpdateProject(ctx, projectID, req)
	if err != nil {
		return fmt.Errorf("failed to update project: %w", err)
	}

	fmt.Printf("✏️  Project updated successfully!\n")
	if newName != "" {
		fmt.Printf("   New name: %s\n", newName)
	}
	if color != "" {
		fmt.Printf("   Color: %s\n", color)
	}
	if cmd.Flags().Changed("favorite") {
		if favorite {
			fmt.Printf("   Favorite: Yes ⭐\n")
		} else {
			fmt.Printf("   Favorite: No\n")
		}
	}
	if verbose {
		fmt.Printf("   Sync token: %s\n", resp.SyncToken)
	}

	return nil
}

// runProjectDelete はプロジェクト削除の実際の処理
func runProjectDelete(cmd *cobra.Command, args []string) error {
	cfg, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	client, err := cfg.NewAPIClient()
	if err != nil {
		return fmt.Errorf("failed to create API client: %w", err)
	}

	ctx := context.Background()
	projectIDOrName := args[0]

	// プロジェクトIDを解決
	projectID, err := findProjectIDByNameInProject(ctx, client, projectIDOrName)
	if err != nil {
		return fmt.Errorf("failed to find project: %w", err)
	}

	// プロジェクトの詳細を取得（確認用）
	projects, err := client.GetAllProjects(ctx)
	if err != nil {
		return fmt.Errorf("failed to get projects: %w", err)
	}

	var targetProject *api.Project
	for i := range projects {
		if projects[i].ID == projectID {
			targetProject = &projects[i]
			break
		}
	}

	if targetProject == nil {
		return fmt.Errorf("project not found: %s", projectIDOrName)
	}

	// 確認フラグをチェック
	force, _ := cmd.Flags().GetBool("force")
	if !force {
		fmt.Printf("⚠️  Are you sure you want to delete this project? (y/N)\n")
		fmt.Printf("    ID: %s\n", targetProject.ID)
		fmt.Printf("    Name: %s\n", targetProject.Name)
		fmt.Printf("    Color: %s\n", targetProject.Color)
		if targetProject.IsFavorite {
			fmt.Printf("    Favorite: Yes ⭐\n")
		}
		if targetProject.Shared {
			fmt.Printf("    Shared: Yes 👥\n")
		}
		fmt.Printf("Enter your choice: ")

		var confirmation string
		_, err := fmt.Scanln(&confirmation)
		if err != nil {
			fmt.Println("❌ Project deletion canceled")
			return nil
		}
		if confirmation != "y" && confirmation != "Y" {
			fmt.Println("❌ Project deletion canceled")
			return nil
		}
	}

	// プロジェクトを削除する
	resp, err := client.DeleteProject(ctx, projectID)
	if err != nil {
		return fmt.Errorf("failed to delete project: %w", err)
	}

	fmt.Printf("🗑️  Project deleted successfully!\n")
	fmt.Printf("    Deleted: %s\n", targetProject.Name)
	if verbose {
		fmt.Printf("    Sync token: %s\n", resp.SyncToken)
	}

	return nil
}

// runProjectArchive はプロジェクトアーカイブの実際の処理
func runProjectArchive(cmd *cobra.Command, args []string) error {
	cfg, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	client, err := cfg.NewAPIClient()
	if err != nil {
		return fmt.Errorf("failed to create API client: %w", err)
	}

	ctx := context.Background()
	projectIDOrName := args[0]

	// プロジェクトIDを解決
	projectID, err := findProjectIDByNameInProject(ctx, client, projectIDOrName)
	if err != nil {
		return fmt.Errorf("failed to find project: %w", err)
	}

	// プロジェクトをアーカイブする
	resp, err := client.ArchiveProject(ctx, projectID)
	if err != nil {
		return fmt.Errorf("failed to archive project: %w", err)
	}

	fmt.Printf("📦 Project archived successfully!\n")
	if verbose {
		fmt.Printf("   Sync token: %s\n", resp.SyncToken)
	}

	return nil
}

// runProjectUnarchive はプロジェクトアーカイブ解除の実際の処理
func runProjectUnarchive(cmd *cobra.Command, args []string) error {
	cfg, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	client, err := cfg.NewAPIClient()
	if err != nil {
		return fmt.Errorf("failed to create API client: %w", err)
	}

	ctx := context.Background()
	projectIDOrName := args[0]

	// プロジェクトIDを解決
	projectID, err := findProjectIDByNameInProject(ctx, client, projectIDOrName)
	if err != nil {
		return fmt.Errorf("failed to find project: %w", err)
	}

	// プロジェクトのアーカイブを解除する
	resp, err := client.UnarchiveProject(ctx, projectID)
	if err != nil {
		return fmt.Errorf("failed to unarchive project: %w", err)
	}

	fmt.Printf("📁 Project unarchived successfully!\n")
	if verbose {
		fmt.Printf("   Sync token: %s\n", resp.SyncToken)
	}

	return nil
}
