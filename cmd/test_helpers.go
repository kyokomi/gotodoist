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
	"github.com/stretchr/testify/require"
)

// testExecutorSetup はテスト用のexecutor設定を保持する共通構造体
type testExecutorSetup struct {
	stdout     *bytes.Buffer
	stderr     *bytes.Buffer
	cleanup    func()
	mockClient *api.MockClient
	dbPath     string
	output     *cli.Output
	repository *repository.Repository
	cfg        *config.Config
}

// setupTestExecutorBase はテスト用のexecutorをセットアップする共通ヘルパー関数
func setupTestExecutorBase(t *testing.T) *testExecutorSetup {
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

	cleanup := func() {
		if err := repo.Close(); err != nil {
			t.Logf("failed to close repo: %v", err)
		}
	}

	return &testExecutorSetup{
		stdout:     stdout,
		stderr:     stderr,
		cleanup:    cleanup,
		mockClient: mockClient,
		dbPath:     dbPath,
		output:     output,
		repository: repo,
		cfg:        cfg,
	}
}

// insertTestProjectsIntoDB はテスト用のプロジェクトを直接DBに挿入するヘルパー関数
func insertTestProjectsIntoDB(t *testing.T, dbPath string, projects []api.Project) {
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
	for i, project := range projects {
		if project.ID == "" {
			project.ID = fmt.Sprintf("project-%d", i+1)
		}

		// 直接DBにプロジェクトを挿入
		err := db.InsertProject(project)
		require.NoError(t, err)
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
