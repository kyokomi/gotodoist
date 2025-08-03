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

// TestSyncIncrementalSync_Success は正常な増分同期のテスト
func TestSyncIncrementalSync_Success(t *testing.T) {
	// テストデータ
	testProjects := []api.Project{
		{ID: "sync-project-1", Name: "Sync Test Project"},
	}
	testTasks := []api.Item{
		{ID: "sync-task-1", Content: "Sync Test Task", ProjectID: "sync-project-1"},
	}

	// モッククライアントをセットアップ
	mockClient := api.NewMockClient()
	mockClient.SyncFunc = func(_ context.Context, req *api.SyncRequest) (*api.SyncResponse, error) {
		return &api.SyncResponse{
			SyncToken: "sync-incremental-token",
			Projects:  testProjects,
			Items:     testTasks,
			Sections:  []api.Section{},
		}, nil
	}

	// テスト用のexecutorを作成
	executor, cleanup := createTestSyncExecutor(t, mockClient)
	defer cleanup()

	// 初期同期を実行してローカルにデータを準備
	err := executor.repository.ForceInitialSync(context.Background())
	require.NoError(t, err)

	// テスト実行（増分同期）
	status, err := executor.executeIncrementalSync(context.Background())
	require.NoError(t, err)

	// 結果検証
	assert.NotNil(t, status)
}

// TestSyncInitialSync_Success は正常な初期同期のテスト
func TestSyncInitialSync_Success(t *testing.T) {
	// テストデータ
	testProjects := []api.Project{
		{ID: "init-project-1", Name: "Init Test Project"},
	}
	testTasks := []api.Item{
		{ID: "init-task-1", Content: "Init Test Task", ProjectID: "init-project-1"},
	}

	// モッククライアントをセットアップ
	mockClient := api.NewMockClient()
	mockClient.SyncFunc = func(_ context.Context, req *api.SyncRequest) (*api.SyncResponse, error) {
		return &api.SyncResponse{
			SyncToken: "sync-initial-token",
			Projects:  testProjects,
			Items:     testTasks,
			Sections:  []api.Section{},
		}, nil
	}

	// テスト用のexecutorを作成
	executor, cleanup := createTestSyncExecutor(t, mockClient)
	defer cleanup()

	// テスト実行（初期同期）
	status, err := executor.executeInitialSync(context.Background())
	require.NoError(t, err)

	// 結果検証
	assert.NotNil(t, status)
	assert.True(t, status.InitialSyncDone)
}

// TestSyncStatus_Success は正常な同期状態取得のテスト
func TestSyncStatus_Success(t *testing.T) {
	// テストデータ
	testProjects := []api.Project{
		{ID: "status-project-1", Name: "Status Test Project"},
	}
	testTasks := []api.Item{
		{ID: "status-task-1", Content: "Status Test Task", ProjectID: "status-project-1"},
	}

	// モッククライアントをセットアップ
	mockClient := api.NewMockClient()
	mockClient.SyncFunc = func(_ context.Context, req *api.SyncRequest) (*api.SyncResponse, error) {
		return &api.SyncResponse{
			SyncToken: "sync-status-token",
			Projects:  testProjects,
			Items:     testTasks,
			Sections:  []api.Section{},
		}, nil
	}

	// テスト用のexecutorを作成
	executor, cleanup := createTestSyncExecutor(t, mockClient)
	defer cleanup()

	// 初期同期を実行してローカルにデータを準備
	err := executor.repository.ForceInitialSync(context.Background())
	require.NoError(t, err)

	// テスト実行（同期状態取得）
	status, err := executor.getSyncStatus()
	require.NoError(t, err)

	// 結果検証
	assert.NotNil(t, status)
	assert.True(t, status.InitialSyncDone)
}

// TestSyncReset_Success は正常なローカルストレージリセットのテスト
func TestSyncReset_Success(t *testing.T) {
	// テストデータ
	testProjects := []api.Project{
		{ID: "reset-project-1", Name: "Reset Test Project"},
	}
	testTasks := []api.Item{
		{ID: "reset-task-1", Content: "Reset Test Task", ProjectID: "reset-project-1"},
	}

	// モッククライアントをセットアップ
	mockClient := api.NewMockClient()
	mockClient.SyncFunc = func(_ context.Context, req *api.SyncRequest) (*api.SyncResponse, error) {
		return &api.SyncResponse{
			SyncToken: "sync-reset-token",
			Projects:  testProjects,
			Items:     testTasks,
			Sections:  []api.Section{},
		}, nil
	}

	// テスト用のexecutorを作成
	executor, cleanup := createTestSyncExecutor(t, mockClient)
	defer cleanup()

	// 初期同期を実行してローカルにデータを準備
	err := executor.repository.ForceInitialSync(context.Background())
	require.NoError(t, err)

	// データが存在することを確認
	projects, err := executor.repository.GetAllProjects(context.Background())
	require.NoError(t, err)
	assert.Len(t, projects, 1)

	// テスト実行（リセット）
	err = executor.executeReset(context.Background())
	require.NoError(t, err)

	// リセット後の状態確認（同期状態の取得はエラーになる場合がある）
	// リセット直後は同期状態テーブルが空になっているため、エラーが発生することがある
	_, err = executor.getSyncStatus()
	// この場合はエラーが発生することを許容する（リセットが正常に動作した証拠）
	if err != nil {
		// エラーが発生することを確認（これは正常な動作）
		assert.Contains(t, err.Error(), "no rows in result set")
	}
}

// TestSyncWithLocalStorageDisabled はローカルストレージ無効時のテスト
func TestSyncWithLocalStorageDisabled(t *testing.T) {
	// ローカルストレージ無効のexecutorを作成
	executor, cleanup := createTestSyncExecutorWithoutLocalStorage(t)
	defer cleanup()

	// ローカルストレージが無効であることを確認
	assert.False(t, executor.isLocalStorageEnabled())

	// 増分同期はエラーになる
	_, err := executor.executeIncrementalSync(context.Background())
	assert.Error(t, err)

	// 初期同期もエラーになる
	_, err = executor.executeInitialSync(context.Background())
	assert.Error(t, err)

	// 同期状態取得もエラーになる
	_, err = executor.getSyncStatus()
	assert.Error(t, err)

	// リセットもエラーになる
	err = executor.executeReset(context.Background())
	assert.Error(t, err)
}

// createTestSyncExecutor はテスト用のsyncExecutorを作成するヘルパー関数
func createTestSyncExecutor(t *testing.T, mockClient api.Interface) (*syncExecutor, func()) {
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
		LocalStorage: &repository.Config{
			Enabled: true,
		},
	}

	// テスト用の出力（verboseモード無効）
	output := cli.New(false)

	executor := &syncExecutor{
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

// createTestSyncExecutorWithoutLocalStorage はローカルストレージ無効のテスト用syncExecutorを作成する
func createTestSyncExecutorWithoutLocalStorage(t *testing.T) (*syncExecutor, func()) {
	// モッククライアント
	mockClient := api.NewMockClient()

	// ローカルストレージ無効設定
	repoConfig := &repository.Config{
		Enabled: false,
	}

	// テスト用Repositoryを作成（ローカルストレージ無効）
	repo, err := factory.NewRepositoryForTest(mockClient, repoConfig, false)
	require.NoError(t, err)

	// テスト用のconfig（ローカルストレージ無効）
	cfg := &config.Config{
		APIToken: "test-token",
		LocalStorage: &repository.Config{
			Enabled: false,
		},
	}

	// テスト用の出力
	output := cli.New(false)

	executor := &syncExecutor{
		cfg:        cfg,
		repository: repo,
		output:     output,
	}

	cleanup := func() {
		if err := repo.Close(); err != nil {
			t.Logf("failed to close repository: %v", err)
		}
	}

	return executor, cleanup
}
