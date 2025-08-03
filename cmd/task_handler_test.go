package cmd

import (
	"context"
	"testing"

	"github.com/kyokomi/gotodoist/internal/api"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// testTaskExecutorSetup はテスト用のtaskExecutor設定を保持する構造体
type testTaskExecutorSetup struct {
	executor *taskExecutor
	*testExecutorSetup
}

// setupTestTaskExecutor はテスト用のtaskExecutorをセットアップするヘルパー関数
func setupTestTaskExecutor(t *testing.T) *testTaskExecutorSetup {
	t.Helper()

	base := setupTestExecutorBase(t)

	executor := &taskExecutor{
		cfg:        base.cfg,
		repository: base.repository,
		output:     base.output,
	}

	return &testTaskExecutorSetup{
		executor:          executor,
		testExecutorSetup: base,
	}
}

func TestExecuteTaskAddWithOutput_Success(t *testing.T) {
	tests := []struct {
		name           string
		params         *taskAddParams
		expectedOutput []string
	}{
		{
			name: "通常のタスク追加",
			params: &taskAddParams{
				content:     "Test Task",
				projectID:   "",
				priority:    "2",
				dueDate:     "",
				description: "",
				labels:      "",
			},
			expectedOutput: []string{
				"Task created successfully!",
			},
		},
		{
			name: "プロジェクト指定のタスク追加",
			params: &taskAddParams{
				content:     "Project Task",
				projectID:   "Test Project", // プロジェクト名で指定
				priority:    "1",
				dueDate:     "today",
				description: "Test description",
				labels:      "urgent,work",
			},
			expectedOutput: []string{
				"Task created successfully!",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange: テスト環境を準備
			setup := setupTestTaskExecutor(t)
			defer setup.cleanup()

			// プロジェクトが指定されている場合は事前にプロジェクトを作成
			if tt.params.projectID != "" {
				testProject := api.Project{
					ID:   "test-project-id",
					Name: "Test Project",
				}
				insertTestProjectsIntoDB(t, setup.dbPath, []api.Project{testProject})
			}

			// CreateTaskのmockを設定
			setup.mockClient.CreateTaskFunc = func(_ context.Context, req *api.CreateTaskRequest) (*api.SyncResponse, error) {
				return &api.SyncResponse{
					Items: []api.Item{
						{
							ID:      "new-task-id",
							Content: req.Content,
						},
					},
					SyncToken: "test-token",
				}, nil
			}

			// Act: テスト対象を実行
			err := setup.executor.executeTaskAddWithOutput(context.Background(), tt.params)

			// Assert: 結果を検証
			require.NoError(t, err)

			outputStr := setup.stdout.String()
			for _, expected := range tt.expectedOutput {
				assert.Contains(t, outputStr, expected, "期待される出力が含まれていません: %s", expected)
			}
		})
	}
}

// setupTaskTestWithMock はタスクテスト用の共通セットアップヘルパー
func setupTaskTestWithMock(t *testing.T, projectID, taskID string, mockSetup func(*api.MockClient)) *testTaskExecutorSetup {
	t.Helper()

	// テストデータ準備
	testProject := api.Project{
		ID:   projectID,
		Name: "Test Project",
	}
	existingTask := api.Item{
		ID:        taskID,
		Content:   "Test Task",
		ProjectID: projectID,
	}

	// Arrange: テスト環境を準備
	setup := setupTestTaskExecutor(t)

	// プロジェクトとタスクをDBに挿入
	insertTestProjectsIntoDB(t, setup.dbPath, []api.Project{testProject})
	insertTestTasksIntoDB(t, setup.dbPath, []api.Item{existingTask})

	// mockを設定
	mockSetup(setup.mockClient)

	return setup
}

func TestExecuteTaskCompleteWithOutput_Success(t *testing.T) {
	setup := setupTaskTestWithMock(t, "test-project", "task-complete", func(mockClient *api.MockClient) {
		mockClient.CloseTaskFunc = func(_ context.Context, _ string) (*api.SyncResponse, error) {
			return &api.SyncResponse{
				SyncToken: "completed-token",
			}, nil
		}
	})
	defer setup.cleanup()

	params := &taskCompleteParams{
		taskID: "task-complete",
	}

	// Act: テスト対象を実行
	err := setup.executor.executeTaskCompleteWithOutput(context.Background(), params)

	// Assert: 結果を検証
	require.NoError(t, err)

	outputStr := setup.stdout.String()
	assert.Contains(t, outputStr, "Task completed successfully!", "期待される出力が含まれていません")
}

func TestExecuteTaskUncompleteWithOutput_Success(t *testing.T) {
	setup := setupTaskTestWithMock(t, "test-project-2", "task-uncomplete", func(mockClient *api.MockClient) {
		mockClient.ReopenTaskFunc = func(_ context.Context, _ string) (*api.SyncResponse, error) {
			return &api.SyncResponse{
				SyncToken: "uncompleted-token",
			}, nil
		}
	})
	defer setup.cleanup()

	params := &taskCompleteParams{
		taskID: "task-uncomplete",
	}

	// Act: テスト対象を実行
	err := setup.executor.executeTaskUncompleteWithOutput(context.Background(), params)

	// Assert: 結果を検証
	require.NoError(t, err)

	outputStr := setup.stdout.String()
	assert.Contains(t, outputStr, "Task marked as uncompleted successfully!", "期待される出力が含まれていません")
}

func TestExecuteTaskDeleteWithOutput_Success(t *testing.T) {
	// 先にプロジェクトを作成
	testProject := api.Project{
		ID:   "test-project-3",
		Name: "Test Project 3",
	}
	existingTask := api.Item{
		ID:        "task-delete",
		Content:   "Task to Delete",
		ProjectID: "test-project-3",
	}
	params := &taskDeleteParams{
		taskID: "task-delete",
		force:  true, // 確認プロンプトをスキップ
	}

	// Arrange: テスト環境を準備
	setup := setupTestTaskExecutor(t)
	defer setup.cleanup()

	// プロジェクトとタスクをDBに挿入
	insertTestProjectsIntoDB(t, setup.dbPath, []api.Project{testProject})
	insertTestTasksIntoDB(t, setup.dbPath, []api.Item{existingTask})

	// DeleteTaskのmockを設定
	setup.mockClient.DeleteTaskFunc = func(_ context.Context, _ string) (*api.SyncResponse, error) {
		return &api.SyncResponse{
			SyncToken: "deleted-token",
		}, nil
	}

	// Act: テスト対象を実行
	err := setup.executor.executeTaskDeleteWithOutput(context.Background(), params)

	// Assert: 結果を検証
	require.NoError(t, err)

	outputStr := setup.stdout.String()
	assert.Contains(t, outputStr, "Task deleted successfully!", "期待される出力が含まれていません")
}

func TestExecuteTaskListWithOutput_Success(t *testing.T) {
	// 先にプロジェクトを作成
	testProject := api.Project{
		ID:   "test-project-4",
		Name: "Test Project 4",
	}
	testTasks := []api.Item{
		{
			ID:        "task-1",
			Content:   "Task 1",
			ProjectID: "test-project-4",
		},
		{
			ID:        "task-2",
			Content:   "Task 2",
			ProjectID: "test-project-4",
		},
	}
	params := &taskListParams{
		projectFilter:    "",
		filterExpression: "",
		showAll:          false,
	}

	// Arrange: テスト環境を準備
	setup := setupTestTaskExecutor(t)
	defer setup.cleanup()

	// プロジェクトとタスクをDBに挿入
	insertTestProjectsIntoDB(t, setup.dbPath, []api.Project{testProject})
	insertTestTasksIntoDB(t, setup.dbPath, testTasks)

	// Act: テスト対象を実行
	err := setup.executor.executeTaskListWithOutput(context.Background(), params)

	// Assert: 結果を検証
	require.NoError(t, err)

	outputStr := setup.stdout.String()
	assert.Contains(t, outputStr, "Task 1", "タスク1が出力に含まれていません")
	assert.Contains(t, outputStr, "Task 2", "タスク2が出力に含まれていません")
}

func TestExecuteTaskUpdateWithOutput_Success(t *testing.T) {
	// 先にプロジェクトを作成
	testProject := api.Project{
		ID:   "test-project-5",
		Name: "Test Project 5",
	}
	existingTask := api.Item{
		ID:        "task-update",
		Content:   "Task to Update",
		ProjectID: "test-project-5",
	}
	params := &taskUpdateParams{
		taskID:      "task-update",
		content:     "Updated Task Content",
		priority:    "3",
		dueDate:     "today",
		description: "Updated description",
		labels:      "updated,test",
	}

	// Arrange: テスト環境を準備
	setup := setupTestTaskExecutor(t)
	defer setup.cleanup()

	// プロジェクトとタスクをDBに挿入
	insertTestProjectsIntoDB(t, setup.dbPath, []api.Project{testProject})
	insertTestTasksIntoDB(t, setup.dbPath, []api.Item{existingTask})

	// UpdateTaskのmockを設定
	setup.mockClient.UpdateTaskFunc = func(_ context.Context, _ string, _ *api.UpdateTaskRequest) (*api.SyncResponse, error) {
		return &api.SyncResponse{
			SyncToken: "updated-token",
		}, nil
	}

	// Act: テスト対象を実行
	err := setup.executor.executeTaskUpdateWithOutput(context.Background(), params)

	// Assert: 結果を検証
	require.NoError(t, err)

	outputStr := setup.stdout.String()
	assert.Contains(t, outputStr, "Task updated successfully!", "期待される出力が含まれていません")
}
