package cmd

import (
	"context"
	"path/filepath"
	"strings"
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

	// 初期同期用（親プロジェクトをローカルに保存）
	parentProjects := []api.Project{
		{
			ID:   "parent-id-123",
			Name: "Parent Project",
		},
	}
	mockClient.SyncFunc = func(_ context.Context, _ *api.SyncRequest) (*api.SyncResponse, error) {
		return &api.SyncResponse{
			SyncToken: "initial-token",
			Projects:  parentProjects,
			Items:     []api.Item{},
			Sections:  []api.Section{},
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

	// 初期同期を実行して親プロジェクトをローカルに保存
	err := executor.repository.ForceInitialSync(context.Background())
	require.NoError(t, err)

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

	// モッククライアントをセットアップ（初期同期用）
	mockClient := api.NewMockClient()
	mockClient.SyncFunc = func(_ context.Context, _ *api.SyncRequest) (*api.SyncResponse, error) {
		// 初期同期時にモックデータを返す
		return &api.SyncResponse{
			SyncToken: "test-token",
			Projects:  mockProjects,
			Items:     []api.Item{},
			Sections:  []api.Section{},
		}, nil
	}

	// テスト用のexecutorを作成
	executor, cleanup := createTestProjectExecutor(t, mockClient)
	defer cleanup()

	// 初期同期を実行してモックデータをローカルに保存
	err := executor.repository.ForceInitialSync(context.Background())
	require.NoError(t, err)

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

	// モッククライアントをセットアップ（初期同期用）
	mockClient := api.NewMockClient()
	mockClient.SyncFunc = func(_ context.Context, _ *api.SyncRequest) (*api.SyncResponse, error) {
		// 初期同期時にモックデータを返す
		return &api.SyncResponse{
			SyncToken: "test-token",
			Projects:  mockProjects,
			Items:     []api.Item{},
			Sections:  []api.Section{},
		}, nil
	}

	// テスト用のexecutorを作成
	executor, cleanup := createTestProjectExecutor(t, mockClient)
	defer cleanup()

	// 初期同期を実行してモックデータをローカルに保存
	err := executor.repository.ForceInitialSync(context.Background())
	require.NoError(t, err)

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

	// モッククライアントをセットアップ（初期同期用）
	mockClient := api.NewMockClient()
	mockClient.SyncFunc = func(_ context.Context, _ *api.SyncRequest) (*api.SyncResponse, error) {
		return &api.SyncResponse{
			SyncToken: "initial-token",
			Projects:  mockProjects,
			Items:     []api.Item{},
			Sections:  []api.Section{},
		}, nil
	}
	mockClient.DeleteProjectFunc = func(_ context.Context, projectID string) (*api.SyncResponse, error) {
		assert.Equal(t, "test-id", projectID)
		return &api.SyncResponse{SyncToken: "delete-token"}, nil
	}

	// テスト用のexecutorを作成
	executor, cleanup := createTestProjectExecutor(t, mockClient)
	defer cleanup()

	// 初期同期を実行してモックデータをローカルに保存
	err := executor.repository.ForceInitialSync(context.Background())
	require.NoError(t, err)

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

	// モッククライアントをセットアップ（初期同期用）
	mockClient := api.NewMockClient()
	mockClient.SyncFunc = func(_ context.Context, _ *api.SyncRequest) (*api.SyncResponse, error) {
		return &api.SyncResponse{
			SyncToken: "initial-token",
			Projects:  mockProjects,
			Items:     []api.Item{},
			Sections:  []api.Section{},
		}, nil
	}

	// テスト用のexecutorを作成
	executor, cleanup := createTestProjectExecutor(t, mockClient)
	defer cleanup()

	// 初期同期を実行してモックデータをローカルに保存
	err := executor.repository.ForceInitialSync(context.Background())
	require.NoError(t, err)

	// 削除対象の確認
	params := &projectDeleteParams{
		projectIDOrName: "Nonexistent Project",
		force:           true,
	}
	_, _, err = executor.confirmProjectDeletion(context.Background(), params)

	// 結果検証（エラーが発生することを確認）
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to find project")
}

// TestProjectUpdate_Success は正常なプロジェクト更新のテスト
func TestProjectUpdate_Success(t *testing.T) {
	// 既存プロジェクトのテストデータ
	existingProjects := []api.Project{
		{ID: "update-id", Name: "Old Project Name", Color: "red", IsFavorite: false},
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
	mockClient.UpdateProjectFunc = func(_ context.Context, projectID string, req *api.UpdateProjectRequest) (*api.SyncResponse, error) {
		assert.Equal(t, "update-id", projectID)
		assert.Equal(t, "New Project Name", req.Name)
		assert.Equal(t, "blue", req.Color)
		assert.True(t, req.IsFavorite)

		return &api.SyncResponse{
			SyncToken: "update-token",
		}, nil
	}

	// テスト用のexecutorを作成
	executor, cleanup := createTestProjectExecutor(t, mockClient)
	defer cleanup()

	// 初期同期を実行して既存プロジェクトをローカルに保存
	err := executor.repository.ForceInitialSync(context.Background())
	require.NoError(t, err)

	// テスト実行
	params := &projectUpdateParams{
		projectIDOrName: "Old Project Name",
		newName:         "New Project Name",
		color:           "blue",
		isFavorite:      true,
		favoriteChanged: true,
	}
	resp, err := executor.executeProjectUpdate(context.Background(), params)

	// 結果検証
	require.NoError(t, err)
	assert.Equal(t, "update-token", resp.SyncToken)
}

// TestProjectArchive_Success は正常なプロジェクトアーカイブのテスト
func TestProjectArchive_Success(t *testing.T) {
	testProjectArchiveOperation(t, false, "archive", "archive-token")
}

// TestProjectUnarchive_Success は正常なプロジェクトアーカイブ解除のテスト
func TestProjectUnarchive_Success(t *testing.T) {
	testProjectArchiveOperation(t, true, "unarchive", "unarchive-token")
}

// testProjectArchiveOperation はアーカイブ/アンアーカイブ操作の共通テストヘルパー
func testProjectArchiveOperation(t *testing.T, isArchived bool, operation, expectedToken string) {
	t.Helper()

	// テストデータの準備
	projectID := operation + "-id"
	projectName := strings.ToUpper(operation[:1]) + operation[1:] + " Test Project"
	testProjects := []api.Project{
		{ID: projectID, Name: projectName, IsArchived: isArchived},
	}

	// モッククライアントをセットアップ
	mockClient := api.NewMockClient()
	mockClient.SyncFunc = func(_ context.Context, _ *api.SyncRequest) (*api.SyncResponse, error) {
		return &api.SyncResponse{
			SyncToken: "initial-token",
			Projects:  testProjects,
			Items:     []api.Item{},
			Sections:  []api.Section{},
		}, nil
	}

	// 操作に応じてモック関数を設定
	switch operation {
	case "archive":
		mockClient.ArchiveProjectFunc = func(_ context.Context, id string) (*api.SyncResponse, error) {
			assert.Equal(t, projectID, id)
			return &api.SyncResponse{SyncToken: expectedToken}, nil
		}
	case "unarchive":
		mockClient.UnarchiveProjectFunc = func(_ context.Context, id string) (*api.SyncResponse, error) {
			assert.Equal(t, projectID, id)
			return &api.SyncResponse{SyncToken: expectedToken}, nil
		}
	}

	// テスト用のexecutorを作成
	executor, cleanup := createTestProjectExecutor(t, mockClient)
	defer cleanup()

	// 初期同期を実行
	err := executor.repository.ForceInitialSync(context.Background())
	require.NoError(t, err)

	// テスト実行
	params := &projectArchiveParams{
		projectIDOrName: projectName,
	}

	var resp *api.SyncResponse
	switch operation {
	case "archive":
		resp, err = executor.executeProjectArchive(context.Background(), params)
	case "unarchive":
		resp, err = executor.executeProjectUnarchive(context.Background(), params)
	}

	// 結果検証
	require.NoError(t, err)
	assert.Equal(t, expectedToken, resp.SyncToken)
}

// TestFindProjectIDByName_Success はプロジェクト名からID検索のテスト
func TestFindProjectIDByName_Success(t *testing.T) {
	// テストデータ
	testProjects := []api.Project{
		{ID: "id-1", Name: "Project Alpha"},
		{ID: "id-2", Name: "Project Beta"},
		{ID: "id-3", Name: "Project Gamma"},
	}

	// モッククライアントをセットアップ
	mockClient := api.NewMockClient()
	mockClient.SyncFunc = func(_ context.Context, _ *api.SyncRequest) (*api.SyncResponse, error) {
		return &api.SyncResponse{
			SyncToken: "initial-token",
			Projects:  testProjects,
			Items:     []api.Item{},
			Sections:  []api.Section{},
		}, nil
	}

	// テスト用のexecutorを作成
	executor, cleanup := createTestProjectExecutor(t, mockClient)
	defer cleanup()

	// 初期同期を実行
	err := executor.repository.ForceInitialSync(context.Background())
	require.NoError(t, err)

	// テストケース: 名前で検索
	t.Run("名前で検索", func(t *testing.T) {
		projectID, err := executor.findProjectIDByName(context.Background(), "Project Beta")
		require.NoError(t, err)
		assert.Equal(t, "id-2", projectID)
	})

	// テストケース: IDで検索（そのまま返される）
	t.Run("IDで検索", func(t *testing.T) {
		projectID, err := executor.findProjectIDByName(context.Background(), "id-3")
		require.NoError(t, err)
		assert.Equal(t, "id-3", projectID)
	})

	// テストケース: 存在しない名前
	t.Run("存在しない名前", func(t *testing.T) {
		_, err := executor.findProjectIDByName(context.Background(), "Nonexistent Project")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "project not found")
	})
}

// TestProjectList_TreeView はツリー表示のテスト
func TestProjectList_TreeView(t *testing.T) {
	// 親子関係があるプロジェクトのテストデータ
	mockProjects := []api.Project{
		{ID: "parent-1", Name: "Parent Project 1", ParentID: ""},
		{ID: "child-1", Name: "Child Project 1", ParentID: "parent-1"},
		{ID: "child-2", Name: "Child Project 2", ParentID: "parent-1"},
		{ID: "parent-2", Name: "Parent Project 2", ParentID: ""},
	}

	// モッククライアントをセットアップ
	mockClient := api.NewMockClient()
	mockClient.SyncFunc = func(_ context.Context, _ *api.SyncRequest) (*api.SyncResponse, error) {
		return &api.SyncResponse{
			SyncToken: "test-token",
			Projects:  mockProjects,
			Items:     []api.Item{},
			Sections:  []api.Section{},
		}, nil
	}

	// テスト用のexecutorを作成
	executor, cleanup := createTestProjectExecutor(t, mockClient)
	defer cleanup()

	// 初期同期を実行
	err := executor.repository.ForceInitialSync(context.Background())
	require.NoError(t, err)

	// テスト実行（ツリー表示）
	params := &projectListParams{
		showTree:      true,
		showArchived:  false,
		showFavorites: false,
	}
	data, err := executor.fetchProjectListData(context.Background(), params)
	require.NoError(t, err)

	// 結果検証（プロジェクトが取得できることを確認）
	assert.Len(t, data.projects, 4)

	// 親プロジェクトが含まれることを確認
	var projectNames []string
	for _, project := range data.projects {
		projectNames = append(projectNames, project.Name)
	}
	assert.Contains(t, projectNames, "Parent Project 1")
	assert.Contains(t, projectNames, "Child Project 1")
	assert.Contains(t, projectNames, "Child Project 2")
	assert.Contains(t, projectNames, "Parent Project 2")
}

// createTestProjectExecutor はテスト用のprojectExecutorを作成するヘルパー関数
func createTestProjectExecutor(t *testing.T, mockClient api.Interface) (*projectExecutor, func()) {
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

	executor := &projectExecutor{
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
