package cmd

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kyokomi/gotodoist/internal/api"
	"github.com/kyokomi/gotodoist/internal/cli"
	"github.com/kyokomi/gotodoist/internal/config"
	"github.com/kyokomi/gotodoist/internal/factory"
	"github.com/kyokomi/gotodoist/internal/repository"
)

// TestTaskAdd_Success は正常なタスク作成のテスト
func TestTaskAdd_Success(t *testing.T) {
	// モッククライアントをセットアップ
	mockClient := api.NewMockClient()
	mockClient.CreateTaskFunc = func(_ context.Context, req *api.CreateTaskRequest) (*api.SyncResponse, error) {
		// リクエストの検証
		assert.Equal(t, "Test Task", req.Content)
		assert.Equal(t, "Task description", req.Description)
		assert.Equal(t, 2, req.Priority)

		return &api.SyncResponse{
			SyncToken: "task-create-token",
		}, nil
	}

	// テスト用のexecutorを作成
	executor, cleanup := createTestTaskExecutor(t, mockClient)
	defer cleanup()

	// テスト実行
	params := &taskAddParams{
		content:     "Test Task",
		description: "Task description",
		priority:    "2",
	}
	resp, err := executor.executeTaskAdd(context.Background(), params)

	// 結果検証
	require.NoError(t, err)
	assert.Equal(t, "task-create-token", resp.SyncToken)
}

// TestTaskAdd_WithProject はプロジェクト指定付きのタスク作成テスト
func TestTaskAdd_WithProject(t *testing.T) {
	// 既存プロジェクトのテストデータ
	existingProjects := []api.Project{
		{ID: "project-123", Name: "Work Project"},
	}

	// モッククライアントをセットアップ
	mockClient := api.NewMockClient()
	mockClient.SyncFunc = func(_ context.Context, _ *api.SyncRequest) (*api.SyncResponse, error) {
		return &api.SyncResponse{
			SyncToken: "initial-token",
			Projects:  existingProjects,
			Items:     []api.Item{},
			Sections:  []api.Section{},
		}, nil
	}
	mockClient.CreateTaskFunc = func(_ context.Context, req *api.CreateTaskRequest) (*api.SyncResponse, error) {
		assert.Equal(t, "Project Task", req.Content)
		assert.Equal(t, "project-123", req.ProjectID)

		return &api.SyncResponse{
			SyncToken: "task-project-token",
		}, nil
	}

	// テスト用のexecutorを作成
	executor, cleanup := createTestTaskExecutor(t, mockClient)
	defer cleanup()

	// 初期同期を実行
	err := executor.repository.ForceInitialSync(context.Background())
	require.NoError(t, err)

	// テスト実行
	params := &taskAddParams{
		content:   "Project Task",
		projectID: "Work Project", // プロジェクト名で指定
	}
	resp, err := executor.executeTaskAdd(context.Background(), params)

	// 結果検証
	require.NoError(t, err)
	assert.Equal(t, "task-project-token", resp.SyncToken)
}

// TestTaskComplete_Success は正常なタスク完了のテスト
func TestTaskComplete_Success(t *testing.T) {
	// テスト用プロジェクト（外部キー制約のため必要）
	testProjects := []api.Project{
		{ID: "project-1", Name: "Test Project"},
	}

	// 既存タスクのテストデータ
	existingTasks := []api.Item{
		{ID: "task-456", Content: "Incomplete Task", ProjectID: "project-1", DateCompleted: nil},
	}

	// モッククライアントをセットアップ
	mockClient := api.NewMockClient()
	mockClient.SyncFunc = func(_ context.Context, _ *api.SyncRequest) (*api.SyncResponse, error) {
		return &api.SyncResponse{
			SyncToken: "initial-token",
			Projects:  testProjects,
			Items:     existingTasks,
			Sections:  []api.Section{},
		}, nil
	}
	mockClient.CloseTaskFunc = func(_ context.Context, taskID string) (*api.SyncResponse, error) {
		assert.Equal(t, "task-456", taskID)
		return &api.SyncResponse{SyncToken: "task-complete-token"}, nil
	}

	// テスト用のexecutorを作成
	executor, cleanup := createTestTaskExecutor(t, mockClient)
	defer cleanup()

	// 初期同期を実行
	err := executor.repository.ForceInitialSync(context.Background())
	require.NoError(t, err)

	// テスト実行
	params := &taskCompleteParams{
		taskID: "task-456",
	}
	resp, err := executor.executeTaskComplete(context.Background(), params)

	// 結果検証
	require.NoError(t, err)
	assert.Equal(t, "task-complete-token", resp.SyncToken)
}

// TestTaskUncomplete_Success は正常なタスク未完了のテスト
func TestTaskUncomplete_Success(t *testing.T) {
	// テスト用プロジェクト（外部キー制約のため必要）
	testProjects := []api.Project{
		{ID: "project-2", Name: "Test Project 2"},
	}

	// 完了済みタスクのテストデータ
	now := api.TodoistTime{}
	completedTasks := []api.Item{
		{ID: "task-789", Content: "Completed Task", ProjectID: "project-2", DateCompleted: &now},
	}

	// モッククライアントをセットアップ
	mockClient := api.NewMockClient()
	mockClient.SyncFunc = func(_ context.Context, _ *api.SyncRequest) (*api.SyncResponse, error) {
		return &api.SyncResponse{
			SyncToken: "initial-token",
			Projects:  testProjects,
			Items:     completedTasks,
			Sections:  []api.Section{},
		}, nil
	}
	mockClient.ReopenTaskFunc = func(_ context.Context, taskID string) (*api.SyncResponse, error) {
		assert.Equal(t, "task-789", taskID)
		return &api.SyncResponse{SyncToken: "task-reopen-token"}, nil
	}

	// テスト用のexecutorを作成
	executor, cleanup := createTestTaskExecutor(t, mockClient)
	defer cleanup()

	// 初期同期を実行
	err := executor.repository.ForceInitialSync(context.Background())
	require.NoError(t, err)

	// テスト実行
	params := &taskCompleteParams{
		taskID: "task-789",
	}
	resp, err := executor.executeTaskUncomplete(context.Background(), params)

	// 結果検証
	require.NoError(t, err)
	assert.Equal(t, "task-reopen-token", resp.SyncToken)
}

// TestTaskUpdate_Success は正常なタスク更新のテスト
func TestTaskUpdate_Success(t *testing.T) {
	// テスト用プロジェクト（外部キー制約のため必要）
	testProjects := []api.Project{
		{ID: "project-3", Name: "Test Project 3"},
	}

	// 既存タスクのテストデータ
	existingTasks := []api.Item{
		{ID: "task-update", Content: "Old Task Content", ProjectID: "project-3", Priority: 1},
	}

	// モッククライアントをセットアップ
	mockClient := api.NewMockClient()
	mockClient.SyncFunc = func(_ context.Context, _ *api.SyncRequest) (*api.SyncResponse, error) {
		return &api.SyncResponse{
			SyncToken: "initial-token",
			Projects:  testProjects,
			Items:     existingTasks,
			Sections:  []api.Section{},
		}, nil
	}
	mockClient.UpdateTaskFunc = func(_ context.Context, taskID string, req *api.UpdateTaskRequest) (*api.SyncResponse, error) {
		assert.Equal(t, "task-update", taskID)
		assert.Equal(t, "New Task Content", req.Content)
		assert.Equal(t, 3, req.Priority)

		return &api.SyncResponse{
			SyncToken: "task-update-token",
		}, nil
	}

	// テスト用のexecutorを作成
	executor, cleanup := createTestTaskExecutor(t, mockClient)
	defer cleanup()

	// 初期同期を実行
	err := executor.repository.ForceInitialSync(context.Background())
	require.NoError(t, err)

	// テスト実行
	params := &taskUpdateParams{
		taskID:   "task-update",
		content:  "New Task Content",
		priority: "3",
	}
	resp, err := executor.executeTaskUpdate(context.Background(), params)

	// 結果検証
	require.NoError(t, err)
	assert.Equal(t, "task-update-token", resp.SyncToken)
}

// TestTaskDelete_Success は正常なタスク削除のテスト
func TestTaskDelete_Success(t *testing.T) {
	// テスト用プロジェクト（外部キー制約のため必要）
	testProjects := []api.Project{
		{ID: "project-4", Name: "Test Project 4"},
	}

	// 既存タスクのテストデータ
	existingTasks := []api.Item{
		{ID: "task-delete", Content: "Task to Delete", ProjectID: "project-4"},
	}

	// モッククライアントをセットアップ
	mockClient := api.NewMockClient()
	mockClient.SyncFunc = func(_ context.Context, _ *api.SyncRequest) (*api.SyncResponse, error) {
		return &api.SyncResponse{
			SyncToken: "initial-token",
			Projects:  testProjects,
			Items:     existingTasks,
			Sections:  []api.Section{},
		}, nil
	}
	mockClient.DeleteTaskFunc = func(_ context.Context, taskID string) (*api.SyncResponse, error) {
		assert.Equal(t, "task-delete", taskID)
		return &api.SyncResponse{SyncToken: "task-delete-token"}, nil
	}

	// テスト用のexecutorを作成
	executor, cleanup := createTestTaskExecutor(t, mockClient)
	defer cleanup()

	// 初期同期を実行
	err := executor.repository.ForceInitialSync(context.Background())
	require.NoError(t, err)

	// 削除対象の確認（force=trueなので確認処理はスキップされる）
	params := &taskDeleteParams{
		taskID: "task-delete",
		force:  true,
	}
	task, shouldDelete, err := executor.confirmTaskDeletion(context.Background(), params)
	require.NoError(t, err)
	assert.True(t, shouldDelete)
	assert.Equal(t, "Task to Delete", task.Content)

	// 実際の削除処理
	resp, err := executor.deleteTask(context.Background(), task.ID)
	require.NoError(t, err)
	assert.Equal(t, "task-delete-token", resp.SyncToken)
}

// TestTaskList_Success は正常なタスク一覧取得のテスト
func TestTaskList_Success(t *testing.T) {
	// テストデータ
	now := api.TodoistTime{}
	mockTasks := []api.Item{
		{ID: "task-1", Content: "Task 1", ProjectID: "project-1", DateCompleted: nil, Priority: 1},
		{ID: "task-2", Content: "Task 2", ProjectID: "project-1", DateCompleted: &now, Priority: 2},
		{ID: "task-3", Content: "Task 3", ProjectID: "project-1", DateCompleted: nil, Priority: 3},
	}
	mockProjects := []api.Project{
		{ID: "project-1", Name: "Project Alpha"},
	}

	// モッククライアントをセットアップ
	mockClient := api.NewMockClient()
	mockClient.SyncFunc = func(_ context.Context, _ *api.SyncRequest) (*api.SyncResponse, error) {
		return &api.SyncResponse{
			SyncToken: "initial-token",
			Projects:  mockProjects,
			Items:     mockTasks,
			Sections:  []api.Section{},
		}, nil
	}

	// テスト用のexecutorを作成
	executor, cleanup := createTestTaskExecutor(t, mockClient)
	defer cleanup()

	// 初期同期を実行
	err := executor.repository.ForceInitialSync(context.Background())
	require.NoError(t, err)

	// テスト実行
	params := &taskListParams{
		showAll: false, // 未完了のみ
	}
	data, err := executor.fetchAllTaskListData(context.Background(), params)
	require.NoError(t, err)

	// フィルタリング適用（未完了のみ）
	filteredTasks := applyTaskFilters(data.tasks, params)

	// 結果検証（未完了タスクのみ）
	assert.Len(t, filteredTasks, 2) // task-1とtask-3のみ
	var taskContents []string
	for _, task := range filteredTasks {
		taskContents = append(taskContents, task.Content)
	}
	assert.Contains(t, taskContents, "Task 1")
	assert.Contains(t, taskContents, "Task 3")
	assert.NotContains(t, taskContents, "Task 2") // 完了済みなので除外
}

// TestFindTaskByID_Success はタスクID検索のテスト
func TestFindTaskByID_Success(t *testing.T) {
	// テスト用プロジェクト（外部キー制約のため必要）
	testProjects := []api.Project{
		{ID: "project-5", Name: "Test Project 5"},
	}

	// テストデータ
	testTasks := []api.Item{
		{ID: "task-alpha", Content: "Alpha Task", ProjectID: "project-5"},
		{ID: "task-beta", Content: "Beta Task", ProjectID: "project-5"},
		{ID: "task-gamma", Content: "Gamma Task", ProjectID: "project-5"},
	}

	// モッククライアントをセットアップ
	mockClient := api.NewMockClient()
	mockClient.SyncFunc = func(_ context.Context, _ *api.SyncRequest) (*api.SyncResponse, error) {
		return &api.SyncResponse{
			SyncToken: "initial-token",
			Projects:  testProjects,
			Items:     testTasks,
			Sections:  []api.Section{},
		}, nil
	}

	// テスト用のexecutorを作成
	executor, cleanup := createTestTaskExecutor(t, mockClient)
	defer cleanup()

	// 初期同期を実行
	err := executor.repository.ForceInitialSync(context.Background())
	require.NoError(t, err)

	// テストケース: 存在するタスクID
	t.Run("存在するタスクID", func(t *testing.T) {
		task, err := executor.findTaskByID(context.Background(), "task-beta")
		require.NoError(t, err)
		assert.Equal(t, "Beta Task", task.Content)
	})

	// テストケース: 存在しないタスクID
	t.Run("存在しないタスクID", func(t *testing.T) {
		task, err := executor.findTaskByID(context.Background(), "nonexistent-task")
		require.NoError(t, err)
		assert.Nil(t, task) // タスクが見つからない場合はnilが返される
	})
}

// createTestTaskExecutor はテスト用のtaskExecutorを作成するヘルパー関数
func createTestTaskExecutor(t *testing.T, mockClient api.Interface) (*taskExecutor, func()) {
	// テスト用の一時ディレクトリを作成
	tempDir := t.TempDir()

	// テスト用設定（ローカルストレージ有効、一時ディレクトリ使用）
	repoConfig := &repository.Config{
		Enabled:      true,
		DatabasePath: filepath.Join(tempDir, "test.db"),
	}

	// テスト用Repositoryを作成
	repo, err := factory.NewRepositoryForTest(mockClient, repoConfig, false)
	require.NoError(t, err)

	// Repositoryの初期化
	ctx := context.Background()
	err = repo.Initialize(ctx)
	require.NoError(t, err)

	// テスト用のconfig
	cfg := &config.Config{
		APIToken: "test-token",
	}

	// テスト用の出力（verboseモード無効）
	output := cli.New(false)

	executor := &taskExecutor{
		cfg:        cfg,
		repository: repo,
		output:     output,
	}

	cleanup := func() {
		if err := repo.Close(); err != nil {
			t.Logf("failed to close repository: %v", err)
		}
		// t.TempDir()で作成されたディレクトリは自動的にクリーンアップされる
	}

	return executor, cleanup
}
