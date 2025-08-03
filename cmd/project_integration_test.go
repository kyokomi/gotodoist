package cmd

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kyokomi/gotodoist/internal/api"
	"github.com/kyokomi/gotodoist/internal/cli"
	"github.com/kyokomi/gotodoist/internal/config"
	"github.com/kyokomi/gotodoist/internal/factory"
	"github.com/kyokomi/gotodoist/internal/repository"
)

// TestProjectAdd_Success は正常なプロジェクト作成のテスト
func TestProjectAdd_Success(t *testing.T) {
	// モッククライアントをセットアップ
	mockClient := api.NewMockClient()
	mockClient.CreateProjectFunc = func(_ context.Context, req *api.CreateProjectRequest) (*api.SyncResponse, error) {
		// リクエストの検証
		assert.Equal(t, "Test Project", req.Name)
		assert.Equal(t, "blue", req.Color)
		assert.True(t, req.IsFavorite)

		return &api.SyncResponse{
			SyncToken: "test-sync-token-12345",
		}, nil
	}

	// テスト用のexecutorを作成
	executor, cleanup := createTestProjectExecutor(t, mockClient)
	defer cleanup()

	// テスト実行
	params := &projectAddParams{
		name:       "Test Project",
		color:      "blue",
		isFavorite: true,
	}
	resp, err := executor.executeProjectAdd(context.Background(), params)

	// 結果検証
	require.NoError(t, err)
	assert.Equal(t, "test-sync-token-12345", resp.SyncToken)
}

// TestProjectAdd_WithParent は親プロジェクト指定付きの作成テスト
func TestProjectAdd_WithParent(t *testing.T) {
	// モッククライアントをセットアップ
	mockClient := api.NewMockClient()

	// 親プロジェクト検索のモック
	mockClient.GetAllProjectsFunc = func(_ context.Context) ([]api.Project, error) {
		return []api.Project{
			{
				ID:   "parent-id-123",
				Name: "Parent Project",
			},
		}, nil
	}

	mockClient.CreateProjectFunc = func(_ context.Context, req *api.CreateProjectRequest) (*api.SyncResponse, error) {
		// 親プロジェクトIDが正しく設定されているか確認
		assert.Equal(t, "Child Project", req.Name)
		assert.Equal(t, "parent-id-123", req.ParentID)

		return &api.SyncResponse{
			SyncToken: "child-project-token",
		}, nil
	}

	// テスト用のexecutorを作成
	executor, cleanup := createTestProjectExecutor(t, mockClient)
	defer cleanup()

	// テスト実行
	params := &projectAddParams{
		name:       "Child Project",
		parentName: "Parent Project",
	}
	resp, err := executor.executeProjectAdd(context.Background(), params)

	// 結果検証
	require.NoError(t, err)
	assert.Equal(t, "child-project-token", resp.SyncToken)
}

// TestProjectAdd_Error はAPIエラーハンドリングのテスト
func TestProjectAdd_Error(t *testing.T) {
	// モッククライアントをセットアップ
	mockClient := api.NewMockClient()
	mockClient.CreateProjectFunc = func(_ context.Context, _ *api.CreateProjectRequest) (*api.SyncResponse, error) {
		return nil, &api.Error{
			StatusCode: 400,
			Message:    "Invalid project name",
		}
	}

	// テスト用のexecutorを作成
	executor, cleanup := createTestProjectExecutor(t, mockClient)
	defer cleanup()

	// テスト実行
	params := &projectAddParams{
		name: "Error Project",
	}
	_, err := executor.executeProjectAdd(context.Background(), params)

	// 結果検証
	assert.Error(t, err)
	var apiErr *api.Error
	assert.ErrorAs(t, err, &apiErr)
	assert.Equal(t, 400, apiErr.StatusCode)
}

// TestProjectList_AllProjects は全プロジェクト表示のテスト
func TestProjectList_AllProjects(t *testing.T) {
	// テストデータ
	mockProjects := []api.Project{
		{ID: "1", Name: "Project 1", IsArchived: false, IsFavorite: false},
		{ID: "2", Name: "Project 2", IsArchived: false, IsFavorite: true},
		{ID: "3", Name: "Project 3", IsArchived: true, IsFavorite: false},
	}

	// モッククライアントをセットアップ
	mockClient := api.NewMockClient()
	mockClient.GetAllProjectsFunc = func(_ context.Context) ([]api.Project, error) {
		return mockProjects, nil
	}

	// テスト用のexecutorを作成
	executor, cleanup := createTestProjectExecutor(t, mockClient)
	defer cleanup()

	// テスト実行
	params := &projectListParams{
		showArchived:  false,
		showFavorites: false,
	}
	data, err := executor.fetchProjectListData(context.Background(), params)
	require.NoError(t, err)

	// フィルタリング適用
	filteredProjects := applyProjectFilters(data.projects, params)

	// 結果検証（アーカイブされていないもののみ）
	assert.Len(t, filteredProjects, 2)
	assert.Equal(t, "Project 1", filteredProjects[0].Name)
	assert.Equal(t, "Project 2", filteredProjects[1].Name)
}

// TestProjectList_FavoritesOnly はお気に入りプロジェクトのみ表示のテスト
func TestProjectList_FavoritesOnly(t *testing.T) {
	// テストデータ
	mockProjects := []api.Project{
		{ID: "1", Name: "Project 1", IsArchived: false, IsFavorite: false},
		{ID: "2", Name: "Project 2", IsArchived: false, IsFavorite: true},
		{ID: "3", Name: "Project 3", IsArchived: true, IsFavorite: false},
	}

	// モッククライアントをセットアップ
	mockClient := api.NewMockClient()
	mockClient.GetAllProjectsFunc = func(_ context.Context) ([]api.Project, error) {
		// お気に入りのみフィルタリング
		var favorites []api.Project
		for _, project := range mockProjects {
			if project.IsFavorite {
				favorites = append(favorites, project)
			}
		}
		return favorites, nil
	}

	// テスト用のexecutorを作成
	executor, cleanup := createTestProjectExecutor(t, mockClient)
	defer cleanup()

	// テスト実行
	params := &projectListParams{
		showArchived:  false,
		showFavorites: true,
	}
	data, err := executor.fetchProjectListData(context.Background(), params)
	require.NoError(t, err)

	// 結果検証（お気に入りのみ）
	assert.Len(t, data.projects, 1)
	assert.Equal(t, "Project 2", data.projects[0].Name)
	assert.True(t, data.projects[0].IsFavorite)
}

// TestProjectDelete_Success は正常な削除のテスト
func TestProjectDelete_Success(t *testing.T) {
	// テストデータ
	mockProjects := []api.Project{
		{ID: "test-id", Name: "Test Project"},
	}

	// モッククライアントをセットアップ
	mockClient := api.NewMockClient()
	mockClient.GetAllProjectsFunc = func(_ context.Context) ([]api.Project, error) {
		return mockProjects, nil
	}
	mockClient.DeleteProjectFunc = func(_ context.Context, projectID string) (*api.SyncResponse, error) {
		assert.Equal(t, "test-id", projectID)
		return &api.SyncResponse{SyncToken: "delete-token"}, nil
	}

	// テスト用のexecutorを作成
	executor, cleanup := createTestProjectExecutor(t, mockClient)
	defer cleanup()

	// 削除対象の確認（force=trueなので確認処理はスキップされる）
	params := &projectDeleteParams{
		projectIDOrName: "Test Project",
		force:           true,
	}
	project, shouldDelete, err := executor.confirmProjectDeletion(context.Background(), params)
	require.NoError(t, err)
	assert.True(t, shouldDelete)
	assert.Equal(t, "Test Project", project.Name)

	// 実際の削除処理
	resp, err := executor.deleteProject(context.Background(), project.ID)
	require.NoError(t, err)
	assert.Equal(t, "delete-token", resp.SyncToken)
}

// TestProjectDelete_NotFound は存在しないプロジェクトの削除テスト
func TestProjectDelete_NotFound(t *testing.T) {
	// テストデータ（対象プロジェクトが存在しない）
	mockProjects := []api.Project{
		{ID: "other-id", Name: "Other Project"},
	}

	// モッククライアントをセットアップ
	mockClient := api.NewMockClient()
	mockClient.GetAllProjectsFunc = func(_ context.Context) ([]api.Project, error) {
		return mockProjects, nil
	}

	// テスト用のexecutorを作成
	executor, cleanup := createTestProjectExecutor(t, mockClient)
	defer cleanup()

	// 削除対象の確認
	params := &projectDeleteParams{
		projectIDOrName: "Nonexistent Project",
		force:           true,
	}
	_, _, err := executor.confirmProjectDeletion(context.Background(), params)

	// 結果検証（エラーが発生することを確認）
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to find project")
}

// createTestProjectExecutor はテスト用のprojectExecutorを作成するヘルパー関数
func createTestProjectExecutor(t *testing.T, mockClient api.Interface) (*projectExecutor, func()) {
	// テスト用設定（ローカルストレージ無効）
	repoConfig := &repository.Config{
		Enabled: false, // ローカルストレージ無効でAPIのみ使用
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

	executor := &projectExecutor{
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
