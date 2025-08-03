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

// insertTestProjectsIntoDB はテスト用のプロジェクトを直接DBに挿入するヘルパー関数
func insertTestProjectsIntoDB(t *testing.T, dbPath string, projects []api.Project) {
	t.Helper()

	// SQLiteDBを直接開く
	db, err := storage.NewSQLiteDB(dbPath)
	require.NoError(t, err)
	defer db.Close()

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
			tempDir := t.TempDir()
			dbPath := filepath.Join(tempDir, "test.db")

			// テストデータを直接DBに挿入
			insertTestProjectsIntoDB(t, dbPath, tt.projects)

			mockClient := api.NewMockClient()

			cfg := &config.Config{
				APIToken: "test-token",
				LocalStorage: &repository.Config{
					Enabled:      true, // ローカルストレージを有効にしてテスト
					DatabasePath: dbPath,
				},
			}

			stdout := &bytes.Buffer{}
			stderr := &bytes.Buffer{}
			output := cli.NewWithWriters(stdout, stderr, false)

			repo, err := factory.NewRepositoryForTest(mockClient, cfg.LocalStorage, false)
			require.NoError(t, err)
			defer repo.Close()

			err = repo.Initialize(context.Background())
			require.NoError(t, err)

			executor := &projectExecutor{
				cfg:        cfg,
				repository: repo,
				output:     output,
			}

			// Act: テスト対象を実行
			err = executor.executeProjectList(context.Background(), tt.params)

			// Assert: 結果を検証
			require.NoError(t, err)

			outputStr := stdout.String()
			for _, expected := range tt.expectedOutput {
				assert.Contains(t, outputStr, expected, "期待される出力が含まれていません: %s", expected)
			}
		})
	}
}
