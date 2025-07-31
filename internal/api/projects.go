package api

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

// CreateProjectRequest はプロジェクト作成用のリクエスト構造体
type CreateProjectRequest struct {
	Name       string `json:"name"`
	ParentID   string `json:"parent_id,omitempty"`
	Color      string `json:"color,omitempty"`
	IsFavorite bool   `json:"is_favorite,omitempty"`
}

// UpdateProjectRequest はプロジェクト更新用のリクエスト構造体
type UpdateProjectRequest struct {
	Name       string `json:"name,omitempty"`
	Color      string `json:"color,omitempty"`
	IsFavorite bool   `json:"is_favorite,omitempty"`
}

// CreateProject は新しいプロジェクトを作成する
func (c *Client) CreateProject(ctx context.Context, req *CreateProjectRequest) (*SyncResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("create project request is required")
	}
	if req.Name == "" {
		return nil, fmt.Errorf("project name is required")
	}

	args := map[string]interface{}{
		"name": req.Name,
	}

	if req.ParentID != "" {
		args["parent_id"] = req.ParentID
	}
	if req.Color != "" {
		args["color"] = req.Color
	}
	if req.IsFavorite {
		args["is_favorite"] = req.IsFavorite
	}

	tempID := uuid.New().String()
	args["temp_id"] = tempID
	cmd := Command{
		Type:   CommandProjectAdd,
		UUID:   uuid.New().String(),
		TempID: tempID,
		Args:   args,
	}

	request := &SyncRequest{
		SyncToken: "*",
		Commands:  []Command{cmd},
	}

	return c.Sync(ctx, request)
}

// UpdateProject は既存のプロジェクトを更新する
func (c *Client) UpdateProject(ctx context.Context, projectID string, req *UpdateProjectRequest) (*SyncResponse, error) {
	if projectID == "" {
		return nil, fmt.Errorf("project ID is required")
	}
	if req == nil {
		return nil, fmt.Errorf("update project request is required")
	}

	args := map[string]interface{}{
		"id": projectID,
	}

	if req.Name != "" {
		args["name"] = req.Name
	}
	if req.Color != "" {
		args["color"] = req.Color
	}
	args["is_favorite"] = req.IsFavorite

	cmd := Command{
		Type: CommandProjectUpdate,
		UUID: uuid.New().String(),
		Args: args,
	}

	request := &SyncRequest{
		SyncToken: "*",
		Commands:  []Command{cmd},
	}

	return c.Sync(ctx, request)
}

// GetAllProjects は全てのプロジェクトを取得する
func (c *Client) GetAllProjects(ctx context.Context) ([]Project, error) {
	resp, err := c.GetProjects(ctx, "*")
	if err != nil {
		return nil, fmt.Errorf("failed to get projects: %w", err)
	}

	// アクティブなプロジェクトのみを返す
	var activeProjects []Project
	for _, project := range resp.Projects {
		if !project.IsDeleted {
			activeProjects = append(activeProjects, project)
		}
	}

	return activeProjects, nil
}

// DeleteProject はプロジェクトを削除する
func (c *Client) DeleteProject(ctx context.Context, projectID string) (*SyncResponse, error) {
	return c.DeleteProjectSync(ctx, projectID)
}

// ArchiveProject はプロジェクトをアーカイブする
func (c *Client) ArchiveProject(ctx context.Context, projectID string) (*SyncResponse, error) {
	if projectID == "" {
		return nil, fmt.Errorf("project ID is required")
	}

	cmd := Command{
		Type: CommandProjectArchive,
		UUID: uuid.New().String(),
		Args: map[string]interface{}{
			"id": projectID,
		},
	}

	request := &SyncRequest{
		SyncToken: "*",
		Commands:  []Command{cmd},
	}

	return c.Sync(ctx, request)
}

// UnarchiveProject はプロジェクトのアーカイブを解除する
func (c *Client) UnarchiveProject(ctx context.Context, projectID string) (*SyncResponse, error) {
	if projectID == "" {
		return nil, fmt.Errorf("project ID is required")
	}

	cmd := Command{
		Type: CommandProjectUnarchive,
		UUID: uuid.New().String(),
		Args: map[string]interface{}{
			"id": projectID,
		},
	}

	request := &SyncRequest{
		SyncToken: "*",
		Commands:  []Command{cmd},
	}

	return c.Sync(ctx, request)
}

// GetFavoriteProjects はお気に入りプロジェクトを取得する
func (c *Client) GetFavoriteProjects(ctx context.Context) ([]Project, error) {
	projects, err := c.GetAllProjects(ctx)
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
	projects, err := c.GetAllProjects(ctx)
	if err != nil {
		return nil, err
	}

	var shared []Project
	for _, project := range projects {
		if project.Shared {
			shared = append(shared, project)
		}
	}

	return shared, nil
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
