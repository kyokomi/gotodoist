package api

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

// GetTasks は全てのタスクを取得する
func (c *Client) GetTasks(ctx context.Context, filters *TaskFilters) ([]Task, error) {
	path := "/tasks"

	// クエリパラメータを構築
	if filters != nil {
		params := url.Values{}
		if filters.ProjectID != "" {
			params.Set("project_id", filters.ProjectID)
		}
		if filters.SectionID != "" {
			params.Set("section_id", filters.SectionID)
		}
		if filters.Label != "" {
			params.Set("label", filters.Label)
		}
		if filters.Filter != "" {
			params.Set("filter", filters.Filter)
		}
		if filters.Lang != "" {
			params.Set("lang", filters.Lang)
		}
		if len(filters.IDs) > 0 {
			params.Set("ids", strings.Join(filters.IDs, ","))
		}

		if len(params) > 0 {
			path += "?" + params.Encode()
		}
	}

	req, err := c.newRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}

	var tasks []Task
	if err := c.do(req, &tasks); err != nil {
		return nil, fmt.Errorf("failed to get tasks: %w", err)
	}

	return tasks, nil
}

// GetTask は指定されたIDのタスクを取得する
func (c *Client) GetTask(ctx context.Context, taskID string) (*Task, error) {
	if taskID == "" {
		return nil, fmt.Errorf("task ID is required")
	}

	path := "/tasks/" + taskID

	req, err := c.newRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}

	var task Task
	if err := c.do(req, &task); err != nil {
		return nil, fmt.Errorf("failed to get task: %w", err)
	}

	return &task, nil
}

// CreateTask は新しいタスクを作成する
func (c *Client) CreateTask(ctx context.Context, req *CreateTaskRequest) (*Task, error) {
	if req == nil {
		return nil, fmt.Errorf("create task request is required")
	}
	if req.Content == "" {
		return nil, fmt.Errorf("task content is required")
	}

	httpReq, err := c.newRequest(ctx, http.MethodPost, "/tasks", req)
	if err != nil {
		return nil, err
	}

	var task Task
	if err := c.do(httpReq, &task); err != nil {
		return nil, fmt.Errorf("failed to create task: %w", err)
	}

	return &task, nil
}

// UpdateTask は既存のタスクを更新する
func (c *Client) UpdateTask(ctx context.Context, taskID string, req *UpdateTaskRequest) (*Task, error) {
	if taskID == "" {
		return nil, fmt.Errorf("task ID is required")
	}
	if req == nil {
		return nil, fmt.Errorf("update task request is required")
	}

	path := "/tasks/" + taskID

	httpReq, err := c.newRequest(ctx, http.MethodPost, path, req)
	if err != nil {
		return nil, err
	}

	var task Task
	if err := c.do(httpReq, &task); err != nil {
		return nil, fmt.Errorf("failed to update task: %w", err)
	}

	return &task, nil
}

// CloseTask はタスクを完了状態にする
func (c *Client) CloseTask(ctx context.Context, taskID string) error {
	if taskID == "" {
		return fmt.Errorf("task ID is required")
	}

	path := "/tasks/" + taskID + "/close"

	req, err := c.newRequest(ctx, http.MethodPost, path, nil)
	if err != nil {
		return err
	}

	if err := c.do(req, nil); err != nil {
		return fmt.Errorf("failed to close task: %w", err)
	}

	return nil
}

// ReopenTask はタスクを未完了状態に戻す
func (c *Client) ReopenTask(ctx context.Context, taskID string) error {
	if taskID == "" {
		return fmt.Errorf("task ID is required")
	}

	path := "/tasks/" + taskID + "/reopen"

	req, err := c.newRequest(ctx, http.MethodPost, path, nil)
	if err != nil {
		return err
	}

	if err := c.do(req, nil); err != nil {
		return fmt.Errorf("failed to reopen task: %w", err)
	}

	return nil
}

// DeleteTask はタスクを削除する
func (c *Client) DeleteTask(ctx context.Context, taskID string) error {
	if taskID == "" {
		return fmt.Errorf("task ID is required")
	}

	path := "/tasks/" + taskID

	req, err := c.newRequest(ctx, http.MethodDelete, path, nil)
	if err != nil {
		return err
	}

	if err := c.do(req, nil); err != nil {
		return fmt.Errorf("failed to delete task: %w", err)
	}

	return nil
}

// MoveTask はタスクを他のプロジェクトに移動する
func (c *Client) MoveTask(ctx context.Context, taskID, projectID string) error {
	if taskID == "" {
		return fmt.Errorf("task ID is required")
	}
	if projectID == "" {
		return fmt.Errorf("project ID is required")
	}

	// UpdateTaskを使用してプロジェクトIDを変更
	updateReq := &UpdateTaskRequest{}

	path := "/tasks/" + taskID
	params := url.Values{}
	params.Set("project_id", projectID)

	httpReq, err := c.newRequest(ctx, http.MethodPost, path+"?"+params.Encode(), updateReq)
	if err != nil {
		return err
	}

	if err := c.do(httpReq, nil); err != nil {
		return fmt.Errorf("failed to move task: %w", err)
	}

	return nil
}

// GetTaskComments はタスクのコメントを取得する
func (c *Client) GetTaskComments(ctx context.Context, taskID string) ([]Comment, error) {
	if taskID == "" {
		return nil, fmt.Errorf("task ID is required")
	}

	path := "/comments?task_id=" + taskID

	req, err := c.newRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}

	var comments []Comment
	if err := c.do(req, &comments); err != nil {
		return nil, fmt.Errorf("failed to get task comments: %w", err)
	}

	return comments, nil
}

// Priority はタスクの優先度を表す型
type Priority int

// Priority constants for tasks
const (
	PriorityNormal   Priority = 1
	PriorityHigh     Priority = 2
	PriorityVeryHigh Priority = 3
	PriorityUrgent   Priority = 4
)

// String はPriorityの文字列表現を返す
func (p Priority) String() string {
	switch p {
	case PriorityNormal:
		return "Normal"
	case PriorityHigh:
		return "High"
	case PriorityVeryHigh:
		return "Very High"
	case PriorityUrgent:
		return "Urgent"
	default:
		return fmt.Sprintf("Priority(%d)", int(p))
	}
}

// IsValid はPriorityが有効な値かどうかを判定する
func (p Priority) IsValid() bool {
	return p >= PriorityNormal && p <= PriorityUrgent
}

// GetTasksByPriority は指定された優先度のタスクを取得する
func (c *Client) GetTasksByPriority(ctx context.Context, priority Priority) ([]Task, error) {
	if !priority.IsValid() {
		return nil, fmt.Errorf("invalid priority: %s", priority.String())
	}

	filter := fmt.Sprintf("p%d", int(priority))
	filters := &TaskFilters{
		Filter: filter,
	}

	return c.GetTasks(ctx, filters)
}
