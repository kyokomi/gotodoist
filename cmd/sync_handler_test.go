package cmd

import (
	"bytes"
	"context"
	"path/filepath"
	"testing"

	"github.com/kyokomi/gotodoist/internal/api"
	"github.com/kyokomi/gotodoist/internal/cli"
	"github.com/kyokomi/gotodoist/internal/config"
	"github.com/kyokomi/gotodoist/internal/factory"
	"github.com/kyokomi/gotodoist/internal/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// testSyncExecutorSetup はテスト用のsyncExecutor設定を保持する構造体
type testSyncExecutorSetup struct {
	executor   *syncExecutor
	stdout     *bytes.Buffer
	stderr     *bytes.Buffer
	cleanup    func()
	mockClient *api.MockClient
	dbPath     string
}

// setupTestSyncExecutor はテスト用のsyncExecutorをセットアップするヘルパー関数
func setupTestSyncExecutor(t *testing.T) *testSyncExecutorSetup {
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

	executor := &syncExecutor{
		cfg:        cfg,
		repository: repo,
		output:     output,
	}

	cleanup := func() {
		if err := repo.Close(); err != nil {
			t.Logf("failed to close repo: %v", err)
		}
	}

	return &testSyncExecutorSetup{
		executor:   executor,
		stdout:     stdout,
		stderr:     stderr,
		cleanup:    cleanup,
		mockClient: mockClient,
		dbPath:     dbPath,
	}
}

func TestExecuteSyncWithOutput_Success(t *testing.T) {
	params := &syncParams{}

	// Arrange: テスト環境を準備
	setup := setupTestSyncExecutor(t)
	defer setup.cleanup()

	// Act: テスト対象を実行
	err := setup.executor.executeSyncWithOutput(context.Background(), params)

	// Assert: 結果を検証
	require.NoError(t, err)

	outputStr := setup.stdout.String()
	assert.Contains(t, outputStr, "Synchronization completed successfully!", "期待される出力が含まれていません")
}

func TestExecuteSyncInitWithOutput_Success(t *testing.T) {
	params := &syncInitParams{}

	// Arrange: テスト環境を準備
	setup := setupTestSyncExecutor(t)
	defer setup.cleanup()

	// SyncFuncのmockを設定
	setup.mockClient.SyncFunc = func(_ context.Context, req *api.SyncRequest) (*api.SyncResponse, error) {
		return &api.SyncResponse{
			SyncToken: "initial-sync-token",
			Items:     []api.Item{},
			Projects:  []api.Project{},
			Sections:  []api.Section{},
		}, nil
	}

	// Act: テスト対象を実行
	err := setup.executor.executeSyncInitWithOutput(context.Background(), params)

	// Assert: 結果を検証
	require.NoError(t, err)

	outputStr := setup.stdout.String()
	assert.Contains(t, outputStr, "Initial synchronization completed successfully!", "期待される出力が含まれていません")
}

func TestExecuteSyncStatusWithOutput_Success(t *testing.T) {
	params := &syncStatusParams{}

	// Arrange: テスト環境を準備
	setup := setupTestSyncExecutor(t)
	defer setup.cleanup()

	// Act: テスト対象を実行
	err := setup.executor.executeSyncStatusWithOutput(context.Background(), params)

	// Assert: 結果を検証
	require.NoError(t, err)

	outputStr := setup.stdout.String()
	assert.Contains(t, outputStr, "Sync Status", "期待される出力が含まれていません")
}

func TestExecuteSyncResetWithOutput_Success(t *testing.T) {
	params := &syncResetParams{
		force: true, // 確認プロンプトをスキップ
	}

	// Arrange: テスト環境を準備
	setup := setupTestSyncExecutor(t)
	defer setup.cleanup()

	// Act: テスト対象を実行
	err := setup.executor.executeSyncResetWithOutput(context.Background(), params)

	// Assert: 結果を検証
	require.NoError(t, err)

	outputStr := setup.stdout.String()
	assert.Contains(t, outputStr, "Local storage reset completed!", "期待される出力が含まれていません")
}
