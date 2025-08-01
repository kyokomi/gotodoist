package api

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

// GetAllData は全てのデータを同期で取得する
func (c *Client) GetAllData(ctx context.Context) (*SyncResponse, error) {
	req := &SyncRequest{
		SyncToken:     "*",
		ResourceTypes: []string{ResourceAll},
	}
	return c.Sync(ctx, req)
}

// GetItems はアイテム（タスク）のみを取得する
func (c *Client) GetItems(ctx context.Context, syncToken string) (*SyncResponse, error) {
	req := &SyncRequest{
		SyncToken:     syncToken,
		ResourceTypes: []string{ResourceItems},
	}
	return c.Sync(ctx, req)
}

// GetProjects はプロジェクトのみを取得する
func (c *Client) GetProjects(ctx context.Context, syncToken string) (*SyncResponse, error) {
	req := &SyncRequest{
		SyncToken:     syncToken,
		ResourceTypes: []string{ResourceProjects},
	}
	return c.Sync(ctx, req)
}

// GetSections はセクションのみを取得する
func (c *Client) GetSections(ctx context.Context, syncToken string) (*SyncResponse, error) {
	req := &SyncRequest{
		SyncToken:     syncToken,
		ResourceTypes: []string{ResourceSections},
	}
	return c.Sync(ctx, req)
}

// GetAllSections は全セクション情報を取得する
func (c *Client) GetAllSections(ctx context.Context) ([]Section, error) {
	resp, err := c.GetSections(ctx, "*")
	if err != nil {
		return nil, fmt.Errorf("failed to get sections: %w", err)
	}
	return resp.Sections, nil
}

// AddItem は新しいタスクを追加する
func (c *Client) AddItem(ctx context.Context, content, projectID string) (*SyncResponse, error) {
	cmd := Command{
		Type: CommandItemAdd,
		UUID: uuid.New().String(),
		Args: map[string]interface{}{
			"content": content,
		},
	}

	if projectID != "" {
		cmd.Args["project_id"] = projectID
	}

	req := &SyncRequest{
		SyncToken: "*",
		Commands:  []Command{cmd},
	}

	return c.Sync(ctx, req)
}

// UpdateItem はタスクを更新する
func (c *Client) UpdateItem(ctx context.Context, itemID string, updates map[string]interface{}) (*SyncResponse, error) {
	args := make(map[string]interface{})
	args["id"] = itemID
	for k, v := range updates {
		args[k] = v
	}

	cmd := Command{
		Type: CommandItemUpdate,
		UUID: uuid.New().String(),
		Args: args,
	}

	req := &SyncRequest{
		SyncToken: "*",
		Commands:  []Command{cmd},
	}

	return c.Sync(ctx, req)
}

// CompleteItem はタスクを完了にする
func (c *Client) CompleteItem(ctx context.Context, itemID string) (*SyncResponse, error) {
	cmd := Command{
		Type: CommandItemComplete,
		UUID: uuid.New().String(),
		Args: map[string]interface{}{
			"id": itemID,
		},
	}

	req := &SyncRequest{
		SyncToken: "*",
		Commands:  []Command{cmd},
	}

	return c.Sync(ctx, req)
}

// DeleteItem はタスクを削除する
func (c *Client) DeleteItem(ctx context.Context, itemID string) (*SyncResponse, error) {
	cmd := Command{
		Type: CommandItemDelete,
		UUID: uuid.New().String(),
		Args: map[string]interface{}{
			"id": itemID,
		},
	}

	req := &SyncRequest{
		SyncToken: "*",
		Commands:  []Command{cmd},
	}

	return c.Sync(ctx, req)
}

// AddProject は新しいプロジェクトを追加する
func (c *Client) AddProject(ctx context.Context, name string) (*SyncResponse, error) {
	cmd := Command{
		Type: CommandProjectAdd,
		UUID: uuid.New().String(),
		Args: map[string]interface{}{
			"name": name,
		},
	}

	req := &SyncRequest{
		SyncToken: "*",
		Commands:  []Command{cmd},
	}

	return c.Sync(ctx, req)
}

// UpdateProjectSync はプロジェクトを更新する（低レベルAPI）
func (c *Client) UpdateProjectSync(ctx context.Context, projectID string, updates map[string]interface{}) (*SyncResponse, error) {
	args := make(map[string]interface{})
	args["id"] = projectID
	for k, v := range updates {
		args[k] = v
	}

	cmd := Command{
		Type: CommandProjectUpdate,
		UUID: uuid.New().String(),
		Args: args,
	}

	req := &SyncRequest{
		SyncToken: "*",
		Commands:  []Command{cmd},
	}

	return c.Sync(ctx, req)
}

// DeleteProjectSync はプロジェクトを削除する（低レベルAPI）
func (c *Client) DeleteProjectSync(ctx context.Context, projectID string) (*SyncResponse, error) {
	cmd := Command{
		Type: CommandProjectDelete,
		UUID: uuid.New().String(),
		Args: map[string]interface{}{
			"id": projectID,
		},
	}

	req := &SyncRequest{
		SyncToken: "*",
		Commands:  []Command{cmd},
	}

	return c.Sync(ctx, req)
}

// ExecuteCommands は複数のコマンドを同時に実行する
func (c *Client) ExecuteCommands(ctx context.Context, commands []Command) (*SyncResponse, error) {
	if len(commands) == 0 {
		return nil, fmt.Errorf("no commands provided")
	}

	req := &SyncRequest{
		SyncToken: "*",
		Commands:  commands,
	}

	return c.Sync(ctx, req)
}
