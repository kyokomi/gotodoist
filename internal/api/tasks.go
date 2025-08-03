package api

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

// CreateTaskRequest はタスク作成用のリクエスト構造体
type CreateTaskRequest struct {
	Content     string   `json:"content"`
	Description string   `json:"description,omitempty"`
	ProjectID   string   `json:"project_id,omitempty"`
	SectionID   string   `json:"section_id,omitempty"`
	ParentID    string   `json:"parent_id,omitempty"`
	Order       int      `json:"order,omitempty"`
	Labels      []string `json:"labels,omitempty"`
	Priority    int      `json:"priority,omitempty"`
	DueString   string   `json:"due_string,omitempty"`
	DueDate     string   `json:"due_date,omitempty"`
	DueDatetime string   `json:"due_datetime,omitempty"`
	DueLang     string   `json:"due_lang,omitempty"`
	AssigneeID  string   `json:"assignee_id,omitempty"`
}

// UpdateTaskRequest はタスク更新用のリクエスト構造体
type UpdateTaskRequest struct {
	Content     string   `json:"content,omitempty"`
	Description string   `json:"description,omitempty"`
	Labels      []string `json:"labels,omitempty"`
	Priority    int      `json:"priority,omitempty"`
	DueString   string   `json:"due_string,omitempty"`
	DueDate     string   `json:"due_date,omitempty"`
	DueDatetime string   `json:"due_datetime,omitempty"`
	DueLang     string   `json:"due_lang,omitempty"`
	AssigneeID  string   `json:"assignee_id,omitempty"`
}

// CreateTask は新しいタスクを作成する
func (c *Client) CreateTask(ctx context.Context, req *CreateTaskRequest) (*SyncResponse, error) {
	if err := validateCreateTaskRequest(req); err != nil {
		return nil, err
	}

	args := map[string]interface{}{
		"content": req.Content,
	}

	if req.Description != "" {
		args["description"] = req.Description
	}
	if req.ProjectID != "" {
		args["project_id"] = req.ProjectID
	}
	if req.SectionID != "" {
		args["section_id"] = req.SectionID
	}
	if req.ParentID != "" {
		args["parent_id"] = req.ParentID
	}
	if req.Order > 0 {
		args["child_order"] = req.Order
	}
	if len(req.Labels) > 0 {
		args["labels"] = req.Labels
	}
	if req.Priority > 0 {
		args["priority"] = req.Priority
	}
	switch {
	case req.DueString != "":
		args["due"] = map[string]interface{}{
			"string": req.DueString,
		}
		if req.DueLang != "" {
			args["due"].(map[string]interface{})["lang"] = req.DueLang
		}
	case req.DueDate != "":
		args["due"] = map[string]interface{}{
			"date": req.DueDate,
		}
	case req.DueDatetime != "":
		args["due"] = map[string]interface{}{
			"datetime": req.DueDatetime,
		}
	}
	if req.AssigneeID != "" {
		args["responsible_uid"] = req.AssigneeID
	}

	tempID := uuid.New().String()
	args["temp_id"] = tempID

	cmd := Command{
		Type:   CommandItemAdd,
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

// UpdateTask は既存のタスクを更新する
func (c *Client) UpdateTask(ctx context.Context, taskID string, req *UpdateTaskRequest) (*SyncResponse, error) {
	if err := validateTaskID(taskID); err != nil {
		return nil, err
	}
	if err := validateUpdateTaskRequest(req); err != nil {
		return nil, err
	}

	args := map[string]interface{}{
		"id": taskID,
	}

	if req.Content != "" {
		args["content"] = req.Content
	}
	if req.Description != "" {
		args["description"] = req.Description
	}
	if len(req.Labels) > 0 {
		args["labels"] = req.Labels
	}
	if req.Priority > 0 {
		args["priority"] = req.Priority
	}
	switch {
	case req.DueString != "":
		args["due"] = map[string]interface{}{
			"string": req.DueString,
		}
		if req.DueLang != "" {
			args["due"].(map[string]interface{})["lang"] = req.DueLang
		}
	case req.DueDate != "":
		args["due"] = map[string]interface{}{
			"date": req.DueDate,
		}
	case req.DueDatetime != "":
		args["due"] = map[string]interface{}{
			"datetime": req.DueDatetime,
		}
	}
	if req.AssigneeID != "" {
		args["responsible_uid"] = req.AssigneeID
	}

	cmd := Command{
		Type: CommandItemUpdate,
		UUID: uuid.New().String(),
		Args: args,
	}

	request := &SyncRequest{
		SyncToken: "*",
		Commands:  []Command{cmd},
	}

	return c.Sync(ctx, request)
}

// GetTasks は全てのタスクを取得する
func (c *Client) GetTasks(ctx context.Context) ([]Item, error) {
	resp, err := c.GetItems(ctx, "*")
	if err != nil {
		return nil, fmt.Errorf("failed to get tasks: %w", err)
	}

	// アクティブなタスクのみを返す
	var activeTasks []Item
	for i := range resp.Items {
		if !resp.Items[i].IsDeleted {
			activeTasks = append(activeTasks, resp.Items[i])
		}
	}

	return activeTasks, nil
}

// GetTasksByProject は指定されたプロジェクトのタスクを取得する
func (c *Client) GetTasksByProject(ctx context.Context, projectID string) ([]Item, error) {
	if err := validateProjectID(projectID); err != nil {
		return nil, err
	}

	tasks, err := c.GetTasks(ctx)
	if err != nil {
		return nil, err
	}

	var projectTasks []Item
	for i := range tasks {
		if tasks[i].ProjectID == projectID {
			projectTasks = append(projectTasks, tasks[i])
		}
	}

	return projectTasks, nil
}

// CloseTask はタスクを完了にする
func (c *Client) CloseTask(ctx context.Context, taskID string) (*SyncResponse, error) {
	return c.CompleteItem(ctx, taskID)
}

// ReopenTask はタスクを未完了に戻す
func (c *Client) ReopenTask(ctx context.Context, taskID string) (*SyncResponse, error) {
	cmd := Command{
		Type: CommandItemUncomplete,
		UUID: uuid.New().String(),
		Args: map[string]interface{}{
			"id": taskID,
		},
	}

	req := &SyncRequest{
		SyncToken: "*",
		Commands:  []Command{cmd},
	}

	return c.Sync(ctx, req)
}

// DeleteTask はタスクを削除する
func (c *Client) DeleteTask(ctx context.Context, taskID string) (*SyncResponse, error) {
	return c.DeleteItem(ctx, taskID)
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
// NOTE: 現在のCLIでは未使用ですが、将来的に優先度による
// タスクフィルタリング機能を実装する際に使用するため残しています。
func (c *Client) GetTasksByPriority(ctx context.Context, priority Priority) ([]Item, error) {
	if !priority.IsValid() {
		return nil, fmt.Errorf("invalid priority: %s", priority.String())
	}

	tasks, err := c.GetTasks(ctx)
	if err != nil {
		return nil, err
	}

	var priorityTasks []Item
	for i := range tasks {
		if tasks[i].Priority == int(priority) {
			priorityTasks = append(priorityTasks, tasks[i])
		}
	}

	return priorityTasks, nil
}
