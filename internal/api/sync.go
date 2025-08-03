package api

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

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
