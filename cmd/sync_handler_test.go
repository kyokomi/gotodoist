package cmd

import (
	"context"
	"testing"

	"github.com/kyokomi/gotodoist/internal/api"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// testSyncExecutorSetup はテスト用のsyncExecutor設定を保持する構造体
type testSyncExecutorSetup struct {
	executor *syncExecutor
	*testExecutorSetup
}

// setupTestSyncExecutor はテスト用のsyncExecutorをセットアップするヘルパー関数
func setupTestSyncExecutor(t *testing.T) *testSyncExecutorSetup {
	t.Helper()

	base := setupTestExecutorBase(t)

	executor := &syncExecutor{
		cfg:        base.cfg,
		repository: base.repository,
		output:     base.output,
	}

	return &testSyncExecutorSetup{
		executor:          executor,
		testExecutorSetup: base,
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
	setup.mockClient.SyncFunc = func(_ context.Context, _ *api.SyncRequest) (*api.SyncResponse, error) {
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
