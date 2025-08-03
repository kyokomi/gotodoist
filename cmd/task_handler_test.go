package cmd

import (
	"bytes"
	"context"
	"fmt"
	"path/filepath"
	"testing"

	"github.com/kyokomi/gotodoist/internal/api"
	"github.com/kyokomi/gotodoist/internal/cli"
	"github.com/kyokomi/gotodoist/internal/config"
	"github.com/kyokomi/gotodoist/internal/factory"
	"github.com/kyokomi/gotodoist/internal/repository"
	"github.com/kyokomi/gotodoist/internal/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// testTaskExecutorSetup はテスト用のtaskExecutor設定を保持する構造体
type testTaskExecutorSetup struct {
	executor   *taskExecutor
	stdout     *bytes.Buffer
	stderr     *bytes.Buffer
	cleanup    func()
	mockClient *api.MockClient
	dbPath     string
}

// setupTestTaskExecutor はテスト用のtaskExecutorをセットアップするヘルパー関数
func setupTestTaskExecutor(t *testing.T) *testTaskExecutorSetup {
	t.Helper()

	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")

	mockClient := api.NewMockClient()

	cfg := &config.Config{
		APIToken: "test-token",
		LocalStorage: &repository.Config{
			Enabled:      true,
			DatabasePath: dbPath,
		},
	}

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	output := cli.NewWithWriters(stdout, stderr, false)

	repo, err := factory.NewRepositoryForTest(mockClient, cfg.LocalStorage, false)
	require.NoError(t, err)

	err = repo.Initialize(context.Background())
	require.NoError(t, err)

	executor := &taskExecutor{
		cfg:        cfg,
		repository: repo,
		output:     output,
	}

	cleanup := func() {
		if err := repo.Close(); err != nil {
			t.Logf("failed to close repo: %v", err)
		}
	}

	return &testTaskExecutorSetup{
		executor:   executor,
		stdout:     stdout,
		stderr:     stderr,
		cleanup:    cleanup,
		mockClient: mockClient,
		dbPath:     dbPath,
	}
}

// insertTestTasksIntoDB はテスト用のタスクを直接DBに挿入するヘルパー関数
func insertTestTasksIntoDB(t *testing.T, dbPath string, tasks []api.Item) {
	t.Helper()

	// SQLiteDBを直接開く
	db, err := storage.NewSQLiteDB(dbPath)
	require.NoError(t, err)
	defer func() {
		if err := db.Close(); err != nil {
			t.Logf("failed to close db: %v", err)
		}
	}()

	// IDが空の場合は自動採番
	for i, task := range tasks {
		if task.ID == "" {
			task.ID = fmt.Sprintf("task-%d", i+1)
		}

		// 直接DBにタスクを挿入
		err := db.InsertTask(task)
		require.NoError(t, err)
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

func TestExecuteTaskCompleteWithOutput_Success(t *testing.T) {
	// 先にプロジェクトを作成
	testProject := api.Project{
		ID:   "test-project",
		Name: "Test Project",
	}
	existingTask := api.Item{
		ID:        "task-complete",
		Content:   "Task to Complete",
		ProjectID: "test-project",
	}
	params := &taskCompleteParams{
		taskID: "task-complete",
	}

	// Arrange: テスト環境を準備
	setup := setupTestTaskExecutor(t)
	defer setup.cleanup()

	// プロジェクトとタスクをDBに挿入
	insertTestProjectsIntoDB(t, setup.dbPath, []api.Project{testProject})
	insertTestTasksIntoDB(t, setup.dbPath, []api.Item{existingTask})

	// CloseTaskのmockを設定
	setup.mockClient.CloseTaskFunc = func(_ context.Context, _ string) (*api.SyncResponse, error) {
		return &api.SyncResponse{
			SyncToken: "completed-token",
		}, nil
	}

	// Act: テスト対象を実行
	err := setup.executor.executeTaskCompleteWithOutput(context.Background(), params)

	// Assert: 結果を検証
	require.NoError(t, err)

	outputStr := setup.stdout.String()
	assert.Contains(t, outputStr, "Task completed successfully!", "期待される出力が含まれていません")
}

func TestExecuteTaskUncompleteWithOutput_Success(t *testing.T) {
	// 先にプロジェクトを作成
	testProject := api.Project{
		ID:   "test-project-2",
		Name: "Test Project 2",
	}
	existingTask := api.Item{
		ID:        "task-uncomplete",
		Content:   "Task to Uncomplete",
		ProjectID: "test-project-2",
	}
	params := &taskCompleteParams{
		taskID: "task-uncomplete",
	}

	// Arrange: テスト環境を準備
	setup := setupTestTaskExecutor(t)
	defer setup.cleanup()

	// プロジェクトとタスクをDBに挿入
	insertTestProjectsIntoDB(t, setup.dbPath, []api.Project{testProject})
	insertTestTasksIntoDB(t, setup.dbPath, []api.Item{existingTask})

	// ReopenTaskのmockを設定
	setup.mockClient.ReopenTaskFunc = func(_ context.Context, _ string) (*api.SyncResponse, error) {
		return &api.SyncResponse{
			SyncToken: "uncompleted-token",
		}, nil
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
