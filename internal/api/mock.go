//nolint:revive // MockClientのメソッドコメントは省略
package api

import (
	"context"
	"time"
)

// MockClient は Interface のテスト用モック実装
type MockClient struct {
	// 各メソッドに対応するカスタム関数
	SyncFunc                func(ctx context.Context, req *SyncRequest) (*SyncResponse, error)
	CreateProjectFunc       func(ctx context.Context, req *CreateProjectRequest) (*SyncResponse, error)
	UpdateProjectFunc       func(ctx context.Context, projectID string, req *UpdateProjectRequest) (*SyncResponse, error)
	DeleteProjectFunc       func(ctx context.Context, projectID string) (*SyncResponse, error)
	ArchiveProjectFunc      func(ctx context.Context, projectID string) (*SyncResponse, error)
	UnarchiveProjectFunc    func(ctx context.Context, projectID string) (*SyncResponse, error)
	GetAllProjectsFunc      func(ctx context.Context) ([]Project, error)
	GetProjectsFunc         func(ctx context.Context, syncToken string) (*SyncResponse, error)
	GetFavoriteProjectsFunc func(ctx context.Context) ([]Project, error)
	GetSharedProjectsFunc   func(ctx context.Context) ([]Project, error)

	CreateTaskFunc         func(ctx context.Context, req *CreateTaskRequest) (*SyncResponse, error)
	UpdateTaskFunc         func(ctx context.Context, taskID string, req *UpdateTaskRequest) (*SyncResponse, error)
	DeleteTaskFunc         func(ctx context.Context, taskID string) (*SyncResponse, error)
	CloseTaskFunc          func(ctx context.Context, taskID string) (*SyncResponse, error)
	ReopenTaskFunc         func(ctx context.Context, taskID string) (*SyncResponse, error)
	GetTasksFunc           func(ctx context.Context) ([]Item, error)
	GetTasksByProjectFunc  func(ctx context.Context, projectID string) ([]Item, error)
	GetTasksByPriorityFunc func(ctx context.Context, priority Priority) ([]Item, error)
	GetItemsFunc           func(ctx context.Context, syncToken string) (*SyncResponse, error)
	CompleteItemFunc       func(ctx context.Context, itemID string) (*SyncResponse, error)
	DeleteItemFunc         func(ctx context.Context, itemID string) (*SyncResponse, error)

	GetSectionsFunc    func(ctx context.Context, syncToken string) (*SyncResponse, error)
	GetAllSectionsFunc func(ctx context.Context) ([]Section, error)

	SetBaseURLFunc func(baseURL string) error
	SetTimeoutFunc func(timeout time.Duration)

	// デフォルトレスポンス (Funcが未設定の場合に使用)
	DefaultSyncResponse *SyncResponse
	DefaultProjects     []Project
	DefaultSections     []Section
	DefaultItems        []Item
}

// NewMockClient は新しいMockClientを作成する
func NewMockClient() *MockClient {
	return &MockClient{
		DefaultSyncResponse: &SyncResponse{
			SyncToken: "mock-sync-token",
		},
		DefaultProjects: []Project{},
		DefaultSections: []Section{},
		DefaultItems:    []Item{},
	}
}

// Interfaceインターフェースの実装

// Sync は同期APIを実行する
func (m *MockClient) Sync(ctx context.Context, req *SyncRequest) (*SyncResponse, error) {
	if m.SyncFunc != nil {
		return m.SyncFunc(ctx, req)
	}
	return m.DefaultSyncResponse, nil
}

// CreateProject は新しいプロジェクトを作成する
func (m *MockClient) CreateProject(ctx context.Context, req *CreateProjectRequest) (*SyncResponse, error) {
	if m.CreateProjectFunc != nil {
		return m.CreateProjectFunc(ctx, req)
	}
	return m.DefaultSyncResponse, nil
}

// UpdateProject は既存のプロジェクトを更新する
func (m *MockClient) UpdateProject(ctx context.Context, projectID string, req *UpdateProjectRequest) (*SyncResponse, error) {
	if m.UpdateProjectFunc != nil {
		return m.UpdateProjectFunc(ctx, projectID, req)
	}
	return m.DefaultSyncResponse, nil
}

func (m *MockClient) DeleteProject(ctx context.Context, projectID string) (*SyncResponse, error) {
	if m.DeleteProjectFunc != nil {
		return m.DeleteProjectFunc(ctx, projectID)
	}
	return m.DefaultSyncResponse, nil
}

func (m *MockClient) ArchiveProject(ctx context.Context, projectID string) (*SyncResponse, error) {
	if m.ArchiveProjectFunc != nil {
		return m.ArchiveProjectFunc(ctx, projectID)
	}
	return m.DefaultSyncResponse, nil
}

func (m *MockClient) UnarchiveProject(ctx context.Context, projectID string) (*SyncResponse, error) {
	if m.UnarchiveProjectFunc != nil {
		return m.UnarchiveProjectFunc(ctx, projectID)
	}
	return m.DefaultSyncResponse, nil
}

func (m *MockClient) GetAllProjects(ctx context.Context) ([]Project, error) {
	if m.GetAllProjectsFunc != nil {
		return m.GetAllProjectsFunc(ctx)
	}
	return m.DefaultProjects, nil
}

func (m *MockClient) GetProjects(ctx context.Context, syncToken string) (*SyncResponse, error) {
	if m.GetProjectsFunc != nil {
		return m.GetProjectsFunc(ctx, syncToken)
	}
	resp := *m.DefaultSyncResponse
	resp.Projects = m.DefaultProjects
	return &resp, nil
}

func (m *MockClient) GetFavoriteProjects(ctx context.Context) ([]Project, error) {
	if m.GetFavoriteProjectsFunc != nil {
		return m.GetFavoriteProjectsFunc(ctx)
	}

	// デフォルトはお気に入りプロジェクトをフィルタリング
	var favorites []Project
	for _, project := range m.DefaultProjects {
		if project.IsFavorite {
			favorites = append(favorites, project)
		}
	}
	return favorites, nil
}

func (m *MockClient) GetSharedProjects(ctx context.Context) ([]Project, error) {
	if m.GetSharedProjectsFunc != nil {
		return m.GetSharedProjectsFunc(ctx)
	}

	// デフォルトは共有プロジェクトをフィルタリング
	var shared []Project
	for _, project := range m.DefaultProjects {
		if project.Shared {
			shared = append(shared, project)
		}
	}
	return shared, nil
}

// Task operations
func (m *MockClient) CreateTask(ctx context.Context, req *CreateTaskRequest) (*SyncResponse, error) {
	if m.CreateTaskFunc != nil {
		return m.CreateTaskFunc(ctx, req)
	}
	return m.DefaultSyncResponse, nil
}

func (m *MockClient) UpdateTask(ctx context.Context, taskID string, req *UpdateTaskRequest) (*SyncResponse, error) {
	if m.UpdateTaskFunc != nil {
		return m.UpdateTaskFunc(ctx, taskID, req)
	}
	return m.DefaultSyncResponse, nil
}

func (m *MockClient) DeleteTask(ctx context.Context, taskID string) (*SyncResponse, error) {
	if m.DeleteTaskFunc != nil {
		return m.DeleteTaskFunc(ctx, taskID)
	}
	return m.DefaultSyncResponse, nil
}

func (m *MockClient) CloseTask(ctx context.Context, taskID string) (*SyncResponse, error) {
	if m.CloseTaskFunc != nil {
		return m.CloseTaskFunc(ctx, taskID)
	}
	return m.DefaultSyncResponse, nil
}

func (m *MockClient) ReopenTask(ctx context.Context, taskID string) (*SyncResponse, error) {
	if m.ReopenTaskFunc != nil {
		return m.ReopenTaskFunc(ctx, taskID)
	}
	return m.DefaultSyncResponse, nil
}

func (m *MockClient) GetTasks(ctx context.Context) ([]Item, error) {
	if m.GetTasksFunc != nil {
		return m.GetTasksFunc(ctx)
	}
	return m.DefaultItems, nil
}

func (m *MockClient) GetTasksByProject(ctx context.Context, projectID string) ([]Item, error) {
	if m.GetTasksByProjectFunc != nil {
		return m.GetTasksByProjectFunc(ctx, projectID)
	}

	// デフォルトはプロジェクトIDでフィルタリング
	var tasks []Item
	for _, item := range m.DefaultItems {
		if item.ProjectID == projectID {
			tasks = append(tasks, item)
		}
	}
	return tasks, nil
}

func (m *MockClient) GetTasksByPriority(ctx context.Context, priority Priority) ([]Item, error) {
	if m.GetTasksByPriorityFunc != nil {
		return m.GetTasksByPriorityFunc(ctx, priority)
	}

	// デフォルトは優先度でフィルタリング
	var tasks []Item
	for _, item := range m.DefaultItems {
		if item.Priority == int(priority) {
			tasks = append(tasks, item)
		}
	}
	return tasks, nil
}

func (m *MockClient) GetItems(ctx context.Context, syncToken string) (*SyncResponse, error) {
	if m.GetItemsFunc != nil {
		return m.GetItemsFunc(ctx, syncToken)
	}
	resp := *m.DefaultSyncResponse
	resp.Items = m.DefaultItems
	return &resp, nil
}

func (m *MockClient) CompleteItem(ctx context.Context, itemID string) (*SyncResponse, error) {
	if m.CompleteItemFunc != nil {
		return m.CompleteItemFunc(ctx, itemID)
	}
	return m.DefaultSyncResponse, nil
}

func (m *MockClient) DeleteItem(ctx context.Context, itemID string) (*SyncResponse, error) {
	if m.DeleteItemFunc != nil {
		return m.DeleteItemFunc(ctx, itemID)
	}
	return m.DefaultSyncResponse, nil
}

// Section operations
func (m *MockClient) GetSections(ctx context.Context, syncToken string) (*SyncResponse, error) {
	if m.GetSectionsFunc != nil {
		return m.GetSectionsFunc(ctx, syncToken)
	}
	resp := *m.DefaultSyncResponse
	resp.Sections = m.DefaultSections
	return &resp, nil
}

func (m *MockClient) GetAllSections(ctx context.Context) ([]Section, error) {
	if m.GetAllSectionsFunc != nil {
		return m.GetAllSectionsFunc(ctx)
	}
	return m.DefaultSections, nil
}

// Utility methods
func (m *MockClient) SetBaseURL(baseURL string) error {
	if m.SetBaseURLFunc != nil {
		return m.SetBaseURLFunc(baseURL)
	}
	return nil // デフォルトでは何もしない
}

func (m *MockClient) SetTimeout(timeout time.Duration) {
	if m.SetTimeoutFunc != nil {
		m.SetTimeoutFunc(timeout)
	}
	// デフォルトでは何もしない
}

// MockClientがInterfaceインターフェースを満たすことを確認
var _ Interface = (*MockClient)(nil)
