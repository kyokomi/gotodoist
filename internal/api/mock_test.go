package api

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestMockClient_BasicFunctionality はMockClientの基本機能テスト
func TestMockClient_BasicFunctionality(t *testing.T) {
	mock := NewMockClient()
	require.NotNil(t, mock)

	// インターフェースが正しく実装されていることを確認
	var _ Interface = mock
}

// TestMockClient_SyncFunc はSyncFunc機能のテスト
func TestMockClient_SyncFunc(t *testing.T) {
	mock := NewMockClient()

	// SyncFunc が設定されていない場合のデフォルト動作
	resp, err := mock.Sync(context.Background(), &SyncRequest{})
	assert.NoError(t, err)
	assert.NotNil(t, resp)

	// カスタム SyncFunc を設定
	mock.SyncFunc = func(_ context.Context, req *SyncRequest) (*SyncResponse, error) {
		assert.Equal(t, "test-token", req.SyncToken)
		return &SyncResponse{
			SyncToken: "mock-response-token",
		}, nil
	}

	resp, err = mock.Sync(context.Background(), &SyncRequest{SyncToken: "test-token"})
	assert.NoError(t, err)
	assert.Equal(t, "mock-response-token", resp.SyncToken)
}

// TestMockClient_ProjectOperations はプロジェクト操作のテスト
func TestMockClient_ProjectOperations(t *testing.T) {
	mock := NewMockClient()
	ctx := context.Background()

	// CreateProject のテスト
	mock.CreateProjectFunc = func(_ context.Context, req *CreateProjectRequest) (*SyncResponse, error) {
		assert.Equal(t, "Test Project", req.Name)
		return &SyncResponse{SyncToken: "create-token"}, nil
	}

	resp, err := mock.CreateProject(ctx, &CreateProjectRequest{Name: "Test Project"})
	assert.NoError(t, err)
	assert.Equal(t, "create-token", resp.SyncToken)

	// UpdateProject のテスト
	mock.UpdateProjectFunc = func(_ context.Context, projectID string, req *UpdateProjectRequest) (*SyncResponse, error) {
		assert.Equal(t, "project-123", projectID)
		assert.Equal(t, "Updated Project", req.Name)
		return &SyncResponse{SyncToken: "update-token"}, nil
	}

	resp, err = mock.UpdateProject(ctx, "project-123", &UpdateProjectRequest{Name: "Updated Project"})
	assert.NoError(t, err)
	assert.Equal(t, "update-token", resp.SyncToken)

	// DeleteProject のテスト
	mock.DeleteProjectFunc = func(_ context.Context, projectID string) (*SyncResponse, error) {
		assert.Equal(t, "project-456", projectID)
		return &SyncResponse{SyncToken: "delete-token"}, nil
	}

	resp, err = mock.DeleteProject(ctx, "project-456")
	assert.NoError(t, err)
	assert.Equal(t, "delete-token", resp.SyncToken)
}

// TestMockClient_TaskOperations はタスク操作のテスト
func TestMockClient_TaskOperations(t *testing.T) {
	mock := NewMockClient()
	ctx := context.Background()

	// CreateTask のテスト
	mock.CreateTaskFunc = func(_ context.Context, req *CreateTaskRequest) (*SyncResponse, error) {
		assert.Equal(t, "Test Task", req.Content)
		return &SyncResponse{SyncToken: "task-create-token"}, nil
	}

	resp, err := mock.CreateTask(ctx, &CreateTaskRequest{Content: "Test Task"})
	assert.NoError(t, err)
	assert.Equal(t, "task-create-token", resp.SyncToken)

	// CloseTask のテスト
	mock.CloseTaskFunc = func(_ context.Context, taskID string) (*SyncResponse, error) {
		assert.Equal(t, "task-789", taskID)
		return &SyncResponse{SyncToken: "task-close-token"}, nil
	}

	resp, err = mock.CloseTask(ctx, "task-789")
	assert.NoError(t, err)
	assert.Equal(t, "task-close-token", resp.SyncToken)
}

// TestMockClient_GetOperations はデータ取得操作のテスト
func TestMockClient_GetOperations(t *testing.T) {
	mock := NewMockClient()
	ctx := context.Background()

	// GetAllProjects のテスト
	expectedProjects := []Project{
		{ID: "1", Name: "Project 1"},
		{ID: "2", Name: "Project 2"},
	}

	mock.GetAllProjectsFunc = func(_ context.Context) ([]Project, error) {
		return expectedProjects, nil
	}

	projects, err := mock.GetAllProjects(ctx)
	assert.NoError(t, err)
	assert.Len(t, projects, 2)
	assert.Equal(t, "Project 1", projects[0].Name)

	// GetTasks のテスト
	expectedTasks := []Item{
		{ID: "task-1", Content: "Task 1"},
		{ID: "task-2", Content: "Task 2"},
	}

	mock.GetTasksFunc = func(_ context.Context) ([]Item, error) {
		return expectedTasks, nil
	}

	tasks, err := mock.GetTasks(ctx)
	assert.NoError(t, err)
	assert.Len(t, tasks, 2)
	assert.Equal(t, "Task 1", tasks[0].Content)
}

// TestMockClient_ConfigOperations は設定操作のテスト
func TestMockClient_ConfigOperations(t *testing.T) {
	mock := NewMockClient()

	// SetBaseURL のテスト
	mock.SetBaseURLFunc = func(baseURL string) error {
		assert.Equal(t, "https://test-api.example.com", baseURL)
		return nil
	}

	err := mock.SetBaseURL("https://test-api.example.com")
	assert.NoError(t, err)

	// SetTimeout のテスト
	mock.SetTimeoutFunc = func(timeout time.Duration) {
		assert.Equal(t, 30*time.Second, timeout)
	}

	mock.SetTimeout(30 * time.Second)
}
