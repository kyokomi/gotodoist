package api

import (
	"time"
)

// Task はTodoistのタスクを表す
type Task struct {
	ID           string    `json:"id"`
	ProjectID    string    `json:"project_id"`
	SectionID    string    `json:"section_id,omitempty"`
	Content      string    `json:"content"`
	Description  string    `json:"description,omitempty"`
	IsCompleted  bool      `json:"is_completed"`
	Labels       []string  `json:"labels,omitempty"`
	Priority     int       `json:"priority"`
	CommentCount int       `json:"comment_count"`
	CreatedAt    time.Time `json:"created_at"`
	URL          string    `json:"url"`
	CreatorID    string    `json:"creator_id"`
	AssigneeID   string    `json:"assignee_id,omitempty"`
	AssignerID   string    `json:"assigner_id,omitempty"`
	Order        int       `json:"order"`
	Due          *Due      `json:"due,omitempty"`
	Duration     *Duration `json:"duration,omitempty"`
}

// Due はタスクの期限を表す
type Due struct {
	String      string `json:"string"`
	Date        string `json:"date"`
	IsRecurring bool   `json:"is_recurring"`
	Datetime    string `json:"datetime,omitempty"`
	Timezone    string `json:"timezone,omitempty"`
}

// Duration はタスクの実行時間を表す
type Duration struct {
	Amount int    `json:"amount"`
	Unit   string `json:"unit"`
}

// CreateTaskRequest はタスク作成リクエスト
type CreateTaskRequest struct {
	Content     string    `json:"content"`
	Description string    `json:"description,omitempty"`
	ProjectID   string    `json:"project_id,omitempty"`
	SectionID   string    `json:"section_id,omitempty"`
	ParentID    string    `json:"parent_id,omitempty"`
	Order       int       `json:"order,omitempty"`
	Labels      []string  `json:"labels,omitempty"`
	Priority    int       `json:"priority,omitempty"`
	DueString   string    `json:"due_string,omitempty"`
	DueDate     string    `json:"due_date,omitempty"`
	DueDatetime string    `json:"due_datetime,omitempty"`
	DueLang     string    `json:"due_lang,omitempty"`
	AssigneeID  string    `json:"assignee_id,omitempty"`
	Duration    *Duration `json:"duration,omitempty"`
}

// UpdateTaskRequest はタスク更新リクエスト
type UpdateTaskRequest struct {
	Content     string    `json:"content,omitempty"`
	Description string    `json:"description,omitempty"`
	Labels      []string  `json:"labels,omitempty"`
	Priority    int       `json:"priority,omitempty"`
	DueString   string    `json:"due_string,omitempty"`
	DueDate     string    `json:"due_date,omitempty"`
	DueDatetime string    `json:"due_datetime,omitempty"`
	DueLang     string    `json:"due_lang,omitempty"`
	AssigneeID  string    `json:"assignee_id,omitempty"`
	Duration    *Duration `json:"duration,omitempty"`
}

// Project はTodoistのプロジェクトを表す
type Project struct {
	ID             string `json:"id"`
	Name           string `json:"name"`
	CommentCount   int    `json:"comment_count"`
	Order          int    `json:"order"`
	Color          string `json:"color"`
	IsShared       bool   `json:"is_shared"`
	IsFavorite     bool   `json:"is_favorite"`
	IsInboxProject bool   `json:"is_inbox_project"`
	IsTeamInbox    bool   `json:"is_team_inbox"`
	ViewStyle      string `json:"view_style"`
	URL            string `json:"url"`
	ParentID       string `json:"parent_id,omitempty"`
}

// CreateProjectRequest はプロジェクト作成リクエスト
type CreateProjectRequest struct {
	Name       string `json:"name"`
	ParentID   string `json:"parent_id,omitempty"`
	Color      string `json:"color,omitempty"`
	IsFavorite bool   `json:"is_favorite,omitempty"`
	ViewStyle  string `json:"view_style,omitempty"`
}

// UpdateProjectRequest はプロジェクト更新リクエスト
type UpdateProjectRequest struct {
	Name       string `json:"name,omitempty"`
	Color      string `json:"color,omitempty"`
	IsFavorite bool   `json:"is_favorite,omitempty"`
	ViewStyle  string `json:"view_style,omitempty"`
}

// Section はTodoistのセクションを表す
type Section struct {
	ID        string `json:"id"`
	ProjectID string `json:"project_id"`
	Order     int    `json:"order"`
	Name      string `json:"name"`
}

// Label はTodoistのラベルを表す
type Label struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	Color      string `json:"color"`
	Order      int    `json:"order"`
	IsFavorite bool   `json:"is_favorite"`
}

// Comment はTodoistのコメントを表す
type Comment struct {
	ID         string                 `json:"id"`
	TaskID     string                 `json:"task_id,omitempty"`
	ProjectID  string                 `json:"project_id,omitempty"`
	PostedAt   time.Time              `json:"posted_at"`
	Content    string                 `json:"content"`
	Attachment map[string]interface{} `json:"attachment,omitempty"`
}

// TaskFilters はタスク取得時のフィルター
type TaskFilters struct {
	ProjectID string
	SectionID string
	Label     string
	Filter    string
	Lang      string
	IDs       []string
}

// ProjectFilters はプロジェクト取得時のフィルター
type ProjectFilters struct {
	IDs []string
}
