package api

import (
	"context"
	"time"
)

// Interface は Todoist API クライアントのインターフェース
// テスト時にはモックとして実装可能
type Interface interface {
	// Sync API
	Sync(ctx context.Context, req *SyncRequest) (*SyncResponse, error)

	// Project operations
	CreateProject(ctx context.Context, req *CreateProjectRequest) (*SyncResponse, error)
	UpdateProject(ctx context.Context, projectID string, req *UpdateProjectRequest) (*SyncResponse, error)
	DeleteProject(ctx context.Context, projectID string) (*SyncResponse, error)
	ArchiveProject(ctx context.Context, projectID string) (*SyncResponse, error)
	UnarchiveProject(ctx context.Context, projectID string) (*SyncResponse, error)
	GetAllProjects(ctx context.Context) ([]Project, error)
	GetProjects(ctx context.Context, syncToken string) (*SyncResponse, error)
	GetFavoriteProjects(ctx context.Context) ([]Project, error)
	GetSharedProjects(ctx context.Context) ([]Project, error)

	// Task operations
	CreateTask(ctx context.Context, req *CreateTaskRequest) (*SyncResponse, error)
	UpdateTask(ctx context.Context, taskID string, req *UpdateTaskRequest) (*SyncResponse, error)
	DeleteTask(ctx context.Context, taskID string) (*SyncResponse, error)
	CloseTask(ctx context.Context, taskID string) (*SyncResponse, error)
	ReopenTask(ctx context.Context, taskID string) (*SyncResponse, error)
	GetTasks(ctx context.Context) ([]Item, error)
	GetTasksByProject(ctx context.Context, projectID string) ([]Item, error)
	GetTasksByPriority(ctx context.Context, priority Priority) ([]Item, error)
	GetItems(ctx context.Context, syncToken string) (*SyncResponse, error)
	CompleteItem(ctx context.Context, itemID string) (*SyncResponse, error)
	DeleteItem(ctx context.Context, itemID string) (*SyncResponse, error)

	// Section operations
	GetSections(ctx context.Context, syncToken string) (*SyncResponse, error)
	GetAllSections(ctx context.Context) ([]Section, error)

	// Utility methods
	SetBaseURL(baseURL string) error
	SetTimeout(timeout time.Duration)
}

// 既存のClientがInterfaceインターフェースを満たすことを確認
var _ Interface = (*Client)(nil)
