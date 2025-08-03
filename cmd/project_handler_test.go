package cmd

import (
	"context"
	"testing"

	"github.com/kyokomi/gotodoist/internal/api"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// testProjectExecutorSetup はテスト用のprojectExecutor設定を保持する構造体
type testProjectExecutorSetup struct {
	executor *projectExecutor
	*testExecutorSetup
}

// setupTestProjectExecutor はテスト用のprojectExecutorをセットアップするヘルパー関数
func setupTestProjectExecutor(t *testing.T) *testProjectExecutorSetup {
	t.Helper()

	base := setupTestExecutorBase(t)

	executor := &projectExecutor{
		cfg:        base.cfg,
		repository: base.repository,
		output:     base.output,
	}

	return &testProjectExecutorSetup{
		executor:          executor,
		testExecutorSetup: base,
	}
}

func TestExecuteProjectList_Success(t *testing.T) {
	tests := []struct {
		name           string
		projects       []api.Project
		params         *projectListParams
		expectedOutput []string
	}{
		{
			name: "通常のプロジェクト一覧表示",
			projects: []api.Project{
				{ID: "1", Name: "Project 1", Color: "red"},
				{ID: "2", Name: "Project 2", Color: "blue"},
			},
			params: &projectListParams{
				showTree:      false,
				showArchived:  false,
				showFavorites: false,
			},
			expectedOutput: []string{
				"📁 Projects (2):",
				"1. 📁 Project 1",
				"2. 📁 Project 2",
			},
		},
		{
			name:     "プロジェクトが0件の場合",
			projects: []api.Project{},
			params: &projectListParams{
				showTree:      false,
				showArchived:  false,
				showFavorites: false,
			},
			expectedOutput: []string{
				"📁 No projects found",
			},
		},
		{
			name: "お気に入りプロジェクトのみ表示",
			projects: []api.Project{
				{ID: "1", Name: "Project 1", IsFavorite: true},
				{ID: "2", Name: "Project 2", IsFavorite: false},
				{ID: "3", Name: "Project 3", IsFavorite: true},
			},
			params: &projectListParams{
				showTree:      false,
				showArchived:  false,
				showFavorites: true,
			},
			expectedOutput: []string{
				"⭐ Favorite Projects (2):",
				"1. 📁 Project 1 ⭐",
				"2. 📁 Project 3 ⭐",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange: テスト環境を準備
			setup := setupTestProjectExecutor(t)
			defer setup.cleanup()

			// テストデータを直接DBに挿入
			insertTestProjectsIntoDB(t, setup.dbPath, tt.projects)

			// Act: テスト対象を実行
			err := setup.executor.executeProjectList(context.Background(), tt.params)

			// Assert: 結果を検証
			require.NoError(t, err)

			outputStr := setup.stdout.String()
			for _, expected := range tt.expectedOutput {
				assert.Contains(t, outputStr, expected, "期待される出力が含まれていません: %s", expected)
			}
		})
	}
}

func TestExecuteProjectAddWithOutput_Success(t *testing.T) {
	tests := []struct {
		name           string
		params         *projectAddParams
		expectedOutput []string
	}{
		{
			name: "通常のプロジェクト追加",
			params: &projectAddParams{
				name:       "New Project",
				color:      "blue",
				isFavorite: false,
			},
			expectedOutput: []string{
				"✅ 📁 Project created successfully!",
				"Name: New Project",
				"Color: blue",
			},
		},
		{
			name: "お気に入りプロジェクトの追加",
			params: &projectAddParams{
				name:       "Favorite Project",
				color:      "red",
				isFavorite: true,
			},
			expectedOutput: []string{
				"✅ 📁 Project created successfully!",
				"Name: Favorite Project",
				"Color: red",
				"Favorite: Yes ⭐",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange: テスト環境を準備
			setup := setupTestProjectExecutor(t)
			defer setup.cleanup()

			// CreateProjectのmockを設定
			setup.mockClient.CreateProjectFunc = func(_ context.Context, req *api.CreateProjectRequest) (*api.SyncResponse, error) {
				return &api.SyncResponse{
					Projects: []api.Project{
						{
							ID:         "new-project-id",
							Name:       req.Name,
							Color:      req.Color,
							IsFavorite: req.IsFavorite,
						},
					},
					SyncToken: "test-token",
				}, nil
			}

			// Act: テスト対象を実行
			err := setup.executor.executeProjectAddWithOutput(context.Background(), tt.params)

			// Assert: 結果を検証
			require.NoError(t, err)

			outputStr := setup.stdout.String()
			for _, expected := range tt.expectedOutput {
				assert.Contains(t, outputStr, expected, "期待される出力が含まれていません: %s", expected)
			}
		})
	}
}

func TestExecuteProjectUpdateWithOutput_Success(t *testing.T) {
	tests := []struct {
		name            string
		existingProject api.Project
		params          *projectUpdateParams
		expectedOutput  []string
	}{
		{
			name: "プロジェクト名の更新",
			existingProject: api.Project{
				ID:         "project-1",
				Name:       "Old Project",
				Color:      "blue",
				IsFavorite: false,
			},
			params: &projectUpdateParams{
				projectIDOrName: "project-1",
				newName:         "Updated Project",
				color:           "",
				isFavorite:      false,
				favoriteChanged: false,
			},
			expectedOutput: []string{
				"✏️  Project updated successfully!",
				"New name: Updated Project",
			},
		},
		{
			name: "プロジェクトの色とお気に入りを更新",
			existingProject: api.Project{
				ID:         "project-2",
				Name:       "Test Project",
				Color:      "red",
				IsFavorite: false,
			},
			params: &projectUpdateParams{
				projectIDOrName: "project-2",
				newName:         "",
				color:           "green",
				isFavorite:      true,
				favoriteChanged: true,
			},
			expectedOutput: []string{
				"✏️  Project updated successfully!",
				"Color: green",
				"Favorite: Yes ⭐",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange: テスト環境を準備
			setup := setupTestProjectExecutor(t)
			defer setup.cleanup()

			// 既存プロジェクトをDBに挿入
			insertTestProjectsIntoDB(t, setup.dbPath, []api.Project{tt.existingProject})

			// UpdateProjectのmockを設定
			setup.mockClient.UpdateProjectFunc = func(_ context.Context, _ string, _ *api.UpdateProjectRequest) (*api.SyncResponse, error) {
				return &api.SyncResponse{
					SyncToken: "updated-token",
				}, nil
			}

			// Act: テスト対象を実行
			err := setup.executor.executeProjectUpdateWithOutput(context.Background(), tt.params)

			// Assert: 結果を検証
			require.NoError(t, err)

			outputStr := setup.stdout.String()
			for _, expected := range tt.expectedOutput {
				assert.Contains(t, outputStr, expected, "期待される出力が含まれていません: %s", expected)
			}
		})
	}
}

func TestExecuteProjectDeleteWithOutput_Success(t *testing.T) {
	tests := []struct {
		name            string
		existingProject api.Project
		params          *projectDeleteParams
		expectedOutput  []string
	}{
		{
			name: "プロジェクト削除（force=true）",
			existingProject: api.Project{
				ID:         "project-to-delete",
				Name:       "Test Project",
				Color:      "red",
				IsFavorite: true,
			},
			params: &projectDeleteParams{
				projectIDOrName: "project-to-delete",
				force:           true, // 確認プロンプトをスキップ
			},
			expectedOutput: []string{
				"🗑️  Project deleted successfully!",
				"Deleted: Test Project",
			},
		},
		{
			name: "プロジェクト削除（名前で指定）",
			existingProject: api.Project{
				ID:         "another-project",
				Name:       "Another Project",
				Color:      "blue",
				IsFavorite: false,
			},
			params: &projectDeleteParams{
				projectIDOrName: "Another Project",
				force:           true,
			},
			expectedOutput: []string{
				"🗑️  Project deleted successfully!",
				"Deleted: Another Project",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange: テスト環境を準備
			setup := setupTestProjectExecutor(t)
			defer setup.cleanup()

			// 既存プロジェクトをDBに挿入
			insertTestProjectsIntoDB(t, setup.dbPath, []api.Project{tt.existingProject})

			// DeleteProjectのmockを設定
			setup.mockClient.DeleteProjectFunc = func(_ context.Context, _ string) (*api.SyncResponse, error) {
				return &api.SyncResponse{
					SyncToken: "deleted-token",
				}, nil
			}

			// Act: テスト対象を実行
			err := setup.executor.executeProjectDeleteWithOutput(context.Background(), tt.params)

			// Assert: 結果を検証
			require.NoError(t, err)

			outputStr := setup.stdout.String()
			for _, expected := range tt.expectedOutput {
				assert.Contains(t, outputStr, expected, "期待される出力が含まれていません: %s", expected)
			}
		})
	}
}

func TestExecuteProjectArchiveWithOutput_Success(t *testing.T) {
	existingProject := api.Project{
		ID:         "project-archive",
		Name:       "Project To Archive",
		Color:      "green",
		IsArchived: false,
	}
	params := &projectArchiveParams{
		projectIDOrName: "project-archive",
	}

	// Arrange: テスト環境を準備
	setup := setupTestProjectExecutor(t)
	defer setup.cleanup()

	// 既存プロジェクトをDBに挿入
	insertTestProjectsIntoDB(t, setup.dbPath, []api.Project{existingProject})

	// ArchiveProjectのmockを設定
	setup.mockClient.ArchiveProjectFunc = func(_ context.Context, _ string) (*api.SyncResponse, error) {
		return &api.SyncResponse{
			SyncToken: "archived-token",
		}, nil
	}

	// Act: テスト対象を実行
	err := setup.executor.executeProjectArchiveWithOutput(context.Background(), params)

	// Assert: 結果を検証
	require.NoError(t, err)

	outputStr := setup.stdout.String()
	assert.Contains(t, outputStr, "📦 Project archived successfully!", "期待される出力が含まれていません")
}

func TestExecuteProjectUnarchiveWithOutput_Success(t *testing.T) {
	existingProject := api.Project{
		ID:         "project-unarchive",
		Name:       "Project To Unarchive",
		Color:      "purple",
		IsArchived: true,
	}
	params := &projectArchiveParams{
		projectIDOrName: "project-unarchive",
	}

	// Arrange: テスト環境を準備
	setup := setupTestProjectExecutor(t)
	defer setup.cleanup()

	// 既存プロジェクトをDBに挿入
	insertTestProjectsIntoDB(t, setup.dbPath, []api.Project{existingProject})

	// UnarchiveProjectのmockを設定
	setup.mockClient.UnarchiveProjectFunc = func(_ context.Context, _ string) (*api.SyncResponse, error) {
		return &api.SyncResponse{
			SyncToken: "unarchived-token",
		}, nil
	}

	// Act: テスト対象を実行
	err := setup.executor.executeProjectUnarchiveWithOutput(context.Background(), params)

	// Assert: 結果を検証
	require.NoError(t, err)

	outputStr := setup.stdout.String()
	assert.Contains(t, outputStr, "📁 Project unarchived successfully!", "期待される出力が含まれていません")
}
