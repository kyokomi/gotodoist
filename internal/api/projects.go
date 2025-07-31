package api

import (
	"context"
	"fmt"
	"net/http"
	"strings"
)

// GetProjects は全てのプロジェクトを取得する
func (c *Client) GetProjects(ctx context.Context, filters *ProjectFilters) ([]Project, error) {
	path := "/projects"

	// IDsフィルタがある場合はクエリパラメータを追加
	if filters != nil && len(filters.IDs) > 0 {
		path += "?ids=" + strings.Join(filters.IDs, ",")
	}

	req, err := c.newRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}

	var projects []Project
	if err := c.do(req, &projects); err != nil {
		return nil, fmt.Errorf("failed to get projects: %w", err)
	}

	return projects, nil
}

// GetProject は指定されたIDのプロジェクトを取得する
func (c *Client) GetProject(ctx context.Context, projectID string) (*Project, error) {
	if projectID == "" {
		return nil, fmt.Errorf("project ID is required")
	}

	path := "/projects/" + projectID

	req, err := c.newRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}

	var project Project
	if err := c.do(req, &project); err != nil {
		return nil, fmt.Errorf("failed to get project: %w", err)
	}

	return &project, nil
}

// CreateProject は新しいプロジェクトを作成する
func (c *Client) CreateProject(ctx context.Context, req *CreateProjectRequest) (*Project, error) {
	if req == nil {
		return nil, fmt.Errorf("create project request is required")
	}
	if req.Name == "" {
		return nil, fmt.Errorf("project name is required")
	}

	httpReq, err := c.newRequest(ctx, http.MethodPost, "/projects", req)
	if err != nil {
		return nil, err
	}

	var project Project
	if err := c.do(httpReq, &project); err != nil {
		return nil, fmt.Errorf("failed to create project: %w", err)
	}

	return &project, nil
}

// UpdateProject は既存のプロジェクトを更新する
func (c *Client) UpdateProject(ctx context.Context, projectID string, req *UpdateProjectRequest) (*Project, error) {
	if projectID == "" {
		return nil, fmt.Errorf("project ID is required")
	}
	if req == nil {
		return nil, fmt.Errorf("update project request is required")
	}

	path := "/projects/" + projectID

	httpReq, err := c.newRequest(ctx, http.MethodPost, path, req)
	if err != nil {
		return nil, err
	}

	var project Project
	if err := c.do(httpReq, &project); err != nil {
		return nil, fmt.Errorf("failed to update project: %w", err)
	}

	return &project, nil
}

// DeleteProject はプロジェクトを削除する
func (c *Client) DeleteProject(ctx context.Context, projectID string) error {
	if projectID == "" {
		return fmt.Errorf("project ID is required")
	}

	path := "/projects/" + projectID

	req, err := c.newRequest(ctx, http.MethodDelete, path, nil)
	if err != nil {
		return err
	}

	if err := c.do(req, nil); err != nil {
		return fmt.Errorf("failed to delete project: %w", err)
	}

	return nil
}

// GetProjectTasks は指定されたプロジェクトのタスクを取得する
func (c *Client) GetProjectTasks(ctx context.Context, projectID string) ([]Task, error) {
	if projectID == "" {
		return nil, fmt.Errorf("project ID is required")
	}

	filters := &TaskFilters{
		ProjectID: projectID,
	}

	return c.GetTasks(ctx, filters)
}

// GetProjectComments はプロジェクトのコメントを取得する
func (c *Client) GetProjectComments(ctx context.Context, projectID string) ([]Comment, error) {
	if projectID == "" {
		return nil, fmt.Errorf("project ID is required")
	}

	path := "/comments?project_id=" + projectID

	req, err := c.newRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}

	var comments []Comment
	if err := c.do(req, &comments); err != nil {
		return nil, fmt.Errorf("failed to get project comments: %w", err)
	}

	return comments, nil
}

// GetProjectSections はプロジェクトのセクションを取得する
func (c *Client) GetProjectSections(ctx context.Context, projectID string) ([]Section, error) {
	if projectID == "" {
		return nil, fmt.Errorf("project ID is required")
	}

	path := "/sections?project_id=" + projectID

	req, err := c.newRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}

	var sections []Section
	if err := c.do(req, &sections); err != nil {
		return nil, fmt.Errorf("failed to get project sections: %w", err)
	}

	return sections, nil
}

// ArchiveProject はプロジェクトをアーカイブする（削除の代替手段）
// 注意: Todoist APIにはarchive機能がないため、これは削除を行う
func (c *Client) ArchiveProject(ctx context.Context, projectID string) error {
	return c.DeleteProject(ctx, projectID)
}

// Project colors (Todoist API で利用可能な色)
const (
	ColorBerryRed   = "berry_red"
	ColorRed        = "red"
	ColorOrange     = "orange"
	ColorYellow     = "yellow"
	ColorOliveGreen = "olive_green"
	ColorLimeGreen  = "lime_green"
	ColorGreen      = "green"
	ColorMintGreen  = "mint_green"
	ColorTeal       = "teal"
	ColorSkyBlue    = "sky_blue"
	ColorLightBlue  = "light_blue"
	ColorBlue       = "blue"
	ColorGrape      = "grape"
	ColorViolet     = "violet"
	ColorLavender   = "lavender"
	ColorMagenta    = "magenta"
	ColorSalmon     = "salmon"
	ColorCharcoal   = "charcoal"
	ColorGrey       = "grey"
	ColorTaupe      = "taupe"
)

// View styles
const (
	ViewStyleList  = "list"
	ViewStyleBoard = "board"
)

// GetFavoriteProjects はお気に入りプロジェクトを取得する
func (c *Client) GetFavoriteProjects(ctx context.Context) ([]Project, error) {
	projects, err := c.GetProjects(ctx, nil)
	if err != nil {
		return nil, err
	}

	var favorites []Project
	for _, project := range projects {
		if project.IsFavorite {
			favorites = append(favorites, project)
		}
	}

	return favorites, nil
}

// GetSharedProjects は共有プロジェクトを取得する
func (c *Client) GetSharedProjects(ctx context.Context) ([]Project, error) {
	projects, err := c.GetProjects(ctx, nil)
	if err != nil {
		return nil, err
	}

	var shared []Project
	for _, project := range projects {
		if project.IsShared {
			shared = append(shared, project)
		}
	}

	return shared, nil
}
