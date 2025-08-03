package api

import (
	"context"
	"fmt"
)

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
